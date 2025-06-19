package s3

import (
	"fmt"
	"strings"
)

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
