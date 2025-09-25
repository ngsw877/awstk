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

// ListIamRoles cmdから呼ばれるメイン関数（Get + Display）
func ListIamRoles(client *sdkiam.Client, opts ListOptions) error {
	if client == nil {
		return fmt.Errorf("iam client is nil")
	}

	// Get: データ取得
	// -1: 引数なし（never used のみ）  /  >0: 指定日数以上未使用
	if opts.UnusedDays == -1 {
		items, err := getNeverUsedIamRoles(client, opts)
		if err != nil {
			return err
		}
		// Display: 共通表示処理
		return common.DisplayList(items, "未使用のIAMロール", unusedIamRolesToTableData, &common.DisplayOptions{
			ShowCount:    true,
			EmptyMessage: common.FormatEmptyMessage("未使用のIAMロール"),
		})
	}

	if opts.UnusedDays > 0 {
		items, err := getUnusedIamRoles(client, opts)
		if err != nil {
			return err
		}
		// Display: 共通表示処理
		return common.DisplayList(items, "未使用のIAMロール", unusedIamRolesToTableData, &common.DisplayOptions{
			ShowCount:    true,
			EmptyMessage: common.FormatEmptyMessage("未使用のIAMロール"),
		})
	}

	items, err := getAllIamRoles(client, opts)
	if err != nil {
		return err
	}
	// Display: 共通表示処理
	return common.DisplayList(items, "IAMロール一覧", iamRolesToTableData, &common.DisplayOptions{ShowCount: true})
}

// getAllIamRoles IAMロール一覧を取得
func getAllIamRoles(client *sdkiam.Client, opts ListOptions) ([]RoleItem, error) {
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
	for _, role := range roles {
		name := aws.ToString(role.RoleName)

		// staticcheck誤検知回避のため、変数に格納してから使用
		isServiceLinked := isServiceLinkedRole(role)
		isServiceLinkedByName := isServiceLinkedRoleByName(name)
		if isServiceLinked || isServiceLinkedByName {
			// サービスリンクは表示には残すが備考フラグにするので継続
			continue // 処理をスキップ
		}
		if matchesAnyFilter(name, excludes) {
			continue
		}
		roleItems = append(roleItems, RoleItem{
			Name:            name,
			Arn:             aws.ToString(role.Arn),
			IsServiceLinked: isServiceLinkedRole(role) || isServiceLinkedRoleByName(name), // サービスリンク判定
		})
	}

	// 最終使用日時を並列取得
	exec := common.NewParallelExecutor(8)
	var mu sync.Mutex
	for index := range roleItems {
		itemIndex := index
		exec.Execute(func() {
			outRole, err := client.GetRole(context.Background(), &sdkiam.GetRoleInput{RoleName: aws.String(roleItems[itemIndex].Name)})
			if err != nil {
				return
			}
			var last *time.Time
			if outRole.Role.RoleLastUsed != nil && outRole.Role.RoleLastUsed.LastUsedDate != nil {
				lastUsedTime := *outRole.Role.RoleLastUsed.LastUsedDate
				last = &lastUsedTime
			}
			mu.Lock()
			roleItems[itemIndex].LastUsed = last
			mu.Unlock()
		})
	}
	exec.Wait()
	return roleItems, nil
}

// getUnusedIamRoles 指定日数以上未使用のIAMロールを取得
func getUnusedIamRoles(client *sdkiam.Client, opts ListOptions) ([]UnusedRole, error) {
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

	roleNames := make([]string, 0, len(roles))
	for _, role := range roles {
		name := aws.ToString(role.RoleName)

		// staticcheck誤検知回避のため、変数に格納してから使用
		isServiceLinked := isServiceLinkedRole(role)
		isServiceLinkedByName := isServiceLinkedRoleByName(name)
		if isServiceLinked || isServiceLinkedByName {
			continue
		}
		if matchesAnyFilter(name, excludes) {
			continue
		}
		roleNames = append(roleNames, name)
	}

	exec := common.NewParallelExecutor(8)
	var mu sync.Mutex
	var unusedRoles []UnusedRole
	for _, roleName := range roleNames {
		roleNameCopy := roleName
		exec.Execute(func() {
			outRole, err := client.GetRole(context.Background(), &sdkiam.GetRoleInput{RoleName: aws.String(roleNameCopy)})
			if err != nil {
				return
			}
			var last *time.Time
			if outRole.Role.RoleLastUsed != nil && outRole.Role.RoleLastUsed.LastUsedDate != nil {
				lastUsedTime := *outRole.Role.RoleLastUsed.LastUsedDate
				last = &lastUsedTime
			}
			if last == nil || last.Before(cutoff) {
				mu.Lock()
				unusedRoles = append(unusedRoles, UnusedRole{Name: roleNameCopy, Arn: aws.ToString(outRole.Role.Arn), LastUsed: last})
				mu.Unlock()
			}
		})
	}
	exec.Wait()
	return unusedRoles, nil
}

