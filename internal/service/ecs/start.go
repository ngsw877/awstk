package ecs

import (
	"fmt"
	
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

// StartEcsService はECSサービスを起動します
func StartEcsService(ecsClient *ecs.Client, aasClient *applicationautoscaling.Client, opts StartServiceOptions) error {
	capacityOpts := ServiceCapacityOptions{
		ClusterName: opts.ClusterName,
		ServiceName: opts.ServiceName,
		MinCapacity: opts.MinCapacity,
		MaxCapacity: opts.MaxCapacity,
	}

	fmt.Println("🚀 サービスの起動を開始します...")
	err := SetEcsServiceCapacity(aasClient, capacityOpts)
	if err != nil {
		return fmt.Errorf("❌ エラー: %w", err)
	}

	waitOpts := waitOptions{
		ClusterName:        opts.ClusterName,
		ServiceName:        opts.ServiceName,
		TargetRunningCount: opts.MinCapacity,
		TimeoutSeconds:     opts.TimeoutSeconds,
	}
	err = waitForServiceStatus(ecsClient, waitOpts)
	if err != nil {
		return fmt.Errorf("❌ サービス起動監視エラー: %w", err)
	}

	return nil
}
