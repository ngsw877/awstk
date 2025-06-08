package internal

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// LoadAwsConfig はAWS設定を読み込む共通関数
func LoadAwsConfig(ctx AwsContext) (aws.Config, error) {
	opts := make([]func(*config.LoadOptions) error, 0)

	if ctx.Profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(ctx.Profile))
	}
	if ctx.Region != "" {
		opts = append(opts, config.WithRegion(ctx.Region))
	}
	return config.LoadDefaultConfig(context.Background(), opts...)
}

// SelectFromOptions はユーザーに複数の選択肢を提示して、1つを選択させる対話的な関数
func SelectFromOptions(title string, options []string) (int, error) {
	if len(options) == 0 {
		return -1, fmt.Errorf("選択肢がありません")
	}

	// 選択肢が1つの場合は自動選択
	if len(options) == 1 {
		fmt.Printf("✅ %s: %s (自動選択)\n", title, options[0])
		return 0, nil
	}

	// 複数の選択肢がある場合は対話的に選択
	fmt.Printf("\n🔍 %s:\n", title)
	for i, option := range options {
		fmt.Printf("  %d) %s\n", i+1, option)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("選択してください (1-%d): ", len(options))
		input, err := reader.ReadString('\n')
		if err != nil {
			return -1, fmt.Errorf("入力読み取りエラー: %w", err)
		}

		input = strings.TrimSpace(input)
		choice, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println("❌ 数値を入力してください")
			continue
		}

		if choice < 1 || choice > len(options) {
			fmt.Printf("❌ 1から%dの範囲で入力してください\n", len(options))
			continue
		}

		selectedIndex := choice - 1
		fmt.Printf("✅ 選択されました: %s\n", options[selectedIndex])
		return selectedIndex, nil
	}
}
