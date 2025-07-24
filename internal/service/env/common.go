package env

import "fmt"

// ValidateVariable は変数名が有効かチェック
func ValidateVariable(variable string) error {
	if _, ok := SupportedVariables[variable]; !ok {
		return fmt.Errorf("❌ エラー: '%s' はサポートされていない変数です。stack または profile を指定してください", variable)
	}
	return nil
}
