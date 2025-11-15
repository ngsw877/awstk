package cmd

import (
	elbsvc "awstk/internal/service/elb"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/spf13/cobra"
)

var elbv2Client *elasticloadbalancingv2.Client

// ElbCmd represents the elb command
var ElbCmd = &cobra.Command{
	Use:          "elb",
	Short:        "ELBリソース操作コマンド",
	Long:         `ELB（Elastic Load Balancing - ALB/NLB/GWLB）を操作するためのコマンド群です。`,
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// 親のPersistentPreRunEを実行（awsCtx設定とAWS設定読み込み）
		if err := RootCmd.PersistentPreRunE(cmd, args); err != nil {
			return err
		}

		// ELBv2用クライアント生成
		elbv2Client = elasticloadbalancingv2.NewFromConfig(awsCfg)

		return nil
	},
}

// elbLsCmd represents the ls command
var elbLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "ロードバランサー一覧を表示するコマンド",
	Long: `ロードバランサー（ALB/NLB/GWLB）の一覧を表示します。
削除保護の状態やターゲットグループ数などの情報も含めて表示します。

【使い方】
  ` + AppName + ` elb ls                    # 全ロードバランサー一覧を表示
  ` + AppName + ` elb ls --type alb         # ALBのみを表示
  ` + AppName + ` elb ls --type nlb         # NLBのみを表示
  ` + AppName + ` elb ls --type gwlb        # GWLBのみを表示
  ` + AppName + ` elb ls -p                 # 削除保護が有効なもののみを表示
  ` + AppName + ` elb ls --details          # 詳細情報付きで表示

【例】
  ` + AppName + ` elb ls --type nlb -p
  → 削除保護が有効なNLBのみを一覧表示します。`,
	RunE: func(cmdCobra *cobra.Command, args []string) error {
		protectedOnly, _ := cmdCobra.Flags().GetBool("protected-only")
		showDetails, _ := cmdCobra.Flags().GetBool("details")
		lbType, _ := cmdCobra.Flags().GetString("type")

		opts := elbsvc.ListOptions{
			ProtectedOnly:    protectedOnly,
			ShowDetails:      showDetails,
			LoadBalancerType: lbType,
		}

		return elbsvc.ListLoadBalancers(elbv2Client, opts)
	},
	SilenceUsage: true,
}

// elbCleanupCmd represents the cleanup command
var elbCleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "ロードバランサーを削除するコマンド",
	Long: `指定したキーワードを含むロードバランサー（ALB/NLB/GWLB）を削除します。
削除保護が有効な場合は自動的に保護を解除してから削除します。

例:
  ` + AppName + ` elb cleanup -f "test-" -P my-profile
  ` + AppName + ` elb cleanup -f "dev" --type alb
  ` + AppName + ` elb cleanup -f "stg" --with-target-groups`,
	RunE: func(cmd *cobra.Command, args []string) error {
		filter, _ := cmd.Flags().GetString("filter")
		withTargetGroups, _ := cmd.Flags().GetBool("with-target-groups")
		lbType, _ := cmd.Flags().GetString("type")

		printAwsContextWithInfo("検索文字列", filter)

		return elbsvc.CleanupLoadBalancersByFilter(elbv2Client, filter, withTargetGroups, lbType)
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(ElbCmd)
	ElbCmd.AddCommand(elbLsCmd)
	ElbCmd.AddCommand(elbCleanupCmd)

	// ls コマンドのフラグ
	elbLsCmd.Flags().BoolP("protected-only", "p", false, "削除保護が有効なもののみを表示")
	elbLsCmd.Flags().BoolP("details", "d", false, "詳細情報を表示")
	elbLsCmd.Flags().String("type", "", "ロードバランサータイプでフィルタ (alb, nlb, gwlb)")

	// cleanup コマンドのフラグ
	elbCleanupCmd.Flags().StringP("filter", "f", "", "削除対象のフィルターパターン")
	elbCleanupCmd.Flags().Bool("with-target-groups", false, "関連するターゲットグループも削除")
	elbCleanupCmd.Flags().String("type", "", "ロードバランサータイプでフィルタ (alb, nlb, gwlb)")
	_ = elbCleanupCmd.MarkFlagRequired("filter")
}
