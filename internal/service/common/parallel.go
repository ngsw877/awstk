package common

import (
	"sync"
)

// ParallelExecutor は並列処理を管理する構造体
type ParallelExecutor struct {
	maxWorkers int
	wg         sync.WaitGroup
	semaphore  chan struct{}
}

// NewParallelExecutor は新しいParallelExecutorを作成
func NewParallelExecutor(maxWorkers int) *ParallelExecutor {
	return &ParallelExecutor{
		maxWorkers: maxWorkers,
		semaphore:  make(chan struct{}, maxWorkers),
	}
}

// Execute はタスクを並列で実行
func (p *ParallelExecutor) Execute(task func()) {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		p.semaphore <- struct{}{}        // セマフォ取得（同時実行数制限）
		defer func() { <-p.semaphore }() // セマフォ解放
		task()
	}()
}

// Wait はすべてのタスクの完了を待つ
func (p *ParallelExecutor) Wait() {
	p.wg.Wait()
}

// ProcessResult は処理結果を保持する構造体
type ProcessResult struct {
	Item    string
	Success bool
	Error   error
}

// CollectResults は並列処理の結果を収集するヘルパー関数
func CollectResults(results []ProcessResult) (successCount, failCount int) {
	for _, result := range results {
		if result.Success {
			successCount++
		} else {
			failCount++
		}
	}
	return
}

// CleanupResult はクリーンアップ処理の結果を保持する構造体
type CleanupResult struct {
	ResourceType string   // リソースタイプ（例: "S3バケット", "ECRリポジトリ"）
	Deleted      []string // 削除成功したリソース名
	Failed       []string // 削除失敗したリソース名
}

// TotalCount は対象リソースの総数を返します
func (r CleanupResult) TotalCount() int {
	return len(r.Deleted) + len(r.Failed)
}

// CollectCleanupResult はProcessResultからCleanupResultを生成します
func CollectCleanupResult(resourceType string, results []ProcessResult) CleanupResult {
	result := CleanupResult{
		ResourceType: resourceType,
		Deleted:      []string{},
		Failed:       []string{},
	}
	for _, r := range results {
		if r.Success {
			result.Deleted = append(result.Deleted, r.Item)
		} else {
			result.Failed = append(result.Failed, r.Item)
		}
	}
	return result
}
