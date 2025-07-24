package ssm

import (
	"awstk/internal/aws"
	"awstk/internal/cli"
	ec2svc "awstk/internal/service/ec2"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// StartSsmSession 指定したEC2インスタンスIDにSSMセッションで接続する
func StartSsmSession(awsCtx aws.Context, opts SessionOptions) error {
	// AWS CLIのssm start-sessionコマンドを呼び出す
	args := []string{
		"ssm", "start-session",
		"--target", opts.InstanceId,
	}

	// cli層の共通関数を使用してコマンドを実行
	return cli.ExecuteAwsCommand(awsCtx, args)
}

// SelectAndStartSession はインスタンスを選択してSSMセッションを開始する
func SelectAndStartSession(awsCtx aws.Context, ec2Client *ec2.Client, instanceId string) error {
	// インスタンスIDが指定されていない場合は、インタラクティブに選択
	if instanceId == "" {
		fmt.Println("🖥️  利用可能なEC2インスタンスから選択してください:")

		selectedInstanceId, err := ec2svc.SelectInstanceInteractively(ec2Client)
		if err != nil {
			return fmt.Errorf("❌ インスタンス選択でエラー: %w", err)
		}
		instanceId = selectedInstanceId
	}

	fmt.Printf("EC2インスタンス (%s) にSSMで接続します...\n", instanceId)

	opts := SessionOptions{
		InstanceId: instanceId,
	}

	err := StartSsmSession(awsCtx, opts)
	if err != nil {
		return fmt.Errorf("❌ SSMセッションの開始に失敗しました: %w", err)
	}

	fmt.Println("✅ SSMセッションを開始しました。")
	return nil
}
