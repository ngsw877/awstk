package cmd

import (
	"awstk/internal/aws"
	"awstk/internal/service/cfn"
	ecrsvc "awstk/internal/service/ecr"
	s3svc "awstk/internal/service/s3"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"
)

// CfnCmd represents the cfn command
var CfnCmd = &cobra.Command{
	Use:   "cfn",
	Short: "CloudFormationリソース操作コマンド",
	Long:  `CloudFormationリソースを操作するためのコマンド群です。`,
}

var cfnLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "CloudFormationスタック一覧を表示するコマンド",
	Long:  `CloudFormationスタック一覧を表示します。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfnClient, err := aws.NewClient[*cloudformation.Client](awsCtx)
		if err != nil {
			return fmt.Errorf("AWS設定の読み込みエラー: %w", err)
		}

		stackNames, err := cfn.ListCfnStacks(cfnClient)
		if err != nil {
			return fmt.Errorf("❌ CloudFormationスタック一覧取得でエラー: %w", err)
		}

		if len(stackNames) == 0 {
			fmt.Println("CloudFormationスタックが見つかりませんでした")
			return nil
		}

		fmt.Printf("CloudFormationスタック一覧: (全%d件)\n", len(stackNames))
		for i, name := range stackNames {
			fmt.Printf("  %3d. %s\n", i+1, name)
		}

		return nil
	},
	SilenceUsage: true,
}

var cfnStartCmd = &cobra.Command{
	Use:   "start",
	Short: "CloudFormationスタック内のリソースを一括起動するコマンド",
	Long: `CloudFormationスタック内の起動・停止可能なリソースを一括起動します。
対象リソース: EC2インスタンス、RDSインスタンス、Aurora DBクラスター、ECSサービス

例:
  ` + AppName + ` cfn start -S my-stack -P my-profile`,
	RunE: func(cmd *cobra.Command, args []string) error {
		stackName, _ := cmd.Flags().GetString("stack")
		if stackName == "" {
			return fmt.Errorf("❌ エラー: スタック名 (-S) を指定してください")
		}

		fmt.Printf("Profile: %s\n", awsCtx.Profile)
		fmt.Printf("Region: %s\n", awsCtx.Region)
		fmt.Printf("Stack: %s\n", stackName)

		// 各種クライアントを作成
		cfnClient, err := aws.NewClient[*cloudformation.Client](awsCtx)
		if err != nil {
			return fmt.Errorf("CloudFormationクライアント作成エラー: %w", err)
		}

		ec2Client, err := aws.NewClient[*ec2.Client](awsCtx)
		if err != nil {
			return fmt.Errorf("EC2クライアント作成エラー: %w", err)
		}

		rdsClient, err := aws.NewClient[*rds.Client](awsCtx)
		if err != nil {
			return fmt.Errorf("RDSクライアント作成エラー: %w", err)
		}

		autoScalingClient, err := aws.NewClient[*applicationautoscaling.Client](awsCtx)
		if err != nil {
			return fmt.Errorf("ApplicationAutoScalingクライアント作成エラー: %w", err)
		}

		// start用のオプションを作成
		startOpts := cfn.StackStartStopOptions{
			CfnClient:                    cfnClient,
			Ec2Client:                    ec2Client,
			RdsClient:                    rdsClient,
			ApplicationAutoScalingClient: autoScalingClient,
			StackName:                    stackName,
		}

		err = cfn.StartAllStackResources(startOpts)
		if err != nil {
			return fmt.Errorf("❌ リソース起動処理でエラー: %w", err)
		}

		fmt.Println("✅ リソース起動処理が完了しました")
		return nil
	},
	SilenceUsage: true,
}

var cfnStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "CloudFormationスタック内のリソースを一括停止するコマンド",
	Long: `CloudFormationスタック内の起動・停止可能なリソースを一括停止します。
対象リソース: EC2インスタンス、RDSインスタンス、Aurora DBクラスター、ECSサービス

