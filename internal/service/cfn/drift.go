package cfn

import (
	"awstk/internal/service/common"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

// driftStatusString はドリフト状態を文字列で返します
func driftStatusString(status types.StackDriftStatus) string {
	switch status {
	case types.StackDriftStatusDrifted:
		return "ドリフトあり"
	case types.StackDriftStatusInSync:
		return "同期中"
	case types.StackDriftStatusNotChecked:
		return "未確認"
	default:
		return string(status)
	}
}

// isDriftDetectable はスタックがドリフト検出可能な状態かを判定します
func isDriftDetectable(status types.StackStatus) bool {
	return status == types.StackStatusCreateComplete ||
		status == types.StackStatusUpdateComplete ||
		status == types.StackStatusUpdateRollbackComplete
}

// getDriftDetectableStatuses はドリフト検出可能なステータスのリストを返します
func getDriftDetectableStatuses() []types.StackStatus {
	return []types.StackStatus{
		types.StackStatusCreateComplete,
		types.StackStatusUpdateComplete,
		types.StackStatusUpdateRollbackComplete,
	}
}

// DetectDrift は指定した条件に一致するスタックのドリフト検出を実行します
func DetectDrift(cfnClient *cloudformation.Client, opts DriftOptions) error {
	// 対象のスタックを検索
	stacks, err := findStacksForDrift(cfnClient, opts)
	if err != nil {
		return err
	}

	if len(stacks) == 0 {
		fmt.Println("対象のスタックが見つかりませんでした")
		return nil
	}

	// 検出対象のスタック一覧を表示
	fmt.Printf("🔍 ドリフト検出を実行するスタック:\n")
	for _, stack := range stacks {
		fmt.Printf("  - %s\n", aws.ToString(stack.StackName))
	}
	fmt.Printf("\n合計 %d 個のスタックでドリフト検出を実行します\n", len(stacks))

	// ドリフト検出を実行
	fmt.Println("\nドリフト検出を開始します...")
	detectionIds := make(map[string]string) // stackName -> detectionId

	for _, stack := range stacks {
		stackName := aws.ToString(stack.StackName)
		fmt.Printf("スタック %s のドリフト検出を開始中...", stackName)

		output, err := cfnClient.DetectStackDrift(context.Background(), &cloudformation.DetectStackDriftInput{
			StackName: aws.String(stackName),
		})
		if err != nil {
			fmt.Printf("\n❌ スタック %s のドリフト検出開始に失敗しました: %v\n", stackName, err)
			continue
		}

		detectionIds[stackName] = aws.ToString(output.StackDriftDetectionId)
		fmt.Printf(" ✅ (検出ID: %s)\n", aws.ToString(output.StackDriftDetectionId))
	}

	if len(detectionIds) > 0 {
		fmt.Printf("\n✅ %d 個のスタックでドリフト検出を開始しました\n", len(detectionIds))
		fmt.Println("ℹ️  検出結果は 'awstk cfn drift-status' コマンドで確認できます")
	}

	return nil
}

// ShowDriftStatus は指定した条件に一致するスタックのドリフト状態を表示します
func ShowDriftStatus(cfnClient *cloudformation.Client, opts DriftStatusOptions) error {
	// 対象のスタックを検索
	stacks, err := findStacksForDrift(cfnClient, DriftOptions{
		Stacks: opts.Stacks,
		Filter: opts.Filter,
		All:    opts.All,
		Exact:  opts.Exact,
	})
	if err != nil {
		return err
	}

	if len(stacks) == 0 {
		fmt.Println("対象のスタックが見つかりませんでした")
		return nil
	}

	// ドリフト状態を確認
	fmt.Println("🔍 スタックのドリフト状態を確認中...")
	driftedCount := 0
	notCheckedCount := 0

	for _, stack := range stacks {
		stackName := aws.ToString(stack.StackName)

		// スタックの詳細情報を取得（ドリフト情報を含む）
		describeOutput, err := cfnClient.DescribeStacks(context.Background(), &cloudformation.DescribeStacksInput{
			StackName: aws.String(stackName),
		})
		if err != nil {
			fmt.Printf("❌ スタック %s の情報取得に失敗しました: %v\n", stackName, err)
			continue
		}

		if len(describeOutput.Stacks) == 0 {
			continue
		}

		stackInfo := describeOutput.Stacks[0]
		driftInfo := stackInfo.DriftInformation
		if driftInfo == nil {
			continue
		}

		// ドリフト状態を確認
		driftStatus := driftInfo.StackDriftStatus
		switch driftStatus {
		case types.StackDriftStatusNotChecked:
			notCheckedCount++
		case types.StackDriftStatusDrifted:
			driftedCount++
		}

		// --drifted-onlyが指定されている場合、ドリフトしていないスタックはスキップ
		if opts.DriftedOnly && driftStatus != types.StackDriftStatusDrifted {
			continue
		}

		// ドリフト状態を表示
		statusStr := driftStatusString(driftStatus)
		statusIcon := "✅"
		switch driftStatus {
		case types.StackDriftStatusDrifted:
			statusIcon = "⚠️ "
		case types.StackDriftStatusNotChecked:
			statusIcon = "❓"
		}

		fmt.Printf("%s %s: %s", statusIcon, stackName, statusStr)

		// 最終チェック時刻を表示
		if driftInfo.LastCheckTimestamp != nil {
			checkTime := aws.ToTime(driftInfo.LastCheckTimestamp)
			fmt.Printf(" (最終チェック: %s)", checkTime.Format("2006-01-02 15:04:05"))
		}
		fmt.Println()
	}

	// サマリーを表示
	fmt.Printf("\n📊 サマリー:\n")
	fmt.Printf("  - 合計: %d スタック\n", len(stacks))
	fmt.Printf("  - ドリフトあり: %d スタック\n", driftedCount)
	fmt.Printf("  - 未確認: %d スタック\n", notCheckedCount)

	if notCheckedCount > 0 {
		fmt.Println("\nℹ️  未確認のスタックがあります。'awstk cfn drift-detect' でドリフト検出を実行してください")
	}

	return nil
}

// findStacksForDrift はドリフト検出対象のスタックを検索します
func findStacksForDrift(cfnClient *cloudformation.Client, opts DriftOptions) ([]types.Stack, error) {
	var allStacks []types.Stack

	// スタック名が指定されている場合
	if len(opts.Stacks) > 0 {
		for _, stackName := range opts.Stacks {
			if stackName == "" {
				continue
			}

			// スタックの詳細情報を取得
			describeOutput, err := cfnClient.DescribeStacks(context.Background(), &cloudformation.DescribeStacksInput{
				StackName: aws.String(stackName),
			})
			if err != nil {
				fmt.Printf("⚠️  スタック %s が見つかりません: %v\n", stackName, err)
				continue
			}
			if len(describeOutput.Stacks) > 0 {
				stack := describeOutput.Stacks[0]
				// ドリフト検出可能なステータスかチェック
				if isDriftDetectable(stack.StackStatus) {
					allStacks = append(allStacks, stack)
				} else {
					fmt.Printf("⚠️  スタック %s はドリフト検出できない状態です (Status: %s)\n", stackName, stack.StackStatus)
				}
			}
		}
		return allStacks, nil
	}

	// --filterまたは--allの場合
	var nextToken *string
	for {
		input := &cloudformation.ListStacksInput{
			NextToken: nextToken,
			// ドリフト検出可能なステータスのみ
			StackStatusFilter: getDriftDetectableStatuses(),
		}

		output, err := cfnClient.ListStacks(context.Background(), input)
		if err != nil {
			return nil, fmt.Errorf("スタック一覧の取得に失敗しました: %w", err)
		}

		// フィルター処理
		for _, summary := range output.StackSummaries {
			if opts.All || (opts.Filter != "" && common.MatchesFilter(aws.ToString(summary.StackName), opts.Filter, opts.Exact)) {
				// スタックの詳細情報を取得
				describeOutput, err := cfnClient.DescribeStacks(context.Background(), &cloudformation.DescribeStacksInput{
					StackName: summary.StackName,
				})
				if err != nil {
					continue
				}
				if len(describeOutput.Stacks) > 0 {
					allStacks = append(allStacks, describeOutput.Stacks[0])
				}
			}
		}

		if output.NextToken == nil {
			break
		}
		nextToken = output.NextToken
	}

	return allStacks, nil
}
