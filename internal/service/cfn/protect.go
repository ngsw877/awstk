package cfn

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

// protectionStatus は削除保護の状態を文字列で返します
func protectionStatus(enabled bool) string {
	if enabled {
		return "有効"
	}
	return "無効"
}

// protectionAction は削除保護の操作を文字列で返します
func protectionAction(enable bool) string {
	if enable {
		return "有効化"
	}
	return "無効化"
}

// UpdateProtection は指定した条件に一致するスタックの削除保護を更新します
func UpdateProtection(cfnClient *cloudformation.Client, opts ProtectOptions) error {
	// 対象のスタックを検索
	stacks, err := findStacksForProtect(cfnClient, opts)
	if err != nil {
		return err
	}

	if len(stacks) == 0 {
		fmt.Println("対象のスタックが見つかりませんでした")
		return nil
	}

	// 変更対象のスタック一覧を表示
	action := protectionAction(opts.Enable)

	fmt.Printf("🔍 削除保護を%sするスタック:\n", action)
	for _, stack := range stacks {
		currentStatus := protectionStatus(aws.ToBool(stack.EnableTerminationProtection))
		fmt.Printf("  - %s (現在の削除保護: %s)\n", aws.ToString(stack.StackName), currentStatus)
	}
	fmt.Printf("\n合計 %d 個のスタックの削除保護を%sします\n", len(stacks), action)

	// 削除保護を更新
	fmt.Printf("\n削除保護の%sを開始します...\n", action)
	updateCount := 0
	skipCount := 0

	for _, stack := range stacks {
		stackName := aws.ToString(stack.StackName)
		currentProtection := aws.ToBool(stack.EnableTerminationProtection)

		// 既に希望の状態になっている場合はスキップ
		if currentProtection == opts.Enable {
			fmt.Printf("⏭️  スタック %s は既に削除保護が%s状態です。スキップします\n",
				stackName,
				protectionStatus(opts.Enable))
			skipCount++
			continue
		}

		fmt.Printf("スタック %s の削除保護を%s中...", stackName, action)

		_, err := cfnClient.UpdateTerminationProtection(context.Background(), &cloudformation.UpdateTerminationProtectionInput{
			StackName:                   aws.String(stackName),
			EnableTerminationProtection: aws.Bool(opts.Enable),
		})
		if err != nil {
			fmt.Printf("\n❌ スタック %s の削除保護更新に失敗しました: %v\n", stackName, err)
			continue
		}
		fmt.Printf(" ✅\n")
		updateCount++
	}

	fmt.Printf("\n✅ %d 個のスタックの削除保護を%sしました\n", updateCount, action)
	if skipCount > 0 {
		fmt.Printf("ℹ️  %d 個のスタックは既に希望の状態のためスキップされました\n", skipCount)
	}

	return nil
}

// findStacksForProtect は削除保護変更対象のスタックを検索します
func findStacksForProtect(cfnClient *cloudformation.Client, opts ProtectOptions) ([]types.Stack, error) {
	var allStacks []types.Stack

	// スタック名が指定されている場合
	if len(opts.Stacks) > 0 {
		for _, stackName := range opts.Stacks {
			if stackName == "" {
				continue
			}

			// スタックの詳細情報を取得
			describeOutput, err := cfnClient.DescribeStacks(context.Background(), &cloudformation.DescribeStacksInput{
				StackName: aws.String(stackName),
			})
			if err != nil {
				fmt.Printf("⚠️  スタック %s が見つかりません: %v\n", stackName, err)
				continue
			}
			if len(describeOutput.Stacks) > 0 {
				allStacks = append(allStacks, describeOutput.Stacks[0])
			}
		}
		return allStacks, nil
	}

	// --filterまたは--statusの場合はfindStacksForCleanupのロジックを使用
	return findStacksForCleanup(cfnClient, CleanupOptions{
		Filter: opts.Filter,
		Status: opts.Status,
	})
}
