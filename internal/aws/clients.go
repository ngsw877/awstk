package aws

import (
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ses"
)

// NewClient は指定されたAWSサービスクライアントを作成
// 使用例:
//
//	cfnClient, err := aws.NewClient[*cloudformation.Client](ctx)
//	ecsClient, err := aws.NewClient[*ecs.Client](ctx)
//	s3Client, err := aws.NewClient[*s3.Client](ctx)
func NewClient[T any](ctx Context) (T, error) {
	var zero T

	// AWS設定を読み込み
	cfg, err := LoadAwsConfig(ctx)
	if err != nil {
		return zero, err
	}

	// クライアントを作成
	clientType := reflect.TypeOf(zero)
	newClient := createClient(cfg, clientType)
	if newClient == nil {
		return zero, nil // サポートされていない型
	}

	return newClient.(T), nil
}

// createClient は型に基づいてクライアントを作成
func createClient(cfg aws.Config, clientType reflect.Type) interface{} {
	switch clientType.String() {
	case "*ecs.Client":
		return ecs.NewFromConfig(cfg)
	case "*s3.Client":
		return s3.NewFromConfig(cfg)
	case "*ec2.Client":
		return ec2.NewFromConfig(cfg)
	case "*rds.Client":
		return rds.NewFromConfig(cfg)
	case "*cloudformation.Client":
		return cloudformation.NewFromConfig(cfg)
	case "*ecr.Client":
		return ecr.NewFromConfig(cfg)
	case "*secretsmanager.Client":
		return secretsmanager.NewFromConfig(cfg)
	case "*ses.Client":
		return ses.NewFromConfig(cfg)
	case "*applicationautoscaling.Client":
		return applicationautoscaling.NewFromConfig(cfg)
	default:
		return nil
	}
}
