package cfn

import (
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
)

// GetEc2FromStack はCloudFormationスタックからEC2インスタンスIDを取得します
func GetEc2FromStack(cfnClient *cloudformation.Client, stackName string) (string, error) {
	allInstances, err := GetAllEc2FromStack(cfnClient, stackName)
	if err != nil {
		return "", err
	}

	if len(allInstances) == 0 {
		return "", fmt.Errorf("スタック '%s' にEC2インスタンスが見つかりませんでした", stackName)
	}

	// 複数のインスタンスがある場合は最初の要素を返す
	return allInstances[0], nil
}

// GetAllEc2FromStack はCloudFormationスタックからすべてのEC2インスタンス識別子を取得します
func GetAllEc2FromStack(cfnClient *cloudformation.Client, stackName string) ([]string, error) {
	// 共通関数を使用してスタックリソースを取得
	stackResources, err := GetStackResources(cfnClient, stackName)
	if err != nil {
		return nil, err
	}

	var instanceIds []string
	for _, resource := range stackResources {
		if *resource.ResourceType == "AWS::EC2::Instance" && resource.PhysicalResourceId != nil {
			instanceIds = append(instanceIds, *resource.PhysicalResourceId)
			fmt.Printf("🔍 検出されたEC2インスタンス: %s\n", *resource.PhysicalResourceId)
		}
	}

	return instanceIds, nil
}

// GetRdsFromStack はCloudFormationスタックからRDSインスタンス識別子を取得します
func GetRdsFromStack(cfnClient *cloudformation.Client, stackName string) (string, error) {
	allInstances, err := GetAllRdsFromStack(cfnClient, stackName)
	if err != nil {
		return "", err
	}

	if len(allInstances) == 0 {
		return "", fmt.Errorf("スタック '%s' にRDSインスタンスが見つかりませんでした", stackName)
	}

	// 複数のインスタンスがある場合は最初の要素を返す
	return allInstances[0], nil
}

// GetAllRdsFromStack はCloudFormationスタックからすべてのRDSインスタンス識別子を取得します
func GetAllRdsFromStack(cfnClient *cloudformation.Client, stackName string) ([]string, error) {
	// 共通関数を使用してスタックリソースを取得
	stackResources, err := GetStackResources(cfnClient, stackName)
	if err != nil {
		return nil, err
	}

	var instanceIds []string
	for _, resource := range stackResources {
		if *resource.ResourceType == "AWS::RDS::DBInstance" && resource.PhysicalResourceId != nil {
			instanceIds = append(instanceIds, *resource.PhysicalResourceId)
			fmt.Printf("🔍 検出されたRDSインスタンス: %s\n", *resource.PhysicalResourceId)
		}
	}

	return instanceIds, nil
}

// GetAuroraFromStack はCloudFormationスタックからAuroraクラスター識別子を取得します
func GetAuroraFromStack(cfnClient *cloudformation.Client, stackName string) (string, error) {
	allClusters, err := GetAllAuroraFromStack(cfnClient, stackName)
	if err != nil {
		return "", err
	}

	if len(allClusters) == 0 {
		return "", fmt.Errorf("スタック '%s' にAuroraクラスターが見つかりませんでした", stackName)
	}

	// 複数のクラスターがある場合は最初の要素を返す
	return allClusters[0], nil
}

// GetAllAuroraFromStack はCloudFormationスタックからすべてのAuroraクラスター識別子を取得します
func GetAllAuroraFromStack(cfnClient *cloudformation.Client, stackName string) ([]string, error) {
	// 共通関数を使用してスタックリソースを取得
	stackResources, err := GetStackResources(cfnClient, stackName)
	if err != nil {
		return nil, err
	}

	var clusterIds []string
	for _, resource := range stackResources {
		if *resource.ResourceType == "AWS::RDS::DBCluster" && resource.PhysicalResourceId != nil {
			clusterIds = append(clusterIds, *resource.PhysicalResourceId)
			fmt.Printf("🔍 検出されたAuroraクラスター: %s\n", *resource.PhysicalResourceId)
		}
	}

	return clusterIds, nil
}

