package canary

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/synthetics"
)

// EnableCanary 指定したCanaryを有効化
func EnableCanary(client *synthetics.Client, name string) error {
	// 現在の状態を確認
	canaries, err := getAllCanaries(client)
	if err != nil {
		return err
	}

	var targetCanary *Canary
	for _, c := range canaries {
		if c.Name == name {
			targetCanary = &c
			break
		}
	}

	if targetCanary == nil {
		return fmt.Errorf("Canary '%s' が見つかりませんでした", name)
	}

	// 既に実行中の場合
	if targetCanary.State == CanaryStateRunning {
		fmt.Printf("ℹ️  %s は既に実行中です\n", name)
		return nil
	}

	// 有効化可能な状態かチェック
	if !canBeEnabled(targetCanary.State) {
		return fmt.Errorf("Canary '%s' は現在の状態(%s)では有効化できません", name, targetCanary.State)
	}

	// 有効化実行
	if err := startCanary(client, name); err != nil {
		return err
	}

	fmt.Printf("✅ %s を有効化しました\n", name)
	return nil
}

// EnableCanariesByFilter フィルタに一致するCanaryを有効化
func EnableCanariesByFilter(client *synthetics.Client, filter string, skipConfirm bool) error {
	// フィルタに一致するCanaryを取得
	canaries, err := getCanariesByFilter(client, filter)
	if err != nil {
		return err
	}

	if len(canaries) == 0 {
		return fmt.Errorf("フィルタ '%s' に一致するCanaryが見つかりませんでした", filter)
	}

	// 有効化対象のCanaryを選別
	var toEnable []Canary
	var alreadyRunning []string
	var cannotEnable []string

	for _, c := range canaries {
		if c.State == CanaryStateRunning {
			alreadyRunning = append(alreadyRunning, c.Name)
		} else if canBeEnabled(c.State) {
			toEnable = append(toEnable, c)
		} else {
			cannotEnable = append(cannotEnable, fmt.Sprintf("%s (%s)", c.Name, c.State))
		}
	}

	// 有効化対象がない場合
	if len(toEnable) == 0 {
		if len(alreadyRunning) > 0 {
			fmt.Printf("ℹ️  全てのCanaryが既に実行中です\n")
		}
		if len(cannotEnable) > 0 {
			fmt.Printf("⚠️  以下のCanaryは現在の状態では有効化できません:\n")
			for _, name := range cannotEnable {
				fmt.Printf("  - %s\n", name)
			}
		}
		return nil
	}

	// 確認プロンプト
	if !skipConfirm {
		fmt.Printf("以下の%d個のCanaryを有効化します:\n", len(toEnable))
		for _, c := range toEnable {
			fmt.Printf("  - %s (現在: %s)\n", c.Name, formatState(c.State))
		}
		if !confirmAction("続行しますか？") {
			return fmt.Errorf("キャンセルされました")
		}
	}

	// 有効化実行
	var errors []error
	successCount := 0
	for _, canary := range toEnable {
		if err := startCanary(client, canary.Name); err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", canary.Name, err))
		} else {
			fmt.Printf("✅ %s を有効化しました\n", canary.Name)
			successCount++
		}
	}

	// 結果サマリー
	if len(alreadyRunning) > 0 {
		fmt.Printf("\nℹ️  既に実行中: %d個\n", len(alreadyRunning))
	}
	if successCount > 0 {
		fmt.Printf("✅ 有効化成功: %d個\n", successCount)
	}
	if len(errors) > 0 {
		fmt.Printf("❌ 有効化失敗: %d個\n", len(errors))
		for _, err := range errors {
			fmt.Printf("  - %v\n", err)
		}
		return fmt.Errorf("一部のCanaryの有効化に失敗しました")
	}

	return nil
}

// EnableAllCanaries 全てのCanaryを有効化
func EnableAllCanaries(client *synthetics.Client, skipConfirm bool) error {
	canaries, err := getAllCanaries(client)
	if err != nil {
		return err
	}

	if len(canaries) == 0 {
		return fmt.Errorf("Canaryが見つかりませんでした")
	}

	// 有効化対象のCanaryを選別
	var toEnable []Canary
	var alreadyRunning []string
	var cannotEnable []string

	for _, c := range canaries {
		if c.State == CanaryStateRunning {
			alreadyRunning = append(alreadyRunning, c.Name)
		} else if canBeEnabled(c.State) {
			toEnable = append(toEnable, c)
		} else {
			cannotEnable = append(cannotEnable, fmt.Sprintf("%s (%s)", c.Name, c.State))
		}
	}

	// 有効化対象がない場合
	if len(toEnable) == 0 {
		if len(alreadyRunning) > 0 {
			fmt.Printf("ℹ️  全てのCanaryが既に実行中です (%d個)\n", len(alreadyRunning))
		}
		if len(cannotEnable) > 0 {
			fmt.Printf("⚠️  以下のCanaryは現在の状態では有効化できません:\n")
			for _, name := range cannotEnable {
				fmt.Printf("  - %s\n", name)
			}
		}
		return nil
	}

	// 確認プロンプト
	if !skipConfirm {
		fmt.Printf("以下の%d個のCanaryを有効化します:\n", len(toEnable))
		for _, c := range toEnable {
			fmt.Printf("  - %s (現在: %s)\n", c.Name, formatState(c.State))
		}
		if !confirmAction("続行しますか？") {
			return fmt.Errorf("キャンセルされました")
		}
	}

	// 有効化実行
	var errors []error
	successCount := 0
	for _, canary := range toEnable {
		if err := startCanary(client, canary.Name); err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", canary.Name, err))
		} else {
			fmt.Printf("✅ %s を有効化しました\n", canary.Name)
			successCount++
		}
	}

	// 結果サマリー
	fmt.Printf("\n--- 実行結果 ---\n")
	if len(alreadyRunning) > 0 {
		fmt.Printf("ℹ️  既に実行中: %d個\n", len(alreadyRunning))
	}
	if successCount > 0 {
		fmt.Printf("✅ 有効化成功: %d個\n", successCount)
	}
	if len(cannotEnable) > 0 {
		fmt.Printf("⚠️  状態により対象外: %d個\n", len(cannotEnable))
	}
	if len(errors) > 0 {
		fmt.Printf("❌ 有効化失敗: %d個\n", len(errors))
		for _, err := range errors {
			fmt.Printf("  - %v\n", err)
		}
		return fmt.Errorf("一部のCanaryの有効化に失敗しました")
	}

	return nil
}
