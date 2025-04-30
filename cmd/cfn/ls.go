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
		stacks, err := cfn.ListStacks(cmd.Region, cmd.Profile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "CloudFormationスタック一覧取得でエラー: %v\n", err)
			os.Exit(1)
		}
		if len(stacks) == 0 {
			fmt.Println("CloudFormationスタックが見つかりませんでした")
			return
		}
		fmt.Println("CloudFormationスタック一覧:")
		for _, name := range stacks {
			fmt.Println("  -", name)
		}
	},
}

func init() {
	CfnCmd.AddCommand(lsCmd)
}
