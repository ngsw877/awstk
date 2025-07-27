package schedule

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/scheduler"
	"github.com/schollz/progressbar/v3"
)

// TriggerOptions はトリガー実行時のオプション
type TriggerOptions struct {
	Timeout int  // 実行待機時間（秒）
	NoWait  bool // 実行を待たずに終了
}

// TriggerSchedule はスケジュールを手動実行する
func TriggerSchedule(eventBridgeClient *eventbridge.Client, schedulerClient *scheduler.Client, name string, opts TriggerOptions) error {
	ctx := context.Background()

	// スケジュールタイプの判別
	scheduleType, err := detectScheduleType(ctx, eventBridgeClient, schedulerClient, name)
	if err != nil {
		return err
	}

	// タイプに応じて処理を分岐
	if scheduleType == "rule" {
		return triggerEventBridgeRule(ctx, eventBridgeClient, name, opts)
	}
	return triggerEventBridgeScheduler(ctx, schedulerClient, name, opts)
}

// detectScheduleType はスケジュールのタイプを自動判別する
func detectScheduleType(ctx context.Context, eventBridgeClient *eventbridge.Client, schedulerClient *scheduler.Client, name string) (string, error) {
	// 並列でチェック
	type result struct {
		scheduleType string
		err          error
	}

	ch := make(chan result, 2)

	// EventBridge Ruleチェック
	go func() {
		_, err := eventBridgeClient.DescribeRule(ctx, &eventbridge.DescribeRuleInput{
			Name: aws.String(name),
		})
		if err == nil {
			ch <- result{"rule", nil}
		} else {
			ch <- result{"", err}
		}
	}()

	// EventBridge Schedulerチェック
	go func() {
		_, err := schedulerClient.GetSchedule(ctx, &scheduler.GetScheduleInput{
			Name: aws.String(name),
		})
		if err == nil {
			ch <- result{"scheduler", nil}
		} else {
			ch <- result{"", err}
		}
	}()

	// 結果を確認
	var hasRule, hasScheduler bool
	for i := 0; i < 2; i++ {
		res := <-ch
		if res.err == nil {
			switch res.scheduleType {
			case "rule":
				hasRule = true
			case "scheduler":
				hasScheduler = true
			}
		}
	}

	// 両方存在する場合は対話的に選択
	if hasRule && hasScheduler {
		return selectScheduleTypeInteractive(name)
	}

	if hasRule {
		return "rule", nil
	}
	if hasScheduler {
		return "scheduler", nil
	}

	return "", fmt.Errorf("スケジュール '%s' が見つかりません", name)
}

// selectScheduleTypeInteractive は対話的にスケジュールタイプを選択する
func selectScheduleTypeInteractive(name string) (string, error) {
	fmt.Printf("\n⚠️  '%s' はEventBridge RuleとSchedulerの両方に存在します。\n", name)
	fmt.Println("どちらを操作しますか？")
	fmt.Println()
	fmt.Println("  1) EventBridge Rule")
	fmt.Println("  2) EventBridge Scheduler")
	fmt.Println()
	fmt.Print("選択してください (1 or 2): ")

	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		fmt.Println("→ EventBridge Rule を選択しました")
		return "rule", nil
	case "2":
		fmt.Println("→ EventBridge Scheduler を選択しました")
		return "scheduler", nil
	default:
		return "", fmt.Errorf("無効な選択です: %s", choice)
	}
}

// triggerEventBridgeRule はEventBridge Ruleを手動実行する
func triggerEventBridgeRule(ctx context.Context, client *eventbridge.Client, name string, opts TriggerOptions) error {
	// 1. 現在のルール情報を取得
	fmt.Printf("📝 現在のスケジュール設定を取得中...\n")
	describeOutput, err := client.DescribeRule(ctx, &eventbridge.DescribeRuleInput{
		Name: aws.String(name),
	})
	if err != nil {
		return fmt.Errorf("ルールの取得に失敗: %w", err)
	}

	// スケジュールルールでない場合はエラー
	if describeOutput.ScheduleExpression == nil || *describeOutput.ScheduleExpression == "" {
		return fmt.Errorf("'%s' はスケジュールルールではありません", name)
	}

	originalSchedule := *describeOutput.ScheduleExpression
	fmt.Printf("  └─ 現在の設定: %s\n", originalSchedule)

	// 元の状態を保存
	originalState := describeOutput.State

	// 2. 確実に元に戻すためのdefer
	shouldRestore := false
	defer func() {
		if shouldRestore && !opts.NoWait {
			fmt.Println("\n🔄 元のスケジュールに復元中...")
			putRuleInput := &eventbridge.PutRuleInput{
				Name:               aws.String(name),
				ScheduleExpression: aws.String(originalSchedule),
				State:              originalState,
			}
			if describeOutput.Description != nil {
				putRuleInput.Description = describeOutput.Description
			}
			if describeOutput.EventBusName != nil {
				putRuleInput.EventBusName = describeOutput.EventBusName
			}

			if _, err := client.PutRule(ctx, putRuleInput); err != nil {
				fmt.Printf("⚠️  スケジュールの復元に失敗: %v\n", err)
			} else {
				fmt.Printf("  └─ 復元後: %s\n", originalSchedule)
			}
		}
	}()

	// 3. スケジュールを rate(1 minute) に変更
	fmt.Println("\n🔄 スケジュールを1分後実行に変更中...")
	newSchedule := "rate(1 minute)"
	putRuleInput := &eventbridge.PutRuleInput{
		Name:               aws.String(name),
		ScheduleExpression: aws.String(newSchedule),
		State:              "ENABLED",
	}
	if describeOutput.Description != nil {
		putRuleInput.Description = describeOutput.Description
	}
	if describeOutput.EventBusName != nil {
		putRuleInput.EventBusName = describeOutput.EventBusName
	}

	if _, err := client.PutRule(ctx, putRuleInput); err != nil {
		return fmt.Errorf("スケジュール変更に失敗: %w", err)
	}
	fmt.Printf("  └─ 新しい設定: %s\n", newSchedule)
	shouldRestore = true

	// 4. 実行待機
	if !opts.NoWait {
		if err := waitForExecution(name, opts.Timeout); err != nil {
			return err
		}
	} else {
		fmt.Println("\n⚠️  --no-waitが指定されました。スケジュールは自動的に復元されません。")
		shouldRestore = false
	}

	fmt.Println("\n✅ 処理が完了しました")
	return nil
}

