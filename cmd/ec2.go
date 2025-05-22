package cmd

import (
	"awsfunc/internal"
	"fmt"
	"github.com/spf13/cobra"
)

var (
	ec2InstanceId string
)

var ec2Cmd = &cobra.Command{
	Use:   "ec2",
	Short: "EC2関連の操作を行うコマンド群",
	Long:  "AWS EC2インスタンスの操作を行うCLIコマンド群です。",
}

var ec2StartInstanceCmd = &cobra.Command{
	Use:   "start",
	Short: "EC2インスタンスを起動する",
	Long: `指定したEC2インスタンスを起動します。

例:
  awsfunc ec2 start -i <ec2-instance-id> [-P <aws-profile>]
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if ec2InstanceId == "" {
			return fmt.Errorf("❌ EC2インスタンスIDは必須です")
		}
		fmt.Printf("EC2インスタンス (%s) を起動します...\n", ec2InstanceId)

		awsCtx := getAwsContext()
		err := internal.StartEc2Instance(awsCtx, ec2InstanceId)
		if err != nil {
			fmt.Printf("❌ EC2インスタンスの起動に失敗しました。")
			return err
		}

		fmt.Println("✅ EC2インスタンスの起動を開始しました。")
		return nil
	},
	SilenceUsage: true,
}

var ec2StopInstanceCmd = &cobra.Command{
	Use:   "stop",
	Short: "EC2インスタンスを停止する",
	Long: `指定したEC2インスタンスを停止します。

例:
  awsfunc ec2 stop -i <ec2-instance-id> [-P <aws-profile>]
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if ec2InstanceId == "" {
			return fmt.Errorf("❌ EC2インスタンスIDは必須です")
		}
		fmt.Printf("EC2インスタンス (%s) を停止します...\n", ec2InstanceId)

		awsCtx := getAwsContext()
		err := internal.StopEc2Instance(awsCtx, ec2InstanceId)
		if err != nil {
			fmt.Printf("❌ EC2インスタンスの停止に失敗しました。")
			return err
		}

		fmt.Println("✅ EC2インスタンスの停止を開始しました。")
		return nil
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(ec2Cmd)
	ec2Cmd.AddCommand(ec2StartInstanceCmd)
	ec2Cmd.AddCommand(ec2StopInstanceCmd)
	ec2Cmd.PersistentFlags().StringVarP(&ec2InstanceId, "instance-id", "i", "", "EC2インスタンスID（必須）")
}
