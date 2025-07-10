package s3

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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

// FilterEmptyBuckets は指定されたバケットの中から空のバケットのみを返す関数
func FilterEmptyBuckets(s3Client *s3.Client, buckets []string) ([]string, error) {
	var emptyBuckets []string

	for _, bucket := range buckets {
		// バケットが空かどうかをチェック
		isEmpty, err := isBucketEmpty(s3Client, bucket)
		if err != nil {
			// エラーが発生してもスキップして続行
			continue
		}
		if isEmpty {
			emptyBuckets = append(emptyBuckets, bucket)
		}
	}

	return emptyBuckets, nil
}

// isBucketEmpty はバケットが空かどうかをチェックする関数
func isBucketEmpty(s3Client *s3.Client, bucketName string) (bool, error) {
	// MaxKeys=1で最初のオブジェクトのみ取得を試みる
	result, err := s3Client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
		Bucket:  aws.String(bucketName),
		MaxKeys: aws.Int32(1),
	})
	if err != nil {
		return false, err
	}

	// オブジェクトが0個なら空
	return len(result.Contents) == 0, nil
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
		page, err := paginator.NextPage(context.Background())
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