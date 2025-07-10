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

// EcsExecOptions はECS execute-commandのパラメータを格納する構造体
type EcsExecOptions struct {
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
