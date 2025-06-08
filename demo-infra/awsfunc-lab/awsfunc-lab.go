package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type AwsfuncLabStackProps struct {
	awscdk.StackProps
	ResourceCounts *ResourceCounts
}

func NewAwsfuncLabStack(scope constructs.Construct, id string, props *AwsfuncLabStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// デフォルト値設定
	counts := DefaultResourceCounts()
	if props != nil && props.ResourceCounts != nil {
		counts = props.ResourceCounts
	}

	// Create a VPC with public and private subnets
	vpc := awsec2.NewVpc(stack, jsii.String("MainVPC"), &awsec2.VpcProps{
		MaxAzs:      jsii.Number(2),
		NatGateways: jsii.Number(1),
		SubnetConfiguration: &[]*awsec2.SubnetConfiguration{
			{
				Name:       jsii.String("PublicSubnet"),
				SubnetType: awsec2.SubnetType_PUBLIC,
				CidrMask:   jsii.Number(24),
			},
			{
				Name:       jsii.String("PrivateSubnet"),
				SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS,
				CidrMask:   jsii.Number(24),
			},
		},
	})

	// 統合リソースグループを作成
	NewMultiResourceGroup(stack, "MultiResourceGroup", &MultiResourceGroupProps{
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
	NewAwsfuncLabStack(app, "AwsfuncLab", &AwsfuncLabStackProps{
		StackProps:     awscdk.StackProps{Env: env()},
		ResourceCounts: DefaultResourceCounts(),
	})

	// // テスト用デプロイ（複数リソース）
	// NewAwsfuncLabStack(app, "AwsfuncLabTest", &AwsfuncLabStackProps{
	// 	StackProps:     awscdk.StackProps{Env: env()},
	// 	ResourceCounts: TestResourceCounts(),
	// })

	// // カスタム設定例
	// NewAwsfuncLabStack(app, "AwsfuncLabCustom", &AwsfuncLabStackProps{
	// 	StackProps: awscdk.StackProps{Env: env()},
	// 	ResourceCounts: &ResourceCounts{
	// 		EcsCount:    3,
	// 		Ec2Count:    2,
	// 		RdsCount:    1,
	// 		S3Count:     5,
	// 		AuroraCount: 1,
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
	//---------------------------------------------------------------------------
	return nil
}
