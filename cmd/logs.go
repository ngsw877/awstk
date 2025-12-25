package cmd

import (
	"awstk/internal/service/common"
	logssvc "awstk/internal/service/logs"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/spf13/cobra"
)

var (
	logsClient      *cloudwatchlogs.Client
	logsDeleteExact bool
	logsDeleteForce bool
)

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

// logsDeleteCmd represents the delete command
var logsDeleteCmd = &cobra.Command{
	Use:   "delete [log-group-names...]",
	Short: "CloudWatch Logsグループを削除するコマンド",
	Long: `指定したCloudWatch Logsグループを削除します。
ロググループ名の直接指定と検索パターンの両方に対応しています。
削除保護が有効な場合は --force オプションで保護を解除して削除できます。

【使い方】
  ` + AppName + ` logs delete my-log-group                    # 単一のロググループを削除
  ` + AppName + ` logs delete log1 log2 log3                  # 複数のロググループを削除
  ` + AppName + ` logs delete --search "/aws/lambda/*"        # パターンに一致するロググループを削除
  ` + AppName + ` logs delete --search "test-*" prod-log      # 検索パターンと直接指定の組み合わせ
  ` + AppName + ` logs delete --search "*" --empty-only       # 空のロググループをすべて削除
  ` + AppName + ` logs delete --search "*" --no-retention     # 保存期間未設定のロググループを削除
  ` + AppName + ` logs delete -s "prod-*" --force             # 削除保護を解除して削除

【例】
  ` + AppName + ` logs delete /aws/lambda/my-function
  → 指定したLambda関数のロググループを削除します。

  ` + AppName + ` logs delete --search "test-*" --empty-only
  → test-で始まる空のロググループのみを削除します。`,
	RunE: func(cmdCobra *cobra.Command, args []string) error {
		search, _ := cmdCobra.Flags().GetString("search")
		emptyOnly, _ := cmdCobra.Flags().GetBool("empty-only")
		noRetention, _ := cmdCobra.Flags().GetBool("no-retention")

		// 引数も検索パターンも指定されていない場合はエラー
		if len(args) == 0 && search == "" {
			return fmt.Errorf("削除対象のロググループ名または検索パターンを指定してください")
		}

		opts := logssvc.DeleteOptions{
			Filter:      search,
			LogGroups:   args,
			EmptyOnly:   emptyOnly,
			NoRetention: noRetention,
			Exact:       logsDeleteExact,
			Force:       logsDeleteForce,
		}

		return logssvc.DeleteLogGroups(logsClient, opts)
	},
	SilenceUsage: true,
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
	LogsCmd.AddCommand(logsDeleteCmd)

	// ls コマンドのフラグ
	logsLsCmd.Flags().BoolP("empty-only", "e", false, "空のログループのみを表示")
	logsLsCmd.Flags().BoolP("no-retention", "n", false, "保存期間が未設定のログのみを表示")
	logsLsCmd.Flags().BoolP("details", "d", false, "詳細情報を表示")

	// delete コマンドのフラグ
	logsDeleteCmd.Flags().StringP("search", "s", "", "削除対象の検索パターン（ワイルドカード対応）")
	logsDeleteCmd.Flags().BoolP("empty-only", "e", false, "空のログループのみを削除")
	logsDeleteCmd.Flags().BoolP("no-retention", "n", false, "保存期間が未設定のログのみを削除")
	logsDeleteCmd.Flags().BoolVar(&logsDeleteExact, "exact", false, "大文字小文字を区別してマッチ")
	logsDeleteCmd.Flags().BoolVar(&logsDeleteForce, "force", false, "削除保護を解除して削除")
}
