package ses

import "github.com/aws/aws-sdk-go-v2/service/ses"

// VerifyOptions はメールアドレス検証のオプション
type VerifyOptions struct {
	SesClient *ses.Client
	FilePath  string
}

// VerifyResult は検証結果を表す構造体
type VerifyResult struct {
	TotalEmails         int
	SuccessfulEmails    int
	FailedEmails        []string
	DuplicateRemoved    int
	VerificationDetails []EmailVerificationDetail
}

// EmailVerificationDetail は個別のメール検証詳細
type EmailVerificationDetail struct {
	Email   string
	Success bool
	Error   error
}

