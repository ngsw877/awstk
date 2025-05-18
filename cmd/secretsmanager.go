package cmd

import (
	"encoding/json"
	"github.com/spf13/cobra"

	"awsfunc/internal"
)

var SecretsManagerCmd = &cobra.Command{
	Use:   "secrets",
	Short: "Secrets Managerリソース操作コマンド",
	Long:  "AWS Secrets Managerのリソースを操作するためのコマンド群です。",
}

var secretsManagerGetCmd = &cobra.Command{
	Use:   "get <secret-name>",
	Short: "Secrets Managerからシークレット値を取得して全て出力する",
	Long: `指定したSecrets Managerのシークレット名またはARNから全ての値を取得し、標準出力にJSON形式で出力します。

【使用例】
  awsfunc secrets get my-secret-name
  awsfunc secrets get arn:aws:secretsmanager:ap-northeast-1:123456789012:secret:my-secret-abc123
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		secretName := args[0]
		secretMap, err := internal.GetSecretValues(region, profile, secretName)
		if err != nil {
			cmd.PrintErrf("エラー: %v\n", err)
			return err
		}
		secretJson, err := json.MarshalIndent(secretMap, "", " ")
		if err != nil {
			cmd.PrintErrf("JSON変換エラー: %v\n", err)
			return err
		}
		cmd.Println(string(secretJson))
		return nil
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(SecretsManagerCmd)
	SecretsManagerCmd.AddCommand(secretsManagerGetCmd)
}
