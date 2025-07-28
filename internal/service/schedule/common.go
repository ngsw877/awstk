package schedule

import (
	"awstk/internal/service/common"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/scheduler"
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

// listEventBridgeRulesWithFilter はフィルターにマッチするEventBridge Rulesを取得する
func listEventBridgeRulesWithFilter(client *eventbridge.Client, filter string) ([]*eventbridge.DescribeRuleOutput, error) {
	ctx := context.Background()
	var matchedRules []*eventbridge.DescribeRuleOutput

	// 全ルールを取得
	listInput := &eventbridge.ListRulesInput{}
	for {
		listOutput, err := client.ListRules(ctx, listInput)
		if err != nil {
			return nil, err
		}

		// フィルターにマッチするルールを抽出
		for _, rule := range listOutput.Rules {
			if rule.Name != nil && common.MatchPattern(*rule.Name, filter) && rule.ScheduleExpression != nil {
				// 詳細情報を取得
				describeOutput, err := client.DescribeRule(ctx, &eventbridge.DescribeRuleInput{
					Name: rule.Name,
				})
				if err == nil {
					matchedRules = append(matchedRules, describeOutput)
				}
			}
		}

		if listOutput.NextToken == nil {
			break
		}
		listInput.NextToken = listOutput.NextToken
	}

	return matchedRules, nil
}

// listEventBridgeSchedulersWithFilter はフィルターにマッチするEventBridge Schedulersを取得する
func listEventBridgeSchedulersWithFilter(client *scheduler.Client, filter string) ([]*scheduler.GetScheduleOutput, error) {
	ctx := context.Background()
	var matchedSchedules []*scheduler.GetScheduleOutput

	// 全スケジュールを取得
	listInput := &scheduler.ListSchedulesInput{}
	for {
		listOutput, err := client.ListSchedules(ctx, listInput)
		if err != nil {
			return nil, err
		}

		// フィルターにマッチするスケジュールを抽出
		for _, schedule := range listOutput.Schedules {
			if schedule.Name != nil && common.MatchPattern(*schedule.Name, filter) {
				// 詳細情報を取得
				getOutput, err := client.GetSchedule(ctx, &scheduler.GetScheduleInput{
					Name: schedule.Name,
				})
				if err == nil {
					matchedSchedules = append(matchedSchedules, getOutput)
				}
			}
		}

		if listOutput.NextToken == nil {
			break
		}
		listInput.NextToken = listOutput.NextToken
	}

	return matchedSchedules, nil
}
