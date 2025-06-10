// Package cmd /*
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of " + AppName,
	Long:  `All software has versions. This is ` + AppName + `'s`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s version %s\n", AppName, version)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
