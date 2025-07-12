package aurora

import (
	"context"
	"fmt"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/rds"

	"awstk/internal/service/cfn"
)

// ListAuroraClusters 現在のリージョンのAuroraクラスター一覧を取得する
func ListAuroraClusters(rdsClient *rds.Client) ([]Cluster, error) {
	resp, err := rdsClient.DescribeDBClusters(context.Background(), &rds.DescribeDBClustersInput{})
	if err != nil {
		return nil, fmt.Errorf("Auroraクラスター一覧の取得に失敗: %w", err)
	}

	clusters := make([]Cluster, 0, len(resp.DBClusters))
	for _, c := range resp.DBClusters {
		clusters = append(clusters, Cluster{
			ClusterId: awssdk.ToString(c.DBClusterIdentifier),
			Engine:    awssdk.ToString(c.Engine),
			Status:    awssdk.ToString(c.Status),
		})
	}

	return clusters, nil
}

// ListAuroraClustersFromStack 指定されたCloudFormationスタックに属するAuroraクラスター一覧を取得する
func ListAuroraClustersFromStack(rdsClient *rds.Client, cfnClient *cloudformation.Client, stackName string) ([]Cluster, error) {
	ids, err := cfn.GetAllAuroraFromStack(cfnClient, stackName)
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return []Cluster{}, nil
	}

	all, err := ListAuroraClusters(rdsClient)
	if err != nil {
		return nil, err
	}

	idSet := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
	}

	var clusters []Cluster
	for _, cl := range all {
		if _, ok := idSet[cl.ClusterId]; ok {
			clusters = append(clusters, cl)
		}
	}

	return clusters, nil
}
