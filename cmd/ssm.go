package cmd

import (
	"awstk/internal/aws"
	ssmsvc "awstk/internal/service/ssm"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/spf13/cobra"
)

var ssmInstanceId string
var ssmParamsPrefix string
var ssmParamsDryRun bool
var ssmDeleteForce bool
var ssmClient *ssm.Client

var ssmCmd = &cobra.Command{
	Use:   "ssm",
	Short: "SSM関連の操作を行うコマンド群",
	Long:  "AWS SSMセッションマネージャーを利用したEC2インスタンスへの接続やParameter Storeの操作を行うCLIコマンド群です。",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// 親のPersistentPreRunEを実行（awsCtx設定とAWS設定読み込み）
		if err := RootCmd.PersistentPreRunE(cmd, args); err != nil {
			return err
		}

		// クライアント生成
		ssmClient = ssm.NewFromConfig(awsCfg)

		return nil
	},
}

var ssmSessionStartCmd = &cobra.Command{
	Use:   "session",
	Short: "EC2インスタンスにSSMで接続する",
	Long: `指定したEC2インスタンスIDにSSMセッションで接続します。

例:
  ` + AppName + ` ssm session -i <ec2-instance-id> [-P <aws-profile>]
  ` + AppName + ` ssm session [-P <aws-profile>]  # インスタンス一覧から選択
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		awsCtx := aws.Context{Region: region, Profile: profile}
		ec2Client := ec2.NewFromConfig(awsCfg)

		return ssmsvc.SelectAndStartSession(awsCtx, ec2Client, ssmInstanceId)
	},
	SilenceUsage: true,
}

var ssmPutParamsCmd = &cobra.Command{
	Use:   "put-params <file>",
	Short: "ファイルからParameter Storeに一括登録",
	Long: `CSV/JSONファイルからAWS Systems Manager Parameter Storeにパラメータを一括登録します。

対応ファイル形式:
  - CSV (.csv): name,value,type,description の形式
  - JSON (.json): {"parameters": [{"name": "...", "value": "...", "type": "...", "description": "..."}]}

例:
  ` + AppName + ` ssm put-params params.csv
  ` + AppName + ` ssm put-params params.json --prefix /myapp/
  ` + AppName + ` ssm put-params params.csv --dry-run
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		// ファイル拡張子のバリデーション
		if !strings.HasSuffix(filePath, ".csv") && !strings.HasSuffix(filePath, ".json") {
			return fmt.Errorf("❌ サポートされていないファイル形式です。.csv または .json ファイルを指定してください")
		}

		opts := ssmsvc.PutParamsOptions{
			FilePath: filePath,
			Prefix:   ssmParamsPrefix,
			DryRun:   ssmParamsDryRun,
		}

		err := ssmsvc.PutParametersFromFile(ssmClient, opts)
		if err != nil {
			return fmt.Errorf("❌ パラメータの登録に失敗しました: %w", err)
		}

		if ssmParamsDryRun {
			fmt.Println("✅ ドライラン完了")
		} else {
			fmt.Println("✅ パラメータの登録が完了しました")
		}
		return nil
	},
	SilenceUsage: true,
}

var ssmDeleteParamsCmd = &cobra.Command{
	Use:   "delete-params <file>",
	Short: "ファイルからParameter Storeを一括削除",
	Long: `テキストファイルに記載されたパラメータ名のリストから、AWS Systems Manager Parameter Storeのパラメータを一括削除します。

ファイル形式:
  - 1行に1つのパラメータ名を記載
  - 空行と#で始まるコメント行は無視されます

例:
  ` + AppName + ` ssm delete-params params.txt
  ` + AppName + ` ssm delete-params params.txt --force
  ` + AppName + ` ssm delete-params params.txt --dry-run
  ` + AppName + ` ssm delete-params params.txt --prefix /myapp/  # 削除対象パラメータ名に/myapp/を付加
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		opts := ssmsvc.DeleteParamsOptions{
			FilePath: filePath,
			Prefix:   ssmParamsPrefix,
			DryRun:   ssmParamsDryRun,
			Force:    ssmDeleteForce,
		}

		err := ssmsvc.DeleteParametersFromFile(ssmClient, opts)
		if err != nil {
			return fmt.Errorf("❌ パラメータの削除に失敗しました: %w", err)
		}

		if ssmParamsDryRun {
			fmt.Println("✅ ドライラン完了")
		} else {
			fmt.Println("✅ パラメータの削除が完了しました")
		}
		return nil
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(ssmCmd)
	ssmCmd.AddCommand(ssmSessionStartCmd)
	ssmCmd.AddCommand(ssmPutParamsCmd)
	ssmCmd.AddCommand(ssmDeleteParamsCmd)

	// session サブコマンドのフラグ
	ssmSessionStartCmd.Flags().StringVarP(&ssmInstanceId, "instance-id", "i", "", "EC2インスタンスID（省略時は一覧から選択）")

	// put-params サブコマンドのフラグ
	ssmPutParamsCmd.Flags().StringVarP(&ssmParamsPrefix, "prefix", "p", "", "パラメータ名のプレフィックス")
	ssmPutParamsCmd.Flags().BoolVarP(&ssmParamsDryRun, "dry-run", "d", false, "実際には登録せず、登録内容を確認")

	// delete-params サブコマンドのフラグ
	ssmDeleteParamsCmd.Flags().StringVarP(&ssmParamsPrefix, "prefix", "p", "", "パラメータ名のプレフィックス")
	ssmDeleteParamsCmd.Flags().BoolVarP(&ssmParamsDryRun, "dry-run", "d", false, "実際には削除せず、削除対象を確認")
	ssmDeleteParamsCmd.Flags().BoolVarP(&ssmDeleteForce, "force", "f", false, "確認プロンプトをスキップ")
}
