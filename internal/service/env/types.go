package env

// Variable は環境変数の情報を表す構造体
type Variable struct {
	Name        string // 環境変数名 (e.g., AWS_STACK_NAME)
	ShortName   string // 短縮名 (e.g., stack)
	Description string // 説明
}

// SupportedVariables はサポートされている環境変数のマップ
var SupportedVariables = map[string]Variable{
	"stack": {
		Name:        "AWS_STACK_NAME",
		ShortName:   "stack",
		Description: "スタック名",
	},
	"profile": {
		Name:        "AWS_PROFILE",
		ShortName:   "profile",
		Description: "プロファイル",
	},
}
