package ecr

import "time"

// RepositoryInfo はリポジトリの詳細情報を保持する構造体
type RepositoryInfo struct {
	RepositoryName string
	RepositoryUri  string
	ImageCount     int
	SizeInBytes    int64
	CreatedAt      *time.Time
	HasLifecycle   bool
}

// ListOptions はリポジトリ一覧表示のオプション
type ListOptions struct {
	EmptyOnly    bool // 空のリポジトリのみを表示
	NoLifecycle  bool // ライフサイクルポリシー未設定のリポジトリのみを表示
	ShowDetails  bool // 詳細情報を表示
}