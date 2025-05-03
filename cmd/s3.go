package cmd

import (
	"awsfunc/internal"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// S3Cmd represents the s3 command
var S3Cmd = &cobra.Command{
	Use:   "s3",
	Short: "S3リソース操作コマンド",
}

// s3LsCmd represents the ls command
var s3LsCmd = &cobra.Command{
	Use:   "ls",
	Short: "S3バケット一覧を表示するコマンド",
	Run: func(cmdCobra *cobra.Command, args []string) {
		buckets, err := internal.ListS3Buckets(Region, Profile)
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
	RootCmd.AddCommand(S3Cmd)
	S3Cmd.AddCommand(s3LsCmd)
}
