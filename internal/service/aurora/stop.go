package aurora

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/rds"
)

// StopAuroraCluster Auroraクラスターを停止する
func StopAuroraCluster(rdsClient *rds.Client, clusterId string) error {
	input := &rds.StopDBClusterInput{
		DBClusterIdentifier: &clusterId,
	}

	_, err := rdsClient.StopDBCluster(context.Background(), input)
	if err != nil {
		return fmt.Errorf("Auroraクラスター停止エラー: %w", err)
	}

	return nil
}
