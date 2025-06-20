package cmd

import (
	"awstk/internal/aws"
	ec2svc "awstk/internal/service/ec2"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/spf13/cobra"
)

var (
	ec2InstanceId string
)

// Ec2Cmd represents the ec2 command
var Ec2Cmd = &cobra.Command{
	Use:   "ec2",
	Short: "EC2インスタンス操作コマンド",
	Long:  `EC2インスタンスを操作するためのコマンド群です。`,
}

var ec2StartCmd = &cobra.Command{
	Use:   "start",
	Short: "EC2インスタンスを起動するコマンド",
	Long: `EC2インスタンスを起動します。
インスタンスIDを直接指定することができます。

例:
  ` + AppName + ` ec2 start -i i-1234567890abcdef0`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if ec2InstanceId == "" {
			return fmt.Errorf("❌ エラー: インスタンスID (-i) を指定してください")
		}

		cfg, err := aws.LoadAwsConfig(awsCtx)
		if err != nil {
			return fmt.Errorf("AWS設定の読み込みエラー: %w", err)
		}
		ec2Client := ec2.NewFromConfig(cfg)

		fmt.Printf("🚀 EC2インスタンス (%s) を起動します...\n", ec2InstanceId)
		err = ec2svc.StartEc2Instance(ec2Client, ec2InstanceId)
		if err != nil {
			return fmt.Errorf("❌ EC2インスタンス起動エラー: %w", err)
		}

		fmt.Printf("✅ EC2インスタンス (%s) の起動を開始しました\n", ec2InstanceId)
		return nil
	},
	SilenceUsage: true,
}

var ec2StopCmd = &cobra.Command{
	Use:   "stop",
	Short: "EC2インスタンスを停止するコマンド",
	Long: `EC2インスタンスを停止します。
インスタンスIDを直接指定することができます。

例:
  ` + AppName + ` ec2 stop -i i-1234567890abcdef0`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if ec2InstanceId == "" {
			return fmt.Errorf("❌ エラー: インスタンスID (-i) を指定してください")
		}

		cfg, err := aws.LoadAwsConfig(awsCtx)
		if err != nil {
			return fmt.Errorf("AWS設定の読み込みエラー: %w", err)
		}
		ec2Client := ec2.NewFromConfig(cfg)

		fmt.Printf("🛑 EC2インスタンス (%s) を停止します...\n", ec2InstanceId)
		err = ec2svc.StopEc2Instance(ec2Client, ec2InstanceId)
		if err != nil {
			return fmt.Errorf("❌ EC2インスタンス停止エラー: %w", err)
		}

		fmt.Printf("✅ EC2インスタンス (%s) の停止を開始しました\n", ec2InstanceId)
		return nil
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(Ec2Cmd)
	Ec2Cmd.AddCommand(ec2StartCmd)
	Ec2Cmd.AddCommand(ec2StopCmd)

	// フラグの追加
	ec2StartCmd.Flags().StringVarP(&ec2InstanceId, "instance", "i", "", "EC2インスタンスID")
	ec2StopCmd.Flags().StringVarP(&ec2InstanceId, "instance", "i", "", "EC2インスタンスID")
}
