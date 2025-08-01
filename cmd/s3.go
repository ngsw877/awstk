package cmd

import (
	"awstk/internal/service/common"
	s3svc "awstk/internal/service/s3"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"
)

var s3Client *s3.Client

// S3Cmd represents the s3 command
var S3Cmd = &cobra.Command{
	Use:          "s3",
	Short:        "S3リソース操作コマンド",
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// 親のPersistentPreRunEを実行（awsCtx設定とAWS設定読み込み）
		if err := RootCmd.PersistentPreRunE(cmd, args); err != nil {
			return err
		}

		// S3用クライアント生成
		s3Client = s3.NewFromConfig(awsCfg)

		return nil
	},
}

// s3LsCmd represents the ls command
var s3LsCmd = &cobra.Command{
	Use:   "ls [s3-path]",
	Short: "S3バケット一覧、または指定S3パスをツリー形式で表示するコマンド",
	Long: `S3バケット一覧または指定されたS3パス以下のオブジェクトをツリー形式で表示します。
S3パスを指定した場合、デフォルトでファイルサイズが表示されます。

【使い方】
  ` + AppName + ` s3 ls                          # バケット一覧を表示
  ` + AppName + ` s3 ls -e                       # 空のバケットのみを表示
  ` + AppName + ` s3 ls my-bucket                # バケット内をツリー形式で表示（サイズ付き）
  ` + AppName + ` s3 ls my-bucket/prefix/        # 指定プレフィックス以下をツリー形式で表示（サイズ付き）
  ` + AppName + ` s3 ls my-bucket -t             # 更新日時も一緒に表示

【例】
  ` + AppName + ` s3 ls -e
  → 空のS3バケットのみを一覧表示します。
  
  ` + AppName + ` s3 ls my-bucket/logs/ -t
  → my-bucket/logs/ 配下のオブジェクトをツリー形式でサイズ + 更新日時付きで表示します。`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmdCobra *cobra.Command, args []string) error {
		showTime, _ := cmdCobra.Flags().GetBool("time")
		emptyOnly, _ := cmdCobra.Flags().GetBool("empty-only")

		if len(args) == 0 {
			// 引数がない場合はバケット一覧表示
			buckets, err := s3svc.ListS3Buckets(s3Client)
			if err != nil {
				return common.FormatListError("S3バケット", err)
			}
			if len(buckets) == 0 {
				fmt.Println(common.FormatEmptyMessage("S3バケット"))
				return nil
			}

			// 空バケットのみ表示する場合
			if emptyOnly {
				emptyBuckets, err := s3svc.FilterEmptyBuckets(s3Client, buckets)
				if err != nil {
					return fmt.Errorf("❌ 空バケットのチェックでエラー: %w", err)
				}
				common.PrintSimpleList(common.ListOutput{
					Title:        "空のS3バケット一覧",
					Items:        emptyBuckets,
					ResourceName: "バケット",
					ShowCount:    false,
				})
			} else {
				common.PrintSimpleList(common.ListOutput{
					Title:        "S3バケット一覧",
					Items:        buckets,
					ResourceName: "バケット",
					ShowCount:    false,
				})
			}
		} else {
			// 引数がある場合は指定S3パスをツリー形式で表示
			s3Path := args[0]
			err := s3svc.ListS3TreeView(s3Client, s3Path, showTime)
			if err != nil {
				return fmt.Errorf("❌ %w", err)
			}
		}

		return nil
	},
	SilenceUsage: true,
}

