package internal

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	ecrtypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// CleanupOptions はクリーンアップ処理のパラメータを格納する構造体
type CleanupOptions struct {
	SearchString string
	Region       string
	Profile      string
}

// CleanupResources は指定した文字列を含むAWSリソースをクリーンアップします
func CleanupResources(opts CleanupOptions) error {
	fmt.Printf("AWS Profile: %s\n", opts.Profile)
	fmt.Printf("検索文字列: %s\n", opts.SearchString)
	fmt.Println("削除を開始します...")

	// S3バケットの削除
	fmt.Println("S3バケットの削除を開始...")
	err := cleanupS3Buckets(opts)
	if err != nil {
		// エラーが発生しても、他のリソースのクリーンアップは続行するためにエラーを返さない
		fmt.Printf("❌ S3バケットのクリーンアップ中にエラーが発生しました: %v\n", err)
	}

	// ECRリポジトリの削除
	fmt.Println("ECRリポジトリの削除を開始...")
	err = cleanupEcrRepositories(opts)
	if err != nil {
		// エラーが発生しても、他のリソースのクリーンアップは続行するためにエラーを返さない
		fmt.Printf("❌ ECRリポジトリのクリーンアップ中にエラーが発生しました: %v\n", err)
	}

	fmt.Println("クリーンアップ完了！")
	return nil
}

// cleanupS3Buckets は指定した文字列を含むS3バケットを検索・削除します (パッケージプライベート)
func cleanupS3Buckets(opts CleanupOptions) error {
	cfg, err := LoadAwsConfig(opts.Region, opts.Profile)
	if err != nil {
		return fmt.Errorf("AWS設定の読み込みエラー: %w", err)
	}

	// S3クライアントを作成
	s3Client := s3.NewFromConfig(cfg)

	// バケット一覧を取得
	listBucketsOutput, err := s3Client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		return fmt.Errorf("S3バケット一覧取得エラー: %w", err)
	}

	foundBuckets := []string{}
	for _, bucket := range listBucketsOutput.Buckets {
		if strings.Contains(*bucket.Name, opts.SearchString) {
			foundBuckets = append(foundBuckets, *bucket.Name)
		}
	}

	if len(foundBuckets) == 0 {
		fmt.Printf("  検索文字列 '%s' にマッチするS3バケットは見つかりませんでした。\n", opts.SearchString)
		return nil
	}

	for _, bucket := range foundBuckets {
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

// cleanupEcrRepositories は指定した文字列を含むECRリポジトリを検索・削除します (パッケージプライベート)
func cleanupEcrRepositories(opts CleanupOptions) error {
	cfg, err := LoadAwsConfig(opts.Region, opts.Profile)
	if err != nil {
		return fmt.Errorf("AWS設定の読み込みエラー: %w", err)
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
			return fmt.Errorf("ECRリポジトリ一覧取得エラー: %w", err)
		}

		for _, repo := range listReposOutput.Repositories {
			if strings.Contains(*repo.RepositoryName, opts.SearchString) {
				foundRepos = append(foundRepos, *repo.RepositoryName)
			}
		}

		if listReposOutput.NextToken == nil {
			break
		}
		listReposInput.NextToken = listReposOutput.NextToken
	}

	if len(foundRepos) == 0 {
		fmt.Printf("  検索文字列 '%s' にマッチするECRリポジトリは見つかりませんでした。\n", opts.SearchString)
		return nil
	}

	for _, repoName := range foundRepos {
		fmt.Printf("リポジトリ %s を空にして削除中...\n", repoName)

		// リポジトリ内のイメージをすべて削除 (ページネーション対応)
		listImagesInput := &ecr.ListImagesInput{
			RepositoryName: aws.String(repoName),
		}
		imageIDsToDelete := []ecrtypes.ImageIdentifier{}

		for {
			listImagesOutput, err := ecrClient.ListImages(context.TODO(), listImagesInput)
			if err != nil {
				// イメージリスト取得エラーはこのリポジトリをスキップ
				fmt.Printf("❌ リポジトリ %s のイメージ一覧取得エラー: %v\n", repoName, err)
				break // このリポジトリの処理を中断
			}

			for _, imageId := range listImagesOutput.ImageIds {
				imageIDsToDelete = append(imageIDsToDelete, imageId)
			}

			if listImagesOutput.NextToken == nil {
				break
			}
			listImagesInput.NextToken = listImagesOutput.NextToken
		}

		// イメージ削除対象がなければスキップ
		if len(imageIDsToDelete) > 0 {
			// イメージを一括削除 (最大100個ずつ)
			chunkSize := 100
			for i := 0; i < len(imageIDsToDelete); i += chunkSize {
				end := i + chunkSize
				if end > len(imageIDsToDelete) {
					end = len(imageIDsToDelete)
				}
				batch := imageIDsToDelete[i:end]

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
