package cmd

import (
	"awsfunc/internal"
	"fmt"

	"github.com/spf13/cobra"
)

var SesCmd = &cobra.Command{
	Use:   "ses",
	Short: "SESリソース操作コマンド",
}

var sesVerifyCmd = &cobra.Command{
	Use:   "verify [email1] [email2] ...",
	Short: "メールアドレスをSESで検証する",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		failedEmails, err := internal.VerifySesEmails(region, profile, args)
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
}
