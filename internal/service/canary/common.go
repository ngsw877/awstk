package canary

import (
	"awstk/internal/service/common"
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
		if common.MatchesFilter(canary.Name, filter, false) {
			filtered = append(filtered, canary)
		}
	}

	return filtered, nil
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
		return fmt.Errorf("canaryの開始に失敗: %w", err)
	}
	return nil
}

// stopCanary Canaryを停止
func stopCanary(client *synthetics.Client, name string) error {
	_, err := client.StopCanary(context.Background(), &synthetics.StopCanaryInput{
		Name: awssdk.String(name),
	})
	if err != nil {
		return fmt.Errorf("canaryの停止に失敗: %w", err)
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
