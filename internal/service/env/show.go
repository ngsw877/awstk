package env

import (
	"fmt"
	"os"
)

// ShowAllVariables はすべてのサポートされている環境変数を表示
func ShowAllVariables() {
	fmt.Println("📋 AWS関連の環境変数の状態:")
	fmt.Println()

	// 順序を固定するため、明示的に順番を定義
	order := []string{"profile", "stack"}

	for _, key := range order {
		v := SupportedVariables[key]
		value := os.Getenv(v.Name)

		if value != "" {
			fmt.Printf("  %s (%s): %s\n", v.Description, v.Name, value)
		} else {
			fmt.Printf("  %s (%s): 未設定\n", v.Description, v.Name)
		}
	}
}