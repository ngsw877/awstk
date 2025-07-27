package schedule

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/scheduler"
)

// EnableSchedule は単一のスケジュールを有効化する
func EnableSchedule(eventBridgeClient *eventbridge.Client, schedulerClient *scheduler.Client, name string) error {
	ctx := context.Background()

	// スケジュールタイプの判別
	scheduleType, err := detectScheduleType(ctx, eventBridgeClient, schedulerClient, name)
	if err != nil {
		return err
	}

	// タイプに応じて処理を分岐
	if scheduleType == "rule" {
		return enableEventBridgeRule(ctx, eventBridgeClient, name)
	}
	return enableEventBridgeScheduler(ctx, schedulerClient, name)
}

// DisableSchedule は単一のスケジュールを無効化する
func DisableSchedule(eventBridgeClient *eventbridge.Client, schedulerClient *scheduler.Client, name string) error {
	ctx := context.Background()

	// スケジュールタイプの判別
	scheduleType, err := detectScheduleType(ctx, eventBridgeClient, schedulerClient, name)
	if err != nil {
		return err
	}

	// タイプに応じて処理を分岐
	if scheduleType == "rule" {
		return disableEventBridgeRule(ctx, eventBridgeClient, name)
	}
	return disableEventBridgeScheduler(ctx, schedulerClient, name)
}

// EnableSchedulesWithFilter はフィルターにマッチする全スケジュールを有効化する
func EnableSchedulesWithFilter(eventBridgeClient *eventbridge.Client, schedulerClient *scheduler.Client, filter string) error {
	ctx := context.Background()
	enabledCount := 0

	fmt.Printf("フィルター '%s' にマッチするスケジュールを検索中...\n", filter)

	// EventBridge Rulesの有効化
	rules, err := listEventBridgeRulesWithFilter(ctx, eventBridgeClient, filter)
	if err != nil {
		return fmt.Errorf("EventBridge Rulesの取得に失敗: %w", err)
	}

	for _, rule := range rules {
		if rule.State == "DISABLED" {
			if err := enableEventBridgeRule(ctx, eventBridgeClient, *rule.Name); err != nil {
				fmt.Printf("  ⚠️  %s (Rule) の有効化に失敗: %v\n", *rule.Name, err)
			} else {
				enabledCount++
			}
		}
	}

	// EventBridge Schedulerの有効化
	schedules, err := listEventBridgeSchedulersWithFilter(ctx, schedulerClient, filter)
	if err != nil {
		return fmt.Errorf("EventBridge Schedulersの取得に失敗: %w", err)
	}

	for _, schedule := range schedules {
		if schedule.State == "DISABLED" {
			if err := enableEventBridgeScheduler(ctx, schedulerClient, *schedule.Name); err != nil {
				fmt.Printf("  ⚠️  %s (Scheduler) の有効化に失敗: %v\n", *schedule.Name, err)
			} else {
				enabledCount++
			}
		}
	}

	fmt.Printf("\n✅ %d 個のスケジュールを有効化しました\n", enabledCount)
	return nil
}

// DisableSchedulesWithFilter はフィルターにマッチする全スケジュールを無効化する
func DisableSchedulesWithFilter(eventBridgeClient *eventbridge.Client, schedulerClient *scheduler.Client, filter string) error {
	ctx := context.Background()
	disabledCount := 0

	fmt.Printf("フィルター '%s' にマッチするスケジュールを検索中...\n", filter)

	// EventBridge Rulesの無効化
	rules, err := listEventBridgeRulesWithFilter(ctx, eventBridgeClient, filter)
	if err != nil {
		return fmt.Errorf("EventBridge Rulesの取得に失敗: %w", err)
	}

	for _, rule := range rules {
		if rule.State == "ENABLED" || rule.State == "ENABLED_WITH_ALL_CLOUDTRAIL_MANAGEMENT_EVENTS" {
			if err := disableEventBridgeRule(ctx, eventBridgeClient, *rule.Name); err != nil {
				fmt.Printf("  ⚠️  %s (Rule) の無効化に失敗: %v\n", *rule.Name, err)
			} else {
				disabledCount++
			}
		}
	}

	// EventBridge Schedulerの無効化
	schedules, err := listEventBridgeSchedulersWithFilter(ctx, schedulerClient, filter)
	if err != nil {
		return fmt.Errorf("EventBridge Schedulersの取得に失敗: %w", err)
	}

	for _, schedule := range schedules {
		if schedule.State == "ENABLED" {
			if err := disableEventBridgeScheduler(ctx, schedulerClient, *schedule.Name); err != nil {
				fmt.Printf("  ⚠️  %s (Scheduler) の無効化に失敗: %v\n", *schedule.Name, err)
			} else {
				disabledCount++
			}
		}
	}

	fmt.Printf("\n✅ %d 個のスケジュールを無効化しました\n", disabledCount)
	return nil
}

// enableEventBridgeRule はEventBridge Ruleを有効化する
func enableEventBridgeRule(ctx context.Context, client *eventbridge.Client, name string) error {
	fmt.Printf("  ✓ %s (Rule) を有効化中...\n", name)
	_, err := client.EnableRule(ctx, &eventbridge.EnableRuleInput{
		Name: aws.String(name),
	})
	if err != nil {
		return err
	}
	fmt.Printf("    → 有効化完了\n")
	return nil
}

// disableEventBridgeRule はEventBridge Ruleを無効化する
func disableEventBridgeRule(ctx context.Context, client *eventbridge.Client, name string) error {
	fmt.Printf("  ✓ %s (Rule) を無効化中...\n", name)
	_, err := client.DisableRule(ctx, &eventbridge.DisableRuleInput{
		Name: aws.String(name),
	})
	if err != nil {
		return err
	}
	fmt.Printf("    → 無効化完了\n")
	return nil
}

// enableEventBridgeScheduler はEventBridge Schedulerを有効化する
func enableEventBridgeScheduler(ctx context.Context, client *scheduler.Client, name string) error {
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

// disableEventBridgeScheduler はEventBridge Schedulerを無効化する
func disableEventBridgeScheduler(ctx context.Context, client *scheduler.Client, name string) error {
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

// listEventBridgeRulesWithFilter はフィルターにマッチするEventBridge Rulesを取得する
func listEventBridgeRulesWithFilter(ctx context.Context, client *eventbridge.Client, filter string) ([]*eventbridge.DescribeRuleOutput, error) {
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
			if rule.Name != nil && matchPattern(*rule.Name, filter) && rule.ScheduleExpression != nil {
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
func listEventBridgeSchedulersWithFilter(ctx context.Context, client *scheduler.Client, filter string) ([]*scheduler.GetScheduleOutput, error) {
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
			if schedule.Name != nil && matchPattern(*schedule.Name, filter) {
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

// matchPattern はワイルドカードパターンマッチングを行う
func matchPattern(name, pattern string) bool {
	// ワイルドカードを含む場合
	if strings.Contains(pattern, "*") {
		// glob パターンマッチング
		matched, _ := filepath.Match(pattern, name)
		return matched
	}
	// ワイルドカードなしの場合は部分一致
	return strings.Contains(name, pattern)
}
