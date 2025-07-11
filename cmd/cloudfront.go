package cmd

import (
	cfsvc "awstk/internal/service/cloudfront"
	"awstk/internal/service/cloudfront/tenant"
	"awstk/internal/service/cfn"
	"awstk/internal/service/common"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/spf13/cobra"
)

var cfClient *cloudfront.Client

// CfCmd represents the cf command
var CfCmd = &cobra.Command{
	Use:          "cf",
	Short:        "CloudFrontリソース操作コマンド",
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// 親のPersistentPreRunEを実行（awsCtx設定とAWS設定読み込み）
		if err := RootCmd.PersistentPreRunE(cmd, args); err != nil {
			return err
		}

		// CloudFront用クライアント生成
		cfClient = cloudfront.NewFromConfig(awsCfg)
		cfnClient = cloudformation.NewFromConfig(awsCfg)

		return nil
	},
}

// cfInvalidateCmd represents the invalidate command
var cfInvalidateCmd = &cobra.Command{
	Use:   "invalidate [distribution-id]",
	Short: "CloudFrontのキャッシュを無効化するコマンド",
	Long: `CloudFrontディストリビューションのキャッシュを無効化します。
ディストリビューションIDを直接指定するか、CloudFormationスタック名から自動検出できます。

【使い方】
  ` + AppName + ` cf invalidate ABCD1234EFGH                    # 全体を無効化（/*）
  ` + AppName + ` cf invalidate ABCD1234EFGH -p "/images/*"     # 特定パスを無効化
  ` + AppName + ` cf invalidate -S my-stack                      # スタックから自動検出
  ` + AppName + ` cf invalidate -S my-stack -p "/api/*" -w       # 完了まで待機

【例】
  ` + AppName + ` cf invalidate E2ABC123DEF456 -p "/images/*" -p "/api/*"
  → 複数のパスを同時に無効化します`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmdCobra *cobra.Command, args []string) error {
		resolveStackName()
		paths, _ := cmdCobra.Flags().GetStringSlice("path")
		wait, _ := cmdCobra.Flags().GetBool("wait")

		var distributionId string
		var err error

		// ディストリビューションIDの取得
		if len(args) > 0 {
			distributionId = args[0]
		} else if stackName != "" {
			// スタックからCloudFrontディストリビューションを取得
			distributions, err := cfn.GetAllCloudFrontFromStack(cfnClient, stackName)
			if err != nil {
				return fmt.Errorf("❌ CloudFormationスタックからディストリビューションの取得に失敗: %w", err)
			}

			if len(distributions) == 0 {
				return fmt.Errorf("❌ スタック '%s' にCloudFrontディストリビューションが見つかりませんでした", stackName)
			}

			if len(distributions) == 1 {
				distributionId = distributions[0]
				fmt.Printf("✅ CloudFormationスタック '%s' からCloudFrontディストリビューション '%s' を検出しました\n", stackName, distributionId)
			} else {
				// 複数のディストリビューションがある場合は選択
				distributionId, err = cfsvc.SelectDistribution(cfClient, distributions)
				if err != nil {
					return fmt.Errorf("❌ ディストリビューション選択エラー: %w", err)
				}
			}
		} else {
			return fmt.Errorf("❌ エラー: ディストリビューションID またはスタック名 (-S) を指定してください")
		}

		fmt.Printf("🚀 CloudFrontディストリビューション (%s) のキャッシュを無効化します...\n", distributionId)
		fmt.Printf("   対象パス: %v\n", paths)

		// キャッシュ無効化の実行
		invalidationId, err := cfsvc.CreateInvalidation(cfClient, distributionId, paths)
		if err != nil {
			return fmt.Errorf("❌ キャッシュ無効化エラー: %w", err)
		}

		fmt.Printf("✅ キャッシュ無効化を開始しました (ID: %s)\n", invalidationId)

		// 待機オプションが有効な場合
		if wait {
			fmt.Println("⏳ 無効化の完了を待機しています...")
			err = cfsvc.WaitForInvalidation(cfClient, distributionId, invalidationId)
			if err != nil {
				return fmt.Errorf("❌ 無効化待機エラー: %w", err)
			}
			fmt.Println("✅ キャッシュ無効化が完了しました")
		}

		return nil
	},
}

// cfTenantCmd represents the tenant command
var cfTenantCmd = &cobra.Command{
	Use:   "tenant",
	Short: "CloudFrontマルチテナントディストリビューション操作",
	Long:  `CloudFrontマルチテナントディストリビューションのテナントを操作するためのコマンド群です。`,
}

