package cmd

import (
	"awstk/internal"
	"fmt"

	"github.com/spf13/cobra"
)

// CfnCmd represents the cfn command
var CfnCmd = &cobra.Command{
	Use:   "cfn",
	Short: "CloudFormationリソース操作コマンド",
	Long:  "CloudFormationスタックおよびスタック内リソースを操作するCLIコマンド群です。",
}

// lsCmd represents the ls command
var cfnLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "CloudFormationスタック一覧を表示するコマンド",
	RunE: func(cmdCobra *cobra.Command, args []string) error {
		fmt.Println("CloudFormationスタックを取得中...")

		awsCtx := getAwsContext()
		stackNames, err := internal.ListCfnStacks(awsCtx)
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

// cfnStartCmd はCloudFormationスタック内のリソースを一括起動するコマンド
var cfnStartCmd = &cobra.Command{
	Use:   "start",
	Short: "スタック内のリソースを一括起動するコマンド",
	Long: `指定したCloudFormationスタック内のすべての操作可能なリソース（EC2、RDS、Aurora、ECS）を一括で起動します。

例:
  awstk cfn start -S <stack-name> [-P <aws-profile>]
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if stackName == "" {
			return fmt.Errorf("❌ エラー: CloudFormationスタック名は必須です (-S で指定)")
		}

		fmt.Printf("CloudFormationスタック (%s) 内のリソースを一括起動します...\n", stackName)

		awsCtx := getAwsContext()
		err := internal.StartAllStackResources(awsCtx, stackName)
		if err != nil {
			fmt.Println("❌ スタック内の一部またはすべてのリソースの起動に失敗しました。")
			return err
		}

		fmt.Println("✅ スタック内のリソースの起動を開始しました。")
		return nil
	},
	SilenceUsage: true,
}

// cfnStopCmd はCloudFormationスタック内のリソースを一括停止するコマンド
var cfnStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "スタック内のリソースを一括停止するコマンド",
	Long: `指定したCloudFormationスタック内のすべての操作可能なリソース（EC2、RDS、Aurora、ECS）を一括で停止します。

例:
  awstk cfn stop -S <stack-name> [-P <aws-profile>]
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if stackName == "" {
			return fmt.Errorf("❌ エラー: CloudFormationスタック名は必須です (-S で指定)")
		}

		fmt.Printf("CloudFormationスタック (%s) 内のリソースを一括停止します...\n", stackName)

		awsCtx := getAwsContext()
		err := internal.StopAllStackResources(awsCtx, stackName)
		if err != nil {
			fmt.Println("❌ スタック内の一部またはすべてのリソースの停止に失敗しました。")
			return err
		}

		fmt.Println("✅ スタック内のリソースの停止を開始しました。")
		return nil
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(CfnCmd)
	CfnCmd.AddCommand(cfnLsCmd)
	CfnCmd.AddCommand(cfnStartCmd)
	CfnCmd.AddCommand(cfnStopCmd)

	// スタック名フラグを追加
	CfnCmd.PersistentFlags().StringVarP(&stackName, "stack", "S", "", "CloudFormationスタック名")
}
