package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecspatterns"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type AwsfuncLabStackProps struct {
	awscdk.StackProps
}

func NewAwsfuncLabStack(scope constructs.Construct, id string, props *AwsfuncLabStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

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

	// Create a bastion host in the private subnet with SSM access
	userData := awsec2.UserData_ForLinux(nil)
	userData.AddCommands(
		jsii.String("sudo dnf update -y"),
		jsii.String("sudo dnf install -y postgresql16"),
	)

	bastionHost := awsec2.NewBastionHostLinux(stack, jsii.String("BastionHost"), &awsec2.BastionHostLinuxProps{
		Vpc:          vpc,
		InstanceType: awsec2.InstanceType_Of(awsec2.InstanceClass_BURSTABLE4_GRAVITON, awsec2.InstanceSize_MICRO),
		SubnetSelection: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS,
		},
		MachineImage: awsec2.NewAmazonLinuxImage(&awsec2.AmazonLinuxImageProps{
			Generation: awsec2.AmazonLinuxGeneration_AMAZON_LINUX_2023,
			CpuType:    awsec2.AmazonLinuxCpuType_ARM_64,
			UserData:   userData,
		}),
	})

	// Create an RDS instance in the private subnet
	dbInstance := awsrds.NewDatabaseInstance(stack, jsii.String("RDSInstance"), &awsrds.DatabaseInstanceProps{
		Engine: awsrds.DatabaseInstanceEngine_Postgres(&awsrds.PostgresInstanceEngineProps{
			Version: awsrds.PostgresEngineVersion_VER_16_4(),
		}),
		InstanceType: awsec2.InstanceType_Of(awsec2.InstanceClass_T4G, awsec2.InstanceSize_MICRO),
		Vpc:          vpc,
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS,
		},
		DatabaseName:           jsii.String("demo_db"),
		AllocatedStorage:       jsii.Number(20),
		StorageType:            awsrds.StorageType_GP3,
		RemovalPolicy:          awscdk.RemovalPolicy_DESTROY,
		DeletionProtection:     jsii.Bool(false),
		DeleteAutomatedBackups: jsii.Bool(true),
		StorageEncrypted:       jsii.Bool(true),
	})

	// Allow connections from the bastion host to the RDS instance
	dbInstance.Connections().AllowFrom(
		bastionHost,
		awsec2.Port_Tcp(jsii.Number(5432)),
		jsii.String("Allow access from Bastion host"),
	)

	// Create an S3 bucket
	awss3.NewBucket(stack, jsii.String("DataBucket"), &awss3.BucketProps{
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
		AutoDeleteObjects: jsii.Bool(true),
	})

	// Create an ALB in the public subnet with a Fargate service in the private subnet
	fargateService := awsecspatterns.NewApplicationLoadBalancedFargateService(stack, jsii.String("FargateService"), &awsecspatterns.ApplicationLoadBalancedFargateServiceProps{
		Vpc:            vpc,
		MemoryLimitMiB: jsii.Number(512),
		TaskImageOptions: &awsecspatterns.ApplicationLoadBalancedTaskImageOptions{
			Image:         awsecs.ContainerImage_FromRegistry(jsii.String("nginx:latest"), nil),
			ContainerPort: jsii.Number(80),
			ContainerName: jsii.String("web"),
		},
		PublicLoadBalancer: jsii.Bool(true),
		AssignPublicIp:     jsii.Bool(false),
		TaskSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS,
		},
		MinHealthyPercent: jsii.Number(100),
		RuntimePlatform: &awsecs.RuntimePlatform{
			CpuArchitecture:       awsecs.CpuArchitecture_ARM64(),
			OperatingSystemFamily: awsecs.OperatingSystemFamily_LINUX(),
		},
		EnableExecuteCommand: jsii.Bool(true),
	})

	// Output the ALB DNS name
	awscdk.NewCfnOutput(stack, jsii.String("LoadBalancerDNS"), &awscdk.CfnOutputProps{
		Value:       fargateService.LoadBalancer().LoadBalancerDnsName(),
		Description: jsii.String("The DNS name of the load balancer"),
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewAwsfuncLabStack(app, "AwsfuncLab", &AwsfuncLabStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

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

	// Uncomment if you know exactly what account and region you want to deploy
	// the stack to. This is the recommendation for production stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String("123456789012"),
	//  Region:  jsii.String("us-east-1"),
	// }

	// Uncomment to specialize this stack for the AWS Account and Region that are
	// implied by the current CLI configuration. This is recommended for dev
	// stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
	//  Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	// }
}
