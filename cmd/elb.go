package cmd

import (
	elbsvc "awstk/internal/service/elb"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/spf13/cobra"
)

var (
	elbv2Client    *elasticloadbalancingv2.Client
	elbDeleteExact bool
	elbDeleteForce bool
)

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

// elbDeleteCmd represents the delete command
var elbDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "ロードバランサーを削除するコマンド",
	Long: `指定したキーワードを含むロードバランサー（ALB/NLB/GWLB）を削除します。
削除保護が有効な場合は --force オプションで保護を解除して削除できます。

例:
  ` + AppName + ` elb delete -s "test-" -P my-profile
  ` + AppName + ` elb delete -s "dev" --type alb
  ` + AppName + ` elb delete -s "stg" --with-target-groups
  ` + AppName + ` elb delete -s "prod" --force    # 削除保護を解除して削除`,
	RunE: func(cmd *cobra.Command, args []string) error {
		search, _ := cmd.Flags().GetString("search")
		withTargetGroups, _ := cmd.Flags().GetBool("with-target-groups")
		lbType, _ := cmd.Flags().GetString("type")

		printAwsContextWithInfo("検索文字列", search)

		return elbsvc.DeleteLoadBalancersByFilter(elbv2Client, search, withTargetGroups, lbType, elbDeleteExact, elbDeleteForce)
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(ElbCmd)
	ElbCmd.AddCommand(elbLsCmd)
	ElbCmd.AddCommand(elbDeleteCmd)

	// ls コマンドのフラグ
	elbLsCmd.Flags().BoolP("protected-only", "p", false, "削除保護が有効なもののみを表示")
	elbLsCmd.Flags().BoolP("details", "d", false, "詳細情報を表示")
	elbLsCmd.Flags().String("type", "", "ロードバランサータイプでフィルタ (alb, nlb, gwlb)")

	// delete コマンドのフラグ
	elbDeleteCmd.Flags().StringP("search", "s", "", "削除対象の検索パターン")
	elbDeleteCmd.Flags().Bool("with-target-groups", false, "関連するターゲットグループも削除")
	elbDeleteCmd.Flags().String("type", "", "ロードバランサータイプでフィルタ (alb, nlb, gwlb)")
	elbDeleteCmd.Flags().BoolVar(&elbDeleteExact, "exact", false, "大文字小文字を区別してマッチ")
	elbDeleteCmd.Flags().BoolVar(&elbDeleteForce, "force", false, "削除保護を解除して削除")
	_ = elbDeleteCmd.MarkFlagRequired("search")
}
