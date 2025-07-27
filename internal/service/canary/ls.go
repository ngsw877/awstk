package canary

import (
	"context"
	"fmt"
	"time"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/synthetics"

	"awstk/internal/service/common"
)

// ListCanaries cmdから呼ばれるメイン関数（Get + Display）
func ListCanaries(client *synthetics.Client) error {
	// Get: データ取得
	canaries, err := getAllCanaries(client)
	if err != nil {
		return common.FormatListError("Canary", err)
	}

	// Display: 共通表示処理
	return common.DisplayList(
		canaries,
		"Canary一覧",
		canariesToTableData,
		&common.DisplayOptions{
			ShowCount:    true,
			EmptyMessage: "Canaryが見つかりませんでした",
		},
	)
}

// getAllCanaries 全てのCanaryを取得
func getAllCanaries(client *synthetics.Client) ([]Canary, error) {
	resp, err := client.DescribeCanaries(context.Background(), &synthetics.DescribeCanariesInput{})
	if err != nil {
		return nil, fmt.Errorf("canary一覧の取得に失敗: %w", err)
	}

	canaries := make([]Canary, 0, len(resp.Canaries))
	for _, c := range resp.Canaries {
		canary := Canary{
			Name:           awssdk.ToString(c.Name),
			State:          string(c.Status.State),
			StateReason:    awssdk.ToString(c.Status.StateReason),
			RuntimeVersion: awssdk.ToString(c.RuntimeVersion),
		}

		// スケジュール情報
		if c.Schedule != nil && c.Schedule.Expression != nil {
			canary.Schedule = awssdk.ToString(c.Schedule.Expression)
		}

		// タイムライン情報（最終実行時刻）
		if c.Timeline != nil && c.Timeline.LastModified != nil {
			canary.LastRunTime = awssdk.ToTime(c.Timeline.LastModified)
		}

		// 最新の実行結果を取得
		runs, err := client.DescribeCanariesLastRun(context.Background(), &synthetics.DescribeCanariesLastRunInput{})
		if err == nil && len(runs.CanariesLastRun) > 0 {
			for _, lastRun := range runs.CanariesLastRun {
				if awssdk.ToString(lastRun.CanaryName) == canary.Name {
					if lastRun.LastRun != nil && lastRun.LastRun.Status != nil {
						canary.LastRunStatus = string(lastRun.LastRun.Status.State)
					}
					break
				}
			}
		}

		// 成功率の計算（直近の実行結果から）
		canary.SuccessRate = calculateSuccessRate(client, canary.Name)

		canaries = append(canaries, canary)
	}

	return canaries, nil
}

// calculateSuccessRate 直近の実行結果から成功率を計算
func calculateSuccessRate(client *synthetics.Client, canaryName string) float64 {
	// 直近100件の実行結果を取得
	runs, err := client.GetCanaryRuns(context.Background(), &synthetics.GetCanaryRunsInput{
		Name:       awssdk.String(canaryName),
		MaxResults: awssdk.Int32(100),
	})
	if err != nil || len(runs.CanaryRuns) == 0 {
		return 0.0
	}

	successCount := 0
	for _, run := range runs.CanaryRuns {
		if run.Status != nil {
			if string(run.Status.State) == CanaryRunStatePassed {
				successCount++
			}
		}
	}

	return float64(successCount) / float64(len(runs.CanaryRuns)) * 100
}

// canariesToTableData Canary情報をテーブルデータに変換
func canariesToTableData(canaries []Canary) ([]common.TableColumn, [][]string) {
	columns := []common.TableColumn{
		{Header: "名前"},
		{Header: "状態"},
		{Header: "スケジュール"},
		{Header: "成功率"},
		{Header: "最終実行"},
		{Header: "最終結果"},
	}

	data := make([][]string, len(canaries))
	for i, c := range canaries {
		data[i] = []string{
			c.Name,
			formatState(c.State),
			formatSchedule(c.Schedule),
			fmt.Sprintf("%.1f%%", c.SuccessRate),
			formatLastRunTime(c.LastRunTime),
			formatRunStatus(c.LastRunStatus),
		}
	}
	return columns, data
}

// formatState 状態を見やすくフォーマット
func formatState(state string) string {
	switch state {
	case CanaryStateRunning:
		return "[実行中]"
	case CanaryStateStopped:
		return "[停止中]"
	case CanaryStateError:
		return "[エラー]"
	case CanaryStateStarting:
		return "[開始中]"
	case CanaryStateStopping:
		return "[停止中]"
	default:
		return state
	}
}

// formatSchedule スケジュールを見やすくフォーマット
func formatSchedule(schedule string) string {
	if schedule == "" {
		return "なし"
	}
	// rate(5 minutes) -> 5分ごと
	if len(schedule) > 5 && schedule[:5] == "rate(" {
		return "定期: " + schedule[5:len(schedule)-1]
	}
	// cron式はそのまま表示
	return schedule
}

// formatLastRunTime 最終実行時刻をフォーマット
func formatLastRunTime(t time.Time) string {
	if t.IsZero() {
		return "未実行"
	}
	return t.Format("01/02 15:04")
}

// formatRunStatus 実行結果ステータスをフォーマット
func formatRunStatus(status string) string {
	switch status {
	case CanaryRunStatePassed:
		return "[成功]"
	case CanaryRunStateFailed:
		return "[失敗]"
	case CanaryRunStateRunning:
		return "[実行中]"
	default:
		if status == "" {
			return "-"
		}
		return status
	}
}
