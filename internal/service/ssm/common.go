package ssm

import "strings"

// normalizeParameterName はプレフィックスとパラメータ名を正しく結合する
func normalizeParameterName(prefix, name string) string {
	if prefix == "" {
		return name
	}
	
	// プレフィックス末尾の/を除去
	prefix = strings.TrimSuffix(prefix, "/")
	
	// パラメータ名先頭の/を確保
	if !strings.HasPrefix(name, "/") {
		name = "/" + name
	}
	
	return prefix + name
}