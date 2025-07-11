package cmd

import (
	cleanup "awstk/internal/service/cleanup"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"
)

// cleanupCmd represents the cleanup command
var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "AWSリソースのクリーンアップコマンド",
	Long:  `AWSリソースを削除するためのコマンド群です。`,
}

// allCleanupCmd represents the all subcommand
var allCleanupCmd = &cobra.Command{
	Use:   "all",
	Short: "S3バケットとECRリポジトリを横断削除",
	Long: `指定した文字列を含むS3バケットやECRリポジトリを一括削除するコマンドです。
CloudFormationスタック名を指定することで、スタック内のリソースを対象にすることもできます。

例:
  ` + AppName + ` cleanup all -k "test" -P my-profile
  ` + AppName + ` cleanup all -S my-stack -P my-profile`,
	RunE: func(cmd *cobra.Command, args []string) error {
		resolveStackName()
		keyword, _ := cmd.Flags().GetString("keyword")

		if keyword == "" && stackName == "" {
			return fmt.Errorf("❌ エラー: キーワード (-k) またはスタック名 (-S) のいずれかを指定してください")
		}

		fmt.Printf("Profile: %s\n", awsCtx.Profile)
		fmt.Printf("Region: %s\n", awsCtx.Region)

		// 各種クライアントを作成
		s3Client := s3.NewFromConfig(awsCfg)
		ecrClient := ecr.NewFromConfig(awsCfg)
		cfnClient := cloudformation.NewFromConfig(awsCfg)

		opts := cleanup.Options{
			S3Client:     s3Client,
			EcrClient:    ecrClient,
			CfnClient:    cfnClient,
			SearchString: keyword,
			StackName:    stackName,
		}

		err := cleanup.CleanupResources(opts)
		if err != nil {
			return fmt.Errorf("❌ クリーンアップ処理でエラー: %w", err)
		}

		fmt.Println("✅ クリーンアップが完了しました")
		return nil
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(cleanupCmd)
	cleanupCmd.AddCommand(allCleanupCmd)
	allCleanupCmd.Flags().StringP("keyword", "k", "", "削除対象のキーワード")
	allCleanupCmd.Flags().StringVarP(&stackName, "stack", "S", "", "CloudFormationスタック名")
}
