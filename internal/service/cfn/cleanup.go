package cfn

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

// CleanupStacks は指定した条件に一致するスタックを削除します
func CleanupStacks(cfnClient *cloudformation.Client, opts CleanupOptions) error {
	// 削除対象のスタックを検索
	stacks, err := findStacksForCleanup(cfnClient, opts)
	if err != nil {
		return err
	}

	if len(stacks) == 0 {
		fmt.Println("削除対象のスタックが見つかりませんでした")
		return nil
	}

	// 削除対象のスタック一覧を表示
	fmt.Println("🔍 削除対象のスタック:")
	for _, stack := range stacks {
		fmt.Printf("  - %s (Status: %s)\n", aws.ToString(stack.StackName), stack.StackStatus)
	}
	fmt.Printf("\n合計 %d 個のスタックが削除されます\n", len(stacks))

	// 確認プロンプト
	if !opts.Force {
		fmt.Print("\n本当に削除しますか？ [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("削除をキャンセルしました")
			return nil
		}
	}

	// スタックを削除
	fmt.Println("\n削除を開始します...")
	deleteCount := 0
	for _, stack := range stacks {
		stackName := aws.ToString(stack.StackName)
		fmt.Printf("スタック %s を削除中...", stackName)

		// 削除保護の確認
		if aws.ToBool(stack.EnableTerminationProtection) {
			fmt.Printf("\n⚠️  スタック %s は削除保護が有効です。スキップします\n", stackName)
			continue
		}

		_, err := cfnClient.DeleteStack(context.Background(), &cloudformation.DeleteStackInput{
			StackName: aws.String(stackName),
		})
		if err != nil {
			fmt.Printf("\n❌ スタック %s の削除に失敗しました: %v\n", stackName, err)
			continue
		}
		fmt.Printf(" ✅\n")
		deleteCount++
	}

	fmt.Printf("\n✅ %d 個のスタックの削除リクエストを送信しました\n", deleteCount)
	if deleteCount < len(stacks) {
		fmt.Printf("⚠️  %d 個のスタックはスキップされました\n", len(stacks)-deleteCount)
	}

	return nil
}

// findStacksForCleanup は指定した条件に一致するスタックを検索します
func findStacksForCleanup(cfnClient *cloudformation.Client, opts CleanupOptions) ([]types.Stack, error) {
	// ステータスフィルターの解析
	var targetStatuses []types.StackStatus
	if opts.Status != "" {
		statusList := strings.Split(opts.Status, ",")
		for _, status := range statusList {
			status = strings.TrimSpace(status)
			// 文字列をStackStatus型に変換
			targetStatuses = append(targetStatuses, types.StackStatus(status))
		}
	}

	var allStacks []types.Stack
	var nextToken *string

	// ページネーション対応でスタック一覧を取得
	for {
		input := &cloudformation.ListStacksInput{
			NextToken: nextToken,
		}

		// ステータスフィルターが指定されている場合は適用
		if len(targetStatuses) > 0 {
			input.StackStatusFilter = targetStatuses
		} else {
			// ステータス指定がない場合は、削除可能なステータスのスタックのみを取得
			// InProgress系は削除できないため除外（ただしREVIEW_IN_PROGRESSは削除可能）
			input.StackStatusFilter = []types.StackStatus{
				types.StackStatusCreateComplete,
				types.StackStatusCreateFailed,
				types.StackStatusDeleteFailed,
				types.StackStatusImportComplete,
				types.StackStatusImportRollbackComplete,
				types.StackStatusImportRollbackFailed,
				types.StackStatusReviewInProgress, // 変更セット作成中（実際の変更は未実行）なので削除可能
				types.StackStatusRollbackComplete,
				types.StackStatusRollbackFailed,
				types.StackStatusUpdateComplete,
				types.StackStatusUpdateFailed,
				types.StackStatusUpdateRollbackComplete,
				types.StackStatusUpdateRollbackFailed,
			}
		}

		output, err := cfnClient.ListStacks(context.Background(), input)
		if err != nil {
			return nil, fmt.Errorf("スタック一覧の取得に失敗しました: %w", err)
		}

		// 名前フィルターを適用
		for _, summary := range output.StackSummaries {
			if opts.Filter == "" || strings.Contains(aws.ToString(summary.StackName), opts.Filter) {
				// スタックの詳細情報を取得（削除保護の確認のため）
				describeOutput, err := cfnClient.DescribeStacks(context.Background(), &cloudformation.DescribeStacksInput{
					StackName: summary.StackName,
				})
				if err != nil {
					// スタックが削除中などで取得できない場合はスキップ
					continue
				}
				if len(describeOutput.Stacks) > 0 {
					allStacks = append(allStacks, describeOutput.Stacks[0])
				}
			}
		}

		if output.NextToken == nil {
			break
		}
		nextToken = output.NextToken
	}

	return allStacks, nil
}
