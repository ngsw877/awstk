package ses

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
)

// VerifyEmailsFromFile はファイルからメールアドレスを読み込んで検証する
func VerifyEmailsFromFile(opts VerifyOptions) (*VerifyResult, error) {
	// ファイルからメールアドレスを読み込み
	emails, err := readEmailsFromFile(opts.FilePath)
	if err != nil {
		return nil, fmt.Errorf("ファイル読み込みエラー: %w", err)
	}

	if len(emails) == 0 {
		return nil, fmt.Errorf("ファイルにメールアドレスが見つかりませんでした")
	}

	originalCount := len(emails)

	// 重複を除去
	emails = removeDuplicates(emails)
	duplicateRemoved := originalCount - len(emails)

	// メールアドレスを検証
	failedEmails, details, err := verifySesEmails(opts.SesClient, emails)
	if err != nil {
		return nil, fmt.Errorf("SES検証エラー: %w", err)
	}

	result := &VerifyResult{
		TotalEmails:         len(emails),
		SuccessfulEmails:    len(emails) - len(failedEmails),
		FailedEmails:        failedEmails,
		DuplicateRemoved:    duplicateRemoved,
		VerificationDetails: details,
	}

	return result, nil
}

// readEmailsFromFile はファイルからメールアドレス一覧を読み込む
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

		// 空行とコメント行をスキップ
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 基本的なメールアドレス検証（@を含む）
		if strings.Contains(line, "@") {
			emails = append(emails, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return emails, nil
}

// removeDuplicates は文字列スライスから重複を除去する
func removeDuplicates(emails []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, email := range emails {
		normalizedEmail := strings.ToLower(strings.TrimSpace(email))
		if !seen[normalizedEmail] {
			seen[normalizedEmail] = true
			result = append(result, email)
		}
	}

	return result
}

// verifySesEmails 指定されたメールアドレス一覧をSESで検証する
func verifySesEmails(sesClient *ses.Client, emails []string) ([]string, []EmailVerificationDetail, error) {
	var failedEmails []string
	var details []EmailVerificationDetail

	for _, email := range emails {
		_, err := sesClient.VerifyEmailIdentity(context.Background(), &ses.VerifyEmailIdentityInput{
			EmailAddress: aws.String(email),
		})

		detail := EmailVerificationDetail{
			Email:   email,
			Success: err == nil,
			Error:   err,
		}
		details = append(details, detail)

		if err != nil {
			failedEmails = append(failedEmails, email)
		}
	}

	return failedEmails, details, nil
}

// DisplayVerifyResult は検証結果を表示する
func DisplayVerifyResult(result *VerifyResult) {
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
}
