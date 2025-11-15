package elb

// ListOptions はロードバランサー一覧表示時のオプション
type ListOptions struct {
	ProtectedOnly    bool   // 削除保護が有効なもののみ表示
	ShowDetails      bool   // 詳細情報を表示
	LoadBalancerType string // ロードバランサータイプ (alb, nlb, gwlb, 空文字で全て)
}

// LoadBalancerInfo はロードバランサーの情報を保持する構造体
type LoadBalancerInfo struct {
	Name               string
	ARN                string
	DNSName            string
	State              string
	Type               string
	Scheme             string
	VPCId              string
	DeletionProtection bool
	TargetGroupCount   int
	ListenerCount      int
	AvailabilityZones  []string
	CreatedTime        string
}