例:
  ` + AppName + ` cfn stop -S my-stack -P my-profile`,
	RunE: func(cmd *cobra.Command, args []string) error {
		stackName, _ := cmd.Flags().GetString("stack")
		if stackName == "" {
			return fmt.Errorf("❌ エラー: スタック名 (-S) を指定してください")
		}

		fmt.Printf("Profile: %s\n", awsCtx.Profile)
		fmt.Printf("Region: %s\n", awsCtx.Region)
		fmt.Printf("Stack: %s\n", stackName)

		// 各種クライアントを作成
		cfnClient, err := aws.NewClient[*cloudformation.Client](awsCtx)
		if err != nil {
			return fmt.Errorf("CloudFormationクライアント作成エラー: %w", err)
		}

		ec2Client, err := aws.NewClient[*ec2.Client](awsCtx)
		if err != nil {
			return fmt.Errorf("EC2クライアント作成エラー: %w", err)
		}

		rdsClient, err := aws.NewClient[*rds.Client](awsCtx)
		if err != nil {
			return fmt.Errorf("RDSクライアント作成エラー: %w", err)
		}

		autoScalingClient, err := aws.NewClient[*applicationautoscaling.Client](awsCtx)
		if err != nil {
			return fmt.Errorf("ApplicationAutoScalingクライアント作成エラー: %w", err)
		}

		// stop用のオプションを作成
		stopOpts := cfn.StackStartStopOptions{
			CfnClient:                    cfnClient,
			Ec2Client:                    ec2Client,
			RdsClient:                    rdsClient,
			ApplicationAutoScalingClient: autoScalingClient,
			StackName:                    stackName,
		}

		err = cfn.StopAllStackResources(stopOpts)
		if err != nil {
			return fmt.Errorf("❌ リソース停止処理でエラー: %w", err)
		}

		fmt.Println("✅ リソース停止処理が完了しました")
		return nil
	},
	SilenceUsage: true,
}

var cfnCleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "CloudFormationスタック内のS3/ECRリソースを削除",
	Long: `指定したCloudFormationスタック内のS3バケットとECRリポジトリを削除します。

例:
  ` + AppName + ` cfn cleanup -S my-stack -P my-profile`,
	RunE: func(cmd *cobra.Command, args []string) error {
		stackName, _ := cmd.Flags().GetString("stack")
		if stackName == "" {
			return fmt.Errorf("❌ エラー: スタック名 (-S) を指定してください")
		}

		fmt.Printf("Profile: %s\n", awsCtx.Profile)
		fmt.Printf("Region: %s\n", awsCtx.Region)
		fmt.Printf("Stack: %s\n", stackName)

		// 各種クライアントを作成
		cfnClient, err := aws.NewClient[*cloudformation.Client](awsCtx)
		if err != nil {
			return fmt.Errorf("CloudFormationクライアント作成エラー: %w", err)
		}

		s3Client, err := aws.NewClient[*s3.Client](awsCtx)
		if err != nil {
			return fmt.Errorf("S3クライアント作成エラー: %w", err)
		}

		ecrClient, err := aws.NewClient[*ecr.Client](awsCtx)
		if err != nil {
			return fmt.Errorf("ECRクライアント作成エラー: %w", err)
		}

		// スタックからリソース情報を取得
		fmt.Println("スタックに関連するリソースの削除を開始します...")
		s3BucketNames, ecrRepoNames, err := cfn.GetCleanupResourcesFromStack(cfnClient, stackName)
		if err != nil {
			return fmt.Errorf("スタックからのリソース取得エラー: %w", err)
		}

		// S3バケットの削除
		if len(s3BucketNames) > 0 {
			fmt.Println("S3バケットの削除を開始...")
			err = s3svc.CleanupS3Buckets(s3Client, s3BucketNames)
			if err != nil {
				fmt.Printf("❌ S3バケットのクリーンアップ中にエラーが発生しました: %v\n", err)
			}
		} else {
			fmt.Println("スタックに関連するS3バケットは見つかりませんでした。")
		}

		// ECRリポジトリの削除
		if len(ecrRepoNames) > 0 {
			fmt.Println("ECRリポジトリの削除を開始...")
			err = ecrsvc.CleanupEcrRepositories(ecrClient, ecrRepoNames)
			if err != nil {
				fmt.Printf("❌ ECRリポジトリのクリーンアップ中にエラーが発生しました: %v\n", err)
			}
		} else {
			fmt.Println("スタックに関連するECRリポジトリは見つかりませんでした。")
		}

		fmt.Println("✅ クリーンアップが完了しました")
		return nil
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(CfnCmd)
	CfnCmd.AddCommand(cfnLsCmd)
	CfnCmd.AddCommand(cfnStartCmd)
	CfnCmd.AddCommand(cfnStopCmd)
	CfnCmd.AddCommand(cfnCleanupCmd)

	// cfn start/stopコマンド用のフラグ
	cfnStartCmd.Flags().StringP("stack", "S", "", "CloudFormationスタック名")
	cfnStopCmd.Flags().StringP("stack", "S", "", "CloudFormationスタック名")

	// cfn cleanupコマンド用のフラグ
	cfnCleanupCmd.Flags().StringP("stack", "S", "", "CloudFormationスタック名")
}
