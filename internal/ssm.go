package internal

import (
	"os"
	"os/exec"
)

// StartSsmSession 指定したEC2インスタンスIDにSSMセッションで接続する
func StartSsmSession(awsCtx AwsContext, instanceId string) error {
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
