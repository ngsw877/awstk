package cmd

import (
	"awsfunc/internal"
	"fmt"

	"github.com/spf13/cobra"
)

var ssmInstanceId string

var ssmCmd = &cobra.Command{
	Use:   "ssm",
	Short: "SSM関連の操作を行うコマンド群",
	Long:  "AWS SSMセッションマネージャーを利用したEC2インスタンスへの接続などを行うCLIコマンド群です。",
}

var ssmSessionStartCmd = &cobra.Command{
	Use:   "session",
	Short: "EC2インスタンスにSSMで接続する",
	Long: `指定したEC2インスタンスIDにSSMセッションで接続します。

例:
  awsfunc ssm session -i <ec2-instance-id> [-P <aws-profile>]
  awsfunc ssm session [-P <aws-profile>]  # インスタンス一覧から選択
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		awsCtx := getAwsContext()

		// -iオプションが指定されていない場合、インスタンス一覧から選択
		if ssmInstanceId == "" {
			selectedInstanceId, err := internal.SelectInstanceInteractively(awsCtx)
			if err != nil {
				return err
			}
			ssmInstanceId = selectedInstanceId
		}

		fmt.Printf("EC2インスタンス (%s) にSSMで接続します...\n", ssmInstanceId)

		err := internal.StartSsmSession(awsCtx, ssmInstanceId)
		if err != nil {
			fmt.Printf("❌ SSMセッションの開始に失敗しました。")
			return err
		}

		fmt.Println("✅ SSMセッションを開始しました。")
		return nil
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(ssmCmd)
	ssmCmd.AddCommand(ssmSessionStartCmd)
	ssmCmd.PersistentFlags().StringVarP(&ssmInstanceId, "instance-id", "i", "", "EC2インスタンスID（省略時は一覧から選択）")
}
