package ecs

import (
	awsCtx "awstk/internal/aws"
	"awstk/internal/cli"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

// GetRunningTask 実行中のタスクを取得する
func GetRunningTask(ecsClient *ecs.Client, clusterName, serviceName string) (string, error) {
	fmt.Println("🔍 実行中のタスクを検索中...")

	// タスク一覧を取得
	taskList, err := ecsClient.ListTasks(context.Background(), &ecs.ListTasksInput{
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

// ExecuteEcsCommand はECS execute-commandを実行する
func ExecuteEcsCommand(awsCtx awsCtx.Context, opts ExecOptions) error {
	// aws ecs execute-commandコマンドを構築
	args := []string{
		"ecs", "execute-command",
		"--cluster", opts.ClusterName,
		"--task", opts.TaskId,
		"--container", opts.ContainerName,
		"--interactive",
		"--command", "/bin/bash",
	}

	// cli層の共通関数を使用してコマンドを実行
	return cli.ExecuteAwsCommand(awsCtx, args)
}
