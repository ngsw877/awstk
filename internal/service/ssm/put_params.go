package ssm

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

// PutParametersFromFile はファイルからパラメータを読み込んでParameter Storeに登録する
func PutParametersFromFile(ssmClient *ssm.Client, opts PutParamsOptions) error {
	// ファイルの存在確認
	if _, err := os.Stat(opts.FilePath); os.IsNotExist(err) {
		return fmt.Errorf("ファイルが見つかりません: %s", opts.FilePath)
	}

	// パラメータの読み込み
	params, err := loadParametersFromFile(opts.FilePath)
	if err != nil {
		return fmt.Errorf("ファイルの読み込みに失敗しました: %w", err)
	}

	if len(params) == 0 {
		return fmt.Errorf("登録するパラメータが見つかりません")
	}

	// プレフィックスの適用
	if opts.Prefix != "" {
		for i := range params {
			params[i].Name = normalizeParameterName(opts.Prefix, params[i].Name)
		}
	}

	// ドライランの場合は内容を表示して終了
	if opts.DryRun {
		fmt.Println("📋 以下のパラメータが登録されます:")
		fmt.Println(strings.Repeat("-", 80))
		for _, param := range params {
			fmt.Printf("Name: %s\n", param.Name)
			fmt.Printf("Type: %s\n", param.Type)
			if param.Type != "SecureString" {
				fmt.Printf("Value: %s\n", param.Value)
			} else {
				fmt.Printf("Value: ****** (SecureString)\n")
			}
			if param.Description != "" {
				fmt.Printf("Description: %s\n", param.Description)
			}
			fmt.Println(strings.Repeat("-", 80))
		}
		return nil
	}

	// パラメータの登録
	var successCount, failCount int
	for _, param := range params {
		err := putParameter(ssmClient, param)
		if err != nil {
			fmt.Printf("❌ %s の登録に失敗しました: %v\n", param.Name, err)
			failCount++
		} else {
			fmt.Printf("✅ %s を登録しました\n", param.Name)
			successCount++
		}
	}

	fmt.Printf("\n📊 登録結果: 成功 %d / 失敗 %d / 合計 %d\n", successCount, failCount, len(params))

	if failCount > 0 {
		return fmt.Errorf("%d 件のパラメータ登録に失敗しました", failCount)
	}

	return nil
}

// loadParametersFromFile はファイルからパラメータを読み込む
func loadParametersFromFile(filePath string) ([]parameter, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".json":
		return loadParametersFromJSON(filePath)
	case ".csv":
		return loadParametersFromCSV(filePath)
	default:
		return nil, fmt.Errorf("サポートされていないファイル形式: %s", ext)
	}
}

// loadParametersFromJSON はJSONファイルからパラメータを読み込む
func loadParametersFromJSON(filePath string) ([]parameter, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("ファイルを開けません: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("⚠️  ファイルのクローズに失敗: %v\n", err)
		}
	}()

	var paramFile parametersFile
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&paramFile); err != nil {
		return nil, fmt.Errorf("JSONの解析に失敗しました: %w", err)
	}

	// バリデーション
	for i, param := range paramFile.Parameters {
		if err := validateParameter(param); err != nil {
			return nil, fmt.Errorf("パラメータ[%d]のバリデーションエラー: %w", i, err)
		}
	}

	return paramFile.Parameters, nil
}

// loadParametersFromCSV はCSVファイルからパラメータを読み込む
func loadParametersFromCSV(filePath string) ([]parameter, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("ファイルを開けません: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("⚠️  ファイルのクローズに失敗: %v\n", err)
		}
	}()

	reader := csv.NewReader(file)

	// ヘッダー行を読み込む
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("CSVヘッダーの読み込みに失敗しました: %w", err)
	}

	// ヘッダーの検証
	expectedHeaders := []string{"name", "value", "type", "description"}
	if len(headers) < 3 {
		return nil, fmt.Errorf("CSVヘッダーが不正です。最低限 name, value, type が必要です")
	}
	for i, expected := range expectedHeaders[:3] {
		if i < len(headers) && strings.ToLower(headers[i]) != expected {
			return nil, fmt.Errorf("CSVヘッダーが不正です。期待: %s, 実際: %s", expected, headers[i])
		}
	}

	var params []parameter
	lineNum := 1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("CSV行 %d の読み込みに失敗しました: %w", lineNum+1, err)
		}
		lineNum++

		if len(record) < 3 {
			return nil, fmt.Errorf("CSV行 %d のカラム数が不足しています", lineNum)
		}

		param := parameter{
			Name:  strings.TrimSpace(record[0]),
			Value: strings.TrimSpace(record[1]),
			Type:  strings.TrimSpace(record[2]),
		}

		// descriptionカラムがある場合
		if len(record) > 3 {
			param.Description = strings.TrimSpace(record[3])
		}

		// バリデーション
		if err := validateParameter(param); err != nil {
			return nil, fmt.Errorf("CSV行 %d のバリデーションエラー: %w", lineNum, err)
		}

		params = append(params, param)
	}

	return params, nil
}

// validateParameter はパラメータのバリデーションを行う
func validateParameter(param parameter) error {
	if param.Name == "" {
		return fmt.Errorf("nameが空です")
	}
	if param.Value == "" {
		return fmt.Errorf("valueが空です")
	}
	if param.Type == "" {
		return fmt.Errorf("typeが空です")
	}

	// 型の検証
	validTypes := []string{"String", "SecureString", "StringList"}
	isValidType := false
	for _, vt := range validTypes {
		if param.Type == vt {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return fmt.Errorf("無効なtype: %s (有効な値: %s)", param.Type, strings.Join(validTypes, ", "))
	}

	// パラメータ名の検証
	if !strings.HasPrefix(param.Name, "/") {
		return fmt.Errorf("パラメータ名は / で始まる必要があります: %s", param.Name)
	}

	return nil
}

// putParameter は単一のパラメータをParameter Storeに登録する
func putParameter(client *ssm.Client, param parameter) error {
	input := &ssm.PutParameterInput{
		Name:      aws.String(param.Name),
		Value:     aws.String(param.Value),
		Type:      types.ParameterType(param.Type),
		Overwrite: aws.Bool(true),
	}

	if param.Description != "" {
		input.Description = aws.String(param.Description)
	}

	_, err := client.PutParameter(context.Background(), input)
	return err
}
