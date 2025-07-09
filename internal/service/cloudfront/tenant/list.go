package tenant

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
)

// ListTenants はディストリビューションに関連付けられたテナント一覧を取得します
func ListTenants(client *cloudfront.Client, distributionId string) ([]TenantInfo, error) {
	var tenants []TenantInfo
	var nextMarker *string

	for {
		input := &cloudfront.ListDistributionTenantsInput{
			AssociationFilter: &types.DistributionTenantAssociationFilter{
				DistributionId: aws.String(distributionId),
			},
		}

		if nextMarker != nil {
			input.Marker = nextMarker
		}

		result, err := client.ListDistributionTenants(context.Background(), input)
		if err != nil {
			return nil, err
		}

		// DistributionTenantListは直接テナントのスライス
		if result.DistributionTenantList != nil {
			for _, item := range result.DistributionTenantList {
				tenant := TenantInfo{
					Id: aws.ToString(item.Id),
					AssociatedDistributionId: distributionId,
				}
				
				// エイリアスがあれば設定（現在のAPIではエイリアスフィールドが存在しない可能性がある）
				// TODO: AWS SDK更新時に確認
				
				tenants = append(tenants, tenant)
			}
		}

		// 次のページがなければ終了（ページネーションのサポートを確認）
		if result.NextMarker == nil || *result.NextMarker == "" {
			break
		}

		// 次のマーカーを設定
		nextMarker = result.NextMarker
	}

	return tenants, nil
}