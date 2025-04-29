/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"awsfunc/cmd"
	"fmt"

	"github.com/spf13/cobra"
)

// s3Cmd represents the s3 command
var s3Cmd = &cobra.Command{
	Use:   "s3",
	Short: "S3リソース操作コマンド",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("s3 called")
	},
}

func init() {
	cmd.RootCmd.AddCommand(s3Cmd)
}
