package cmd

import (
	"awstk/internal/service/aurora"
	"awstk/internal/service/cfn"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
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

func init() {
	RootCmd.AddCommand(AuroraCmd)
	AuroraCmd.AddCommand(auroraStartCmd)
	AuroraCmd.AddCommand(auroraStopCmd)
	AuroraCmd.AddCommand(auroraLsCmd)

	// フラグの追加
	auroraStartCmd.Flags().StringP("cluster", "c", "", "Aurora DBクラスター名")
	auroraStartCmd.Flags().StringP("stack", "S", "", "CloudFormationスタック名")
	auroraStopCmd.Flags().StringP("cluster", "c", "", "Aurora DBクラスター名")
	auroraStopCmd.Flags().StringP("stack", "S", "", "CloudFormationスタック名")
	auroraLsCmd.Flags().StringP("stack", "S", "", "CloudFormationスタック名")
}
