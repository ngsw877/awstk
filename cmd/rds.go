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
  awsfunc rds start -S <stack-name> [-P <aws-profile>]
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		instanceId, err := resolveRdsInstanceIdentifier()
		if err != nil {
			cmd.Help()
			return err
		}

		fmt.Printf("RDSインスタンス (%s) を起動します...\n", instanceId)

		err = internal.StartRdsInstance(instanceId, region, profile)
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
  awsfunc rds stop -S <stack-name> [-P <aws-profile>]
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		instanceId, err := resolveRdsInstanceIdentifier()
		if err != nil {
			cmd.Help()
			return err
		}

		fmt.Printf("RDSインスタンス (%s) を停止します...\n", instanceId)

		err = internal.StopRdsInstance(instanceId, region, profile)
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
	rdsCmd.PersistentFlags().StringVarP(&rdsInstanceId, "db-instance-identifier", "d", "", "RDSインスタンス識別子 (-Sが指定されていない場合に必須)")
	rdsCmd.PersistentFlags().StringVarP(&stackName, "stack", "S", "", "CloudFormationスタック名 (-dが指定されていない場合に必須)")
}

// resolveRdsInstanceIdentifier はフラグの値に基づいて
// 操作対象のRDSインスタンス識別子を取得するプライベートヘルパー関数。
// ECSコマンドの resolveEcsClusterAndService 関数を参考に作成。
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
	instanceId, stackErr := internal.GetRdsFromStack(stackName, region, profile)
	if stackErr != nil {
		return "", fmt.Errorf("❌ エラー: スタックからRDSインスタンス識別子の取得に失敗しました: %w", stackErr)
	}
	fmt.Println("🔍 検出されたRDSインスタンス識別子: " + instanceId)
	return instanceId, nil
}