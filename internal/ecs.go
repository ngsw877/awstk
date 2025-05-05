package internal

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

// ServiceCapacityOptions はECSサービスキャパシティ設定のパラメータを格納する構造体
type ServiceCapacityOptions struct {
	ClusterName string
	ServiceName string
	Region      string
	Profile     string
	MinCapacity int
	MaxCapacity int
}

type EcsServiceInfo struct {
	ClusterName string
	ServiceName string
}

func GetEcsFromStack(stackName, region, profile string) (EcsServiceInfo, error) {
	var result EcsServiceInfo

	cfg, err := LoadAwsConfig(region, profile)
	if err != nil {
		return result, fmt.Errorf("AWS設定の読み込みエラー: %w", err)
	}

	// CloudFormationクライアントを作成
	cfnClient := cloudformation.NewFromConfig(cfg)

	// スタックからクラスターリソースを取得
	fmt.Println("🔍 スタック '" + stackName + "' からECSクラスターを検索中...")
	clusterResources, err := cfnClient.DescribeStackResources(context.TODO(), &cloudformation.DescribeStackResourcesInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return result, fmt.Errorf("スタックリソース取得エラー: %w", err)
	}

	// クラスターリソースをフィルタリング
	var clusterPhysicalIds []string
	for _, resource := range clusterResources.StackResources {
		if *resource.ResourceType == "AWS::ECS::Cluster" {
			clusterPhysicalIds = append(clusterPhysicalIds, *resource.PhysicalResourceId)
		}
	}

	if len(clusterPhysicalIds) == 0 {
		return result, errors.New("スタック '" + stackName + "' からECSクラスターを検出できませんでした")
	}

	// 複数のクラスターがある場合は警告を表示
	if len(clusterPhysicalIds) > 1 {
		fmt.Println("⚠️ 警告: スタック '" + stackName + "' に複数のECSクラスターが見つかりました。最初のクラスターを使用します:")
		for i, id := range clusterPhysicalIds {
			if i == 0 {
				fmt.Println(" * " + id + " (使用するクラスター)")
			} else {
				fmt.Println(" * " + id)
			}
		}
	}

	// 最初のクラスターを使用
	result.ClusterName = clusterPhysicalIds[0]

	// サービスリソースをフィルタリング
	fmt.Println("🔍 スタック '" + stackName + "' からECSサービスを検索中...")
	var servicePhysicalIds []string
	for _, resource := range clusterResources.StackResources {
		if *resource.ResourceType == "AWS::ECS::Service" {
			servicePhysicalIds = append(servicePhysicalIds, *resource.PhysicalResourceId)
		}
	}

	if len(servicePhysicalIds) == 0 {
		return result, errors.New("スタック '" + stackName + "' からECSサービスを検出できませんでした")
	}

	// サービス名を抽出 (形式: arn:aws:ecs:REGION:ACCOUNT:service/CLUSTER/SERVICE_NAME)
	serviceName := servicePhysicalIds[0]
	parts := strings.Split(serviceName, "/")
	if len(parts) > 0 {
		result.ServiceName = parts[len(parts)-1]
	} else {
		result.ServiceName = serviceName
	}

	// 複数のサービスがある場合は警告を表示
	if len(servicePhysicalIds) > 1 {
		fmt.Println("⚠️ 警告: スタック '" + stackName + "' に複数のECSサービスが見つかりました。最初のサービスを使用します:")
		for i, id := range servicePhysicalIds {
			serviceName := id
			parts := strings.Split(serviceName, "/")
			if len(parts) > 0 {
				serviceName = parts[len(parts)-1]
			}

			if i == 0 {
				fmt.Println(" * " + serviceName + " (使用するサービス)")
			} else {
				fmt.Println(" * " + serviceName)
			}
		}
	}

	return result, nil
}

func GetRunningTask(clusterName, serviceName, region, profile string) (string, error) {
	fmt.Println("🔍 実行中のタスクを検索中...")

	cfg, err := LoadAwsConfig(region, profile)
	if err != nil {
		return "", fmt.Errorf("AWS設定の読み込みエラー: %w", err)
	}

	// ECSクライアントを作成
	ecsClient := ecs.NewFromConfig(cfg)

	// タスク一覧を取得
	taskList, err := ecsClient.ListTasks(context.TODO(), &ecs.ListTasksInput{
		Cluster:     aws.String(clusterName),
		ServiceName: aws.String(serviceName),
	})
	if err != nil {
		return "", fmt.Errorf("タスク一覧取得エラー: %w", err)
	}

	if len(taskList.TaskArns) == 0 {
		return "", fmt.Errorf("クラスター '%s' のサービス '%s' で実行中のタスクが見つかりませんでした", clusterName, serviceName)
	}

	// 最初のタスクを使用
	taskId := taskList.TaskArns[0]
	fmt.Println("✅ 実行中のタスク '" + taskId + "' を検出しました")

	return taskId, nil
}

func ExecuteCommand(clusterName, taskId, containerName, region, profile string) error {
	// aws ecs execute-commandコマンドを構築
	args := []string{
		"ecs", "execute-command",
		"--region", region,
		"--cluster", clusterName,
		"--task", taskId,
		"--container", containerName,
		"--interactive",
		"--command", "/bin/bash",
	}

	if profile != "" {
		args = append(args, "--profile", profile)
	}

	// コマンドを実行
	cmd := exec.Command("aws", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// SetEcsServiceCapacity はECSサービスの最小・最大キャパシティを設定します
func SetEcsServiceCapacity(opts ServiceCapacityOptions) error {
	fmt.Printf("🔍 🚀 Fargate (ECSサービス: %s) のDesiredCountを%d～%dに設定します...\n",
		opts.ServiceName, opts.MinCapacity, opts.MaxCapacity)

	cfg, err := LoadAwsConfig(opts.Region, opts.Profile)
	if err != nil {
		return fmt.Errorf("AWS設定の読み込みエラー: %w", err)
	}

	// Application Auto Scalingクライアントを作成
	client := applicationautoscaling.NewFromConfig(cfg)

	// リソースIDを構築
	resourceId := fmt.Sprintf("service/%s/%s", opts.ClusterName, opts.ServiceName)

	// スケーラブルターゲットを登録
	_, err = client.RegisterScalableTarget(context.TODO(), &applicationautoscaling.RegisterScalableTargetInput{
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
