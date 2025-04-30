package cfn

import (
	"awsfunc/cmd"

	"github.com/spf13/cobra"
)

// CfnCmd represents the cfn command
var CfnCmd = &cobra.Command{
	Use:   "cfn",
	Short: "CloudFormationリソース操作コマンド",
}

func init() {
	cmd.RootCmd.AddCommand(CfnCmd)
}
