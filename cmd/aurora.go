package cmd

import (
	"awstk/internal/service/aurora"
	"awstk/internal/service/cfn"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/spf13/cobra"
)

// AuroraCmd represents the aurora command
var AuroraCmd = &cobra.Command{
	Use:   "aurora",
	Short: "Aurora DBクラスター操作コマンド",
	Long:  `Aurora DBクラスターを操作するためのコマンド群です。`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// 親のPersistentPreRunEを実行（awsCtx設定とAWS設定読み込み）
		if err := RootCmd.PersistentPreRunE(cmd, args); err != nil {
			return err
		}

		// Aurora用クライアント生成
		rdsClient = rds.NewFromConfig(awsCfg)
		cfnClient = cloudformation.NewFromConfig(awsCfg)

		return nil
	},
}

var cwClient *cloudwatch.Client

var auroraStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Aurora DBクラスターを起動するコマンド",
	Long: `Aurora DBクラスターを起動します。
CloudFormationスタック名を指定するか、クラスター名を直接指定することができます。

例:
  ` + AppName + ` aurora start -P my-profile -S my-stack
  ` + AppName + ` aurora start -P my-profile -c my-cluster`,
	RunE: func(cmd *cobra.Command, args []string) error {
		resolveStackName()
		clusterName, _ := cmd.Flags().GetString("cluster")
		var err error

		if stackName != "" {
			clusterName, err = cfn.GetAuroraFromStack(cfnClient, stackName)
			if err != nil {
				return fmt.Errorf("❌ CloudFormationスタックからクラスター名の取得に失敗: %w", err)
			}
			fmt.Printf("✅ CloudFormationスタック '%s' からAuroraクラスター '%s' を検出しました\n", stackName, clusterName)
		} else if clusterName == "" {
			return fmt.Errorf("❌ エラー: Auroraクラスター名 (-c) またはスタック名 (-S) を指定してください")
		}

		fmt.Printf("🚀 Aurora DBクラスター (%s) を起動します...\n", clusterName)
		err = aurora.StartAuroraCluster(rdsClient, clusterName)
		if err != nil {
			return fmt.Errorf("❌ Aurora DBクラスター起動エラー: %w", err)
		}

		fmt.Printf("✅ Aurora DBクラスター (%s) の起動を開始しました\n", clusterName)
		return nil
	},
	SilenceUsage: true,
}

var auroraStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Aurora DBクラスターを停止するコマンド",
	Long: `Aurora DBクラスターを停止します。
CloudFormationスタック名を指定するか、クラスター名を直接指定することができます。

例:
  ` + AppName + ` aurora stop -P my-profile -S my-stack
  ` + AppName + ` aurora stop -P my-profile -c my-cluster`,
	RunE: func(cmd *cobra.Command, args []string) error {
		resolveStackName()
		clusterName, _ := cmd.Flags().GetString("cluster")
		var err error

		if stackName != "" {
			clusterName, err = cfn.GetAuroraFromStack(cfnClient, stackName)
			if err != nil {
				return fmt.Errorf("❌ CloudFormationスタックからクラスター名の取得に失敗: %w", err)
			}
			fmt.Printf("✅ CloudFormationスタック '%s' からAuroraクラスター '%s' を検出しました\n", stackName, clusterName)
		} else if clusterName == "" {
			return fmt.Errorf("❌ エラー: Auroraクラスター名 (-c) またはスタック名 (-S) を指定してください")
		}

		fmt.Printf("🛑 Aurora DBクラスター (%s) を停止します...\n", clusterName)
		err = aurora.StopAuroraCluster(rdsClient, clusterName)
		if err != nil {
			return fmt.Errorf("❌ Aurora DBクラスター停止エラー: %w", err)
		}

		fmt.Printf("✅ Aurora DBクラスター (%s) の停止を開始しました\n", clusterName)
		return nil
	},
	SilenceUsage: true,
}

var auroraLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "Auroraクラスター一覧を表示するコマンド",
	Long:  `Auroraクラスター一覧を表示します。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		resolveStackName()
		var (
			clusters []aurora.AuroraCluster
			err      error
		)

		if stackName != "" {
			clusters, err = aurora.ListAuroraClustersFromStack(rdsClient, cfnClient, stackName)
			if err != nil {
				return fmt.Errorf("❌ CloudFormationスタックからクラスター名の取得に失敗: %w", err)
			}
		} else {
			clusters, err = aurora.ListAuroraClusters(rdsClient)
			if err != nil {
				return fmt.Errorf("❌ Auroraクラスター一覧取得でエラー: %w", err)
			}
		}

		if len(clusters) == 0 {
			fmt.Println("Auroraクラスターが見つかりませんでした")
			return nil
		}

		fmt.Printf("Auroraクラスター一覧: (全%d件)\n", len(clusters))
		for i, cl := range clusters {
			fmt.Printf("  %3d. %s (%s) [%s]\n", i+1, cl.ClusterId, cl.Engine, cl.Status)
		}

		return nil
	},
	SilenceUsage: true,
}

var auroraAcuCmd = &cobra.Command{
	Use:   "acu",
	Short: "Aurora Serverless v2のAcu使用状況を表示",
	Long: `Aurora Serverless v2クラスターの現在のAcu（Aurora Capacity Units）使用状況を表示します。

