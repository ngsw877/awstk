// Package cmd /*
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of awsfunc",
	Long:  `All software has versions. This is awsfunc's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("awsfunc version %s\n", Version)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
