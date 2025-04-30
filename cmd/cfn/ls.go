package cfn

import (
	"awsfunc/internal/cfn"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var region string
var profile string

// lsCmd represents the ls command
var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "CloudFormationスタック一覧を表示するコマンド",
	Run: func(cmd *cobra.Command, args []string) {
		stacks, err := cfn.ListStacks(region, profile)
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
	lsCmd.Flags().StringVarP(&region, "region", "r", "ap-northeast-1", "AWSリージョン")
	lsCmd.Flags().StringVarP(&profile, "profile", "P", "", "AWSプロファイル")
}
