package ses

import (
	"awstk/internal/service/common"
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

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
	emails = removeDuplicateEmails(emails)
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
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("⚠️  ファイルのクローズに失敗: %v\n", err)
		}
	}()

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

// removeDuplicateEmails はメールアドレスの重複を除去（大文字小文字を無視）
func removeDuplicateEmails(emails []string) []string {
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
	if len(emails) == 0 {
		return nil, nil, nil
	}

	// 並列実行数を設定（最大10並列）
	maxWorkers := 10
	if len(emails) < maxWorkers {
		maxWorkers = len(emails)
	}

	executor := common.NewParallelExecutor(maxWorkers)
	details := make([]EmailVerificationDetail, len(emails))
	detailsMutex := &sync.Mutex{}
	failedEmailsMutex := &sync.Mutex{}
	var failedEmails []string

	fmt.Printf("🚀 %d件のメールアドレスを最大%d並列で検証します...\n\n", len(emails), maxWorkers)

	for i, email := range emails {
		idx := i
		emailAddr := email
		executor.Execute(func() {
			_, err := sesClient.VerifyEmailIdentity(context.Background(), &ses.VerifyEmailIdentityInput{
				EmailAddress: aws.String(emailAddr),
			})

			detail := EmailVerificationDetail{
				Email:   emailAddr,
				Success: err == nil,
				Error:   err,
			}

			detailsMutex.Lock()
			details[idx] = detail
			detailsMutex.Unlock()

			if err != nil {
				failedEmailsMutex.Lock()
				failedEmails = append(failedEmails, emailAddr)
				failedEmailsMutex.Unlock()
			}
		})
	}

	executor.Wait()

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
