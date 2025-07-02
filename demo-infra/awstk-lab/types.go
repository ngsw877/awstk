package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecspatterns"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
)

// AwstkLabStackProps はAwstkLabStackのプロパティ
type AwstkLabStackProps struct {
	awscdk.StackProps
	ResourceCounts *ResourceCounts
}

// ResourceCounts リソースの作成数を制御する構造体
type ResourceCounts struct {
	EcsCount    int
	Ec2Count    int
	S3Count     int
	RdsCount    int
	AuroraCount int
}

// MultiResourceGroup 複数のAWSリソースを管理するコンストラクト
type MultiResourceGroup struct {
	constructs.Construct
	EcsServices    []awsecspatterns.ApplicationLoadBalancedFargateService
	Ec2Instances   []awsec2.BastionHostLinux
	S3Buckets      []awss3.Bucket
	RdsInstances   []awsrds.DatabaseInstance
	AuroraClusters []awsrds.DatabaseCluster
}

// MultiResourceGroupProps コンストラクトのプロパティ
type MultiResourceGroupProps struct {
	Vpc         awsec2.Vpc
	EcsCount    int
	Ec2Count    int
	S3Count     int
	RdsCount    int
	AuroraCount int
}