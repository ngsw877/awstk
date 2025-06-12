package cmd

import (
	"awstk/internal"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/spf13/cobra"
)

var (
	ec2InstanceId string
)

var Ec2Cmd = &cobra.Command{
	Use:   "ec2",
	Short: "EC2リソース操作コマンド",
	Long:  `EC2リソースを操作するためのコマンド群です。`,
}

var ec2StartCmd = &cobra.Command{
	Use:   "start",
	Short: "EC2インスタンスを起動するコマンド",
	Long: `EC2インスタンスを起動するコマンドです。
インスタンスIDを直接指定して操作します。

例:
  ` + AppName + ` ec2 start -P my-profile -i i-1234567890abcdef0`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if ec2InstanceId == "" {
			cmd.Help()
			return fmt.Errorf("❌ エラー: EC2インスタンスID (-i) が必須です")
		}

		awsCtx := getAwsContext()
		// AWS設定を読み込んでEC2クライアントを作成
		cfg, err := internal.LoadAwsConfig(awsCtx)
		if err != nil {
			return fmt.Errorf("AWS設定の読み込みエラー: %w", err)
		}
		ec2Client := ec2.NewFromConfig(cfg)

		fmt.Printf("🚀 EC2インスタンス '%s' を起動します...\n", ec2InstanceId)
		err = internal.StartEc2Instance(ec2Client, ec2InstanceId)
		if err != nil {
			return fmt.Errorf("❌ エラー: %w", err)
		}

		fmt.Printf("✅ EC2インスタンス '%s' の起動を開始しました\n", ec2InstanceId)
		return nil
	},
	SilenceUsage: true,
}

var ec2StopCmd = &cobra.Command{
	Use:   "stop",
	Short: "EC2インスタンスを停止するコマンド",
	Long: `EC2インスタンスを停止するコマンドです。
インスタンスIDを直接指定して操作します。

例:
  ` + AppName + ` ec2 stop -P my-profile -i i-1234567890abcdef0`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if ec2InstanceId == "" {
			cmd.Help()
			return fmt.Errorf("❌ エラー: EC2インスタンスID (-i) が必須です")
		}

		awsCtx := getAwsContext()
		// AWS設定を読み込んでEC2クライアントを作成
		cfg, err := internal.LoadAwsConfig(awsCtx)
		if err != nil {
			return fmt.Errorf("AWS設定の読み込みエラー: %w", err)
		}
		ec2Client := ec2.NewFromConfig(cfg)

		fmt.Printf("🛑 EC2インスタンス '%s' を停止します...\n", ec2InstanceId)
		err = internal.StopEc2Instance(ec2Client, ec2InstanceId)
		if err != nil {
			return fmt.Errorf("❌ エラー: %w", err)
		}

		fmt.Printf("✅ EC2インスタンス '%s' の停止を開始しました\n", ec2InstanceId)
		return nil
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(Ec2Cmd)
	Ec2Cmd.AddCommand(ec2StartCmd)
	Ec2Cmd.AddCommand(ec2StopCmd)

	// startコマンドのフラグを設定
	ec2StartCmd.Flags().StringVarP(&ec2InstanceId, "instance", "i", "", "EC2インスタンスID（必須）")

	// stopコマンドのフラグを設定
	ec2StopCmd.Flags().StringVarP(&ec2InstanceId, "instance", "i", "", "EC2インスタンスID（必須）")
}
