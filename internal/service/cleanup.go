package service

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// CleanupOptions はクリーンアップ処理のパラメータを格納する構造体
type CleanupOptions struct {
	S3Client     *s3.Client
	EcrClient    *ecr.Client
	CfnClient    *cloudformation.Client
	SearchString string // 検索文字列
	StackName    string // CloudFormationスタック名
}

// CleanupResources は指定した文字列を含むAWSリソースをクリーンアップします
func CleanupResources(opts CleanupOptions) error {
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

		// スタックからリソース情報を取得
		s3BucketNames, ecrRepoNames, err = getCleanupResourcesFromStack(opts.CfnClient, opts.StackName)
		if err != nil {
			return fmt.Errorf("スタックからのリソース取得エラー: %w", err)
		}
	} else {
		// キーワードから検索する場合
		fmt.Printf("検索文字列: %s\n", opts.SearchString)
		fmt.Println("検索文字列に一致するリソースの削除を開始します...")

		// S3バケット名を取得
		s3BucketNames, err = getS3BucketsByKeyword(opts.S3Client, opts.SearchString)
		if err != nil {
			fmt.Printf("❌ S3バケット一覧取得中にエラーが発生しました: %v\n", err)
			// エラーが発生しても続行
			s3BucketNames = []string{} // 空のリストで初期化
		}

		// ECRリポジトリ名を取得
		ecrRepoNames, err = getEcrRepositoriesByKeyword(opts.EcrClient, opts.SearchString)
		if err != nil {
			fmt.Printf("❌ ECRリポジトリ一覧取得中にエラーが発生しました: %v\n", err)
			// エラーが発生しても続行
			ecrRepoNames = []string{} // 空のリストで初期化
		}
	}

	// S3バケットの削除（共通処理）
	fmt.Println("S3バケットの削除を開始...")
	if len(s3BucketNames) > 0 {
		err = cleanupS3Buckets(opts.S3Client, s3BucketNames)
		if err != nil {
			fmt.Printf("❌ S3バケットのクリーンアップ中にエラーが発生しました: %v\n", err)
		}
	} else {
		if opts.StackName != "" {
			fmt.Println("スタックに関連するS3バケットは見つかりませんでした。")
		} else {
			fmt.Printf("  検索文字列 '%s' にマッチするS3バケットは見つかりませんでした。\n", opts.SearchString)
		}
	}

	// ECRリポジトリの削除（共通処理）
	fmt.Println("ECRリポジトリの削除を開始...")
	if len(ecrRepoNames) > 0 {
		err = cleanupEcrRepositories(opts.EcrClient, ecrRepoNames)
		if err != nil {
			fmt.Printf("❌ ECRリポジトリのクリーンアップ中にエラーが発生しました: %v\n", err)
		}
	} else {
		if opts.StackName != "" {
			fmt.Println("スタックに関連するECRリポジトリは見つかりませんでした。")
		} else {
			fmt.Printf("  検索文字列 '%s' にマッチするECRリポジトリは見つかりませんでした。\n", opts.SearchString)
		}
	}

	fmt.Println("クリーンアップ完了！")
	return nil
}

// ValidateCleanupOptions はクリーンアップオプションのバリデーションを行います
func validateCleanupOptions(opts CleanupOptions) error {
	// クライアントのnilチェック
	if opts.S3Client == nil {
		return fmt.Errorf("S3クライアントが指定されていません")
	}
	if opts.EcrClient == nil {
		return fmt.Errorf("ECRクライアントが指定されていません")
	}
	if opts.CfnClient == nil {
		return fmt.Errorf("CloudFormationクライアントが指定されていません")
	}

	// キーワードとスタック名の両方が指定された場合はエラー
	if opts.SearchString != "" && opts.StackName != "" {
		return fmt.Errorf("検索キーワードとスタック名は同時に指定できません。いずれか一方を指定してください")
	}

	// 少なくとも一方が指定されている必要がある
	if opts.SearchString == "" && opts.StackName == "" {
		return fmt.Errorf("検索キーワードまたはスタック名のいずれかを指定してください")
	}

	return nil
}
