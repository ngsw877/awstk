/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"awsfunc/cmd"
	"awsfunc/internal/s3"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// lsCmd represents the ls command
var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "S3バケット一覧を表示するコマンド",
	Run: func(cmdCobra *cobra.Command, args []string) {
		buckets, err := s3.ListBuckets(cmd.Region, cmd.Profile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "S3バケット一覧取得でエラー: %v\n", err)
			os.Exit(1)
		}
		if len(buckets) == 0 {
			fmt.Println("S3バケットが見つかりませんでした")
			return
		}
		fmt.Println("S3バケット一覧:")
		for _, name := range buckets {
			fmt.Println("  -", name)
		}
	},
}

func init() {
	s3Cmd.AddCommand(lsCmd)
}
