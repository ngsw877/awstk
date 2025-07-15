package cli

import (
	"os/exec"
)

// ExecuteGitCommand はGitコマンドを実行する共通関数
func ExecuteGitCommand(args []string) error {
	cmd := exec.Command("git", args...)
	return cmd.Run()
}

// SetGitConfig はgit configを設定する
func SetGitConfig(key, value string) error {
	return ExecuteGitCommand([]string{"config", "--local", key, value})
}

// UnsetGitConfig はgit configを削除する
func UnsetGitConfig(key string) error {
	return ExecuteGitCommand([]string{"config", "--local", "--unset", key})
}

// GetGitConfig はgit configの値を取得する
func GetGitConfig(key string) (string, error) {
	cmd := exec.Command("git", "config", "--local", "--get", key)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
