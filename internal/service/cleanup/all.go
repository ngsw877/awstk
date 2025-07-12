package cleanup

import (
	"awstk/internal/service/cfn"
	ecrsvc "awstk/internal/service/ecr"
	s3svc "awstk/internal/service/s3"
	"fmt"
)

// CleanupResources は指定した文字列を含むAWSリソースをクリーンアップします
func CleanupResources(opts Options) error {
	// 事前条件チェック
	if err := validateCleanupOptions(opts); err != nil {
		return err
	}

	var s3BucketNames, ecrRepoNames []string
	var err error

	// 検索方法によって取得ロジックを分岐
	if opts.StackName != "" {
		// スタック名から検索する場合
		fmt.Printf("CloudFormationスタック: %s\n", opts.StackName)
		fmt.Println("スタックに関連するリソースの削除を開始します...")

		s3BucketNames, ecrRepoNames, err = cfn.GetCleanupResourcesFromStack(opts.CfnClient, opts.StackName)
		if err != nil {
			return fmt.Errorf("スタックからのリソース取得エラー: %w", err)
		}
	} else {
		// キーワードから検索する場合
		fmt.Printf("検索文字列: %s\n", opts.SearchString)
		fmt.Println("検索文字列に一致するリソースの削除を開始します...")

		s3BucketNames, err = s3svc.GetS3BucketsByKeyword(opts.S3Client, opts.SearchString)
		if err != nil {
			fmt.Printf("❌ S3バケット一覧取得中にエラーが発生しました: %v\n", err)
			s3BucketNames = []string{}
		}

		ecrRepoNames, err = ecrsvc.GetEcrRepositoriesByKeyword(opts.EcrClient, opts.SearchString)
		if err != nil {
			fmt.Printf("❌ ECRリポジトリ一覧取得中にエラーが発生しました: %v\n", err)
			ecrRepoNames = []string{}
		}
	}

	// S3バケットの削除
	fmt.Println("S3バケットの削除を開始...")
	if len(s3BucketNames) > 0 {
		if err := s3svc.CleanupS3Buckets(opts.S3Client, s3BucketNames); err != nil {
			fmt.Printf("❌ S3バケットのクリーンアップ中にエラーが発生しました: %v\n", err)
		}
	} else {
		fmt.Println("  削除対象のS3バケットはありません")
	}

	// ECRリポジトリの削除
	fmt.Println("ECRリポジトリの削除を開始...")
	if len(ecrRepoNames) > 0 {
		if err := ecrsvc.CleanupEcrRepositories(opts.EcrClient, ecrRepoNames); err != nil {
			fmt.Printf("❌ ECRリポジトリのクリーンアップ中にエラーが発生しました: %v\n", err)
		}
	} else {
		fmt.Println("  削除対象のECRリポジトリはありません")
	}

	fmt.Println("🎉 クリーンアップ完了！")
	return nil
}

// validateCleanupOptions はクリーンアップオプションのバリデーションを行います
func validateCleanupOptions(opts Options) error {
	if opts.S3Client == nil {
		return fmt.Errorf("S3クライアントが指定されていません")
	}
	if opts.EcrClient == nil {
		return fmt.Errorf("ECRクライアントが指定されていません")
	}
	if opts.CfnClient == nil {
		return fmt.Errorf("CloudFormationクライアントが指定されていません")
	}

	if opts.SearchString != "" && opts.StackName != "" {
		return fmt.Errorf("検索キーワードとスタック名は同時に指定できません。いずれか一方を指定してください")
	}
	if opts.SearchString == "" && opts.StackName == "" {
		return fmt.Errorf("検索キーワードまたはスタック名のいずれかを指定してください")
	}
	return nil
}
