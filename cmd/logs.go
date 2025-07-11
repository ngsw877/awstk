package cmd

import (
	"awstk/internal/service/common"
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
			return common.FormatListError("CloudWatch Logsグループ", err)
		}

		if len(logGroups) == 0 {
			fmt.Println(common.FormatEmptyMessage("CloudWatch Logsグループ"))
			return nil
		}

		// フィルタリング処理とタイトル生成
		filteredGroups := logGroups
		var conditions []string
		
		if emptyOnly {
			conditions = append(conditions, "空の")
			filteredGroups = logssvc.FilterEmptyLogGroups(filteredGroups)
		}
		if noRetention {
			conditions = append(conditions, "保存期間未設定の")
			filteredGroups = logssvc.FilterNoRetentionLogGroups(filteredGroups)
		}
		
		title := common.GenerateFilteredTitle("CloudWatch Logsグループ", conditions...)

		// 結果表示
		if !showDetails {
			// シンプル表示
			names := make([]string, len(filteredGroups))
			for i, group := range filteredGroups {
				names[i] = *group.LogGroupName
			}
			common.PrintSimpleList(common.ListOutput{
				Title:        title,
				Items:        names,
				ResourceName: "ログループ",
				ShowCount:    true,
			})
		} else {
			// 詳細表示
			fmt.Printf("%s:\n", title)
			if len(filteredGroups) == 0 {
				fmt.Println("該当するログループはありませんでした")
				return nil
			}
			for _, group := range filteredGroups {
				logssvc.DisplayLogGroupDetails(group)
			}
			fmt.Printf("\n合計: %d個のログループ\n", len(filteredGroups))
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
	logsLsCmd.Flags().BoolP("no-retention", "n", false, "保存期間が未設定のログのみを表示")
	logsLsCmd.Flags().BoolP("details", "d", false, "詳細情報を表示")
}