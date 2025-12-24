package role

import (
	"awstk/internal/service/common"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	sdkiam "github.com/aws/aws-sdk-go-v2/service/iam"
)

// DeleteRoles はフィルター条件に一致するIAMロールを削除します
func DeleteRoles(client *sdkiam.Client, opts DeleteOptions) error {
	if client == nil {
		return fmt.Errorf("iam client is nil")
	}
	if opts.Filter == "" {
		return fmt.Errorf("フィルターパターンは必須です")
	}

	// 削除対象のロールを取得
	roleNames, err := getRolesForDeletion(client, opts)
	if err != nil {
		return err
	}

	if len(roleNames) == 0 {
		fmt.Printf("フィルター '%s' に一致するIAMロールが見つかりませんでした\n", opts.Filter)
		return nil
	}

	// 並列実行数を設定（最大8並列）
	maxWorkers := 8
	if len(roleNames) < maxWorkers {
		maxWorkers = len(roleNames)
	}

	executor := common.NewParallelExecutor(maxWorkers)
	results := make([]common.ProcessResult, len(roleNames))
	resultsMutex := &sync.Mutex{}

	fmt.Printf("🚀 %d個のロールを最大%d並列で削除します...\n\n", len(roleNames), maxWorkers)

	for i, roleName := range roleNames {
		idx := i
		name := roleName
		executor.Execute(func() {
			fmt.Printf("ロール %s を削除中...\n", name)

			err := deleteRole(client, name)

			resultsMutex.Lock()
			if err != nil {
				fmt.Printf("❌ ロール %s の削除に失敗しました: %v\n", name, err)
				results[idx] = common.ProcessResult{Item: name, Success: false, Error: err}
			} else {
				fmt.Printf("✅ ロール %s を削除しました\n", name)
				results[idx] = common.ProcessResult{Item: name, Success: true}
			}
			resultsMutex.Unlock()
		})
	}

	executor.Wait()

	// 結果の集計
	successCount, failCount := common.CollectResults(results)
	fmt.Printf("\n✅ 削除完了: 成功 %d個, 失敗 %d個\n", successCount, failCount)

	return nil
}

// getRolesForDeletion は削除対象のロール名一覧を取得します
func getRolesForDeletion(client *sdkiam.Client, opts DeleteOptions) ([]string, error) {
	paginator := sdkiam.NewListRolesPaginator(client, &sdkiam.ListRolesInput{})

	excludes := common.RemoveDuplicates(opts.Exclude)
	cutoff := time.Now().AddDate(0, 0, -opts.UnusedDays)

	var candidateRoles []string

	// 全ロールを取得してフィルタリング
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.Background())
		if err != nil {
			return nil, common.FormatListError("IAMロール", err)
		}

		for _, role := range page.Roles {
			name := aws.ToString(role.RoleName)

			// サービスリンクロールはスキップ
			if isServiceLinkedRole(role) || isServiceLinkedRoleByName(name) {
				continue
			}

			// 除外パターンチェック
			if matchesAnyFilter(name, excludes) {
				continue
			}

			// フィルターパターンチェック
			if !common.MatchesFilter(name, opts.Filter) {
				continue
			}

			candidateRoles = append(candidateRoles, name)
		}
	}

	// UnusedDaysフィルターが無効（0）の場合はそのまま返す
	if opts.UnusedDays == 0 {
		for _, name := range candidateRoles {
			fmt.Printf("🔍 検出されたIAMロール: %s\n", name)
		}
		return candidateRoles, nil
	}

	// UnusedDaysフィルターが有効な場合は最終使用日時でフィルタリング
	exec := common.NewParallelExecutor(8)
	var mu sync.Mutex
	var filteredRoles []string

	for _, roleName := range candidateRoles {
		roleNameCopy := roleName
		exec.Execute(func() {
			outRole, err := client.GetRole(context.Background(), &sdkiam.GetRoleInput{
				RoleName: aws.String(roleNameCopy),
			})
			if err != nil {
				return
			}

			var lastUsed *time.Time
			if outRole.Role.RoleLastUsed != nil && outRole.Role.RoleLastUsed.LastUsedDate != nil {
				lastUsedTime := *outRole.Role.RoleLastUsed.LastUsedDate
				lastUsed = &lastUsedTime
			}

			// -1: never used のみ
			if opts.UnusedDays == -1 {
				if lastUsed == nil {
					mu.Lock()
					filteredRoles = append(filteredRoles, roleNameCopy)
					fmt.Printf("🔍 検出されたIAMロール: %s (未使用)\n", roleNameCopy)
					mu.Unlock()
				}
				return
			}

			// >0: 指定日数以上未使用
			if lastUsed == nil || lastUsed.Before(cutoff) {
				mu.Lock()
				filteredRoles = append(filteredRoles, roleNameCopy)
				if lastUsed == nil {
					fmt.Printf("🔍 検出されたIAMロール: %s (未使用)\n", roleNameCopy)
				} else {
					fmt.Printf("🔍 検出されたIAMロール: %s (最終使用: %s)\n", roleNameCopy, lastUsed.Format("2006-01-02"))
				}
				mu.Unlock()
			}
		})
	}

	exec.Wait()
	return filteredRoles, nil
}

// deleteRole は単一のIAMロールを削除します（前処理含む）
func deleteRole(client *sdkiam.Client, roleName string) error {
	ctx := context.Background()

	// 1. インスタンスプロファイルからロールを削除
	profilesOutput, err := client.ListInstanceProfilesForRole(ctx, &sdkiam.ListInstanceProfilesForRoleInput{
		RoleName: aws.String(roleName),
	})
	if err == nil {
		for _, profile := range profilesOutput.InstanceProfiles {
			_, err := client.RemoveRoleFromInstanceProfile(ctx, &sdkiam.RemoveRoleFromInstanceProfileInput{
				InstanceProfileName: profile.InstanceProfileName,
				RoleName:            aws.String(roleName),
			})
			if err != nil {
				return fmt.Errorf("インスタンスプロファイルからのロール削除エラー: %w", err)
			}
		}
	}

	// 2. アタッチされた管理ポリシーをデタッチ
	attachedPoliciesOutput, err := client.ListAttachedRolePolicies(ctx, &sdkiam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(roleName),
	})
	if err == nil {
		for _, policy := range attachedPoliciesOutput.AttachedPolicies {
			_, err := client.DetachRolePolicy(ctx, &sdkiam.DetachRolePolicyInput{
				RoleName:  aws.String(roleName),
				PolicyArn: policy.PolicyArn,
			})
			if err != nil {
				return fmt.Errorf("ポリシーのデタッチエラー: %w", err)
			}
		}
	}

	// 3. インラインポリシーを削除
	inlinePoliciesOutput, err := client.ListRolePolicies(ctx, &sdkiam.ListRolePoliciesInput{
		RoleName: aws.String(roleName),
	})
	if err == nil {
		for _, policyName := range inlinePoliciesOutput.PolicyNames {
			_, err := client.DeleteRolePolicy(ctx, &sdkiam.DeleteRolePolicyInput{
				RoleName:   aws.String(roleName),
				PolicyName: aws.String(policyName),
			})
			if err != nil {
				return fmt.Errorf("インラインポリシーの削除エラー: %w", err)
			}
		}
	}

	// 4. ロールを削除
	_, err = client.DeleteRole(ctx, &sdkiam.DeleteRoleInput{
		RoleName: aws.String(roleName),
	})
	if err != nil {
		return fmt.Errorf("ロール削除エラー: %w", err)
	}

	return nil
}
