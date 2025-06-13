package service

import (
	"awstk/internal/aws"
	"context"
	"fmt"
	"strings"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/rds"
)

// ListCfnStacks はアクティブなCloudFormationスタック名一覧を返す
func ListCfnStacks(cfnClient *cloudformation.Client) ([]string, error) {
	activeStatusStrs := []string{
		"CREATE_COMPLETE",
		"UPDATE_COMPLETE",
		"UPDATE_ROLLBACK_COMPLETE",
		"ROLLBACK_COMPLETE",
		"IMPORT_COMPLETE",
	}
	activeStatuses := make([]types.StackStatus, 0, len(activeStatusStrs))
	for _, s := range activeStatusStrs {
		activeStatuses = append(activeStatuses, types.StackStatus(s))
	}

	// すべてのスタックを格納するスライス
	var allStackNames []string

	// ページネーション用のトークン
	var nextToken *string

	// すべてのページを取得するまでループ
	for {
		input := &cloudformation.ListStacksInput{
			StackStatusFilter: activeStatuses,
			NextToken:         nextToken,
		}

		resp, err := cfnClient.ListStacks(context.Background(), input)
		if err != nil {
			return nil, fmt.Errorf("スタック一覧取得エラー: %w", err)
		}

		// 現在のページのスタック名をスライスに追加
		for _, summary := range resp.StackSummaries {
			allStackNames = append(allStackNames, awssdk.ToString(summary.StackName))
		}

		// 次のページがあるかチェック
		nextToken = resp.NextToken
		if nextToken == nil {
			// 次のページがなければループを抜ける
			break
		}
	}
	return allStackNames, nil
}

