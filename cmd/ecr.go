package cmd

import (
	"awstk/internal/service/common"
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

		fmt.Printf("Profile: %s\n", awsCtx.Profile)
		fmt.Printf("Region: %s\n", awsCtx.Region)
		fmt.Printf("検索文字列: %s\n", keyword)

		// キーワードに一致するリポジトリを取得
		repositories, err := ecrsvc.GetEcrRepositoriesByKeyword(ecrClient, keyword)
		if err != nil {
			return fmt.Errorf("❌ ECRリポジトリ一覧取得エラー: %w", err)
		}

		if len(repositories) == 0 {
			fmt.Printf("キーワード '%s' に一致するECRリポジトリが見つかりませんでした\n", keyword)
			return nil
		}

		// リポジトリを削除
		err = ecrsvc.CleanupEcrRepositories(ecrClient, repositories)
		if err != nil {
			return fmt.Errorf("❌ ECRリポジトリ削除エラー: %w", err)
		}

		fmt.Println("✅ ECRリポジトリの削除が完了しました")
		return nil
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

		// リポジトリ一覧を取得
		repositories, err := ecrsvc.ListEcrRepositories(ecrClient)
		if err != nil {
			return common.FormatListError("ECRリポジトリ", err)
		}

		if len(repositories) == 0 {
			fmt.Println(common.FormatEmptyMessage("ECRリポジトリ"))
			return nil
		}

		// フィルタリング処理
		filteredRepos := repositories

		// フィルタリング処理とタイトル生成
		var conditions []string
		if emptyOnly {
			conditions = append(conditions, "空の")
			filteredRepos, err = ecrsvc.FilterEmptyRepositories(ecrClient, filteredRepos)
			if err != nil {
				return fmt.Errorf("❌ 空リポジトリチェックでエラー: %w", err)
			}
		}
		if noLifecycle {
			conditions = append(conditions, "ライフサイクルポリシー未設定の")
			filteredRepos, err = ecrsvc.FilterNoLifecycleRepositories(ecrClient, filteredRepos)
			if err != nil {
				return fmt.Errorf("❌ ライフサイクルポリシーチェックでエラー: %w", err)
			}
		}
		
		title := common.GenerateFilteredTitle("ECRリポジトリ", conditions...)

		// 結果表示
		if !showDetails {
			// シンプル表示
			names := make([]string, len(filteredRepos))
			for i, repo := range filteredRepos {
				names[i] = repo.RepositoryName
			}
			common.PrintSimpleList(common.ListOutput{
				Title:        title,
				Items:        names,
				ResourceName: "リポジトリ",
				ShowCount:    true,
			})
		} else {
			// 詳細表示
			fmt.Printf("%s:\n", title)
			if len(filteredRepos) == 0 {
				fmt.Println("該当するリポジトリはありませんでした")
				return nil
			}
			for i := range filteredRepos {
				if err := ecrsvc.EnrichRepositoryDetails(ecrClient, &filteredRepos[i]); err != nil {
					fmt.Printf("  - %s (詳細取得エラー: %v)\n", filteredRepos[i].RepositoryName, err)
					continue
				}
				ecrsvc.DisplayRepositoryDetails(filteredRepos[i])
			}
			fmt.Printf("\n合計: %d個のリポジトリ\n", len(filteredRepos))
		}

		return nil
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
