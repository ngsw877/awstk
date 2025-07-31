package common

import (
	"strings"

	"github.com/gobwas/glob"
)

// MatchesFilter は文字列がフィルターパターンにマッチするか判定します
// ワイルドカード文字(*?[])が含まれていればglobパターンマッチング、なければ部分一致
func MatchesFilter(text, filter string) bool {
	if strings.ContainsAny(filter, "*?[]") {
		// ワイルドカードパターンマッチング
		pattern := glob.MustCompile(filter)
		return pattern.Match(text)
	}
	// 部分一致
	return strings.Contains(text, filter)
}

// RemoveDuplicates は文字列スライスから重複を除去します
func RemoveDuplicates(items []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}
