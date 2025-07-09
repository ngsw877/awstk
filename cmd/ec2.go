package cmd

import (
	ec2svc "awstk/internal/service/ec2"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/spf13/cobra"
)

var (
	ec2InstanceId string
	ec2Client     *ec2.Client
)

// Ec2Cmd represents the ec2 command
var Ec2Cmd = &cobra.Command{
	Use:   "ec2",
	Short: "EC2インスタンス操作コマンド",
	Long:  `EC2インスタンスを操作するためのコマンド群です。`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// 親のPersistentPreRunEを実行（awsCtx設定とAWS設定読み込み）
		if err := RootCmd.PersistentPreRunE(cmd, args); err != nil {
			return err
		}

		// クライアント生成
		ec2Client = ec2.NewFromConfig(awsCfg)
		cfnClient = cloudformation.NewFromConfig(awsCfg)

		return nil
	},
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

		fmt.Printf("🚀 EC2インスタンス (%s) を起動します...\n", ec2InstanceId)
		err := ec2svc.StartEc2Instance(ec2Client, ec2InstanceId)
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

		fmt.Printf("🛑 EC2インスタンス (%s) を停止します...\n", ec2InstanceId)
		err := ec2svc.StopEc2Instance(ec2Client, ec2InstanceId)
		if err != nil {
			return fmt.Errorf("❌ EC2インスタンス停止エラー: %w", err)
		}

		fmt.Printf("✅ EC2インスタンス (%s) の停止を開始しました\n", ec2InstanceId)
		return nil
	},
	SilenceUsage: true,
}

var ec2LsCmd = &cobra.Command{
	Use:   "ls",
	Short: "EC2インスタンス一覧を表示するコマンド",
	Long:  `EC2インスタンス一覧を表示します。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			instances []ec2svc.Ec2Instance
			err       error
		)

		if stackName != "" {
			instances, err = ec2svc.ListEc2InstancesFromStack(ec2Client, cfnClient, stackName)
			if err != nil {
				return fmt.Errorf("❌ CloudFormationスタックからインスタンスIDの取得に失敗: %w", err)
			}
		} else {
			instances, err = ec2svc.ListEc2Instances(ec2Client)
			if err != nil {
				return fmt.Errorf("❌ EC2インスタンス一覧取得でエラー: %w", err)
			}
		}

		if len(instances) == 0 {
			fmt.Println("EC2インスタンスが見つかりませんでした")
			return nil
		}

		fmt.Printf("EC2インスタンス一覧: (全%d件)\n", len(instances))
		for i, ins := range instances {
			fmt.Printf("  %3d. %s (%s) [%s]\n", i+1, ins.InstanceId, ins.InstanceName, ins.State)
		}
		return nil
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(Ec2Cmd)
	Ec2Cmd.AddCommand(ec2StartCmd)
	Ec2Cmd.AddCommand(ec2StopCmd)
	Ec2Cmd.AddCommand(ec2LsCmd)

	// フラグの追加
	ec2StartCmd.Flags().StringVarP(&ec2InstanceId, "instance", "i", "", "EC2インスタンスID")
	ec2StopCmd.Flags().StringVarP(&ec2InstanceId, "instance", "i", "", "EC2インスタンスID")
	ec2LsCmd.Flags().StringVarP(&stackName, "stack", "S", "", "CloudFormationスタック名")
}
