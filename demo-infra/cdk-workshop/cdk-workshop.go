package main

import (
	"cdk-workshop/hitcounter"
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3assets"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdklabs/cdk-dynamo-table-viewer-go/dynamotableviewer"
)

type CdkWorkshopStackProps struct {
	awscdk.StackProps
}

func NewCdkWorkshopStack(scope constructs.Construct, id string, props *CdkWorkshopStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Build commands for better readability
	buildCommands := []string{
		"CGO_ENABLED=0",
		"GOOS=linux",
		"GOARCH=arm64",
		// Dockerコンテナ内での権限問題を避けるため、Goキャッシュとワークスペースを/tmpに設定
		// CDKはデフォルトユーザー(503:20)で実行され、/.cacheなどのデフォルトの場所への書き込み権限がない
		"export GOCACHE=/tmp/go-cache",
		"export GOPATH=/tmp/go-path",
		// /asset-outputはCDKの特別なパス。ここに出力されたファイルはcdk.outディレクトリに保存され、
		// 最終的にS3経由でLambda実行環境にデプロイされる
		"go build -tags lambda.norpc -o /asset-output/bootstrap main.go",
	}

	helloHandler := awslambda.NewFunction(stack, jsii.String("HelloHandler"), &awslambda.FunctionProps{
		Code: awslambda.Code_FromAsset(jsii.String("lambda/hello"), &awss3assets.AssetOptions{
			Bundling: &awscdk.BundlingOptions{
				Image: awscdk.DockerImage_FromRegistry(jsii.String("golang:1.24")),
				Command: &[]*string{
					jsii.String("bash"),
					jsii.String("-c"),
					jsii.String(strings.Join(buildCommands, " && ")),
				},
			},
		}),
		Runtime:      awslambda.Runtime_PROVIDED_AL2023(),
		Handler:      jsii.String("bootstrap"),
		Architecture: awslambda.Architecture_ARM_64(),
	})

	hitcounter := hitcounter.NewHitCounter(stack, "HelloHitCounter", &hitcounter.HitCounterProps{
		Downstream:   helloHandler,
		ReadCapacity: 10,
	})

	awsapigateway.NewLambdaRestApi(stack, jsii.String("Endpoint"), &awsapigateway.LambdaRestApiProps{
		Handler: hitcounter.Handler(),
	})

	// https://pkg.go.dev/github.com/cdklabs/cdk-dynamo-table-viewer-go/dynamotableviewer#section-readme
	dynamotableviewer.NewTableViewer(stack, jsii.String("ViewHitCounter"), &dynamotableviewer.TableViewerProps{
		Title:  jsii.String("Hello Hits"),
		Table:  hitcounter.Table(),
		SortBy: jsii.String("-hits"),
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewCdkWorkshopStack(app, "CdkWorkshopStack", &CdkWorkshopStackProps{})

	app.Synth(nil)
}
