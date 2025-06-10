package cmd

import (
	"awstk/internal"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	auroraClusterId string
)

var AuroraCmd = &cobra.Command{
	Use:   "aurora",
	Short: "Auroraリソース操作コマンド",
	Long:  `Auroraリソースを操作するためのコマンド群です。`,
}

var auroraStartClusterCmd = &cobra.Command{
	Use:   "start",
	Short: "Aurora DBクラスターを起動するコマンド",
	Long: `Aurora DBクラスターを起動するコマンドです。
CloudFormationスタック名を指定するか、クラスター識別子を直接指定することができます。

例:
  ` + AppName + ` aurora start -P my-profile -S my-stack
  ` + AppName + ` aurora start -P my-profile -c my-aurora-cluster`,
	RunE: func(cmd *cobra.Command, args []string) error {
		awsCtx := getAwsContext()

		var clusterId string
		var err error

		if stackName != "" {
			fmt.Println("CloudFormationスタックからAurora情報を取得します...")
			clusterId, err = internal.GetAuroraFromStack(awsCtx, stackName)
			if err != nil {
				return fmt.Errorf("❌ エラー: %w", err)
			}
			fmt.Printf("🔍 検出されたAuroraクラスター: %s\n", clusterId)
		} else if auroraClusterId != "" {
			clusterId = auroraClusterId
		} else {
			cmd.Help()
			return fmt.Errorf("❌ エラー: スタック名 (-S) またはAuroraクラスター識別子 (-c) が必須です")
		}

		fmt.Printf("🚀 Aurora DBクラスター '%s' を起動します...\n", clusterId)
		err = internal.StartAuroraCluster(awsCtx, clusterId)
		if err != nil {
			return fmt.Errorf("❌ エラー: %w", err)
		}

		fmt.Printf("✅ Aurora DBクラスター '%s' の起動を開始しました。起動完了まで数十分かかります。\n", clusterId)
		return nil
	},
	SilenceUsage: true,
}

var auroraStopClusterCmd = &cobra.Command{
	Use:   "stop",
	Short: "Aurora DBクラスターを停止するコマンド",
	Long: `Aurora DBクラスターを停止するコマンドです。
CloudFormationスタック名を指定するか、クラスター識別子を直接指定することができます。

例:
  ` + AppName + ` aurora stop -P my-profile -S my-stack
  ` + AppName + ` aurora stop -P my-profile -c my-aurora-cluster`,
	RunE: func(cmd *cobra.Command, args []string) error {
		awsCtx := getAwsContext()

		var clusterId string
		var err error

		if stackName != "" {
			fmt.Println("CloudFormationスタックからAurora情報を取得します...")
			clusterId, err = internal.GetAuroraFromStack(awsCtx, stackName)
			if err != nil {
				return fmt.Errorf("❌ エラー: %w", err)
			}
			fmt.Printf("🔍 検出されたAuroraクラスター: %s\n", clusterId)
		} else if auroraClusterId != "" {
			clusterId = auroraClusterId
		} else {
			cmd.Help()
			return fmt.Errorf("❌ エラー: スタック名 (-S) またはAuroraクラスター識別子 (-c) が必須です")
		}

		fmt.Printf("🛑 Aurora DBクラスター '%s' を停止します...\n", clusterId)
		err = internal.StopAuroraCluster(awsCtx, clusterId)
		if err != nil {
			return fmt.Errorf("❌ エラー: %w", err)
		}

		fmt.Printf("✅ Aurora DBクラスター '%s' の停止を開始しました\n", clusterId)
		return nil
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(AuroraCmd)
	AuroraCmd.AddCommand(auroraStartClusterCmd)
	AuroraCmd.AddCommand(auroraStopClusterCmd)

	// startコマンドのフラグを設定
	auroraStartClusterCmd.Flags().StringVarP(&stackName, "stack", "S", "", "CloudFormationスタック名")
	auroraStartClusterCmd.Flags().StringVarP(&auroraClusterId, "cluster", "c", "", "Auroraクラスター識別子 (-Sが指定されていない場合に必須)")

	// stopコマンドのフラグを設定
	auroraStopClusterCmd.Flags().StringVarP(&stackName, "stack", "S", "", "CloudFormationスタック名")
	auroraStopClusterCmd.Flags().StringVarP(&auroraClusterId, "cluster", "c", "", "Auroraクラスター識別子 (-Sが指定されていない場合に必須)")
}
