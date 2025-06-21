package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// LoadAwsConfig は認証情報からAWS設定を読み込む
func LoadAwsConfig(ctx Context) (aws.Config, error) {
	opts := make([]func(*config.LoadOptions) error, 0)

	if ctx.Profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(ctx.Profile))
	}
	if ctx.Region != "" {
		opts = append(opts, config.WithRegion(ctx.Region))
	}
	return config.LoadDefaultConfig(context.Background(), opts...)
}
