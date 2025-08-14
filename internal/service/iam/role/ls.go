package role

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

// ListOptions is defined in types.go

// List prints IAM roles. If UnusedDays > 0, it prints only unused roles.
func List(client *sdkiam.Client, opts ListOptions) error {
	if client == nil {
		return fmt.Errorf("iam client is nil")
	}

	if opts.UnusedDays > 0 {
		items, err := listUnusedRoles(client, opts)
		if err != nil {
			return err
		}
		_ = common.DisplayList(items, "未使用のIAMロール", toUnusedRolesTable, &common.DisplayOptions{
			ShowCount:    true,
			EmptyMessage: common.FormatEmptyMessage("未使用のIAMロール"),
		})
		return nil
	}

	items, err := listAllRoles(client, opts)
	if err != nil {
		return err
	}
	_ = common.DisplayList(items, "IAMロール一覧", toRoleItemsTable, &common.DisplayOptions{ShowCount: true})
	return nil
}

// ===== データ取得 =====

// RoleItem and UnusedRole are defined in types.go

func listAllRoles(client *sdkiam.Client, opts ListOptions) ([]RoleItem, error) {
	paginator := sdkiam.NewListRolesPaginator(client, &sdkiam.ListRolesInput{})
	var roles []types.Role
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.Background())
		if err != nil {
			return nil, common.FormatListError("IAMロール", err)
		}
		roles = append(roles, page.Roles...)
	}

	roleItems := make([]RoleItem, 0, len(roles))
	excludes := common.RemoveDuplicates(opts.Exclude)
	for _, r := range roles {
		name := aws.ToString(r.RoleName)
		if isServiceLinkedRole(r) || isServiceLinkedRoleName(name) {
			// サービスリンクは表示には残すが備考フラグにするので継続
		}
		if matchesAnyFilter(name, excludes) {
			continue
		}
		roleItems = append(roleItems, RoleItem{
			Name:            name,
			Arn:             aws.ToString(r.Arn),
			IsServiceLinked: isServiceLinkedRole(r) || isServiceLinkedRoleName(name),
		})
	}

	// last-used 並列取得
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
	return roleItems, nil
}

func listUnusedRoles(client *sdkiam.Client, opts ListOptions) ([]UnusedRole, error) {
	paginator := sdkiam.NewListRolesPaginator(client, &sdkiam.ListRolesInput{})
	var roles []types.Role
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.Background())
		if err != nil {
			return nil, common.FormatListError("IAMロール", err)
		}
		roles = append(roles, page.Roles...)
	}

	excludes := common.RemoveDuplicates(opts.Exclude)
	cutoff := time.Now().AddDate(0, 0, -opts.UnusedDays)

	names := make([]string, 0, len(roles))
	for _, r := range roles {
		name := aws.ToString(r.RoleName)
		if isServiceLinkedRole(r) || isServiceLinkedRoleName(name) {
			continue
		}
		if matchesAnyFilter(name, excludes) {
			continue
		}
		names = append(names, name)
	}

	exec := common.NewParallelExecutor(8)
	var mu sync.Mutex
	var out []UnusedRole
	for _, rn := range names {
		roleName := rn
		exec.Execute(func() {
			outRole, err := client.GetRole(context.Background(), &sdkiam.GetRoleInput{RoleName: aws.String(roleName)})
			if err != nil {
				return
			}
			var last *time.Time
			if outRole.Role.RoleLastUsed != nil && outRole.Role.RoleLastUsed.LastUsedDate != nil {
				t := *outRole.Role.RoleLastUsed.LastUsedDate
				last = &t
			}
			if last == nil || last.Before(cutoff) {
				mu.Lock()
				out = append(out, UnusedRole{Name: roleName, Arn: aws.ToString(outRole.Role.Arn), LastUsed: last})
				mu.Unlock()
			}
		})
	}
	exec.Wait()
	return out, nil
}

// ===== 判定/整形ヘルパー =====

func isServiceLinkedRole(r types.Role) bool    { return aws.ToString(r.Path) == "/aws-service-role/" }
func isServiceLinkedRoleName(name string) bool { return strings.HasPrefix(name, "AWSServiceRoleFor") }
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

// ===== 表示ヘルパー =====

func formatLastUsedLocal(t *time.Time) string {
	if t == nil {
		return "never"
	}
	return t.In(time.Local).Format("2006-01-02 15:04:05 MST")
}

func toRoleItemsTable(items []RoleItem) ([]common.TableColumn, [][]string) {
	cols := []common.TableColumn{{Header: "ロール名"}, {Header: "最終使用"}, {Header: "備考"}}
	rows := make([][]string, len(items))
	for i, r := range items {
		note := ""
		if r.IsServiceLinked {
			note = "service-linked"
		}
		rows[i] = []string{r.Name, formatLastUsedLocal(r.LastUsed), note}
	}
	return cols, rows
}

func toUnusedRolesTable(items []UnusedRole) ([]common.TableColumn, [][]string) {
	cols := []common.TableColumn{{Header: "ロール名"}, {Header: "最終使用"}}
	rows := make([][]string, len(items))
	for i, r := range items {
		rows[i] = []string{r.Name, formatLastUsedLocal(r.LastUsed)}
	}
	return cols, rows
}
