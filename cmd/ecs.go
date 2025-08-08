package cmd

import (
	"awstk/internal/aws"
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
	ecsClient      *ecs.Client
)

var EcsCmd = &cobra.Command{
	Use:   "ecs",
	Short: "ECSリソース操作コマンド",
	Long:  `ECSリソースを操作するためのコマンド群です。`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// 親のPersistentPreRunEを実行（awsCtx設定とAWS設定読み込み）
		if err := RootCmd.PersistentPreRunE(cmd, args); err != nil {
			return err
		}

		// ECS用クライアント生成
		ecsClient = ecs.NewFromConfig(awsCfg)

		return nil
	},
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

		resolveStackName()
		opts := ecssvc.ResolveOptions{
			StackName:   stackName,
			ClusterName: clusterName,
			ServiceName: serviceName,
		}
		cfnClient := cloudformation.NewFromConfig(awsCfg)
		clusterName, serviceName, err = ecssvc.ResolveClusterAndService(cfnClient, opts)
		if err != nil {
			return err
		}

		// タスクIDを取得
		taskId, err := ecssvc.GetRunningTask(ecsClient, clusterName, serviceName)
		if err != nil {
			return fmt.Errorf("❌ エラー: %w", err)
		}

		// シェル接続を実行
		fmt.Printf("🔍 コンテナ '%s' に接続しています...\n", containerName)
		awsCtx := aws.Context{Region: region, Profile: profile}
		err = ecssvc.ExecuteEcsCommand(awsCtx, ecssvc.ExecOptions{
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

		resolveStackName()
		opts := ecssvc.ResolveOptions{
			StackName:   stackName,
			ClusterName: clusterName,
			ServiceName: serviceName,
		}
		cfnClient := cloudformation.NewFromConfig(awsCfg)
		clusterName, serviceName, err = ecssvc.ResolveClusterAndService(cfnClient, opts)
		if err != nil {
			return err
		}

		aasClient := applicationautoscaling.NewFromConfig(awsCfg)

		startOpts := ecssvc.StartServiceOptions{
			ClusterName:    clusterName,
			ServiceName:    serviceName,
			MinCapacity:    minCapacity,
			MaxCapacity:    maxCapacity,
			TimeoutSeconds: timeoutSeconds,
		}
		err = ecssvc.StartEcsService(ecsClient, aasClient, startOpts)
		if err != nil {
			return err
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

		resolveStackName()
		opts := ecssvc.ResolveOptions{
			StackName:   stackName,
			ClusterName: clusterName,
			ServiceName: serviceName,
		}
		cfnClient := cloudformation.NewFromConfig(awsCfg)
		clusterName, serviceName, err = ecssvc.ResolveClusterAndService(cfnClient, opts)
		if err != nil {
			return err
		}

		aasClient := applicationautoscaling.NewFromConfig(awsCfg)

		stopOpts := ecssvc.StopServiceOptions{
			ClusterName:    clusterName,
			ServiceName:    serviceName,
			TimeoutSeconds: timeoutSeconds,
		}
		err = ecssvc.StopEcsService(ecsClient, aasClient, stopOpts)
		if err != nil {
			return err
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

		resolveStackName()
		opts := ecssvc.ResolveOptions{
			StackName:   stackName,
			ClusterName: clusterName,
			ServiceName: serviceName,
		}
		cfnClient := cloudformation.NewFromConfig(awsCfg)
		clusterName, serviceName, err = ecssvc.ResolveClusterAndService(cfnClient, opts)
		if err != nil {
			return err
		}

		// タスク実行オプションを作成
		runOpts := ecssvc.RunAndWaitForTaskOptions{
			ClusterName:    clusterName,
			ServiceName:    serviceName,
			TaskDefinition: taskDefinition,
			ContainerName:  containerName,
			Command:        commandString,
			TimeoutSeconds: timeoutSeconds,
		}

		// タスクを実行して完了を待機
		fmt.Println("🚀 ECSタスクを実行します...")
		exitCode, err := ecssvc.RunAndWaitForTask(ecsClient, runOpts)
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

		resolveStackName()
		opts := ecssvc.ResolveOptions{
			StackName:   stackName,
			ClusterName: clusterName,
			ServiceName: serviceName,
		}
		cfnClient := cloudformation.NewFromConfig(awsCfg)
		clusterName, serviceName, err = ecssvc.ResolveClusterAndService(cfnClient, opts)
		if err != nil {
			return err
		}

		// 強制再デプロイを実行
		err = ecssvc.ForceRedeployService(ecsClient, clusterName, serviceName)
		if err != nil {
			return fmt.Errorf("❌ エラー: %w", err)
		}

		// --no-waitフラグが指定されていない場合はデプロイ完了まで待機
		noWait, _ := cmd.Flags().GetBool("no-wait")
		if !noWait {
			waitOpts := ecssvc.WaitDeploymentOptions{
				ClusterName:    clusterName,
				ServiceName:    serviceName,
				TimeoutSeconds: timeoutSeconds,
			}
			err = ecssvc.WaitForDeploymentComplete(ecsClient, waitOpts)
			if err != nil {
				return fmt.Errorf("❌ デプロイ完了待機エラー: %w", err)
			}
		}
		return nil
	},
	SilenceUsage: true,
}

// ecsStatusCmd はECSサービスの状態を表示するコマンドです
var ecsStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "ECSサービスの状態を表示するコマンド",
	Long: `ECSサービスのタスク稼働状況を表示するコマンドです。
CloudFormationスタック名を指定するか、クラスター名とサービス名を直接指定することができます。

例:
  ` + AppName + ` ecs status -P my-profile -S my-stack
  ` + AppName + ` ecs status -P my-profile -c my-cluster -s my-service`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		resolveStackName()
		opts := ecssvc.ResolveOptions{
			StackName:   stackName,
			ClusterName: clusterName,
			ServiceName: serviceName,
		}
		cfnClient := cloudformation.NewFromConfig(awsCfg)
		clusterName, serviceName, err = ecssvc.ResolveClusterAndService(cfnClient, opts)
		if err != nil {
			return err
		}

		aasClient := applicationautoscaling.NewFromConfig(awsCfg)

		statusOpts := ecssvc.StatusOptions{
			ClusterName: clusterName,
			ServiceName: serviceName,
		}

		// サービス状態を取得
		status, err := ecssvc.GetServiceStatus(ecsClient, aasClient, statusOpts)
		if err != nil {
			return fmt.Errorf("❌ エラー: %w", err)
		}

		// 状態を表示
		ecssvc.ShowServiceStatus(status)
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
	EcsCmd.AddCommand(ecsStatusCmd)

	// execコマンドのフラグを設定
	ecsExecCmd.Flags().StringVarP(&stackName, "stack", "S", "", "CloudFormationスタック名")
	ecsExecCmd.Flags().StringVarP(&clusterName, "cluster", "c", "", "ECSクラスター名 (-Sが指定されていない場合に必須)")
	ecsExecCmd.Flags().StringVarP(&serviceName, "service", "s", "", "ECSサービス名 (-Sが指定されていない場合に必須)")
	ecsExecCmd.Flags().StringVarP(&containerName, "container", "t", "app", "接続するコンテナ名")
	ecsExecCmd.MarkFlagsMutuallyExclusive("stack", "cluster")
	ecsExecCmd.MarkFlagsMutuallyExclusive("stack", "service")
	ecsExecCmd.MarkFlagsRequiredTogether("cluster", "service")

	// startコマンドのフラグを設定
	ecsStartCmd.Flags().StringVarP(&stackName, "stack", "S", "", "CloudFormationスタック名")
	ecsStartCmd.Flags().StringVarP(&clusterName, "cluster", "c", "", "ECSクラスター名 (-Sが指定されていない場合に必須)")
	ecsStartCmd.Flags().StringVarP(&serviceName, "service", "s", "", "ECSサービス名 (-Sが指定されていない場合に必須)")
	ecsStartCmd.Flags().IntVarP(&minCapacity, "min", "m", 1, "最小キャパシティ")
	ecsStartCmd.Flags().IntVarP(&maxCapacity, "max", "M", 2, "最大キャパシティ")
	ecsStartCmd.Flags().IntVar(&timeoutSeconds, "timeout", 300, "待機タイムアウト（秒）")
	ecsStartCmd.MarkFlagsMutuallyExclusive("stack", "cluster")
	ecsStartCmd.MarkFlagsMutuallyExclusive("stack", "service")
	ecsStartCmd.MarkFlagsRequiredTogether("cluster", "service")

	// stopコマンドのフラグを設定
	ecsStopCmd.Flags().StringVarP(&stackName, "stack", "S", "", "CloudFormationスタック名")
	ecsStopCmd.Flags().StringVarP(&clusterName, "cluster", "c", "", "ECSクラスター名 (-Sが指定されていない場合に必須)")
	ecsStopCmd.Flags().StringVarP(&serviceName, "service", "s", "", "ECSサービス名 (-Sが指定されていない場合に必須)")
	ecsStopCmd.Flags().IntVar(&timeoutSeconds, "timeout", 300, "待機タイムアウト（秒）")
	ecsStopCmd.MarkFlagsMutuallyExclusive("stack", "cluster")
	ecsStopCmd.MarkFlagsMutuallyExclusive("stack", "service")
	ecsStopCmd.MarkFlagsRequiredTogether("cluster", "service")

	// runコマンドのフラグを設定
	ecsRunCmd.Flags().StringVarP(&stackName, "stack", "S", "", "CloudFormationスタック名")
	ecsRunCmd.Flags().StringVarP(&clusterName, "cluster", "c", "", "ECSクラスター名 (-Sが指定されていない場合に必須)")
	ecsRunCmd.Flags().StringVarP(&serviceName, "service", "s", "", "ECSサービス名 (-Sが指定されていない場合に必須)")
	ecsRunCmd.Flags().StringVarP(&containerName, "container", "t", "app", "実行するコンテナ名")
	ecsRunCmd.Flags().StringVarP(&taskDefinition, "task-definition", "d", "", "タスク定義 (指定しない場合はサービスのタスク定義を使用)")
	ecsRunCmd.Flags().StringVarP(&commandString, "command", "C", "", "実行するコマンド")
	ecsRunCmd.Flags().IntVar(&timeoutSeconds, "timeout", 300, "待機タイムアウト（秒）")
	ecsRunCmd.MarkFlagsMutuallyExclusive("stack", "cluster")
	ecsRunCmd.MarkFlagsMutuallyExclusive("stack", "service")
	ecsRunCmd.MarkFlagsRequiredTogether("cluster", "service")

	// redeployコマンドのフラグを設定
	ecsRedeployCmd.Flags().StringVarP(&stackName, "stack", "S", "", "CloudFormationスタック名")
	ecsRedeployCmd.Flags().StringVarP(&clusterName, "cluster", "c", "", "ECSクラスター名 (-Sが指定されていない場合に必須)")
	ecsRedeployCmd.Flags().StringVarP(&serviceName, "service", "s", "", "ECSサービス名 (-Sが指定されていない場合に必須)")
	ecsRedeployCmd.Flags().IntVar(&timeoutSeconds, "timeout", 300, "待機タイムアウト（秒）")
	ecsRedeployCmd.Flags().Bool("no-wait", false, "デプロイ完了を待機せずに即座に終了する")
	ecsRedeployCmd.MarkFlagsMutuallyExclusive("stack", "cluster")
	ecsRedeployCmd.MarkFlagsMutuallyExclusive("stack", "service")
	ecsRedeployCmd.MarkFlagsRequiredTogether("cluster", "service")

	// statusコマンドのフラグを設定
	ecsStatusCmd.Flags().StringVarP(&stackName, "stack", "S", "", "CloudFormationスタック名")
	ecsStatusCmd.Flags().StringVarP(&clusterName, "cluster", "c", "", "ECSクラスター名 (-Sが指定されていない場合に必須)")
	ecsStatusCmd.Flags().StringVarP(&serviceName, "service", "s", "", "ECSサービス名 (-Sが指定されていない場合に必須)")
	ecsStatusCmd.MarkFlagsMutuallyExclusive("stack", "cluster")
	ecsStatusCmd.MarkFlagsMutuallyExclusive("stack", "service")
	ecsStatusCmd.MarkFlagsRequiredTogether("cluster", "service")
}
