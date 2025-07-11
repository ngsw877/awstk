package ecr

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
)

// ListEcrRepositories はECRリポジトリの一覧を取得する関数
func ListEcrRepositories(client *ecr.Client) ([]RepositoryInfo, error) {
	var repositories []RepositoryInfo
	var nextToken *string

	for {
		input := &ecr.DescribeRepositoriesInput{
			NextToken: nextToken,
		}

		result, err := client.DescribeRepositories(context.Background(), input)
		if err != nil {
			return nil, fmt.Errorf("リポジトリ一覧取得エラー: %w", err)
		}

		// 各リポジトリの基本情報を取得（追加APIリクエストなし）
		for _, repo := range result.Repositories {
			info := RepositoryInfo{
				RepositoryName: aws.ToString(repo.RepositoryName),
				RepositoryUri:  aws.ToString(repo.RepositoryUri),
				CreatedAt:      repo.CreatedAt,
			}

			repositories = append(repositories, info)
		}

		if result.NextToken == nil {
			break
		}
		nextToken = result.NextToken
	}

	return repositories, nil
}

// GetRepositoryImageCount はリポジトリ内のイメージ数を取得する関数
func GetRepositoryImageCount(client *ecr.Client, repoName string) (int, error) {
	input := &ecr.DescribeImagesInput{
		RepositoryName: aws.String(repoName),
		MaxResults:     aws.Int32(1), // カウントだけ必要なので最小限に
	}

	result, err := client.DescribeImages(context.Background(), input)
	if err != nil {
		return 0, err
	}

	// NextTokenがある場合は、全件取得してカウント
	if result.NextToken != nil {
		imageDetails, err := getRepositoryImageDetails(client, repoName)
		if err != nil {
			return 0, err
		}
		return len(imageDetails), nil
	}

	return len(result.ImageDetails), nil
}

// getRepositoryImageDetails はリポジトリ内のイメージ詳細を取得する関数
func getRepositoryImageDetails(client *ecr.Client, repoName string) ([]types.ImageDetail, error) {
	var imageDetails []types.ImageDetail
	var nextToken *string

	for {
		input := &ecr.DescribeImagesInput{
			RepositoryName: aws.String(repoName),
			NextToken:      nextToken,
		}

		result, err := client.DescribeImages(context.Background(), input)
		if err != nil {
			return nil, err
		}

		imageDetails = append(imageDetails, result.ImageDetails...)

		if result.NextToken == nil {
			break
		}
		nextToken = result.NextToken
	}

	return imageDetails, nil
}

// FilterEmptyRepositories は空のリポジトリのみを返す関数
func FilterEmptyRepositories(client *ecr.Client, repositories []RepositoryInfo) ([]RepositoryInfo, error) {
	var emptyRepos []RepositoryInfo

	for _, repo := range repositories {
		imageCount, err := GetRepositoryImageCount(client, repo.RepositoryName)
		if err != nil {
			return nil, fmt.Errorf("イメージ数取得エラー (%s): %w", repo.RepositoryName, err)
		}
		
		repo.ImageCount = imageCount
		if imageCount == 0 {
			emptyRepos = append(emptyRepos, repo)
		}
	}

	return emptyRepos, nil
}

// FilterNoLifecycleRepositories はライフサイクルポリシーが未設定のリポジトリのみを返す関数
func FilterNoLifecycleRepositories(client *ecr.Client, repositories []RepositoryInfo) ([]RepositoryInfo, error) {
	var noLifecycleRepos []RepositoryInfo

	for _, repo := range repositories {
		hasLifecycle, err := CheckLifecyclePolicy(client, repo.RepositoryName)
		if err != nil {
			return nil, fmt.Errorf("ライフサイクルポリシー確認エラー (%s): %w", repo.RepositoryName, err)
		}

		repo.HasLifecycle = hasLifecycle
		if !hasLifecycle {
			noLifecycleRepos = append(noLifecycleRepos, repo)
		}
	}

	return noLifecycleRepos, nil
}

// CheckLifecyclePolicy はリポジトリにライフサイクルポリシーが設定されているか確認する関数
func CheckLifecyclePolicy(client *ecr.Client, repoName string) (bool, error) {
	_, err := client.GetLifecyclePolicy(context.Background(), &ecr.GetLifecyclePolicyInput{
		RepositoryName: aws.String(repoName),
	})

	if err != nil {
		var notFoundErr *types.LifecyclePolicyNotFoundException
		if errors.As(err, &notFoundErr) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// EnrichRepositoryDetails はリポジトリの詳細情報を取得して追加する関数
func EnrichRepositoryDetails(client *ecr.Client, repo *RepositoryInfo) error {
	// イメージ詳細を取得
	imageDetails, err := getRepositoryImageDetails(client, repo.RepositoryName)
	if err != nil {
		return fmt.Errorf("イメージ詳細取得エラー: %w", err)
	}
	
	repo.ImageCount = len(imageDetails)
	repo.SizeInBytes = 0
	for _, image := range imageDetails {
		if image.ImageSizeInBytes != nil {
			repo.SizeInBytes += *image.ImageSizeInBytes
		}
	}
	
	// ライフサイクルポリシーの有無を確認
	repo.HasLifecycle, _ = CheckLifecyclePolicy(client, repo.RepositoryName)
	
	return nil
}

// DisplayRepositoryDetails はリポジトリの詳細情報を表示する関数
func DisplayRepositoryDetails(repo RepositoryInfo) {
	fmt.Printf("  - %s\n", repo.RepositoryName)
	fmt.Printf("    URI: %s\n", repo.RepositoryUri)
	fmt.Printf("    イメージ数: %d個\n", repo.ImageCount)
	fmt.Printf("    サイズ: %s\n", formatBytes(repo.SizeInBytes))
	
	if repo.CreatedAt != nil {
		fmt.Printf("    作成日: %s\n", repo.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	
	lifecycleStatus := "設定済み"
	if !repo.HasLifecycle {
		lifecycleStatus = "未設定"
	}
	fmt.Printf("    ライフサイクルポリシー: %s\n", lifecycleStatus)
}

// formatBytes はバイト数を人間が読みやすい形式に変換する関数
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}