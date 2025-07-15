package precommit

import (
	"awstk/internal/cli"
	"fmt"
	"os"
	"strings"
)

// GetStatus はpre-commitフックの現在の状態を取得する
func GetStatus() (*Status, error) {
	status := &Status{}

	// Git hooks の参照先設定を取得
	hooksPath, err := cli.GetGitConfig("core.hooksPath")
	if err != nil {
		// 設定がない場合はデフォルトの .git/hooks を使用
		status.Enabled = false
		status.HooksPath = ".git/hooks (default)"
		return status, nil
	}

	// 改行を削除
	hooksPath = strings.TrimSpace(hooksPath)
	status.HooksPath = hooksPath

	// .githooks を使用しているかチェック
	if hooksPath == ".githooks" {
		// pre-commit ファイルが存在し、実行可能かチェック
		fileInfo, err := os.Stat(".githooks/pre-commit")
		if err != nil {
			status.Enabled = false
			status.ErrorMessage = ".githooks/pre-commit not found"
			return status, nil
		}

		// 実行権限をチェック
		if fileInfo.Mode()&0111 == 0 {
			status.Enabled = false
			status.ErrorMessage = ".githooks/pre-commit is not executable"
			return status, nil
		}

		status.Enabled = true
	} else {
		status.Enabled = false
	}

	return status, nil
}

// ShowStatus はpre-commitフックの状態を表示する
func ShowStatus() error {
	status, err := GetStatus()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	if status.Enabled {
		fmt.Println("✅ Pre-commit hook is ENABLED")
		fmt.Printf("   Hooks path: %s\n", status.HooksPath)
		fmt.Println("")
		fmt.Println("The following pre-commit hooks are active:")
		fmt.Println("  - Cursor Rules → CLAUDE.md sync")
	} else {
		fmt.Println("❌ Pre-commit hook is DISABLED")
		fmt.Printf("   Hooks path: %s\n", status.HooksPath)
		if status.ErrorMessage != "" {
			fmt.Printf("   Issue: %s\n", status.ErrorMessage)
		}
		fmt.Println("")
		fmt.Println("To enable pre-commit hook, run:")
		fmt.Println("  awstk precommit enable")
	}

	return nil
}
