package internal

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// LoadAwsConfig はAWS設定を読み込む共通関数
func LoadAwsConfig(region, profile string) (aws.Config, error) {
	if profile != "" {
		os.Setenv("AWS_PROFILE", profile)
	}
	if region != "" {
		os.Setenv("AWS_REGION", region)
	}
	// AWS設定を取得
	return config.LoadDefaultConfig(context.TODO())
}
