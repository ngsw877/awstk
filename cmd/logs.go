package cmd

import (
	logssvc "awstk/internal/service/logs"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/spf13/cobra"
)

var logsClient *cloudwatchlogs.Client

// LogsCmd represents the logs command
var LogsCmd = &cobra.Command{
	Use:          "logs",
	Short:        "CloudWatch Logsリソース操作コマンド",
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// 親のPersistentPreRunEを実行（awsCtx設定とAWS設定読み込み）
		if err := RootCmd.PersistentPreRunE(cmd, args); err != nil {
			return err
		}

		// CloudWatch Logs用クライアント生成
		logsClient = cloudwatchlogs.NewFromConfig(awsCfg)

		return nil
	},
}

// logsLsCmd represents the ls command
var logsLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "CloudWatch Logsグループ一覧を表示するコマンド",
	Long: `CloudWatch Logsグループの一覧を表示します。
ログサイズやストリーム数、保存期間などの情報も含めて表示します。

【使い方】
  ` + AppName + ` logs ls                    # ログループ一覧を表示
  ` + AppName + ` logs ls -e                 # 空のログループのみを表示
  ` + AppName + ` logs ls --details          # 詳細情報付きで表示

【例】
  ` + AppName + ` logs ls -e
  → 空のCloudWatch Logsグループのみを一覧表示します。`,
	RunE: func(cmdCobra *cobra.Command, args []string) error {
		emptyOnly, _ := cmdCobra.Flags().GetBool("empty-only")
		showDetails, _ := cmdCobra.Flags().GetBool("details")

		// ログループ一覧を取得
		logGroups, err := logssvc.ListLogGroups(logsClient)
		if err != nil {
			return fmt.Errorf("❌ ログループ一覧取得でエラー: %w", err)
		}

		if len(logGroups) == 0 {
			fmt.Println("CloudWatch Logsグループが見つかりませんでした")
			return nil
		}

		// 空ロググループのみ表示する場合
		if emptyOnly {
			fmt.Println("空のCloudWatch Logsグループ一覧:")
			emptyGroups := logssvc.FilterEmptyLogGroups(logGroups)
			if len(emptyGroups) == 0 {
				fmt.Println("空のログループはありませんでした")
				return nil
			}
			for _, group := range emptyGroups {
				if showDetails {
					fmt.Printf("  - %s (作成日: %s)\n", 
						*group.LogGroupName, 
						logssvc.FormatTimestamp(group.CreationTime))
				} else {
					fmt.Printf("  - %s\n", *group.LogGroupName)
				}
			}
			fmt.Printf("\n合計: %d個の空ログループ\n", len(emptyGroups))
		} else {
			// 全ログループを表示
			fmt.Println("CloudWatch Logsグループ一覧:")
			for _, group := range logGroups {
				if showDetails {
					logssvc.DisplayLogGroupDetails(group)
				} else {
					fmt.Printf("  - %s\n", *group.LogGroupName)
				}
			}
			fmt.Printf("\n合計: %d個のログループ\n", len(logGroups))
		}

		return nil
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(LogsCmd)
	LogsCmd.AddCommand(logsLsCmd)

	// ls コマンドのフラグ
	logsLsCmd.Flags().BoolP("empty-only", "e", false, "空のログループのみを表示")
	logsLsCmd.Flags().BoolP("details", "d", false, "詳細情報を表示")
}