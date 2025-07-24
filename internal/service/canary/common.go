package canary

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/synthetics"
)

// getCanariesByFilter フィルタパターンに一致するCanaryを取得
func getCanariesByFilter(client *synthetics.Client, filter string) ([]Canary, error) {
	allCanaries, err := getAllCanaries(client)
	if err != nil {
		return nil, err
	}

	var filtered []Canary
	for _, canary := range allCanaries {
		if matchPattern(canary.Name, filter) {
			filtered = append(filtered, canary)
		}
	}

	return filtered, nil
}

// matchPattern ワイルドカードパターンマッチング
func matchPattern(name, pattern string) bool {
	// 単純なワイルドカード実装（* のみサポート）
	pattern = strings.ReplaceAll(pattern, "*", ".*")
	pattern = "^" + pattern + "$"

	// 簡易的なマッチング（正規表現は使わずに実装）
	if !strings.Contains(pattern, ".*") {
		return name == strings.Trim(pattern, "^$")
	}

	// ワイルドカードがある場合
	parts := strings.Split(strings.Trim(pattern, "^$"), ".*")

	// 先頭の一致確認
	if parts[0] != "" && !strings.HasPrefix(name, parts[0]) {
		return false
	}

	// 末尾の一致確認
	if len(parts) > 1 && parts[len(parts)-1] != "" && !strings.HasSuffix(name, parts[len(parts)-1]) {
		return false
	}

	// 中間部分の確認
	remaining := name
	for i, part := range parts {
		if part == "" {
			continue
		}
		idx := strings.Index(remaining, part)
		if idx == -1 {
			return false
		}
		// 最初のパートは先頭一致が必要
		if i == 0 && idx != 0 {
			return false
		}
		remaining = remaining[idx+len(part):]
	}

	return true
}

// confirmAction ユーザーに確認を求める
func confirmAction(message string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [y/N]: ", message)

	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

// startCanary Canaryを開始
func startCanary(client *synthetics.Client, name string) error {
	_, err := client.StartCanary(context.Background(), &synthetics.StartCanaryInput{
		Name: awssdk.String(name),
	})
	if err != nil {
		return fmt.Errorf("Canaryの開始に失敗: %w", err)
	}
	return nil
}

// stopCanary Canaryを停止
func stopCanary(client *synthetics.Client, name string) error {
	_, err := client.StopCanary(context.Background(), &synthetics.StopCanaryInput{
		Name: awssdk.String(name),
	})
	if err != nil {
		return fmt.Errorf("Canaryの停止に失敗: %w", err)
	}
	return nil
}

// canBeEnabled Canaryが有効化可能な状態か確認
func canBeEnabled(state string) bool {
	return state == CanaryStateStopped || state == CanaryStateReady
}

// canBeDisabled Canaryが無効化可能な状態か確認
func canBeDisabled(state string) bool {
	return state == CanaryStateRunning
}
