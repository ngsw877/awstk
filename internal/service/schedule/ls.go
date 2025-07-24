package schedule

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	eventbridgetypes "github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/aws/aws-sdk-go-v2/service/scheduler"
	schedulertypes "github.com/aws/aws-sdk-go-v2/service/scheduler/types"
)

// ListSchedules はスケジュール一覧を取得する
func ListSchedules(eventBridgeClient *eventbridge.Client, schedulerClient *scheduler.Client, opts ListOptions) ([]Schedule, error) {
	var schedules []Schedule
	ctx := context.Background()

	// EventBridge Rulesを取得
	if opts.Type == "all" || opts.Type == "rule" {
		rules, err := listEventBridgeRules(ctx, eventBridgeClient)
		if err != nil {
			return nil, fmt.Errorf("EventBridge Rules取得エラー: %w", err)
		}
		schedules = append(schedules, rules...)
	}

	// EventBridge Schedulerを取得
	if opts.Type == "all" || opts.Type == "scheduler" {
		schedulerList, err := listEventBridgeSchedulers(ctx, schedulerClient)
		if err != nil {
			return nil, fmt.Errorf("EventBridge Scheduler取得エラー: %w", err)
		}
		schedules = append(schedules, schedulerList...)
	}

	return schedules, nil
}

// listEventBridgeRules はEventBridge Rules（スケジュールタイプ）を取得
func listEventBridgeRules(ctx context.Context, client *eventbridge.Client) ([]Schedule, error) {
	var schedules []Schedule

	// ルール一覧を取得
	listOutput, err := client.ListRules(ctx, &eventbridge.ListRulesInput{})
	if err != nil {
		return nil, err
	}

	for _, rule := range listOutput.Rules {
		// スケジュール式を持つルールのみ対象
		if rule.ScheduleExpression == nil || *rule.ScheduleExpression == "" {
			continue
		}

		// ターゲット情報を取得
		targetsOutput, err := client.ListTargetsByRule(ctx, &eventbridge.ListTargetsByRuleInput{
			Rule: rule.Name,
		})
		if err != nil {
			return nil, fmt.Errorf("ルール %s のターゲット取得エラー: %w", *rule.Name, err)
		}

		// ターゲットを簡潔に表現
		target := formatTargets(targetsOutput.Targets)

		schedules = append(schedules, Schedule{
			Name:       getString(rule.Name),
			Type:       "rule",
			Expression: getString(rule.ScheduleExpression),
			State:      string(rule.State),
			Target:     target,
			Arn:        getString(rule.Arn),
		})
	}

	return schedules, nil
}

// listEventBridgeSchedulers はEventBridge Schedulerを取得
func listEventBridgeSchedulers(ctx context.Context, client *scheduler.Client) ([]Schedule, error) {
	var schedules []Schedule

	// スケジュール一覧を取得
	listOutput, err := client.ListSchedules(ctx, &scheduler.ListSchedulesInput{})
	if err != nil {
		return nil, err
	}

	for _, sched := range listOutput.Schedules {
		// 詳細情報を取得
		getOutput, err := client.GetSchedule(ctx, &scheduler.GetScheduleInput{
			Name: sched.Name,
		})
		if err != nil {
			return nil, fmt.Errorf("スケジュール %s の詳細取得エラー: %w", *sched.Name, err)
		}

		// スケジュール式を構築
		expression := formatScheduleExpression(getOutput)

		// ターゲットを簡潔に表現
		target := formatSchedulerTarget(getOutput.Target)

		schedules = append(schedules, Schedule{
			Name:       getString(sched.Name),
			Type:       "scheduler",
			Expression: expression,
			State:      string(getOutput.State),
			Target:     target,
			Arn:        getString(sched.Arn),
		})
	}

	return schedules, nil
}

// formatTargets はEventBridge Rulesのターゲットを簡潔に表現
func formatTargets(targets []eventbridgetypes.Target) string {
	if len(targets) == 0 {
		return "なし"
	}

	var targetStrs []string
	for _, target := range targets {
		if target.Arn != nil {
			// ARNからサービスとリソース名を抽出
			arnParts := strings.Split(*target.Arn, ":")
			if len(arnParts) >= 6 {
				service := arnParts[2]
				resourceType := arnParts[5]

				// Lambda関数の場合
				if service == "lambda" && strings.HasPrefix(resourceType, "function:") {
					funcName := strings.TrimPrefix(resourceType, "function:")
					targetStrs = append(targetStrs, fmt.Sprintf("Lambda:%s", funcName))
				} else if service == "sns" {
					targetStrs = append(targetStrs, fmt.Sprintf("SNS:%s", resourceType))
				} else if service == "sqs" {
					targetStrs = append(targetStrs, fmt.Sprintf("SQS:%s", resourceType))
				} else {
					targetStrs = append(targetStrs, fmt.Sprintf("%s:%s", service, resourceType))
				}
			} else {
				targetStrs = append(targetStrs, *target.Arn)
			}
		}
	}

	return strings.Join(targetStrs, ", ")
}

// formatSchedulerTarget はEventBridge Schedulerのターゲットを簡潔に表現
func formatSchedulerTarget(target *schedulertypes.Target) string {
	if target == nil || target.Arn == nil {
		return "なし"
	}

	// ARNからサービスとリソース名を抽出
	arnParts := strings.Split(*target.Arn, ":")
	if len(arnParts) >= 6 {
		service := arnParts[2]
		resourceType := arnParts[5]

		// サービス別の表現
		switch service {
		case "lambda":
			if strings.HasPrefix(resourceType, "function:") {
				funcName := strings.TrimPrefix(resourceType, "function:")
				return fmt.Sprintf("Lambda:%s", funcName)
			}
		case "states":
			return fmt.Sprintf("StepFunc:%s", resourceType)
		case "sns":
			return fmt.Sprintf("SNS:%s", resourceType)
		case "sqs":
			return fmt.Sprintf("SQS:%s", resourceType)
		case "events":
			return fmt.Sprintf("EventBus:%s", resourceType)
		}

		return fmt.Sprintf("%s:%s", service, resourceType)
	}

	return *target.Arn
}

// formatScheduleExpression はEventBridge Schedulerのスケジュール式を構築
func formatScheduleExpression(schedule *scheduler.GetScheduleOutput) string {
	if schedule.ScheduleExpression != nil {
		return *schedule.ScheduleExpression
	}

	// FlexibleTimeWindowがある場合
	if schedule.FlexibleTimeWindow != nil && schedule.FlexibleTimeWindow.Mode == schedulertypes.FlexibleTimeWindowModeFlexible {
		if schedule.FlexibleTimeWindow.MaximumWindowInMinutes != nil {
			return fmt.Sprintf("flexible(%d min)", *schedule.FlexibleTimeWindow.MaximumWindowInMinutes)
		}
	}

	return "不明"
}

// getString は*stringを安全にstringに変換
func getString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
