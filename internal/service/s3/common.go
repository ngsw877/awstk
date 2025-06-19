package s3

import (
	"fmt"
	"strings"
)

// parseS3Url はユーザー入力の S3 パスをバケット名とプレフィックスに分解します。
// 以下 2 つの形式をサポートします。
//  1. s3://bucket/prefix/
//  2. bucket/prefix/
//
// prefix 省略や末尾スラッシュなしも許容します。
func parseS3Url(s3path string) (bucket, prefix string, err error) {
	// s3:// プレフィックスが付いている場合は除去
	if strings.HasPrefix(s3path, "s3://") {
		s3path = strings.TrimPrefix(s3path, "s3://")
	}

	// バケット名とプレフィックスを分割
	parts := strings.SplitN(s3path, "/", 2)
	bucket = parts[0]
	if bucket == "" {
		return "", "", fmt.Errorf("バケット名が空です")
	}

	if len(parts) > 1 {
		prefix = parts[1]
	}

	// プレフィックスが存在し、末尾にスラッシュが無ければ付加
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	return bucket, prefix, nil
}
