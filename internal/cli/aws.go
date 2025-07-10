package cli

import (
	"awstk/internal/aws"
	"os"
	"os/exec"
)

// ExecuteAwsCommand はAWS CLIコマンドを実行する共通関数
func ExecuteAwsCommand(ctx aws.Context, args []string) error {
	// AWS CLIコマンドを構築
	// プロファイルが指定されている場合、引数に追加
	if ctx.Profile != "" {
		args = append(args, "--profile", ctx.Profile)
	}

	// リージョンが指定されている場合、引数に追加
	if ctx.Region != "" {
		args = append(args, "--region", ctx.Region)
	}

	cmd := exec.Command("aws", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
