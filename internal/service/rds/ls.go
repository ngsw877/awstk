package rds

import (
	"context"
	"fmt"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/rds"

	"awstk/internal/service/cfn"
	"awstk/internal/service/common"
)

// ListRdsInstances cmdから呼ばれるメイン関数（Get + Display）
func ListRdsInstances(rdsClient *rds.Client, cfnClient *cloudformation.Client, stackName string) error {
	// Get: データ取得
	instances, err := getRdsInstances(rdsClient, cfnClient, stackName)
	if err != nil {
		if stackName != "" {
			return fmt.Errorf("❌ CloudFormationスタックからインスタンス名の取得に失敗: %w", err)
		}
		return common.FormatListError("RDSインスタンス", err)
	}

	// Display: 共通表示処理
	return common.DisplayList(
		instances,
		"RDSインスタンス一覧",
		rdsInstancesToTableData,
		&common.DisplayOptions{
			ShowCount:    true,
			EmptyMessage: "RDSインスタンスが見つかりませんでした",
		},
	)
}

// getRdsInstances データ取得内部関数
func getRdsInstances(rdsClient *rds.Client, cfnClient *cloudformation.Client, stackName string) ([]Instance, error) {
	if stackName != "" {
		return getRdsInstancesByStackName(rdsClient, cfnClient, stackName)
	}
	return getAllRdsInstances(rdsClient)
}

// getAllRdsInstances 現在のリージョンの全RDSインスタンスを取得
func getAllRdsInstances(rdsClient *rds.Client) ([]Instance, error) {
	resp, err := rdsClient.DescribeDBInstances(context.Background(), &rds.DescribeDBInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("RDSインスタンス一覧の取得に失敗: %w", err)
	}

	instances := make([]Instance, 0, len(resp.DBInstances))
	for _, db := range resp.DBInstances {
		instances = append(instances, Instance{
			InstanceId: awssdk.ToString(db.DBInstanceIdentifier),
			Engine:     awssdk.ToString(db.Engine),
			Status:     awssdk.ToString(db.DBInstanceStatus),
		})
	}

	return instances, nil
}

// getRdsInstancesByStackName 指定されたCloudFormationスタック名でフィルタリングしたRDSインスタンス一覧を取得
func getRdsInstancesByStackName(rdsClient *rds.Client, cfnClient *cloudformation.Client, stackName string) ([]Instance, error) {
	ids, err := cfn.GetAllRdsFromStack(cfnClient, stackName)
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return []Instance{}, nil
	}

	all, err := getAllRdsInstances(rdsClient)
	if err != nil {
		return nil, err
	}

	idSet := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
	}

	var instances []Instance
	for _, ins := range all {
		if _, ok := idSet[ins.InstanceId]; ok {
			instances = append(instances, ins)
		}
	}

	return instances, nil
}

// rdsInstancesToTableData RDSインスタンス情報をテーブルデータに変換
func rdsInstancesToTableData(instances []Instance) ([]common.TableColumn, [][]string) {
	columns := []common.TableColumn{
		{Header: "インスタンスID"},
		{Header: "エンジン"},
		{Header: "ステータス"},
	}
	
	data := make([][]string, len(instances))
	for i, ins := range instances {
		data[i] = []string{
			ins.InstanceId,
			ins.Engine,
			ins.Status,
		}
	}
	return columns, data
}