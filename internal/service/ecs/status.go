package ecs

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	autoscalingtypes "github.com/aws/aws-sdk-go-v2/service/applicationautoscaling/types"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
)


// GetServiceStatus はECSサービスの状態を取得する
func GetServiceStatus(ecsClient *ecs.Client, aasClient *applicationautoscaling.Client, opts StatusOptions) (*serviceStatus, error) {
	// サービス情報を取得
	serviceResp, err := ecsClient.DescribeServices(context.Background(), &ecs.DescribeServicesInput{
		Cluster:  &opts.ClusterName,
		Services: []string{opts.ServiceName},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe service: %w", err)
	}

	if len(serviceResp.Services) == 0 {
		return nil, fmt.Errorf("service '%s' not found in cluster '%s'", opts.ServiceName, opts.ClusterName)
	}

	service := serviceResp.Services[0]
	statusStr := ""
	if service.Status != nil {
		statusStr = *service.Status
	}
	
	taskDef := ""
	if service.TaskDefinition != nil {
		taskDef = *service.TaskDefinition
	}

	status := &serviceStatus{
		ServiceName:    opts.ServiceName,
		ClusterName:    opts.ClusterName,
		Status:         statusStr,
		TaskDefinition: taskDef,
		DesiredCount:   service.DesiredCount,
		RunningCount:   service.RunningCount,
		PendingCount:   service.PendingCount,
	}

	// タスク詳細を取得
	tasks, err := getTaskDetails(ecsClient, opts.ClusterName, opts.ServiceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get task details: %w", err)
	}
	status.Tasks = tasks

	// Auto Scaling設定を取得
	autoScaling, err := getAutoScalingInfo(aasClient, opts.ClusterName, opts.ServiceName)
	if err != nil {
		// Auto Scalingが設定されていない場合はエラーではない
		fmt.Printf("ℹ️  Auto Scaling情報の取得に失敗しました（設定されていない可能性があります）: %v\n", err)
	} else {
		status.AutoScaling = autoScaling
	}

	return status, nil
}

// getTaskDetails はサービスに関連するタスクの詳細を取得する
func getTaskDetails(ecsClient *ecs.Client, clusterName, serviceName string) ([]taskInfo, error) {
	// サービスのタスクARNを取得
	tasksResp, err := ecsClient.ListTasks(context.Background(), &ecs.ListTasksInput{
		Cluster:     &clusterName,
		ServiceName: &serviceName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	if len(tasksResp.TaskArns) == 0 {
		return []taskInfo{}, nil
	}

	// タスクの詳細情報を取得
	taskDetailsResp, err := ecsClient.DescribeTasks(context.Background(), &ecs.DescribeTasksInput{
		Cluster: &clusterName,
		Tasks:   tasksResp.TaskArns,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe tasks: %w", err)
	}

	var tasks []taskInfo
	for _, task := range taskDetailsResp.Tasks {
		taskId := extractTaskId(*task.TaskArn)
		healthStatus := "UNKNOWN"
		if task.HealthStatus != "" {
			healthStatus = string(task.HealthStatus)
		}

		createdAt := ""
		if task.CreatedAt != nil {
			createdAt = task.CreatedAt.Format("2006-01-02 15:04:05")
		}

		lastStatus := ""
		if task.LastStatus != nil {
			lastStatus = *task.LastStatus
		}

		tasks = append(tasks, taskInfo{
			TaskId:       taskId,
			Status:       lastStatus,
			HealthStatus: healthStatus,
			CreatedAt:    createdAt,
		})
	}

	return tasks, nil
}

// getAutoScalingInfo はAuto Scalingの設定情報を取得する
func getAutoScalingInfo(autoScalingClient *applicationautoscaling.Client, clusterName, serviceName string) (*autoScalingInfo, error) {
	resourceId := fmt.Sprintf("service/%s/%s", clusterName, serviceName)

	// Scalable Targetsを取得
	targetsResp, err := autoScalingClient.DescribeScalableTargets(context.Background(), &applicationautoscaling.DescribeScalableTargetsInput{
		ServiceNamespace: autoscalingtypes.ServiceNamespaceEcs,
		ResourceIds:      []string{resourceId},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe scalable targets: %w", err)
	}

	if len(targetsResp.ScalableTargets) == 0 {
		return nil, fmt.Errorf("no scalable targets found")
	}

	target := targetsResp.ScalableTargets[0]
	
	minCap := int32(0)
	if target.MinCapacity != nil {
		minCap = *target.MinCapacity
	}
	
	maxCap := int32(0)
	if target.MaxCapacity != nil {
		maxCap = *target.MaxCapacity
	}
	
	return &autoScalingInfo{
		MinCapacity: minCap,
		MaxCapacity: maxCap,
	}, nil
}

// extractTaskId はタスクARNからタスクIDを抽出する
func extractTaskId(taskArn string) string {
	// arn:aws:ecs:region:account:task/cluster-name/task-id の形式からtask-idを抽出
	parts := strings.Split(taskArn, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-1]
	}
	return taskArn
}

// ShowServiceStatus はECSサービスの状態を表示する
func ShowServiceStatus(status *serviceStatus) {
	fmt.Printf("🔍 ECSサービス状態: %s/%s\n\n", status.ClusterName, status.ServiceName)

	// サービス基本情報
	fmt.Printf("📊 サービス情報:\n")
	fmt.Printf("  状態:           %s\n", status.Status)
	fmt.Printf("  タスク定義:      %s\n", status.TaskDefinition)
	fmt.Printf("  期待数:         %d\n", status.DesiredCount)
	fmt.Printf("  実行中:         %d\n", status.RunningCount)
	fmt.Printf("  起動中:         %d\n", status.PendingCount)

	// Auto Scaling情報
	if status.AutoScaling != nil {
		fmt.Printf("\n⚖️  Auto Scaling設定:\n")
		fmt.Printf("  最小キャパシティ: %d\n", status.AutoScaling.MinCapacity)
		fmt.Printf("  最大キャパシティ: %d\n", status.AutoScaling.MaxCapacity)
	}

	// タスク詳細
	fmt.Printf("\n📋 タスク詳細:\n")
	if len(status.Tasks) == 0 {
		fmt.Println("  実行中のタスクはありません")
	} else {
		for i, task := range status.Tasks {
			fmt.Printf("  %d. タスクID: %s\n", i+1, task.TaskId)
			fmt.Printf("     状態:     %s\n", task.Status)
			fmt.Printf("     ヘルス:   %s\n", task.HealthStatus)
			if task.CreatedAt != "" {
				fmt.Printf("     作成日時: %s\n", task.CreatedAt)
			}
			fmt.Println()
		}
	}
}