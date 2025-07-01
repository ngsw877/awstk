package ssm

import (
	"awstk/internal/cli"
)

// StartSsmSession 指定したEC2インスタンスIDにSSMセッションで接続する
func StartSsmSession(opts SessionOptions) error {
	// AWS CLIのssm start-sessionコマンドを呼び出す
	args := []string{
		"ssm", "start-session",
		"--target", opts.InstanceId,
		"--region", opts.Region,
	}
	if opts.Profile != "" {
		args = append(args, "--profile", opts.Profile)
	}

	// cli層の共通関数を使用してコマンドを実行
	return cli.ExecuteAwsCommand(args)
}
