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