例:
  ` + AppName + ` aurora acu -P my-profile -S my-stack
  ` + AppName + ` aurora acu -P my-profile -c my-cluster
  ` + AppName + ` aurora acu -P my-profile --all`,
	RunE: func(cmd *cobra.Command, args []string) error {
		resolveStackName()
		clusterName, _ := cmd.Flags().GetString("cluster")
		showAll, _ := cmd.Flags().GetBool("all")

		// CloudWatchクライアントを初期化
		cwClient = cloudwatch.NewFromConfig(awsCfg)

		if showAll {
			// 全Serverless v2クラスターのAcu情報を表示
			capacityInfos, err := aurora.ListAuroraCapacityInfo(rdsClient, cwClient)
			if err != nil {
				return fmt.Errorf("❌ Acu情報取得でエラー: %w", err)
			}

			if len(capacityInfos) == 0 {
				fmt.Println("Aurora Serverless v2クラスターが見つかりませんでした")
				return nil
			}

			fmt.Printf("Aurora Serverless v2 Acu使用状況: (全%d件)\n\n", len(capacityInfos))
			for _, info := range capacityInfos {
				displayCapacityInfo(&info)
				fmt.Println()
			}
			return nil
		}

		// 単一クラスターの処理
		if stackName != "" {
			var err error
			clusterName, err = cfn.GetAuroraFromStack(cfnClient, stackName)
			if err != nil {
				return fmt.Errorf("❌ CloudFormationスタックからクラスター名の取得に失敗: %w", err)
			}
			fmt.Printf("✅ CloudFormationスタック '%s' からAuroraクラスター '%s' を検出しました\n\n", stackName, clusterName)
		} else if clusterName == "" {
			return fmt.Errorf("❌ エラー: Auroraクラスター名 (-c) またはスタック名 (-S) を指定してください")
		}

		// Acu情報を取得
		info, err := aurora.GetAuroraCapacityInfo(rdsClient, cwClient, clusterName)
		if err != nil {
			return fmt.Errorf("❌ ACU情報取得でエラー: %w", err)
		}

		if !info.IsServerless {
			fmt.Printf("ℹ️ クラスター '%s' はServerless v2ではありません\n", clusterName)
			return nil
		}

		displayCapacityInfo(info)
		return nil
	},
	SilenceUsage: true,
}

func displayCapacityInfo(info *aurora.AuroraCapacityInfo) {
	fmt.Printf("📊 %s\n", info.ClusterId)
	if info.CurrentAcu >= 0 {
		if info.CurrentAcu == 0 {
			fmt.Printf("   Acu使用量: %.1f (過去5分間の平均 - アイドル状態)\n", info.CurrentAcu)
		} else {
			fmt.Printf("   Acu使用量: %.1f (過去5分間の平均値)\n", info.CurrentAcu)
		}
		fmt.Printf("   設定範囲: %.1f - %.1f Acu\n", info.MinAcu, info.MaxAcu)
	} else {
		fmt.Printf("   設定範囲: %.1f - %.1f Acu\n", info.MinAcu, info.MaxAcu)
		fmt.Println("   ⚠️  Acu使用量を取得できませんでした")
		fmt.Println("   💡 ヒント: クラスターが停止中、または CloudWatch にメトリクスがまだ記録されていない可能性があります")
	}
	fmt.Printf("   ステータス: %s\n", info.Status)
}

func init() {
	RootCmd.AddCommand(AuroraCmd)
	AuroraCmd.AddCommand(auroraStartCmd)
	AuroraCmd.AddCommand(auroraStopCmd)
	AuroraCmd.AddCommand(auroraLsCmd)
	AuroraCmd.AddCommand(auroraAcuCmd)

	// フラグの追加
	auroraStartCmd.Flags().StringP("cluster", "c", "", "Aurora DBクラスター名")
	auroraStartCmd.Flags().StringVarP(&stackName, "stack", "S", "", "CloudFormationスタック名")
	auroraStopCmd.Flags().StringP("cluster", "c", "", "Aurora DBクラスター名")
	auroraStopCmd.Flags().StringVarP(&stackName, "stack", "S", "", "CloudFormationスタック名")
	auroraLsCmd.Flags().StringVarP(&stackName, "stack", "S", "", "CloudFormationスタック名")
	auroraAcuCmd.Flags().StringP("cluster", "c", "", "Aurora DBクラスター名")
	auroraAcuCmd.Flags().StringVarP(&stackName, "stack", "S", "", "CloudFormationスタック名")
	auroraAcuCmd.Flags().BoolP("all", "a", false, "全てのServerless v2クラスターを表示")
}
