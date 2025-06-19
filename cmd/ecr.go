package cmd

import (
	"awstk/internal/aws"
	ecrsvc "awstk/internal/service/ecr"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/spf13/cobra"
)

// EcrCmd represents the ecr command
var EcrCmd = &cobra.Command{
	Use:   "ecr",
	Short: "ECRリソース操作コマンド",
	Long:  `ECR（Elastic Container Registry）を操作するためのコマンド群です。`,
}

// ecrCleanupCmd represents the cleanup command
var ecrCleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "ECRリポジトリを削除するコマンド",
	Long: `指定したキーワードを含むECRリポジトリを削除します。

例:
  ` + AppName + ` ecr cleanup -k "test-ecr-repo" -P my-profile`,
	RunE: func(cmd *cobra.Command, args []string) error {
		keyword, _ := cmd.Flags().GetString("keyword")
		if keyword == "" {
			return fmt.Errorf("❌ エラー: キーワード (-k) を指定してください")
		}

		fmt.Printf("Profile: %s\n", awsCtx.Profile)
		fmt.Printf("Region: %s\n", awsCtx.Region)
		fmt.Printf("検索文字列: %s\n", keyword)

		ecrClient, err := aws.NewClient[*ecr.Client](awsCtx)
		if err != nil {
			return fmt.Errorf("ECRクライアント作成エラー: %w", err)
		}

		// キーワードに一致するリポジトリを取得
		repos, err := ecrsvc.GetEcrRepositoriesByKeyword(ecrClient, keyword)
		if err != nil {
			return fmt.Errorf("❌ ECRリポジトリ一覧取得エラー: %w", err)
		}

		if len(repos) == 0 {
			fmt.Printf("キーワード '%s' に一致するECRリポジトリが見つかりませんでした\n", keyword)
			return nil
		}

		// リポジトリを削除
		err = ecrsvc.CleanupEcrRepositories(ecrClient, repos)
		if err != nil {
			return fmt.Errorf("❌ ECRリポジトリ削除エラー: %w", err)
		}

		fmt.Println("✅ ECRリポジトリの削除が完了しました")
		return nil
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(EcrCmd)
	EcrCmd.AddCommand(ecrCleanupCmd)

	// cleanup コマンドのフラグ
	ecrCleanupCmd.Flags().StringP("keyword", "k", "", "削除対象のキーワード")
}
