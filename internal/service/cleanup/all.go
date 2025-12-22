package cleanup

import (
	"awstk/internal/service/cfn"
	"awstk/internal/service/common"
	ecrsvc "awstk/internal/service/ecr"
	logssvc "awstk/internal/service/logs"
	s3svc "awstk/internal/service/s3"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/smithy-go"
)

// CleanupResources は指定した文字列を含むAWSリソースをクリーンアップします
func CleanupResources(clients ClientSet, opts Options) error {
	// 事前条件チェック
	if err := validateCleanupOptions(clients); err != nil {
		return err
	}
	if err := validateOptions(opts); err != nil {
		return err
	}

	var s3BucketNames, ecrRepoNames, logGroupNames []string
	var err error

	// 検索方法によって取得ロジックを分岐
	if opts.StackId != "" {
		// スタックIDから検索する場合
		fmt.Printf("CloudFormationスタックID: %s\n", opts.StackId)
		fmt.Println("スタックに関連するリソースの削除を開始します...")

		s3BucketNames, ecrRepoNames, logGroupNames, err = cfn.GetCleanupResourcesFromStack(clients.CfnClient, opts.StackId)
		if err != nil {
			return fmt.Errorf("スタックからのリソース取得エラー: %w", err)
		}
	} else if opts.StackName != "" {
		// スタック名から検索する場合
		fmt.Printf("CloudFormationスタック: %s\n", opts.StackName)
		fmt.Println("スタックに関連するリソースの削除を開始します...")

		s3BucketNames, ecrRepoNames, logGroupNames, err = cfn.GetCleanupResourcesFromStack(clients.CfnClient, opts.StackName)
		if err != nil {
			var apiErr smithy.APIError
			if errors.As(err, &apiErr) && apiErr.ErrorCode() == "ValidationError" && strings.Contains(apiErr.ErrorMessage(), "does not exist") {
				fmt.Printf("❌ スタック '%s' は存在しません。\n", opts.StackName)
				fmt.Println("ℹ️ 削除済みスタックの履歴が90日以内にある場合、--stack-id に削除済みスタックのID(ARN)を指定してください")
				return fmt.Errorf("スタック '%s' が見つかりません", opts.StackName)
			}
			return fmt.Errorf("スタックからのリソース取得エラー: %w", err)
		}
	} else {
		// キーワードから検索する場合
		fmt.Printf("検索文字列: %s\n", opts.SearchString)
		fmt.Println("検索文字列に一致するリソースの削除を開始します...")

		s3BucketNames, err = s3svc.GetS3BucketsByFilter(clients.S3Client, opts.SearchString)
		if err != nil {
			fmt.Printf("❌ S3バケット一覧取得中にエラーが発生しました: %v\n", err)
			s3BucketNames = []string{}
		}

		ecrRepoNames, err = ecrsvc.GetEcrRepositoriesByFilter(clients.EcrClient, opts.SearchString)
		if err != nil {
			fmt.Printf("❌ ECRリポジトリ一覧取得中にエラーが発生しました: %v\n", err)
			ecrRepoNames = []string{}
		}

		logGroupNames, err = logssvc.GetLogGroupsByFilter(clients.LogsClient, opts.SearchString)
		if err != nil {
			fmt.Printf("❌ CloudWatch Logsグループ一覧取得中にエラーが発生しました: %v\n", err)
			logGroupNames = []string{}
		}
	}

	// 結果を格納するスライス
	var results []common.CleanupResult

	// S3バケットの削除
	fmt.Println("S3バケットの削除を開始...")
	if len(s3BucketNames) > 0 {
		s3Result := s3svc.CleanupS3Buckets(clients.S3Client, s3BucketNames)
		results = append(results, s3Result)
	} else {
		fmt.Println("  削除対象のS3バケットはありません")
	}

	// ECRリポジトリの削除
	fmt.Println("ECRリポジトリの削除を開始...")
	if len(ecrRepoNames) > 0 {
		ecrResult := ecrsvc.CleanupEcrRepositories(clients.EcrClient, ecrRepoNames)
		results = append(results, ecrResult)
	} else {
		fmt.Println("  削除対象のECRリポジトリはありません")
	}

	// CloudWatch Logsグループの削除
	fmt.Println("CloudWatch Logsグループの削除を開始...")
	if len(logGroupNames) > 0 {
		logsResult := logssvc.CleanupLogGroups(clients.LogsClient, logGroupNames)
		results = append(results, logsResult)
	} else {
		fmt.Println("  削除対象のCloudWatch Logsグループはありません")
	}

	// サマリー表示
	printCleanupSummary(results)

	return nil
}

// printCleanupSummary はクリーンアップ結果のサマリーを表示します
func printCleanupSummary(results []common.CleanupResult) {
	fmt.Println()
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("                    クリーンアップ サマリー")
	fmt.Println("════════════════════════════════════════════════════════════")

	totalDeleted := 0
	totalFailed := 0

	for _, result := range results {
		if result.TotalCount() == 0 {
			continue
		}

		fmt.Printf("\n【%s】\n", result.ResourceType)

		if len(result.Deleted) > 0 {
			fmt.Printf("  ✅ 削除成功: %d件\n", len(result.Deleted))
			for _, name := range result.Deleted {
				fmt.Printf("     - %s\n", name)
			}
		}

		if len(result.Failed) > 0 {
			fmt.Printf("  ❌ 削除失敗: %d件\n", len(result.Failed))
			for _, name := range result.Failed {
				fmt.Printf("     - %s\n", name)
			}
		}

		totalDeleted += len(result.Deleted)
		totalFailed += len(result.Failed)
	}

	fmt.Println()
	fmt.Println("────────────────────────────────────────────────────────────")
	fmt.Printf("合計: 削除成功 %d件 / 削除失敗 %d件\n", totalDeleted, totalFailed)
	fmt.Println("════════════════════════════════════════════════════════════")
}

// validateCleanupOptions はクリーンアップオプションのバリデーションを行います
func validateCleanupOptions(clients ClientSet) error {
	if clients.S3Client == nil {
		return fmt.Errorf("s3クライアントが指定されていません")
	}
	if clients.EcrClient == nil {
		return fmt.Errorf("ecrクライアントが指定されていません")
	}
	if clients.CfnClient == nil {
		return fmt.Errorf("cloudFormationクライアントが指定されていません")
	}
	if clients.LogsClient == nil {
		return fmt.Errorf("cloudWatchLogsクライアントが指定されていません")
	}
	return nil
}

// validateOptions はオプションの論理バリデーションを行います
func validateOptions(opts Options) error {
	count := 0
	if opts.SearchString != "" {
		count++
	}
	if opts.StackName != "" {
		count++
	}
	if opts.StackId != "" {
		count++
	}
	if count == 0 {
		return fmt.Errorf("検索キーワード、スタック名、またはスタックIDのいずれかを指定してください")
	}
	if count > 1 {
		return fmt.Errorf("検索キーワード、スタック名、スタックIDは同時に指定できません。いずれか一つのみ指定してください")
	}
	return nil
}
