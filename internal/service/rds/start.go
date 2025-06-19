package rds

import (
	"context"
	"fmt"

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
