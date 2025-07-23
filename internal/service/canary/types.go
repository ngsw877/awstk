package canary

import (
	"time"
)

// Canary はAWS Synthetics Canaryの情報を表す構造体
type Canary struct {
	Name           string
	State          string    // RUNNING, STOPPED, ERROR等
	StateReason    string    // 状態の理由
	Schedule       string    // cron式またはrate式
	SuccessRate    float64   // 成功率（パーセント）
	LastRunTime    time.Time // 最終実行時刻
	LastRunStatus  string    // PASSED, FAILED等
	RuntimeVersion string    // syn-nodejs-puppeteer-x.x等
}

// CanaryState はCanaryの実行状態を表す定数
const (
	CanaryStateRunning     = "RUNNING"
	CanaryStateStopped     = "STOPPED"
	CanaryStateError       = "ERROR"
	CanaryStateReady       = "READY"
	CanaryStateStopping    = "STOPPING"
	CanaryStateStarting    = "STARTING"
	CanaryStateDeleting    = "DELETING"
	CanaryStateUpdating    = "UPDATING"
	CanaryStateRollbackFailed = "ROLLBACK_FAILED"
)

// CanaryRunState はCanary実行結果の状態を表す定数
const (
	CanaryRunStatePassed = "PASSED"
	CanaryRunStateFailed = "FAILED"
	CanaryRunStateRunning = "RUNNING"
)