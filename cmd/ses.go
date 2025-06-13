package cmd

import (
	"awstk/internal/aws"
	"awstk/internal/service"
	"bufio"
	"fmt"
	"os"
	"strings"

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

		// ファイルからメールアドレスを読み込み
		emails, err := readEmailsFromFile(emailFile)
		if err != nil {
			return fmt.Errorf("❌ ファイル読み込みエラー: %w", err)
		}

		if len(emails) == 0 {
			return fmt.Errorf("❌ エラー: ファイルにメールアドレスが見つかりませんでした")
		}

		// 重複を除去
		filtered := removeDuplicates(emails)
		if len(filtered) != len(emails) {
			fmt.Printf("重複するメールアドレスを除去しました: %d件 → %d件\n", len(emails), len(filtered))
		}

		awsClients, err := aws.NewAwsClients(aws.AwsContext{Region: region, Profile: profile})
		if err != nil {
			return fmt.Errorf("AWS設定の読み込みエラー: %w", err)
		}

		sesClient := awsClients.Ses()

		failedEmails, err := service.VerifySesEmails(sesClient, filtered)
		if err != nil {
			return fmt.Errorf("❌ SES検証エラー: %w", err)
		}

		if len(failedEmails) > 0 {
			fmt.Printf("❌ 検証に失敗したメールアドレス: %d件\n", len(failedEmails))
			for _, email := range failedEmails {
				fmt.Printf("  - %s\n", email)
			}
		} else {
			fmt.Println("✅ すべてのメールアドレスの検証リクエストが完了しました")
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

// readEmailsFromFile はファイルからメールアドレス一覧を読み込みます
func readEmailsFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var emails []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && strings.Contains(line, "@") {
			emails = append(emails, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return emails, nil
}

// removeDuplicates は文字列スライスから重複を除去します
func removeDuplicates(emails []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, email := range emails {
		if !seen[email] {
			seen[email] = true
			result = append(result, email)
		}
	}

	return result
}
