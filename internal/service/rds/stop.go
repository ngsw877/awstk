package rds

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/rds"
)

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
