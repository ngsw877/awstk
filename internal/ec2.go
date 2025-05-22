package internal

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// StartEc2Instance EC2インスタンスを起動する
func StartEc2Instance(awsCtx AwsContext, instanceId string) error {
	cfg, err := LoadAwsConfig(awsCtx)
	if err != nil {
		return fmt.Errorf("AWS設定のロードに失敗: %w", err)
	}

	client := ec2.NewFromConfig(cfg)
	_, err = client.StartInstances(context.Background(), &ec2.StartInstancesInput{
		InstanceIds: []string{instanceId},
	})
	if err != nil {
		return fmt.Errorf("❌ EC2インスタンスの起動に失敗: %w", err)
	}
	return nil
}

// StopEc2Instance EC2インスタンスを停止する
func StopEc2Instance(awsCtx AwsContext, instanceId string) error {
	cfg, err := LoadAwsConfig(awsCtx)
	if err != nil {
		return fmt.Errorf("AWS設定のロードに失敗: %w", err)
	}

	client := ec2.NewFromConfig(cfg)
	_, err = client.StopInstances(context.Background(), &ec2.StopInstancesInput{
		InstanceIds: []string{instanceId},
	})
	if err != nil {
		return fmt.Errorf("❌ EC2インスタンスの停止に失敗: %w", err)
	}
	return nil
}
