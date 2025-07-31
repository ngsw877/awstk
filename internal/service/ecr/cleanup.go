package ecr

import (
	"awstk/internal/service/common"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
)

// GetEcrRepositoriesByFilter はフィルターに一致するECRリポジトリ名の一覧を取得します
func GetEcrRepositoriesByFilter(ecrClient *ecr.Client, searchString string) ([]string, error) {
	// リポジトリ一覧を取得
	listReposInput := &ecr.DescribeRepositoriesInput{}
	foundRepos := []string{}

	// ページネーション対応
	for {
		listReposOutput, err := ecrClient.DescribeRepositories(context.Background(), listReposInput)
		if err != nil {
			return nil, fmt.Errorf("ecrリポジトリ一覧取得エラー: %w", err)
		}

		for _, repo := range listReposOutput.Repositories {
			if common.MatchesFilter(*repo.RepositoryName, searchString) {
				foundRepos = append(foundRepos, *repo.RepositoryName)
				fmt.Printf("🔍 検出されたECRリポジトリ: %s\n", *repo.RepositoryName)
			}
		}

		if listReposOutput.NextToken == nil {
			break
		}
		listReposInput.NextToken = listReposOutput.NextToken
	}

	return foundRepos, nil
}

// CleanupEcrRepositories は指定したECRリポジトリ一覧を削除します
func CleanupEcrRepositories(ecrClient *ecr.Client, repoNames []string) error {
	for _, repoName := range repoNames {
		fmt.Printf("リポジトリ %s を削除中...\n", repoName)

		// リポジトリの削除（強制削除フラグで内部のイメージも含めて削除）
		_, err := ecrClient.DeleteRepository(context.Background(), &ecr.DeleteRepositoryInput{
			RepositoryName: aws.String(repoName),
			Force:          true, // 強制削除（イメージが残っていても削除）
		})
		if err != nil {
			fmt.Printf("❌ リポジトリ %s の削除に失敗しました: %v\n", repoName, err)
			// このリポジトリの削除はスキップし、次のリポジトリへ
			continue
		}
		fmt.Printf("✅ リポジトリ %s を削除しました\n", repoName)
	}

	return nil
}

// CleanupRepositoriesByFilter はフィルターに基づいてリポジトリを削除する
func CleanupRepositoriesByFilter(ecrClient *ecr.Client, filter string) error {
	// フィルターに一致するリポジトリを取得
	repositories, err := GetEcrRepositoriesByFilter(ecrClient, filter)
	if err != nil {
		return fmt.Errorf("❌ ECRリポジトリ一覧取得エラー: %w", err)
	}

	if len(repositories) == 0 {
		fmt.Printf("フィルター '%s' に一致するECRリポジトリが見つかりませんでした\n", filter)
		return nil
	}

	// リポジトリを削除
	err = CleanupEcrRepositories(ecrClient, repositories)
	if err != nil {
		return fmt.Errorf("❌ ECRリポジトリ削除エラー: %w", err)
	}

	fmt.Println("✅ ECRリポジトリの削除が完了しました")
	return nil
}
