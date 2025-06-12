package internal

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
)

// StartAuroraCluster Aurora DBクラスターを起動する
func StartAuroraCluster(rdsClient *rds.Client, clusterId string) error {
	_, err := rdsClient.StartDBCluster(context.Background(), &rds.StartDBClusterInput{
		DBClusterIdentifier: aws.String(clusterId),
	})
	if err != nil {
		return fmt.Errorf("❌ Aurora DBクラスターの起動に失敗: %w", err)
	}
	return nil
}

// StopAuroraCluster Aurora DBクラスターを停止する
func StopAuroraCluster(rdsClient *rds.Client, clusterId string) error {
	_, err := rdsClient.StopDBCluster(context.Background(), &rds.StopDBClusterInput{
		DBClusterIdentifier: aws.String(clusterId),
	})
	if err != nil {
		return fmt.Errorf("❌ Aurora DBクラスターの停止に失敗: %w", err)
	}
	return nil
}

// GetAuroraFromStack はCloudFormationスタック名からAurora DBクラスター識別子を取得します。
func GetAuroraFromStack(awsCtx AwsContext, stackName string) (string, error) {
	clusters, err := GetAllAuroraFromStack(awsCtx, stackName)
	if err != nil {
		return "", err
	}

	// 対話的選択機能を使用
	selectedIndex, err := SelectFromOptions("複数のAurora DBクラスターが見つかりました", clusters)
	if err != nil {
		return "", err
	}

	return clusters[selectedIndex], nil
}

// GetAllAuroraFromStack はスタック内のすべてのAurora DBクラスター識別子を取得します
func GetAllAuroraFromStack(awsCtx AwsContext, stackName string) ([]string, error) {
	var results []string

	stackResources, err := getStackResources(awsCtx, stackName)
	if err != nil {
		return results, fmt.Errorf("CloudFormationスタックのリソース取得に失敗: %w", err)
	}

	// リソースの中からRDS DBClusterを探す
	for _, resource := range stackResources {
		if resource.ResourceType != nil && *resource.ResourceType == "AWS::RDS::DBCluster" {
			if resource.PhysicalResourceId != nil && *resource.PhysicalResourceId != "" {
				results = append(results, *resource.PhysicalResourceId)
				fmt.Printf("🔍 検出されたAurora DBクラスター: %s\n", *resource.PhysicalResourceId)
			}
		}
	}

	if len(results) == 0 {
		return results, fmt.Errorf("指定されたスタック (%s) にAurora DBクラスターが見つかりませんでした", stackName)
	}

	return results, nil
}
