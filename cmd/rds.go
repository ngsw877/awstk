package cmd

import (
	"awstk/internal/aws"
	"awstk/internal/service"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/spf13/cobra"
)

var (
	rdsInstanceId string
	rdsStackName  string
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
		instanceId, err := resolveRdsInstance()
		if err != nil {
			return err
		}

		rdsClient, err := aws.NewClient[*rds.Client](awsCtx)
		if err != nil {
			return fmt.Errorf("AWS設定の読み込みエラー: %w", err)
		}

		fmt.Printf("🚀 RDSインスタンス (%s) を起動します...\n", instanceId)
		err = service.StartRdsInstance(rdsClient, instanceId)
		if err != nil {
			return fmt.Errorf("❌ RDSインスタンス起動エラー: %w", err)
		}

		fmt.Printf("✅ RDSインスタンス (%s) の起動を開始しました\n", instanceId)
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
		instanceId, err := resolveRdsInstance()
		if err != nil {
			return err
		}

		rdsClient, err := aws.NewClient[*rds.Client](awsCtx)
		if err != nil {
			return fmt.Errorf("AWS設定の読み込みエラー: %w", err)
		}

		fmt.Printf("🛑 RDSインスタンス (%s) を停止します...\n", instanceId)
		err = service.StopRdsInstance(rdsClient, instanceId)
		if err != nil {
			return fmt.Errorf("❌ RDSインスタンス停止エラー: %w", err)
		}

		fmt.Printf("✅ RDSインスタンス (%s) の停止を開始しました\n", instanceId)
		return nil
	},
	SilenceUsage: true,
}

// resolveRdsInstance はRDSインスタンスIDを解決する（直接指定またはスタックから取得）
func resolveRdsInstance() (string, error) {
	if rdsInstanceId != "" {
		return rdsInstanceId, nil
	}

	if rdsStackName != "" {
		cfnClient, err := aws.NewClient[*cloudformation.Client](awsCtx)
		if err != nil {
			return "", fmt.Errorf("CloudFormationクライアント作成エラー: %w", err)
		}

		instanceId, err := service.GetRdsFromStack(cfnClient, rdsStackName)
		if err != nil {
			return "", fmt.Errorf("スタックからRDSインスタンス取得エラー: %w", err)
		}
		return instanceId, nil
	}

	return "", fmt.Errorf("RDSインスタンスID (-i) またはスタック名 (-S) を指定してください")
}

func init() {
	RootCmd.AddCommand(RdsCmd)
	RdsCmd.AddCommand(rdsStartCmd)
	RdsCmd.AddCommand(rdsStopCmd)

	// 共通フラグをRdsCmd（親コマンド）に定義
	RdsCmd.PersistentFlags().StringVarP(&rdsInstanceId, "instance", "i", "", "RDSインスタンス名")
	RdsCmd.PersistentFlags().StringVarP(&rdsStackName, "stack", "S", "", "CloudFormationスタック名")
}
