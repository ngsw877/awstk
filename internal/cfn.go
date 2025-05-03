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
