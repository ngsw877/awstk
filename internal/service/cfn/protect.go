package cfn

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
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
	// 対象のスタックを検索（cleanupと同じロジックを再利用）
	stacks, err := findStacksForCleanup(cfnClient, CleanupOptions{
		Filter: opts.Filter,
		Status: opts.Status,
	})
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

	// 確認プロンプト
	if !opts.Force {
		fmt.Printf("\n本当に削除保護を%sしますか？ [y/N]: ", action)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("処理をキャンセルしました")
			return nil
		}
	}

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
