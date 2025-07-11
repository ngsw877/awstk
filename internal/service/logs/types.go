package logs

import (
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

// LogGroupInfo はログループの情報を保持する構造体
type LogGroupInfo struct {
	LogGroupName     string
	StoredBytes      int64
	CreationTime     int64
	RetentionInDays  *int32
	LogStreamCount   int32
}

// LogGroupDetail はロググループの詳細情報
type LogGroupDetail struct {
	types.LogGroup
	StreamCount int32
	IsEmpty     bool
}