// isServiceLinkedRole サービスリンクロールか判定
func isServiceLinkedRole(role types.Role) bool {
	return aws.ToString(role.Path) == "/aws-service-role/"
}

// isServiceLinkedRoleByName ロール名からサービスリンクロールか判定
func isServiceLinkedRoleByName(name string) bool { return strings.HasPrefix(name, "AWSServiceRoleFor") }

// matchesAnyFilter 除外パターンに一致するか判定
func matchesAnyFilter(name string, filters []string) bool {
	if len(filters) == 0 {
		return false
	}
	for _, filter := range filters {
		if filter == "" {
			continue
		}
		if common.MatchesFilter(name, filter) {
			return true
		}
	}
	return false
}

// formatIamRoleLastUsedLocal IAMロールの最終使用日時をローカル時刻に整形
func formatIamRoleLastUsedLocal(lastUsedTime *time.Time) string {
	if lastUsedTime == nil {
		return "never"
	}
	return lastUsedTime.In(time.Local).Format("2006-01-02 15:04:05 MST")
}

// iamRolesToTableData IAMロール情報をテーブルデータに変換
func iamRolesToTableData(items []RoleItem) ([]common.TableColumn, [][]string) {
	cols := []common.TableColumn{{Header: "ロール名"}, {Header: "最終使用"}, {Header: "備考"}}
	rows := make([][]string, len(items))
	for index, roleItem := range items {
		note := ""
		if roleItem.IsServiceLinked {
			note = "service-linked"
		}
		rows[index] = []string{roleItem.Name, formatIamRoleLastUsedLocal(roleItem.LastUsed), note}
	}
	return cols, rows
}

// unusedIamRolesToTableData 未使用IAMロール情報をテーブルデータに変換
func unusedIamRolesToTableData(items []UnusedRole) ([]common.TableColumn, [][]string) {
	cols := []common.TableColumn{{Header: "ロール名"}, {Header: "最終使用"}}
	rows := make([][]string, len(items))
	for index, unusedRole := range items {
		rows[index] = []string{unusedRole.Name, formatIamRoleLastUsedLocal(unusedRole.LastUsed)}
	}
	return cols, rows
}

// getNeverUsedIamRoles 一度も使用されていないIAMロールを取得
func getNeverUsedIamRoles(client *sdkiam.Client, opts ListOptions) ([]UnusedRole, error) {
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
	roleNames := make([]string, 0, len(roles))
	for _, role := range roles {
		name := aws.ToString(role.RoleName)

		// staticcheck誤検知回避のため、変数に格納してから使用
		isServiceLinked := isServiceLinkedRole(role)
		isServiceLinkedByName := isServiceLinkedRoleByName(name)
		if isServiceLinked || isServiceLinkedByName {
			continue
		}
		if matchesAnyFilter(name, excludes) {
			continue
		}
		roleNames = append(roleNames, name)
	}

	exec := common.NewParallelExecutor(8)
	var mu sync.Mutex
	var unusedRoles []UnusedRole
	for _, roleName := range roleNames {
		roleNameCopy := roleName
		exec.Execute(func() {
			outRole, err := client.GetRole(context.Background(), &sdkiam.GetRoleInput{RoleName: aws.String(roleNameCopy)})
			if err != nil {
				return
			}
			var last *time.Time
			if outRole.Role.RoleLastUsed != nil && outRole.Role.RoleLastUsed.LastUsedDate != nil {
				lastUsedTime := *outRole.Role.RoleLastUsed.LastUsedDate
				last = &lastUsedTime
			}
			if last == nil {
				mu.Lock()
				unusedRoles = append(unusedRoles, UnusedRole{Name: roleNameCopy, Arn: aws.ToString(outRole.Role.Arn), LastUsed: last})
				mu.Unlock()
			}
		})
	}
	exec.Wait()
	return unusedRoles, nil
}
