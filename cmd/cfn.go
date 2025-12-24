package cmd

import (
	"awstk/internal/aws"
	"awstk/internal/service/cfn"
	"awstk/internal/service/common"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/spf13/cobra"
)

// CfnCmd represents the cfn command
var CfnCmd = &cobra.Command{
	Use:   "cfn",
	Short: "CloudFormationリソース操作コマンド",
	Long:  `CloudFormationリソースを操作するためのコマンド群です。`,
}

var showAll bool

var cfnLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "CloudFormationスタック一覧を表示するコマンド",
	Long:  `CloudFormationスタック一覧を表示します。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfnClient := cloudformation.NewFromConfig(awsCfg)

		stacks, err := cfn.ListCfnStacks(cfnClient, showAll)
		if err != nil {
			return common.FormatListError("CloudFormationスタック", err)
		}

		if len(stacks) == 0 {
			fmt.Println(common.FormatEmptyMessage("CloudFormationスタック"))
			return nil
		}

		// ステータス付きリストとして表示
		items := make([]common.ListItem, len(stacks))
		for i, stk := range stacks {
			items[i] = common.ListItem{
				Name:   stk.Name,
				Status: stk.Status,
			}
		}
		common.PrintStatusList("CloudFormationスタック一覧", items, "スタック")

		return nil
	},
	SilenceUsage: true,
}

var cfnStartCmd = &cobra.Command{
	Use:   "start",
	Short: "CloudFormationスタック内のリソースを一括起動するコマンド",
	Long: `CloudFormationスタック内の起動・停止可能なリソースを一括起動します。
対象リソース: EC2インスタンス、RDSインスタンス、Aurora DBクラスター、ECSサービス

例:
  ` + AppName + ` cfn start -S my-stack -P my-profile`,
	RunE: func(cmd *cobra.Command, args []string) error {
		resolveStackName()
		if stackName == "" {
			return fmt.Errorf("❌ エラー: スタック名 (-S) を指定してください")
		}

		printAwsContextWithInfo("Stack", stackName)

		cfnClient := cloudformation.NewFromConfig(awsCfg)
		ec2Client := ec2.NewFromConfig(awsCfg)
		rdsClient := rds.NewFromConfig(awsCfg)
		aasClient := applicationautoscaling.NewFromConfig(awsCfg)

		err := cfn.StartAllStackResources(cfnClient, ec2Client, rdsClient, aasClient, stackName)
		if err != nil {
			return fmt.Errorf("❌ リソース起動処理でエラー: %w", err)
		}

		fmt.Println("✅ リソース起動処理が完了しました")
		return nil
	},
	SilenceUsage: true,
}

var cfnStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "CloudFormationスタック内のリソースを一括停止するコマンド",
	Long: `CloudFormationスタック内の起動・停止可能なリソースを一括停止します。
対象リソース: EC2インスタンス、RDSインスタンス、Aurora DBクラスター、ECSサービス

例:
  ` + AppName + ` cfn stop -S my-stack -P my-profile`,
	RunE: func(cmd *cobra.Command, args []string) error {
		resolveStackName()
		if stackName == "" {
			return fmt.Errorf("❌ エラー: スタック名 (-S) を指定してください")
		}

		printAwsContextWithInfo("Stack", stackName)

		cfnClient := cloudformation.NewFromConfig(awsCfg)
		ec2Client := ec2.NewFromConfig(awsCfg)
		rdsClient := rds.NewFromConfig(awsCfg)
		aasClient := applicationautoscaling.NewFromConfig(awsCfg)

		err := cfn.StopAllStackResources(cfnClient, ec2Client, rdsClient, aasClient, stackName)
		if err != nil {
			return fmt.Errorf("❌ リソース停止処理でエラー: %w", err)
		}

		fmt.Println("✅ リソース停止処理が完了しました")
		return nil
	},
	SilenceUsage: true,
}

var (
	cleanupFilter string
	cleanupStatus string
	cleanupForce  bool
	cleanupExact  bool
)

var cfnCleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "CloudFormationスタックを一括削除するコマンド",
	Long: `指定した条件に一致するCloudFormationスタックを一括削除します。
フィルターによる名前の部分一致検索、またはステータスによる絞り込みが可能です。

