package cmd

import (
	imPolicy "awstk/internal/service/iam/policy"
	imRole "awstk/internal/service/iam/role"

	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/spf13/cobra"
)

var (
	iamClient *awsiam.Client
	// role ls flags
	iamRoleUnusedDays int
	iamRoleExclude    []string
	// role delete flags
	iamRoleDeleteSearch     string
	iamRoleDeleteUnusedDays int
	iamRoleDeleteExclude    []string
	iamRoleDeleteExact      bool
	// policy ls flags
	iamPolicyUnattached bool
	iamPolicyExclude    []string
	// policy delete flags
	iamPolicyDeleteSearch     string
	iamPolicyDeleteUnattached bool
	iamPolicyDeleteExclude    []string
	iamPolicyDeleteExact      bool
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
  ` + AppName + ` iam role ls -u               # 一度も使用されていないロールのみ
  ` + AppName + ` iam role ls -u 180          # 180日以上未使用のロールのみ
  ` + AppName + ` iam role ls -x AWSServiceRoleFor -x AWSReservedSSO`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return imRole.ListIamRoles(iamClient, imRole.ListOptions{
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
		return imPolicy.ListIamPolicies(iamClient, imPolicy.ListOptions{
			UnattachedOnly: iamPolicyUnattached,
			Exclude:        iamPolicyExclude,
		})
	},
	SilenceUsage: true,
}

var iamRoleDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "IAMロールを削除",
	Long: `検索パターンに一致するIAMロールを削除します。

例:
  ` + AppName + ` iam role delete -s "test-*"              # パターンマッチで削除
  ` + AppName + ` iam role delete -s "test" -u 180         # 180日未使用 AND "test"含む
  ` + AppName + ` iam role delete -s "test" -u             # 一度も未使用 AND "test"含む
  ` + AppName + ` iam role delete -s "test" -x AWSReserved # 除外パターン指定`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return imRole.DeleteRoles(iamClient, imRole.DeleteOptions{
			Filter:     iamRoleDeleteSearch,
			UnusedDays: iamRoleDeleteUnusedDays,
			Exclude:    iamRoleDeleteExclude,
			Exact:      iamRoleDeleteExact,
		})
	},
	SilenceUsage: true,
}

var iamPolicyDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "IAMポリシーを削除",
	Long: `検索パターンに一致するカスタマー管理ポリシーを削除します。

例:
  ` + AppName + ` iam policy delete -s "test-*"              # パターンマッチで削除
  ` + AppName + ` iam policy delete -s "test" --unattached   # 未アタッチ AND "test"含む
  ` + AppName + ` iam policy delete -s "test" -x AWSReserved # 除外パターン指定`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return imPolicy.DeletePolicies(iamClient, imPolicy.DeleteOptions{
			Filter:         iamPolicyDeleteSearch,
			UnattachedOnly: iamPolicyDeleteUnattached,
			Exclude:        iamPolicyDeleteExclude,
			Exact:          iamPolicyDeleteExact,
		})
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(IamCmd)
	IamCmd.AddCommand(IamRoleCmd)
	IamCmd.AddCommand(IamPolicyCmd)
	IamRoleCmd.AddCommand(iamRoleLsCmd)
	IamRoleCmd.AddCommand(iamRoleDeleteCmd)
	IamPolicyCmd.AddCommand(iamPolicyLsCmd)
	IamPolicyCmd.AddCommand(iamPolicyDeleteCmd)

	// iam role ls flags
	iamRoleLsCmd.Flags().IntVarP(&iamRoleUnusedDays, "unused-days", "u", 0, "未使用とみなす経過日数（引数なし=一度も使用なし、数値指定=指定日数以上未使用、0=全件）")
	if unusedDaysFlag := iamRoleLsCmd.Flags().Lookup("unused-days"); unusedDaysFlag != nil {
		unusedDaysFlag.NoOptDefVal = "-1"
	}
	iamRoleLsCmd.Flags().StringSliceVarP(&iamRoleExclude, "exclude", "x", []string{}, "除外パターン（名前に含む文字列、複数指定可）")

	// iam role delete flags
	iamRoleDeleteCmd.Flags().StringVarP(&iamRoleDeleteSearch, "search", "s", "", "削除対象の検索パターン（必須）")
	_ = iamRoleDeleteCmd.MarkFlagRequired("search")
	iamRoleDeleteCmd.Flags().IntVarP(&iamRoleDeleteUnusedDays, "unused-days", "u", 0, "未使用とみなす経過日数（引数なし=一度も使用なし、数値指定=指定日数以上未使用、0=全件）")
	if unusedDaysFlag := iamRoleDeleteCmd.Flags().Lookup("unused-days"); unusedDaysFlag != nil {
		unusedDaysFlag.NoOptDefVal = "-1"
	}
	iamRoleDeleteCmd.Flags().StringSliceVarP(&iamRoleDeleteExclude, "exclude", "x", []string{}, "除外パターン（名前に含む文字列、複数指定可）")
	iamRoleDeleteCmd.Flags().BoolVar(&iamRoleDeleteExact, "exact", false, "大文字小文字を区別してマッチ")

	// iam policy ls flags
	iamPolicyLsCmd.Flags().BoolVarP(&iamPolicyUnattached, "unattached", "u", false, "未アタッチのポリシーのみ表示")
	iamPolicyLsCmd.Flags().StringSliceVarP(&iamPolicyExclude, "exclude", "x", []string{}, "除外パターン（名前に含む文字列、複数指定可）")

	// iam policy delete flags
	iamPolicyDeleteCmd.Flags().StringVarP(&iamPolicyDeleteSearch, "search", "s", "", "削除対象の検索パターン（必須）")
	_ = iamPolicyDeleteCmd.MarkFlagRequired("search")
	iamPolicyDeleteCmd.Flags().BoolVarP(&iamPolicyDeleteUnattached, "unattached", "u", false, "未アタッチのポリシーのみ削除")
	iamPolicyDeleteCmd.Flags().StringSliceVarP(&iamPolicyDeleteExclude, "exclude", "x", []string{}, "除外パターン（名前に含む文字列、複数指定可）")
	iamPolicyDeleteCmd.Flags().BoolVar(&iamPolicyDeleteExact, "exact", false, "大文字小文字を区別してマッチ")
}
