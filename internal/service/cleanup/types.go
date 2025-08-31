package cleanup

import (
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// ClientSet はクリーンアップ処理に必要なクライアントをまとめた構造体
type ClientSet struct {
	S3Client   *s3.Client
	EcrClient  *ecr.Client
	CfnClient  *cloudformation.Client
	LogsClient *cloudwatchlogs.Client
}

// Options はクリーンアップ処理のパラメータを格納する構造体
type Options struct {
	SearchString string // 検索文字列
	StackName    string // CloudFormationスタック名
	StackId      string // CloudFormationスタックID (ARN可)
}
