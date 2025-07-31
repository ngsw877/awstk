package cleanup

import (
	"awstk/internal/service/cfn"
	ecrsvc "awstk/internal/service/ecr"
	logssvc "awstk/internal/service/logs"
	s3svc "awstk/internal/service/s3"
	"fmt"
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
	if opts.StackName != "" {
		// スタック名から検索する場合
		fmt.Printf("CloudFormationスタック: %s\n", opts.StackName)
		fmt.Println("スタックに関連するリソースの削除を開始します...")

		s3BucketNames, ecrRepoNames, err = cfn.GetCleanupResourcesFromStack(clients.CfnClient, opts.StackName)
		if err != nil {
			return fmt.Errorf("スタックからのリソース取得エラー: %w", err)
		}
		// スタックからの削除では現時点でCloudWatch Logsは対象外
		logGroupNames = []string{}
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

	// S3バケットの削除
	fmt.Println("S3バケットの削除を開始...")
	if len(s3BucketNames) > 0 {
		if err := s3svc.CleanupS3Buckets(clients.S3Client, s3BucketNames); err != nil {
			fmt.Printf("❌ S3バケットのクリーンアップ中にエラーが発生しました: %v\n", err)
		}
	} else {
		fmt.Println("  削除対象のS3バケットはありません")
	}

	// ECRリポジトリの削除
	fmt.Println("ECRリポジトリの削除を開始...")
	if len(ecrRepoNames) > 0 {
		if err := ecrsvc.CleanupEcrRepositories(clients.EcrClient, ecrRepoNames); err != nil {
			fmt.Printf("❌ ECRリポジトリのクリーンアップ中にエラーが発生しました: %v\n", err)
		}
	} else {
		fmt.Println("  削除対象のECRリポジトリはありません")
	}

	// CloudWatch Logsグループの削除
	fmt.Println("CloudWatch Logsグループの削除を開始...")
	if len(logGroupNames) > 0 {
		if err := logssvc.CleanupLogGroups(clients.LogsClient, logGroupNames); err != nil {
			fmt.Printf("❌ CloudWatch Logsグループのクリーンアップ中にエラーが発生しました: %v\n", err)
		}
	} else {
		fmt.Println("  削除対象のCloudWatch Logsグループはありません")
	}

	fmt.Println("🎉 クリーンアップ完了！")
	return nil
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
	if opts.SearchString != "" && opts.StackName != "" {
		return fmt.Errorf("検索キーワードとスタック名は同時に指定できません。いずれか一方を指定してください")
	}
	if opts.SearchString == "" && opts.StackName == "" {
		return fmt.Errorf("検索キーワードまたはスタック名のいずれかを指定してください")
	}
	return nil
}
