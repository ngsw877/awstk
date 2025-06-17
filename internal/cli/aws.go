package cli

import (
	"os"
	"os/exec"
)

// ExecuteAwsCommand はAWS CLIコマンドを実行する共通関数
func ExecuteAwsCommand(args []string) error {
	cmd := exec.Command("aws", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
