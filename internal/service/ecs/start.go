package ecs

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

// StartEcsService はECSサービスを起動します
func StartEcsService(autoScalingClient *applicationautoscaling.Client, ecsClient *ecs.Client, clusterName, serviceName string, minCapacity, maxCapacity, timeoutSeconds int) error {
	opts := ServiceCapacityOptions{
		ClusterName: clusterName,
		ServiceName: serviceName,
		MinCapacity: minCapacity,
		MaxCapacity: maxCapacity,
	}

	fmt.Println("🚀 サービスの起動を開始します...")
	err := SetEcsServiceCapacity(autoScalingClient, opts)
	if err != nil {
		return fmt.Errorf("❌ エラー: %w", err)
	}

	err = waitForServiceStatus(ecsClient, opts, minCapacity, timeoutSeconds)
	if err != nil {
		return fmt.Errorf("❌ サービス起動監視エラー: %w", err)
	}

	return nil
}
