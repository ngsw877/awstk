package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func NewAwstkLabStack(scope constructs.Construct, id string, props *AwstkLabStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// ResourceCountsを取得（nil の場合はデフォルト値を使用）
	counts := props.ResourceCounts
	if counts == nil {
		counts = DefaultResourceCounts()
	}

	// VPCを作成
	vpc := awsec2.NewVpc(stack, jsii.String("Vpc"), &awsec2.VpcProps{
		MaxAzs: jsii.Number(2),
		SubnetConfiguration: &[]*awsec2.SubnetConfiguration{
			{
				Name:       jsii.String("public"),
				SubnetType: awsec2.SubnetType_PUBLIC,
				CidrMask:   jsii.Number(24),
			},
			{
				Name:       jsii.String("private"),
				SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS,
				CidrMask:   jsii.Number(24),
			},
		},
	})

	// MultiResourceGroupを作成
	NewMultiResourceGroup(stack, "Resources", &MultiResourceGroupProps{
		Vpc:         vpc,
		EcsCount:    counts.EcsCount,
		Ec2Count:    counts.Ec2Count,
		S3Count:     counts.S3Count,
		RdsCount:    counts.RdsCount,
		AuroraCount: counts.AuroraCount,
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	// 通常デプロイ（1つずつ）
	NewAwstkLabStack(app, "AwstkLab", &AwstkLabStackProps{
		StackProps:     awscdk.StackProps{Env: env()},
		ResourceCounts: DefaultResourceCounts(),
	})

	// // テスト用デプロイ（複数リソース）
	// NewAwstkLabStack(app, "AwstkLabTest", &AwstkLabStackProps{
	// 	StackProps:     awscdk.StackProps{Env: env()},
	// 	ResourceCounts: TestResourceCounts(),
	// })

	// // カスタム設定例
	// NewAwstkLabStack(app, "AwstkLabCustom", &AwstkLabStackProps{
	// 	StackProps: awscdk.StackProps{Env: env()},
	// 	ResourceCounts: &ResourceCounts{
	// 		S3Buckets:       10,
	// 		ECRRepositories: 5,
	// 		ECSClusters:     2,
	// 		AuroraClusters:  1,
	// 		RDSInstances:    3,
	// 		EC2Instances:    4,
	// 		ECSServices:     6,
	// 	},
	// })

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	// return nil

	// Uncomment if you know exactly what account and region you want to deploy
	// the stack to. This is the recommendation for production stacks.
	// return &awscdk.Environment{
	//  Account: jsii.String("123456789012"),
	//  Region:  jsii.String("us-east-1"),
	// }

	// Uncomment to specialize this stack for the AWS Account and Region that are
	// implied by the current CLI configuration. This is recommended for dev
	// stacks.
	// return &awscdk.Environment{
	//  Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
	//  Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	// }
	return nil
}
