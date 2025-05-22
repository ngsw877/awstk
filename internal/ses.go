package internal

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
)

// VerifySesEmails はSESでメールアドレス検証リクエストを送信する
func VerifySesEmails(awsCtx AwsContext, emails []string) ([]string, error) {
	cfg, err := LoadAwsConfig(awsCtx)
	if err != nil {
		return nil, fmt.Errorf("AWS設定の読み込みエラー: %w", err)
	}
	client := ses.NewFromConfig(cfg)

	var failedEmails []string
	for _, email := range emails {
		_, err := client.VerifyEmailIdentity(context.Background(), &ses.VerifyEmailIdentityInput{
			EmailAddress: aws.String(email),
		})
		if err != nil {
			fmt.Printf("❌ %s の検証に失敗: %v\n", email, err)
			failedEmails = append(failedEmails, email)
		} else {
			fmt.Printf("✅ %s の検証リクエストを送信しました\n", email)
		}
	}
	return failedEmails, nil
}
