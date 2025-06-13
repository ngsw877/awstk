// Package cmd /*
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = "dev" // ビルド時に設定される

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "バージョン情報を表示",
	Long:  `awstkのバージョン情報を表示します。`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("awstk version %s\n", Version)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
