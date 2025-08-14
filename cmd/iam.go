package cmd

import (
	imPolicy "awstk/internal/service/iam/policy"
	imRole "awstk/internal/service/iam/role"

	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/spf13/cobra"
)

var (
	iamClient *awsiam.Client
	// role flags
	iamRoleUnusedDays int
	iamRoleExclude    []string
	// policy flags
	iamPolicyUnattached bool
	iamPolicyExclude    []string
)

// IamCmd represents the iam command
var IamCmd = &cobra.Command{
	Use:   "iam",
	Short: "IAMリソース操作コマンド",
	Long:  `IAMリソース（ユーザー/グループ/ロール/ポリシー）に関する操作コマンド群です。未使用のロール/ポリシーの一覧表示に対応しています。`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// 親のPersistentPreRunEを実行（awsCtx設定とAWS設定読み込み）
		if err := RootCmd.PersistentPreRunE(cmd, args); err != nil {
			return err
		}

		// IAMクライアントを初期化
		iamClient = awsiam.NewFromConfig(awsCfg)
		return nil
	},
}

// iam role ...
var IamRoleCmd = &cobra.Command{
	Use:   "role",
	Short: "IAMロール操作",
}

var iamRoleLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "IAMロール一覧を表示",
	Long: `IAMロールの一覧を表示します。

例:
  ` + AppName + ` iam role ls                 # 全ロール（最終使用日時つき）
  ` + AppName + ` iam role ls -u 180          # 180日以上未使用のロールのみ
  ` + AppName + ` iam role ls -x AWSServiceRoleFor -x AWSReservedSSO`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return imRole.List(iamClient, imRole.ListOptions{
			UnusedDays: iamRoleUnusedDays,
			Exclude:    iamRoleExclude,
		})
	},
	SilenceUsage: true,
}

// iam policy ...
var IamPolicyCmd = &cobra.Command{
	Use:   "policy",
	Short: "IAMポリシー操作",
}

var iamPolicyLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "カスタマー管理ポリシー一覧を表示",
	Long: `カスタマー管理ポリシーの一覧を表示します。

例:
  ` + AppName + ` iam policy ls               # 全カスタマー管理ポリシー
  ` + AppName + ` iam policy ls --unattached  # 未アタッチのみ
  ` + AppName + ` iam policy ls -x AWSReserved`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return imPolicy.List(iamClient, imPolicy.ListOptions{
			UnattachedOnly: iamPolicyUnattached,
			Exclude:        iamPolicyExclude,
		})
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(IamCmd)
	IamCmd.AddCommand(IamRoleCmd)
	IamCmd.AddCommand(IamPolicyCmd)
	IamRoleCmd.AddCommand(iamRoleLsCmd)
	IamPolicyCmd.AddCommand(iamPolicyLsCmd)

	// iam role ls flags
	iamRoleLsCmd.Flags().IntVarP(&iamRoleUnusedDays, "unused-days", "u", 0, "未使用とみなす経過日数（指定時は未使用のみ、0で全件）")
	iamRoleLsCmd.Flags().StringSliceVarP(&iamRoleExclude, "exclude", "x", []string{}, "除外パターン（名前に含む文字列、複数指定可）")

	// iam policy ls flags
	iamPolicyLsCmd.Flags().BoolVarP(&iamPolicyUnattached, "unattached", "u", false, "未アタッチのポリシーのみ表示")
	iamPolicyLsCmd.Flags().StringSliceVarP(&iamPolicyExclude, "exclude", "x", []string{}, "除外パターン（名前に含む文字列、複数指定可）")
}
