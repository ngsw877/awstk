package cfn

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/rds"
)

// StopAllStackResources はスタック内のすべてのリソースを停止します
func StopAllStackResources(cfnClient *cloudformation.Client, ec2Client *ec2.Client, rdsClient *rds.Client, aasClient *applicationautoscaling.Client, stackName string) error {
	// スタックからリソースを取得
	resources, err := getStartStopResourcesFromStack(cfnClient, stackName)
	if err != nil {
		return err
	}

	// 検出されたリソースのサマリーを表示
	printResourcesSummary(resources)

	errorsOccurred := false

	// EC2インスタンスを停止
	if len(resources.Ec2InstanceIds) > 0 {
		for _, instanceId := range resources.Ec2InstanceIds {
			fmt.Printf("🛑 EC2インスタンス (%s) を停止します...\n", instanceId)
			if err := stopEc2Instance(ec2Client, instanceId); err != nil {
				fmt.Printf("❌ EC2インスタンス (%s) の停止中にエラーが発生しました: %v\n", instanceId, err)
				errorsOccurred = true
			} else {
				fmt.Printf("✅ EC2インスタンス (%s) の停止を開始しました\n", instanceId)
			}
		}
	}

	// RDSインスタンスとAuroraクラスターを停止
	if len(resources.RdsInstanceIds) > 0 || len(resources.AuroraClusterIds) > 0 {
		// RDSインスタンスを停止
		for _, instanceId := range resources.RdsInstanceIds {
			fmt.Printf("🛑 RDSインスタンス (%s) を停止します...\n", instanceId)
			if err := stopRdsInstance(rdsClient, instanceId); err != nil {
				fmt.Printf("❌ RDSインスタンス (%s) の停止中にエラーが発生しました: %v\n", instanceId, err)
				errorsOccurred = true
			} else {
				fmt.Printf("✅ RDSインスタンス (%s) の停止を開始しました\n", instanceId)
			}
		}

		// Auroraクラスターを停止
		for _, clusterId := range resources.AuroraClusterIds {
			fmt.Printf("🛑 Aurora DBクラスター (%s) を停止します...\n", clusterId)
			if err := stopAuroraCluster(rdsClient, clusterId); err != nil {
				fmt.Printf("❌ Aurora DBクラスター (%s) の停止中にエラーが発生しました: %v\n", clusterId, err)
				errorsOccurred = true
			} else {
				fmt.Printf("✅ Aurora DBクラスター (%s) の停止を開始しました\n", clusterId)
			}
		}
	}

	// ECSサービスを停止
	if len(resources.EcsServiceInfo) > 0 {
		for _, ecsInfo := range resources.EcsServiceInfo {
			fmt.Printf("🛑 ECSサービス (%s/%s) を停止します...\n", ecsInfo.ClusterName, ecsInfo.ServiceName)
			capacityOpts := ServiceCapacityOptions{
				ClusterName: ecsInfo.ClusterName,
				ServiceName: ecsInfo.ServiceName,
				MinCapacity: 0, // 停止するために0に設定
				MaxCapacity: 0, // 停止するために0に設定
			}

			if err := setEcsServiceCapacity(aasClient, capacityOpts); err != nil {
				fmt.Printf("❌ ECSサービス (%s/%s) の停止中にエラーが発生しました: %v\n",
					ecsInfo.ClusterName, ecsInfo.ServiceName, err)
				errorsOccurred = true
			} else {
				fmt.Printf("✅ ECSサービス (%s/%s) の停止を開始しました\n",
					ecsInfo.ClusterName, ecsInfo.ServiceName)
			}
		}
	}

	if errorsOccurred {
		return fmt.Errorf("一部のリソースの停止中にエラーが発生しました")
	}
	return nil
}

// stopEc2Instance はEC2インスタンスを停止します
func stopEc2Instance(ec2Client *ec2.Client, instanceId string) error {
	input := &ec2.StopInstancesInput{
		InstanceIds: []string{instanceId},
	}

	_, err := ec2Client.StopInstances(context.Background(), input)
	if err != nil {
		return fmt.Errorf("EC2インスタンス停止エラー: %w", err)
	}

	return nil
}

// stopRdsInstance はRDSインスタンスを停止します
func stopRdsInstance(rdsClient *rds.Client, instanceId string) error {
	input := &rds.StopDBInstanceInput{
		DBInstanceIdentifier: &instanceId,
	}

	_, err := rdsClient.StopDBInstance(context.Background(), input)
	if err != nil {
		return fmt.Errorf("RDSインスタンス停止エラー: %w", err)
	}

	return nil
}

// stopAuroraCluster はAuroraクラスターを停止します
func stopAuroraCluster(rdsClient *rds.Client, clusterId string) error {
	input := &rds.StopDBClusterInput{
		DBClusterIdentifier: &clusterId,
	}

	_, err := rdsClient.StopDBCluster(context.Background(), input)
	if err != nil {
		return fmt.Errorf("auroraクラスター停止エラー: %w", err)
	}

	return nil
}
