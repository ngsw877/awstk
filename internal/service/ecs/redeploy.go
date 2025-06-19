package ecs

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

// ForceRedeployService はECSサービスを強制再デプロイします
func ForceRedeployService(ecsClient *ecs.Client, clusterName, serviceName string) error {
	fmt.Printf("🚀 ECSサービス '%s' を強制再デプロイします...\n", serviceName)

	updateInput := &ecs.UpdateServiceInput{
		Cluster:            aws.String(clusterName),
		Service:            aws.String(serviceName),
		ForceNewDeployment: true,
	}

	_, err := ecsClient.UpdateService(context.Background(), updateInput)

	if err != nil {
		return fmt.Errorf("サービスの強制再デプロイに失敗しました: %w", err)
	}

	fmt.Println("✅ 強制再デプロイを開始しました")
	return nil
}

// WaitForDeploymentComplete はECSサービスのデプロイが完了するまで待機します
func WaitForDeploymentComplete(ecsClient *ecs.Client, clusterName, serviceName string, timeoutSeconds int) error {
	fmt.Println("⏳ デプロイ完了を待機しています...")

	start := time.Now()
	timeout := time.Duration(timeoutSeconds) * time.Second
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C

		// サービスの詳細を取得
		resp, err := ecsClient.DescribeServices(context.Background(), &ecs.DescribeServicesInput{
			Cluster:  aws.String(clusterName),
			Services: []string{serviceName},
		})
		if err != nil {
			return fmt.Errorf("サービス情報の取得に失敗しました: %w", err)
		}

		if len(resp.Services) == 0 {
			return fmt.Errorf("サービス '%s' が見つかりません", serviceName)
		}

		service := resp.Services[0]

		// デプロイメント状況をチェック
		var primaryDeployment *types.Deployment
		for _, deployment := range service.Deployments {
			if *deployment.Status == "PRIMARY" {
				primaryDeployment = &deployment
				break
			}
		}

		if primaryDeployment == nil {
			return fmt.Errorf("プライマリデプロイメントが見つかりません")
		}

		runningCount := int(primaryDeployment.RunningCount)
		desiredCount := int(primaryDeployment.DesiredCount)
		deploymentStatus := *primaryDeployment.Status

		// 経過時間と進捗状況を表示
		elapsed := time.Since(start).Round(time.Second)
		fmt.Printf("⏱️ 経過時間: %s - デプロイ状況: %s - 実行中タスク: %d / 希望タスク数: %d\n",
			elapsed, deploymentStatus, runningCount, desiredCount)

		// デプロイ完了の確認
		if deploymentStatus == "PRIMARY" && runningCount == desiredCount && desiredCount > 0 {
			fmt.Println("✅ デプロイが完了しました")
			return nil
		}

		// タイムアウトのチェック
		if time.Since(start) > timeout {
			return fmt.Errorf("タイムアウト: %d秒経過しましたがデプロイは完了していません", timeoutSeconds)
		}
	}
}
