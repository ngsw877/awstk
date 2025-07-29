package s3

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// GetS3BucketsByFilter はフィルターに一致するS3バケット名の一覧を取得します
func GetS3BucketsByFilter(s3Client *s3.Client, searchString string) ([]string, error) {
	// バケット一覧を取得
	listBucketsOutput, err := s3Client.ListBuckets(context.Background(), &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("s3バケット一覧取得エラー: %w", err)
	}

	foundBuckets := []string{}
	for _, bucket := range listBucketsOutput.Buckets {
		if strings.Contains(*bucket.Name, searchString) {
			foundBuckets = append(foundBuckets, *bucket.Name)
			fmt.Printf("🔍 検出されたS3バケット: %s\n", *bucket.Name)
		}
	}

	return foundBuckets, nil
}

// CleanupS3Buckets は指定したS3バケット一覧を削除します
func CleanupS3Buckets(s3Client *s3.Client, bucketNames []string) error {
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
		_, err = s3Client.DeleteBucket(context.Background(), &s3.DeleteBucketInput{
			Bucket: aws.String(bucket),
		})
		if err != nil {
			fmt.Printf("❌ バケット %s の削除に失敗しました: %v\n", bucket, err)
			// このバケットの削除はスキップし、次のバケットへ
			continue
		}
		fmt.Printf("✅ バケット %s を削除しました\n", bucket)
	}
	return nil
}

// emptyS3Bucket は指定したS3バケットの中身をすべて削除します (バージョン管理対応)
func emptyS3Bucket(s3Client *s3.Client, bucketName string) error {
	// ページネーション対応のループ
	var keyMarker *string
	var versionIdMarker *string

	for {
		// バケット内のオブジェクトとバージョンをリスト
		listVersionsInput := &s3.ListObjectVersionsInput{
			Bucket: aws.String(bucketName),
		}
		if keyMarker != nil {
			listVersionsInput.KeyMarker = keyMarker
			listVersionsInput.VersionIdMarker = versionIdMarker
		}

		listVersionsOutput, err := s3Client.ListObjectVersions(context.Background(), listVersionsInput)
		if err != nil {
			return fmt.Errorf("バケット内のオブジェクトバージョン一覧取得エラー: %w", err)
		}

		// 削除対象のオブジェクトと削除マーカーのリストを作成
		deleteObjects := []types.ObjectIdentifier{}
		if listVersionsOutput.Versions != nil {
			for _, version := range listVersionsOutput.Versions {
				deleteObjects = append(deleteObjects, types.ObjectIdentifier{
					Key:       version.Key,
					VersionId: version.VersionId,
				})
			}
		}
		if listVersionsOutput.DeleteMarkers != nil {
			for _, marker := range listVersionsOutput.DeleteMarkers {
				deleteObjects = append(deleteObjects, types.ObjectIdentifier{
					Key:       marker.Key,
					VersionId: marker.VersionId,
				})
			}
		}

		// 削除対象がある場合は削除
		if len(deleteObjects) > 0 {
			// オブジェクトを一括削除 (最大1000個ずつ)
			chunkSize := 1000
			for i := 0; i < len(deleteObjects); i += chunkSize {
				end := i + chunkSize
				if end > len(deleteObjects) {
					end = len(deleteObjects)
				}
				batch := deleteObjects[i:end]

				fmt.Printf("  %d件のオブジェクトを削除中...\n", len(batch))
				deleteOutput, err := s3Client.DeleteObjects(context.Background(), &s3.DeleteObjectsInput{
					Bucket: aws.String(bucketName),
					Delete: &types.Delete{
						Objects: batch,
						Quiet:   aws.Bool(false),
					},
				})
				if err != nil {
					return fmt.Errorf("オブジェクトの一括削除エラー: %w", err)
				}

				// 削除エラーがあった場合は警告を表示
				if len(deleteOutput.Errors) > 0 {
					for _, deleteErr := range deleteOutput.Errors {
						fmt.Printf("  ⚠️  オブジェクト削除エラー: %s (バージョンID: %s) - %s\n",
							*deleteErr.Key,
							aws.ToString(deleteErr.VersionId),
							aws.ToString(deleteErr.Message))
					}
				}
			}
		}

		// 次のページがない場合は終了
		if !aws.ToBool(listVersionsOutput.IsTruncated) {
			break
		}

		// 次のページのマーカーを設定
		keyMarker = listVersionsOutput.NextKeyMarker
		versionIdMarker = listVersionsOutput.NextVersionIdMarker
	}

	fmt.Println("  バケットを空にしました。")
	return nil
}
