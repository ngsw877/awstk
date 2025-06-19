package s3

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// checkS3BucketAvailability は指定バケット名の利用可否判定・メッセージ生成まで行う
func checkS3BucketAvailability(s3Client *s3.Client, bucketName string) S3BucketAvailabilityResult {
	ctx := context.Background()
	input := &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	}
	_, err := s3Client.HeadBucket(ctx, input)

	if err == nil {
		return S3BucketAvailabilityResult{
			BucketName: bucketName,
			StatusCode: 200,
			Message:    "利用不可（すでに存在）",
		}
	}

	// HTTPレスポンスエラーからステータスコードを取得
	statusCode := 0
	var respErr *awshttp.ResponseError
	if errors.As(err, &respErr) {
		statusCode = respErr.HTTPStatusCode()
	}

	var msg string
	switch statusCode {
	case 404:
		msg = "利用可能"
	case 403:
		msg = "利用不可（存在するがアクセス権限なし）"
	case 301:
		msg = "利用不可（リージョン不一致）"
	default:
		msg = fmt.Sprintf("利用不可（エラー: %v）", err)
	}

	return S3BucketAvailabilityResult{
		BucketName: bucketName,
		StatusCode: statusCode,
		Message:    msg,
	}
}

// CheckS3BucketsAvailability 複数バケットの利用可否をまとめて判定
func CheckS3BucketsAvailability(s3Client *s3.Client, buckets []string) []S3BucketAvailabilityResult {
	results := make([]S3BucketAvailabilityResult, 0, len(buckets))
	for _, bucket := range buckets {
		results = append(results, checkS3BucketAvailability(s3Client, bucket))
	}
	return results
}
