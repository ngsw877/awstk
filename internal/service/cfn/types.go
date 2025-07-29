package cfn

// StackResources はCloudFormationスタック内のリソース識別子を格納する構造体
type StackResources struct {
	Ec2InstanceIds   []string
	RdsInstanceIds   []string
	AuroraClusterIds []string
	EcsServiceInfo   []EcsServiceInfo
}

// Stack CfnStack はCloudFormationスタックの名前とステータスを表す構造体
type Stack struct {
	Name   string
	Status string
}

// CleanupOptions はクリーンアップコマンドのオプション
type CleanupOptions struct {
	Filter string // スタック名のフィルター（部分一致）
	Status string // 削除対象のステータス（カンマ区切り）
	Force  bool   // 確認プロンプトをスキップ
}

// ProtectOptions は削除保護コマンドのオプション
type ProtectOptions struct {
	Filter string // スタック名のフィルター（部分一致）
	Status string // 対象のステータス（カンマ区切り）
	Enable bool   // 削除保護を有効化するかどうか
	Force  bool   // 確認プロンプトをスキップ
}

// DriftOptions はドリフト検出コマンドのオプション
type DriftOptions struct {
	Stacks []string // スタック名のリスト
	Filter string   // スタック名のフィルター（部分一致）
	All    bool     // すべてのスタックを対象
}

// DriftStatusOptions はドリフト状態確認コマンドのオプション
type DriftStatusOptions struct {
	Stacks      []string // スタック名のリスト
	Filter      string   // スタック名のフィルター（部分一致）
	All         bool     // すべてのスタックを対象
	DriftedOnly bool     // ドリフトしているスタックのみ表示
}
