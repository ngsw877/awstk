package precommit

import (
	"awstk/internal/cli"
	"fmt"
	"os"
)

// Enable はpre-commitフックを有効化する
func Enable() error {
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
