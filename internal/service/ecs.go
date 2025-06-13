package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	awsinternal "awstk/internal/aws"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

// ServiceCapacityOptions はECSサービスキャパシティ設定のパラメータを格納する構造体
type ServiceCapacityOptions struct {
	ClusterName string
	ServiceName string
	MinCapacity int
	MaxCapacity int
}

type EcsServiceInfo struct {
	ClusterName string
	ServiceName string
}

// GetEcsFromStack はCloudFormationスタックからECSサービス情報を取得します
func GetEcsFromStack(awsCtx awsinternal.AwsContext, stackName string) (EcsServiceInfo, error) {
	allServices, err := getAllEcsFromStack(awsCtx, stackName)
	if err != nil {
		return EcsServiceInfo{}, err
	}

	if len(allServices) == 0 {
		return EcsServiceInfo{}, fmt.Errorf("スタック '%s' にECSサービスが見つかりませんでした", stackName)
	}

	// 複数のサービスがある場合は最初の要素を返す
	return allServices[0], nil
}

// GetAllEcsFromStack はスタック内のすべてのECSサービス情報を取得します
func getAllEcsFromStack(awsCtx awsinternal.AwsContext, stackName string) ([]EcsServiceInfo, error) {
	var results []EcsServiceInfo

	stackResources, err := getStackResources(awsCtx, stackName)
	if err != nil {
		return results, fmt.Errorf("CloudFormationスタックのリソース取得に失敗: %w", err)
	}

	// クラスターリソースをフィルタリング
	var clusterPhysicalIds []string
	for _, resource := range stackResources {
		if *resource.ResourceType == "AWS::ECS::Cluster" {
			clusterPhysicalIds = append(clusterPhysicalIds, *resource.PhysicalResourceId)
			fmt.Printf("🔍 検出されたECSクラスター: %s\n", *resource.PhysicalResourceId)
		}
	}

	if len(clusterPhysicalIds) == 0 {
		return results, errors.New("スタック '" + stackName + "' からECSクラスターを検出できませんでした")
	}

	// サービスリソースをフィルタリング
	fmt.Println("🔍 スタック '" + stackName + "' からECSサービスを検索中...")
	var servicePhysicalIds []string
	for _, resource := range stackResources {
		if *resource.ResourceType == "AWS::ECS::Service" {
			servicePhysicalIds = append(servicePhysicalIds, *resource.PhysicalResourceId)
		}
	}

	if len(servicePhysicalIds) == 0 {
		return results, errors.New("スタック '" + stackName + "' からECSサービスを検出できませんでした")
	}

	// 各サービスについてクラスターとの組み合わせを作成
	for _, serviceArn := range servicePhysicalIds {
		// サービス名を抽出 (形式: arn:aws:ecs:REGION:ACCOUNT:service/CLUSTER/SERVICE_NAME)
		parts := strings.Split(serviceArn, "/")
		if len(parts) < 2 {
			continue // 不正な形式はスキップ
		}

		clusterNameFromArn := parts[len(parts)-2]
		serviceName := parts[len(parts)-1]

		// ARNから抽出したクラスター名がスタック内のクラスターと一致するかチェック
		var matchedClusterName string
		for _, clusterId := range clusterPhysicalIds {
			// クラスター名の完全一致またはクラスターARNの末尾一致をチェック
			if clusterId == clusterNameFromArn || strings.HasSuffix(clusterId, "/"+clusterNameFromArn) {
				matchedClusterName = clusterId
				break
			}
		}

		// マッチしたクラスターがある場合のみ追加
		if matchedClusterName != "" {
			// クラスター名を正規化（ARNの場合は名前部分のみ抽出）
			displayClusterName := matchedClusterName
			if strings.Contains(matchedClusterName, "/") {
				clusterParts := strings.Split(matchedClusterName, "/")
				displayClusterName = clusterParts[len(clusterParts)-1]
			}

			results = append(results, EcsServiceInfo{
				ClusterName: displayClusterName,
				ServiceName: serviceName,
			})
			fmt.Printf("🔍 検出されたECSサービス: %s/%s\n", displayClusterName, serviceName)
		} else {
			fmt.Printf("⚠️ 警告: サービス %s のクラスター %s がスタック内で見つかりませんでした\n", serviceName, clusterNameFromArn)
		}
	}

	if len(results) == 0 {
		return results, errors.New("スタック '" + stackName + "' から有効なECSサービスを検出できませんでした")
	}

	return results, nil
}

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

func ExecuteCommand(awsCtx awsinternal.AwsContext, clusterName, taskId, containerName string) error {
	// aws ecs execute-commandコマンドを構築
	args := []string{
		"ecs", "execute-command",
		"--region", awsCtx.Region,
		"--cluster", clusterName,
		"--task", taskId,
		"--container", containerName,
		"--interactive",
		"--command", "/bin/bash",
	}

	if awsCtx.Profile != "" {
		args = append(args, "--profile", awsCtx.Profile)
	}

	// コマンドを実行
	cmd := exec.Command("aws", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
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

// WaitForServiceStatus はECSサービスの状態が目標とする状態になるまで待機します
func WaitForServiceStatus(ecsClient *ecs.Client, opts ServiceCapacityOptions, targetRunningCount int, timeoutSeconds int) error {
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

// RunAndWaitForTaskOptions はECSタスク実行のパラメータを格納する構造体
type RunAndWaitForTaskOptions struct {
	ClusterName    string
	ServiceName    string
	TaskDefinition string
	ContainerName  string
	Command        string
	Region         string
	Profile        string
	TimeoutSeconds int
}

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
