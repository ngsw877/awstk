package cfn

import (
	"context"
	"fmt"
	"strings"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

// EcsServiceInfo はECSサービスの情報を格納する構造体（ローカル定義）
type EcsServiceInfo struct {
	ClusterName string
	ServiceName string
}

// GetStackResources はスタックからリソース一覧を取得する関数
func GetStackResources(cfnClient *cloudformation.Client, stackName string) ([]types.StackResource, error) {
	ctx := context.Background()

	// スタックからリソースを取得
	fmt.Printf("🔍 スタック '%s' からリソースを検索中...\n", stackName)
	resp, err := cfnClient.DescribeStackResources(ctx, &cloudformation.DescribeStackResourcesInput{
		StackName: awssdk.String(stackName),
	})
	if err != nil {
		return nil, fmt.Errorf("CloudFormationスタックのリソース取得に失敗: %w", err)
	}

	// スタック存在確認
	if len(resp.StackResources) == 0 {
		return nil, fmt.Errorf("スタック '%s' にリソースが見つかりませんでした", stackName)
	}

	return resp.StackResources, nil
}

// GetCleanupResourcesFromStack はCloudFormationスタックからS3バケットとECRリポジトリのリソース一覧を取得します
func GetCleanupResourcesFromStack(cfnClient *cloudformation.Client, stackName string) ([]string, []string, error) {
	// 共通関数を使用してスタックリソースを取得
	stackResources, err := GetStackResources(cfnClient, stackName)
	if err != nil {
		return nil, nil, err
	}

	// S3バケットとECRリポジトリを抽出
	s3Resources := []string{}
	ecrResources := []string{}

	for _, resource := range stackResources {
		// リソースタイプに基づいて振り分け
		resourceType := *resource.ResourceType

		// S3バケット
		if resourceType == "AWS::S3::Bucket" && resource.PhysicalResourceId != nil {
			s3Resources = append(s3Resources, *resource.PhysicalResourceId)
			fmt.Printf("🔍 検出されたS3バケット: %s\n", *resource.PhysicalResourceId)
		}

		// ECRリポジトリ
		if resourceType == "AWS::ECR::Repository" && resource.PhysicalResourceId != nil {
			ecrResources = append(ecrResources, *resource.PhysicalResourceId)
			fmt.Printf("🔍 検出されたECRリポジトリ: %s\n", *resource.PhysicalResourceId)
		}
	}

	return s3Resources, ecrResources, nil
}

// getStartStopResourcesFromStack はCloudFormationスタックから起動・停止可能なリソースの識別子を取得します
func getStartStopResourcesFromStack(cfnClient *cloudformation.Client, stackName string) (StackResources, error) {
	var result StackResources

	// 共通関数を使用してスタックリソースを取得
	stackResources, err := GetStackResources(cfnClient, stackName)
	if err != nil {
		return result, err
	}

	// Auroraクラスターの存在フラグ
	hasAuroraCluster := false

	// 各リソースタイプをフィルタリング
	for _, resource := range stackResources {
		if resource.PhysicalResourceId == nil || *resource.PhysicalResourceId == "" {
			continue
		}

		switch *resource.ResourceType {
		case "AWS::RDS::DBCluster":
			// Aurora DBクラスターを検出した場合、フラグを立てる
			hasAuroraCluster = true
			result.AuroraClusterIds = append(result.AuroraClusterIds, *resource.PhysicalResourceId)
		case "AWS::RDS::DBInstance":
			// Aurora DBクラスターが存在しない場合のみ、純粋なRDSインスタンスとして扱う
			if !hasAuroraCluster {
				result.RdsInstanceIds = append(result.RdsInstanceIds, *resource.PhysicalResourceId)
			}
		case "AWS::EC2::Instance":
			result.Ec2InstanceIds = append(result.Ec2InstanceIds, *resource.PhysicalResourceId)
		case "AWS::ECS::Service":
			// ECSサービスARNからクラスター名とサービス名を抽出
			serviceArn := *resource.PhysicalResourceId
			parts := strings.Split(serviceArn, "/")
			if len(parts) >= 2 {
				clusterName := parts[len(parts)-2]
				serviceName := parts[len(parts)-1]

				// クラスター名を正規化（ARNの場合は名前部分のみ抽出）
				if strings.Contains(clusterName, "/") {
					clusterParts := strings.Split(clusterName, "/")
					clusterName = clusterParts[len(clusterParts)-1]
				}

				result.EcsServiceInfo = append(result.EcsServiceInfo, EcsServiceInfo{
					ClusterName: clusterName,
					ServiceName: serviceName,
				})
			}
		}
	}

	return result, nil
}

// printResourcesSummary はスタック内の検出されたリソースサマリーを表示します
func printResourcesSummary(resources StackResources) {
	fmt.Println("📋 検出されたリソース:")

	if len(resources.Ec2InstanceIds) > 0 {
		fmt.Println("  EC2インスタンス:")
		for _, id := range resources.Ec2InstanceIds {
			fmt.Println("   - " + id)
		}
	}

	if len(resources.RdsInstanceIds) > 0 {
		fmt.Println("  RDSインスタンス:")
		for _, id := range resources.RdsInstanceIds {
			fmt.Println("   - " + id)
		}
	}

	if len(resources.AuroraClusterIds) > 0 {
		fmt.Println("  Aurora DBクラスター:")
		for _, id := range resources.AuroraClusterIds {
			fmt.Println("   - " + id)
		}
	}

	if len(resources.EcsServiceInfo) > 0 {
		fmt.Println("  ECSサービス:")
		for _, info := range resources.EcsServiceInfo {
			fmt.Printf("   - %s/%s\n", info.ClusterName, info.ServiceName)
		}
	}

	if len(resources.Ec2InstanceIds) == 0 &&
		len(resources.RdsInstanceIds) == 0 &&
		len(resources.AuroraClusterIds) == 0 &&
		len(resources.EcsServiceInfo) == 0 {
		fmt.Println("  操作可能なリソースは見つかりませんでした")
	}
}
