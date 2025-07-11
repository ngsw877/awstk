package common

import (
	"fmt"
	"time"
)

// FormatBytes はバイト数を人間が読みやすい形式に変換する関数
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatTimestamp はUnixミリ秒のタイムスタンプをフォーマットする関数
func FormatTimestamp(timestamp *int64) string {
	if timestamp == nil {
		return "不明"
	}
	t := time.Unix(*timestamp/1000, (*timestamp%1000)*1000000)
	return t.Format("2006-01-02 15:04:05")
}