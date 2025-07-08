package cloudfront

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
)

// CreateInvalidation はCloudFrontディストリビューションのキャッシュを無効化します
func CreateInvalidation(client *cloudfront.Client, distributionId string, paths []string) (string, error) {
	// パスをAWS SDKの形式に変換
	var items []string
	for _, path := range paths {
		items = append(items, path)
	}

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