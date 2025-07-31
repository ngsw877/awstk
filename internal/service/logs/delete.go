package logs

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/gobwas/glob"
)

// DeleteLogGroups は指定されたオプションに基づいてロググループを削除します
func DeleteLogGroups(client *cloudwatchlogs.Client, opts DeleteOptions) error {
	// 削除対象のロググループを収集
	targetGroups, err := collectTargetLogGroups(client, opts)
	if err != nil {
		return fmt.Errorf("削除対象の収集に失敗: %w", err)
	}

	if len(targetGroups) == 0 {
		fmt.Println("削除対象のロググループがありません")
		return nil
	}

	// 削除実行
	successCount := 0
	failCount := 0

	fmt.Printf("🗑️  %d個のロググループを削除します...\n\n", len(targetGroups))

	for _, logGroupName := range targetGroups {
		fmt.Printf("削除中: %s ... ", logGroupName)

		_, err := client.DeleteLogGroup(context.Background(), &cloudwatchlogs.DeleteLogGroupInput{
			LogGroupName: &logGroupName,
		})

		if err != nil {
			fmt.Printf("❌ 失敗 (%v)\n", err)
			failCount++
		} else {
			fmt.Println("✅ 完了")
			successCount++
		}
	}

	// サマリー表示
	fmt.Printf("\n削除完了: 成功 %d個, 失敗 %d個\n", successCount, failCount)

	if failCount > 0 {
		return fmt.Errorf("%d個のロググループの削除に失敗しました", failCount)
	}

	return nil
}

// collectTargetLogGroups は削除対象のロググループを収集します
func collectTargetLogGroups(client *cloudwatchlogs.Client, opts DeleteOptions) ([]string, error) {
	var targetGroups []string

	// 位置引数で指定されたロググループを追加
	if len(opts.LogGroups) > 0 {
		targetGroups = append(targetGroups, opts.LogGroups...)
	}

	// フィルターが指定されている場合
	if opts.Filter != "" {
		// すべてのロググループを取得
		allGroups, err := ListLogGroups(client)
		if err != nil {
			return nil, err
		}

		// フィルター適用（まず追加フィルターを適用）
		filteredGroups := allGroups
		if opts.EmptyOnly {
			filteredGroups = FilterEmptyLogGroups(filteredGroups)
		}
		if opts.NoRetention {
			filteredGroups = FilterNoRetentionLogGroups(filteredGroups)
		}

		// パターンマッチングを適用
		// ワイルドカードが含まれている場合はglobパターン、そうでない場合は部分一致
		if strings.ContainsAny(opts.Filter, "*?[]") {
			pattern := glob.MustCompile(opts.Filter)
			for _, group := range filteredGroups {
				if pattern.Match(*group.LogGroupName) {
					targetGroups = append(targetGroups, *group.LogGroupName)
				}
			}
		} else {
			// ワイルドカードがない場合は部分一致
			for _, group := range filteredGroups {
				if strings.Contains(*group.LogGroupName, opts.Filter) {
					targetGroups = append(targetGroups, *group.LogGroupName)
				}
			}
		}
	}

	// 重複を除去
	return removeDuplicates(targetGroups), nil
}

// removeDuplicates は文字列スライスから重複を除去します
func removeDuplicates(items []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// GetLogGroupsByFilter はフィルターに一致するロググループを取得します（cleanup allから呼ばれる用）
func GetLogGroupsByFilter(client *cloudwatchlogs.Client, searchString string) ([]string, error) {
	// すべてのロググループを取得
	allGroups, err := ListLogGroups(client)
	if err != nil {
		return nil, fmt.Errorf("ロググループ一覧取得エラー: %w", err)
	}

	var matchedGroups []string
	for _, group := range allGroups {
		if strings.Contains(*group.LogGroupName, searchString) {
			matchedGroups = append(matchedGroups, *group.LogGroupName)
			fmt.Printf("🔍 検出されたロググループ: %s\n", *group.LogGroupName)
		}
	}

	return matchedGroups, nil
}

// CleanupLogGroups は指定したロググループ一覧を削除します（cleanup allから呼ばれる用）
func CleanupLogGroups(client *cloudwatchlogs.Client, logGroupNames []string) error {
	for _, logGroupName := range logGroupNames {
		fmt.Printf("ロググループ %s を削除中...\n", logGroupName)

		_, err := client.DeleteLogGroup(context.Background(), &cloudwatchlogs.DeleteLogGroupInput{
			LogGroupName: &logGroupName,
		})

		if err != nil {
			fmt.Printf("❌ ロググループ %s の削除に失敗しました: %v\n", logGroupName, err)
			// エラーをログに記録して続行
			continue
		}

		fmt.Printf("✅ ロググループ %s を削除しました\n", logGroupName)
	}

	return nil
}
