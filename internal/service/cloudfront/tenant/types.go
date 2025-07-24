package tenant

// TenantInfo はテナント情報を保持する構造体
type TenantInfo struct {
	Id                       string
	Alias                    string // エイリアスがあれば
	AssociatedDistributionId string // 関連するディストリビューションID
}

// InvalidateOptions はキャッシュ無効化のオプションを保持する構造体
type InvalidateOptions struct {
	DistributionId string   // 必須: ディストリビューションID
	TenantId       string   // 必須: テナントID（InvalidateAllTenantsでは不要）
	Paths          []string // 必須: 無効化するパス（デフォルト: [/*]）
	Wait           bool     // オプション: 無効化完了まで待機
}
