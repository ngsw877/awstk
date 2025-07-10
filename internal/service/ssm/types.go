package ssm

import (
	"awstk/internal/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// SessionOptions SsmSessionOptions はSSMセッション開始のパラメータを格納する構造体
type SessionOptions struct {
	AwsCtx     aws.Context
	InstanceId string
}

// PutParamsOptions はパラメータ一括登録のオプション
type PutParamsOptions struct {
	SsmClient *ssm.Client
	FilePath  string
	Prefix    string
	DryRun    bool
}

// parameter はSSMパラメータを表す構造体
type parameter struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	Type        string `json:"type"` // String, SecureString, StringList
	Description string `json:"description,omitempty"`
}

// parametersFile はJSONファイルの構造を表す
type parametersFile struct {
	Parameters []parameter `json:"parameters"`
}

// DeleteParamsOptions はパラメータ一括削除のオプション
type DeleteParamsOptions struct {
	SsmClient *ssm.Client
	FilePath  string
	Prefix    string
	DryRun    bool
	Force     bool
}
