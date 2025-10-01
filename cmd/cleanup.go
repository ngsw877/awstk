package cmd

import (
	cleanup "awstk/internal/service/cleanup"
	"fmt"
	"os"

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
  ` + AppName + ` cleanup all -f "test" -P my-profile
  ` + AppName + ` cleanup all -S my-stack -P my-profile
  ` + AppName + ` cleanup all --stack-id arn:aws:cloudformation:... -P my-profile`,
	RunE: func(cmd *cobra.Command, args []string) error {
		resolveStackName()
		filter, _ := cmd.Flags().GetString("filter")
		stackID, _ := cmd.Flags().GetString("stack-id")
		if stackID == "" {
			if v := os.Getenv("AWS_STACK_ID"); v != "" {
				fmt.Println("🔍 環境変数 AWS_STACK_ID の値を使用します")
				stackID = v
			}
		}

		printAwsContext()

		// クライアントセットを作成
		clients := cleanup.ClientSet{
			S3Client:   s3.NewFromConfig(awsCfg),
			EcrClient:  ecr.NewFromConfig(awsCfg),
			CfnClient:  cloudformation.NewFromConfig(awsCfg),
			LogsClient: cloudwatchlogs.NewFromConfig(awsCfg),
		}

		opts := cleanup.Options{
			SearchString: filter,
			StackName:    stackName,
			StackId:      stackID,
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
	allCleanupCmd.Flags().StringP("filter", "f", "", "削除対象のフィルターパターン")
	allCleanupCmd.Flags().StringVarP(&stackName, "stack-name", "S", "", "CloudFormationスタック名")
	allCleanupCmd.Flags().StringP("stack-id", "i", "", "CloudFormationスタックID(ARN可)")
}
