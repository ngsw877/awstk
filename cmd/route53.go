package cmd

import (
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/spf13/cobra"

	route53Service "awstk/internal/service/route53"
)

var (
	route53Client *route53.Client
	useId         bool
)

// route53Cmd represents the route53 command
var route53Cmd = &cobra.Command{
	Use:   "route53",
	Short: "Route53ホストゾーン操作コマンド",
	Long:  `Route53のホストゾーンを管理するコマンドです。ホストゾーンの一覧表示や削除が可能です。`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// 親のPersistentPreRunEを実行（awsCtx設定とAWS設定読み込み）
		if err := RootCmd.PersistentPreRunE(cmd, args); err != nil {
			return err
		}

		// Route53用クライアント生成
		route53Client = route53.NewFromConfig(awsCfg)

		return nil
	},
}

// lsCmd represents the ls command
var route53LsCmd = &cobra.Command{
	Use:   "ls",
	Short: "ホストゾーン一覧を表示",
	Long:  `アカウント内のすべてのRoute53ホストゾーンを一覧表示します。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return route53Service.ListHostedZones(route53Client)
	},
}

// deleteCmd represents the delete command
var route53DeleteCmd = &cobra.Command{
	Use:   "delete <ドメイン名またはゾーンID>",
	Short: "ホストゾーンを削除",
	Long: `Route53のホストゾーンを削除します。デフォルトではドメイン名を指定します。
ホストゾーンIDを指定する場合は --id フラグを使用してください。

このコマンドは以下の処理を実行します：
1. すべてのリソースレコードセットを削除（NSとSOAレコードを除く）
2. ホストゾーン自体を削除

【使用例】
  ` + AppName + ` route53 delete example.com
  ` + AppName + ` route53 delete --id Z1234567890ABC`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		identifier := args[0]
		force, _ := cmd.Flags().GetBool("force")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		opts := route53Service.DeleteOptions{
			UseId:  useId,
			Force:  force,
			DryRun: dryRun,
		}

		return route53Service.DeleteHostedZone(route53Client, identifier, opts)
	},
}

func init() {
	RootCmd.AddCommand(route53Cmd)
	route53Cmd.AddCommand(route53LsCmd)
	route53Cmd.AddCommand(route53DeleteCmd)

	// delete command flags
	route53DeleteCmd.Flags().BoolVarP(&useId, "id", "i", false, "引数をホストゾーンIDとして扱う（デフォルト：ドメイン名）")
	route53DeleteCmd.Flags().BoolP("force", "f", false, "確認プロンプトをスキップ")
	route53DeleteCmd.Flags().BoolP("dry-run", "d", false, "削除対象を表示するのみ（実際には削除しない）")
}
