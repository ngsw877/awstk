package cfn

import (
	"context"
	"fmt"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

// ListCfnStacks はCloudFormationスタック一覧を返す
// activeOnly が true の場合はアクティブなスタックのみを取得する
func ListCfnStacks(cfnClient *cloudformation.Client, activeOnly bool) ([]CfnStack, error) {
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

	var stacks []CfnStack
	var nextToken *string

	for {
		input := &cloudformation.ListStacksInput{NextToken: nextToken}
		if activeOnly {
			input.StackStatusFilter = activeStatuses
		}

		resp, err := cfnClient.ListStacks(context.Background(), input)
		if err != nil {
			return nil, fmt.Errorf("スタック一覧取得エラー: %w", err)
		}

		for _, summary := range resp.StackSummaries {
			stacks = append(stacks, CfnStack{
				Name:   awssdk.ToString(summary.StackName),
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
