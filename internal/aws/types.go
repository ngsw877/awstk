package aws

import "github.com/aws/aws-sdk-go-v2/aws"

// Context AwsContext は認証情報を保持
type Context struct {
	Profile string
	Region  string
	config  *aws.Config // AWS設定のキャッシュ（非公開）
}
