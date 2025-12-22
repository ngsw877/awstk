package logs

import (
	"awstk/internal/service/common"
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
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

	// 並列実行数を設定（最大20並列）
	maxWorkers := 20
	if len(targetGroups) < maxWorkers {
		maxWorkers = len(targetGroups)
	}

	executor := common.NewParallelExecutor(maxWorkers)
	results := make([]common.ProcessResult, len(targetGroups))
	resultsMutex := &sync.Mutex{}

	fmt.Printf("🗑️  %d個のロググループを最大%d並列で削除します...\n\n", len(targetGroups), maxWorkers)

	for i, logGroupName := range targetGroups {
		idx := i
		groupName := logGroupName
		executor.Execute(func() {
			fmt.Printf("削除中: %s ... ", groupName)

			_, err := client.DeleteLogGroup(context.Background(), &cloudwatchlogs.DeleteLogGroupInput{
				LogGroupName: &groupName,
			})

			resultsMutex.Lock()
			if err != nil {
				fmt.Printf("❌ 失敗 (%v)\n", err)
				results[idx] = common.ProcessResult{Item: groupName, Success: false, Error: err}
			} else {
				fmt.Println("✅ 完了")
				results[idx] = common.ProcessResult{Item: groupName, Success: true}
			}
			resultsMutex.Unlock()
		})
	}

	executor.Wait()

	// 結果の集計
	successCount, failCount := common.CollectResults(results)
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
		for _, group := range filteredGroups {
			if common.MatchesFilter(*group.LogGroupName, opts.Filter) {
				targetGroups = append(targetGroups, *group.LogGroupName)
			}
		}
	}

	// 重複を除去
	return common.RemoveDuplicates(targetGroups), nil
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
		if common.MatchesFilter(*group.LogGroupName, searchString) {
			matchedGroups = append(matchedGroups, *group.LogGroupName)
			fmt.Printf("🔍 検出されたロググループ: %s\n", *group.LogGroupName)
		}
	}

	return matchedGroups, nil
}

// CleanupLogGroups は指定したロググループ一覧を削除します（cleanup allから呼ばれる用）
func CleanupLogGroups(client *cloudwatchlogs.Client, logGroupNames []string) common.CleanupResult {
	result := common.CleanupResult{
		ResourceType: "CloudWatch Logsグループ",
		Deleted:      []string{},
		Failed:       []string{},
	}

	if len(logGroupNames) == 0 {
		return result
	}

	// 並列実行数を設定（最大20並列）
	maxWorkers := 20
	if len(logGroupNames) < maxWorkers {
		maxWorkers = len(logGroupNames)
	}

	executor := common.NewParallelExecutor(maxWorkers)
	results := make([]common.ProcessResult, len(logGroupNames))
	resultsMutex := &sync.Mutex{}

	fmt.Printf("🚀 %d個のロググループを最大%d並列で削除します...\n\n", len(logGroupNames), maxWorkers)

	for i, logGroupName := range logGroupNames {
		idx := i
		groupName := logGroupName
		executor.Execute(func() {
			fmt.Printf("ロググループ %s を削除中...\n", groupName)

			_, err := client.DeleteLogGroup(context.Background(), &cloudwatchlogs.DeleteLogGroupInput{
				LogGroupName: &groupName,
			})

			resultsMutex.Lock()
			if err != nil {
				fmt.Printf("❌ ロググループ %s の削除に失敗しました: %v\n", groupName, err)
				results[idx] = common.ProcessResult{Item: groupName, Success: false, Error: err}
			} else {
				fmt.Printf("✅ ロググループ %s を削除しました\n", groupName)
				results[idx] = common.ProcessResult{Item: groupName, Success: true}
			}
			resultsMutex.Unlock()
		})
	}

	executor.Wait()

	// 結果の集計
	successCount, failCount := common.CollectResults(results)
	fmt.Printf("\n✅ 削除完了: 成功 %d個, 失敗 %d個\n", successCount, failCount)

	return common.CollectCleanupResult("CloudWatch Logsグループ", results)
}
