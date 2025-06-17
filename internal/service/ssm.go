package service

import (
	"awstk/internal/cli"
)

// SsmSessionOptions はSSMセッション開始のパラメータを格納する構造体
type SsmSessionOptions struct {
	Region     string
	Profile    string
	InstanceId string
}

// StartSsmSession 指定したEC2インスタンスIDにSSMセッションで接続する
func StartSsmSession(opts SsmSessionOptions) error {
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
