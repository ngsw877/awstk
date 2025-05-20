package internal

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// ListS3Buckets はS3バケット名の一覧を返す関数
func ListS3Buckets(region, profile string) ([]string, error) {
	cfg, err := LoadAwsConfig(region, profile)
	if err != nil {
		return nil, err
	}

	// Create an Amazon S3 service client
	client := s3.NewFromConfig(cfg)

	// List all S3 buckets
	result, err := client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}

	buckets := make([]string, 0, len(result.Buckets))
	for _, bucket := range result.Buckets {
		buckets = append(buckets, *bucket.Name)
	}
	return buckets, nil
}

// getS3BucketsByKeyword はキーワードに一致す���S3バケット名の一覧を取得します
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

// emptyS3Bucket は指定したS3バケットの中身をすべて削除します (バージョン管理対応)
func emptyS3Bucket(s3Client *s3.Client, bucketName string) error {
	// バケット内のオブジェクトとバージョンをリスト
	listVersionsOutput, err := s3Client.ListObjectVersions(context.TODO(), &s3.ListObjectVersionsInput{
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
		_, err = s3Client.DeleteObjects(context.TODO(), &s3.DeleteObjectsInput{
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

// DownloadAndExtractGzFiles 指定S3パス配下の.gzファイルを一括ダウンロード＆解凍
func DownloadAndExtractGzFiles(s3url, outDir, region, profile string) error {
	ctx := context.Background()
	cfg, err := LoadAwsConfig(region, profile)
	if err != nil {
		return fmt.Errorf("AWS設定のロードに失敗: %w", err)
	}
	bucket, prefix, err := parseS3Url(s3url)
	if err != nil {
		return err
	}
	client := s3.NewFromConfig(cfg)
	// .gzファイル一覧取得
	listInput := &s3.ListObjectsV2Input{
		Bucket: &bucket,
		Prefix: &prefix,
	}
	resp, err := client.ListObjectsV2(ctx, listInput)
	if err != nil {
		return fmt.Errorf("S3リスト取得失敗: %w", err)
	}
	if len(resp.Contents) == 0 {
		return fmt.Errorf("指定パスに.gzファイルが見つかりませんでした")
	}
	for _, obj := range resp.Contents {
		if !strings.HasSuffix(*obj.Key, ".gz") {
			continue
		}
		getObjInput := &s3.GetObjectInput{
			Bucket: &bucket,
			Key:    obj.Key,
		}
		getObjOut, err := client.GetObject(ctx, getObjInput)
		if err != nil {
			return fmt.Errorf("ダウンロード失敗: %w", err)
		}
		defer getObjOut.Body.Close()
		// ローカルパス生成
		relPath := strings.TrimPrefix(*obj.Key, prefix)
		if strings.HasPrefix(relPath, "/") {
			relPath = relPath[1:]
		}
		outPath := filepath.Join(outDir, strings.TrimSuffix(relPath, ".gz"))
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return fmt.Errorf("ディレクトリ作成失敗: %w", err)
		}
		// 解凍して保存
		gzr, err := gzip.NewReader(getObjOut.Body)
		if err != nil {
			return fmt.Errorf("gzip解凍失敗: %w", err)
		}
		f, err := os.Create(outPath)
		if err != nil {
			return fmt.Errorf("ファイル作成失敗: %w", err)
		}
		_, err = io.Copy(f, gzr)
		gzr.Close()
		f.Close()
		if err != nil {
			return fmt.Errorf("ファイル書き込み失敗: %w", err)
		}
		fmt.Printf("✅ %s を %s に保存しました\n", *obj.Key, outPath)
	}
	return nil
}

// parseS3Url s3://bucket/prefix/ 形式を分解
func parseS3Url(s3url string) (bucket, prefix string, err error) {
	if !strings.HasPrefix(s3url, "s3://") {
		return "", "", fmt.Errorf("⚠️ S3パスは s3:// で始めてください")
	}
	noPrefix := strings.TrimPrefix(s3url, "s3://")
	parts := strings.SplitN(noPrefix, "/", 2)
	bucket = parts[0]
	if len(parts) > 1 {
		prefix = parts[1]
	} else {
		prefix = ""
	}
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	return bucket, prefix, nil
}
