package common

import (
	"fmt"
	"strings"
)

// ListOutput はリスト表示の共通構造体
type ListOutput struct {
	Title        string   // 例: "S3バケット一覧"
	Items        []string // 表示するアイテムのリスト
	ResourceName string   // 例: "バケット", "リポジトリ"
	ShowCount    bool     // 合計数を表示するか
}

// PrintSimpleList はシンプルな箇条書きリストを表示
func PrintSimpleList(output ListOutput) {
	// タイトル表示
	fmt.Printf("%s:\n", output.Title)

	// アイテムがない場合
	if len(output.Items) == 0 {
		fmt.Printf("該当する%sはありませんでした\n", output.ResourceName)
		return
	}

	// 各アイテムを表示
	for _, item := range output.Items {
		fmt.Printf("  - %s\n", item)
	}

	// 合計数表示
	if output.ShowCount {
		fmt.Printf("\n合計: %d個の%s\n", len(output.Items), output.ResourceName)
	}
}

// PrintNumberedList は番号付きリストを表示
func PrintNumberedList(output ListOutput) {
	// タイトル表示（件数付き）
	fmt.Printf("%s: (全%d件)\n", output.Title, len(output.Items))

	// アイテムがない場合
	if len(output.Items) == 0 {
		fmt.Printf("%sが見つかりませんでした\n", output.ResourceName)
		return
	}

	// 各アイテムを番号付きで表示
	for i, item := range output.Items {
		fmt.Printf("  %3d. %s\n", i+1, item)
	}
}

// ListItem は詳細情報を持つリストアイテム
type ListItem struct {
	Name   string
	Status string // オプション: ステータス情報
}

// PrintStatusList はステータス付きリストを表示
func PrintStatusList(title string, items []ListItem, resourceName string) {
	fmt.Printf("%s: (全%d件)\n", title, len(items))

	if len(items) == 0 {
		fmt.Printf("%sが見つかりませんでした\n", resourceName)
		return
	}

	for i, item := range items {
		if item.Status != "" {
			fmt.Printf("  %3d. %s [%s]\n", i+1, item.Name, item.Status)
		} else {
			fmt.Printf("  %3d. %s\n", i+1, item.Name)
		}
	}
}

// GenerateFilteredTitle はフィルタ条件に基づいてタイトルを生成
func GenerateFilteredTitle(resourceType string, conditions ...string) string {
	if len(conditions) == 0 {
		return fmt.Sprintf("%s一覧", resourceType)
	}

	// 空文字列を除外
	var validConditions []string
	for _, cond := range conditions {
		if cond != "" {
			validConditions = append(validConditions, cond)
		}
	}

	if len(validConditions) == 0 {
		return fmt.Sprintf("%s一覧", resourceType)
	}

	return fmt.Sprintf("%s%s一覧", strings.Join(validConditions, ""), resourceType)
}

// FormatListError はリスト取得エラーを統一フォーマットで返す
func FormatListError(service string, err error) error {
	return fmt.Errorf("❌ %s一覧取得でエラー: %w", service, err)
}

// FormatEmptyMessage は該当リソースがない場合のメッセージを返す
func FormatEmptyMessage(resourceType string) string {
	return fmt.Sprintf("%sが見つかりませんでした", resourceType)
}
