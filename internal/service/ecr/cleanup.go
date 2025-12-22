package ecr

import (
	"awstk/internal/service/common"
	"context"
	"fmt"
	"sync"

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
func CleanupEcrRepositories(ecrClient *ecr.Client, repoNames []string) common.CleanupResult {
	result := common.CleanupResult{
		ResourceType: "ECRリポジトリ",
		Deleted:      []string{},
		Failed:       []string{},
	}

	if len(repoNames) == 0 {
		return result
	}

	// 並列実行数を設定（最大10並列）
	maxWorkers := 10
	if len(repoNames) < maxWorkers {
		maxWorkers = len(repoNames)
	}

	executor := common.NewParallelExecutor(maxWorkers)
	results := make([]common.ProcessResult, len(repoNames))
	resultsMutex := &sync.Mutex{}

	fmt.Printf("🚀 %d個のリポジトリを最大%d並列で削除します...\n\n", len(repoNames), maxWorkers)

	for i, repoName := range repoNames {
		idx := i
		repo := repoName
		executor.Execute(func() {
			fmt.Printf("リポジトリ %s を削除中...\n", repo)

			// リポジトリの削除（強制削除フラグで内部のイメージも含めて削除）
			_, err := ecrClient.DeleteRepository(context.Background(), &ecr.DeleteRepositoryInput{
				RepositoryName: aws.String(repo),
				Force:          true, // 強制削除（イメージが残っていても削除）
			})

			resultsMutex.Lock()
			if err != nil {
				fmt.Printf("❌ リポジトリ %s の削除に失敗しました: %v\n", repo, err)
				results[idx] = common.ProcessResult{Item: repo, Success: false, Error: err}
			} else {
				fmt.Printf("✅ リポジトリ %s を削除しました\n", repo)
				results[idx] = common.ProcessResult{Item: repo, Success: true}
			}
			resultsMutex.Unlock()
		})
	}

	executor.Wait()

	// 結果の集計
	successCount, failCount := common.CollectResults(results)
	fmt.Printf("\n✅ 削除完了: 成功 %d個, 失敗 %d個\n", successCount, failCount)

	return common.CollectCleanupResult("ECRリポジトリ", results)
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
	result := CleanupEcrRepositories(ecrClient, repositories)
	if len(result.Failed) > 0 {
		return fmt.Errorf("❌ %d個のECRリポジトリの削除に失敗しました", len(result.Failed))
	}

	fmt.Println("✅ ECRリポジトリの削除が完了しました")
	return nil
}
