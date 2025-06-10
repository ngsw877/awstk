// Package cmd /*
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of awstk",
	Long:  `All software has versions. This is awstk's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("awstk version %s\n", version)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
