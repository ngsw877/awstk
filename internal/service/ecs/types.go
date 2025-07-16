package ecs

import (
	"awstk/internal/aws"
)

// ServiceCapacityOptions はECSサービスのキャパシティ設定用パラメータを格納する構造体
type ServiceCapacityOptions struct {
	ClusterName string
	ServiceName string
	MinCapacity int
	MaxCapacity int
}

// ExecOptions EcsExecOptions はECS execute-commandのパラメータを格納する構造体
type ExecOptions struct {
	AwsCtx        aws.Context
	ClusterName   string
	TaskId        string
	ContainerName string
}

// RunAndWaitForTaskOptions はECSタスク実行のパラメータを格納する構造体
type RunAndWaitForTaskOptions struct {
	ClusterName    string
	ServiceName    string
	TaskDefinition string
	ContainerName  string
	Command        string
	AwsCtx         aws.Context
	TimeoutSeconds int
}

// serviceStatus はECSサービスの状態情報を格納する構造体
type serviceStatus struct {
	ServiceName     string
	ClusterName     string
	Status          string
	TaskDefinition  string
	DesiredCount    int32
	RunningCount    int32
	PendingCount    int32
	Tasks           []taskInfo
	AutoScaling     *autoScalingInfo
}

// taskInfo はECSタスクの情報を格納する構造体
type taskInfo struct {
	TaskId        string
	Status        string
	HealthStatus  string
	CreatedAt     string
}

// autoScalingInfo はAuto Scalingの設定情報を格納する構造体
type autoScalingInfo struct {
	MinCapacity int32
	MaxCapacity int32
}
