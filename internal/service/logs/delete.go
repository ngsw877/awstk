package logs

import (
	"awstk/internal/service/common"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
)

// DeleteLogGroups は指定されたオプションに基づいてロググループを削除します
// Force=true の場合、削除保護が有効なロググループも保護を解除して削除します
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

	// 削除保護の状態を事前チェック
	var protectedGroups []string
	for _, groupName := range targetGroups {
		protected, err := isDeletionProtected(client, groupName)
		if err != nil {
			return fmt.Errorf("削除保護状態の確認エラー (%s): %w", groupName, err)
		}
		if protected {
			protectedGroups = append(protectedGroups, groupName)
		}
	}

	// 削除保護が有効なロググループがある場合
	if len(protectedGroups) > 0 {
		if !opts.Force {
			fmt.Printf("⚠️  %d件のロググループで削除保護が有効です:\n", len(protectedGroups))
			for _, name := range protectedGroups {
				fmt.Printf("   🔒 %s\n", name)
			}
			fmt.Println("\n削除保護を解除して削除するには --force オプションを指定してください")
			return fmt.Errorf("削除保護が有効なロググループがあります")
		}
		fmt.Printf("⚠️  %d件のロググループで削除保護が有効です。--force により削除前に解除されます。\n\n", len(protectedGroups))
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
			err := deleteLogGroupWithProtectionCheck(client, groupName, opts.Force)

			resultsMutex.Lock()
			if err != nil {
				fmt.Printf("❌ %s ... 失敗 (%v)\n", groupName, err)
				results[idx] = common.ProcessResult{Item: groupName, Success: false, Error: err}
			} else {
				fmt.Printf("✅ %s ... 完了\n", groupName)
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

// deleteLogGroupWithProtectionCheck は削除保護を確認・解除してからロググループを削除します
// force=true の場合、削除保護が有効でも解除して削除します
func deleteLogGroupWithProtectionCheck(client *cloudwatchlogs.Client, logGroupName string, force bool) error {
	// 削除保護の確認
	protected, err := isDeletionProtected(client, logGroupName)
	if err != nil {
		return fmt.Errorf("削除保護状態の確認エラー: %w", err)
	}

	// 削除保護が有効な場合
	if protected && force {
		fmt.Printf("🔓 %s ... 削除保護を解除中\n", logGroupName)
		if err := disableDeletionProtection(client, logGroupName); err != nil {
			return fmt.Errorf("削除保護の解除エラー: %w", err)
		}
		// 削除保護解除が反映されるまで少し待つ
		time.Sleep(1 * time.Second)
	}

	// ロググループ削除
	_, err = client.DeleteLogGroup(context.Background(), &cloudwatchlogs.DeleteLogGroupInput{
		LogGroupName: &logGroupName,
	})
	return err
}

// isDeletionProtected はロググループの削除保護が有効かどうかを確認します
func isDeletionProtected(client *cloudwatchlogs.Client, logGroupName string) (bool, error) {
	output, err := client.DescribeLogGroups(context.Background(), &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: &logGroupName,
	})
	if err != nil {
		return false, err
	}

	// 完全一致するロググループを探す
	for _, lg := range output.LogGroups {
		if lg.LogGroupName != nil && *lg.LogGroupName == logGroupName {
			if lg.DeletionProtectionEnabled != nil {
				return *lg.DeletionProtectionEnabled, nil
			}
			return false, nil
		}
	}

	return false, nil
}

// disableDeletionProtection はロググループの削除保護を無効化します
func disableDeletionProtection(client *cloudwatchlogs.Client, logGroupName string) error {
	_, err := client.PutLogGroupDeletionProtection(context.Background(), &cloudwatchlogs.PutLogGroupDeletionProtectionInput{
		LogGroupIdentifier:        aws.String(logGroupName),
		DeletionProtectionEnabled: aws.Bool(false),
	})
	return err
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
			if common.MatchesFilter(*group.LogGroupName, opts.Filter, opts.Exact) {
				targetGroups = append(targetGroups, *group.LogGroupName)
			}
		}
	}

	// 重複を除去
	return common.RemoveDuplicates(targetGroups), nil
}

// GetLogGroupsByFilter はフィルターに一致するロググループを取得します（cleanup allから呼ばれる用）
// exact が true の場合、大文字小文字を区別します
func GetLogGroupsByFilter(client *cloudwatchlogs.Client, searchString string, exact bool) ([]string, error) {
	// すべてのロググループを取得
	allGroups, err := ListLogGroups(client)
	if err != nil {
		return nil, fmt.Errorf("ロググループ一覧取得エラー: %w", err)
	}

	var matchedGroups []string
	for _, group := range allGroups {
		if common.MatchesFilter(*group.LogGroupName, searchString, exact) {
			matchedGroups = append(matchedGroups, *group.LogGroupName)
			fmt.Printf("🔍 検出されたロググループ: %s\n", *group.LogGroupName)
		}
	}

	return matchedGroups, nil
}

// CleanupLogGroups は指定したロググループ一覧を削除します（cleanup allから呼ばれる用）
// cleanup allでは削除保護を自動的に解除して削除します（force=true相当）
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
			// cleanup allでは削除保護を自動解除（force=true）
			err := deleteLogGroupWithProtectionCheck(client, groupName, true)

			resultsMutex.Lock()
			if err != nil {
				fmt.Printf("❌ %s ... 失敗 (%v)\n", groupName, err)
				results[idx] = common.ProcessResult{Item: groupName, Success: false, Error: err}
			} else {
				fmt.Printf("✅ %s ... 完了\n", groupName)
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
