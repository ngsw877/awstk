package ecr

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	ecrtypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"
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
			return nil, fmt.Errorf("ECRリポジトリ一覧取得エラー: %w", err)
		}

		for _, repo := range listReposOutput.Repositories {
			if strings.Contains(*repo.RepositoryName, searchString) {
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
		fmt.Printf("リポジトリ %s を空にして削除中...\n", repoName)

		// リポジトリ内のイメージをすべて削除 (ページネーション対応)
		listImagesInput := &ecr.ListImagesInput{
			RepositoryName: aws.String(repoName),
		}
		imageIdsToDelete := []ecrtypes.ImageIdentifier{}

		for {
			listImagesOutput, err := ecrClient.ListImages(context.Background(), listImagesInput)
			if err != nil {
				// イメージリスト取得エラーはこのリポジトリをスキップ
				fmt.Printf("❌ リポジトリ %s のイメージ一覧取得エラー: %v\n", repoName, err)
				break // このリポジトリの処理を中断
			}

			imageIdsToDelete = append(imageIdsToDelete, listImagesOutput.ImageIds...)

			if listImagesOutput.NextToken == nil {
				break
			}
			listImagesInput.NextToken = listImagesOutput.NextToken
		}

		// イメージ削除対象がなければスキップ
		if len(imageIdsToDelete) > 0 {
			// イメージを一括削除 (最大100個ずつ)
			chunkSize := 100
			for i := 0; i < len(imageIdsToDelete); i += chunkSize {
				end := i + chunkSize
				if end > len(imageIdsToDelete) {
					end = len(imageIdsToDelete)
				}
				batch := imageIdsToDelete[i:end]

				fmt.Printf("  %d件のイメージを削除中...\n", len(batch))
				_, err := ecrClient.BatchDeleteImage(context.Background(), &ecr.BatchDeleteImageInput{
					RepositoryName: aws.String(repoName),
					ImageIds:       batch,
				})
				if err != nil {
					fmt.Printf("❌ リポジトリ %s のイメージ一括削除エラー: %v\n", repoName, err)
					// イメージ削除エラーでもリポジトリ削除は試みる
				}
			}
		} else {
			fmt.Println("  削除するイメージがありません。")
		}

		// リポジトリの削除
		fmt.Printf("  リポジトリ削除中: %s\n", repoName)
		_, err := ecrClient.DeleteRepository(context.Background(), &ecr.DeleteRepositoryInput{
			RepositoryName: aws.String(repoName),
			Force:          true, // 強制削除
		})
		if err != nil {
			fmt.Printf("❌ リポジトリ %s の削除に失敗しました: %v\n", repoName, err)
			// このリポジトリの削除はスキップし、次のリポジトリへ
			continue
		}
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
