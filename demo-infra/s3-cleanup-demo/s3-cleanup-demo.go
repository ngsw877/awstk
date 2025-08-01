package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3assets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3deployment"
	"github.com/aws/aws-cdk-go/awscdk/v2/customresources"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type S3CleanupDemoStackProps struct {
	awscdk.StackProps
}

// 定数定義
const (
	bucketPrefix = "awstk-s3cleanupdemo-"
	stackPurpose = "S3CleanupDemo"
	environment  = "Demo"
)

func NewS3CleanupDemoStack(scope constructs.Construct, id string, props *S3CleanupDemoStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// スタック全体にタグを追加（コスト追跡用）
	awscdk.Tags_Of(stack).Add(jsii.String("Purpose"), jsii.String(stackPurpose), nil)
	awscdk.Tags_Of(stack).Add(jsii.String("Environment"), jsii.String(environment), nil)

	// 1. 空のバケット
	awss3.NewBucket(stack, jsii.String("EmptyBucket"), &awss3.BucketProps{
		BucketName:        jsii.String(bucketPrefix + "empty-bucket"),
		Encryption:        awss3.BucketEncryption_S3_MANAGED,
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
		AutoDeleteObjects: jsii.Bool(true),
	})

	// 2. 通常バケット（10個のオブジェクト）
	normalBucket := awss3.NewBucket(stack, jsii.String("NormalBucket"), &awss3.BucketProps{
		BucketName:        jsii.String(bucketPrefix + "normal-bucket"),
		Encryption:        awss3.BucketEncryption_S3_MANAGED,
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
		AutoDeleteObjects: jsii.Bool(true),
	})

	// demo-dataディレクトリの内容をデプロイ
	awss3deployment.NewBucketDeployment(stack, jsii.String("DeployTestData"), &awss3deployment.BucketDeploymentProps{
		Sources: &[]awss3deployment.ISource{
			awss3deployment.Source_Asset(jsii.String("./demo-data"), nil),
		},
		DestinationBucket: normalBucket,
	})

	// 3. ネストされたフォルダ構造
	nestedBucket := awss3.NewBucket(stack, jsii.String("NestedBucket"), &awss3.BucketProps{
		BucketName:        jsii.String(bucketPrefix + "nested-bucket"),
		Encryption:        awss3.BucketEncryption_S3_MANAGED,
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
		AutoDeleteObjects: jsii.Bool(true),
	})

	// ネストされた構造もdemo-dataから取得
	awss3deployment.NewBucketDeployment(stack, jsii.String("DeployNestedData"), &awss3deployment.BucketDeploymentProps{
		Sources: &[]awss3deployment.ISource{
			awss3deployment.Source_Asset(jsii.String("./demo-data"), nil),
		},
		DestinationBucket:    nestedBucket,
		DestinationKeyPrefix: jsii.String("deep/nested/folder/"),
	})

	// 4. バージョニング有効バケット（複数バージョン）
	versionedBucket := awss3.NewBucket(stack, jsii.String("VersionedBucket"), &awss3.BucketProps{
		BucketName:        jsii.String(bucketPrefix + "versioned-bucket"),
		Versioned:         jsii.Bool(true),
		Encryption:        awss3.BucketEncryption_S3_MANAGED,
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
		AutoDeleteObjects: jsii.Bool(true),
	})

	// 5. バージョニング有効バケット（削除マーカー）
	deletedMarkerBucket := awss3.NewBucket(stack, jsii.String("DeletedMarkerBucket"), &awss3.BucketProps{
		BucketName:        jsii.String(bucketPrefix + "deleted-marker-bucket"),
		Versioned:         jsii.Bool(true),
		Encryption:        awss3.BucketEncryption_S3_MANAGED,
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
		AutoDeleteObjects: jsii.Bool(true),
	})

	// 6. 大量オブジェクトバケット（1000個以上）
	largeBucket := awss3.NewBucket(stack, jsii.String("LargeBucket"), &awss3.BucketProps{
		BucketName:        jsii.String(bucketPrefix + "large-bucket"),
		Encryption:        awss3.BucketEncryption_S3_MANAGED,
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
		AutoDeleteObjects: jsii.Bool(true),
	})

	// Lambda関数を作成（データ作成用）
	dataCreatorFunction := awslambda.NewFunction(stack, jsii.String("DataCreatorFunction"), &awslambda.FunctionProps{
		Runtime: awslambda.Runtime_PROVIDED_AL2023(),
		Handler: jsii.String("bootstrap"),
		Code:    awslambda.Code_FromAsset(jsii.String("lambda"), &awss3assets.AssetOptions{
			Bundling: &awscdk.BundlingOptions{
				Image: awscdk.DockerImage_FromRegistry(jsii.String("golang:1.24")),
				Command: &[]*string{
					jsii.String("bash"),
					jsii.String("-c"),
					jsii.String("export GOCACHE=/tmp/go-cache && export GOPATH=/tmp/go-path && cd /asset-input && go mod download && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o /asset-output/bootstrap data-creator.go"),
				},
			},
		}),
		Timeout: awscdk.Duration_Minutes(jsii.Number(10)),
		MemorySize: jsii.Number(512),
		LogGroup: awslogs.NewLogGroup(stack, jsii.String("DataCreatorLogGroup"), &awslogs.LogGroupProps{
			Retention: awslogs.RetentionDays_ONE_WEEK,
			RemovalPolicy: awscdk.RemovalPolicy_DESTROY, // スタック削除時にログも削除
		}),
	})

	// Lambda関数にS3への権限を付与
	dataCreatorFunction.AddToRolePolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Actions: &[]*string{
			jsii.String("s3:PutObject"),
			jsii.String("s3:DeleteObject"),
		},
		Resources: &[]*string{
			versionedBucket.ArnForObjects(jsii.String("*")),
			deletedMarkerBucket.ArnForObjects(jsii.String("*")),
			largeBucket.ArnForObjects(jsii.String("*")),
		},
	}))

	// CloudFormation Custom Resourceプロバイダー
	provider := customresources.NewProvider(stack, jsii.String("DataCreatorProvider"), &customresources.ProviderProps{
		OnEventHandler: dataCreatorFunction,
	})

	// カスタムリソースの作成ヘルパー関数
	createDataResource := func(id string, bucket awss3.IBucket, objectCount int, pattern string) {
		resource := awscdk.NewCustomResource(stack, jsii.String(id), &awscdk.CustomResourceProps{
			ServiceToken: provider.ServiceToken(),
			Properties: &map[string]interface{}{
				"BucketName":  bucket.BucketName(),
				"ObjectCount": objectCount,
				"Pattern":     pattern,
			},
		})
		resource.Node().AddDependency(bucket)
	}

	// 各バケット用のカスタムリソースを作成
	createDataResource("VersionedData", versionedBucket, 5, "versioned")
	createDataResource("DeletedMarkerData", deletedMarkerBucket, 3, "deleted-markers")
	createDataResource("LargeData", largeBucket, 1200, "normal")

	// クリーンアップコマンドを出力として追加
	awscdk.NewCfnOutput(stack, jsii.String("CleanupCommand"), &awscdk.CfnOutputProps{
		Value:       jsii.String("awstk s3 cleanup --filter \"" + bucketPrefix + "\""),
		Description: jsii.String("Command to cleanup all demo buckets"),
	})

	// スタック名を出力として追加
	awscdk.NewCfnOutput(stack, jsii.String("StackName"), &awscdk.CfnOutputProps{
		Value:       stack.StackName(),
		Description: jsii.String("Stack name for reference"),
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewS3CleanupDemoStack(app, "S3CleanupDemo", &S3CleanupDemoStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	//---------------------------------------------------------------------------
	return nil

	// Uncomment if you know exactly what account and region you want to deploy
	// the stack to. This is the recommendation for production stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String("123456789012"),
	//  Region:  jsii.String("us-east-1"),
	// }

	// Uncomment to specialize this stack for the AWS Account and Region that are
	// implied by the current CLI configuration. This is recommended for dev
	// stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
	//  Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	// }
}
