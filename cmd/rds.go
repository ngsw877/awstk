package cmd

import (
	"awsfunc/internal"
	"fmt"
	"github.com/spf13/cobra"
)

var (
	rdsInstanceId string
)

var rdsCmd = &cobra.Command{
	Use:   "rds",
	Short: "RDS関連の操作を行うコマンド群",
	Long:  "AWS RDSインスタンスの操作を行うCLIコマンド群です。",
}

var rdsStartInstanceCmd = &cobra.Command{
	Use:   "start",
	Short: "RDSインスタンスを起動する",
	Long: `指定したRDSインスタンスを起動します。

例:
  awsfunc rds start -d <rds-instance-identifier> [-P <aws-profile>]
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if rdsInstanceId == "" {
			return fmt.Errorf("❌ RDSインスタンス識別子は必須です")
		}
		fmt.Printf("RDSインスタンス (%s) を起動します...\n", rdsInstanceId)

		err := internal.StartRdsInstance(rdsInstanceId, region, profile)
		if err != nil {
			fmt.Printf("❌ RDSインスタンスの起動に失敗しました。")
			return err
		}

		fmt.Println("✅ RDSインスタンスの起動を開始しました。")
		return nil
	},
	SilenceUsage: true,
}

var rdsStopInstanceCmd = &cobra.Command{
	Use:   "stop",
	Short: "RDSインスタンスを停止する",
	Long: `指定したRDSインスタンスを停止します。

例:
  awsfunc rds stop -d <rds-instance-identifier> [-P <aws-profile>]
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if rdsInstanceId == "" {
			return fmt.Errorf("❌ RDSインスタンス識別子は必須です")
		}
		fmt.Printf("RDSインスタンス (%s) を停止します...\n", rdsInstanceId)

		err := internal.StopRdsInstance(rdsInstanceId, region, profile)
		if err != nil {
			fmt.Printf("❌ RDSインスタンスの停止に失敗しました。")
			return err
		}

		fmt.Println("✅ RDSインスタンスの停止を開始しました。")
		return nil
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(rdsCmd)
	rdsCmd.AddCommand(rdsStartInstanceCmd)
	rdsCmd.AddCommand(rdsStopInstanceCmd)
	rdsCmd.PersistentFlags().StringVarP(&rdsInstanceId, "db-instance-identifier", "d", "", "RDSインスタンス識別子（必須）")
}
