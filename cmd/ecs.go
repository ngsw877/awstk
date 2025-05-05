package cmd

import (
	"awsfunc/internal"
	"fmt"
	"github.com/spf13/cobra"
)

var (
	stackName     string
	clusterName   string
	serviceName   string
	containerName string
	minCapacity   int
	maxCapacity   int
)

var EcsCmd = &cobra.Command{
	Use:   "ecs",
	Short: "ECSリソース操作コマンド",
	Long:  `ECSリソースを操作するためのコマンド群です。`,
}

var ecsExecCmd = &cobra.Command{
	Use:   "exec",
	Short: "Fargateコンテナに接続するコマンド",
	Long: `Fargateコンテナにシェル接続するコマンドです。
CloudFormationスタック名を指定するか、クラスター名とサービス名を直接指定することができます。

例:
  awsfunc ecs exec -P my-profile -S my-stack
  awsfunc ecs exec -P my-profile -c my-cluster -s my-service -t app`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var cluster, service string

		// スタック名から情報取得
		if stackName != "" {
			fmt.Println("CloudFormationスタックからECS情報を取得します...")
			serviceInfo, err := internal.GetEcsFromStack(stackName, Region, Profile)
			if err != nil {
				return fmt.Errorf("❌ エラー: %w", err)
			}
			cluster = serviceInfo.ClusterName
			service = serviceInfo.ServiceName

			fmt.Println("🔍 検出されたクラスター: " + cluster)
			fmt.Println("🔍 検出されたサービス: " + service)
		} else if clusterName != "" && serviceName != "" {
			// クラスター名とサービス名が直接指定された場合
			cluster = clusterName
			service = serviceName
		} else {
			cmd.Help()
			return fmt.Errorf("❌ エラー: スタック名が指定されていない場合は、クラスター名 (-c) とサービス名 (-s) が必須です")
		}

		// タスクIDを取得
		taskId, err := internal.GetRunningTask(cluster, service, Region, Profile)
		if err != nil {
			return fmt.Errorf("❌ エラー: %w", err)
		}

		// シェル接続を実行
		fmt.Printf("🔍 コンテナ '%s' に接続しています...\n", containerName)
		err = internal.ExecuteCommand(cluster, taskId, containerName, Region, Profile)
		if err != nil {
			return fmt.Errorf("❌ コンテナへの接続に失敗しました: %w", err)
		}
		return nil
	},
	SilenceUsage: true,
}

// ecsStartCmd はECSサービスのキャパシティを設定して起動するコマンドです
var ecsStartCmd = &cobra.Command{
	Use:   "start",
	Short: "ECSサービスのキャパシティを設定して起動するコマンド",
	Long: `ECSサービスの最小・最大キャパシティを設定して起動するコマンドです。
CloudFormationスタック名を指定するか、クラスター名とサービス名を直接指定することができます。

例:
  awsfunc ecs start -P my-profile -S my-stack -m 1 -M 2
  awsfunc ecs start -P my-profile -c my-cluster -s my-service -m 1 -M 3`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var cluster, service string

		// スタック名から情報取得
		if stackName != "" {
			fmt.Println("CloudFormationスタックからECS情報を取得します...")
			serviceInfo, err := internal.GetEcsFromStack(stackName, Region, Profile)
			if err != nil {
				return fmt.Errorf("❌ エラー: %w", err)
			}
			cluster = serviceInfo.ClusterName
			service = serviceInfo.ServiceName

			fmt.Println("🔍 検出されたクラスター: " + cluster)
			fmt.Println("🔍 検出されたサービス: " + service)
		} else if clusterName != "" && serviceName != "" {
			// クラスター名とサービス名が直接指定された場合
			cluster = clusterName
			service = serviceName
		} else {
			cmd.Help()
			return fmt.Errorf("❌ エラー: スタック名が指定されていない場合は、クラスター名 (-c) とサービス名 (-s) が必須です")
		}

		// キャパシティ設定オプションを作成
		opts := internal.ServiceCapacityOptions{
			ClusterName: cluster,
			ServiceName: service,
			Region:      Region,
			Profile:     Profile,
			MinCapacity: minCapacity,
			MaxCapacity: maxCapacity,
		}

		// キャパシティを設定
		err := internal.SetEcsServiceCapacity(opts)
		if err != nil {
			return fmt.Errorf("❌ エラー: %w", err)
		}

		fmt.Println("✅ サービスが起動中です。")
		return nil
	},
	SilenceUsage: true,
}

