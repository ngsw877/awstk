package ec2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// StartEc2Instance はEC2インスタンスを起動します
func StartEc2Instance(ec2Client *ec2.Client, instanceId string) error {
	input := &ec2.StartInstancesInput{
		InstanceIds: []string{instanceId},
	}

	_, err := ec2Client.StartInstances(context.Background(), input)
	if err != nil {
		return fmt.Errorf("EC2インスタンス起動エラー: %w", err)
	}

	return nil
}
