package schedule

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/scheduler"
)

// DisableSchedule は単一のスケジュールを無効化する
func DisableSchedule(eventBridgeClient *eventbridge.Client, schedulerClient *scheduler.Client, name string) error {
	// スケジュールタイプの判別
	scheduleType, err := detectScheduleType(eventBridgeClient, schedulerClient, name)
	if err != nil {
		return err
	}

	// タイプに応じて処理を分岐
	if scheduleType == "rule" {
		return disableEventBridgeRule(eventBridgeClient, name)
	}
	return disableEventBridgeScheduler(schedulerClient, name)
}

// DisableSchedulesWithFilter はフィルターにマッチする全スケジュールを無効化する
func DisableSchedulesWithFilter(eventBridgeClient *eventbridge.Client, schedulerClient *scheduler.Client, filter string) error {
	disabledCount := 0

	fmt.Printf("フィルター '%s' にマッチするスケジュールを検索中...\n", filter)

	// EventBridge Rulesの無効化
	rules, err := listEventBridgeRulesWithFilter(eventBridgeClient, filter)
	if err != nil {
		return fmt.Errorf("EventBridge Rulesの取得に失敗: %w", err)
	}

	for _, rule := range rules {
		if rule.State == "ENABLED" || rule.State == "ENABLED_WITH_ALL_CLOUDTRAIL_MANAGEMENT_EVENTS" {
			if err := disableEventBridgeRule(eventBridgeClient, *rule.Name); err != nil {
				fmt.Printf("  ⚠️  %s (Rule) の無効化に失敗: %v\n", *rule.Name, err)
			} else {
				disabledCount++
			}
		}
	}

	// EventBridge Schedulerの無効化
	schedules, err := listEventBridgeSchedulersWithFilter(schedulerClient, filter)
	if err != nil {
		return fmt.Errorf("EventBridge Schedulersの取得に失敗: %w", err)
	}

	for _, schedule := range schedules {
		if schedule.State == "ENABLED" {
			if err := disableEventBridgeScheduler(schedulerClient, *schedule.Name); err != nil {
				fmt.Printf("  ⚠️  %s (Scheduler) の無効化に失敗: %v\n", *schedule.Name, err)
			} else {
				disabledCount++
			}
		}
	}

	fmt.Printf("\n✅ %d 個のスケジュールを無効化しました\n", disabledCount)
	return nil
}

// disableEventBridgeRule はEventBridge Ruleを無効化する
func disableEventBridgeRule(client *eventbridge.Client, name string) error {
	fmt.Printf("  ✓ %s (Rule) を無効化中...\n", name)
	_, err := client.DisableRule(context.Background(), &eventbridge.DisableRuleInput{
		Name: aws.String(name),
	})
	if err != nil {
		return err
	}
	fmt.Printf("    → 無効化完了\n")
	return nil
}

// disableEventBridgeScheduler はEventBridge Schedulerを無効化する
func disableEventBridgeScheduler(client *scheduler.Client, name string) error {
	ctx := context.Background()
	fmt.Printf("  ✓ %s (Scheduler) を無効化中...\n", name)

	// 現在の設定を取得
	getOutput, err := client.GetSchedule(ctx, &scheduler.GetScheduleInput{
		Name: aws.String(name),
	})
	if err != nil {
		return err
	}

	// 無効化
	updateInput := &scheduler.UpdateScheduleInput{
		Name:               aws.String(name),
		ScheduleExpression: getOutput.ScheduleExpression,
		State:              "DISABLED",
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
	fmt.Printf("    → 無効化完了\n")
	return nil
}
