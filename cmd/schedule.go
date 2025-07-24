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

func init() {
	RootCmd.AddCommand(ScheduleCmd)
	ScheduleCmd.AddCommand(scheduleLsCmd)

	// フラグ定義
	scheduleLsCmd.Flags().StringVarP(&scheduleType, "type", "t", "all", "表示タイプ (all|rule|scheduler)")
}
