package cmd

import (
	"awsfunc/internal"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// stackName は root.go でグローバル変数として宣言されているため削除
	clusterName    string
	serviceName    string
	containerName  string
	minCapacity    int
	maxCapacity    int
	timeoutSeconds int
	taskDefinition string
	commandString  string
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
		var clusterName, serviceName string
		var err error

		clusterName, serviceName, err = resolveEcsClusterAndService()
		if err != nil {
			cmd.Help()
			return err
		}

		// タスクIDを取得
		taskId, err := internal.GetRunningTask(clusterName, serviceName, region, profile)
		if err != nil {
			return fmt.Errorf("❌ エラー: %w", err)
		}

		// シェル接続を実行
		fmt.Printf("🔍 コンテナ '%s' に接続しています...\n", containerName)
		err = internal.ExecuteCommand(clusterName, taskId, containerName, region, profile)
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
サービスが指定したキャパシティになるまで必ず待機します。待機タイムアウトは-t/--timeoutで秒数指定できます（デフォルト: 300秒）。

例:
  awsfunc ecs start -P my-profile -S my-stack -m 1 -M 2
  awsfunc ecs start -P my-profile -c my-cluster -s my-service -m 1 -M 3
  awsfunc ecs start -P my-profile -S my-stack -m 1 -M 2`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var clusterName, serviceName string
		var err error

		clusterName, serviceName, err = resolveEcsClusterAndService()
		if err != nil {
			cmd.Help()
			return err
		}

		// キャパシティ設定オプションを作成
		opts := internal.ServiceCapacityOptions{
			ClusterName: clusterName,
			ServiceName: serviceName,
			Region:      region,
			Profile:     profile,
			MinCapacity: minCapacity,
			MaxCapacity: maxCapacity,
		}

		// キャパシティを設定
		fmt.Println(" サービスの起動を開始します...")
		err = internal.SetEcsServiceCapacity(opts)
		if err != nil {
			return fmt.Errorf("❌ エラー: %w", err)
		}

		// 起動完了を必ず待機
		err = internal.WaitForServiceStatus(opts, minCapacity, timeoutSeconds)
		if err != nil {
			return fmt.Errorf("❌ サービス起動監視エラー: %w", err)
		}
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
サービスが完全に停止するまで必ず待機します。待機タイムアウトは-t/--timeoutで秒数指定できます（デフォルト: 300秒）。

例:
  awsfunc ecs stop -P my-profile -S my-stack
  awsfunc ecs stop -P my-profile -c my-cluster -s my-service
  awsfunc ecs stop -P my-profile -S my-stack`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var clusterName, serviceName string
		var err error

		clusterName, serviceName, err = resolveEcsClusterAndService()
		if err != nil {
			cmd.Help()
			return err
		}

		// キャパシティ設定オプションを作成（停止のため0に設定）
		opts := internal.ServiceCapacityOptions{
			ClusterName: clusterName,
			ServiceName: serviceName,
			Region:      region,
			Profile:     profile,
			MinCapacity: 0,
			MaxCapacity: 0,
		}

		// キャパシティを設定
		fmt.Println(" サービスの停止を開始します...")
		err = internal.SetEcsServiceCapacity(opts)
		if err != nil {
			return fmt.Errorf("❌ エラー: %w", err)
		}

		// 停止完了を必ず待機
		err = internal.WaitForServiceStatus(opts, 0, timeoutSeconds)
		if err != nil {
			return fmt.Errorf("❌ サービス停止監視エラー: %w", err)
		}
		return nil
	},
	SilenceUsage: true,
}

