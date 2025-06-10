package main

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecspatterns"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// ResourceCounts リソースの作成数を制御する構造体
type ResourceCounts struct {
	EcsCount    int
	Ec2Count    int
	S3Count     int
	RdsCount    int
	AuroraCount int
}

// DefaultResourceCounts デフォルト設定（1つずつ）
func DefaultResourceCounts() *ResourceCounts {
	return &ResourceCounts{
		EcsCount:    1,
		Ec2Count:    1,
		S3Count:     1,
		RdsCount:    1,
		AuroraCount: 0,
	}
}

// TestResourceCounts テスト用設定（複数リソース）
func TestResourceCounts() *ResourceCounts {
	return &ResourceCounts{
		EcsCount:    2,
		Ec2Count:    2,
		S3Count:     2,
		RdsCount:    2,
		AuroraCount: 2,
	}
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

// NewMultiResourceGroup 新しいMultiResourceGroupを作成
func NewMultiResourceGroup(scope constructs.Construct, id string, props *MultiResourceGroupProps) *MultiResourceGroup {
	construct := constructs.NewConstruct(scope, &id)

	group := &MultiResourceGroup{
		Construct: construct,
	}

	// ECSサービスを作成
	group.createEcsServices(construct, props)

	// EC2インスタンスを作成
	group.createEc2Instances(construct, props)

	// S3バケットを作成
	group.createS3Buckets(construct, props)

	// RDSインスタンスを作成
	group.createRdsInstances(construct, props)

	// Auroraクラスターを作成
	group.createAuroraClusters(construct, props)

	return group
}

// createEcsServices ECSサービスを作成
func (g *MultiResourceGroup) createEcsServices(scope constructs.Construct, props *MultiResourceGroupProps) {
	for i := 0; i < props.EcsCount; i++ {
		serviceName := fmt.Sprintf("FargateService%d", i+1)
		containerName := "web"

		service := awsecspatterns.NewApplicationLoadBalancedFargateService(
			scope,
			jsii.String(serviceName),
			&awsecspatterns.ApplicationLoadBalancedFargateServiceProps{
				Vpc:            props.Vpc,
				MemoryLimitMiB: jsii.Number(512),
				TaskImageOptions: &awsecspatterns.ApplicationLoadBalancedTaskImageOptions{
					Image:         awsecs.ContainerImage_FromRegistry(jsii.String("nginx:latest"), nil),
					ContainerPort: jsii.Number(80),
					ContainerName: jsii.String(containerName),
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
			},
		)
		g.EcsServices = append(g.EcsServices, service)

		// ALB DNS名を出力
		awscdk.NewCfnOutput(scope, jsii.String(fmt.Sprintf("LoadBalancerDNS%d", i+1)), &awscdk.CfnOutputProps{
			Value:       service.LoadBalancer().LoadBalancerDnsName(),
			Description: jsii.String(fmt.Sprintf("The DNS name of load balancer %d", i+1)),
		})
	}
}

// createRdsInstances RDSインスタンスを作成
func (g *MultiResourceGroup) createRdsInstances(scope constructs.Construct, props *MultiResourceGroupProps) {
	for i := 0; i < props.RdsCount; i++ {
		instanceName := fmt.Sprintf("RDSInstance%d", i+1)
		dbName := fmt.Sprintf("demo_db_%d", i+1)

		instance := awsrds.NewDatabaseInstance(
			scope,
			jsii.String(instanceName),
			&awsrds.DatabaseInstanceProps{
				Engine: awsrds.DatabaseInstanceEngine_Postgres(&awsrds.PostgresInstanceEngineProps{
					Version: awsrds.PostgresEngineVersion_VER_16_4(),
				}),
				InstanceType: awsec2.InstanceType_Of(awsec2.InstanceClass_T4G, awsec2.InstanceSize_MICRO),
				Vpc:          props.Vpc,
				VpcSubnets: &awsec2.SubnetSelection{
					SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS,
				},
				DatabaseName:           jsii.String(dbName),
				AllocatedStorage:       jsii.Number(20),
				StorageType:            awsrds.StorageType_GP3,
				RemovalPolicy:          awscdk.RemovalPolicy_DESTROY,
				DeletionProtection:     jsii.Bool(false),
				DeleteAutomatedBackups: jsii.Bool(true),
				StorageEncrypted:       jsii.Bool(true),
			},
		)
		g.RdsInstances = append(g.RdsInstances, instance)
	}
}

// createEc2Instances EC2インスタンス（Bastion Host）を作成
func (g *MultiResourceGroup) createEc2Instances(scope constructs.Construct, props *MultiResourceGroupProps) {
	// 共通のUserData
	userData := awsec2.UserData_ForLinux(nil)
	userData.AddCommands(
		jsii.String("sudo dnf update -y"),
		jsii.String("sudo dnf install -y postgresql16"),
	)

	for i := 0; i < props.Ec2Count; i++ {
		instanceName := fmt.Sprintf("BastionHost%d", i+1)

		instance := awsec2.NewBastionHostLinux(
			scope,
			jsii.String(instanceName),
			&awsec2.BastionHostLinuxProps{
				Vpc:          props.Vpc,
				InstanceType: awsec2.InstanceType_Of(awsec2.InstanceClass_BURSTABLE4_GRAVITON, awsec2.InstanceSize_MICRO),
				SubnetSelection: &awsec2.SubnetSelection{
					SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS,
				},
				MachineImage: awsec2.NewAmazonLinuxImage(&awsec2.AmazonLinuxImageProps{
					Generation: awsec2.AmazonLinuxGeneration_AMAZON_LINUX_2023,
					CpuType:    awsec2.AmazonLinuxCpuType_ARM_64,
					UserData:   userData,
				}),
			},
		)
		g.Ec2Instances = append(g.Ec2Instances, instance)

		// 作成したRDSインスタンスへの接続を許可
		for _, rdsInstance := range g.RdsInstances {
			rdsInstance.Connections().AllowFrom(
				instance,
				awsec2.Port_Tcp(jsii.Number(5432)),
				jsii.String(fmt.Sprintf("Allow access from %s", instanceName)),
			)
		}
	}
}

// createAuroraClusters Auroraクラスターを作成
func (g *MultiResourceGroup) createAuroraClusters(scope constructs.Construct, props *MultiResourceGroupProps) {
	for i := 0; i < props.AuroraCount; i++ {
		clusterName := fmt.Sprintf("AuroraCluster%d", i+1)

		cluster := awsrds.NewDatabaseCluster(
			scope,
			jsii.String(clusterName),
			&awsrds.DatabaseClusterProps{
				Engine: awsrds.DatabaseClusterEngine_AuroraPostgres(&awsrds.AuroraPostgresClusterEngineProps{
					Version: awsrds.AuroraPostgresEngineVersion_VER_16_4(),
				}),
				Vpc: props.Vpc,
				VpcSubnets: &awsec2.SubnetSelection{
					SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS,
				},
				RemovalPolicy:           awscdk.RemovalPolicy_DESTROY,
				DeletionProtection:      jsii.Bool(false),
				StorageEncrypted:        jsii.Bool(true),
				ServerlessV2MinCapacity: jsii.Number(0.5),
				ServerlessV2MaxCapacity: jsii.Number(1),
				Writer: awsrds.ClusterInstance_ServerlessV2(jsii.String("writer"), &awsrds.ServerlessV2ClusterInstanceProps{
					InstanceIdentifier: jsii.String(fmt.Sprintf("%s-writer", clusterName)),
				}),
			},
		)
		g.AuroraClusters = append(g.AuroraClusters, cluster)

		// EC2インスタンスからの接続を許可
		for _, ec2Instance := range g.Ec2Instances {
			cluster.Connections().AllowFrom(
				ec2Instance,
				awsec2.Port_Tcp(jsii.Number(5432)),
				jsii.String(fmt.Sprintf("Allow access from EC2 to %s", clusterName)),
			)
		}
	}
}

// createS3Buckets S3バケットを作成
func (g *MultiResourceGroup) createS3Buckets(scope constructs.Construct, props *MultiResourceGroupProps) {
	for i := 0; i < props.S3Count; i++ {
		bucketName := fmt.Sprintf("DataBucket%d", i+1)

		bucket := awss3.NewBucket(
			scope,
			jsii.String(bucketName),
			&awss3.BucketProps{
				RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
				AutoDeleteObjects: jsii.Bool(true),
			},
		)
		g.S3Buckets = append(g.S3Buckets, bucket)
	}
}
