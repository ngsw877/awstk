package internal

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cftypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	ecrtypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// CleanupOptions はクリーンアップ処理のパラメータを格納する構造体
type CleanupOptions struct {
	SearchString string // 検索文字列
	StackName    string // CloudFormationスタック名
	Region       string
	Profile      string
}

// ValidateCleanupOptions はクリーンアップオプションのバリデーションを行います
func ValidateCleanupOptions(opts CleanupOptions) error {
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

// CleanupResources は指定した文字列を含むAWSリソースをクリーンアップします
func CleanupResources(opts CleanupOptions) error {
	// 事前条件チェック
	if err := ValidateCleanupOptions(opts); err != nil {
		return err
	}

	fmt.Printf("AWS Profile: %s\n", opts.Profile)

	var s3BucketNames, ecrRepoNames []string
	var err error

	// 検索方法によって取得ロジックを分岐
	if opts.StackName != "" {
		// スタック名から検索する場合
		fmt.Printf("CloudFormationスタック: %s\n", opts.StackName)
		fmt.Println("スタックに関連するリソースの削除を開始します...")

		// スタックからリソース情報を取得
		s3BucketNames, ecrRepoNames, err = getResourcesFromStack(opts)
		if err != nil {
			return fmt.Errorf("スタックからのリソース取得エラー: %w", err)
		}
	} else {
		// キーワードから検索する場合
		fmt.Printf("検索文字列: %s\n", opts.SearchString)
		fmt.Println("検索文字列に一致するリソースの削除を開始します...")

		// S3バケット名を取得
		s3BucketNames, err = getS3BucketsByKeyword(opts)
		if err != nil {
			fmt.Printf("❌ S3バケット一覧取得中にエラーが発生しました: %v\n", err)
			// エラーが発生しても続行
			s3BucketNames = []string{} // 空のリストで初期化
		}

		// ECRリポジトリ名を取得
		ecrRepoNames, err = getEcrRepositoriesByKeyword(opts)
		if err != nil {
			fmt.Printf("❌ ECRリポジトリ一覧取得中にエラーが発生しました: %v\n", err)
			// エラーが発生しても続行
			ecrRepoNames = []string{} // 空のリストで初期化
		}
	}

	// S3バケットの削除（共通処理）
	fmt.Println("S3バケットの削除を開始...")
	if len(s3BucketNames) > 0 {
		err = cleanupS3Buckets(opts, s3BucketNames)
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
		err = cleanupEcrRepositories(opts, ecrRepoNames)
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

// getResourcesFromStack はCloudFormationスタックからS3バケットとECRリポジトリのリソース一覧を取得します
func getResourcesFromStack(opts CleanupOptions) ([]string, []string, error) {
	cfg, err := LoadAwsConfig(opts.Region, opts.Profile)
	if err != nil {
		return nil, nil, fmt.Errorf("AWS設定の読み込みエラー: %w", err)
	}

	// CloudFormationクライアントを作成
	cfnClient := cloudformation.NewFromConfig(cfg)

	// スタックリソース一覧の取得
	stackResources := []cftypes.StackResourceSummary{}
	var nextToken *string

	for {
		resp, err := cfnClient.ListStackResources(context.TODO(), &cloudformation.ListStackResourcesInput{
			StackName: aws.String(opts.StackName),
			NextToken: nextToken,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("スタックリソース一覧取得エラー: %w", err)
		}

		stackResources = append(stackResources, resp.StackResourceSummaries...)

		if resp.NextToken == nil {
			break
		}
		nextToken = resp.NextToken
	}

	// S3バケットとECRリポジトリを抽出
	s3Resources := []string{}
	ecrResources := []string{}

	for _, resource := range stackResources {
		// リソースタイプに基づいて振り分け
		resourceType := *resource.ResourceType

		// S3バケット
		if resourceType == "AWS::S3::Bucket" && resource.PhysicalResourceId != nil {
			s3Resources = append(s3Resources, *resource.PhysicalResourceId)
			fmt.Printf("🔍 検出されたS3バケット: %s\n", *resource.PhysicalResourceId)
		}

		// ECRリポジトリ
		if resourceType == "AWS::ECR::Repository" && resource.PhysicalResourceId != nil {
			ecrResources = append(ecrResources, *resource.PhysicalResourceId)
			fmt.Printf("🔍 検出されたECRリポジトリ: %s\n", *resource.PhysicalResourceId)
		}
	}

	return s3Resources, ecrResources, nil
}

// getS3BucketsByKeyword はキーワードに一致するS3バケット名の一覧を取得します
func getS3BucketsByKeyword(opts CleanupOptions) ([]string, error) {
	cfg, err := LoadAwsConfig(opts.Region, opts.Profile)
	if err != nil {
		return nil, fmt.Errorf("AWS設定の読み込みエラー: %w", err)
	}

	// S3クライアントを作成
	s3Client := s3.NewFromConfig(cfg)

	// バケット一覧を取得
	listBucketsOutput, err := s3Client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("S3バケット一覧取得エラー: %w", err)
	}

	foundBuckets := []string{}
	for _, bucket := range listBucketsOutput.Buckets {
		if strings.Contains(*bucket.Name, opts.SearchString) {
			foundBuckets = append(foundBuckets, *bucket.Name)
			fmt.Printf("🔍 検出されたS3バケット: %s\n", *bucket.Name)
		}
	}

	return foundBuckets, nil
}

// getEcrRepositoriesByKeyword はキーワードに一致するECRリポジトリ名の一覧を取得します
func getEcrRepositoriesByKeyword(opts CleanupOptions) ([]string, error) {
	cfg, err := LoadAwsConfig(opts.Region, opts.Profile)
	if err != nil {
		return nil, fmt.Errorf("AWS設定の読み込みエラー: %w", err)
	}

	// ECRクライアントを作成
	ecrClient := ecr.NewFromConfig(cfg)

	// リポジトリ一覧を取得
	listReposInput := &ecr.DescribeRepositoriesInput{}
	foundRepos := []string{}

	// ページネーション対応
	for {
		listReposOutput, err := ecrClient.DescribeRepositories(context.TODO(), listReposInput)
		if err != nil {
			return nil, fmt.Errorf("ECRリポジトリ一覧取得エラー: %w", err)
		}

		for _, repo := range listReposOutput.Repositories {
			if strings.Contains(*repo.RepositoryName, opts.SearchString) {
				foundRepos = append(foundRepos, *repo.RepositoryName)
				fmt.Printf("🔍 検出されたECRリポジトリ: %s\n", *repo.RepositoryName)
			}
		}

		if listReposOutput.NextToken == nil {
			break
		}
		listReposInput.NextToken = listReposOutput.NextToken
	}

	return foundRepos, nil
}

// cleanupS3Buckets は指定したS3バケット一覧を削除します
func cleanupS3Buckets(opts CleanupOptions, bucketNames []string) error {
	cfg, err := LoadAwsConfig(opts.Region, opts.Profile)
	if err != nil {
		return fmt.Errorf("AWS設定の読み込みエラー: %w", err)
	}

	// S3クライアントを作成
	s3Client := s3.NewFromConfig(cfg)

	for _, bucket := range bucketNames {
		fmt.Printf("バケット %s を空にして削除中...\n", bucket)

		// バケットを空にする (バージョン管理対応)
		err := emptyS3Bucket(s3Client, bucket)
		if err != nil {
			fmt.Printf("❌ バケット %s を空にするのに失敗しました: %v\n", bucket, err)
			// このバケットの削除はスキップし、次のバケットへ
			continue
		}

		// バケットの削除
		fmt.Printf("  バケット削除中: %s\n", bucket)
		_, err = s3Client.DeleteBucket(context.TODO(), &s3.DeleteBucketInput{
			Bucket: aws.String(bucket),
		})
		if err != nil {
			fmt.Printf("❌ バケット %s の削除に失敗しました: %v\n", bucket, err)
			// このバケットの削除はスキップし、次のバケットへ
			continue
		}
	}
	return nil
}

// emptyS3Bucket は指定したS3バケットの中身をすべて削除します (バージョン管理対応) (パッケージプライベート)
func emptyS3Bucket(s3Client *s3.Client, bucketName string) error {
	// バケット内のオブジェクトとバージョンをリスト
	listVersionsOutput, err := s3Client.ListObjectVersions(context.TODO(), &s3.ListObjectVersionsInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return fmt.Errorf("バケット内のオブジェクトバージョン一覧取得エラー: %w", err)
	}

	// 削除対象のオブジェクトと削除マーカーのリストを作成
	deleteObjects := []s3types.ObjectIdentifier{}
	if listVersionsOutput.Versions != nil {
		for _, version := range listVersionsOutput.Versions {
			deleteObjects = append(deleteObjects, s3types.ObjectIdentifier{
				Key:       version.Key,
				VersionId: version.VersionId,
			})
		}
	}
	if listVersionsOutput.DeleteMarkers != nil {
		for _, marker := range listVersionsOutput.DeleteMarkers {
			deleteObjects = append(deleteObjects, s3types.ObjectIdentifier{
				Key:       marker.Key,
				VersionId: marker.VersionId,
			})
		}
	}

	// 削除対象がなければ終了
	if len(deleteObjects) == 0 {
		fmt.Println("  削除するオブジェクトがありません。")
		return nil
	}

	// オブジェクトを一括削除 (最大1000個ずつ)
	chunkSize := 1000
	for i := 0; i < len(deleteObjects); i += chunkSize {
		end := i + chunkSize
		if end > len(deleteObjects) {
			end = len(deleteObjects)
		}
		batch := deleteObjects[i:end]

		fmt.Printf("  %d件のオブジェクトを削除中...\n", len(batch))
		_, err = s3Client.DeleteObjects(context.TODO(), &s3.DeleteObjectsInput{
			Bucket: aws.String(bucketName),
			Delete: &s3types.Delete{
				Objects: batch,
				Quiet:   aws.Bool(false),
			},
		})
		if err != nil {
			return fmt.Errorf("オブジェクトの一括削除エラー: %w", err)
		}
		// TODO: DeleteObjectsのErrorsを確認して処理を検討
	}

	// まだオブジェクトが残っている場合は再帰的に呼び出す（NextToken対応は一旦しない）
	// 簡易的な対応のため、削除後に再度リストして空になるまで繰り返す（非効率だがシンプル）
	// 実際にはListObjectVersionsのNextTokenを使うのが正しいが、今回は簡易実装
	// TODO: ページネーション対応
	time.Sleep(1 * time.Second) // 反映を待つ
	remainingObjects, err := s3Client.ListObjectVersions(context.TODO(), &s3.ListObjectVersionsInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return fmt.Errorf("削除後のオブジェクト確認エラー: %w", err)
	}

	if len(remainingObjects.Versions) > 0 || len(remainingObjects.DeleteMarkers) > 0 {
		// 残っている場合は再度空にする処理を実行（簡易的な再帰）
		// 無限ループにならないように注意が必要だが、ここでは単純化
		return emptyS3Bucket(s3Client, bucketName) // 簡易的な再帰呼び出し
	}

	return nil
}

// cleanupEcrRepositories は指定したECRリポジトリ一覧を削除します
func cleanupEcrRepositories(opts CleanupOptions, repoNames []string) error {
	cfg, err := LoadAwsConfig(opts.Region, opts.Profile)
	if err != nil {
		return fmt.Errorf("AWS設定の読み込みエラー: %w", err)
	}

	// ECRクライアントを作成
	ecrClient := ecr.NewFromConfig(cfg)

	for _, repoName := range repoNames {
		fmt.Printf("リポジトリ %s を空にして削除中...\n", repoName)

		// リポジトリ内のイメージをすべて削除 (ページネーション対応)
		listImagesInput := &ecr.ListImagesInput{
			RepositoryName: aws.String(repoName),
		}
		imageIdsToDelete := []ecrtypes.ImageIdentifier{}

		for {
			listImagesOutput, err := ecrClient.ListImages(context.TODO(), listImagesInput)
			if err != nil {
				// イメージリスト取得エラーはこのリポジトリをスキップ
				fmt.Printf("❌ リポジトリ %s のイメージ一覧取得エラー: %v\n", repoName, err)
				break // このリポジトリの処理を中断
			}

			imageIdsToDelete = append(imageIdsToDelete, listImagesOutput.ImageIds...)

			if listImagesOutput.NextToken == nil {
				break
			}
			listImagesInput.NextToken = listImagesOutput.NextToken
		}

		// イメージ削除対象がなければスキップ
		if len(imageIdsToDelete) > 0 {
			// イメージを一括削除 (最大100個ずつ)
			chunkSize := 100
			for i := 0; i < len(imageIdsToDelete); i += chunkSize {
				end := i + chunkSize
				if end > len(imageIdsToDelete) {
					end = len(imageIdsToDelete)
				}
				batch := imageIdsToDelete[i:end]

				fmt.Printf("  %d件のイメージを削除中...\n", len(batch))
				_, err = ecrClient.BatchDeleteImage(context.TODO(), &ecr.BatchDeleteImageInput{
					RepositoryName: aws.String(repoName),
					ImageIds:       batch,
				})
				if err != nil {
					fmt.Printf("❌ リポジトリ %s のイメージ一括削除エラー: %v\n", repoName, err)
					// イメージ削除エラーでもリポジトリ削除は試みる
				}
			}
		} else {
			fmt.Println("  削除するイメージがありません。")
		}

		// リポジトリの削除
		fmt.Printf("  リポジトリ削除中: %s\n", repoName)
		_, err = ecrClient.DeleteRepository(context.TODO(), &ecr.DeleteRepositoryInput{
			RepositoryName: aws.String(repoName),
			Force:          true, // 強制削除
		})
		if err != nil {
			fmt.Printf("❌ リポジトリ %s の削除に失敗しました: %v\n", repoName, err)
			// このリポジトリの削除はスキップし、次のリポジトリへ
			continue
		}
	}

	return nil
}
