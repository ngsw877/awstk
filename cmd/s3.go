package cmd

import (
	"awstk/internal/aws"
	"awstk/internal/service"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
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
	Use:   "ls [s3-path]",
	Short: "S3バケット一覧、または指定S3パスをツリー形式で表示するコマンド",
	Long: `S3バケット一覧または指定されたS3パス以下のオブジェクトをツリー形式で表示します。
S3パスを指定した場合、デフォルトでファイルサイズが表示されます。

【使い方】
  ` + AppName + ` s3 ls                          # バケット一覧を表示
  ` + AppName + ` s3 ls s3://my-bucket           # バケット内をツリー形式で表示（サイズ付き）
  ` + AppName + ` s3 ls s3://my-bucket/prefix/   # 指定プレフィックス以下をツリー形式で表示（サイズ付き）
  ` + AppName + ` s3 ls s3://my-bucket -t        # 更新日時も一緒に表示

【例】
  ` + AppName + ` s3 ls s3://my-bucket/logs/ -t
  → my-bucket/logs/ 配下のオブジェクトをツリー形式でサイズ + 更新日時付きで表示します。`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmdCobra *cobra.Command, args []string) error {
		showTime, _ := cmdCobra.Flags().GetBool("time")

		s3Client, err := aws.NewClient[*s3.Client](aws.Context{Region: region, Profile: profile})
		if err != nil {
			return fmt.Errorf("AWS設定の読み込みエラー: %w", err)
		}

		if len(args) == 0 {
			// 引数がない場合はバケット一覧表示
			buckets, err := service.ListS3Buckets(s3Client)
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
		} else {
			// 引数がある場合は指定S3パスをツリー形式で表示
			s3Path := args[0]
			err := service.ListS3TreeView(s3Client, s3Path, showTime)
			if err != nil {
				return fmt.Errorf("❌ %w", err)
			}
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
  ` + AppName + ` s3 gunzip s3://my-bucket/some/prefix/ [-o 出力先ディレクトリ]

【例】
  ` + AppName + ` s3 gunzip s3://my-bucket/logs/ -o ./logs/
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

		s3Client, err := aws.NewClient[*s3.Client](aws.Context{Region: region, Profile: profile})
		if err != nil {
			return fmt.Errorf("AWS設定の読み込みエラー: %w", err)
		}

		if err := service.DownloadAndExtractGzFiles(s3Client, s3url, outDir); err != nil {
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

	// ls コマンドに --time フラグを追加
	s3LsCmd.Flags().BoolP("time", "t", false, "ファイルの更新日時も一緒に表示")
}
