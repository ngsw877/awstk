package cmd

import (
	"awsfunc/internal"
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var emailFile string

var SesCmd = &cobra.Command{
	Use:   "ses",
	Short: "SESリソース操作コマンド",
	Long:  "AWS SES（Simple Email Service）のリソースを操作するためのコマンド群です。",
}

var sesVerifyCmd = &cobra.Command{
	Use:   "verify [email1] [email2] ...",
	Short: "メールアドレスをSESで検証する",
	Long: `指定したメールアドレスをAWS SESで検証します。
コマンドライン引数または--fileオプションでメールアドレスを指定できます。
--fileオプションを使う場合は、1行1メールアドレスのテキストファイルを指定してください。
例:
  awsfunc ses verify user1@example.com user2@example.com
  awsfunc ses verify --file emails.txt
  awsfunc ses verify user1@example.com --file emails.txt`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var emails []string

		// --fileオプションでファイルから読み込み
		if emailFile != "" {
			file, err := os.Open(emailFile)
			if err != nil {
				return fmt.Errorf("ファイルのオープンに失敗: %w", err)
			}
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line != "" {
					emails = append(emails, line)
				}
			}
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("ファイルの読み込みに失敗: %w", err)
			}
		}

		// 引数も追加
		emails = append(emails, args...)

		// 重複除去
		uniq := make(map[string]struct{})
		var filtered []string
		for _, e := range emails {
			e = strings.TrimSpace(e)
			if e == "" {
				continue
			}
			if _, ok := uniq[e]; !ok {
				uniq[e] = struct{}{}
				filtered = append(filtered, e)
			}
		}
		if len(filtered) == 0 {
			return fmt.Errorf("メールアドレスが指定されていません")
		}

		failedEmails, err := internal.VerifySesEmails(region, profile, filtered)
		if err != nil {
			return err
		}
		if len(failedEmails) > 0 {
			fmt.Println("検証に失敗したメールアドレス:")
			for _, email := range failedEmails {
				fmt.Println(" -", email)
			}
			return fmt.Errorf("一部のメールアドレスの検証に失敗しました")
		}
		fmt.Println("すべてのメールアドレスの検証が成功しました。")
		return nil
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(SesCmd)
	SesCmd.AddCommand(sesVerifyCmd)
	sesVerifyCmd.Flags().StringVarP(&emailFile, "file", "f", "", "メールアドレス一覧ファイル（1行1メールアドレス）")
}
