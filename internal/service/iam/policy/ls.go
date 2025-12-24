package policy

import (
	"awstk/internal/service/common"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	sdkiam "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

// ListIamPolicies cmdから呼ばれるメイン関数（Get + Display）
func ListIamPolicies(client *sdkiam.Client, opts ListOptions) error {
	if client == nil {
		return fmt.Errorf("iam client is nil")
	}

	// Get: データ取得
	if opts.UnattachedOnly {
		items, err := getUnusedIamPolicies(client, opts)
		if err != nil {
			return err
		}
		// Display: 共通表示処理
		return common.DisplayList(items, "未使用のカスタマー管理ポリシー", unusedIamPoliciesToTableData, &common.DisplayOptions{
			ShowCount:    true,
			EmptyMessage: common.FormatEmptyMessage("未使用のカスタマー管理ポリシー"),
		})
	}

	items, err := getAllIamPolicies(client, opts)
	if err != nil {
		return err
	}
	// Display: 共通表示処理
	return common.DisplayList(items, "カスタマー管理ポリシー一覧", iamPoliciesToTableData, &common.DisplayOptions{ShowCount: true})
}

// getAllIamPolicies カスタマー管理ポリシー一覧を取得
func getAllIamPolicies(client *sdkiam.Client, opts ListOptions) ([]PolicyItem, error) {
	paginator := sdkiam.NewListPoliciesPaginator(client, &sdkiam.ListPoliciesInput{Scope: types.PolicyScopeTypeLocal})
	filters := common.RemoveDuplicates(opts.Exclude)

	var policies []PolicyItem
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.Background())
		if err != nil {
			return nil, common.FormatListError("カスタマー管理ポリシー", err)
		}
		for _, policy := range page.Policies {
			name := aws.ToString(policy.PolicyName)
			if matchesAnyFilter(name, filters) {
				continue
			}
			attachmentCount := int32(0)
			if policy.AttachmentCount != nil {
				attachmentCount = *policy.AttachmentCount
			}
			policies = append(policies, PolicyItem{
				Name:            name,
				Arn:             aws.ToString(policy.Arn),
				AttachmentCount: attachmentCount,
			})
		}
	}
	return policies, nil
}

// getUnusedIamPolicies 未アタッチのカスタマー管理ポリシー一覧を取得
func getUnusedIamPolicies(client *sdkiam.Client, opts ListOptions) ([]UnusedPolicy, error) {
	paginator := sdkiam.NewListPoliciesPaginator(client, &sdkiam.ListPoliciesInput{Scope: types.PolicyScopeTypeLocal})
	filters := common.RemoveDuplicates(opts.Exclude)

	var unusedPolicies []UnusedPolicy
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.Background())
		if err != nil {
			return nil, common.FormatListError("カスタマー管理ポリシー", err)
		}
		for _, policy := range page.Policies {
			name := aws.ToString(policy.PolicyName)
			if matchesAnyFilter(name, filters) {
				continue
			}
			if policy.AttachmentCount != nil && *policy.AttachmentCount == 0 {
				unusedPolicies = append(unusedPolicies, UnusedPolicy{Name: name, Arn: aws.ToString(policy.Arn), Note: "未アタッチ"})
			}
		}
	}
	return unusedPolicies, nil
}

// matchesAnyFilter 除外パターンに一致するか判定
func matchesAnyFilter(name string, filters []string) bool {
	if len(filters) == 0 {
		return false
	}
	for _, filter := range filters {
		if filter == "" {
			continue
		}
		if common.MatchesFilter(name, filter, false) {
			return true
		}
	}
	return false
}

// unusedIamPoliciesToTableData 未使用カスタマー管理ポリシー情報をテーブルデータに変換
func unusedIamPoliciesToTableData(items []UnusedPolicy) ([]common.TableColumn, [][]string) {
	columns := []common.TableColumn{{Header: "ポリシー名"}, {Header: "理由"}}
	rows := make([][]string, len(items))
	for index, policy := range items {
		rows[index] = []string{policy.Name, policy.Note}
	}
	return columns, rows
}

// iamPoliciesToTableData カスタマー管理ポリシー情報をテーブルデータに変換
func iamPoliciesToTableData(items []PolicyItem) ([]common.TableColumn, [][]string) {
	columns := []common.TableColumn{{Header: "ポリシー名"}, {Header: "attachments"}}
	rows := make([][]string, len(items))
	for index, policy := range items {
		rows[index] = []string{policy.Name, fmt.Sprintf("%d", policy.AttachmentCount)}
	}
	return columns, rows
}
