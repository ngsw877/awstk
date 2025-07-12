package cleanup

import (
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Options CleanupOptions はクリーンアップ処理のパラメータを格納する構造体
type Options struct {
	S3Client     *s3.Client
	EcrClient    *ecr.Client
	CfnClient    *cloudformation.Client
	SearchString string // 検索文字列
	StackName    string // CloudFormationスタック名
}
