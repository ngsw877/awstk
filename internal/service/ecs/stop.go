package ecs

import (
	"fmt"
	
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

// StopEcsService はECSサービスを停止します
func StopEcsService(ecsClient *ecs.Client, aasClient *applicationautoscaling.Client, opts StopServiceOptions) error {
	// キャパシティ設定オプションを作成（停止のため0に設定）
	capacityOpts := ServiceCapacityOptions{
		ClusterName: opts.ClusterName,
		ServiceName: opts.ServiceName,
		MinCapacity: 0,
		MaxCapacity: 0,
	}

	// キャパシティを設定
	fmt.Println("🛑 サービスの停止を開始します...")
	err := SetEcsServiceCapacity(aasClient, capacityOpts)
	if err != nil {
		return fmt.Errorf("❌ エラー: %w", err)
	}

	// 停止完了を必ず待機
	waitOpts := waitOptions{
		ClusterName:        opts.ClusterName,
		ServiceName:        opts.ServiceName,
		TargetRunningCount: 0,
		TimeoutSeconds:     opts.TimeoutSeconds,
	}
	err = waitForServiceStatus(ecsClient, waitOpts)
	if err != nil {
		return fmt.Errorf("❌ サービス停止監視エラー: %w", err)
	}

	return nil
}
