package internal

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// LoadAwsConfig はAWS設定を読み込む共通関数
func LoadAwsConfig(region, profile string) (aws.Config, error) {
	opts := make([]func(*config.LoadOptions) error, 0)

	// プロファイル指定があればWithSharedConfigProfileで指定
	if profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}
	// リージョン指定があればWithRegionで指定
	if region != "" {
		opts = append(opts, config.WithRegion(region))
	}
	// オプション未指定時は環境変数やデフォルト設定（~/.aws/config, ~/.aws/credentials, AWS_PROFILE, AWS_REGIONなど）を自動で参照する
	return config.LoadDefaultConfig(context.Background(), opts...)
}
