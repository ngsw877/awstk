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
func ListS3Buckets(s3Client *s3.Client) ([]string, error) {
	result, err := s3Client.ListBuckets(context.Background(), &s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}

	buckets := make([]string, 0, len(result.Buckets))
	for _, bucket := range result.Buckets {
		buckets = append(buckets, *bucket.Name)
	}
	return buckets, nil
}

// getS3BucketsByKeyword はキーワードに一致するS3バケット名の一覧を取得します
func getS3BucketsByKeyword(s3Client *s3.Client, searchString string) ([]string, error) {
	// バケット一覧を取得
	listBucketsOutput, err := s3Client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
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

// cleanupS3Buckets は指定したS3バケット一覧を削除します
func cleanupS3Buckets(s3Client *s3.Client, bucketNames []string) error {
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
		return fmt.Errorf("S3リスト取得失敗: %w", err)
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
			getResp.Body.Close()
			continue
		}

		// ファイル作成
		outFile, err := os.Create(outPath)
		if err != nil {
			fmt.Printf("❌ %s のファイル作成に失敗: %v\n", outPath, err)
			gzr.Close()
			getResp.Body.Close()
			continue
		}

		// 解凍データをファイルに書き込み
		_, err = io.Copy(outFile, gzr)
		gzr.Close()
		outFile.Close()
		if err != nil {
			fmt.Printf("❌ %s の書き込みに失敗: %v\n", outPath, err)
			getResp.Body.Close()
			continue
		}
		getResp.Body.Close()
		fmt.Printf("✅ %s → %s\n", key, outPath)
	}

	if gzCount == 0 {
		return fmt.Errorf("指定されたパス配下に .gz ファイルが見つかりませんでした")
	}
	fmt.Printf("🎉 %d個の .gz ファイルの処理が完了しました\n", gzCount)
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

// S3Object はS3オブジェクトの情報を格納する構造体
type S3Object struct {
	Key          string
	Size         int64
	LastModified time.Time
}

// listS3Objects はS3バケット内のオブジェクト一覧を取得します
func listS3Objects(s3Client *s3.Client, bucketName string, prefix string) ([]S3Object, error) {
	var objects []S3Object

	// ListObjectsV2を使用してオブジェクト一覧を取得
	paginator := s3.NewListObjectsV2Paginator(s3Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("S3オブジェクト一覧取得エラー: %w", err)
		}

		for _, obj := range page.Contents {
			objects = append(objects, S3Object{
				Key:          *obj.Key,
				Size:         *obj.Size,
				LastModified: *obj.LastModified,
			})
		}
	}

	return objects, nil
}

// TreeNode はツリー構造のノードを表現する構造体
type TreeNode struct {
	Name     string
	IsDir    bool
	Children map[string]*TreeNode
	Object   *S3Object // ファイルの場合のみ設定
}

// buildTreeFromObjects S3オブジェクトリストからツリー構造を構築します
func buildTreeFromObjects(objects []S3Object, prefix string) *TreeNode {
	root := &TreeNode{
		Name:     "",
		IsDir:    true,
		Children: make(map[string]*TreeNode),
	}

	for _, obj := range objects {
		// プレフィックスを除去した相対パスを取得
		relativePath := strings.TrimPrefix(obj.Key, prefix)
		if strings.HasPrefix(relativePath, "/") {
			relativePath = relativePath[1:]
		}

		// 空のパスはスキップ
		if relativePath == "" {
			continue
		}

		// パスを分割してツリーに追加
		parts := strings.Split(relativePath, "/")
		current := root

		// ディレクトリ部分を処理
		for _, part := range parts[:len(parts)-1] {
			if part == "" {
				continue
			}

			if current.Children[part] == nil {
				current.Children[part] = &TreeNode{
					Name:     part,
					IsDir:    true,
					Children: make(map[string]*TreeNode),
				}
			}
			current = current.Children[part]
		}

		// ファイル部分を処理
		fileName := parts[len(parts)-1]
		if fileName != "" {
			current.Children[fileName] = &TreeNode{
				Name:   fileName,
				IsDir:  false,
				Object: &obj,
			}
		}
	}

	return root
}

// displayTree ツリー構造を表示します
func displayTree(node *TreeNode, prefix string, isLast bool, humanReadable bool, showTime bool) {
	if node.Name != "" {
		connector := "├── "
		if isLast {
			connector = "└── "
		}

		if node.IsDir {
			fmt.Printf("%s%s%s/\n", prefix, connector, node.Name)
		} else {
			if humanReadable && node.Object != nil {
				// ファイルサイズを人間が読める形式で表示
				sizeStr := formatFileSize(node.Object.Size)
				if showTime {
					// 更新日時も表示（括弧を分ける）
					timeStr := node.Object.LastModified.Format("2006-01-02 15:04:05")
					fmt.Printf("%s%s%s (%s) [%s]\n", prefix, connector, node.Name, sizeStr, timeStr)
				} else {
					fmt.Printf("%s%s%s (%s)\n", prefix, connector, node.Name, sizeStr)
				}
			} else {
				fmt.Printf("%s%s%s\n", prefix, connector, node.Name)
			}
		}
	}

	// 子ノードをソートして表示
	var names []string
	for name := range node.Children {
		names = append(names, name)
	}

	// ディレクトリを先に、ファイルを後に表示するためのソート
	dirs := []string{}
	files := []string{}
	for _, name := range names {
		if node.Children[name].IsDir {
			dirs = append(dirs, name)
		} else {
			files = append(files, name)
		}
	}

	// ディレクトリとファイルそれぞれをアルファベット順にソート
	for i := 0; i < len(dirs); i++ {
		for j := i + 1; j < len(dirs); j++ {
			if dirs[i] > dirs[j] {
				dirs[i], dirs[j] = dirs[j], dirs[i]
			}
		}
	}
	for i := 0; i < len(files); i++ {
		for j := i + 1; j < len(files); j++ {
			if files[i] > files[j] {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	// 統合したリスト
	allNames := append(dirs, files...)

	for i, name := range allNames {
		child := node.Children[name]
		isLastChild := (i == len(allNames)-1)

		var newPrefix string
		if node.Name == "" {
			// ルートノードの場合
			newPrefix = prefix
		} else {
			if isLast {
				newPrefix = prefix + "    "
			} else {
				newPrefix = prefix + "│   "
			}
		}

		displayTree(child, newPrefix, isLastChild, humanReadable, showTime)
	}
}

// formatFileSize ファイルサイズを人間が読める形式でフォーマットします
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(size)/float64(div), "KMGTPE"[exp])
}

// ListS3TreeView 指定されたS3パスをツリー形式で表示します
func ListS3TreeView(s3Client *s3.Client, s3Path string, showTime bool) error {
	bucket, prefix, err := parseS3Url(s3Path)
	if err != nil {
		return err
	}

	// S3オブジェクト一覧を取得
	objects, err := listS3Objects(s3Client, bucket, prefix)
	if err != nil {
		return err
	}

	if len(objects) == 0 {
		fmt.Printf("🔍 %s には何も見つかりませんでした\n", s3Path)
		return nil
	}

	// ツリー構造を構築
	tree := buildTreeFromObjects(objects, prefix)

	// ツリーを表示
	fmt.Printf("📁 %s\n", s3Path)
	displayTree(tree, "", true, true, showTime)

	return nil
}
