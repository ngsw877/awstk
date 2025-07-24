package cloudfront

import (
	"awstk/internal/service/cfn"
	"awstk/internal/service/cloudfront/tenant"
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
)

// CreateInvalidation はCloudFrontディストリビューションのキャッシュを無効化します
func CreateInvalidation(client *cloudfront.Client, distributionId string, paths []string) (string, error) {
	// パスをAWS SDKの形式に変換
	var items []string
	items = append(items, paths...)

	// CallerReferenceとして現在時刻を使用
	callerReference := fmt.Sprintf("awstk-%d", time.Now().Unix())

	input := &cloudfront.CreateInvalidationInput{
		DistributionId: aws.String(distributionId),
		InvalidationBatch: &types.InvalidationBatch{
			CallerReference: aws.String(callerReference),
			Paths: &types.Paths{
				Quantity: aws.Int32(int32(len(items))),
				Items:    items,
			},
		},
	}

	result, err := client.CreateInvalidation(context.Background(), input)
	if err != nil {
		return "", err
	}

	return *result.Invalidation.Id, nil
}

// WaitForInvalidation は無効化が完了するまで待機します
func WaitForInvalidation(client *cloudfront.Client, distributionId, invalidationId string) error {
	for {
		input := &cloudfront.GetInvalidationInput{
			DistributionId: aws.String(distributionId),
			Id:             aws.String(invalidationId),
		}

		result, err := client.GetInvalidation(context.Background(), input)
		if err != nil {
			return err
		}

		status := *result.Invalidation.Status
		fmt.Printf("   現在のステータス: %s\n", status)

		if status == "Completed" {
			return nil
		}

		// 10秒待機してから再確認
		time.Sleep(10 * time.Second)
	}
}

// InvalidateOptions はキャッシュ無効化の共通オプション
type InvalidateOptions struct {
	DistributionId string   // オプション: ディストリビューションID（指定なしの場合はStackNameから解決）
	Paths          []string // 必須: 無効化するパス
	Wait           bool     // オプション: 無効化完了まで待機
	StackName      string   // オプション: CloudFormationスタック名（DistributionId未指定時に使用）
}

// InvalidateByIdOrStack はディストリビューションIDまたはスタック名を使用してキャッシュを無効化します
func InvalidateByIdOrStack(cfClient *cloudfront.Client, cfnClient *cloudformation.Client, opts InvalidateOptions) error {
	// ディストリビューションIDの解決
	resolvedId, err := resolveDistributionId(cfClient, cfnClient, opts.DistributionId, opts.StackName)
	if err != nil {
		return err
	}

	fmt.Printf("🚀 CloudFrontディストリビューション (%s) のキャッシュを無効化します...\n", resolvedId)
	fmt.Printf("   対象パス: %v\n", opts.Paths)

	// キャッシュ無効化の実行
	invalidationId, err := CreateInvalidation(cfClient, resolvedId, opts.Paths)
	if err != nil {
		return fmt.Errorf("キャッシュ無効化エラー: %w", err)
	}

	fmt.Printf("✅ キャッシュ無効化を開始しました (ID: %s)\n", invalidationId)

	// 待機オプションが有効な場合
	if opts.Wait {
		fmt.Println("⏳ 無効化の完了を待機しています...")
		err = WaitForInvalidation(cfClient, resolvedId, invalidationId)
		if err != nil {
			return fmt.Errorf("無効化待機エラー: %w", err)
		}
		fmt.Println("✅ キャッシュ無効化が完了しました")
	}

	return nil
}

// InvalidateTenantByIdOrSelection はテナントIDまたは選択によってテナントキャッシュを無効化します
func InvalidateTenantByIdOrSelection(cfClient *cloudfront.Client, selectFromList bool, opts tenant.InvalidateOptions) error {
	if selectFromList {
		// テナント一覧から選択
		resolvedTenantId, err := tenant.SelectTenant(cfClient, opts.DistributionId)
		if err != nil {
			return fmt.Errorf("テナント選択エラー: %w", err)
		}
		opts.TenantId = resolvedTenantId
	} else {
		if opts.TenantId == "" {
			return fmt.Errorf("テナントID、--all、または --list オプションを指定してください")
		}
		fmt.Printf("🚀 テナント (%s) のキャッシュを無効化します...\n", opts.TenantId)
		fmt.Printf("   対象パス: %v\n", opts.Paths)
	}

	err := tenant.InvalidateTenant(cfClient, opts)
	if err != nil {
		return fmt.Errorf("キャッシュ無効化エラー: %w", err)
	}

	fmt.Printf("✅ テナント '%s' のキャッシュ無効化を開始しました\n", opts.TenantId)
	return nil
}

// InvalidateAllTenantsWithMessage は全テナントのキャッシュを無効化します（メッセージ付き）
func InvalidateAllTenantsWithMessage(cfClient *cloudfront.Client, opts tenant.InvalidateOptions) error {
	fmt.Printf("🚀 CloudFrontディストリビューション (%s) の全テナントのキャッシュを無効化します...\n", opts.DistributionId)

	err := tenant.InvalidateAllTenants(cfClient, opts)
	if err != nil {
		return fmt.Errorf("全テナントキャッシュ無効化エラー: %w", err)
	}

	fmt.Println("✅ 全テナントのキャッシュ無効化を開始しました")
	return nil
}

// resolveDistributionId はディストリビューションIDを解決します
func resolveDistributionId(cfClient *cloudfront.Client, cfnClient *cloudformation.Client, distributionId, stackName string) (string, error) {
	// 既にディストリビューションIDが指定されている場合
	if distributionId != "" {
		return distributionId, nil
	}

	// スタック名が指定されていない場合
	if stackName == "" {
		return "", fmt.Errorf("ディストリビューションID またはスタック名 (-S) を指定してください")
	}

	// スタックからCloudFrontディストリビューションを取得
	distributions, err := cfn.GetAllCloudFrontFromStack(cfnClient, stackName)
	if err != nil {
		return "", fmt.Errorf("CloudFormationスタックからディストリビューションの取得に失敗: %w", err)
	}

	if len(distributions) == 0 {
		return "", fmt.Errorf("スタック '%s' にCloudFrontディストリビューションが見つかりませんでした", stackName)
	}

	if len(distributions) == 1 {
		distributionId = distributions[0]
		fmt.Printf("✅ CloudFormationスタック '%s' からCloudFrontディストリビューション '%s' を検出しました\n", stackName, distributionId)
		return distributionId, nil
	}

	// 複数のディストリビューションがある場合は選択
	return SelectDistribution(cfClient, distributions)
}
