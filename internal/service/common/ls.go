package common

import (
	"fmt"
	"strings"
	"time"
)

// ===== フォーマット関数 =====

// FormatBytes はバイト数を人間が読みやすい形式に変換する関数
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatTimestamp はUnixミリ秒のタイムスタンプをフォーマットする関数
func FormatTimestamp(timestamp *int64) string {
	if timestamp == nil {
		return "不明"
	}
	t := time.Unix(*timestamp/1000, (*timestamp%1000)*1000000)
	return t.Format("2006-01-02 15:04:05")
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

// ===== 低レベル表示関数 =====

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

// PrintTable はテーブル形式でデータを表示する
func PrintTable(title string, columns []TableColumn, data [][]string) {
	if title != "" {
		fmt.Printf("\n%s:\n", title)
	}
	
	// 各列の最大幅を計算（ヘッダーとデータの中で最大値を取得）
	colWidths := make([]int, len(columns))
	
	// ヘッダーの幅で初期化
	for i, col := range columns {
		colWidths[i] = len(col.Header)
	}
	
	// 各データセルと比較して最大値を更新
	for _, row := range data {
		for i, cell := range row {
			if i < len(colWidths) {
				if len(cell) > colWidths[i] {
					colWidths[i] = len(cell)
				}
			}
		}
	}
	
	// ヘッダー表示
	for i, col := range columns {
		fmt.Printf("%-*s ", colWidths[i], col.Header)
	}
	fmt.Println()
	
	// 区切り線
	for i := range columns {
		fmt.Printf("%s ", strings.Repeat("-", colWidths[i]))
	}
	fmt.Println()
	
	// データ行
	for _, row := range data {
		for i, cell := range row {
			if i < len(columns) {
				fmt.Printf("%-*s ", colWidths[i], cell)
			}
		}
		fmt.Println()
	}
}

// ===== 高レベル表示関数 =====

// DisplayList は汎用的なリスト表示関数
func DisplayList[T any](
	items []T,
	title string,
	toTableData func([]T) ([]TableColumn, [][]string),
	opts *DisplayOptions,
) error {
	// デフォルトオプション
	if opts == nil {
		opts = &DisplayOptions{}
	}
	if opts.EmptyMessage == "" {
		opts.EmptyMessage = "リソースが見つかりませんでした"
	}

	// フィルタ条件がある場合はタイトルに追加
	if len(opts.FilterMessages) > 0 {
		title = GenerateFilteredTitle(title, opts.FilterMessages...)
	}

	// 空の場合の処理
	if len(items) == 0 {
		fmt.Println(opts.EmptyMessage)
		return nil
	}

	// テーブル表示
	columns, data := toTableData(items)
	PrintTable(title, columns, data)

	// 件数表示
	if opts.ShowCount {
		fmt.Printf("\n合計: %d件\n", len(items))
	}

	return nil
}

// DisplaySimpleList はシンプルなリスト表示（1列のみ）
func DisplaySimpleList[T any](
	items []T,
	title string,
	getName func(T) string,
	opts *DisplayOptions,
) error {
	toTableData := func(items []T) ([]TableColumn, [][]string) {
		columns := []TableColumn{{Header: "名前"}}
		data := make([][]string, len(items))
		for i, item := range items {
			data[i] = []string{getName(item)}
		}
		return columns, data
	}
	
	return DisplayList(items, title, toTableData, opts)
}