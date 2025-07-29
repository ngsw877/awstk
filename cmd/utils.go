package cmd

import (
	"fmt"
	"os"
)

// resolveStackName はコマンドライン引数または環境変数からスタック名を決定し、グローバル変数 stackName にセットする
func resolveStackName() {
	if stackName != "" {
		fmt.Println("🔍 -Sオプションで指定されたスタック名 '" + stackName + "' を使用します")
		return
	}
	envStack := os.Getenv("AWS_STACK_NAME")
	if envStack != "" {
		fmt.Println("🔍 環境変数 AWS_STACK_NAME の値 '" + envStack + "' を使用します")
		stackName = envStack
	}
	// どちらもなければstackNameは空のまま
}

// printAwsContext はAWSコンテキスト情報を表示する共通関数
func printAwsContext() {
	fmt.Printf("Profile: %s\n", profile)
	fmt.Printf("Region: %s\n", region)
}

// printAwsContextWithInfo はAWSコンテキスト情報と追加情報を表示する共通関数
func printAwsContextWithInfo(infoLabel string, infoValue string) {
	printAwsContext()
	fmt.Printf("%s: %s\n", infoLabel, infoValue)
}

// ValidateStackSelection は位置引数とオプションの排他チェックを行います
func ValidateStackSelection(args []string, hasOptions bool) error {
	hasArgs := len(args) > 0

	if !hasArgs && !hasOptions {
		return fmt.Errorf("❌ エラー: スタック名またはオプションを指定してください")
	}

	if hasArgs && hasOptions {
		return fmt.Errorf("❌ エラー: スタック名とオプションは同時に指定できません")
	}

	return nil
}

// ValidateExclusiveOptions は複数のオプションの排他チェックを行います
// requireOne: true の場合、いずれか1つの指定が必須
// exclusive: true の場合、同時指定不可
func ValidateExclusiveOptions(requireOne, exclusive bool, options ...bool) error {
	count := 0
	for _, opt := range options {
		if opt {
			count++
		}
	}

	if requireOne && count == 0 {
		return fmt.Errorf("❌ エラー: いずれかのオプションを指定してください")
	}

	if exclusive && count > 1 {
		return fmt.Errorf("❌ エラー: オプションは同時に指定できません")
	}

	return nil
}
