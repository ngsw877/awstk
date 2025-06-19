package ecs

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

// describeService はECSサービスの詳細情報を取得します
func describeService(ecsClient *ecs.Client, clusterName, serviceName string) (*types.Service, error) {
	// サービスの詳細を取得
	resp, err := ecsClient.DescribeServices(context.Background(), &ecs.DescribeServicesInput{
		Cluster:  aws.String(clusterName),
		Services: []string{serviceName},
	})
	if err != nil {
		return nil, fmt.Errorf("サービス情報の取得に失敗しました: %w", err)
	}

	if len(resp.Services) == 0 {
		return nil, fmt.Errorf("サービス '%s' が見つかりません", serviceName)
	}

	return &resp.Services[0], nil
}

// SetEcsServiceCapacity はECSサービスの最小・最大キャパシティを設定します
func SetEcsServiceCapacity(autoScalingClient *applicationautoscaling.Client, opts ServiceCapacityOptions) error {
	fmt.Printf("🔍 🚀 Fargate (ECSサービス: %s) のDesiredCountを%d～%dに設定します...\n",
		opts.ServiceName, opts.MinCapacity, opts.MaxCapacity)

	// リソースIDを構築
	resourceId := fmt.Sprintf("service/%s/%s", opts.ClusterName, opts.ServiceName)

	// スケーラブルターゲットを登録
	_, err := autoScalingClient.RegisterScalableTarget(context.Background(), &applicationautoscaling.RegisterScalableTargetInput{
		ServiceNamespace:  "ecs",
		ScalableDimension: "ecs:service:DesiredCount",
		ResourceId:        &resourceId,
		MinCapacity:       aws.Int32(int32(opts.MinCapacity)),
		MaxCapacity:       aws.Int32(int32(opts.MaxCapacity)),
	})
	if err != nil {
		return fmt.Errorf("スケーラブルターゲット登録でエラー: %w", err)
	}

	// 設定完了メッセージを表示（サービスの状態の解釈はcmdパッケージに任せる）
	fmt.Printf("✅ Fargate (ECSサービス) のDesiredCountを%d～%dに設定しました。\n",
		opts.MinCapacity, opts.MaxCapacity)
	return nil
}

// waitForServiceStatus はECSサービスの状態が目標とする状態になるまで待機します
func waitForServiceStatus(ecsClient *ecs.Client, opts ServiceCapacityOptions, targetRunningCount int, timeoutSeconds int) error {
	var status string
	if targetRunningCount == 0 {
		status = "停止"
	} else {
		status = "起動"
	}
	fmt.Printf("⏳ サービスが%s状態になるまで待機しています...\n", status)

	start := time.Now()
	timeout := time.Duration(timeoutSeconds) * time.Second
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		// サービスの状態を取得
		service, err := describeService(ecsClient, opts.ClusterName, opts.ServiceName)
		if err != nil {
			return fmt.Errorf("サービス情報の取得に失敗しました: %w", err)
		}

		runningCount := int(service.RunningCount)
		desiredCount := int(service.DesiredCount)

		// 経過時間と進捗状況を表示
		elapsed := time.Since(start).Round(time.Second)
		fmt.Printf("⏱️ 経過時間: %s - 実行中タスク: %d / 希望タスク数: %d\n",
			elapsed, runningCount, desiredCount)

		// 目標達成の確認
		if runningCount == targetRunningCount && desiredCount == targetRunningCount {
			if targetRunningCount == 0 {
				fmt.Println("✅ サービスが完全に停止しました")
			} else {
				fmt.Println("✅ サービスが完全に起動しました")
			}
			return nil
		}

		// タイムアウトのチェック
		if time.Since(start) > timeout {
			return fmt.Errorf("タイムアウト: %d秒経過しましたがサービスは目標状態に達していません", timeoutSeconds)
		}
	}
}
