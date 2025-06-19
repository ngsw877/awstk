package cmd

import (
	"awstk/internal/aws"
	ec2svc "awstk/internal/service/ec2"
	"awstk/internal/service/ssm"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
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
  ` + AppName + ` ssm session -i <ec2-instance-id> [-P <aws-profile>]
  ` + AppName + ` ssm session [-P <aws-profile>]  # インスタンス一覧から選択
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		awsCtx := aws.Context{Region: region, Profile: profile}

		// -iオプションが指定されていない場合、インスタンス一覧から選択
		if ssmInstanceId == "" {
			// インタラクティブモードでインスタンスを選択
			fmt.Println("🖥️  利用可能なEC2インスタンスから選択してください:")

			ec2Client, err := aws.NewClient[*ec2.Client](awsCtx)
			if err != nil {
				return fmt.Errorf("EC2クライアント作成エラー: %w", err)
			}

			selectedInstanceId, err := ec2svc.SelectInstanceInteractively(ec2Client)
			if err != nil {
				return fmt.Errorf("❌ インスタンス選択でエラー: %w", err)
			}
			ssmInstanceId = selectedInstanceId
		}

		fmt.Printf("EC2インスタンス (%s) にSSMで接続します...\n", ssmInstanceId)

		opts := ssm.SsmSessionOptions{
			Region:     awsCtx.Region,
			Profile:    awsCtx.Profile,
			InstanceId: ssmInstanceId,
		}

		err := ssm.StartSsmSession(opts)
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
