package ssm

// SessionOptions SsmSessionOptions はSSMセッション開始のパラメータを格納する構造体
type SessionOptions struct {
	InstanceId string
}

// PutParamsOptions はパラメータ一括登録のオプション
type PutParamsOptions struct {
	FilePath string // 必須: JSONファイルのパス
	Prefix   string // オプション: パラメータ名のプレフィックス
	DryRun   bool   // オプション: ドライラン実行
}

// parameter はSSMパラメータを表す構造体
type parameter struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	Type        string `json:"type"` // String, SecureString, StringList
	Description string `json:"description,omitempty"`
}

// parametersFile はJSONファイルの構造を表す
type parametersFile struct {
	Parameters []parameter `json:"parameters"`
}

// DeleteParamsOptions はパラメータ一括削除のオプション
type DeleteParamsOptions struct {
	FilePath string // 必須: JSONファイルのパス
	Prefix   string // オプション: パラメータ名のプレフィックス
	DryRun   bool   // オプション: ドライラン実行
	Force    bool   // オプション: 強制削除フラグ
}
