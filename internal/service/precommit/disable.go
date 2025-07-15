package precommit

import (
	"awstk/internal/cli"
	"fmt"
)

// Disable はpre-commitフックを無効化する
func Disable() error {
	// Git hooks の参照先設定を削除
	if err := cli.UnsetGitConfig("core.hooksPath"); err != nil {
		// エラーが発生しても、既に設定がない場合は問題ないので続行
		fmt.Println("ℹ️  No custom hooksPath was set")
	}

	fmt.Println("✅ Pre-commit hook has been disabled")
	fmt.Println("")
	fmt.Println("To skip pre-commit hook temporarily, you can use:")
	fmt.Println("  git commit --no-verify")

	return nil
}
