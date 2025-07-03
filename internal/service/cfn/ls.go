package cfn

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

// ListCfnStacks はCloudFormationスタック一覧を返す
// showAll が true の場合は全てのステータスのスタックを取得する
// showAll が false の場合はアクティブなスタックのみを取得する
func ListCfnStacks(cfnClient *cloudformation.Client, showAll bool) ([]CfnStack, error) {
	activeStatuses := []types.StackStatus{
		types.StackStatusCreateComplete,
		types.StackStatusUpdateComplete,
		types.StackStatusUpdateRollbackComplete,
		types.StackStatusRollbackComplete,
		types.StackStatusImportComplete,
	}

	var stacks []CfnStack
	var nextToken *string

	for {
		input := &cloudformation.ListStacksInput{NextToken: nextToken}
		if !showAll {
			input.StackStatusFilter = activeStatuses
		}

		resp, err := cfnClient.ListStacks(context.Background(), input)
		if err != nil {
			return nil, fmt.Errorf("スタック一覧取得エラー: %w", err)
		}

		for _, summary := range resp.StackSummaries {
			stacks = append(stacks, CfnStack{
				Name:   aws.ToString(summary.StackName),
				Status: string(summary.StackStatus),
			})
		}

		nextToken = resp.NextToken
		if nextToken == nil {
			break
		}
	}

	return stacks, nil
}
