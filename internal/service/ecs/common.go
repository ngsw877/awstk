package ecs

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

// describeService はECSサービスの詳細情報を取得します
func describeService(ecsClient *ecs.Client, clusterName, serviceName string) (*types.Service, error) {
	// サービスの詳細を取得
	resp, err := ecsClient.DescribeServices(context.Background(), &ecs.DescribeServicesInput{
		Cluster:  aws.String(clusterName),
		Services: []string{serviceName},
	})
	if err != nil {
		return nil, fmt.Errorf("サービス情報の取得に失敗しました: %w", err)
	}

	if len(resp.Services) == 0 {
		return nil, fmt.Errorf("サービス '%s' が見つかりません", serviceName)
	}

	return &resp.Services[0], nil
}
