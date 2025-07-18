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
	fmt.Printf("Profile: %s\n", awsCtx.Profile)
	fmt.Printf("Region: %s\n", awsCtx.Region)
}

// printAwsContextWithInfo はAWSコンテキスト情報と追加情報を表示する共通関数
func printAwsContextWithInfo(infoLabel string, infoValue string) {
	printAwsContext()
	fmt.Printf("%s: %s\n", infoLabel, infoValue)
}
