package s3

import "time"

// S3Object はS3オブジェクトの情報を格納する構造体
type S3Object struct {
	Key          string
	Size         int64
	LastModified time.Time
}

// TreeNode はツリー構造のノードを表現する構造体
type TreeNode struct {
	Name     string
	IsDir    bool
	Children map[string]*TreeNode
	Object   *S3Object // ファイルの場合のみ設定
}

// S3BucketAvailabilityResult S3バケット利用可否判定結果構造体
type S3BucketAvailabilityResult struct {
	BucketName string
	StatusCode int
	Message    string
}
