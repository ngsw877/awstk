package logs

import (
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

// LogGroupInfo はログループの情報を保持する構造体
type LogGroupInfo struct {
	LogGroupName    string
	StoredBytes     int64
	CreationTime    int64
	RetentionInDays *int32
	LogStreamCount  int32
}

// LogGroupDetail はロググループの詳細情報
type LogGroupDetail struct {
	types.LogGroup
	StreamCount int32
	IsEmpty     bool
}

// DeleteOptions はログ削除時のオプション
type DeleteOptions struct {
	Filter      string   // フィルターパターン
	LogGroups   []string // 削除対象のロググループ名
	EmptyOnly   bool     // 空のロググループのみ削除
	NoRetention bool     // 保存期間未設定のロググループのみ削除
	Exact       bool     // 大文字小文字を区別してマッチ
}