// cfTenantListCmd represents the tenant list command
var cfTenantListCmd = &cobra.Command{
	Use:   "list <distribution-id>",
	Short: "マルチテナントディストリビューションのテナント一覧を表示",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmdCobra *cobra.Command, args []string) error {
		distributionId := args[0]
		
		tenants, err := tenant.ListTenants(cfClient, distributionId)
		if err != nil {
			return common.FormatListError("テナント", err)
		}

		// テナントIDの一覧を作成
		tenantIds := make([]string, len(tenants))
		for i, t := range tenants {
			tenantIds[i] = t.Id
		}

		// 共通関数で表示
		title := fmt.Sprintf("テナント一覧 (ディストリビューション: %s)", distributionId)
		common.PrintNumberedList(common.ListOutput{
			Title:        title,
			Items:        tenantIds,
			ResourceName: "テナント",
		})

		return nil
	},
	SilenceUsage: true,
}

// cfTenantInvalidateCmd represents the tenant invalidate command
var cfTenantInvalidateCmd = &cobra.Command{
	Use:   "invalidate [distribution-id] [tenant-id]",
	Short: "マルチテナントディストリビューションのキャッシュを無効化",
	Long: `CloudFrontマルチテナントディストリビューションの特定テナントまたは全テナントのキャッシュを無効化します。

【使い方】
  ` + AppName + ` cf tenant invalidate ABCD1234EFGH tenant-123     # 特定テナント
  ` + AppName + ` cf tenant invalidate ABCD1234EFGH --all          # 全テナント
  ` + AppName + ` cf tenant invalidate ABCD1234EFGH --list        # テナント一覧から選択

【例】
  ` + AppName + ` cf tenant invalidate E2ABC123DEF456 --all -p "/api/*"
  → 全テナントの /api/* パスを無効化します`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmdCobra *cobra.Command, args []string) error {
		paths, _ := cmdCobra.Flags().GetStringSlice("path")
		all, _ := cmdCobra.Flags().GetBool("all")
		list, _ := cmdCobra.Flags().GetBool("list")
		wait, _ := cmdCobra.Flags().GetBool("wait")

		var distributionId string
		var tenantId string
		var err error

		// ディストリビューションIDを引数から取得
		distributionId = args[0]
		
		// テナントIDの処理
		if all {
			// 全テナント無効化
			fmt.Printf("🚀 CloudFrontディストリビューション (%s) の全テナントのキャッシュを無効化します...\n", distributionId)
			err = tenant.InvalidateAllTenants(cfClient, distributionId, paths, wait)
			if err != nil {
				return fmt.Errorf("❌ 全テナントキャッシュ無効化エラー: %w", err)
			}
			fmt.Println("✅ 全テナントのキャッシュ無効化を開始しました")
		} else if list {
			// テナント一覧から選択
			tenantId, err = tenant.SelectTenant(cfClient, distributionId)
			if err != nil {
				return fmt.Errorf("❌ テナント選択エラー: %w", err)
			}
			err = tenant.InvalidateTenant(cfClient, distributionId, tenantId, paths, wait)
			if err != nil {
				return fmt.Errorf("❌ キャッシュ無効化エラー: %w", err)
			}
			fmt.Printf("✅ テナント '%s' のキャッシュ無効化を開始しました\n", tenantId)
		} else {
			// 特定テナント
			if len(args) < 2 {
				return fmt.Errorf("❌ エラー: テナントID、--all、または --list オプションを指定してください")
			}
			tenantId = args[1]
			fmt.Printf("🚀 テナント (%s) のキャッシュを無効化します...\n", tenantId)
			fmt.Printf("   対象パス: %v\n", paths)
			
			err = tenant.InvalidateTenant(cfClient, distributionId, tenantId, paths, wait)
			if err != nil {
				return fmt.Errorf("❌ キャッシュ無効化エラー: %w", err)
			}
			fmt.Printf("✅ テナント '%s' のキャッシュ無効化を開始しました\n", tenantId)
		}

		return nil
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(CfCmd)
	CfCmd.AddCommand(cfInvalidateCmd)
	CfCmd.AddCommand(cfTenantCmd)
	
	// tenant サブコマンドに list, invalidate を追加
	cfTenantCmd.AddCommand(cfTenantListCmd)
	cfTenantCmd.AddCommand(cfTenantInvalidateCmd)

	// フラグの追加
	cfInvalidateCmd.Flags().StringSliceP("path", "p", []string{"/*"}, "無効化するパス（デフォルト: /*）")
	cfInvalidateCmd.Flags().BoolP("wait", "w", false, "無効化完了まで待機")
	cfInvalidateCmd.Flags().StringVarP(&stackName, "stack", "S", "", "CloudFormationスタック名")
	
	// tenant invalidate フラグ
	cfTenantInvalidateCmd.Flags().StringSliceP("path", "p", []string{"/*"}, "無効化するパス（デフォルト: /*）")
	cfTenantInvalidateCmd.Flags().BoolP("all", "a", false, "全テナントを無効化")
	cfTenantInvalidateCmd.Flags().BoolP("list", "l", false, "テナント一覧から選択")
	cfTenantInvalidateCmd.Flags().BoolP("wait", "w", false, "無効化完了まで待機")
}