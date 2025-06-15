package cmd

import (
	"fmt"

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

		if secretName == "" {
			return fmt.Errorf("❌ エラー: シークレット名 (-n) を指定してください")
		}

		fmt.Printf("🔍 シークレット (%s) の値を取得します...\n", secretName)
		// TODO: service.GetSecretValue関数を実装する必要があります
		fmt.Printf("⚠️ SecretsManager取得機能は未実装です\n")
		return nil
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(SecretsManagerCmd)
	SecretsManagerCmd.AddCommand(secretsManagerGetCmd)
}
