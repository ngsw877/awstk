package cmd

import (
	"awstk/internal/service/precommit"

	"github.com/spf13/cobra"
)

// precommitCmd はprecommitコマンドを表す
var precommitCmd = &cobra.Command{
	Use:   "precommit",
	Short: "awstkプロジェクトのpre-commitフック管理",
	Long: `awstkプロジェクトのpre-commitフックを管理します。

このコマンドではpre-commitフックの有効化、無効化、状態確認ができます。
現在のpre-commitフックはCursor Rules（.mdcファイル）をCLAUDE.mdと自動同期します。`,
}

// precommitEnableCmd はprecommit enableコマンドを表す
var precommitEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "pre-commitフックを有効化",
	Long:  `core.hooksPathを.githooksに設定してpre-commitフックを有効化します`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return precommit.Enable()
	},
}

// precommitDisableCmd はprecommit disableコマンドを表す
var precommitDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "pre-commitフックを無効化",
	Long:  `core.hooksPathの設定を解除してpre-commitフックを無効化します`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return precommit.Disable()
	},
}

// precommitStatusCmd はprecommit statusコマンドを表す
var precommitStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "pre-commitフックの状態表示",
	Long:  `pre-commitフックの現在の状態を表示します`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return precommit.ShowStatus()
	},
}

func init() {
	RootCmd.AddCommand(precommitCmd)
	precommitCmd.AddCommand(precommitEnableCmd)
	precommitCmd.AddCommand(precommitDisableCmd)
	precommitCmd.AddCommand(precommitStatusCmd)
}
