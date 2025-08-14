package iam

import (
	"awstk/internal/service/common"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	sdkiam "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

// ListUnused returns unused IAM roles and policies based on the provided options.
func ListUnused(client *sdkiam.Client, opts ListOptions) (*UnusedResult, error) {
	res := &UnusedResult{}

	if opts.Roles {
		roles, err := listUnusedRoles(client, opts)
		if err != nil {
			return nil, err
		}
		res.Roles = roles
	}

	if opts.Policies {
		policies, err := listUnusedPolicies(client, opts)
		if err != nil {
			return nil, err
		}
		res.Policies = policies
	}

	return res, nil
}

// ListAll returns all roles and/or policies without unused filtering.
// When opts.Details is true, it enriches role items with last-used timestamps via GetRole (parallelized).
func ListAll(client *sdkiam.Client, opts ListOptions) (*AllResult, error) {
	result := &AllResult{}

	if opts.Roles {
		rolesPaginator := sdkiam.NewListRolesPaginator(client, &sdkiam.ListRolesInput{})
		var roles []types.Role
		for rolesPaginator.HasMorePages() {
			page, err := rolesPaginator.NextPage(context.Background())
			if err != nil {
				return nil, fmt.Errorf("ListRoles 失敗: %w", err)
			}
			roles = append(roles, page.Roles...)
		}

		// Map to RoleItem
		roleItems := make([]RoleItem, 0, len(roles))
		for _, r := range roles {
			roleItems = append(roleItems, RoleItem{
				Name:            aws.ToString(r.RoleName),
				Arn:             aws.ToString(r.Arn),
				IsServiceLinked: isServiceLinkedRole(r) || isServiceLinkedRoleName(aws.ToString(r.RoleName)),
			})
		}

		// 常に最終使用日時を付与
		exec := common.NewParallelExecutor(8)
		var mu sync.Mutex
		for i := range roleItems {
			idx := i
			exec.Execute(func() {
				outRole, err := client.GetRole(context.Background(), &sdkiam.GetRoleInput{RoleName: aws.String(roleItems[idx].Name)})
				if err != nil {
					return
				}
				var last *time.Time
				if outRole.Role.RoleLastUsed != nil && outRole.Role.RoleLastUsed.LastUsedDate != nil {
					t := *outRole.Role.RoleLastUsed.LastUsedDate
					last = &t
				}
				mu.Lock()
				roleItems[idx].LastUsed = last
				mu.Unlock()
			})
		}
		exec.Wait()
		result.Roles = roleItems
	}

	if opts.Policies {
		policiesPaginator := sdkiam.NewListPoliciesPaginator(client, &sdkiam.ListPoliciesInput{
			Scope: types.PolicyScopeTypeLocal,
		})
		var items []PolicyItem
		for policiesPaginator.HasMorePages() {
			page, err := policiesPaginator.NextPage(context.Background())
			if err != nil {
				return nil, fmt.Errorf("ListPolicies 失敗: %w", err)
			}
			for _, p := range page.Policies {
				ac := int32(0)
				if p.AttachmentCount != nil {
					ac = *p.AttachmentCount
				}
				items = append(items, PolicyItem{
					Name:            aws.ToString(p.PolicyName),
					Arn:             aws.ToString(p.Arn),
					AttachmentCount: ac,
				})
			}
		}
		result.Policies = items
	}

	return result, nil
}

func listUnusedRoles(client *sdkiam.Client, opts ListOptions) ([]UnusedRole, error) {
	// まずは全ロール一覧
	rolesPaginator := sdkiam.NewListRolesPaginator(client, &sdkiam.ListRolesInput{})
	var allRoles []types.Role
	for rolesPaginator.HasMorePages() {
		rolesPage, err := rolesPaginator.NextPage(context.Background())
		if err != nil {
			return nil, fmt.Errorf("ListRoles 失敗: %w", err)
		}
		allRoles = append(allRoles, rolesPage.Roles...)
	}

	// フィルタ（サービスリンク/除外）
	var targetRoleNames []string
	for _, r := range allRoles {
		name := aws.ToString(r.RoleName)
		if isServiceLinkedRole(r) || isServiceLinkedRoleName(name) {
			continue
		}
		if matchExclude(opts.Exclude, name) {
			continue
		}
		targetRoleNames = append(targetRoleNames, name)
	}

	// 詳細取得（GetRoleで RoleLastUsed を取得）
	cutoff := time.Now().AddDate(0, 0, -opts.UnusedDays)
	exec := common.NewParallelExecutor(8)
	var mu sync.Mutex
	var out []UnusedRole

	for _, roleName := range targetRoleNames {
		rn := roleName
		exec.Execute(func() {
			outRole, err := client.GetRole(context.Background(), &sdkiam.GetRoleInput{RoleName: aws.String(rn)})
			if err != nil {
				// エラー時はスキップ（一覧表示なので厳格に止めない）
				return
			}
			var last *time.Time
			if outRole.Role.RoleLastUsed != nil && outRole.Role.RoleLastUsed.LastUsedDate != nil {
				t := *outRole.Role.RoleLastUsed.LastUsedDate
				last = &t
			}
			if last == nil || last.Before(cutoff) {
				mu.Lock()
				out = append(out, UnusedRole{
					Name:     rn,
					Arn:      aws.ToString(outRole.Role.Arn),
					LastUsed: last,
				})
				mu.Unlock()
			}
		})
	}
	exec.Wait()

	return out, nil
}

func listUnusedPolicies(client *sdkiam.Client, opts ListOptions) ([]UnusedPolicy, error) {
	var out []UnusedPolicy
	policiesPaginator := sdkiam.NewListPoliciesPaginator(client, &sdkiam.ListPoliciesInput{
		Scope: types.PolicyScopeTypeLocal, // カスタマー管理のみ
	})

	for policiesPaginator.HasMorePages() {
		policiesPage, err := policiesPaginator.NextPage(context.Background())
		if err != nil {
			return nil, fmt.Errorf("ListPolicies 失敗: %w", err)
		}
		for _, pol := range policiesPage.Policies {
			name := aws.ToString(pol.PolicyName)
			if matchExclude(opts.Exclude, name) {
				continue
			}
			// 未アタッチのみを未使用候補に
			if pol.AttachmentCount != nil && *pol.AttachmentCount == 0 {
				out = append(out, UnusedPolicy{
					Name: name,
					Arn:  aws.ToString(pol.Arn),
					Note: "未アタッチ",
				})
			}
		}
	}
	return out, nil
}

func isServiceLinkedRole(r types.Role) bool {
	// サービスリンクロール: Path が "/aws-service-role/" で判定
	return aws.ToString(r.Path) == "/aws-service-role/"
}

func isServiceLinkedRoleName(name string) bool {
	// 名前からの補助的な判定
	return strings.HasPrefix(name, "AWSServiceRoleFor")
}

func matchExclude(ex []string, name string) bool {
	if len(ex) == 0 {
		return false
	}
	for _, pat := range ex {
		if pat == "" {
			continue
		}
		if strings.Contains(name, pat) {
			return true
		}
	}
	return false
}

// DisplayUnused prints unused roles and policies using common list utilities.
func DisplayUnused(result *UnusedResult, opts ListOptions) {
	if opts.Roles {
		if len(result.Roles) == 0 {
			fmt.Println(common.FormatEmptyMessage("未使用のIAMロール"))
		} else {
			items := make([]string, 0, len(result.Roles))
			for _, r := range result.Roles {
				last := "never"
				if r.LastUsed != nil {
					last = r.LastUsed.In(time.Local).Format("2006-01-02 15:04:05 MST")
				}
				items = append(items, fmt.Sprintf("%s (last-used: %s)", r.Name, last))
			}
			common.PrintNumberedList(common.ListOutput{
				Title:        "未使用のIAMロール",
				Items:        items,
				ResourceName: "ロール",
			})
		}
	}

	if opts.Policies {
		if len(result.Policies) == 0 {
			fmt.Println(common.FormatEmptyMessage("未使用のカスタマー管理ポリシー"))
		} else {
			items := make([]string, 0, len(result.Policies))
			for _, p := range result.Policies {
				items = append(items, fmt.Sprintf("%s (%s)", p.Name, p.Note))
			}
			common.PrintNumberedList(common.ListOutput{
				Title:        "未使用のカスタマー管理ポリシー",
				Items:        items,
				ResourceName: "ポリシー",
			})
		}
	}
}

// DisplayAll prints all roles and/or policies.
func DisplayAll(result *AllResult, opts ListOptions) {
	if opts.Roles {
		items := make([]string, 0, len(result.Roles))
		for _, r := range result.Roles {
			last := "never"
			if r.LastUsed != nil {
				last = r.LastUsed.In(time.Local).Format("2006-01-02 15:04:05 MST")
			}
			sl := ""
			if r.IsServiceLinked {
				sl = " [service-linked]"
			}
			items = append(items, fmt.Sprintf("%s (last-used: %s)%s", r.Name, last, sl))
		}
		common.PrintNumberedList(common.ListOutput{
			Title:        "IAMロール一覧",
			Items:        items,
			ResourceName: "ロール",
		})
	}

	if opts.Policies {
		items := make([]string, 0, len(result.Policies))
		for _, p := range result.Policies {
			note := fmt.Sprintf("attachments: %d", p.AttachmentCount)
			items = append(items, fmt.Sprintf("%s (%s)", p.Name, note))
		}
		common.PrintNumberedList(common.ListOutput{
			Title:        "カスタマー管理ポリシー一覧",
			Items:        items,
			ResourceName: "ポリシー",
		})
	}
}
