package cmd

import (
	"awstk/internal/service/cfn"
	rdssvc "awstk/internal/service/rds"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/spf13/cobra"
)

var (
	rdsInstanceId string
)

// RdsCmd represents the rds command
var RdsCmd = &cobra.Command{
	Use:   "rds",
	Short: "RDSリソース操作コマンド",
	Long:  `RDSインスタンスを操作するためのコマンド群です。`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// 親のPersistentPreRunEを実行（awsCtx設定とAWS設定読み込み）
		if err := RootCmd.PersistentPreRunE(cmd, args); err != nil {
			return err
		}

		// RDS用クライアント生成
		rdsClient = rds.NewFromConfig(awsCfg)
		cfnClient = cloudformation.NewFromConfig(awsCfg)

		return nil
	},
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
		instanceName, err := resolveRdsInstanceName(cmd)
		if err != nil {
			return err
		}

		fmt.Printf("🚀 RDSインスタンス (%s) を起動します...\n", instanceName)
		err = rdssvc.StartRdsInstance(rdsClient, instanceName)
		if err != nil {
			return fmt.Errorf("❌ RDSインスタンス起動エラー: %w", err)
		}

		fmt.Printf("✅ RDSインスタンス (%s) の起動を開始しました\n", instanceName)
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
		instanceName, err := resolveRdsInstanceName(cmd)
		if err != nil {
			return err
		}

		fmt.Printf("🚀 RDSインスタンス (%s) を停止します...\n", instanceName)
		err = rdssvc.StopRdsInstance(rdsClient, instanceName)
		if err != nil {
			return fmt.Errorf("❌ RDSインスタンス停止エラー: %w", err)
		}

		fmt.Printf("✅ RDSインスタンス (%s) の停止を開始しました\n", instanceName)
		return nil
	},
	SilenceUsage: true,
}

var rdsLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "RDSインスタンス一覧を表示するコマンド",
	Long:  `RDSインスタンス一覧を表示します。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		resolveStackName()
		return rdssvc.ListRdsInstances(rdsClient, cfnClient, stackName)
	},
	SilenceUsage: true,
}

// resolveRdsInstanceName はRDSインスタンス名を解決する
func resolveRdsInstanceName(cmd *cobra.Command) (string, error) {
	resolveStackName()
	instanceName, _ := cmd.Flags().GetString("instance")

	// スタック名が指定されている場合
	if stackName != "" {
		return getRdsInstanceFromStack(stackName)
	}

	// インスタンス名が直接指定されている場合
	if instanceName != "" {
		return instanceName, nil
	}

	// どちらも指定されていない場合
	return "", fmt.Errorf("❌ エラー: RDSインスタンス名 (-i) またはスタック名 (-S) を指定してください")
}

// getRdsInstanceFromStack はCloudFormationスタックからRDSインスタンス名を取得する
func getRdsInstanceFromStack(stackName string) (string, error) {
	instanceName, err := cfn.GetRdsFromStack(cfnClient, stackName)
	if err != nil {
		return "", fmt.Errorf("❌ CloudFormationスタックからインスタンス名の取得に失敗: %w", err)
	}
	fmt.Printf("✅ CloudFormationスタック '%s' からRDSインスタンス '%s' を検出しました\n", stackName, instanceName)
	return instanceName, nil
}

func init() {
	RootCmd.AddCommand(RdsCmd)
	RdsCmd.AddCommand(rdsStartCmd)
	RdsCmd.AddCommand(rdsStopCmd)
	RdsCmd.AddCommand(rdsLsCmd)

	// 共通フラグをRdsCmd（親コマンド）に定義
	RdsCmd.PersistentFlags().StringVarP(&rdsInstanceId, "instance", "i", "", "RDSインスタンス名")
	RdsCmd.PersistentFlags().StringVarP(&stackName, "stack", "S", "", "CloudFormationスタック名")
}
