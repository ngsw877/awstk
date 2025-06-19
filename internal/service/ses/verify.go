package ses

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
)

// VerifySesEmails 指定されたメールアドレス一覧をSESで検証する
func VerifySesEmails(sesClient *ses.Client, emails []string) ([]string, error) {
	var failedEmails []string

	for _, email := range emails {
		fmt.Printf("検証中: %s\n", email)
		_, err := sesClient.VerifyEmailIdentity(context.Background(), &ses.VerifyEmailIdentityInput{
			EmailAddress: aws.String(email),
		})
		if err != nil {
			fmt.Printf("❌ 失敗: %s - %v\n", email, err)
			failedEmails = append(failedEmails, email)
		} else {
			fmt.Printf("✅ 成功: %s\n", email)
		}
	}

	return failedEmails, nil
}
