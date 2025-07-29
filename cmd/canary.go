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
	canaryAll        bool
	canaryYes        bool
	syntheticsClient *synthetics.Client
)

var CanaryCmd = &cobra.Command{
	Use:   "canary",
	Short: "AWS Synthetics Canary操作コマンド",
	Long: `AWS Synthetics Canaryの一覧表示、有効化/無効化を行います。

使用例:
  ` + AppName + ` canary ls                          # Canary一覧を表示
  ` + AppName + ` canary enable --name my-canary     # 特定のCanaryを有効化
  ` + AppName + ` canary disable --filter "test-*"   # パターンに一致するCanaryを無効化
  ` + AppName + ` canary enable --all                # 全てのCanaryを有効化`,
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
	PreRunE: validateCanaryFlags,
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
	PreRunE: validateCanaryFlags,
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

// validateCanaryFlags は排他的なフラグの検証を行う
func validateCanaryFlags(cmd *cobra.Command, args []string) error {
	return ValidateExclusiveOptions(true, true,
		canaryName != "",
		canaryFilter != "",
		canaryAll)
}

func init() {
	RootCmd.AddCommand(CanaryCmd)
	CanaryCmd.AddCommand(canaryLsCmd)
	CanaryCmd.AddCommand(canaryEnableCmd)
	CanaryCmd.AddCommand(canaryDisableCmd)

	// Enable/Disableコマンドのフラグ設定
	for _, cmd := range []*cobra.Command{canaryEnableCmd, canaryDisableCmd} {
		cmd.Flags().StringVarP(&canaryName, "name", "n", "", "Canary名")
		cmd.Flags().StringVarP(&canaryFilter, "filter", "f", "", "名前パターン（ワイルドカード対応）")
		cmd.Flags().BoolVarP(&canaryAll, "all", "a", false, "全てのCanaryを対象")
		cmd.Flags().BoolVarP(&canaryYes, "yes", "y", false, "確認なしで実行")
	}
}
