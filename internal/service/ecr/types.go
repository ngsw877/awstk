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