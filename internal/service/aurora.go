package service

import (
	"context"
	"fmt"

	"awstk/internal/aws"

	"github.com/aws/aws-sdk-go-v2/service/rds"
)

// StartAuroraCluster Aurora DBクラスターを起動する
func StartAuroraCluster(rdsClient *rds.Client, clusterId string) error {
	input := &rds.StartDBClusterInput{
		DBClusterIdentifier: &clusterId,
	}

	_, err := rdsClient.StartDBCluster(context.Background(), input)
	if err != nil {
		return fmt.Errorf("Aurora DBクラスター起動エラー: %w", err)
	}

	return nil
}

// StopAuroraCluster Aurora DBクラスターを停止する
func StopAuroraCluster(rdsClient *rds.Client, clusterId string) error {
	input := &rds.StopDBClusterInput{
		DBClusterIdentifier: &clusterId,
	}

	_, err := rdsClient.StopDBCluster(context.Background(), input)
	if err != nil {
		return fmt.Errorf("Aurora DBクラスター停止エラー: %w", err)
	}

	return nil
}

// GetAuroraFromStack はCloudFormationスタックからAurora DBクラスター識別子を取得します
func GetAuroraFromStack(awsCtx aws.AwsContext, stackName string) (string, error) {
	allClusters, err := GetAllAuroraFromStack(awsCtx, stackName)
	if err != nil {
		return "", err
	}

	if len(allClusters) == 0 {
		return "", fmt.Errorf("スタック '%s' にAurora DBクラスターが見つかりませんでした", stackName)
	}

	// 複数のクラスターがある場合は最初の要素を返す
	return allClusters[0], nil
}

// GetAllAuroraFromStack はCloudFormationスタックからすべてのAurora DBクラスター識別子を取得します
func GetAllAuroraFromStack(awsCtx aws.AwsContext, stackName string) ([]string, error) {
	// 共通関数を使用してスタックリソースを取得
	stackResources, err := getStackResources(awsCtx, stackName)
	if err != nil {
		return nil, err
	}

	var clusterIds []string
	for _, resource := range stackResources {
		if *resource.ResourceType == "AWS::RDS::DBCluster" && resource.PhysicalResourceId != nil {
			clusterIds = append(clusterIds, *resource.PhysicalResourceId)
			fmt.Printf("🔍 検出されたAurora DBクラスター: %s\n", *resource.PhysicalResourceId)
		}
	}

	return clusterIds, nil
}
