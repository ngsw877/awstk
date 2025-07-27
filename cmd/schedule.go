package cmd

import (
	"awstk/internal/service/schedule"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/scheduler"
	"github.com/spf13/cobra"
)

var (
	scheduleType string
	// trigger サブコマンド用フラグ
	triggerType    string
	triggerTimeout int
	triggerNoWait  bool
)

// ScheduleCmd はscheduleコマンドを表す
var ScheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "EventBridgeスケジュール管理コマンド",
	Long:  `EventBridge RulesとEventBridge Schedulerのスケジュールを管理するためのコマンド群です。`,
}

var scheduleLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "スケジュール一覧を表示",
	Long: `EventBridge Rules（スケジュールタイプ）とEventBridge Schedulerの一覧を表示します。

例:
  ` + AppName + ` schedule ls                    # 両方のスケジュールを表示
  ` + AppName + ` schedule ls --type rule       # EventBridge Rulesのみ表示
  ` + AppName + ` schedule ls --type scheduler  # EventBridge Schedulerのみ表示`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// クライアント生成
		eventBridgeClient := eventbridge.NewFromConfig(awsCfg)
		schedulerClient := scheduler.NewFromConfig(awsCfg)

		// オプション設定
		opts := schedule.ListOptions{
			Type: scheduleType,
		}

		// スケジュール一覧取得
		schedules, err := schedule.ListSchedules(eventBridgeClient, schedulerClient, opts)
		if err != nil {
			return fmt.Errorf("スケジュール一覧の取得に失敗: %w", err)
		}

		// 表示
		schedule.DisplaySchedules(schedules)

		return nil
	},
	SilenceUsage: true,
}

var scheduleTriggerCmd = &cobra.Command{
	Use:   "trigger NAME",
	Short: "スケジュールを手動実行",
	Long: `EventBridge RuleまたはEventBridge Schedulerを手動で実行します。
スケジュールを一時的に"rate(1 minute)"に変更し、実行後に元に戻します。

例:
  ` + AppName + ` schedule trigger my-rule              # 自動でタイプを判別
  ` + AppName + ` schedule trigger my-rule --type rule   # EventBridge Ruleとして実行
  ` + AppName + ` schedule trigger my-scheduler --no-wait # 待機せずに終了`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// クライアント生成
		eventBridgeClient := eventbridge.NewFromConfig(awsCfg)
		schedulerClient := scheduler.NewFromConfig(awsCfg)

		// オプション設定
		opts := schedule.TriggerOptions{
			Type:    triggerType,
			Timeout: triggerTimeout,
			NoWait:  triggerNoWait,
		}

		// スケジュール実行
		return schedule.TriggerSchedule(eventBridgeClient, schedulerClient, name, opts)
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(ScheduleCmd)
	ScheduleCmd.AddCommand(scheduleLsCmd)
	ScheduleCmd.AddCommand(scheduleTriggerCmd)

	// フラグ定義
	scheduleLsCmd.Flags().StringVarP(&scheduleType, "type", "t", "all", "表示タイプ (all|rule|scheduler)")

	// trigger サブコマンドのフラグ
	scheduleTriggerCmd.Flags().StringVar(&triggerType, "type", "", "スケジュールタイプ (rule|scheduler) ※省略時は自動判別")
	scheduleTriggerCmd.Flags().IntVar(&triggerTimeout, "timeout", 90, "実行待機時間（秒）")
	scheduleTriggerCmd.Flags().BoolVar(&triggerNoWait, "no-wait", false, "実行を待たずに終了")
}
