package env

import "fmt"

// GetUnsetCommand は環境変数を削除するコマンドを返す
func GetUnsetCommand(variable string) (string, error) {
	if err := ValidateVariable(variable); err != nil {
		return "", err
	}

	v := SupportedVariables[variable]
	return fmt.Sprintf("unset %s", v.Name), nil
}
