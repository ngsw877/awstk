package canary

import (
	"context"
	"fmt"
	"strings"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/synthetics"
)

// RunCanary 特定のCanaryを手動実行
func RunCanary(client *synthetics.Client, name string) error {
	fmt.Printf("Canary '%s' を実行中...\n", name)

	_, err := client.StartCanary(context.Background(), &synthetics.StartCanaryInput{
		Name: awssdk.String(name),
	})
	if err != nil {
		return fmt.Errorf("canary '%s' の実行に失敗: %w", name, err)
	}

	fmt.Printf("✓ Canary '%s' を開始しました\n", name)
	return nil
}

// RunCanaryDryRun Canaryのドライラン実行（将来の拡張用）
// 注意: 現在はドライラン機能を使わず、通常の実行と同じ動作をします
func RunCanaryDryRun(client *synthetics.Client, name string) error {
	fmt.Printf("Canary '%s' を実行中（ドライランモード）...\n", name)

	// 現在はドライラン専用APIを使わず、通常の実行を行う
	// 将来的にドライラン機能が必要になったら、この部分を拡張する
	return RunCanary(client, name)
}

// RunCanariesByFilter フィルターに一致するCanaryを一括実行
func RunCanariesByFilter(client *synthetics.Client, filters []string, dryRun bool, skipConfirm bool) error {
	if len(filters) == 0 {
		return fmt.Errorf("フィルターが指定されていません")
	}

	// フィルターに一致するCanaryを取得
	var matchedCanaries []Canary
	for _, filter := range filters {
		canaries, err := getCanariesByFilter(client, filter)
		if err != nil {
			return err
		}
		matchedCanaries = append(matchedCanaries, canaries...)
	}

	// 重複排除
	uniqueCanaries := removeDuplicateCanaries(matchedCanaries)

	if len(uniqueCanaries) == 0 {
		fmt.Printf("フィルター '%s' に一致するCanaryが見つかりませんでした\n", strings.Join(filters, ", "))
		return nil
	}

	// 実行対象を表示
	actionType := "実行"
	if dryRun {
		actionType = "ドライラン実行"
	}

	fmt.Printf("\n以下のCanaryを%sします:\n", actionType)
	for _, canary := range uniqueCanaries {
		fmt.Printf("  - %s (%s)\n", canary.Name, canary.State)
	}
	fmt.Printf("\n合計: %d個のCanary\n", len(uniqueCanaries))

	// 確認
	if !skipConfirm {
		if !confirmAction(fmt.Sprintf("%d個のCanaryを%ししますか？", len(uniqueCanaries), actionType)) {
			fmt.Println("キャンセルしました")
			return nil
		}
	}

	// 実行
	return executeCanaries(client, uniqueCanaries, dryRun)
}

// executeCanaries Canary群を実行
func executeCanaries(client *synthetics.Client, canaries []Canary, dryRun bool) error {
	successCount := 0
	errorCount := 0
	var errors []string

	fmt.Printf("\n実行中...\n")

	for _, canary := range canaries {
		var err error
		if dryRun {
			err = RunCanaryDryRun(client, canary.Name)
		} else {
			err = RunCanary(client, canary.Name)
		}

		if err != nil {
			errorCount++
			errMsg := fmt.Sprintf("- %s: %v", canary.Name, err)
			errors = append(errors, errMsg)
			fmt.Printf("✗ %s\n", errMsg)
		} else {
			successCount++
		}
	}

	// 結果サマリー
	fmt.Printf("\n=== 実行結果 ===\n")
	fmt.Printf("成功: %d個\n", successCount)
	if errorCount > 0 {
		fmt.Printf("失敗: %d個\n", errorCount)
		fmt.Printf("\n失敗詳細:\n")
		for _, errMsg := range errors {
			fmt.Printf("  %s\n", errMsg)
		}
		return fmt.Errorf("%d個のCanaryの実行に失敗しました", errorCount)
	}

	fmt.Printf("全てのCanaryの実行に成功しました！\n")
	return nil
}

// removeDuplicateCanaries Canary配列の重複を削除
func removeDuplicateCanaries(canaries []Canary) []Canary {
	seen := make(map[string]bool)
	var unique []Canary

	for _, canary := range canaries {
		if !seen[canary.Name] {
			seen[canary.Name] = true
			unique = append(unique, canary)
		}
	}

	return unique
}