var s3GunzipCmd = &cobra.Command{
	Use:   "gunzip [バケット名/プレフィックス]",
	Short: "S3の.gzファイルを一括ダウンロード＆解凍するコマンド",
	Long: `S3バケット内の指定prefix配下に存在する全ての.gzファイルを一括でダウンロードし、解凍してローカルに保存するコマンドです。

【使い方】
  ` + AppName + ` s3 gunzip <バケット名>[/プレフィックス] [-o 出力先ディレクトリ]

【例】
  ` + AppName + ` s3 gunzip my-bucket/logs/ -o ./logs/
  ` + AppName + ` s3 gunzip my-bucket -o ./data/
  → my-bucket/logs/ 配下の .gz ファイルを全部ダウンロード＆解凍して指定ディレクトリに保存します。

出力先ディレクトリを省略した場合は ./outputs/ に保存されます。`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmdCobra *cobra.Command, args []string) error {
		s3Path := args[0]
		outDir, _ := cmdCobra.Flags().GetString("out")
		if outDir == "" {
			outDir = "./outputs/"
		}

		fmt.Printf("S3パス: %s\n出力先: %s\n", s3Path, outDir)

		if err := s3svc.DownloadAndExtractGzFiles(s3Client, s3Path, outDir); err != nil {
			return fmt.Errorf("❌ gunzip失敗: %w", err)
		}
		return nil
	},
	SilenceUsage: true,
}

// s3AvailCmd represents the avail command
var s3AvailCmd = &cobra.Command{
	Use:   "avail [bucket-names...]",
	Short: "指定したS3バケット名が利用可能かチェック",
	Long:  `指定した複数のS3バケット名が利用可能か（未作成か）を判定します。\n\n【使い方】\n  ` + AppName + ` s3 avail bucket1 bucket2 ...\n\n【出力例】\n  [404] my-bucket-1: 利用可能\n  [200] my-bucket-2: 利用不可（すでに存在）\n  [403] my-bucket-3: 利用不可（存在するがアクセス権限なし）`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmdCobra *cobra.Command, args []string) error {
		return s3svc.CheckAndDisplayBucketsAvailability(s3Client, args)
	},
	SilenceUsage: true,
}

// s3CleanupCmd represents the cleanup command
var s3CleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "S3バケットを削除するコマンド",
	Long: `指定したキーワードを含むS3バケットを削除します。

例:
  ` + AppName + ` s3 cleanup -f "test-bucket" -P my-profile`,
	RunE: func(cmd *cobra.Command, args []string) error {
		filter, _ := cmd.Flags().GetString("filter")
		if filter == "" {
			return fmt.Errorf("❌ エラー: フィルター (-f) を指定してください")
		}

		printAwsContextWithInfo("検索文字列", filter)

		// フィルターに一致するバケットを取得
		buckets, err := s3svc.GetS3BucketsByFilter(s3Client, filter)
		if err != nil {
			return fmt.Errorf("❌ S3バケット一覧取得エラー: %w", err)
		}

		if len(buckets) == 0 {
			fmt.Printf("フィルター '%s' に一致するS3バケットが見つかりませんでした\n", filter)
			return nil
		}

		// バケットを削除
		err = s3svc.CleanupS3Buckets(s3Client, buckets)
		if err != nil {
			return fmt.Errorf("❌ S3バケット削除エラー: %w", err)
		}

		fmt.Println("✅ S3バケットの削除が完了しました")
		return nil
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(S3Cmd)
	S3Cmd.AddCommand(s3LsCmd)
	S3Cmd.AddCommand(s3GunzipCmd)
	S3Cmd.AddCommand(s3AvailCmd)
	S3Cmd.AddCommand(s3CleanupCmd)
	s3GunzipCmd.Flags().StringP("out", "o", "", "解凍ファイルの出力先ディレクトリ (デフォルト: ./outputs/)")

	// ls コマンドに --time フラグを追加
	s3LsCmd.Flags().BoolP("time", "t", false, "ファイルの更新日時も一緒に表示")
	// ls コマンドに --empty-only フラグを追加
	s3LsCmd.Flags().BoolP("empty-only", "e", false, "空のバケットのみを表示")

	// cleanup コマンドのフラグ
	s3CleanupCmd.Flags().StringP("filter", "f", "", "削除対象のフィルターパターン")
}
