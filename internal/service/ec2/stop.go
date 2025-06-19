package ec2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// StopEc2Instance はEC2インスタンスを停止します
func StopEc2Instance(ec2Client *ec2.Client, instanceId string) error {
	input := &ec2.StopInstancesInput{
		InstanceIds: []string{instanceId},
	}

	_, err := ec2Client.StopInstances(context.Background(), input)
	if err != nil {
		return fmt.Errorf("EC2インスタンス停止エラー: %w", err)
	}

	return nil
}
