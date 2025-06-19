package ecs

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

// waitForTaskStopped はタスクが停止するまで待機し、コンテナの終了コードを返します
func waitForTaskStopped(ecsClient *ecs.Client, clusterName, taskArn, containerName string, timeoutSeconds int) (int, error) {
	fmt.Println("⏳ タスクの完了を待機中...")

	timeout := time.Duration(timeoutSeconds) * time.Second
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	startTime := time.Now()

	for {
		select {
		case <-ticker.C:
			// タスクの状態を確認
			resp, err := ecsClient.DescribeTasks(context.Background(), &ecs.DescribeTasksInput{
				Cluster: aws.String(clusterName),
				Tasks:   []string{taskArn},
			})
			if err != nil {
				return -1, fmt.Errorf("タスク情報の取得に失敗しました: %w", err)
			}

			if len(resp.Tasks) == 0 {
				return -1, fmt.Errorf("タスク '%s' が見つかりません", taskArn)
			}

			task := resp.Tasks[0]
			lastStatus := *task.LastStatus

			// 経過時間と状態を表示
			elapsed := time.Since(startTime).Round(time.Second)
			fmt.Printf("⏱️ 経過時間: %s - タスク状態: %s\n", elapsed, lastStatus)

			// タスクが停止した場合
			if lastStatus == "STOPPED" {
				// 指定したコンテナの終了コードを取得
				for _, container := range task.Containers {
					if *container.Name == containerName {
						if container.ExitCode == nil {
							return -1, fmt.Errorf("コンテナ '%s' の終了コードが取得できませんでした", containerName)
						}
						exitCode := int(*container.ExitCode)
						return exitCode, nil
					}
				}

				// 指定したコンテナが見つからない場合
				containerNames := []string{}
				for _, container := range task.Containers {
					containerNames = append(containerNames, *container.Name)
				}
				return -1, fmt.Errorf("コンテナ '%s' がタスク内に見つかりません。利用可能なコンテナ: %s",
					containerName, strings.Join(containerNames, ", "))
			}
		case <-time.After(timeout):
			return -1, fmt.Errorf("タイムアウト: %d秒経過しましたがタスクは停止していません", timeoutSeconds)
		}
	}
}

// RunAndWaitForTask はECSタスクを実行し、完了するまで待機します
func RunAndWaitForTask(ecsClient *ecs.Client, opts RunAndWaitForTaskOptions) (int, error) {
	// タスク定義とネットワーク設定を決定
	var taskDefArn string
	var networkConfig *types.NetworkConfiguration

	if opts.TaskDefinition != "" {
		// タスク定義が直接指定された場合はそれを使用
		taskDefArn = opts.TaskDefinition
		fmt.Println("🔍 指定されたタスク定義を使用します: " + taskDefArn)
	} else {
		// サービスからタスク定義を取得
		fmt.Println("🔍 サービスの情報を取得中...")
		service, err := describeService(ecsClient, opts.ClusterName, opts.ServiceName)
		if err != nil {
			return -1, err
		}

		taskDefArn = *service.TaskDefinition
		networkConfig = service.NetworkConfiguration
		fmt.Println("🔍 サービスのタスク定義を使用します: " + taskDefArn)
	}

	// コマンドをオーバーライド
	var overrides *types.TaskOverride
	if opts.Command != "" {
		// コマンド内の引用符をエスケープ
		escapedCommand := strings.ReplaceAll(opts.Command, "\"", "\\\"")

		containerOverrides := []types.ContainerOverride{
			{
				Name:    aws.String(opts.ContainerName),
				Command: []string{"sh", "-c", escapedCommand},
			},
		}

		overrides = &types.TaskOverride{
			ContainerOverrides: containerOverrides,
		}

		fmt.Printf("🔍 コンテナ '%s' で実行するコマンド: %s\n", opts.ContainerName, opts.Command)
	}

	// タスク実行パラメータを設定
	runTaskInput := &ecs.RunTaskInput{
		Cluster:        aws.String(opts.ClusterName),
		TaskDefinition: aws.String(taskDefArn),
		LaunchType:     types.LaunchTypeFargate,
	}

	// オーバーライドがある場合は設定
	if overrides != nil {
		runTaskInput.Overrides = overrides
	}

	// ネットワーク設定がある場合は設定
	if networkConfig != nil {
		runTaskInput.NetworkConfiguration = networkConfig
	}

	// タスクを実行
	fmt.Println("🚀 タスクを実行中...")
	runResult, err := ecsClient.RunTask(context.Background(), runTaskInput)
	if err != nil {
		return -1, fmt.Errorf("タスクの実行に失敗しました: %w", err)
	}

	if len(runResult.Tasks) == 0 {
		return -1, errors.New("タスクの実行に失敗しました: タスクが作成されませんでした")
	}

	taskArn := *runResult.Tasks[0].TaskArn
	fmt.Println("✅ タスクが開始されました: " + taskArn)

	// タスクが停止するまで待機
	exitCode, err := waitForTaskStopped(ecsClient, opts.ClusterName, taskArn, opts.ContainerName, opts.TimeoutSeconds)
	if err != nil {
		return -1, err
	}

	return exitCode, nil
}
