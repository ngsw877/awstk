package internal

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// LoadAwsConfig はAWS設定を読み込む共通関数
func LoadAwsConfig(ctx AwsContext) (aws.Config, error) {
	opts := make([]func(*config.LoadOptions) error, 0)

	if ctx.Profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(ctx.Profile))
	}
	if ctx.Region != "" {
		opts = append(opts, config.WithRegion(ctx.Region))
	}
	return config.LoadDefaultConfig(context.Background(), opts...)
}
