package aws

import (
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

// Clients AwsClients はAWS設定と各サービスクライアントを管理
type Clients struct {
	cfg aws.Config

	// 遅延初期化されるクライアント群
	ecs            *ecs.Client
	autoScaling    *applicationautoscaling.Client
	cfn            *cloudformation.Client
	s3             *s3.Client
	ec2            *ec2.Client
	ecr            *ecr.Client
	rds            *rds.Client
	secretsManager *secretsmanager.Client
	ses            *ses.Client
}

// NewAwsClients は認証情報からAWS設定を読み込んでクライアント管理構造体を作成
func NewAwsClients(ctx Context) (*Clients, error) {
	cfg, err := LoadAwsConfig(ctx)
	if err != nil {
		return nil, err
	}

	return &Clients{cfg: cfg}, nil
}

// Ecs は遅延初期化でECSクライアントを取得
func (c *Clients) Ecs() *ecs.Client {
	if c.ecs == nil {
		c.ecs = ecs.NewFromConfig(c.cfg)
	}
	return c.ecs
}

// AutoScaling は遅延初期化でAutoScalingクライアントを取得
func (c *Clients) AutoScaling() *applicationautoscaling.Client {
	if c.autoScaling == nil {
		c.autoScaling = applicationautoscaling.NewFromConfig(c.cfg)
	}
	return c.autoScaling
}

// Cfn は遅延初期化でCloudFormationクライアントを取得
func (c *Clients) Cfn() *cloudformation.Client {
	if c.cfn == nil {
		c.cfn = cloudformation.NewFromConfig(c.cfg)
	}
	return c.cfn
}

// S3 は遅延初期化でS3クライアントを取得
func (c *Clients) S3() *s3.Client {
	if c.s3 == nil {
		c.s3 = s3.NewFromConfig(c.cfg)
	}
	return c.s3
}

// Ec2 は遅延初期化でEC2クライアントを取得
func (c *Clients) Ec2() *ec2.Client {
	if c.ec2 == nil {
		c.ec2 = ec2.NewFromConfig(c.cfg)
	}
	return c.ec2
}

// Rds は遅延初期化でRDSクライアントを取得
func (c *Clients) Rds() *rds.Client {
	if c.rds == nil {
		c.rds = rds.NewFromConfig(c.cfg)
	}
	return c.rds
}

// SecretsManager は遅延初期化でSecretsManagerクライアントを取得
func (c *Clients) SecretsManager() *secretsmanager.Client {
	if c.secretsManager == nil {
		c.secretsManager = secretsmanager.NewFromConfig(c.cfg)
	}
	return c.secretsManager
}

// Ecr は遅延初期化でECRクライアントを取得
func (c *Clients) Ecr() *ecr.Client {
	if c.ecr == nil {
		c.ecr = ecr.NewFromConfig(c.cfg)
	}
	return c.ecr
}

// Ses は遅延初期化でSESクライアントを取得
func (c *Clients) Ses() *ses.Client {
	if c.ses == nil {
		c.ses = ses.NewFromConfig(c.cfg)
	}
	return c.ses
}