例:
  # 名前に "test-" を含むスタックを削除
  ` + AppName + ` cfn cleanup --filter test-

  # 削除失敗状態のスタックをクリーンアップ
  ` + AppName + ` cfn cleanup --status DELETE_FAILED,ROLLBACK_COMPLETE

  # 両方の条件を組み合わせ
  ` + AppName + ` cfn cleanup --filter dev- --status CREATE_FAILED

  # 確認プロンプトをスキップ
  ` + AppName + ` cfn cleanup --filter test- --force`,
	RunE: func(cmd *cobra.Command, args []string) error {
		printAwsContext()

		cfnClient := cloudformation.NewFromConfig(awsCfg)

		err := cfn.CleanupStacks(cfnClient, cfn.CleanupOptions{
			Filter: cleanupFilter,
			Status: cleanupStatus,
			Force:  cleanupForce,
			Exact:  cleanupExact,
		})
		if err != nil {
			return fmt.Errorf("❌ スタック削除処理でエラー: %w", err)
		}

		return nil
	},
	SilenceUsage: true,
}

var cfnProtectCmd = &cobra.Command{
	Use:   "protect",
	Short: "CloudFormationスタックの削除保護を一括設定するコマンド",
	Long: `指定した条件に一致するCloudFormationスタックの削除保護を一括で有効化または無効化します。
フィルターによる名前の部分一致検索、またはステータスによる絞り込みが可能です。

例:
  # 名前に "prod-" を含むスタックの削除保護を有効化
  ` + AppName + ` cfn protect --filter prod- --enable

  # 特定ステータスのスタックの削除保護を無効化
  ` + AppName + ` cfn protect --status CREATE_COMPLETE --disable

  # 両方の条件を組み合わせ
  ` + AppName + ` cfn protect --filter dev- --status UPDATE_COMPLETE --enable

  # 特定のスタックを指定
  ` + AppName + ` cfn protect stack-a stack-b --enable`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// フラグの値を取得
		protectFilter, _ := cmd.Flags().GetString("filter")
		protectStatus, _ := cmd.Flags().GetString("status")
		protectEnable, _ := cmd.Flags().GetBool("enable")
		protectExact, _ := cmd.Flags().GetBool("exact")

		// フィルター条件の排他チェック
		if err := ValidateStackSelection(args, protectFilter != "" || protectStatus != ""); err != nil {
			return err
		}

		printAwsContext()

		cfnClient := cloudformation.NewFromConfig(awsCfg)

		err := cfn.UpdateProtection(cfnClient, cfn.ProtectOptions{
			Stacks: args,
			Filter: protectFilter,
			Status: protectStatus,
			Enable: protectEnable, // --enableならtrue、--disableならfalse
			Exact:  protectExact,
		})
		if err != nil {
			return fmt.Errorf("❌ 削除保護の更新処理でエラー: %w", err)
		}

		return nil
	},
	SilenceUsage: true,
}

var cfnDriftDetectCmd = &cobra.Command{
	Use:   "drift-detect",
	Short: "CloudFormationスタックのドリフト検出を一括実行するコマンド",
	Long: `指定した条件に一致するCloudFormationスタックのドリフト検出を一括で実行します。
フィルターによる名前の部分一致検索、または全スタックを対象にできます。

例:
  # 名前に "prod-" を含むスタックのドリフト検出
  ` + AppName + ` cfn drift-detect --filter prod-

  # すべてのスタックのドリフト検出
  ` + AppName + ` cfn drift-detect --all

  # 特定のスタックを指定
  ` + AppName + ` cfn drift-detect stack-a stack-b stack-c

  # 実行例
  ` + AppName + ` cfn drift-detect --filter test-`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// フラグの値を取得
		driftFilter, _ := cmd.Flags().GetString("filter")
		driftAll, _ := cmd.Flags().GetBool("all")
		driftExact, _ := cmd.Flags().GetBool("exact")

		// 排他チェック
		if err := ValidateStackSelection(args, driftFilter != "" || driftAll); err != nil {
			return err
		}

		printAwsContext()

		cfnClient := cloudformation.NewFromConfig(awsCfg)

		err := cfn.DetectDrift(cfnClient, cfn.DriftOptions{
			Stacks: args,
			Filter: driftFilter,
			All:    driftAll,
			Exact:  driftExact,
		})
		if err != nil {
			return fmt.Errorf("❌ ドリフト検出処理でエラー: %w", err)
		}

		return nil
	},
	SilenceUsage: true,
}

var cfnDriftStatusCmd = &cobra.Command{
	Use:   "drift-status",
	Short: "CloudFormationスタックのドリフト状態を一括確認するコマンド",
	Long: `指定した条件に一致するCloudFormationスタックのドリフト状態を一括で確認します。
