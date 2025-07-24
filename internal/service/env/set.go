package env

import "fmt"

// GetExportCommand は環境変数をエクスポートするコマンドを返す
func GetExportCommand(variable, value string) (string, error) {
	if err := ValidateVariable(variable); err != nil {
		return "", err
	}

	v := SupportedVariables[variable]
	return fmt.Sprintf("export %s=%s", v.Name, value), nil
}
