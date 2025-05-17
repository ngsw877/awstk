package internal

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
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
