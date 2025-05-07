package cmd

import (
	"awsfunc/internal"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	searchString string
)

// CleanupCmd represents the cleanup command
var CleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "指定した文字列を含むAWSリソースをクリーンアップする",
	Long: `指定した文字列を含むS3バケットやECRリポジトリなどのAWSリソースを検索し、強制的に削除します。
!!! 注意 !!! このコマンドはリソースを完全に削除します。実行には十分注意してください。

例:
  awsfunc cleanup -P my-profile -r us-east-1 -k my-search-string`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		searchString, err = cmd.Flags().GetString("keyword")
		if err != nil {
			return fmt.Errorf("キーワードオプションの取得エラー: %w", err)
		}

		if searchString == "" {
			return fmt.Errorf("検索キーワードを指定してください (--keyword or -k)")
		}

		fmt.Printf("🔍 検索文字列 '%s' にマッチするリソースのクリーンアップを開始します...\n", searchString)

		// internal パッケージのクリーンアップ関数を呼び出す
		opts := internal.CleanupOptions{
			SearchString: searchString,
			Region:       Region,  // root.goで定義されているRegion
			Profile:      Profile, // root.goで定義されているProfile
		}

		err = internal.CleanupResources(opts)
		if err != nil {
			return fmt.Errorf("❌ クリーンアップ中にエラーが発生しました: %w", err)
		}

		fmt.Println("✅ クリーンアップ完了！")
		return nil
	},
	SilenceUsage: true,
}

func init() {
	// RootCmd に CleanupCmd を追加
	RootCmd.AddCommand(CleanupCmd)

	// cleanupCmd 固有のフラグがあればここに追加
	// --keyword (-k) オプションを追加
	CleanupCmd.Flags().StringVarP(&searchString, "keyword", "k", "", "クリーンアップ対象を絞り込むための検索キーワード")

	// keyword オプションを必須にする
	CleanupCmd.MarkFlagRequired("keyword")

	// ※ profile, region は root.go で定義されたグローバルフラグを使用
}
