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
	triggerTimeout int
	triggerNoWait  bool
	// enable/disable サブコマンド用フラグ
	enableSearch  string
	disableSearch string
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
  ` + AppName + ` schedule trigger my-scheduler --no-wait # 待機せずに終了`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// クライアント生成
		eventBridgeClient := eventbridge.NewFromConfig(awsCfg)
		schedulerClient := scheduler.NewFromConfig(awsCfg)

		// オプション設定
		opts := schedule.TriggerOptions{
			Timeout: triggerTimeout,
			NoWait:  triggerNoWait,
		}

		// スケジュール実行
		return schedule.TriggerSchedule(eventBridgeClient, schedulerClient, name, opts)
	},
	SilenceUsage: true,
}

var scheduleEnableCmd = &cobra.Command{
	Use:   "enable NAME",
	Short: "スケジュールを有効化",
	Long: `EventBridge RuleまたはEventBridge Schedulerを有効化します。

例:
  ` + AppName + ` schedule enable my-rule                # 単一のスケジュールを有効化
  ` + AppName + ` schedule enable --search "batch-*"     # batch-で始まる全てを有効化
  ` + AppName + ` schedule enable --search "Scheduled"   # Scheduledを含む全てを有効化`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// クライアント生成
		eventBridgeClient := eventbridge.NewFromConfig(awsCfg)
		schedulerClient := scheduler.NewFromConfig(awsCfg)

		// 単一指定または検索パターン指定の確認
		if len(args) == 1 && enableSearch == "" {
			// 単一スケジュールの有効化
			return schedule.EnableSchedule(eventBridgeClient, schedulerClient, args[0])
		} else if len(args) == 0 && enableSearch != "" {
			// 検索パターンによる一括有効化
			return schedule.EnableSchedulesWithFilter(eventBridgeClient, schedulerClient, enableSearch)
		} else {
			return fmt.Errorf("スケジュール名または検索パターンのいずれか一方を指定してください")
		}
	},
	SilenceUsage: true,
}

var scheduleDisableCmd = &cobra.Command{
	Use:   "disable NAME",
	Short: "スケジュールを無効化",
	Long: `EventBridge RuleまたはEventBridge Schedulerを無効化します。

例:
  ` + AppName + ` schedule disable my-rule               # 単一のスケジュールを無効化
  ` + AppName + ` schedule disable --search "test-*"     # test-で始まる全てを無効化
  ` + AppName + ` schedule disable --search "Dev"        # Devを含む全てを無効化`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// クライアント生成
		eventBridgeClient := eventbridge.NewFromConfig(awsCfg)
		schedulerClient := scheduler.NewFromConfig(awsCfg)

		// 単一指定または検索パターン指定の確認
		if len(args) == 1 && disableSearch == "" {
			// 単一スケジュールの無効化
			return schedule.DisableSchedule(eventBridgeClient, schedulerClient, args[0])
		} else if len(args) == 0 && disableSearch != "" {
			// 検索パターンによる一括無効化
			return schedule.DisableSchedulesWithFilter(eventBridgeClient, schedulerClient, disableSearch)
		} else {
			return fmt.Errorf("スケジュール名または検索パターンのいずれか一方を指定してください")
		}
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(ScheduleCmd)
	ScheduleCmd.AddCommand(scheduleLsCmd)
	ScheduleCmd.AddCommand(scheduleTriggerCmd)
	ScheduleCmd.AddCommand(scheduleEnableCmd)
	ScheduleCmd.AddCommand(scheduleDisableCmd)

	// フラグ定義
	scheduleLsCmd.Flags().StringVarP(&scheduleType, "type", "t", "all", "表示タイプ (all|rule|scheduler)")

	// trigger サブコマンドのフラグ
	scheduleTriggerCmd.Flags().IntVar(&triggerTimeout, "timeout", 90, "実行待機時間（秒）")
	scheduleTriggerCmd.Flags().BoolVar(&triggerNoWait, "no-wait", false, "実行を待たずに終了")

	// enable サブコマンドのフラグ
	scheduleEnableCmd.Flags().StringVarP(&enableSearch, "search", "s", "", "有効化するスケジュールの検索パターン")

	// disable サブコマンドのフラグ
	scheduleDisableCmd.Flags().StringVarP(&disableSearch, "search", "s", "", "無効化するスケジュールの検索パターン")
}
