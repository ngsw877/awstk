package cmd

import (
	"awstk/internal/aws"
	"awstk/internal/service"
	"fmt"

	"github.com/spf13/cobra"
)

// cleanupCmd represents the cleanup command
var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "AWSリソースのクリーンアップコマンド",
	Long: `指定した文字列を含むS3バケットやECRリポジトリを一括削除するコマンドです。
CloudFormationスタック名を指定することで、スタック内のリソースを対象にすることもできます。

例:
  ` + AppName + ` cleanup -k "test" -P my-profile
  ` + AppName + ` cleanup -S my-stack -P my-profile`,
	RunE: func(cmd *cobra.Command, args []string) error {
		keyword, _ := cmd.Flags().GetString("keyword")
		stackName, _ := cmd.Flags().GetString("stack")

		if keyword == "" && stackName == "" {
			return fmt.Errorf("❌ エラー: キーワード (-k) またはスタック名 (-S) のいずれかを指定してください")
		}

		// クリーンアップオプションを作成（既存の構造体に合わせる）
		awsCtx := aws.AwsContext{Region: region, Profile: profile}
		opts := service.CleanupOptions{
			AwsContext: aws.AwsContext{
				Region:  awsCtx.Region,
				Profile: awsCtx.Profile,
			},
			SearchString: keyword,
			StackName:    stackName,
		}

		// 既存の関数をそのまま使用
		err := service.CleanupResources(opts)
		if err != nil {
			return fmt.Errorf("❌ クリーンアップ中にエラーが発生しました: %w", err)
		}

		fmt.Println("✅ クリーンアップが完了しました")
		return nil
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(cleanupCmd)
	cleanupCmd.Flags().StringP("keyword", "k", "", "削除対象のキーワード")
	cleanupCmd.Flags().StringP("stack", "S", "", "CloudFormationスタック名")
}
