package service

import (
	"os"
	"os/exec"

	"awstk/internal/aws"
)

// StartSsmSession 指定したEC2インスタンスIDにSSMセッションで接続する
func StartSsmSession(awsCtx aws.Context, instanceId string) error {
	// AWS CLIのssm start-sessionコマンドを呼び出す
	args := []string{
		"ssm", "start-session",
		"--target", instanceId,
		"--region", awsCtx.Region,
	}
	if awsCtx.Profile != "" {
		args = append(args, "--profile", awsCtx.Profile)
	}

	// コマンドを実行
	cmd := exec.Command("aws", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
