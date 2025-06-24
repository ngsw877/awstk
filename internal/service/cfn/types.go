package cfn

import (
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/rds"
)

// StackStartStopOptions はCloudFormationスタック内リソースの起動/停止処理のパラメータを格納する構造体
type StackStartStopOptions struct {
	CfnClient                    *cloudformation.Client
	Ec2Client                    *ec2.Client
	RdsClient                    *rds.Client
	ApplicationAutoScalingClient *applicationautoscaling.Client
	StackName                    string
}

// StackResources はCloudFormationスタック内のリソース識別子を格納する構造体
type StackResources struct {
	Ec2InstanceIds   []string
	RdsInstanceIds   []string
	AuroraClusterIds []string
	EcsServiceInfo   []EcsServiceInfo
}

// CfnStack はCloudFormationスタックの名前とステータスを表す構造体
type CfnStack struct {
	Name   string
	Status string
}
