package cmd

import (
	"awstk/internal/service/precommit"

	"github.com/spf13/cobra"
)

// precommitCmd represents the precommit command
var precommitCmd = &cobra.Command{
	Use:   "precommit",
	Short: "Manage pre-commit hooks for the awstk project",
	Long: `Manage pre-commit hooks for the awstk project.

This command helps you enable, disable, and check the status of pre-commit hooks.
Currently, the pre-commit hook automatically syncs Cursor Rules (.mdc files) with CLAUDE.md.`,
}

// precommitEnableCmd represents the precommit enable command
var precommitEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable pre-commit hooks",
	Long:  `Enable pre-commit hooks by setting core.hooksPath to .githooks`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return precommit.Enable()
	},
}

// precommitDisableCmd represents the precommit disable command
var precommitDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable pre-commit hooks",
	Long:  `Disable pre-commit hooks by unsetting core.hooksPath`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return precommit.Disable()
	},
}

// precommitStatusCmd represents the precommit status command
var precommitStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show pre-commit hook status",
	Long:  `Show the current status of pre-commit hooks`,
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