// ecsRunCmd はECSタスクを実行してその完了を待機するコマンドです
var ecsRunCmd = &cobra.Command{
	Use:   "run",
	Short: "ECSタスクを実行するコマンド",
	Long: `ECSタスクを実行してその完了を待機するコマンドです。
CloudFormationスタック名を指定するか、クラスター名とサービス名を直接指定することができます。
タスク定義は指定されていない場合、サービスで使用されている最新のタスク定義が使用されます。
待機タイムアウトは--timeoutで秒数指定できます（デフォルト: 300秒）。

例:
  awsfunc ecs run -P my-profile -S my-stack -t app -C "echo hello"
  awsfunc ecs run -P my-profile -c my-cluster -s my-service -t app -C "echo hello"
  awsfunc ecs run -P my-profile -S my-stack -t app -d my-task-def:1 -C "echo hello"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var clusterName, serviceName string
		var err error

		clusterName, serviceName, err = resolveEcsClusterAndService()
		if err != nil {
			cmd.Help()
			return err
		}

		// タスク実行オプションを作成
		opts := internal.RunAndWaitForTaskOptions{
			ClusterName:    clusterName,
			ServiceName:    serviceName,
			TaskDefinition: taskDefinition,
			ContainerName:  containerName,
			Command:        commandString,
			Region:         region,
			Profile:        profile,
			TimeoutSeconds: timeoutSeconds,
		}

		// タスクを実行して完了を待機
		fmt.Println("🚀 ECSタスクを実行します...")
		exitCode, err := internal.RunAndWaitForTask(opts)
		if err != nil {
			return fmt.Errorf("❌ タスク実行エラー: %w", err)
		}

		fmt.Printf("✅ タスクが完了しました。終了コード: %d\n", exitCode)
		// 終了コードが0以外の場合はエラーとして扱う
		if exitCode != 0 {
			return fmt.Errorf("タスクが非ゼロの終了コード %d で終了しました", exitCode)
		}
		return nil
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(EcsCmd)
	EcsCmd.AddCommand(ecsExecCmd)
	EcsCmd.AddCommand(ecsStartCmd)
	EcsCmd.AddCommand(ecsStopCmd)
	EcsCmd.AddCommand(ecsRunCmd)

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
	ecsStartCmd.Flags().IntVar(&timeoutSeconds, "timeout", 300, "待機タイムアウト（秒）")

	// stopコマンドのフラグを設定
	ecsStopCmd.Flags().StringVarP(&stackName, "stack", "S", "", "CloudFormationスタック名")
	ecsStopCmd.Flags().StringVarP(&clusterName, "cluster", "c", "", "ECSクラスター名 (-Sが指定されていない場合に必須)")
	ecsStopCmd.Flags().StringVarP(&serviceName, "service", "s", "", "ECSサービス名 (-Sが指定されていない場合に必須)")
	ecsStopCmd.Flags().IntVar(&timeoutSeconds, "timeout", 300, "待機タイムアウト（秒）")

	// runコマンドのフラグを設定
	ecsRunCmd.Flags().StringVarP(&stackName, "stack", "S", "", "CloudFormationスタック名")
	ecsRunCmd.Flags().StringVarP(&clusterName, "cluster", "c", "", "ECSクラスター名 (-Sが指定されていない場合に必須)")
	ecsRunCmd.Flags().StringVarP(&serviceName, "service", "s", "", "ECSサービス名 (-Sが指定されていない場合に必須)")
	ecsRunCmd.Flags().StringVarP(&containerName, "container", "t", "app", "実行するコンテナ名")
	ecsRunCmd.Flags().StringVarP(&taskDefinition, "task-definition", "d", "", "タスク定義 (指定しない場合はサービスのタスク定義を使用)")
	ecsRunCmd.Flags().StringVarP(&commandString, "command", "C", "", "実行するコマンド")
	ecsRunCmd.Flags().IntVar(&timeoutSeconds, "timeout", 300, "待機タイムアウト（秒）")
}

// resolveEcsClusterAndService はフラグの値に基づいて
// 操作対象のECSクラスター名とサービス名を取得するプライベートヘルパー関数。
func resolveEcsClusterAndService() (string, string, error) {
	if stackName != "" {
		fmt.Println("CloudFormationスタックからECS情報を取得します...")
		serviceInfo, stackErr := internal.GetEcsFromStack(stackName, region, profile)
		if stackErr != nil {
			return "", "", fmt.Errorf("❌ エラー: %w", stackErr)
		}
		// グローバル変数に値をセットする（必要に応じて）
		clusterName = serviceInfo.ClusterName
		serviceName = serviceInfo.ServiceName
		fmt.Println("🔍 検出されたクラスター: " + clusterName)
		fmt.Println("🔍 検出されたサービス: " + serviceName)
		return clusterName, serviceName, nil
	} else if clusterName != "" && serviceName != "" {
		return clusterName, serviceName, nil
	} else {
		return "", "", fmt.Errorf("❌ エラー: スタック名が指定されていない場合は、クラスター名 (-c) とサービス名 (-s) が必須です")
	}
}
