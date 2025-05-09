package internal

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

// ListCfnStacks はアクティブなCloudFormationスタック名一覧を返す
func ListCfnStacks(region, profile string) ([]string, error) {
	activeStatusStrs := []string{
		"CREATE_COMPLETE",
		"UPDATE_COMPLETE",
		"UPDATE_ROLLBACK_COMPLETE",
		"ROLLBACK_COMPLETE",
		"IMPORT_COMPLETE",
	}
	activeStatuses := make([]types.StackStatus, 0, len(activeStatusStrs))
	for _, s := range activeStatusStrs {
		activeStatuses = append(activeStatuses, types.StackStatus(s))
	}

	cfg, err := LoadAwsConfig(region, profile)
	if err != nil {
		return nil, err
	}

	client := cloudformation.NewFromConfig(cfg)

	// すべてのスタックを格納するスライス
	var allStackNames []string

	// ページネーション用のトークン
	var nextToken *string

	// すべてのページを取得するまでループ
	for {
		input := &cloudformation.ListStacksInput{
			StackStatusFilter: activeStatuses,
			NextToken:         nextToken,
		}

		resp, err := client.ListStacks(context.TODO(), input)
		if err != nil {
			return nil, fmt.Errorf("スタック一覧取得エラー: %w", err)
		}

		// 現在のページのスタック名をスライスに追加
		for _, summary := range resp.StackSummaries {
			allStackNames = append(allStackNames, aws.ToString(summary.StackName))
		}

		// 次のページがあるかチェック
		nextToken = resp.NextToken
		if nextToken == nil {
			// 次のページがなければループを抜ける
			break
		}
	}
	return allStackNames, nil
}

// getResourcesFromStack はCloudFormationスタックからS3バケットとECRリポジトリのリソース一覧を取得します
func getResourcesFromStack(opts CleanupOptions) ([]string, []string, error) {
	cfg, err := LoadAwsConfig(opts.Region, opts.Profile)
	if err != nil {
		return nil, nil, fmt.Errorf("AWS設定の読み込みエラー: %w", err)
	}

	// CloudFormationクライアントを作成
	cfnClient := cloudformation.NewFromConfig(cfg)

	// スタックリソース一覧の取得
	stackResources := []types.StackResourceSummary{}
	var nextToken *string

	for {
		resp, err := cfnClient.ListStackResources(context.TODO(), &cloudformation.ListStackResourcesInput{
			StackName: aws.String(opts.StackName),
			NextToken: nextToken,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("スタックリソース一覧取得エラー: %w", err)
		}

		stackResources = append(stackResources, resp.StackResourceSummaries...)

		if resp.NextToken == nil {
			break
		}
		nextToken = resp.NextToken
	}

	// S3バケットとECRリポジトリを抽出
	s3Resources := []string{}
	ecrResources := []string{}

	for _, resource := range stackResources {
		// リソースタイプに基づいて振り分け
		resourceType := *resource.ResourceType

		// S3バケット
		if resourceType == "AWS::S3::Bucket" && resource.PhysicalResourceId != nil {
			s3Resources = append(s3Resources, *resource.PhysicalResourceId)
			fmt.Printf("🔍 検出されたS3バケット: %s\n", *resource.PhysicalResourceId)
		}

		// ECRリポジトリ
		if resourceType == "AWS::ECR::Repository" && resource.PhysicalResourceId != nil {
			ecrResources = append(ecrResources, *resource.PhysicalResourceId)
			fmt.Printf("🔍 検出されたECRリポジトリ: %s\n", *resource.PhysicalResourceId)
		}
	}

	return s3Resources, ecrResources, nil
}
