package service

import (
	"context"
	"fmt"

	"awstk/internal/aws"

	"github.com/aws/aws-sdk-go-v2/service/rds"
)

// StartRdsInstance RDSインスタンスを起動する
func StartRdsInstance(rdsClient *rds.Client, instanceId string) error {
	input := &rds.StartDBInstanceInput{
		DBInstanceIdentifier: &instanceId,
	}

	_, err := rdsClient.StartDBInstance(context.Background(), input)
	if err != nil {
		return fmt.Errorf("RDSインスタンス起動エラー: %w", err)
	}

	return nil
}

// StopRdsInstance RDSインスタンスを停止する
func StopRdsInstance(rdsClient *rds.Client, instanceId string) error {
	input := &rds.StopDBInstanceInput{
		DBInstanceIdentifier: &instanceId,
	}

	_, err := rdsClient.StopDBInstance(context.Background(), input)
	if err != nil {
		return fmt.Errorf("RDSインスタンス停止エラー: %w", err)
	}

	return nil
}

// GetRdsFromStack はCloudFormationスタックからRDSインスタンス識別子を取得します
func GetRdsFromStack(awsCtx aws.Context, stackName string) (string, error) {
	allInstances, err := GetAllRdsFromStack(awsCtx, stackName)
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
func GetAllRdsFromStack(awsCtx aws.Context, stackName string) ([]string, error) {
	// 共通関数を使用してスタックリソースを取得
	stackResources, err := getStackResources(awsCtx, stackName)
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
