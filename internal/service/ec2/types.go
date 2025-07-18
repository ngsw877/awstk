package ec2

import (
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// ClientSet はEC2関連の操作に必要なクライアントをまとめた構造体
type ClientSet struct {
	Ec2Client *ec2.Client
	CfnClient *cloudformation.Client
}

// Instance Ec2Instance EC2インスタンスの情報を格納する構造体
type Instance struct {
	InstanceId   string
	InstanceName string
	State        string
}
