package common

// ListOutput はリスト表示の共通構造体
type ListOutput struct {
	Title        string   // 例: "S3バケット一覧"
	Items        []string // 表示するアイテムのリスト
	ResourceName string   // 例: "バケット", "リポジトリ"
	ShowCount    bool     // 合計数を表示するか
}

// ListItem は詳細情報を持つリストアイテム
type ListItem struct {
	Name   string
	Status string // オプション: ステータス情報
}