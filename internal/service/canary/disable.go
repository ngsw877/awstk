package canary

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/synthetics"
)

// DisableCanary 指定したCanaryを無効化
func DisableCanary(client *synthetics.Client, name string) error {
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

	// 既に停止中の場合
	if targetCanary.State == CanaryStateStopped {
		fmt.Printf("ℹ️  %s は既に停止しています\n", name)
		return nil
	}

	// 無効化可能な状態かチェック
	if !canBeDisabled(targetCanary.State) {
		return fmt.Errorf("Canary '%s' は現在の状態(%s)では無効化できません", name, targetCanary.State)
	}

	// 無効化実行
	if err := stopCanary(client, name); err != nil {
		return err
	}

	fmt.Printf("✅ %s を無効化しました\n", name)
	return nil
}

// DisableCanariesByFilter フィルタに一致するCanaryを無効化
func DisableCanariesByFilter(client *synthetics.Client, filter string, skipConfirm bool) error {
	// フィルタに一致するCanaryを取得
	canaries, err := getCanariesByFilter(client, filter)
	if err != nil {
		return err
	}

	if len(canaries) == 0 {
		return fmt.Errorf("フィルタ '%s' に一致するCanaryが見つかりませんでした", filter)
	}

	// 無効化対象のCanaryを選別
	var toDisable []Canary
	var alreadyStopped []string
	var cannotDisable []string

	for _, c := range canaries {
		if c.State == CanaryStateStopped {
			alreadyStopped = append(alreadyStopped, c.Name)
		} else if canBeDisabled(c.State) {
			toDisable = append(toDisable, c)
		} else {
			cannotDisable = append(cannotDisable, fmt.Sprintf("%s (%s)", c.Name, c.State))
		}
	}

	// 無効化対象がない場合
	if len(toDisable) == 0 {
		if len(alreadyStopped) > 0 {
			fmt.Printf("ℹ️  全てのCanaryが既に停止しています\n")
		}
		if len(cannotDisable) > 0 {
			fmt.Printf("⚠️  以下のCanaryは現在の状態では無効化できません:\n")
			for _, name := range cannotDisable {
				fmt.Printf("  - %s\n", name)
			}
		}
		return nil
	}

	// 確認プロンプト
	if !skipConfirm {
		fmt.Printf("以下の%d個のCanaryを無効化します:\n", len(toDisable))
		for _, c := range toDisable {
			fmt.Printf("  - %s (現在: %s)\n", c.Name, formatState(c.State))
		}
		if !confirmAction("続行しますか？") {
			return fmt.Errorf("キャンセルされました")
		}
	}

	// 無効化実行
	var errors []error
	successCount := 0
	for _, canary := range toDisable {
		if err := stopCanary(client, canary.Name); err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", canary.Name, err))
		} else {
			fmt.Printf("✅ %s を無効化しました\n", canary.Name)
			successCount++
		}
	}

	// 結果サマリー
	if len(alreadyStopped) > 0 {
		fmt.Printf("\nℹ️  既に停止中: %d個\n", len(alreadyStopped))
	}
	if successCount > 0 {
		fmt.Printf("✅ 無効化成功: %d個\n", successCount)
	}
	if len(errors) > 0 {
		fmt.Printf("❌ 無効化失敗: %d個\n", len(errors))
		for _, err := range errors {
			fmt.Printf("  - %v\n", err)
		}
		return fmt.Errorf("一部のCanaryの無効化に失敗しました")
	}

	return nil
}

// DisableAllCanaries 全てのCanaryを無効化
func DisableAllCanaries(client *synthetics.Client, skipConfirm bool) error {
	canaries, err := getAllCanaries(client)
	if err != nil {
		return err
	}

	if len(canaries) == 0 {
		return fmt.Errorf("Canaryが見つかりませんでした")
	}

	// 無効化対象のCanaryを選別
	var toDisable []Canary
	var alreadyStopped []string
	var cannotDisable []string

	for _, c := range canaries {
		if c.State == CanaryStateStopped {
			alreadyStopped = append(alreadyStopped, c.Name)
		} else if canBeDisabled(c.State) {
			toDisable = append(toDisable, c)
		} else {
			cannotDisable = append(cannotDisable, fmt.Sprintf("%s (%s)", c.Name, c.State))
		}
	}

	// 無効化対象がない場合
	if len(toDisable) == 0 {
		if len(alreadyStopped) > 0 {
			fmt.Printf("ℹ️  全てのCanaryが既に停止しています (%d個)\n", len(alreadyStopped))
		}
		if len(cannotDisable) > 0 {
			fmt.Printf("⚠️  以下のCanaryは現在の状態では無効化できません:\n")
			for _, name := range cannotDisable {
				fmt.Printf("  - %s\n", name)
			}
		}
		return nil
	}

	// 確認プロンプト
	if !skipConfirm {
		fmt.Printf("⚠️  以下の%d個のCanaryを無効化します:\n", len(toDisable))
		for _, c := range toDisable {
			fmt.Printf("  - %s (現在: %s)\n", c.Name, formatState(c.State))
		}
		fmt.Printf("\n🔴 警告: 全てのCanaryが停止すると、監視が行われなくなります。\n")
		if !confirmAction("本当に続行しますか？") {
			return fmt.Errorf("キャンセルされました")
		}
	}

	// 無効化実行
	var errors []error
	successCount := 0
	for _, canary := range toDisable {
		if err := stopCanary(client, canary.Name); err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", canary.Name, err))
		} else {
			fmt.Printf("✅ %s を無効化しました\n", canary.Name)
			successCount++
		}
	}

	// 結果サマリー
	fmt.Printf("\n--- 実行結果 ---\n")
	if len(alreadyStopped) > 0 {
		fmt.Printf("ℹ️  既に停止中: %d個\n", len(alreadyStopped))
	}
	if successCount > 0 {
		fmt.Printf("✅ 無効化成功: %d個\n", successCount)
	}
	if len(cannotDisable) > 0 {
		fmt.Printf("⚠️  状態により対象外: %d個\n", len(cannotDisable))
	}
	if len(errors) > 0 {
		fmt.Printf("❌ 無効化失敗: %d個\n", len(errors))
		for _, err := range errors {
			fmt.Printf("  - %v\n", err)
		}
		return fmt.Errorf("一部のCanaryの無効化に失敗しました")
	}

	return nil
}