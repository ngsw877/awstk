package aurora

import (
	"context"
	"fmt"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/rds"

	"awstk/internal/service/cfn"
	"awstk/internal/service/common"
)

// ListAuroraClusters cmdから呼ばれるメイン関数（Get + Display）
func ListAuroraClusters(rdsClient *rds.Client, cfnClient *cloudformation.Client, stackName string) error {
	// Get: データ取得
	clusters, err := getAuroraClusters(rdsClient, cfnClient, stackName)
	if err != nil {
		if stackName != "" {
			return fmt.Errorf("❌ CloudFormationスタックからクラスター名の取得に失敗: %w", err)
		}
		return common.FormatListError("Auroraクラスター", err)
	}

	// Display: 共通表示処理
	return common.DisplayList(
		clusters,
		"Auroraクラスター一覧",
		auroraClustersToTableData,
		&common.DisplayOptions{
			ShowCount:    true,
			EmptyMessage: "Auroraクラスターが見つかりませんでした",
		},
	)
}

// getAuroraClusters データ取得内部関数
func getAuroraClusters(rdsClient *rds.Client, cfnClient *cloudformation.Client, stackName string) ([]Cluster, error) {
	if stackName != "" {
		return getAuroraClustersByStackName(rdsClient, cfnClient, stackName)
	}
	return getAllAuroraClusters(rdsClient)
}

// getAllAuroraClusters 現在のリージョンの全Auroraクラスターを取得
func getAllAuroraClusters(rdsClient *rds.Client) ([]Cluster, error) {
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

// getAuroraClustersByStackName 指定されたCloudFormationスタック名でフィルタリングしたAuroraクラスター一覧を取得
func getAuroraClustersByStackName(rdsClient *rds.Client, cfnClient *cloudformation.Client, stackName string) ([]Cluster, error) {
	ids, err := cfn.GetAllAuroraFromStack(cfnClient, stackName)
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return []Cluster{}, nil
	}

	all, err := getAllAuroraClusters(rdsClient)
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

// auroraClustersToTableData Auroraクラスター情報をテーブルデータに変換
func auroraClustersToTableData(clusters []Cluster) ([]common.TableColumn, [][]string) {
	columns := []common.TableColumn{
		{Header: "クラスターID"},
		{Header: "エンジン"},
		{Header: "ステータス"},
	}

	data := make([][]string, len(clusters))
	for i, cl := range clusters {
		data[i] = []string{
			cl.ClusterId,
			cl.Engine,
			cl.Status,
		}
	}
	return columns, data
}
