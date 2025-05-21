package cmd

import (
	"awsfunc/internal"
	"fmt"

	"github.com/spf13/cobra"
)

// S3Cmd represents the s3 command
var S3Cmd = &cobra.Command{
	Use:          "s3",
	Short:        "S3リソース操作コマンド",
	SilenceUsage: true,
}

// s3LsCmd represents the ls command
var s3LsCmd = &cobra.Command{
	Use:   "ls",
	Short: "S3バケット一覧を表示するコマンド",
	RunE: func(cmdCobra *cobra.Command, args []string) error {
		buckets, err := internal.ListS3Buckets(region, profile)
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

var s3GunzipCmd = &cobra.Command{
	Use:   "gunzip [s3パス]",
	Short: "S3の.gzファイルを一括ダウンロード＆解凍するコマンド",
	Long: `S3バケット内の指定prefix配下に存在する全ての.gzファイルを一括でダウンロードし、解凍してローカルに保存するコマンドです。

【使い方】
  awsfunc s3 gunzip s3://my-bucket/some/prefix/ [-o 出力先ディレクトリ]

【例】
  awsfunc s3 gunzip s3://my-bucket/logs/ -o ./logs/
  → my-bucket/logs/ 配下の .gz ファイルを全部ダウンロード＆解凍して ./logs/ に保存します。

出力先ディレクトリを省略した場合は ./outputs/ に保存されます。`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmdCobra *cobra.Command, args []string) error {
		if len(args) == 0 {
			cmdCobra.Help()
			return nil
		}
		s3url := args[0]
		outDir, _ := cmdCobra.Flags().GetString("out")
		if outDir == "" {
			outDir = "./outputs/"
		}
		fmt.Printf("S3パス: %s\n出力先: %s\n", s3url, outDir)
		if err := internal.DownloadAndExtractGzFiles(s3url, outDir, region, profile); err != nil {
			return fmt.Errorf("❌ gunzip失敗: %w", err)
		}
		return nil
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(S3Cmd)
	S3Cmd.AddCommand(s3LsCmd)
	S3Cmd.AddCommand(s3GunzipCmd)
	s3GunzipCmd.Flags().StringP("out", "o", "", "解凍ファイルの出力先ディレクトリ (デフォルト: ./outputs/)")
}
