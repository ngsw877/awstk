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
