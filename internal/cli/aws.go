package cli

import (
	"awstk/internal/aws"
	"bytes"
	"io"
	"os"
	"os/exec"
)

// AwsCommandResult はAWS CLIコマンドの実行結果
type AwsCommandResult struct {
	Stdout string
	Stderr string
}

// buildAwsCliArgs はAWS CLIコマンドの引数を構築する（認証情報を追加）
func buildAwsCliArgs(ctx aws.Context, args []string) []string {
	// プロファイルが指定されている場合、引数に追加
	if ctx.Profile != "" {
		args = append(args, "--profile", ctx.Profile)
	}

	// リージョンが指定されている場合、引数に追加
	if ctx.Region != "" {
		args = append(args, "--region", ctx.Region)
	}

	return args
}

// ExecuteAwsCommand はAWS CLIコマンドを実行する共通関数
func ExecuteAwsCommand(ctx aws.Context, args []string) error {
	args = buildAwsCliArgs(ctx, args)

	cmd := exec.Command("aws", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// ExecuteAwsCommandWithCapture はAWS CLIコマンドを実行し、出力をキャプチャする
func ExecuteAwsCommandWithCapture(ctx aws.Context, args []string) (*AwsCommandResult, error) {
	args = buildAwsCliArgs(ctx, args)

	cmd := exec.Command("aws", args...)
	cmd.Stdin = os.Stdin

	// 出力をキャプチャしつつ、リアルタイムで表示
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	err := cmd.Run()

	result := &AwsCommandResult{
		Stdout: stdoutBuf.String(),
		Stderr: stderrBuf.String(),
	}

	return result, err
}
