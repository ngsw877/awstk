package cfn

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/rds"
)

// StartAllStackResources はスタック内のすべてのリソースを起動します
func StartAllStackResources(opts StackStartStopOptions) error {
	// スタックからリソースを取得
	resources, err := getStartStopResourcesFromStack(opts.CfnClient, opts.StackName)
	if err != nil {
		return err
	}

	// 検出されたリソースのサマリーを表示
	printResourcesSummary(resources)

	errorsOccurred := false

	// EC2インスタンスを起動
	if len(resources.Ec2InstanceIds) > 0 {
		for _, instanceId := range resources.Ec2InstanceIds {
			fmt.Printf("🚀 EC2インスタンス (%s) を起動します...\n", instanceId)
			if err := startEc2Instance(opts.Ec2Client, instanceId); err != nil {
				fmt.Printf("❌ EC2インスタンス (%s) の起動中にエラーが発生しました: %v\n", instanceId, err)
				errorsOccurred = true
			} else {
				fmt.Printf("✅ EC2インスタンス (%s) の起動を開始しました\n", instanceId)
			}
		}
	}

	// RDSインスタンスとAuroraクラスターを起動
	if len(resources.RdsInstanceIds) > 0 || len(resources.AuroraClusterIds) > 0 {
		// RDSインスタンスを起動
		for _, instanceId := range resources.RdsInstanceIds {
			fmt.Printf("🚀 RDSインスタンス (%s) を起動します...\n", instanceId)
			if err := startRdsInstance(opts.RdsClient, instanceId); err != nil {
				fmt.Printf("❌ RDSインスタンス (%s) の起動中にエラーが発生しました: %v\n", instanceId, err)
				errorsOccurred = true
			} else {
				fmt.Printf("✅ RDSインスタンス (%s) の起動を開始しました\n", instanceId)
			}
		}

		// Auroraクラスターを起動
		for _, clusterId := range resources.AuroraClusterIds {
			fmt.Printf("🚀 Aurora DBクラスター (%s) を起動します...\n", clusterId)
			if err := startAuroraCluster(opts.RdsClient, clusterId); err != nil {
				fmt.Printf("❌ Aurora DBクラスター (%s) の起動中にエラーが発生しました: %v\n", clusterId, err)
				errorsOccurred = true
			} else {
				fmt.Printf("✅ Aurora DBクラスター (%s) の起動を開始しました\n", clusterId)
			}
		}
	}

	// ECSサービスを起動
	if len(resources.EcsServiceInfo) > 0 {
		for _, ecsInfo := range resources.EcsServiceInfo {
			fmt.Printf("🚀 ECSサービス (%s/%s) を起動します...\n", ecsInfo.ClusterName, ecsInfo.ServiceName)
			capacityOpts := ServiceCapacityOptions{
				ClusterName: ecsInfo.ClusterName,
				ServiceName: ecsInfo.ServiceName,
				MinCapacity: 1, // デフォルト値として1を使用
				MaxCapacity: 2, // デフォルト値として2を使用
			}

			if err := setEcsServiceCapacity(opts.ApplicationAutoScalingClient, capacityOpts); err != nil {
				fmt.Printf("❌ ECSサービス (%s/%s) の起動中にエラーが発生しました: %v\n",
					ecsInfo.ClusterName, ecsInfo.ServiceName, err)
				errorsOccurred = true
			} else {
				fmt.Printf("✅ ECSサービス (%s/%s) の起動を開始しました\n",
					ecsInfo.ClusterName, ecsInfo.ServiceName)
			}
		}
	}

	if errorsOccurred {
		return fmt.Errorf("一部のリソースの起動中にエラーが発生しました")
	}
	return nil
}

// ServiceCapacityOptions はECSサービスのキャパシティ設定用オプション
type ServiceCapacityOptions struct {
	ClusterName string
	ServiceName string
	MinCapacity int
	MaxCapacity int
}

// startEc2Instance はEC2インスタンスを起動します
func startEc2Instance(ec2Client *ec2.Client, instanceId string) error {
	input := &ec2.StartInstancesInput{
		InstanceIds: []string{instanceId},
	}

	_, err := ec2Client.StartInstances(context.Background(), input)
	if err != nil {
		return fmt.Errorf("EC2インスタンス起動エラー: %w", err)
	}

	return nil
}

// startRdsInstance はRDSインスタンスを起動します
func startRdsInstance(rdsClient *rds.Client, instanceId string) error {
	input := &rds.StartDBInstanceInput{
		DBInstanceIdentifier: &instanceId,
	}

	_, err := rdsClient.StartDBInstance(context.Background(), input)
	if err != nil {
		return fmt.Errorf("RDSインスタンス起動エラー: %w", err)
	}

	return nil
}

// startAuroraCluster はAuroraクラスターを起動します
func startAuroraCluster(rdsClient *rds.Client, clusterId string) error {
	input := &rds.StartDBClusterInput{
		DBClusterIdentifier: &clusterId,
	}

	_, err := rdsClient.StartDBCluster(context.Background(), input)
	if err != nil {
		return fmt.Errorf("Auroraクラスター起動エラー: %w", err)
	}

	return nil
}

// setEcsServiceCapacity はECSサービスのキャパシティを設定します
func setEcsServiceCapacity(autoScalingClient *applicationautoscaling.Client, opts ServiceCapacityOptions) error {
	// リソースIDを構築
	resourceId := fmt.Sprintf("service/%s/%s", opts.ClusterName, opts.ServiceName)

	// スケーラブルターゲットを登録
	_, err := autoScalingClient.RegisterScalableTarget(context.Background(), &applicationautoscaling.RegisterScalableTargetInput{
		ServiceNamespace:  "ecs",
		ScalableDimension: "ecs:service:DesiredCount",
		ResourceId:        &resourceId,
		MinCapacity:       aws.Int32(int32(opts.MinCapacity)),
		MaxCapacity:       aws.Int32(int32(opts.MaxCapacity)),
	})
	if err != nil {
		return fmt.Errorf("スケーラブルターゲット登録でエラー: %w", err)
	}

	return nil
}
