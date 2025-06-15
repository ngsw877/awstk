package cmd

import (
	"awstk/internal/aws"
	"awstk/internal/service"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/spf13/cobra"
)

// RdsCmd represents the rds command
var RdsCmd = &cobra.Command{
	Use:   "rds",
	Short: "RDSリソース操作コマンド",
	Long:  `RDSインスタンスを操作するためのコマンド群です。`,
}

var rdsStartCmd = &cobra.Command{
	Use:   "start",
	Short: "RDSインスタンスを起動するコマンド",
	Long: `RDSインスタンスを起動します。
CloudFormationスタック名を指定するか、インスタンス名を直接指定することができます。

例:
  ` + AppName + ` rds start -P my-profile -S my-stack
  ` + AppName + ` rds start -P my-profile -i my-instance`,
	RunE: func(cmd *cobra.Command, args []string) error {
		rdsInstanceId, _ := cmd.Flags().GetString("instance")
		stackName, _ := cmd.Flags().GetString("stack")
		var err error

		if rdsInstanceId == "" && stackName != "" {
			// スタックからRDSインスタンス名を取得
			rdsInstanceId, err = service.GetRdsFromStack(awsCtx, stackName)
			if err != nil {
				return fmt.Errorf("❌ スタックからRDSインスタンス取得エラー: %w", err)
			}
		}

		if rdsInstanceId == "" {
			return fmt.Errorf("❌ エラー: RDSインスタンスID (-i) を指定してください")
		}

		rdsClient, err := aws.NewClient[*rds.Client](awsCtx)
		if err != nil {
			return fmt.Errorf("AWS設定の読み込みエラー: %w", err)
		}

		fmt.Printf("🚀 RDSインスタンス (%s) を起動します...\n", rdsInstanceId)
		err = service.StartRdsInstance(rdsClient, rdsInstanceId)
		if err != nil {
			return fmt.Errorf("❌ RDSインスタンス起動エラー: %w", err)
		}

		fmt.Printf("✅ RDSインスタンス (%s) の起動を開始しました\n", rdsInstanceId)
		return nil
	},
	SilenceUsage: true,
}

var rdsStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "RDSインスタンスを停止するコマンド",
	Long: `RDSインスタンスを停止します。
CloudFormationスタック名を指定するか、インスタンス名を直接指定することができます。

例:
  ` + AppName + ` rds stop -P my-profile -S my-stack
  ` + AppName + ` rds stop -P my-profile -i my-instance`,
	RunE: func(cmd *cobra.Command, args []string) error {
		rdsInstanceId, _ := cmd.Flags().GetString("instance")
		stackName, _ := cmd.Flags().GetString("stack")
		var err error

		if rdsInstanceId == "" && stackName != "" {
			// スタックからRDSインスタンス名を取得
			rdsInstanceId, err = service.GetRdsFromStack(awsCtx, stackName)
			if err != nil {
				return fmt.Errorf("❌ スタックからRDSインスタンス取得エラー: %w", err)
			}
		}

		if rdsInstanceId == "" {
			return fmt.Errorf("❌ エラー: RDSインスタンスID (-i) を指定してください")
		}

		rdsClient, err := aws.NewClient[*rds.Client](awsCtx)
		if err != nil {
			return fmt.Errorf("AWS設定の読み込みエラー: %w", err)
		}

		fmt.Printf("🛑 RDSインスタンス (%s) を停止します...\n", rdsInstanceId)
		err = service.StopRdsInstance(rdsClient, rdsInstanceId)
		if err != nil {
			return fmt.Errorf("❌ RDSインスタンス停止エラー: %w", err)
		}

		fmt.Printf("✅ RDSインスタンス (%s) の停止を開始しました\n", rdsInstanceId)
		return nil
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(RdsCmd)
	RdsCmd.AddCommand(rdsStartCmd)
	RdsCmd.AddCommand(rdsStopCmd)

	// フラグの追加
	rdsStartCmd.Flags().StringP("instance", "i", "", "RDSインスタンス名")
	rdsStartCmd.Flags().StringP("stack", "S", "", "CloudFormationスタック名")
	rdsStopCmd.Flags().StringP("instance", "i", "", "RDSインスタンス名")
	rdsStopCmd.Flags().StringP("stack", "S", "", "CloudFormationスタック名")
}
