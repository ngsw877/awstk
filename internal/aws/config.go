package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// LoadAwsConfig は認証情報からAWS設定を読み込む（既存のutils.goから移植）
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

// GetConfig は遅延初期化でAWS設定を取得（初回のみ認証処理実行）
func (ctx *Context) GetConfig() (aws.Config, error) {
	if ctx.config == nil {
		cfg, err := LoadAwsConfig(*ctx)
		if err != nil {
			return aws.Config{}, err
		}
		ctx.config = &cfg
	}
	return *ctx.config, nil
}
