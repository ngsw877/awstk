package cmd

import (
	"awsfunc/internal"
	"fmt"
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
	RunE: func(cmdCobra *cobra.Command, args []string) error {
		buckets, err := internal.ListS3Buckets(Region, Profile)
		if err != nil {
			return fmt.Errorf("❌ S3バケット一覧取得でエラー: %w", err)
		}
		if len(buckets) == 0 {
			fmt.Println("S3バケットが見つかりませんでした")
			return nil
		}
		fmt.Println("S3バケット一覧:")
		for _, name := range buckets {
			fmt.Println("  -", name)
		}
		return nil
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(S3Cmd)
	S3Cmd.AddCommand(s3LsCmd)
}
