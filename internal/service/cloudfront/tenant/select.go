package tenant

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
)

// SelectTenant は複数のテナントから一つを選択します
func SelectTenant(client *cloudfront.Client, distributionId string) (string, error) {
	// テナント一覧を取得
	tenants, err := ListTenants(client, distributionId)
	if err != nil {
		return "", fmt.Errorf("テナント一覧の取得に失敗: %w", err)
	}

	if len(tenants) == 0 {
		return "", fmt.Errorf("テナントが見つかりませんでした")
	}

	fmt.Println("\nテナントを選択してください:")

	// 各テナントを表示
	for i, tenant := range tenants {
		displayName := tenant.Id
		if tenant.Alias != "" {
			displayName = fmt.Sprintf("%s (%s)", tenant.Id, tenant.Alias)
		}
		fmt.Printf("  %d. %s\n", i+1, displayName)
	}

	// ユーザーの選択を待つ
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\n番号を入力してください (1-" + fmt.Sprintf("%d", len(tenants)) + "): ")
	
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("入力エラー: %w", err)
	}

	// 選択番号を解析
	choice, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil || choice < 1 || choice > len(tenants) {
		return "", fmt.Errorf("無効な選択です")
	}

	selectedTenant := tenants[choice-1]
	fmt.Printf("\n✅ テナント '%s' を選択しました\n", selectedTenant.Id)
	
	return selectedTenant.Id, nil
}