package cfn

// StackResources はCloudFormationスタック内のリソース識別子を格納する構造体
type StackResources struct {
	Ec2InstanceIds   []string
	RdsInstanceIds   []string
	AuroraClusterIds []string
	EcsServiceInfo   []EcsServiceInfo
}

// Stack CfnStack はCloudFormationスタックの名前とステータスを表す構造体
type Stack struct {
	Name   string
	Status string
}
