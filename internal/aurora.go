package internal

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
)

// StartAuroraCluster Auroraクラスタを起動する
func StartAuroraCluster(clusterId, region, profile string) error {
	ctx := context.Background()
	cfg, err := LoadAwsConfig(region, profile)
	if err != nil {
		return fmt.Errorf("AWS設定のロードに失敗: %w", err)
	}

	client := rds.NewFromConfig(cfg)
	_, err = client.StartDBCluster(ctx, &rds.StartDBClusterInput{
		DBClusterIdentifier: aws.String(clusterId),
	})
	if err != nil {
		return fmt.Errorf("❌ Aurora DBクラスターの起動に失敗: %w", err)
	}
	return nil
}

// StopAuroraCluster Auroraクラスタを停止する
func StopAuroraCluster(clusterId, region, profile string) error {
	ctx := context.Background()
	cfg, err := LoadAwsConfig(region, profile)
	if err != nil {
		return fmt.Errorf("AWS設定のロードに失敗: %w", err)
	}

	client := rds.NewFromConfig(cfg)
	_, err = client.StopDBCluster(ctx, &rds.StopDBClusterInput{
		DBClusterIdentifier: aws.String(clusterId),
	})
	if err != nil {
		return fmt.Errorf("❌ Aurora DBクラスターの停止に失敗: %w", err)
	}
	return nil
}
