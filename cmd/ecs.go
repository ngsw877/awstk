package cmd

import (
	"awstk/internal/aws"
	"awstk/internal/service/cfn"
	ecssvc "awstk/internal/service/ecs"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
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
  ` + AppName + ` ecs exec -P my-profile -S my-stack
  ` + AppName + ` ecs exec -P my-profile -c my-cluster -s my-service -t app`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		clusterName, serviceName, err = resolveEcsClusterAndService(awsCtx)
		if err != nil {
			cmd.Help()
			return err
		}

		ecsClient, err := aws.NewClient[*ecs.Client](awsCtx)
		if err != nil {
			return fmt.Errorf("AWS設定の読み込みエラー: %w", err)
		}

		// タスクIDを取得
		taskId, err := ecssvc.GetRunningTask(ecsClient, clusterName, serviceName)
		if err != nil {
			return fmt.Errorf("❌ エラー: %w", err)
		}

		// シェル接続を実行
		fmt.Printf("🔍 コンテナ '%s' に接続しています...\n", containerName)
		err = ecssvc.ExecuteEcsCommand(ecssvc.EcsExecOptions{
			Region:        awsCtx.Region,
			Profile:       awsCtx.Profile,
			ClusterName:   clusterName,
			TaskId:        taskId,
			ContainerName: containerName,
		})
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
  ` + AppName + ` ecs start -P my-profile -S my-stack -m 1 -M 2
  ` + AppName + ` ecs start -P my-profile -c my-cluster -s my-service -m 1 -M 3
  ` + AppName + ` ecs start -P my-profile -S my-stack -m 1 -M 2`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		if stackName != "" {
			cfnClient, err := aws.NewClient[*cloudformation.Client](awsCtx)
			if err != nil {
				return fmt.Errorf("CloudFormationクライアント作成エラー: %w", err)
			}

			serviceInfo, stackErr := cfn.GetEcsFromStack(cfnClient, stackName)
			if stackErr != nil {
				return fmt.Errorf("❌ CloudFormationスタックからECSサービス情報の取得に失敗: %w", stackErr)
			}
			clusterName = serviceInfo.ClusterName
			serviceName = serviceInfo.ServiceName
		}

		autoScalingClient, err := aws.NewClient[*applicationautoscaling.Client](awsCtx)
		if err != nil {
			return fmt.Errorf("AWS設定の読み込みエラー: %w", err)
		}

		opts := ecssvc.ServiceCapacityOptions{
			ClusterName: clusterName,
			ServiceName: serviceName,
			MinCapacity: minCapacity,
			MaxCapacity: maxCapacity,
		}

		fmt.Println("🚀 サービスの起動を開始します...")
		err = ecssvc.SetEcsServiceCapacity(autoScalingClient, opts)
		if err != nil {
			return fmt.Errorf("❌ エラー: %w", err)
		}

		ecsClient, err := aws.NewClient[*ecs.Client](awsCtx)
		if err != nil {
			return fmt.Errorf("AWS設定の読み込みエラー: %w", err)
		}

		err = ecssvc.WaitForServiceStatus(ecsClient, opts, minCapacity, timeoutSeconds)
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
  ` + AppName + ` ecs stop -P my-profile -S my-stack
  ` + AppName + ` ecs stop -P my-profile -c my-cluster -s my-service
  ` + AppName + ` ecs stop -P my-profile -S my-stack`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		clusterName, serviceName, err = resolveEcsClusterAndService(awsCtx)
		if err != nil {
			cmd.Help()
			return err
		}

		autoScalingClient, err := aws.NewClient[*applicationautoscaling.Client](awsCtx)
		if err != nil {
			return fmt.Errorf("AWS設定の読み込みエラー: %w", err)
		}

		ecsClient, err := aws.NewClient[*ecs.Client](awsCtx)
		if err != nil {
			return fmt.Errorf("AWS設定の読み込みエラー: %w", err)
		}

		// キャパシティ設定オプションを作成（停止のため0に設定）
		opts := ecssvc.ServiceCapacityOptions{
			ClusterName: clusterName,
			ServiceName: serviceName,
			MinCapacity: 0,
			MaxCapacity: 0,
		}

		// キャパシティを設定
		fmt.Println("🛑 サービスの停止を開始します...")
		err = ecssvc.SetEcsServiceCapacity(autoScalingClient, opts)
		if err != nil {
			return fmt.Errorf("❌ エラー: %w", err)
		}

		// 停止完了を必ず待機
		err = ecssvc.WaitForServiceStatus(ecsClient, opts, 0, timeoutSeconds)
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
  ` + AppName + ` ecs run -P my-profile -S my-stack -t app -C "echo hello"
  ` + AppName + ` ecs run -P my-profile -c my-cluster -s my-service -t app -C "echo hello"
  ` + AppName + ` ecs run -P my-profile -S my-stack -t app -d my-task-def:1 -C "echo hello"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		clusterName, serviceName, err = resolveEcsClusterAndService(awsCtx)
		if err != nil {
			cmd.Help()
			return err
		}

		ecsClient, err := aws.NewClient[*ecs.Client](awsCtx)
		if err != nil {
			return fmt.Errorf("AWS設定の読み込みエラー: %w", err)
		}

		// タスク実行オプションを作成
		opts := ecssvc.RunAndWaitForTaskOptions{
			ClusterName:    clusterName,
			ServiceName:    serviceName,
			TaskDefinition: taskDefinition,
			ContainerName:  containerName,
			Command:        commandString,
			Region:         awsCtx.Region,
			Profile:        awsCtx.Profile,
			TimeoutSeconds: timeoutSeconds,
		}

		// タスクを実行して完了を待機
		fmt.Println("🚀 ECSタスクを実行します...")
		exitCode, err := ecssvc.RunAndWaitForTask(ecsClient, opts)
		if err != nil {
			return fmt.Errorf("❌ タスク実行エラー: %w", err)
		}

		fmt.Printf("✅ タスクが完了しました。終了コード: %d\n", exitCode)
		// 終了コードが0以外の場合はエラーとして扱う
		if exitCode != 0 {
			return fmt.Errorf("❌ タスクが異常終了しました。終了コード: %d", exitCode)
		}
		return nil
	},
	SilenceUsage: true,
}

// ecsRedeployCmd はECSサービスを強制再デプロイするコマンドです
var ecsRedeployCmd = &cobra.Command{
	Use:   "redeploy",
	Short: "ECSサービスを強制再デプロイするコマンド",
	Long: `ECSサービスを強制再デプロイするコマンドです。
パラメータストアの値を更新した後などに、新しい設定でタスクを再起動したい場合に使用します。
CloudFormationスタック名を指定するか、クラスター名とサービス名を直接指定することができます。
デフォルトでデプロイ完了まで待機します。--no-waitフラグを指定すると、待機せずに即座に終了します。

例:
  ` + AppName + ` ecs redeploy -P my-profile -S my-stack
  ` + AppName + ` ecs redeploy -P my-profile -c my-cluster -s my-service
  ` + AppName + ` ecs redeploy -P my-profile -S my-stack --no-wait`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		clusterName, serviceName, err = resolveEcsClusterAndService(awsCtx)
		if err != nil {
			cmd.Help()
			return err
		}

		ecsClient, err := aws.NewClient[*ecs.Client](awsCtx)
		if err != nil {
			return fmt.Errorf("AWS設定の読み込みエラー: %w", err)
		}

		// 強制再デプロイを実行
		err = ecssvc.ForceRedeployService(ecsClient, clusterName, serviceName)
		if err != nil {
			return fmt.Errorf("❌ エラー: %w", err)
		}

		// --no-waitフラグが指定されていない場合はデプロイ完了まで待機
		noWait, _ := cmd.Flags().GetBool("no-wait")
		if !noWait {
			err = ecssvc.WaitForDeploymentComplete(ecsClient, clusterName, serviceName, timeoutSeconds)
			if err != nil {
				return fmt.Errorf("❌ デプロイ完了待機エラー: %w", err)
			}
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
	EcsCmd.AddCommand(ecsRedeployCmd)

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

	// redeployコマンドのフラグを設定
	ecsRedeployCmd.Flags().StringVarP(&stackName, "stack", "S", "", "CloudFormationスタック名")
	ecsRedeployCmd.Flags().StringVarP(&clusterName, "cluster", "c", "", "ECSクラスター名 (-Sが指定されていない場合に必須)")
	ecsRedeployCmd.Flags().StringVarP(&serviceName, "service", "s", "", "ECSサービス名 (-Sが指定されていない場合に必須)")
	ecsRedeployCmd.Flags().IntVar(&timeoutSeconds, "timeout", 300, "待機タイムアウト（秒）")
	ecsRedeployCmd.Flags().Bool("no-wait", false, "デプロイ完了を待機せずに即座に終了する")
}

// resolveEcsClusterAndService はフラグの値に基づいて
// 操作対象のECSクラスター名とサービス名を取得するプライベートヘルパー関数。
func resolveEcsClusterAndService(awsCtx aws.Context) (string, string, error) {
	if stackName != "" {
		cfnClient, err := aws.NewClient[*cloudformation.Client](awsCtx)
		if err != nil {
			return "", "", fmt.Errorf("CloudFormationクライアント作成エラー: %w", err)
		}

		serviceInfo, stackErr := cfn.GetEcsFromStack(cfnClient, stackName)
		if stackErr != nil {
			return "", "", fmt.Errorf("❌ CloudFormationスタックからECSサービス情報の取得に失敗: %w", stackErr)
		}
		clusterName = serviceInfo.ClusterName
		serviceName = serviceInfo.ServiceName
	}

	// フラグから直接取得
	return clusterName, serviceName, nil
}
