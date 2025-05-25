package hitcounter

import (
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3assets"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type HitCounterProps struct {
	Downstream   awslambda.IFunction
	ReadCapacity float64
}

type hitCounter struct {
	constructs.Construct
	handler awslambda.IFunction
	table   awsdynamodb.Table
}

type HitCounter interface {
	constructs.Construct
	Handler() awslambda.IFunction
	Table() awsdynamodb.Table
}

func NewHitCounter(scope constructs.Construct, id string, props *HitCounterProps) HitCounter {
	if props.ReadCapacity < 5 || props.ReadCapacity > 20 {
		panic("ReadCapacity must be between 5 and 20")
	}
	this := constructs.NewConstruct(scope, &id)

	table := awsdynamodb.NewTable(this, jsii.String("Hits"), &awsdynamodb.TableProps{
		PartitionKey:  &awsdynamodb.Attribute{Name: jsii.String("path"), Type: awsdynamodb.AttributeType_STRING},
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		Encryption:    awsdynamodb.TableEncryption_AWS_MANAGED,
		ReadCapacity:  &props.ReadCapacity,
	})

	handler := awslambda.NewFunction(this, jsii.String("HitCounterHandler"), &awslambda.FunctionProps{
		Runtime: awslambda.Runtime_PROVIDED_AL2023(),
		Code: awslambda.Code_FromAsset(jsii.String("lambda/hitcounter"), &awss3assets.AssetOptions{
			Bundling: &awscdk.BundlingOptions{
				Image: awscdk.DockerImage_FromRegistry(jsii.String("golang:1.24")),
				Command: &[]*string{
					jsii.String("bash"),
					jsii.String("-c"),
					jsii.String(strings.Join([]string{
						"export GOCACHE=/tmp/go-cache",
						"export GOPATH=/tmp/go-path",
						"CGO_ENABLED=0",
						"GOOS=linux",
						"GOARCH=arm64",
						"go build -tags lambda.norpc -o /asset-output/bootstrap main.go",
					}, " && ")),
				},
			},
		}),
		Handler:      jsii.String("bootstrap"),
		Architecture: awslambda.Architecture_ARM_64(),
		Environment: &map[string]*string{
			"DOWNSTREAM_FUNCTION_NAME": props.Downstream.FunctionName(),
			"HITS_TABLE_NAME":          table.TableName(),
		},
	})

	table.GrantReadWriteData(handler)

	props.Downstream.GrantInvoke(handler)

	return &hitCounter{this, handler, table}
}

func (h *hitCounter) Handler() awslambda.IFunction {
	return h.handler
}

func (h *hitCounter) Table() awsdynamodb.Table {
	return h.table
}
