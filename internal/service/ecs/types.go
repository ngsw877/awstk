package ecs

// ServiceCapacityOptions はECSサービスのキャパシティ設定用パラメータを格納する構造体
type ServiceCapacityOptions struct {
	ClusterName string
	ServiceName string
	MinCapacity int
	MaxCapacity int
}

// ExecOptions EcsExecOptions はECS execute-commandのパラメータを格納する構造体
type ExecOptions struct {
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
	TimeoutSeconds int
}

// StartServiceOptions はECSサービス起動のパラメータを格納する構造体
type StartServiceOptions struct {
	ClusterName    string // 必須: ECSクラスター名
	ServiceName    string // 必須: ECSサービス名
	MinCapacity    int    // 必須: 最小キャパシティ
	MaxCapacity    int    // 必須: 最大キャパシティ
	TimeoutSeconds int    // オプション: タイムアウト秒数（デフォルト: 300）
}

// StopServiceOptions はECSサービス停止のパラメータを格納する構造体
type StopServiceOptions struct {
	ClusterName    string // 必須: ECSクラスター名
	ServiceName    string // 必須: ECSサービス名
	TimeoutSeconds int    // オプション: タイムアウト秒数（デフォルト: 300）
}

// StatusOptions はECSサービス状態取得のパラメータを格納する構造体
type StatusOptions struct {
	ClusterName string // 必須: ECSクラスター名
	ServiceName string // 必須: ECSサービス名
}

// waitOptions はサービス状態待機のパラメータを格納する構造体（内部使用）
type waitOptions struct {
	ClusterName        string
	ServiceName        string
	TargetRunningCount int
	TimeoutSeconds     int
}

// waitTaskOptions はタスク停止待機のパラメータを格納する構造体（内部使用）
type waitTaskOptions struct {
	ClusterName    string
	TaskArn        string
	ContainerName  string
	TimeoutSeconds int
}

// WaitDeploymentOptions はデプロイ完了待機のパラメータを格納する構造体
type WaitDeploymentOptions struct {
	ClusterName    string // 必須: ECSクラスター名
	ServiceName    string // 必須: ECSサービス名
	TimeoutSeconds int    // 必須: タイムアウト秒数
}

// ResolveOptions はECSクラスター名とサービス名の解決オプション
type ResolveOptions struct {
	StackName   string // オプション: CloudFormationスタック名
	ClusterName string // オプション: ECSクラスター名（スタック名が指定されていない場合は必須）
	ServiceName string // オプション: ECSサービス名（スタック名が指定されていない場合は必須）
}

// serviceStatus はECSサービスの状態情報を格納する構造体
type serviceStatus struct {
	ServiceName    string
	ClusterName    string
	Status         string
	TaskDefinition string
	DesiredCount   int32
	RunningCount   int32
	PendingCount   int32
	Tasks          []taskInfo
	AutoScaling    *autoScalingInfo
}

// taskInfo はECSタスクの情報を格納する構造体
type taskInfo struct {
	TaskId       string
	Status       string
	HealthStatus string
	CreatedAt    string
}

// autoScalingInfo はAuto Scalingの設定情報を格納する構造体
type autoScalingInfo struct {
	MinCapacity int32
	MaxCapacity int32
}
