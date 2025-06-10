package cmd

import (
	"awstk/internal"
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
検索方法は、キーワード(-k)またはCloudFormationスタック名(-S)のいずれかを指定できます。
!!! 注意 !!! このコマンドはリソースを完全に削除します。実行には十分注意してください。

例:
  awstk cleanup --profile my-profile --region us-east-1 --keyword my-search-string
  awstk cleanup -P my-profile -r us-east-1 -k my-search-string
  awstk cleanup --profile my-profile --stack my-stack
  awstk cleanup -P my-profile -S my-stack`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		searchString, err = cmd.Flags().GetString("keyword")
		if err != nil {
			return fmt.Errorf("キーワードオプションの取得エラー: %w", err)
		}

		// internal パッケージのクリーンアップ関数を呼び出す
		opts := internal.CleanupOptions{
			AwsContext:   getAwsContext(),
			SearchString: searchString,
			StackName:    stackName, // root.goで定義されているStackName
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
	CleanupCmd.Flags().StringVarP(&searchString, "keyword", "k", "", "クリーンアップ対象を絞り込むための検索キーワード")

	// ※ profile, region, stack は root.go で定義されたグローバルフラグを使用
}
