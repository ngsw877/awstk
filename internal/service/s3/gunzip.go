package s3

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// DownloadAndExtractGzFiles 指定S3パス配下の.gzファイルを一括ダウンロード＆解凍
func DownloadAndExtractGzFiles(s3Client *s3.Client, s3url, outDir string) error {
	ctx := context.Background()
	bucket, prefix, err := parseS3Url(s3url)
	if err != nil {
		return err
	}

	// .gzファイル一覧取得
	listInput := &s3.ListObjectsV2Input{
		Bucket: &bucket,
		Prefix: &prefix,
	}
	resp, err := s3Client.ListObjectsV2(ctx, listInput)
	if err != nil {
		return fmt.Errorf("s3リスト取得失敗: %w", err)
	}
	if len(resp.Contents) == 0 {
		return fmt.Errorf("指定されたパス配下に .gz ファイルが見つかりませんでした")
	}

	// 出力ディレクトリを作成
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("出力ディレクトリの作成に失敗: %w", err)
	}

	gzCount := 0
	for _, obj := range resp.Contents {
		key := *obj.Key
		if !strings.HasSuffix(key, ".gz") {
			continue // .gz以外はスキップ
		}
		gzCount++

		fmt.Printf("📦 %s をダウンロード中...\n", key)
		// ダウンロード
		getObjectInput := &s3.GetObjectInput{
			Bucket: &bucket,
			Key:    &key,
		}
		getResp, err := s3Client.GetObject(ctx, getObjectInput)
		if err != nil {
			fmt.Printf("❌ %s のダウンロードに失敗: %v\n", key, err)
			continue
		}

		// 解凍とローカル保存
		baseKey := strings.TrimSuffix(filepath.Base(key), ".gz")
		outPath := filepath.Join(outDir, baseKey)

		// gzip解凍
		gzr, err := gzip.NewReader(getResp.Body)
		if err != nil {
			fmt.Printf("❌ %s のgzip解凍に失敗: %v\n", key, err)
			if closeErr := getResp.Body.Close(); closeErr != nil {
				fmt.Printf("⚠️  S3レスポンスボディのクローズに失敗: %v\n", closeErr)
			}
			continue
		}

		// ファイル作成
		outFile, err := os.Create(outPath)
		if err != nil {
			fmt.Printf("❌ %s のファイル作成に失敗: %v\n", outPath, err)
			if closeErr := gzr.Close(); closeErr != nil {
				fmt.Printf("⚠️  %s のgzipリーダーのクローズに失敗: %v\n", key, closeErr)
			}
			if closeErr := getResp.Body.Close(); closeErr != nil {
				fmt.Printf("⚠️  S3レスポンスボディのクローズに失敗: %v\n", closeErr)
			}
			continue
		}

		// 解凍データをファイルに書き込み
		_, err = io.Copy(outFile, gzr)
		if closeErr := gzr.Close(); closeErr != nil {
			fmt.Printf("⚠️  %s のgzipリーダーのクローズに失敗: %v\n", key, closeErr)
		}
		if closeErr := outFile.Close(); closeErr != nil {
			fmt.Printf("⚠️  %s のファイルクローズに失敗: %v\n", outPath, closeErr)
		}
		if err != nil {
			fmt.Printf("❌ %s の書き込みに失敗: %v\n", outPath, err)
			if closeErr := getResp.Body.Close(); closeErr != nil {
				fmt.Printf("⚠️  S3レスポンスボディのクローズに失敗: %v\n", closeErr)
			}
			continue
		}
		if closeErr := getResp.Body.Close(); closeErr != nil {
			fmt.Printf("⚠️  S3レスポンスボディのクローズに失敗: %v\n", closeErr)
		}
		fmt.Printf("✅ %s → %s\n", key, outPath)
	}

	if gzCount == 0 {
		return fmt.Errorf("指定されたパス配下に .gz ファイルが見つかりませんでした")
	}
	fmt.Printf("🎉 %d個の .gz ファイルの処理が完了しました\n", gzCount)
	return nil
}
