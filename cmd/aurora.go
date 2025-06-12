package cmd

import (
	"awstk/internal"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/spf13/cobra"
)

var (
	auroraClusterId string
)

var AuroraCmd = &cobra.Command{
	Use:   "aurora",
	Short: "Aurora DBクラスター操作コマンド",
	Long:  `Aurora DBクラスターを操作するためのコマンド群です。`,
}

var auroraStartCmd = &cobra.Command{
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
			fmt.Printf("🔍 検出されたAurora DBクラスター: %s\n", clusterId)
		} else if auroraClusterId != "" {
			clusterId = auroraClusterId
		} else {
			cmd.Help()
			return fmt.Errorf("❌ エラー: スタック名 (-S) またはAurora DBクラスター識別子 (-c) が必須です")
		}

		// AWS設定を読み込んでRDSクライアントを作成
		cfg, err := internal.LoadAwsConfig(awsCtx)
		if err != nil {
			return fmt.Errorf("AWS設定の読み込みエラー: %w", err)
		}
		rdsClient := rds.NewFromConfig(cfg)

		fmt.Printf("🚀 Aurora DBクラスター '%s' を起動します...\n", clusterId)
		err = internal.StartAuroraCluster(rdsClient, clusterId)
		if err != nil {
			return fmt.Errorf("❌ エラー: %w", err)
		}

		fmt.Printf("✅ Aurora DBクラスター '%s' の起動を開始しました\n", clusterId)
		return nil
	},
	SilenceUsage: true,
}

var auroraStopCmd = &cobra.Command{
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
			fmt.Printf("🔍 検出されたAurora DBクラスター: %s\n", clusterId)
		} else if auroraClusterId != "" {
			clusterId = auroraClusterId
		} else {
			cmd.Help()
			return fmt.Errorf("❌ エラー: スタック名 (-S) またはAurora DBクラスター識別子 (-c) が必須です")
		}

		// AWS設定を読み込んでRDSクライアントを作成
		cfg, err := internal.LoadAwsConfig(awsCtx)
		if err != nil {
			return fmt.Errorf("AWS設定の読み込みエラー: %w", err)
		}
		rdsClient := rds.NewFromConfig(cfg)

		fmt.Printf("🛑 Aurora DBクラスター '%s' を停止します...\n", clusterId)
		err = internal.StopAuroraCluster(rdsClient, clusterId)
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
	AuroraCmd.AddCommand(auroraStartCmd)
	AuroraCmd.AddCommand(auroraStopCmd)

	// startコマンドのフラグを設定
	auroraStartCmd.Flags().StringVarP(&stackName, "stack", "S", "", "CloudFormationスタック名")
	auroraStartCmd.Flags().StringVarP(&auroraClusterId, "cluster", "c", "", "Aurora DBクラスター識別子 (-Sが指定されていない場合に必須)")

	// stopコマンドのフラグを設定
	auroraStopCmd.Flags().StringVarP(&stackName, "stack", "S", "", "CloudFormationスタック名")
	auroraStopCmd.Flags().StringVarP(&auroraClusterId, "cluster", "c", "", "Aurora DBクラスター識別子 (-Sが指定されていない場合に必須)")
}
