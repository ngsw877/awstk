package cmd

import (
	"awstk/internal/service/canary"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/synthetics"
	"github.com/spf13/cobra"
)

var (
	canaryName       string
	canaryFilter     string
	canaryFilters    []string
	canaryAll        bool
	canaryYes        bool
	canaryDryRun     bool
	syntheticsClient *synthetics.Client
)

var CanaryCmd = &cobra.Command{
	Use:   "canary",
	Short: "AWS Synthetics Canary操作コマンド",
	Long: `AWS Synthetics Canaryの一覧表示、有効化/無効化、手動実行を行います。

使用例:
  ` + AppName + ` canary ls                          # Canary一覧を表示
  ` + AppName + ` canary enable --name my-canary     # 特定のCanaryを有効化
  ` + AppName + ` canary disable --filter "test-*"   # パターンに一致するCanaryを無効化
  ` + AppName + ` canary enable --all                # 全てのCanaryを有効化
  ` + AppName + ` canary run --name my-canary        # 特定のCanaryを手動実行
  ` + AppName + ` canary run --filter "api-*" --yes  # パターンに一致するCanaryを一括実行`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// 親のPersistentPreRunEを実行（awsCtx設定とAWS設定読み込み）
		if err := RootCmd.PersistentPreRunE(cmd, args); err != nil {
			return err
		}

		// Syntheticsクライアントを初期化
		syntheticsClient = synthetics.NewFromConfig(awsCfg)
		return nil
	},
}

var canaryLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "Canary一覧を表示するコマンド",
	Long:  `AWS Synthetics Canaryの一覧を表示します。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return canary.ListCanaries(syntheticsClient)
	},
	SilenceUsage: true,
}

var canaryEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Canaryを有効化するコマンド",
	Long: `指定したCanaryを有効化（開始）します。
    --name, --filter, --all のいずれかを指定してください。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if canaryAll {
			return canary.EnableAllCanaries(syntheticsClient, canaryYes)
		}
		if canaryFilter != "" {
			return canary.EnableCanariesByFilter(syntheticsClient, canaryFilter, canaryYes)
		}
		if canaryName != "" {
			return canary.EnableCanary(syntheticsClient, canaryName)
		}
		return fmt.Errorf("オプションが指定されていません")
	},
	SilenceUsage: true,
}

var canaryDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Canaryを無効化するコマンド",
	Long: `指定したCanaryを無効化（停止）します。
    --name, --filter, --all のいずれかを指定してください。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if canaryAll {
			return canary.DisableAllCanaries(syntheticsClient, canaryYes)
		}
		if canaryFilter != "" {
			return canary.DisableCanariesByFilter(syntheticsClient, canaryFilter, canaryYes)
		}
		if canaryName != "" {
			return canary.DisableCanary(syntheticsClient, canaryName)
		}
		return fmt.Errorf("オプションが指定されていません")
	},
	SilenceUsage: true,
}

var canaryRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Canaryを手動実行するコマンド",
	Long: `指定したCanaryを手動で実行します。
    --name または --filter を指定してください。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if canaryName != "" {
			if canaryDryRun {
				return canary.RunCanaryDryRun(syntheticsClient, canaryName)
			}
			return canary.RunCanary(syntheticsClient, canaryName)
		}
		if len(canaryFilters) > 0 {
			return canary.RunCanariesByFilter(syntheticsClient, canaryFilters, canaryDryRun, canaryYes)
		}
		return fmt.Errorf("--name または --filter のいずれかを指定してください")
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(CanaryCmd)
	CanaryCmd.AddCommand(canaryLsCmd)
	CanaryCmd.AddCommand(canaryEnableCmd)
	CanaryCmd.AddCommand(canaryDisableCmd)
	CanaryCmd.AddCommand(canaryRunCmd)

	// Enable/Disableコマンドのフラグ設定
	for _, cmd := range []*cobra.Command{canaryEnableCmd, canaryDisableCmd} {
		cmd.Flags().StringVarP(&canaryName, "name", "n", "", "Canary名")
		cmd.Flags().StringVarP(&canaryFilter, "filter", "f", "", "名前パターン（ワイルドカード対応）")
		cmd.Flags().BoolVarP(&canaryAll, "all", "a", false, "全てのCanaryを対象")
		cmd.Flags().BoolVarP(&canaryYes, "yes", "y", false, "確認なしで実行")
		// --name / --filter / --all は相互排他かついずれか必須
		cmd.MarkFlagsMutuallyExclusive("name", "filter", "all")
		cmd.MarkFlagsOneRequired("name", "filter", "all")
	}

	// Runコマンドのフラグ設定
	canaryRunCmd.Flags().StringVarP(&canaryName, "name", "n", "", "Canary名")
	canaryRunCmd.Flags().StringSliceVarP(&canaryFilters, "filter", "f", []string{}, "名前パターン（複数指定可能、ワイルドカード対応）")
	canaryRunCmd.Flags().BoolVarP(&canaryDryRun, "dry-run", "d", false, "ドライラン実行")
	canaryRunCmd.Flags().BoolVarP(&canaryYes, "yes", "y", false, "確認なしで実行")
	// --name と --filter は相互排他かついずれか必須
	canaryRunCmd.MarkFlagsMutuallyExclusive("name", "filter")
	canaryRunCmd.MarkFlagsOneRequired("name", "filter")
}
