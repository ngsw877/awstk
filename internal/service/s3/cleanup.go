package s3

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// GetS3BucketsByKeyword はキーワードに一致するS3バケット名の一覧を取得します
func GetS3BucketsByKeyword(s3Client *s3.Client, searchString string) ([]string, error) {
	// バケット一覧を取得
	listBucketsOutput, err := s3Client.ListBuckets(context.Background(), &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("S3バケット一覧取得エラー: %w", err)
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
	}
	return nil
}

// emptyS3Bucket は指定したS3バケットの中身をすべて削除します (バージョン管理対応)
func emptyS3Bucket(s3Client *s3.Client, bucketName string) error {
	// バケット内のオブジェクトとバージョンをリスト
	listVersionsOutput, err := s3Client.ListObjectVersions(context.Background(), &s3.ListObjectVersionsInput{
		Bucket: aws.String(bucketName),
	})
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
		_, err = s3Client.DeleteObjects(context.Background(), &s3.DeleteObjectsInput{
			Bucket: aws.String(bucketName),
			Delete: &types.Delete{
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
	remainingObjects, err := s3Client.ListObjectVersions(context.Background(), &s3.ListObjectVersionsInput{
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
