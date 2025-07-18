package cmd

import (
	ecrsvc "awstk/internal/service/ecr"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/spf13/cobra"
)

var ecrClient *ecr.Client

// EcrCmd represents the ecr command
var EcrCmd = &cobra.Command{
	Use:          "ecr",
	Short:        "ECRリソース操作コマンド",
	Long:         `ECR（Elastic Container Registry）を操作するためのコマンド群です。`,
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// 親のPersistentPreRunEを実行（awsCtx設定とAWS設定読み込み）
		if err := RootCmd.PersistentPreRunE(cmd, args); err != nil {
			return err
		}

		// ECR用クライアント生成
		ecrClient = ecr.NewFromConfig(awsCfg)

		return nil
	},
}

// ecrCleanupCmd represents the cleanup command
var ecrCleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "ECRリポジトリを削除するコマンド",
	Long: `指定したキーワードを含むECRリポジトリを削除します。

例:
  ` + AppName + ` ecr cleanup -k "test-repo" -P my-profile`,
	RunE: func(cmd *cobra.Command, args []string) error {
		keyword, _ := cmd.Flags().GetString("keyword")
		if keyword == "" {
			return fmt.Errorf("❌ エラー: キーワード (-k) を指定してください")
		}

		printAwsContextWithInfo("検索文字列", keyword)

		return ecrsvc.CleanupRepositoriesByKeyword(ecrClient, keyword)
	},
	SilenceUsage: true,
}

// ecrLsCmd represents the ls command
var ecrLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "ECRリポジトリ一覧を表示するコマンド",
	Long: `ECRリポジトリの一覧を表示します。
イメージ数、サイズ、ライフサイクルポリシーの有無などの情報も含めて表示します。

【使い方】
  ` + AppName + ` ecr ls                    # リポジトリ一覧を表示
  ` + AppName + ` ecr ls -e                 # 空のリポジトリのみを表示
  ` + AppName + ` ecr ls -n                 # ライフサイクルポリシー未設定のリポジトリのみを表示
  ` + AppName + ` ecr ls --details          # 詳細情報付きで表示
  ` + AppName + ` ecr ls -e -n              # 空かつポリシー未設定のリポジトリを表示

【例】
  ` + AppName + ` ecr ls -n
  → ライフサイクルポリシーが未設定のECRリポジトリのみを一覧表示します。
  
  ` + AppName + ` ecr ls -e -d
  → 空のリポジトリを詳細情報付きで表示します。`,
	RunE: func(cmdCobra *cobra.Command, args []string) error {
		emptyOnly, _ := cmdCobra.Flags().GetBool("empty-only")
		noLifecycle, _ := cmdCobra.Flags().GetBool("no-lifecycle")
		showDetails, _ := cmdCobra.Flags().GetBool("details")

		opts := ecrsvc.ListOptions{
			EmptyOnly:   emptyOnly,
			NoLifecycle: noLifecycle,
			ShowDetails: showDetails,
		}

		return ecrsvc.ListRepositories(ecrClient, opts)
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(EcrCmd)
	EcrCmd.AddCommand(ecrLsCmd)
	EcrCmd.AddCommand(ecrCleanupCmd)

	// ls コマンドのフラグ
	ecrLsCmd.Flags().BoolP("empty-only", "e", false, "空のリポジトリのみを表示")
	ecrLsCmd.Flags().BoolP("no-lifecycle", "n", false, "ライフサイクルポリシー未設定のリポジトリのみを表示")
	ecrLsCmd.Flags().BoolP("details", "d", false, "詳細情報を表示")

	// cleanup コマンドのフラグ
	ecrCleanupCmd.Flags().StringP("keyword", "k", "", "削除対象のキーワード")
}
