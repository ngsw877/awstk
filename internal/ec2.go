package internal

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// StartEc2Instance EC2インスタンスを起動する
func StartEc2Instance(instanceId, region, profile string) error {
	ctx := context.Background()
	cfg, err := LoadAwsConfig(region, profile)
	if err != nil {
		return fmt.Errorf("AWS設定のロードに失敗: %w", err)
	}

	client := ec2.NewFromConfig(cfg)
	_, err = client.StartInstances(ctx, &ec2.StartInstancesInput{
		InstanceIds: []string{instanceId},
	})
	if err != nil {
		return fmt.Errorf("❌ EC2インスタンスの起動に失敗: %w", err)
	}
	return nil
}

// StopEc2Instance EC2インスタンスを停止する
func StopEc2Instance(instanceId, region, profile string) error {
	ctx := context.Background()
	cfg, err := LoadAwsConfig(region, profile)
	if err != nil {
		return fmt.Errorf("AWS設定のロードに失敗: %w", err)
	}

	client := ec2.NewFromConfig(cfg)
	_, err = client.StopInstances(ctx, &ec2.StopInstancesInput{
		InstanceIds: []string{instanceId},
	})
	if err != nil {
		return fmt.Errorf("❌ EC2インスタンスの停止に失敗: %w", err)
	}
	return nil
}
