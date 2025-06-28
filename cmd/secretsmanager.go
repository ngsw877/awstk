package cmd

import (
	secretsmgrSvc "awstk/internal/service/secretsmanager"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/spf13/cobra"
)

var secretsmanagerClient *secretsmanager.Client

// secretsmanagerCmd represents the secretsmanager command
var secretsmanagerCmd = &cobra.Command{
	Use:   "secrets",
	Short: "AWS Secrets Managerリソース操作コマンド",
	Long:  `AWS Secrets Managerのシークレットを操作するためのコマンド群です。`,
	// サブコマンド実行前にクライアントを初期化
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := RootCmd.PersistentPreRunE(cmd, args); err != nil {
			return err
		}

		secretsmanagerClient = secretsmanager.NewFromConfig(awsCfg)
		return nil
	},
}

var secretsmanagerGetCmd = &cobra.Command{
	Use:   "get <secret-name>",
	Short: "Secrets Managerからシークレット値を取得するコマンド",
	Long: `指定したSecrets Managerのシークレット名またはARNから値を取得し、JSON形式で出力します。

例:
  ` + AppName + ` secrets get my-secret-name
  ` + AppName + ` secrets get arn:aws:secretsmanager:ap-northeast-1:123456789012:secret:my-secret-abc123`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		secretName := args[0]

		fmt.Printf("🔍 シークレット (%s) の値を取得します...\n", secretName)

		secretMap, err := secretsmgrSvc.GetSecretValues(secretsmanagerClient, secretName)
		if err != nil {
			return fmt.Errorf("❌ シークレット取得エラー: %w", err)
		}

		// JSON形式で整形して出力
		jsonBytes, err := json.MarshalIndent(secretMap, "", "  ")
		if err != nil {
			return fmt.Errorf("❌ JSON変換エラー: %w", err)
		}

		fmt.Println(string(jsonBytes))
		return nil
	},
	SilenceUsage: true,
}

// secretsmanagerDeleteCmd represents the delete command
var secretsmanagerDeleteCmd = &cobra.Command{
	Use:   "delete <secret-id>",
	Short: "Secrets Managerのシークレットを即時削除します。",
	Long: `指定したシークレットを復旧期間なしで即時削除します。

この操作は元に戻すことができません。`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		secretId := args[0]

		if err := secretsmgrSvc.DeleteSecret(secretsmanagerClient, secretId); err != nil {
			return err
		}

		fmt.Printf("シークレット %s は正常に削除されました。\n", secretId)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(secretsmanagerCmd)
	secretsmanagerCmd.AddCommand(secretsmanagerGetCmd)
	secretsmanagerCmd.AddCommand(secretsmanagerDeleteCmd)
}
