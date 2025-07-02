package cmd

import (
	sesSvc "awstk/internal/service/ses"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/spf13/cobra"
)

var emailFile string

// SesCmd represents the ses command
var SesCmd = &cobra.Command{
	Use:   "ses",
	Short: "SESリソース操作コマンド",
	Long:  `SES（Simple Email Service）を操作するためのコマンド群です。`,
}

var sesVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "SESメールアドレス検証コマンド",
	Long: `指定されたファイルからメールアドレス一覧を読み込み、SESで検証リクエストを送信します。

例:
  ` + AppName + ` ses verify -f emails.txt`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if emailFile == "" {
			return fmt.Errorf("❌ エラー: メールアドレスファイル (-f) を指定してください")
		}

		sesClient := ses.NewFromConfig(awsCfg)

		opts := sesSvc.VerifyOptions{
			SesClient: sesClient,
			FilePath:  emailFile,
		}

		result, err := sesSvc.VerifyEmailsFromFile(opts)
		if err != nil {
			return fmt.Errorf("❌ %v", err)
		}

		// 成功したメールアドレス
		fmt.Printf("✅ 検証成功: %d件\n", result.SuccessfulEmails)
		for _, detail := range result.VerificationDetails {
			if detail.Success {
				fmt.Printf("  - %s\n", detail.Email)
			}
		}

		// 失敗したメールアドレス
		if len(result.FailedEmails) > 0 {
			fmt.Printf("\n❌ 検証失敗: %d件\n", len(result.FailedEmails))
			for _, email := range result.FailedEmails {
				fmt.Printf("  - %s\n", email)
			}
		}

		return nil
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(SesCmd)
	SesCmd.AddCommand(sesVerifyCmd)
	sesVerifyCmd.Flags().StringVarP(&emailFile, "file", "f", "", "メールアドレス一覧ファイル（1行1メールアドレス）")
}
