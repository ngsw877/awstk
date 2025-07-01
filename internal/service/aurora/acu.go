package aurora

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/rds"
)

// GetAuroraCapacityInfo Aurora Serverless v2のAcu情報を取得
func GetAuroraCapacityInfo(rdsClient *rds.Client, cwClient *cloudwatch.Client, clusterName string) (*AuroraCapacityInfo, error) {
	// まずクラスター情報を取得
	describeInput := &rds.DescribeDBClustersInput{
		DBClusterIdentifier: aws.String(clusterName),
	}

	result, err := rdsClient.DescribeDBClusters(context.Background(), describeInput)
	if err != nil {
		return nil, fmt.Errorf("クラスター情報の取得に失敗: %w", err)
	}

	if len(result.DBClusters) == 0 {
		return nil, fmt.Errorf("クラスター '%s' が見つかりません", clusterName)
	}

	cluster := result.DBClusters[0]
	info := &AuroraCapacityInfo{
		ClusterId:    aws.ToString(cluster.DBClusterIdentifier),
		Status:       aws.ToString(cluster.Status),
		IsServerless: cluster.ServerlessV2ScalingConfiguration != nil,
	}

	// Serverless v2でない場合は終了
	if !info.IsServerless {
		return info, nil
	}

	// 最小・最大Acuを取得
	scaling := cluster.ServerlessV2ScalingConfiguration
	info.MinAcu = aws.ToFloat64(scaling.MinCapacity)
	info.MaxAcu = aws.ToFloat64(scaling.MaxCapacity)

	// CloudWatchから現在のAcuを取得
	currentAcu, err := getCurrentAcuFromCloudWatch(cwClient, clusterName)
	if err != nil {
		// エラーがあっても部分的な情報は返す（エラーメッセージは表示しない）
		// CloudWatchにデータがない場合やアクセス権限がない場合がある
		info.CurrentAcu = -1 // -1 を「取得できなかった」を示す値として使用
		return info, nil
	}
	info.CurrentAcu = currentAcu

	return info, nil
}

// getCurrentAcuFromCloudWatch CloudWatchから現在のAcu値を取得
func getCurrentAcuFromCloudWatch(cwClient *cloudwatch.Client, clusterName string) (float64, error) {
	now := time.Now()
	startTime := now.Add(-5 * time.Minute) // 過去5分間に拡大（データがない可能性を考慮）

	input := &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String("AWS/RDS"),
		MetricName: aws.String("ServerlessDatabaseCapacity"),
		Dimensions: []types.Dimension{
			{
				Name:  aws.String("DBClusterIdentifier"),
				Value: aws.String(clusterName),
			},
		},
		StartTime:  aws.Time(startTime),
		EndTime:    aws.Time(now),
		Period:     aws.Int32(300), // 5分間隔に変更（データポイントを取得しやすくする）
		Statistics: []types.Statistic{types.StatisticAverage},
	}

	result, err := cwClient.GetMetricStatistics(context.Background(), input)
	if err != nil {
		return 0, err
	}

	if len(result.Datapoints) == 0 {
		return 0, fmt.Errorf("Acuメトリクスが取得できません")
	}

	// 最新のデータポイントを探す
	var latestDatapoint *types.Datapoint
	for i := range result.Datapoints {
		dp := &result.Datapoints[i]
		if latestDatapoint == nil || dp.Timestamp.After(*latestDatapoint.Timestamp) {
			latestDatapoint = dp
		}
	}

	return aws.ToFloat64(latestDatapoint.Average), nil
}

// ListAuroraCapacityInfo 複数クラスターのAcu情報を取得
func ListAuroraCapacityInfo(rdsClient *rds.Client, cwClient *cloudwatch.Client) ([]AuroraCapacityInfo, error) {
	// 全クラスターを取得
	clusters, err := ListAuroraClusters(rdsClient)
	if err != nil {
		return nil, err
	}

	var capacityInfos []AuroraCapacityInfo
	for _, cluster := range clusters {
		info, err := GetAuroraCapacityInfo(rdsClient, cwClient, cluster.ClusterId)
		if err != nil {
			// エラーがあっても続行（部分的な情報を含む）
			if info != nil {
				capacityInfos = append(capacityInfos, *info)
			}
			continue
		}
		if info.IsServerless {
			capacityInfos = append(capacityInfos, *info)
		}
	}

	return capacityInfos, nil
}