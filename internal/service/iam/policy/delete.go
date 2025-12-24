package policy

import (
	"awstk/internal/service/common"
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	sdkiam "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

// DeletePolicies はフィルター条件に一致するIAMポリシーを削除します
func DeletePolicies(client *sdkiam.Client, opts DeleteOptions) error {
	if client == nil {
		return fmt.Errorf("iam client is nil")
	}
	if opts.Filter == "" {
		return fmt.Errorf("フィルターパターンは必須です")
	}

	// 削除対象のポリシーを取得
	policies, err := getPoliciesForDeletion(client, opts)
	if err != nil {
		return err
	}

	if len(policies) == 0 {
		fmt.Printf("フィルター '%s' に一致するIAMポリシーが見つかりませんでした\n", opts.Filter)
		return nil
	}

	// 並列実行数を設定（最大8並列）
	maxWorkers := 8
	if len(policies) < maxWorkers {
		maxWorkers = len(policies)
	}

	executor := common.NewParallelExecutor(maxWorkers)
	results := make([]common.ProcessResult, len(policies))
	resultsMutex := &sync.Mutex{}

	fmt.Printf("🚀 %d個のポリシーを最大%d並列で削除します...\n\n", len(policies), maxWorkers)

	for i, policy := range policies {
		idx := i
		p := policy
		executor.Execute(func() {
			fmt.Printf("ポリシー %s を削除中...\n", p.Name)

			err := deletePolicy(client, p.Arn)

			resultsMutex.Lock()
			if err != nil {
				fmt.Printf("❌ ポリシー %s の削除に失敗しました: %v\n", p.Name, err)
				results[idx] = common.ProcessResult{Item: p.Name, Success: false, Error: err}
			} else {
				fmt.Printf("✅ ポリシー %s を削除しました\n", p.Name)
				results[idx] = common.ProcessResult{Item: p.Name, Success: true}
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

// getPoliciesForDeletion は削除対象のポリシー一覧を取得します
func getPoliciesForDeletion(client *sdkiam.Client, opts DeleteOptions) ([]PolicyItem, error) {
	paginator := sdkiam.NewListPoliciesPaginator(client, &sdkiam.ListPoliciesInput{
		Scope: types.PolicyScopeTypeLocal, // カスタマー管理ポリシーのみ
	})

	excludes := common.RemoveDuplicates(opts.Exclude)
	var policies []PolicyItem

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.Background())
		if err != nil {
			return nil, common.FormatListError("IAMポリシー", err)
		}

		for _, policy := range page.Policies {
			name := aws.ToString(policy.PolicyName)

			// 除外パターンチェック
			if matchesAnyFilter(name, excludes) {
				continue
			}

			// フィルターパターンチェック
			if !common.MatchesFilter(name, opts.Filter) {
				continue
			}

			// UnattachedOnlyフィルター
			attachmentCount := int32(0)
			if policy.AttachmentCount != nil {
				attachmentCount = *policy.AttachmentCount
			}

			if opts.UnattachedOnly && attachmentCount > 0 {
				continue
			}

			policies = append(policies, PolicyItem{
				Name:            name,
				Arn:             aws.ToString(policy.Arn),
				AttachmentCount: attachmentCount,
			})

			if attachmentCount == 0 {
				fmt.Printf("🔍 検出されたIAMポリシー: %s (未アタッチ)\n", name)
			} else {
				fmt.Printf("🔍 検出されたIAMポリシー: %s (アタッチ数: %d)\n", name, attachmentCount)
			}
		}
	}

	return policies, nil
}

// deletePolicy は単一のIAMポリシーを削除します（前処理含む）
func deletePolicy(client *sdkiam.Client, policyArn string) error {
	ctx := context.Background()

	// 1. ポリシーがアタッチされているエンティティからデタッチ
	entitiesOutput, err := client.ListEntitiesForPolicy(ctx, &sdkiam.ListEntitiesForPolicyInput{
		PolicyArn: aws.String(policyArn),
	})
	if err == nil {
		// ユーザーからデタッチ
		for _, user := range entitiesOutput.PolicyUsers {
			_, err := client.DetachUserPolicy(ctx, &sdkiam.DetachUserPolicyInput{
				UserName:  user.UserName,
				PolicyArn: aws.String(policyArn),
			})
			if err != nil {
				return fmt.Errorf("ユーザーからのポリシーデタッチエラー: %w", err)
			}
		}

		// グループからデタッチ
		for _, group := range entitiesOutput.PolicyGroups {
			_, err := client.DetachGroupPolicy(ctx, &sdkiam.DetachGroupPolicyInput{
				GroupName: group.GroupName,
				PolicyArn: aws.String(policyArn),
			})
			if err != nil {
				return fmt.Errorf("グループからのポリシーデタッチエラー: %w", err)
			}
		}

		// ロールからデタッチ
		for _, role := range entitiesOutput.PolicyRoles {
			_, err := client.DetachRolePolicy(ctx, &sdkiam.DetachRolePolicyInput{
				RoleName:  role.RoleName,
				PolicyArn: aws.String(policyArn),
			})
			if err != nil {
				return fmt.Errorf("ロールからのポリシーデタッチエラー: %w", err)
			}
		}
	}

	// 2. 非デフォルトバージョンを削除
	versionsOutput, err := client.ListPolicyVersions(ctx, &sdkiam.ListPolicyVersionsInput{
		PolicyArn: aws.String(policyArn),
	})
	if err == nil {
		for _, version := range versionsOutput.Versions {
			if !version.IsDefaultVersion {
				_, err := client.DeletePolicyVersion(ctx, &sdkiam.DeletePolicyVersionInput{
					PolicyArn: aws.String(policyArn),
					VersionId: version.VersionId,
				})
				if err != nil {
					return fmt.Errorf("ポリシーバージョン削除エラー: %w", err)
				}
			}
		}
	}

	// 3. ポリシーを削除
	_, err = client.DeletePolicy(ctx, &sdkiam.DeletePolicyInput{
		PolicyArn: aws.String(policyArn),
	})
	if err != nil {
		return fmt.Errorf("ポリシー削除エラー: %w", err)
	}

	return nil
}
