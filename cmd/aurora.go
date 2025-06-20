package cmd

import (
	"awstk/internal/aws"
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
		clusterName, _ := cmd.Flags().GetString("cluster")
		stackName, _ := cmd.Flags().GetString("stack")
		var err error

		cfg, err := aws.LoadAwsConfig(awsCtx)
		if err != nil {
			return fmt.Errorf("AWS設定の読み込みエラー: %w", err)
		}

		if stackName != "" {
			cfnClient := cloudformation.NewFromConfig(cfg)

			clusterName, err = cfn.GetAuroraFromStack(cfnClient, stackName)
			if err != nil {
				return fmt.Errorf("❌ CloudFormationスタックからクラスター名の取得に失敗: %w", err)
			}
			fmt.Printf("✅ CloudFormationスタック '%s' からAuroraクラスター '%s' を検出しました\n", stackName, clusterName)
		} else if clusterName == "" {
			return fmt.Errorf("❌ エラー: Auroraクラスター名 (-c) またはスタック名 (-S) を指定してください")
		}

		rdsClient := rds.NewFromConfig(cfg)

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
		clusterName, _ := cmd.Flags().GetString("cluster")
		stackName, _ := cmd.Flags().GetString("stack")
		var err error

		cfg, err := aws.LoadAwsConfig(awsCtx)
		if err != nil {
			return fmt.Errorf("AWS設定の読み込みエラー: %w", err)
		}

		if stackName != "" {
			cfnClient := cloudformation.NewFromConfig(cfg)

			clusterName, err = cfn.GetAuroraFromStack(cfnClient, stackName)
			if err != nil {
				return fmt.Errorf("❌ CloudFormationスタックからクラスター名の取得に失敗: %w", err)
			}
			fmt.Printf("✅ CloudFormationスタック '%s' からAuroraクラスター '%s' を検出しました\n", stackName, clusterName)
		} else if clusterName == "" {
			return fmt.Errorf("❌ エラー: Auroraクラスター名 (-c) またはスタック名 (-S) を指定してください")
		}

		rdsClient := rds.NewFromConfig(cfg)

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

func init() {
	RootCmd.AddCommand(AuroraCmd)
	AuroraCmd.AddCommand(auroraStartCmd)
	AuroraCmd.AddCommand(auroraStopCmd)

	// フラグの追加
	auroraStartCmd.Flags().StringP("cluster", "c", "", "Aurora DBクラスター名")
	auroraStartCmd.Flags().StringP("stack", "S", "", "CloudFormationスタック名")
	auroraStopCmd.Flags().StringP("cluster", "c", "", "Aurora DBクラスター名")
	auroraStopCmd.Flags().StringP("stack", "S", "", "CloudFormationスタック名")
}
