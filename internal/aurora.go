package internal

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
)

// StartAuroraCluster Auroraクラスタを起動する
func StartAuroraCluster(awsCtx AwsContext, clusterId string) error {
	cfg, err := LoadAwsConfig(awsCtx)
	if err != nil {
		return fmt.Errorf("AWS設定のロードに失敗: %w", err)
	}

	client := rds.NewFromConfig(cfg)
	_, err = client.StartDBCluster(context.Background(), &rds.StartDBClusterInput{
		DBClusterIdentifier: aws.String(clusterId),
	})
	if err != nil {
		return fmt.Errorf("❌ Aurora DBクラスターの起動に失敗: %w", err)
	}
	return nil
}

// StopAuroraCluster Auroraクラスタを停止する
func StopAuroraCluster(awsCtx AwsContext, clusterId string) error {
	cfg, err := LoadAwsConfig(awsCtx)
	if err != nil {
		return fmt.Errorf("AWS設定のロードに失敗: %w", err)
	}

	client := rds.NewFromConfig(cfg)
	_, err = client.StopDBCluster(context.Background(), &rds.StopDBClusterInput{
		DBClusterIdentifier: aws.String(clusterId),
	})
	if err != nil {
		return fmt.Errorf("❌ Aurora DBクラスターの停止に失敗: %w", err)
	}
	return nil
}
