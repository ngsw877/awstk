package cmd

import (
	"awstk/internal/service/cfn"
	"awstk/internal/service/common"
	"fmt"

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
		if cleanupFilter == "" && cleanupStatus == "" {
			return fmt.Errorf("❌ エラー: --filterまたは--statusのいずれかを指定してください")
		}

		printAwsContext()

		cfnClient := cloudformation.NewFromConfig(awsCfg)

		err := cfn.CleanupStacks(cfnClient, cfn.CleanupOptions{
			Filter: cleanupFilter,
			Status: cleanupStatus,
			Force:  cleanupForce,
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

  # 確認プロンプトをスキップ
  ` + AppName + ` cfn protect --filter test- --disable --force`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// フラグの値を取得
		protectFilter, _ := cmd.Flags().GetString("filter")
		protectStatus, _ := cmd.Flags().GetString("status")
		protectEnable, _ := cmd.Flags().GetBool("enable")
		protectDisable, _ := cmd.Flags().GetBool("disable")
		protectForce, _ := cmd.Flags().GetBool("force")
		// --enableと--disableの排他チェック
		if protectEnable && protectDisable {
			return fmt.Errorf("❌ エラー: --enableと--disableは同時に指定できません")
		}
		if !protectEnable && !protectDisable {
			return fmt.Errorf("❌ エラー: --enableまたは--disableのいずれかを指定してください")
		}

		// フィルター条件のチェック
		if protectFilter == "" && protectStatus == "" {
			return fmt.Errorf("❌ エラー: --filterまたは--statusのいずれかを指定してください")
		}

		printAwsContext()

		cfnClient := cloudformation.NewFromConfig(awsCfg)

		err := cfn.UpdateProtection(cfnClient, cfn.ProtectOptions{
			Filter: protectFilter,
			Status: protectStatus,
			Enable: protectEnable, // --enableならtrue、--disableならfalse
			Force:  protectForce,
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

		// 排他チェック
		specified := 0
		if len(args) > 0 {
			specified++
		}
		if driftFilter != "" {
			specified++
		}
		if driftAll {
			specified++
		}

		if specified == 0 {
			return fmt.Errorf("❌ エラー: スタック名、--filter、--allのいずれかを指定してください")
		}
		if specified > 1 {
			return fmt.Errorf("❌ エラー: スタック名、--filter、--allは同時に指定できません")
		}

		printAwsContext()

		cfnClient := cloudformation.NewFromConfig(awsCfg)

		err := cfn.DetectDrift(cfnClient, cfn.DriftOptions{
			Stacks: args,
			Filter: driftFilter,
			All:    driftAll,
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

		// 排他チェック
		specified := 0
		if len(args) > 0 {
			specified++
		}
		if driftFilter != "" {
			specified++
		}
		if driftAll {
			specified++
		}

		if specified == 0 {
			return fmt.Errorf("❌ エラー: スタック名、--filter、--allのいずれかを指定してください")
		}
		if specified > 1 {
			return fmt.Errorf("❌ エラー: スタック名、--filter、--allは同時に指定できません")
		}

		printAwsContext()

		cfnClient := cloudformation.NewFromConfig(awsCfg)

		err := cfn.ShowDriftStatus(cfnClient, cfn.DriftStatusOptions{
			Stacks:      args,
			Filter:      driftFilter,
			All:         driftAll,
			DriftedOnly: driftedOnly,
		})
		if err != nil {
			return fmt.Errorf("❌ ドリフト状態確認処理でエラー: %w", err)
		}

		return nil
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(CfnCmd)
	CfnCmd.AddCommand(cfnLsCmd)
	CfnCmd.AddCommand(cfnStartCmd)
	CfnCmd.AddCommand(cfnStopCmd)
	CfnCmd.AddCommand(cfnCleanupCmd)
	CfnCmd.AddCommand(cfnProtectCmd)
	CfnCmd.AddCommand(cfnDriftDetectCmd)
	CfnCmd.AddCommand(cfnDriftStatusCmd)

	cfnLsCmd.Flags().BoolVarP(&showAll, "all", "a", false, "全てのステータスのスタックを表示")

	// cfn start/stopコマンド用のフラグ
	cfnStartCmd.Flags().StringVarP(&stackName, "stack", "S", "", "CloudFormationスタック名")
	cfnStopCmd.Flags().StringVarP(&stackName, "stack", "S", "", "CloudFormationスタック名")

	// cfn cleanupコマンド用のフラグ
	cfnCleanupCmd.Flags().StringVar(&cleanupFilter, "filter", "", "スタック名のフィルター（部分一致）")
	cfnCleanupCmd.Flags().StringVar(&cleanupStatus, "status", "", "削除対象のステータス（カンマ区切り）")
	cfnCleanupCmd.Flags().BoolVarP(&cleanupForce, "force", "f", false, "確認プロンプトをスキップ")

	// cfn protectコマンド用のフラグ
	cfnProtectCmd.Flags().String("filter", "", "スタック名のフィルター（部分一致）")
	cfnProtectCmd.Flags().String("status", "", "対象のステータス（カンマ区切り）")
	cfnProtectCmd.Flags().Bool("enable", false, "削除保護を有効化")
	cfnProtectCmd.Flags().Bool("disable", false, "削除保護を無効化")
	cfnProtectCmd.Flags().BoolP("force", "f", false, "確認プロンプトをスキップ")

	// cfn drift-detectコマンド用のフラグ
	cfnDriftDetectCmd.Flags().StringP("filter", "F", "", "スタック名のフィルター（部分一致）")
	cfnDriftDetectCmd.Flags().BoolP("all", "a", false, "すべてのスタックを対象")

	// cfn drift-statusコマンド用のフラグ
	cfnDriftStatusCmd.Flags().StringP("filter", "F", "", "スタック名のフィルター（部分一致）")
	cfnDriftStatusCmd.Flags().BoolP("all", "a", false, "すべてのスタックを対象")
	cfnDriftStatusCmd.Flags().BoolP("drifted-only", "d", false, "ドリフトしているスタックのみ表示")
}
