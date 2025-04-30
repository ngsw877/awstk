package cfn

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

// ListStacksはアクティブなCloudFormationスタック名一覧を返す
func ListStacks(region, profile string) ([]string, error) {
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

	if profile != "" {
		os.Setenv("AWS_PROFILE", profile)
	}
	if region != "" {
		os.Setenv("AWS_REGION", region)
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	client := cloudformation.NewFromConfig(cfg)

	input := &cloudformation.ListStacksInput{
		StackStatusFilter: activeStatuses,
	}

	resp, err := client.ListStacks(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	stacks := make([]string, 0, len(resp.StackSummaries))
	for _, summary := range resp.StackSummaries {
		stacks = append(stacks, aws.ToString(summary.StackName))
	}
	return stacks, nil
}
