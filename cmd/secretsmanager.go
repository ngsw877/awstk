package cmd

import (
	"awstk/internal/aws"
	secretsmgrSvc "awstk/internal/service/secretsmanager"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/spf13/cobra"
)

// SecretsManagerCmd represents the secretsmanager command
var SecretsManagerCmd = &cobra.Command{
	Use:   "secretsmanager",
	Short: "AWS Secrets Managerリソース操作コマンド",
	Long:  `AWS Secrets Managerのシークレットを操作するためのコマンド群です。`,
}

var secretsManagerGetCmd = &cobra.Command{
	Use:   "get <secret-name>",
	Short: "Secrets Managerからシークレット値を取得するコマンド",
	Long: `指定したSecrets Managerのシークレット名またはARNから値を取得し、JSON形式で出力します。

例:
  ` + AppName + ` secretsmanager get my-secret-name
  ` + AppName + ` secretsmanager get arn:aws:secretsmanager:ap-northeast-1:123456789012:secret:my-secret-abc123`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		secretName := args[0]

		cfg, err := aws.LoadAwsConfig(awsCtx)
		if err != nil {
			return fmt.Errorf("AWS設定の読み込みエラー: %w", err)
		}
		secretsClient := secretsmanager.NewFromConfig(cfg)

		fmt.Printf("🔍 シークレット (%s) の値を取得します...\n", secretName)

		secretMap, err := secretsmgrSvc.GetSecretValues(secretsClient, secretName)
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

func init() {
	RootCmd.AddCommand(SecretsManagerCmd)
	SecretsManagerCmd.AddCommand(secretsManagerGetCmd)
}
