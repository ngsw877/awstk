package schedule

import (
	"awstk/internal/service/common"
	"fmt"
)

// DisplaySchedules はスケジュール一覧を表示する
func DisplaySchedules(schedules []Schedule) {
	// タイトル表示
	fmt.Printf("\n📅 スケジュール一覧\n")

	if len(schedules) == 0 {
		fmt.Println("スケジュールが見つかりませんでした")
		return
	}

	// テーブル列定義（Widthは未使用になるが互換性のため残す）
	columns := []common.TableColumn{
		{Header: "Name", Width: 0},
		{Header: "Schedule", Width: 0},
		{Header: "State", Width: 0},
		{Header: "Target", Width: 0},
	}

	// EventBridge RulesとSchedulerでデータを分離
	var ruleData [][]string
	var schedulerData [][]string

	for _, s := range schedules {
		// Stateに絵文字を付ける
		stateWithEmoji := s.State
		switch s.State {
		case "ENABLED":
			stateWithEmoji = "🟢 " + s.State
		case "DISABLED":
			stateWithEmoji = "🔴 " + s.State
		}

		row := []string{s.Name, s.Expression, stateWithEmoji, s.Target}
		if s.Type == "rule" {
			ruleData = append(ruleData, row)
		} else {
			schedulerData = append(schedulerData, row)
		}
	}

	// EventBridge Rules表示
	if len(ruleData) > 0 {
		common.PrintTable("EventBridge Rules (Schedule)", columns, ruleData)
	}

	// EventBridge Scheduler表示
	if len(schedulerData) > 0 {
		if len(ruleData) > 0 {
			fmt.Println() // 改行
		}
		common.PrintTable("EventBridge Scheduler", columns, schedulerData)
	}

	// 合計表示
	fmt.Printf("\n合計: %d個のスケジュール", len(schedules))
	if len(ruleData) > 0 || len(schedulerData) > 0 {
		fmt.Printf(" (Rules: %d, Scheduler: %d)", len(ruleData), len(schedulerData))
	}
	fmt.Println()
}