フィルターによる名前の部分一致検索、または全スタックを対象にできます。

例:
  # 名前に "prod-" を含むスタックのドリフト状態確認
  ` + AppName + ` cfn drift-status --filter prod-

  # すべてのスタックのドリフト状態確認
  ` + AppName + ` cfn drift-status --all

  # 特定のスタックを指定
  ` + AppName + ` cfn drift-status stack-a stack-b

  # ドリフトしているスタックのみ表示
  ` + AppName + ` cfn drift-status --filter prod- --drifted-only`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// フラグの値を取得
		driftFilter, _ := cmd.Flags().GetString("filter")
		driftAll, _ := cmd.Flags().GetBool("all")
		driftedOnly, _ := cmd.Flags().GetBool("drifted-only")
		driftStatusExact, _ := cmd.Flags().GetBool("exact")

		// 排他チェック
		if err := ValidateStackSelection(args, driftFilter != "" || driftAll); err != nil {
			return err
		}

		printAwsContext()

		cfnClient := cloudformation.NewFromConfig(awsCfg)

		err := cfn.ShowDriftStatus(cfnClient, cfn.DriftStatusOptions{
			Stacks:      args,
			Filter:      driftFilter,
			All:         driftAll,
			DriftedOnly: driftedOnly,
			Exact:       driftStatusExact,
		})
		if err != nil {
			return fmt.Errorf("❌ ドリフト状態確認処理でエラー: %w", err)
		}

		return nil
	},
	SilenceUsage: true,
}

var (
	deployTemplatePath    string
	deployStackName       string
	deployParameters      string
	deployIsChangeSetOnly bool
)

var cfnDeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "CloudFormationスタックをデプロイするコマンド",
	Long: `指定したテンプレートファイルからCloudFormationスタックをデプロイします。
内部的に aws cloudformation deploy コマンドを実行します。

