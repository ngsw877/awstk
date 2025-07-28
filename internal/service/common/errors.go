package common

// エラーメッセージの絵文字定数
const (
	ErrorIcon   = "❌"
	SuccessIcon = "✅"
	WarningIcon = "⚠️"
	SearchIcon  = "🔍"
	InfoIcon    = "📋"
	ProcessIcon = "🔄"
	PartyIcon   = "🎉"
)

// エラーメッセージフォーマット定数
const (
	// 一覧取得エラー
	ListErrorFormat = "%s %s一覧の取得に失敗: %w"

	// リソース操作エラー
	EnableErrorFormat  = "%s %s の有効化に失敗: %w"
	DisableErrorFormat = "%s %s の無効化に失敗: %w"
	StartErrorFormat   = "%s %s の起動に失敗: %w"
	StopErrorFormat    = "%s %s の停止に失敗: %w"
	DeleteErrorFormat  = "%s %s の削除に失敗: %w"

	// その他の操作エラー
	CreateErrorFormat = "%s %s の作成に失敗: %w"
	UpdateErrorFormat = "%s %s の更新に失敗: %w"
	GetErrorFormat    = "%s %s の取得に失敗: %w"

	// 成功メッセージ
	EnableSuccessFormat  = "%s %s を有効化しました"
	DisableSuccessFormat = "%s %s を無効化しました"
	StartSuccessFormat   = "%s %s を起動しました"
	StopSuccessFormat    = "%s %s を停止しました"
	DeleteSuccessFormat  = "%s %s を削除しました"
	CreateSuccessFormat  = "%s %s を作成しました"
	UpdateSuccessFormat  = "%s %s を更新しました"

	// 処理中メッセージ
	ProcessingFormat = "%s %s を処理中..."
	SearchingFormat  = "%s %s を検索中..."
)