// triggerEventBridgeScheduler はEventBridge Schedulerを手動実行する
func triggerEventBridgeScheduler(ctx context.Context, client *scheduler.Client, name string, opts TriggerOptions) error {
	// 1. 現在のスケジュール情報を取得
	fmt.Printf("📝 現在のスケジュール設定を取得中...\n")
	getOutput, err := client.GetSchedule(ctx, &scheduler.GetScheduleInput{
		Name: aws.String(name),
	})
	if err != nil {
		return fmt.Errorf("スケジュールの取得に失敗: %w", err)
	}

	originalSchedule := *getOutput.ScheduleExpression
	fmt.Printf("  └─ 現在の設定: %s\n", originalSchedule)

	// 元の状態を保存
	originalState := getOutput.State

	// 2. 確実に元に戻すためのdefer
	shouldRestore := false
	defer func() {
		if shouldRestore && !opts.NoWait {
			fmt.Println("\n🔄 元のスケジュールに復元中...")
			updateInput := &scheduler.UpdateScheduleInput{
				Name:               aws.String(name),
				ScheduleExpression: aws.String(originalSchedule),
				State:              originalState,
				Target:             getOutput.Target,
				FlexibleTimeWindow: getOutput.FlexibleTimeWindow,
			}
			if getOutput.Description != nil {
				updateInput.Description = getOutput.Description
			}
			if getOutput.GroupName != nil {
				updateInput.GroupName = getOutput.GroupName
			}

			if _, err := client.UpdateSchedule(ctx, updateInput); err != nil {
				fmt.Printf("⚠️  スケジュールの復元に失敗: %v\n", err)
			} else {
				fmt.Printf("  └─ 復元後: %s\n", originalSchedule)
			}
		}
	}()

	// 3. スケジュールを rate(1 minute) に変更
	fmt.Println("\n🔄 スケジュールを1分後実行に変更中...")
	newSchedule := "rate(1 minute)"
	updateInput := &scheduler.UpdateScheduleInput{
		Name:               aws.String(name),
		ScheduleExpression: aws.String(newSchedule),
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

	if _, err := client.UpdateSchedule(ctx, updateInput); err != nil {
		return fmt.Errorf("スケジュール変更に失敗: %w", err)
	}
	fmt.Printf("  └─ 新しい設定: %s\n", newSchedule)
	shouldRestore = true

	// 4. 実行待機
	if !opts.NoWait {
		if err := waitForExecution(name, opts.Timeout); err != nil {
			return err
		}
	} else {
		fmt.Println("\n⚠️  --no-waitが指定されました。スケジュールは自動的に復元されません。")
		shouldRestore = false
	}

	fmt.Println("\n✅ 処理が完了しました")
	return nil
}

// waitForExecution は実行を待機する
func waitForExecution(name string, timeout int) error {
	// EventBridgeがスケジュール変更を認識するまでの時間 + rate(1 minute)の実行時間を考慮
	minWaitTime := 70
	actualWaitTime := timeout
	if actualWaitTime < minWaitTime {
		actualWaitTime = minWaitTime
		fmt.Printf("\n⚠️  最低待機時間%d秒に調整しました\n", minWaitTime)
	}

	fmt.Printf("\n⏳ スケジュール実行を待機中（%d秒）...\n", actualWaitTime)

	// プログレスバー表示
	bar := progressbar.NewOptions(actualWaitTime,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(false),
		progressbar.OptionSetWidth(40),
		progressbar.OptionSetDescription("待機中..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionShowCount(),
		progressbar.OptionShowElapsedTimeOnFinish(),
	)

	for i := 0; i < actualWaitTime; i++ {
		time.Sleep(1 * time.Second)
		if err := bar.Add(1); err != nil {
			fmt.Printf("⚠️  プログレスバー更新エラー: %v\n", err)
		}
	}

	if err := bar.Finish(); err != nil {
		fmt.Printf("⚠️  プログレスバー終了エラー: %v\n", err)
	}
	fmt.Println("\n✓ 実行待機完了")

	return nil
}