例:
  # 基本的なデプロイ
  ` + AppName + ` cfn deploy -t template.yaml -S my-stack

  # パラメータを指定してデプロイ（key=value形式）
  ` + AppName + ` cfn deploy -t template.yaml -S my-stack -p KeyName=mykey,InstanceType=t3.micro

  # パラメータをJSONファイルから読み込み
  ` + AppName + ` cfn deploy -t template.yaml -S my-stack -p params.json

  # Change Setの作成のみ（実行は手動）
  ` + AppName + ` cfn deploy -t template.yaml -S my-stack -n`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if deployTemplatePath == "" {
			return fmt.Errorf("❌ エラー: テンプレートファイルパス (--template) を指定してください")
		}
		if deployStackName == "" {
			return fmt.Errorf("❌ エラー: スタック名 (--stack) を指定してください")
		}

		printAwsContext()

		awsCtx := aws.Context{Region: region, Profile: profile}

		// パラメータの処理
		var params map[string]string
		var paramFile string

		if deployParameters != "" {
			// .jsonで終わる場合はファイルパスとして扱う
			if strings.HasSuffix(strings.ToLower(deployParameters), ".json") {
				paramFile = deployParameters
			} else {
				// key=value形式をパース
				params = make(map[string]string)
				pairs := strings.Split(deployParameters, ",")
				for _, pair := range pairs {
					kv := strings.SplitN(pair, "=", 2)
					if len(kv) == 2 {
						params[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
					}
				}
			}
		}

		err := cfn.DeployStack(awsCtx, cfn.DeployOptions{
			TemplatePath:    deployTemplatePath,
			StackName:       deployStackName,
			Parameters:      params,
			ParameterFile:   paramFile,
			IsChangeSetOnly: deployIsChangeSetOnly,
		})
		if err != nil {
			return fmt.Errorf("❌ デプロイ処理でエラー: %w", err)
		}

		return nil
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(CfnCmd)
	CfnCmd.AddCommand(cfnLsCmd)
	CfnCmd.AddCommand(cfnDeployCmd)
	CfnCmd.AddCommand(cfnStartCmd)
	CfnCmd.AddCommand(cfnStopCmd)
	CfnCmd.AddCommand(cfnCleanupCmd)
	CfnCmd.AddCommand(cfnProtectCmd)
	CfnCmd.AddCommand(cfnDriftDetectCmd)
	CfnCmd.AddCommand(cfnDriftStatusCmd)

	cfnLsCmd.Flags().BoolVarP(&showAll, "all", "a", false, "全てのステータスのスタックを表示")

	// cfn deployコマンド用のフラグ
	cfnDeployCmd.Flags().StringVarP(&deployTemplatePath, "template", "t", "", "テンプレートファイルのパス")
	cfnDeployCmd.Flags().StringVarP(&deployStackName, "stack-name", "S", "", "スタック名")
	cfnDeployCmd.Flags().StringVarP(&deployParameters, "parameters", "p", "", "パラメータ（key=value形式またはJSONファイルパス）")
	cfnDeployCmd.Flags().BoolVarP(&deployIsChangeSetOnly, "no-execute", "n", false, "Change Setの作成のみで実行しない")
	_ = cfnDeployCmd.MarkFlagRequired("template")
	_ = cfnDeployCmd.MarkFlagRequired("stack-name")

	// cfn start/stopコマンド用のフラグ
	cfnStartCmd.Flags().StringVarP(&stackName, "stack-name", "S", "", "CloudFormationスタック名")
	cfnStopCmd.Flags().StringVarP(&stackName, "stack-name", "S", "", "CloudFormationスタック名")

	// cfn cleanupコマンド用のフラグ
	cfnCleanupCmd.Flags().StringVar(&cleanupFilter, "filter", "", "スタック名のフィルター（部分一致）")
	cfnCleanupCmd.Flags().StringVar(&cleanupStatus, "status", "", "削除対象のステータス（カンマ区切り）")
	cfnCleanupCmd.Flags().BoolVarP(&cleanupForce, "force", "f", false, "確認プロンプトをスキップ")
	cfnCleanupCmd.Flags().BoolVar(&cleanupExact, "exact", false, "大文字小文字を区別してマッチ")
	// どちらか1つ必須
	cfnCleanupCmd.MarkFlagsOneRequired("filter", "status")

	// cfn protectコマンド用のフラグ
	cfnProtectCmd.Flags().StringP("filter", "F", "", "スタック名のフィルター（部分一致）")
	cfnProtectCmd.Flags().StringP("status", "s", "", "対象のステータス（カンマ区切り）")
	cfnProtectCmd.Flags().BoolP("enable", "e", false, "削除保護を有効化")
	cfnProtectCmd.Flags().BoolP("disable", "d", false, "削除保護を無効化")
	cfnProtectCmd.Flags().Bool("exact", false, "大文字小文字を区別してマッチ")
	// enable/disable は相互排他かつどちらか1つ必須
	cfnProtectCmd.MarkFlagsMutuallyExclusive("enable", "disable")
	cfnProtectCmd.MarkFlagsOneRequired("enable", "disable")

	// cfn drift-detectコマンド用のフラグ
	cfnDriftDetectCmd.Flags().StringP("filter", "F", "", "スタック名のフィルター（部分一致）")
	cfnDriftDetectCmd.Flags().BoolP("all", "a", false, "すべてのスタックを対象")
	cfnDriftDetectCmd.Flags().Bool("exact", false, "大文字小文字を区別してマッチ")

	// cfn drift-statusコマンド用のフラグ
	cfnDriftStatusCmd.Flags().StringP("filter", "F", "", "スタック名のフィルター（部分一致）")
	cfnDriftStatusCmd.Flags().BoolP("all", "a", false, "すべてのスタックを対象")
	cfnDriftStatusCmd.Flags().BoolP("drifted-only", "d", false, "ドリフトしているスタックのみ表示")
	cfnDriftStatusCmd.Flags().Bool("exact", false, "大文字小文字を区別してマッチ")
}
