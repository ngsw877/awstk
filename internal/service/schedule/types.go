package schedule

// Schedule はスケジュール情報を表す構造体
type Schedule struct {
	Name       string // スケジュール名
	Type       string // "rule" or "scheduler"
	Expression string // cron式やrate式
	State      string // "ENABLED" or "DISABLED"
	Target     string // ターゲットの簡潔な表現
	Arn        string // リソースARN
}

// ListOptions はスケジュール一覧取得のオプション
type ListOptions struct {
	Type string // "all", "rule", "scheduler"
}