// ecsStopCmd はECSサービスのキャパシティを0に設定して停止するコマンドです
var ecsStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "ECSサービスを停止するコマンド",
	Long: `ECSサービスの最小・最大キャパシティを0に設定して停止するコマンドです。
CloudFormationスタック名を指定するか、クラスター名とサービス名を直接指定することができます。

例:
  awsfunc ecs stop -P my-profile -S my-stack
  awsfunc ecs stop -P my-profile -c my-cluster -s my-service`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var cluster, service string

		// スタック名から情報取得
		if stackName != "" {
			fmt.Println("CloudFormationスタックからECS情報を取得します...")
			serviceInfo, err := internal.GetEcsFromStack(stackName, Region, Profile)
			if err != nil {
				return fmt.Errorf("❌ エラー: %w", err)
			}
			cluster = serviceInfo.ClusterName
			service = serviceInfo.ServiceName

			fmt.Println("🔍 検出されたクラスター: " + cluster)
			fmt.Println("🔍 検出されたサービス: " + service)
		} else if clusterName != "" && serviceName != "" {
			// クラスター名とサービス名が直接指定された場合
			cluster = clusterName
			service = serviceName
		} else {
			cmd.Help()
			return fmt.Errorf("❌ エラー: スタック名が指定されていない場合は、クラスター名 (-c) とサービス名 (-s) が必須です")
		}

		// キャパシティ設定オプションを作成（停止のため0に設定）
		opts := internal.ServiceCapacityOptions{
			ClusterName: cluster,
			ServiceName: service,
			Region:      Region,
			Profile:     Profile,
			MinCapacity: 0,
			MaxCapacity: 0,
		}

		// キャパシティを設定
		err := internal.SetEcsServiceCapacity(opts)
		if err != nil {
			return fmt.Errorf("❌ エラー: %w", err)
		}

		fmt.Println("✅ サービスが停止中です。")
		return nil
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(EcsCmd)
	EcsCmd.AddCommand(ecsExecCmd)
	EcsCmd.AddCommand(ecsStartCmd)
	EcsCmd.AddCommand(ecsStopCmd)

	// execコマンドのフラグを設定
	ecsExecCmd.Flags().StringVarP(&stackName, "stack", "S", "", "CloudFormationスタック名")
	ecsExecCmd.Flags().StringVarP(&clusterName, "cluster", "c", "", "ECSクラスター名 (-Sが指定されていない場合に必須)")
	ecsExecCmd.Flags().StringVarP(&serviceName, "service", "s", "", "ECSサービス名 (-Sが指定されていない場合に必須)")
	ecsExecCmd.Flags().StringVarP(&containerName, "container", "t", "app", "接続するコンテナ名")

	// startコマンドのフラグを設定
	ecsStartCmd.Flags().StringVarP(&stackName, "stack", "S", "", "CloudFormationスタック名")
	ecsStartCmd.Flags().StringVarP(&clusterName, "cluster", "c", "", "ECSクラスター名 (-Sが指定されていない場合に必須)")
	ecsStartCmd.Flags().StringVarP(&serviceName, "service", "s", "", "ECSサービス名 (-Sが指定されていない場合に必須)")
	ecsStartCmd.Flags().IntVarP(&minCapacity, "min", "m", 1, "最小キャパシティ")
	ecsStartCmd.Flags().IntVarP(&maxCapacity, "max", "M", 2, "最大キャパシティ")

	// stopコマンドのフラグを設定
	ecsStopCmd.Flags().StringVarP(&stackName, "stack", "S", "", "CloudFormationスタック名")
	ecsStopCmd.Flags().StringVarP(&clusterName, "cluster", "c", "", "ECSクラスター名 (-Sが指定されていない場合に必須)")
	ecsStopCmd.Flags().StringVarP(&serviceName, "service", "s", "", "ECSサービス名 (-Sが指定されていない場合に必須)")
}
