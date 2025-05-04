package cmd

import (
	"awsfunc/internal"
	"fmt"
	"github.com/spf13/cobra"
)

// CfnCmd represents the cfn command
var CfnCmd = &cobra.Command{
	Use:   "cfn",
	Short: "CloudFormationリソース操作コマンド",
}

// lsCmd represents the ls command
var cfnLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "CloudFormationスタック一覧を表示するコマンド",
	RunE: func(cmdCobra *cobra.Command, args []string) error {
		fmt.Println("CloudFormationスタックを取得中...")

		stackNames, err := internal.ListCfnStacks(Region, Profile)
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
	RootCmd.AddCommand(CfnCmd)
	CfnCmd.AddCommand(cfnLsCmd)
}
