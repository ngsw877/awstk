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
	rdsStackName  string
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
		instanceName, _ := cmd.Flags().GetString("instance")
		stackName, _ := cmd.Flags().GetString("stack")
		var err error

		if stackName != "" {
			instanceName, err = cfn.GetRdsFromStack(cfnClient, stackName)
			if err != nil {
				return fmt.Errorf("❌ CloudFormationスタックからインスタンス名の取得に失敗: %w", err)
			}
			fmt.Printf("✅ CloudFormationスタック '%s' からRDSインスタンス '%s' を検出しました\n", stackName, instanceName)
		} else if instanceName == "" {
			return fmt.Errorf("❌ エラー: RDSインスタンス名 (-i) またはスタック名 (-S) を指定してください")
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
		instanceName, _ := cmd.Flags().GetString("instance")
		stackName, _ := cmd.Flags().GetString("stack")
		var err error

		if stackName != "" {
			instanceName, err = cfn.GetRdsFromStack(cfnClient, stackName)
			if err != nil {
				return fmt.Errorf("❌ CloudFormationスタックからインスタンス名の取得に失敗: %w", err)
			}
			fmt.Printf("✅ CloudFormationスタック '%s' からRDSインスタンス '%s' を検出しました\n", stackName, instanceName)
		} else if instanceName == "" {
			return fmt.Errorf("❌ エラー: RDSインスタンス名 (-i) またはスタック名 (-S) を指定してください")
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
		stackNames, err := cfn.ListCfnStacks(cfnClient)
		if err != nil {
			return fmt.Errorf("❌ CloudFormationスタック一覧取得でエラー: %w", err)
		}

		if len(stackNames) == 0 {
			fmt.Println("CloudFormationスタックが見つかりませんでした")
			return nil
		}

		fmt.Printf("CloudFormationスタック一覧: (全%d件)\n", len(stackNames))
		for i, name := range stackNames {
			fmt.Printf("  %3d. %s\n", i+1, name)
		}

		return nil
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(RdsCmd)
	RdsCmd.AddCommand(rdsStartCmd)
	RdsCmd.AddCommand(rdsStopCmd)
	RdsCmd.AddCommand(rdsLsCmd)

	// 共通フラグをRdsCmd（親コマンド）に定義
	RdsCmd.PersistentFlags().StringVarP(&rdsInstanceId, "instance", "i", "", "RDSインスタンス名")
	RdsCmd.PersistentFlags().StringVarP(&rdsStackName, "stack", "S", "", "CloudFormationスタック名")
}
