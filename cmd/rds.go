package cmd

import (
	"awsfunc/internal"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	rdsInstanceId string
)

var RdsCmd = &cobra.Command{
	Use:   "rds",
	Short: "RDSリソース操作コマンド",
	Long:  `RDSリソースを操作するためのコマンド群です。`,
}

var rdsStartCmd = &cobra.Command{
	Use:   "start",
	Short: "RDSインスタンスを起動するコマンド",
	Long: `RDSインスタンスを起動するコマンドです。
CloudFormationスタック名を指定するか、インスタンス識別子を直接指定することができます。

例:
  awsfunc rds start -P my-profile -S my-stack
  awsfunc rds start -P my-profile -i my-db-instance`,
	RunE: func(cmd *cobra.Command, args []string) error {
		awsCtx := getAwsContext()

		var instanceId string
		var err error

		if stackName != "" {
			fmt.Println("CloudFormationスタックからRDS情報を取得します...")
			instanceId, err = internal.GetRdsFromStack(awsCtx, stackName)
			if err != nil {
				return fmt.Errorf("❌ エラー: %w", err)
			}
			fmt.Printf("🔍 検出されたRDSインスタンス: %s\n", instanceId)
		} else if rdsInstanceId != "" {
			instanceId = rdsInstanceId
		} else {
			cmd.Help()
			return fmt.Errorf("❌ エラー: スタック名 (-S) またはRDSインスタンス識別子 (-i) が必須です")
		}

		fmt.Printf("🚀 RDSインスタンス '%s' を起動します...\n", instanceId)
		err = internal.StartRdsInstance(awsCtx, instanceId)
		if err != nil {
			return fmt.Errorf("❌ エラー: %w", err)
		}

		fmt.Printf("✅ RDSインスタンス '%s' の起動を開始しました\n", instanceId)
		return nil
	},
	SilenceUsage: true,
}

var rdsStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "RDSインスタンスを停止するコマンド",
	Long: `RDSインスタンスを停止するコマンドです。
CloudFormationスタック名を指定するか、インスタンス識別子を直接指定することができます。

例:
  awsfunc rds stop -P my-profile -S my-stack
  awsfunc rds stop -P my-profile -i my-db-instance`,
	RunE: func(cmd *cobra.Command, args []string) error {
		awsCtx := getAwsContext()

		var instanceId string
		var err error

		if stackName != "" {
			fmt.Println("CloudFormationスタックからRDS情報を取得します...")
			instanceId, err = internal.GetRdsFromStack(awsCtx, stackName)
			if err != nil {
				return fmt.Errorf("❌ エラー: %w", err)
			}
			fmt.Printf("🔍 検出されたRDSインスタンス: %s\n", instanceId)
		} else if rdsInstanceId != "" {
			instanceId = rdsInstanceId
		} else {
			cmd.Help()
			return fmt.Errorf("❌ エラー: スタック名 (-S) またはRDSインスタンス識別子 (-i) が必須です")
		}

		fmt.Printf("🛑 RDSインスタンス '%s' を停止します...\n", instanceId)
		err = internal.StopRdsInstance(awsCtx, instanceId)
		if err != nil {
			return fmt.Errorf("❌ エラー: %w", err)
		}

		fmt.Printf("✅ RDSインスタンス '%s' の停止を開始しました\n", instanceId)
		return nil
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(RdsCmd)
	RdsCmd.AddCommand(rdsStartCmd)
	RdsCmd.AddCommand(rdsStopCmd)

	// startコマンドのフラグを設定
	rdsStartCmd.Flags().StringVarP(&stackName, "stack", "S", "", "CloudFormationスタック名")
	rdsStartCmd.Flags().StringVarP(&rdsInstanceId, "instance", "i", "", "RDSインスタンス識別子 (-Sが指定されていない場合に必須)")

	// stopコマンドのフラグを設定
	rdsStopCmd.Flags().StringVarP(&stackName, "stack", "S", "", "CloudFormationスタック名")
	rdsStopCmd.Flags().StringVarP(&rdsInstanceId, "instance", "i", "", "RDSインスタンス識別子 (-Sが指定されていない場合に必須)")
}

// resolveRdsInstanceIdentifier はフラグの値に基づいて
// 操作対象のRDSインスタンス識別子を取得するプライベートヘルパー関数。
func resolveRdsInstanceIdentifier() (instanceId string, err error) {
	if rdsInstanceId != "" && stackName != "" {
		return "", fmt.Errorf("❌ エラー: RDSインスタンス識別子 (-d) とスタック名 (-S) は同時に指定できません")
	}
	if rdsInstanceId == "" && stackName == "" {
		return "", fmt.Errorf("❌ エラー: RDSインスタンス識別子 (-d) またはスタック名 (-S) のどちらかが必要です")
	}
	// -d で直接指定された場合
	if rdsInstanceId != "" {
		return rdsInstanceId, nil
	}
	// -S でスタック名が指定された場合
	fmt.Println("CloudFormationスタックからRDSインスタンス識別子を取得します...")
	instanceId, err = internal.GetRdsFromStack(getAwsContext(), stackName)
	if err != nil {
		return "", fmt.Errorf("❌ エラー: スタックからRDSインスタンス識別子の取得に失敗しました: %w", err)
	}
	fmt.Println("🔍 検出されたRDSインスタンス識別子: " + instanceId)
	return instanceId, nil
}
