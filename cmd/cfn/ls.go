package cfn

import (
	"awsfunc/cmd"
	"awsfunc/internal/cfn"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// lsCmd represents the ls command
var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "CloudFormationスタック一覧を表示するコマンド",
	Run: func(cmdCobra *cobra.Command, args []string) {
		fmt.Println("CloudFormationスタックを取得中...")

		stackNames, err := cfn.ListStacks(cmd.Region, cmd.Profile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ CloudFormationスタック一覧取得でエラー: %v\n", err)
			os.Exit(1)
		}

		if len(stackNames) == 0 {
			fmt.Println("CloudFormationスタックが見つかりませんでした")
			return
		}

		fmt.Printf("CloudFormationスタック一覧: (全%d件)\n", len(stackNames))
		for i, name := range stackNames {
			fmt.Printf("  %3d. %s\n", i+1, name)
		}
	},
}

func init() {
	CfnCmd.AddCommand(lsCmd)
}
