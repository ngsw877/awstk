package role

import "time"

// ListOptions IamRoleListOptions IAMロール一覧取得時のオプション
type ListOptions struct {
	UnusedDays int
	Exclude    []string
}

// RoleItem IamRole IAMロール一覧表示用の情報
type RoleItem struct {
	Name            string
	Arn             string
	LastUsed        *time.Time
	IsServiceLinked bool
}

// UnusedRole IamRoleUnused 未使用IAMロールの情報
type UnusedRole struct {
	Name     string
	Arn      string
	LastUsed *time.Time
}

// DeleteOptions IAMロール削除時のオプション
type DeleteOptions struct {
	Filter     string   // 必須: 削除対象のフィルターパターン
	UnusedDays int      // 0=無効、-1=never used、>0=指定日数以上未使用
	Exclude    []string // 除外パターン
}
