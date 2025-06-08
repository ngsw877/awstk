package internal

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
)

// StartRdsInstance RDSインスタンスを起動する
func StartRdsInstance(awsContext AwsContext, instanceId string) error {
	cfg, err := LoadAwsConfig(awsContext)
	if err != nil {
		return fmt.Errorf("AWS設定のロードに失敗: %w", err)
	}

	client := rds.NewFromConfig(cfg)
	_, err = client.StartDBInstance(context.Background(), &rds.StartDBInstanceInput{
		DBInstanceIdentifier: aws.String(instanceId),
	})
	if err != nil {
		return fmt.Errorf("❌ RDSインスタンスの起動に失敗: %w", err)
	}
	return nil
}

// StopRdsInstance RDSインスタンスを停止する
func StopRdsInstance(awsContext AwsContext, instanceId string) error {
	cfg, err := LoadAwsConfig(awsContext)
	if err != nil {
		return fmt.Errorf("AWS設定のロードに失敗: %w", err)
	}

	client := rds.NewFromConfig(cfg)
	_, err = client.StopDBInstance(context.Background(), &rds.StopDBInstanceInput{
		DBInstanceIdentifier: aws.String(instanceId),
	})
	if err != nil {
		return fmt.Errorf("❌ RDSインスタンスの停止に失敗: %w", err)
	}
	return nil
}

// GetRdsFromStack はCloudFormationスタック名からRDSインスタンス識別子を取得します。
func GetRdsFromStack(awsCtx AwsContext, stackName string) (string, error) {
	instances, err := GetAllRdsFromStack(awsCtx, stackName)
	if err != nil {
		return "", err
	}

	// 対話的選択機能を使用
	selectedIndex, err := SelectFromOptions("複数のRDSインスタンスが見つかりました", instances)
	if err != nil {
		return "", err
	}

	return instances[selectedIndex], nil
}

// GetAllRdsFromStack はスタック内のすべてのRDSインスタンス識別子を取得します
func GetAllRdsFromStack(awsCtx AwsContext, stackName string) ([]string, error) {
	var results []string

	stackResources, err := getStackResources(awsCtx, stackName)
	if err != nil {
		return results, fmt.Errorf("CloudFormationスタックのリソース取得に失敗: %w", err)
	}

	// リソースの中からRDS DBInstanceを探す
	for _, resource := range stackResources {
		if resource.ResourceType != nil && *resource.ResourceType == "AWS::RDS::DBInstance" {
			if resource.PhysicalResourceId != nil && *resource.PhysicalResourceId != "" {
				results = append(results, *resource.PhysicalResourceId)
				fmt.Printf("🔍 検出されたRDSインスタンス: %s\n", *resource.PhysicalResourceId)
			}
		}
	}

	if len(results) == 0 {
		return results, fmt.Errorf("指定されたスタック (%s) にRDSインスタンスが見つかりませんでした", stackName)
	}

	return results, nil
}
