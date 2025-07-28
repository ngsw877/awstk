package common

import (
	"path/filepath"
	"strings"
)

// MatchPattern はワイルドカードパターンマッチングを行う
// ワイルドカード（*）を含む場合はglob形式でマッチング、
// 含まない場合は部分一致で判定する
func MatchPattern(name, pattern string) bool {
	// ワイルドカードを含む場合
	if strings.Contains(pattern, "*") {
		// glob パターンマッチング
		matched, _ := filepath.Match(pattern, name)
		return matched
	}
	// ワイルドカードなしの場合は部分一致
	return strings.Contains(name, pattern)
}
