package ssm

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// DeleteParametersFromFile はファイルからパラメータ名を読み込んでParameter Storeから削除する
func DeleteParametersFromFile(ssmClient *ssm.Client, opts DeleteParamsOptions) error {
	// ファイルの存在確認
	if _, err := os.Stat(opts.FilePath); os.IsNotExist(err) {
		return fmt.Errorf("ファイルが見つかりません: %s", opts.FilePath)
	}

	// パラメータ名の読み込み
	paramNames, err := loadParameterNamesFromFile(opts.FilePath)
	if err != nil {
		return fmt.Errorf("ファイルの読み込みに失敗しました: %w", err)
	}

	if len(paramNames) == 0 {
		return fmt.Errorf("削除するパラメータが見つかりません")
	}

	// プレフィックスの適用
	if opts.Prefix != "" {
		for i := range paramNames {
			paramNames[i] = normalizeParameterName(opts.Prefix, paramNames[i])
		}
	}

	// ドライランの場合は内容を表示して終了
	if opts.DryRun {
		fmt.Println("🗑️  以下のパラメータが削除されます:")
		fmt.Println(strings.Repeat("-", 80))
		for _, name := range paramNames {
			fmt.Printf("  %s\n", name)
		}
		fmt.Println(strings.Repeat("-", 80))
		fmt.Printf("📊 合計: %d 件\n", len(paramNames))
		return nil
	}

	// 確認プロンプト（--forceでない場合）
	if !opts.Force {
		fmt.Printf("⚠️  %d 件のパラメータを削除しようとしています。\n", len(paramNames))
		fmt.Print("本当に削除しますか？ [y/N]: ")

		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			fmt.Printf("⚠️  入力エラー: %v\n", err)
			fmt.Println("削除をキャンセルしました。")
			return nil
		}
		if strings.ToLower(response) != "y" {
			fmt.Println("削除をキャンセルしました。")
			return nil
		}
	}

	// パラメータの削除
	var successCount, failCount, notFoundCount int
	for _, name := range paramNames {
		err := deleteParameter(ssmClient, name)
		if err != nil {
			if strings.Contains(err.Error(), "ParameterNotFound") {
				fmt.Printf("⚠️  %s は存在しません（スキップ）\n", name)
				notFoundCount++
			} else {
				fmt.Printf("❌ %s の削除に失敗しました: %v\n", name, err)
				failCount++
			}
		} else {
			fmt.Printf("✅ %s を削除しました\n", name)
			successCount++
		}
	}

	fmt.Printf("\n📊 削除結果: 成功 %d / 失敗 %d / 存在しない %d / 合計 %d\n",
		successCount, failCount, notFoundCount, len(paramNames))

	if failCount > 0 {
		return fmt.Errorf("%d 件のパラメータ削除に失敗しました", failCount)
	}

	return nil
}

// loadParameterNamesFromFile はファイルからパラメータ名を読み込む
func loadParameterNamesFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("ファイルを開けません: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("⚠️  ファイルのクローズに失敗: %v\n", err)
		}
	}()

	var paramNames []string
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// 空行とコメント行（#で始まる）をスキップ
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// パラメータ名の妥当性チェック
		if !isValidParameterName(line) {
			fmt.Printf("⚠️  行 %d: 無効なパラメータ名をスキップ: %s\n", lineNum, line)
			continue
		}

		paramNames = append(paramNames, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("ファイルの読み込みエラー: %w", err)
	}

	return paramNames, nil
}

// isValidParameterName はパラメータ名の妥当性をチェック
func isValidParameterName(name string) bool {
	// パラメータ名は / で始まる必要がある
	if !strings.HasPrefix(name, "/") {
		return false
	}

	// 空白が含まれていないことを確認
	if strings.Contains(name, " ") || strings.Contains(name, "\t") {
		return false
	}

	// 最低限の長さチェック（/のみは無効）
	if len(name) < 2 {
		return false
	}

	return true
}

// deleteParameter は単一のパラメータをParameter Storeから削除する
func deleteParameter(client *ssm.Client, name string) error {
	input := &ssm.DeleteParameterInput{
		Name: &name,
	}

	_, err := client.DeleteParameter(context.Background(), input)
	return err
}
