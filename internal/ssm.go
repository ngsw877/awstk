package internal

import (
	"os"
	"os/exec"
)

// StartSsmSession 指定したEC2インスタンスIDにSSMセッションで接続する
func StartSsmSession(instanceId, region, profile string) error {

	// AWS CLIのssm start-sessionコマンドを呼び出す
	args := []string{
		"aws", "ssm", "start-session",
		"--target", instanceId,
		"--region", region,
	}
	if profile != "" {
		args = append(args, "--profile", profile)
	}

	// コマンドを実行
	cmd := exec.Command("aws", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
