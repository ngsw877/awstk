package logs

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

// ListLogGroups はCloudWatch Logsグループの一覧を取得する関数
func ListLogGroups(client *cloudwatchlogs.Client) ([]types.LogGroup, error) {
	var logGroups []types.LogGroup
	var nextToken *string

	for {
		input := &cloudwatchlogs.DescribeLogGroupsInput{
			NextToken: nextToken,
		}

		result, err := client.DescribeLogGroups(context.Background(), input)
		if err != nil {
			return nil, fmt.Errorf("ログループ一覧取得エラー: %w", err)
		}

		logGroups = append(logGroups, result.LogGroups...)

		if result.NextToken == nil {
			break
		}
		nextToken = result.NextToken
	}

	return logGroups, nil
}

// FilterEmptyLogGroups は空のログループのみを返す関数
func FilterEmptyLogGroups(logGroups []types.LogGroup) []types.LogGroup {
	var emptyGroups []types.LogGroup

	for _, group := range logGroups {
		if isLogGroupEmpty(group) {
			emptyGroups = append(emptyGroups, group)
		}
	}

	return emptyGroups
}

// FilterNoRetentionLogGroups は保存期間が未設定のログループのみを返す関数
func FilterNoRetentionLogGroups(logGroups []types.LogGroup) []types.LogGroup {
	var noRetentionGroups []types.LogGroup

	for _, group := range logGroups {
		if group.RetentionInDays == nil {
			noRetentionGroups = append(noRetentionGroups, group)
		}
	}

	return noRetentionGroups
}

// isLogGroupEmpty はログループが空かどうかを判定する関数
func isLogGroupEmpty(group types.LogGroup) bool {
	// StoredBytesが0または存在しない場合は空と判定
	// AWS APIでは、完全に空のログループはStoredBytesフィールドが含まれないことがある
	return group.StoredBytes == nil || *group.StoredBytes == 0
}

// DisplayLogGroupDetails はログループの詳細情報を表示する関数
func DisplayLogGroupDetails(group types.LogGroup) {
	fmt.Printf("  - %s\n", *group.LogGroupName)
	
	// サイズ情報
	if group.StoredBytes != nil {
		fmt.Printf("    サイズ: %s\n", formatBytes(*group.StoredBytes))
	} else {
		fmt.Printf("    サイズ: 0 B (空)\n")
	}

	// 作成日時
	if group.CreationTime != nil {
		fmt.Printf("    作成日: %s\n", FormatTimestamp(group.CreationTime))
	}

	// 保存期間
	if group.RetentionInDays != nil {
		fmt.Printf("    保存期間: %d日\n", *group.RetentionInDays)
	} else {
		fmt.Printf("    保存期間: 無期限\n")
	}

	// メトリクスフィルター数
	if group.MetricFilterCount != nil && *group.MetricFilterCount > 0 {
		fmt.Printf("    メトリクスフィルター: %d個\n", *group.MetricFilterCount)
	}
}

// FormatTimestamp はUnixミリ秒のタイムスタンプをフォーマットする関数
func FormatTimestamp(timestamp *int64) string {
	if timestamp == nil {
		return "不明"
	}
	t := time.Unix(*timestamp/1000, (*timestamp%1000)*1000000)
	return t.Format("2006-01-02 15:04:05")
}

// formatBytes はバイト数を人間が読みやすい形式に変換する関数
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}