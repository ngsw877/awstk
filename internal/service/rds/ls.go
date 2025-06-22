package rds

import (
	"context"
	"fmt"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/rds"

	"awstk/internal/service/cfn"
)

// ListRdsInstances 現在のリージョンのRDSインスタンス一覧を取得する
func ListRdsInstances(rdsClient *rds.Client) ([]RdsInstance, error) {
	resp, err := rdsClient.DescribeDBInstances(context.Background(), &rds.DescribeDBInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("RDSインスタンス一覧の取得に失敗: %w", err)
	}

	instances := make([]RdsInstance, 0, len(resp.DBInstances))
	for _, db := range resp.DBInstances {
		instances = append(instances, RdsInstance{
			InstanceId: awssdk.ToString(db.DBInstanceIdentifier),
			Engine:     awssdk.ToString(db.Engine),
			Status:     awssdk.ToString(db.DBInstanceStatus),
		})
	}

	return instances, nil
}

// ListRdsInstancesFromStack 指定されたCloudFormationスタックに属するRDSインスタンス一覧を取得する
func ListRdsInstancesFromStack(rdsClient *rds.Client, cfnClient *cloudformation.Client, stackName string) ([]RdsInstance, error) {
	ids, err := cfn.GetAllRdsFromStack(cfnClient, stackName)
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return []RdsInstance{}, nil
	}

	all, err := ListRdsInstances(rdsClient)
	if err != nil {
		return nil, err
	}

	idSet := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
	}

	var instances []RdsInstance
	for _, ins := range all {
		if _, ok := idSet[ins.InstanceId]; ok {
			instances = append(instances, ins)
		}
	}

	return instances, nil
}