// 共通処理：スタックからリソース一覧を取得する内部関数
func getStackResources(awsCtx aws.AwsContext, stackName string) ([]types.StackResource, error) {
	ctx := context.Background()
	cfg, err := aws.LoadAwsConfig(aws.AwsContext{
		Profile: awsCtx.Profile,
		Region:  awsCtx.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("AWS設定のロードに失敗: %w", err)
	}

	// CloudFormationクライアントを作成
	cfnClient := cloudformation.NewFromConfig(cfg)

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

// getCleanupResourcesFromStack はCloudFormationスタックからS3バケットとECRリポジトリのリソース一覧を取得します
func getCleanupResourcesFromStack(opts CleanupOptions) ([]string, []string, error) {
	// 共通関数を使用してスタックリソースを取得
	stackResources, err := getStackResources(opts.AwsContext, opts.StackName)
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

// StackResources はCloudFormationスタック内のリソース識別子を格納する構造体
type StackResources struct {
	Ec2InstanceIds   []string
	RdsInstanceIds   []string
	AuroraClusterIds []string
	EcsServiceInfo   []EcsServiceInfo
}

// GetStartStopResourcesFromStack はCloudFormationスタックから起動・停止可能なリソースの識別子を取得します
func GetStartStopResourcesFromStack(awsCtx aws.AwsContext, stackName string) (StackResources, error) {
	var result StackResources

	// 共通関数を使用してスタックリソースを取得
	stackResources, err := getStackResources(awsCtx, stackName)
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
			// 実際のスタックでは、AuroraクラスターとRDSインスタンスが混在することは稀で、
			// Auroraスタックの場合はクラスター単位での操作が基本となる
			hasAuroraCluster = true
			result.AuroraClusterIds = append(result.AuroraClusterIds, *resource.PhysicalResourceId)
		case "AWS::RDS::DBInstance":
			// Aurora DBクラスターが存在しない場合のみ、純粋なRDSインスタンスとして扱う
			// 理由: Auroraスタックでは、DBInstanceはDBClusterの一部として作成されるため、
			// クラスター単位での制御が適切。個別のDBInstance操作は不要かつ非推奨
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

// StartAllStackResources はスタック内のすべてのリソースを起動します
func StartAllStackResources(awsCtx aws.AwsContext, stackName string) error {
	// スタックからリソースを取得（名前変更された関数を使用）
	resources, err := GetStartStopResourcesFromStack(awsCtx, stackName)
	if err != nil {
		return err
	}

	// 検出されたリソースのサマリーを表示
	printResourcesSummary(resources)

	errorsOccurred := false

	// 必要に応じて各種クライアントを作成
	cfg, err := aws.LoadAwsConfig(aws.AwsContext{
		Profile: awsCtx.Profile,
		Region:  awsCtx.Region,
	})
	if err != nil {
		return fmt.Errorf("AWS設定の読み込みエラー: %w", err)
	}

	// EC2インスタンスを起動
	if len(resources.Ec2InstanceIds) > 0 {
		ec2Client := ec2.NewFromConfig(cfg)
		for _, instanceId := range resources.Ec2InstanceIds {
			fmt.Printf("🚀 EC2インスタンス (%s) を起動します...\n", instanceId)
			if err := StartEc2Instance(ec2Client, instanceId); err != nil {
				fmt.Printf("❌ EC2インスタンス (%s) の起動中にエラーが発生しました: %v\n", instanceId, err)
				errorsOccurred = true
			} else {
				fmt.Printf("✅ EC2インスタンス (%s) の起動を開始しました\n", instanceId)
			}
		}
	}

	// RDSインスタンスとAuroraクラスターを起動
	if len(resources.RdsInstanceIds) > 0 || len(resources.AuroraClusterIds) > 0 {
		rdsClient := rds.NewFromConfig(cfg)

		// RDSインスタンスを起動
		for _, instanceId := range resources.RdsInstanceIds {
			fmt.Printf("🚀 RDSインスタンス (%s) を起動します...\n", instanceId)
			if err := StartRdsInstance(rdsClient, instanceId); err != nil {
				fmt.Printf("❌ RDSインスタンス (%s) の起動中にエラーが発生しました: %v\n", instanceId, err)
				errorsOccurred = true
			} else {
				fmt.Printf("✅ RDSインスタンス (%s) の起動を開始しました\n", instanceId)
			}
		}

		// Auroraクラスターを起動
		for _, clusterId := range resources.AuroraClusterIds {
			fmt.Printf("🚀 Aurora DBクラスター (%s) を起動します...\n", clusterId)
			if err := StartAuroraCluster(rdsClient, clusterId); err != nil {
				fmt.Printf("❌ Aurora DBクラスター (%s) の起動中にエラーが発生しました: %v\n", clusterId, err)
				errorsOccurred = true
			} else {
				fmt.Printf("✅ Aurora DBクラスター (%s) の起動を開始しました\n", clusterId)
			}
		}
	}

	// ECSサービスを起動
	if len(resources.EcsServiceInfo) > 0 {
		autoScalingClient := applicationautoscaling.NewFromConfig(cfg)
		for _, ecsInfo := range resources.EcsServiceInfo {
			fmt.Printf("🚀 ECSサービス (%s/%s) を起動します...\n", ecsInfo.ClusterName, ecsInfo.ServiceName)
			opts := ServiceCapacityOptions{
				ClusterName: ecsInfo.ClusterName,
				ServiceName: ecsInfo.ServiceName,
				MinCapacity: 1, // デフォルト値として1を使用
				MaxCapacity: 2, // デフォルト値として2を使用
			}

			if err := SetEcsServiceCapacity(autoScalingClient, opts); err != nil {
				fmt.Printf("❌ ECSサービス (%s/%s) の起動中にエラーが発生しました: %v\n",
					ecsInfo.ClusterName, ecsInfo.ServiceName, err)
				errorsOccurred = true
			} else {
				fmt.Printf("✅ ECSサービス (%s/%s) の起動を開始しました\n",
					ecsInfo.ClusterName, ecsInfo.ServiceName)
			}
		}
	}

	if errorsOccurred {
		return fmt.Errorf("一部のリソースの起動中にエラーが発生しました")
	}
	return nil
}

// StopAllStackResources はスタック内のすべてのリソースを停止します
func StopAllStackResources(awsCtx aws.AwsContext, stackName string) error {
	// スタックからリソースを取得（名前変更された関数を使用）
	resources, err := GetStartStopResourcesFromStack(awsCtx, stackName)
	if err != nil {
		return err
	}

	// 検出されたリソースのサマリーを表示
	printResourcesSummary(resources)

	errorsOccurred := false

	// 必要に応じて各種クライアントを作成
	cfg, err := aws.LoadAwsConfig(aws.AwsContext{
		Profile: awsCtx.Profile,
		Region:  awsCtx.Region,
	})
	if err != nil {
		return fmt.Errorf("AWS設定の読み込みエラー: %w", err)
	}

	// ECSサービスを停止（他のリソースより先に停止）
	if len(resources.EcsServiceInfo) > 0 {
		autoScalingClient := applicationautoscaling.NewFromConfig(cfg)
		for _, ecsInfo := range resources.EcsServiceInfo {
			fmt.Printf("🛑 ECSサービス (%s/%s) を停止します...\n", ecsInfo.ClusterName, ecsInfo.ServiceName)
			opts := ServiceCapacityOptions{
				ClusterName: ecsInfo.ClusterName,
				ServiceName: ecsInfo.ServiceName,
				MinCapacity: 0,
				MaxCapacity: 0,
			}

			if err := SetEcsServiceCapacity(autoScalingClient, opts); err != nil {
				fmt.Printf("❌ ECSサービス (%s/%s) の停止中にエラーが発生しました: %v\n",
					ecsInfo.ClusterName, ecsInfo.ServiceName, err)
				errorsOccurred = true
			} else {
				fmt.Printf("✅ ECSサービス (%s/%s) の停止を開始しました\n",
					ecsInfo.ClusterName, ecsInfo.ServiceName)
			}
		}
	}

	// EC2インスタンスを停止
	if len(resources.Ec2InstanceIds) > 0 {
		ec2Client := ec2.NewFromConfig(cfg)
		for _, instanceId := range resources.Ec2InstanceIds {
			fmt.Printf("🛑 EC2インスタンス (%s) を停止します...\n", instanceId)
			if err := StopEc2Instance(ec2Client, instanceId); err != nil {
				fmt.Printf("❌ EC2インスタンス (%s) の停止中にエラーが発生しました: %v\n", instanceId, err)
				errorsOccurred = true
			} else {
				fmt.Printf("✅ EC2インスタンス (%s) の停止を開始しました\n", instanceId)
			}
		}
	}

	// RDSインスタンスとAuroraクラスターを停止
	if len(resources.RdsInstanceIds) > 0 || len(resources.AuroraClusterIds) > 0 {
		rdsClient := rds.NewFromConfig(cfg)

		// RDSインスタンスを停止
		for _, instanceId := range resources.RdsInstanceIds {
			fmt.Printf("🛑 RDSインスタンス (%s) を停止します...\n", instanceId)
			if err := StopRdsInstance(rdsClient, instanceId); err != nil {
				fmt.Printf("❌ RDSインスタンス (%s) の停止中にエラーが発生しました: %v\n", instanceId, err)
				errorsOccurred = true
			} else {
				fmt.Printf("✅ RDSインスタンス (%s) の停止を開始しました\n", instanceId)
			}
		}

		// Auroraクラスターを停止
		for _, clusterId := range resources.AuroraClusterIds {
			fmt.Printf("🛑 Aurora DBクラスター (%s) を停止します...\n", clusterId)
			if err := StopAuroraCluster(rdsClient, clusterId); err != nil {
				fmt.Printf("❌ Aurora DBクラスター (%s) の停止中にエラーが発生しました: %v\n", clusterId, err)
				errorsOccurred = true
			} else {
				fmt.Printf("✅ Aurora DBクラスター (%s) の停止を開始しました\n", clusterId)
			}
		}
	}

	if errorsOccurred {
		return fmt.Errorf("一部のリソースの停止中にエラーが発生しました")
	}
	return nil
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
