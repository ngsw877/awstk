package main

import (
	"awstk/internal/cli"
	"fmt"
	"os"
	"strings"
)

// status はpre-commitフックの状態を表す
type status struct {
	Enabled      bool   `json:"enabled"`
	HooksPath    string `json:"hooks_path,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run scripts/precommit.go [enable|disable|status]")
		os.Exit(1)
	}

	var err error
	switch os.Args[1] {
	case "enable":
		err = enable()
	case "disable":
		err = disable()
	case "status":
		err = showStatus()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		fmt.Println("Usage: go run scripts/precommit.go [enable|disable|status]")
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// enable はpre-commitフックを有効化する
func enable() error {
	// .githooks ディレクトリが存在するかチェック
	if _, err := os.Stat(".githooks"); os.IsNotExist(err) {
		return fmt.Errorf(".githooks directory not found")
	}

	// pre-commit ファイルが存在するかチェック
	if _, err := os.Stat(".githooks/pre-commit"); os.IsNotExist(err) {
		return fmt.Errorf(".githooks/pre-commit not found")
	}

	// Git hooks の参照先を .githooks に設定
	if err := cli.SetGitConfig("core.hooksPath", ".githooks"); err != nil {
		return fmt.Errorf("failed to set hooksPath: %w", err)
	}

	// pre-commit ファイルに実行権限を付与
	if err := os.Chmod(".githooks/pre-commit", 0755); err != nil {
		return fmt.Errorf("failed to set executable permission: %w", err)
	}

	fmt.Println("✅ Pre-commit hook has been enabled")
	fmt.Println("")
	fmt.Println("Now, when you commit changes to .cursor/rules/*.mdc files,")
	fmt.Println("CLAUDE.md will be automatically updated and included in the commit.")

	return nil
}

// disable はpre-commitフックを無効化する
func disable() error {
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

// getStatus はpre-commitフックの現在の状態を取得する
func getStatus() (*status, error) {
	status := &status{}

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

// showStatus はpre-commitフックの状態を表示する
func showStatus() error {
	status, err := getStatus()
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
		fmt.Println("  make precommit-enable")
	}

	return nil
}
