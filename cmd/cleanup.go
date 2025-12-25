package cmd

import (
	cleanup "awstk/internal/service/cleanup"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
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
	Short: "S3バケット、ECRリポジトリ、CloudWatch Logsを横断削除",
	Long: `指定した文字列を含むS3バケット、ECRリポジトリ、CloudWatch Logsグループを一括削除するコマンドです。
CloudFormationスタック名またはスタックIDを指定することで、スタック内のリソースを対象にすることもできます。

例:
  ` + AppName + ` cleanup all -s "test" -P my-profile
  ` + AppName + ` cleanup all -S my-stack -P my-profile
  ` + AppName + ` cleanup all --stack-id arn:aws:cloudformation:... -P my-profile`,
	RunE: func(cmd *cobra.Command, args []string) error {
		resolveStackName()
		search, _ := cmd.Flags().GetString("search")
		stackID, _ := cmd.Flags().GetString("stack-id")
		exact, _ := cmd.Flags().GetBool("exact")

		printAwsContext()

		// クライアントセットを作成
		clients := cleanup.ClientSet{
			S3Client:   s3.NewFromConfig(awsCfg),
			EcrClient:  ecr.NewFromConfig(awsCfg),
			CfnClient:  cloudformation.NewFromConfig(awsCfg),
			LogsClient: cloudwatchlogs.NewFromConfig(awsCfg),
		}

		opts := cleanup.Options{
			SearchString: search,
			StackName:    stackName,
			StackId:      stackID,
			Exact:        exact,
		}

		if err := cleanup.CleanupResources(clients, opts); err != nil {
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
	allCleanupCmd.Flags().StringP("search", "s", "", "削除対象の検索パターン")
	allCleanupCmd.Flags().StringVarP(&stackName, "stack-name", "S", "", "CloudFormationスタック名")
	allCleanupCmd.Flags().StringP("stack-id", "i", "", "CloudFormationスタックID(ARN可)")
	allCleanupCmd.Flags().Bool("exact", false, "大文字小文字を区別してマッチ")
}