// GetEcsFromStack はCloudFormationスタックからECSサービス情報を取得します
func GetEcsFromStack(cfnClient *cloudformation.Client, stackName string) (EcsServiceInfo, error) {
	allServices, err := GetAllEcsFromStack(cfnClient, stackName)
	if err != nil {
		return EcsServiceInfo{}, err
	}

	if len(allServices) == 0 {
		return EcsServiceInfo{}, fmt.Errorf("スタック '%s' にECSサービスが見つかりませんでした", stackName)
	}

	// 複数のサービスがある場合は最初の要素を返す
	return allServices[0], nil
}

// GetAllEcsFromStack はCloudFormationスタックからすべてのECSサービス識別子を取得します
func GetAllEcsFromStack(cfnClient *cloudformation.Client, stackName string) ([]EcsServiceInfo, error) {
	// 共通関数を使用してスタックリソースを取得
	stackResources, err := GetStackResources(cfnClient, stackName)
	if err != nil {
		return nil, err
	}

	var results []EcsServiceInfo

	// クラスターリソースをフィルタリング
	var clusterPhysicalIds []string
	for _, resource := range stackResources {
		if *resource.ResourceType == "AWS::ECS::Cluster" {
			clusterPhysicalIds = append(clusterPhysicalIds, *resource.PhysicalResourceId)
			fmt.Printf("🔍 検出されたECSクラスター: %s\n", *resource.PhysicalResourceId)
		}
	}

	if len(clusterPhysicalIds) == 0 {
		return results, errors.New("スタック '" + stackName + "' からECSクラスターを検出できませんでした")
	}

	// サービスリソースをフィルタリング
	fmt.Println("🔍 スタック '" + stackName + "' からECSサービスを検索中...")
	var servicePhysicalIds []string
	for _, resource := range stackResources {
		if *resource.ResourceType == "AWS::ECS::Service" {
			servicePhysicalIds = append(servicePhysicalIds, *resource.PhysicalResourceId)
		}
	}

	if len(servicePhysicalIds) == 0 {
		return results, errors.New("スタック '" + stackName + "' からECSサービスを検出できませんでした")
	}

	// 各サービスについてクラスターとの組み合わせを作成
	for _, serviceArn := range servicePhysicalIds {
		// サービス名を抽出 (形式: arn:aws:ecs:REGION:ACCOUNT:service/CLUSTER/SERVICE_NAME)
		parts := strings.Split(serviceArn, "/")
		if len(parts) < 2 {
			continue // 不正な形式はスキップ
		}

		clusterNameFromArn := parts[len(parts)-2]
		serviceName := parts[len(parts)-1]

		// ARNから抽出したクラスター名がスタック内のクラスターと一致するかチェック
		var matchedClusterName string
		for _, clusterId := range clusterPhysicalIds {
			// クラスター名の完全一致またはクラスターARNの末尾一致をチェック
			if clusterId == clusterNameFromArn || strings.HasSuffix(clusterId, "/"+clusterNameFromArn) {
				matchedClusterName = clusterId
				break
			}
		}

		// マッチしたクラスターがある場合のみ追加
		if matchedClusterName != "" {
			// クラスター名を正規化（ARNの場合は名前部分のみ抽出）
			displayClusterName := matchedClusterName
			if strings.Contains(matchedClusterName, "/") {
				clusterParts := strings.Split(matchedClusterName, "/")
				displayClusterName = clusterParts[len(clusterParts)-1]
			}

			results = append(results, EcsServiceInfo{
				ClusterName: displayClusterName,
				ServiceName: serviceName,
			})
			fmt.Printf("🔍 検出されたECSサービス: %s/%s\n", displayClusterName, serviceName)
		} else {
			fmt.Printf("⚠️ 警告: サービス %s のクラスター %s がスタック内で見つかりませんでした\n", serviceName, clusterNameFromArn)
		}
	}

	if len(results) == 0 {
		return results, errors.New("スタック '" + stackName + "' から有効なECSサービスを検出できませんでした")
	}

	return results, nil
}
