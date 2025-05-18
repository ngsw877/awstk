package internal

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/rds"
)

// StartRdsInstance RDSインスタンスを起動する
func StartRdsInstance(instanceId, region, profile string) error {
	ctx := context.Background()
	cfg, err := LoadAwsConfig(region, profile)
	if err != nil {
		return fmt.Errorf("AWS設定のロードに失敗: %w", err)
	}

	client := rds.NewFromConfig(cfg)
	_, err = client.StartDBInstance(ctx, &rds.StartDBInstanceInput{
		DBInstanceIdentifier: aws.String(instanceId),
	})
	if err != nil {
		return fmt.Errorf("❌ RDSインスタンスの起動に失敗: %w", err)
	}
	return nil
}

// StopRdsInstance RDSインスタンスを停止する
func StopRdsInstance(instanceId, region, profile string) error {
	ctx := context.Background()
	cfg, err := LoadAwsConfig(region, profile)
	if err != nil {
		return fmt.Errorf("AWS設定のロードに失敗: %w", err)
	}

	client := rds.NewFromConfig(cfg)
	_, err = client.StopDBInstance(ctx, &rds.StopDBInstanceInput{
		DBInstanceIdentifier: aws.String(instanceId),
	})
	if err != nil {
		return fmt.Errorf("❌ RDSインスタンスの停止に失敗: %w", err)
	}
	return nil
}

// GetRdsFromStack はCloudFormationスタック名からRDSインスタンス識別子を取得します。
func GetRdsFromStack(stackName, region, profile string) (string, error) {
	ctx := context.Background()
	cfg, err := LoadAwsConfig(region, profile) // Assuming LoadAwsConfig is available
	if err != nil {
		return "", fmt.Errorf("AWS設定のロードに失敗: %w", err)
	}

	cfnClient := cloudformation.NewFromConfig(cfg)

	// DescribeStackResources でスタック内のリソース一覧を取得
	resp, err := cfnClient.DescribeStackResources(ctx, &cloudformation.DescribeStackResourcesInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return "", fmt.Errorf("CloudFormationスタックのリソース取得に失敗: %w", err)
	}

	// リソースの中からRDS DBInstanceを探す
	for _, resource := range resp.StackResources {
		if resource.ResourceType != nil && *resource.ResourceType == "AWS::RDS::DBInstance" {
			if resource.PhysicalResourceId != nil && *resource.PhysicalResourceId != "" {
				// 見つかった最初のRDSインスタンスのPhysicalResourceIdを返す
				return *resource.PhysicalResourceId, nil
			}
		}
	}

	// RDSインスタンスが見つからなかった場合
	return "", fmt.Errorf("指定されたスタック (%s) にRDSインスタンスが見つかりませんでした", stackName)
}
