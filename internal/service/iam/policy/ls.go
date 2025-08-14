package policy

import (
	"awstk/internal/service/common"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	sdkiam "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

// ListOptions is defined in types.go

// List prints IAM customer managed policies. If UnattachedOnly is true, only unattached ones are printed.
func List(client *sdkiam.Client, opts ListOptions) error {
	if client == nil {
		return fmt.Errorf("iam client is nil")
	}

	if opts.UnattachedOnly {
		items, err := listUnusedPolicies(client, opts)
		if err != nil {
			return err
		}
		_ = common.DisplayList(items, "未使用のカスタマー管理ポリシー", toUnusedPoliciesTable, &common.DisplayOptions{
			ShowCount:    true,
			EmptyMessage: common.FormatEmptyMessage("未使用のカスタマー管理ポリシー"),
		})
		return nil
	}

	items, err := listAllPolicies(client, opts)
	if err != nil {
		return err
	}
	_ = common.DisplayList(items, "カスタマー管理ポリシー一覧", toPolicyItemsTable, &common.DisplayOptions{ShowCount: true})
	return nil
}

// ===== データ取得 =====

// PolicyItem and UnusedPolicy are defined in types.go

func listAllPolicies(client *sdkiam.Client, opts ListOptions) ([]PolicyItem, error) {
	paginator := sdkiam.NewListPoliciesPaginator(client, &sdkiam.ListPoliciesInput{Scope: types.PolicyScopeTypeLocal})
	var items []PolicyItem
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.Background())
		if err != nil {
			return nil, common.FormatListError("カスタマー管理ポリシー", err)
		}
		for _, p := range page.Policies {
			name := aws.ToString(p.PolicyName)
			if matchesAnyFilter(name, common.RemoveDuplicates(opts.Exclude)) {
				continue
			}
			ac := int32(0)
			if p.AttachmentCount != nil {
				ac = *p.AttachmentCount
			}
			items = append(items, PolicyItem{Name: name, Arn: aws.ToString(p.Arn), AttachmentCount: ac})
		}
	}
	return items, nil
}

func listUnusedPolicies(client *sdkiam.Client, opts ListOptions) ([]UnusedPolicy, error) {
	paginator := sdkiam.NewListPoliciesPaginator(client, &sdkiam.ListPoliciesInput{Scope: types.PolicyScopeTypeLocal})
	var out []UnusedPolicy
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.Background())
		if err != nil {
			return nil, common.FormatListError("カスタマー管理ポリシー", err)
		}
		for _, pol := range page.Policies {
			name := aws.ToString(pol.PolicyName)
			if matchesAnyFilter(name, common.RemoveDuplicates(opts.Exclude)) {
				continue
			}
			if pol.AttachmentCount != nil && *pol.AttachmentCount == 0 {
				out = append(out, UnusedPolicy{Name: name, Arn: aws.ToString(pol.Arn), Note: "未アタッチ"})
			}
		}
	}
	return out, nil
}

// ===== ヘルパー/表示 =====

func matchesAnyFilter(name string, filters []string) bool {
	if len(filters) == 0 {
		return false
	}
	for _, f := range filters {
		if f == "" {
			continue
		}
		if common.MatchesFilter(name, f) {
			return true
		}
	}
	return false
}

func toUnusedPoliciesTable(items []UnusedPolicy) ([]common.TableColumn, [][]string) {
	cols := []common.TableColumn{{Header: "ポリシー名"}, {Header: "理由"}}
	rows := make([][]string, len(items))
	for i, p := range items {
		rows[i] = []string{p.Name, p.Note}
	}
	return cols, rows
}

func toPolicyItemsTable(items []PolicyItem) ([]common.TableColumn, [][]string) {
	cols := []common.TableColumn{{Header: "ポリシー名"}, {Header: "attachments"}}
	rows := make([][]string, len(items))
	for i, p := range items {
		rows[i] = []string{p.Name, fmt.Sprintf("%d", p.AttachmentCount)}
	}
	return cols, rows
}
