package schedule

import (
	"awstk/internal/service/common"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/scheduler"
)

// EnableSchedule は単一のスケジュールを有効化する
func EnableSchedule(eventBridgeClient *eventbridge.Client, schedulerClient *scheduler.Client, name string) error {
	// スケジュールタイプの判別
	scheduleType, err := detectScheduleType(eventBridgeClient, schedulerClient, name)
	if err != nil {
		return err
	}

	// タイプに応じて処理を分岐
	if scheduleType == "rule" {
		return enableEventBridgeRule(eventBridgeClient, name)
	}
	return enableEventBridgeScheduler(schedulerClient, name)
}

// EnableSchedulesWithFilter はフィルターにマッチする全スケジュールを有効化する
func EnableSchedulesWithFilter(eventBridgeClient *eventbridge.Client, schedulerClient *scheduler.Client, filter string) error {
	enabledCount := 0

	fmt.Printf("フィルター '%s' にマッチするスケジュールを検索中...\n", filter)

	// EventBridge Rulesの有効化
	rules, err := listEventBridgeRulesWithFilter(eventBridgeClient, filter)
	if err != nil {
		return fmt.Errorf("EventBridge Rulesの取得に失敗: %w", err)
	}

	for _, rule := range rules {
		if rule.State == "DISABLED" {
			if err := enableEventBridgeRule(eventBridgeClient, *rule.Name); err != nil {
				fmt.Printf("  %s %s (Rule) の有効化に失敗: %v\n", common.WarningIcon, *rule.Name, err)
			} else {
				enabledCount++
			}
		}
	}

	// EventBridge Schedulerの有効化
	schedules, err := listEventBridgeSchedulersWithFilter(schedulerClient, filter)
	if err != nil {
		return fmt.Errorf("EventBridge Schedulersの取得に失敗: %w", err)
	}

	for _, schedule := range schedules {
		if schedule.State == "DISABLED" {
			if err := enableEventBridgeScheduler(schedulerClient, *schedule.Name); err != nil {
				fmt.Printf("  ⚠️  %s (Scheduler) の有効化に失敗: %v\n", *schedule.Name, err)
			} else {
				enabledCount++
			}
		}
	}

	fmt.Printf("\n✅ %d 個のスケジュールを有効化しました\n", enabledCount)
	return nil
}

// enableEventBridgeRule はEventBridge Ruleを有効化する
func enableEventBridgeRule(client *eventbridge.Client, name string) error {
	fmt.Printf("  ✓ %s (Rule) を有効化中...\n", name)
	_, err := client.EnableRule(context.Background(), &eventbridge.EnableRuleInput{
		Name: aws.String(name),
	})
	if err != nil {
		return err
	}
	fmt.Printf("    → 有効化完了\n")
	return nil
}

// enableEventBridgeScheduler はEventBridge Schedulerを有効化する
func enableEventBridgeScheduler(client *scheduler.Client, name string) error {
	ctx := context.Background()
	fmt.Printf("  ✓ %s (Scheduler) を有効化中...\n", name)

	// 現在の設定を取得
	getOutput, err := client.GetSchedule(ctx, &scheduler.GetScheduleInput{
		Name: aws.String(name),
	})
	if err != nil {
		return err
	}

	// 有効化
	updateInput := &scheduler.UpdateScheduleInput{
		Name:               aws.String(name),
		ScheduleExpression: getOutput.ScheduleExpression,
		State:              "ENABLED",
		Target:             getOutput.Target,
		FlexibleTimeWindow: getOutput.FlexibleTimeWindow,
	}
	if getOutput.Description != nil {
		updateInput.Description = getOutput.Description
	}
	if getOutput.GroupName != nil {
		updateInput.GroupName = getOutput.GroupName
	}

	_, err = client.UpdateSchedule(ctx, updateInput)
	if err != nil {
		return err
	}
	fmt.Printf("    → 有効化完了\n")
	return nil
}
