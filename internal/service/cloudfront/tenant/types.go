package tenant

// TenantInfo はテナント情報を保持する構造体
type TenantInfo struct {
	Id          string
	Alias       string // エイリアスがあれば
	AssociatedDistributionId string // 関連するディストリビューションID
}