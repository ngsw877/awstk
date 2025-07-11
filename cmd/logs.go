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
  ` + AppName + ` logs ls -n                 # 保存期間が未設定のログのみを表示
  ` + AppName + ` logs ls --details          # 詳細情報付きで表示
  ` + AppName + ` logs ls -e -n              # 空かつ保存期間未設定のログを表示

【例】
  ` + AppName + ` logs ls -e
  → 空のCloudWatch Logsグループのみを一覧表示します。
  
  ` + AppName + ` logs ls -n -d
  → 保存期間が未設定のログを詳細情報付きで表示します。`,
	RunE: func(cmdCobra *cobra.Command, args []string) error {
		emptyOnly, _ := cmdCobra.Flags().GetBool("empty-only")
		noRetention, _ := cmdCobra.Flags().GetBool("no-retention")
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

		// フィルタリング処理
		filteredGroups := logGroups
		var title string

		// 複数フィルタの組み合わせ対応
		if emptyOnly && noRetention {
			title = "空かつ保存期間未設定のCloudWatch Logsグループ一覧:"
			filteredGroups = logssvc.FilterEmptyLogGroups(filteredGroups)
			filteredGroups = logssvc.FilterNoRetentionLogGroups(filteredGroups)
		} else if emptyOnly {
			title = "空のCloudWatch Logsグループ一覧:"
			filteredGroups = logssvc.FilterEmptyLogGroups(filteredGroups)
		} else if noRetention {
			title = "保存期間未設定のCloudWatch Logsグループ一覧:"
			filteredGroups = logssvc.FilterNoRetentionLogGroups(filteredGroups)
		} else {
			title = "CloudWatch Logsグループ一覧:"
		}

		// 結果表示
		fmt.Println(title)
		if len(filteredGroups) == 0 {
			fmt.Println("該当するログループはありませんでした")
			return nil
		}

		for _, group := range filteredGroups {
			if showDetails {
				logssvc.DisplayLogGroupDetails(group)
			} else {
				fmt.Printf("  - %s\n", *group.LogGroupName)
			}
		}
		fmt.Printf("\n合計: %d個のログループ\n", len(filteredGroups))

		return nil
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(LogsCmd)
	LogsCmd.AddCommand(logsLsCmd)

	// ls コマンドのフラグ
	logsLsCmd.Flags().BoolP("empty-only", "e", false, "空のログループのみを表示")
	logsLsCmd.Flags().BoolP("no-retention", "n", false, "保存期間が未設定のログのみを表示")
	logsLsCmd.Flags().BoolP("details", "d", false, "詳細情報を表示")
}