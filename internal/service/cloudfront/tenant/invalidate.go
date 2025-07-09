package tenant

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
)

// InvalidateTenant は特定テナントのキャッシュを無効化します
func InvalidateTenant(client *cloudfront.Client, distributionId, tenantId string, paths []string, wait bool) error {
	callerReference := fmt.Sprintf("awstk-tenant-%d", time.Now().Unix())

	input := &cloudfront.CreateInvalidationForDistributionTenantInput{
		Id: aws.String(tenantId),
		InvalidationBatch: &types.InvalidationBatch{
			CallerReference: aws.String(callerReference),
			Paths: &types.Paths{
				Quantity: aws.Int32(int32(len(paths))),
				Items:    paths,
			},
		},
	}

	result, err := client.CreateInvalidationForDistributionTenant(context.Background(), input)
	if err != nil {
		return err
	}

	if wait {
		// TODO: テナント無効化の完了待機実装
		fmt.Printf("   無効化ID: %s (待機機能は未実装)\n", *result.Invalidation.Id)
	}

	return nil
}

// InvalidateAllTenants は全テナントのキャッシュを無効化します
func InvalidateAllTenants(client *cloudfront.Client, distributionId string, paths []string, wait bool) error {
	// テナント一覧を取得
	tenants, err := ListTenants(client, distributionId)
	if err != nil {
		return fmt.Errorf("テナント一覧の取得に失敗: %w", err)
	}

	if len(tenants) == 0 {
		return fmt.Errorf("テナントが見つかりませんでした")
	}

	fmt.Printf("   対象テナント数: %d\n", len(tenants))
	fmt.Printf("   対象パス: %v\n", paths)

	// 並列処理で各テナントを無効化
	var wg sync.WaitGroup
	errChan := make(chan error, len(tenants))

	for _, tenant := range tenants {
		wg.Add(1)
		go func(t TenantInfo) {
			defer wg.Done()
			fmt.Printf("   テナント %s を無効化中...\n", t.Id)
			if err := InvalidateTenant(client, distributionId, t.Id, paths, false); err != nil {
				errChan <- fmt.Errorf("テナント %s の無効化に失敗: %w", t.Id, err)
			}
		}(tenant)
	}

	wg.Wait()
	close(errChan)

	// エラーがあれば最初のものを返す
	for err := range errChan {
		return err
	}

	return nil
}