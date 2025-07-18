package main

import (
	"testing"

	"cdk-workshop/hitcounter"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/jsii-runtime-go"
	"github.com/google/go-cmp/cmp"
)

func TestHitCounterConstruct(t *testing.T) {
	defer jsii.Close()
	defer func() {
		if r := recover(); r == nil {
			t.Error("Did not throw ReadCapacity error")
		}
	}()

	// GIVEN
	stack := awscdk.NewStack(nil, nil, nil)

	// WHEN
	testFn := awslambda.NewFunction(stack, jsii.String("TestFunction"), &awslambda.FunctionProps{
		Runtime: awslambda.Runtime_NODEJS_16_X(),
		Handler: jsii.String("hello.handler"),
		Code:    awslambda.Code_FromAsset(jsii.String("lambda"), nil),
	})
	hitcounter.NewHitCounter(stack, "MyTestConstruct", &hitcounter.HitCounterProps{
		Downstream:   testFn,
		ReadCapacity: 21, // This should trigger a panic
	})

	// THEN
	template := assertions.Template_FromStack(stack, nil)

	template.ResourceCountIs(jsii.String("AWS::DynamoDB::Table"), jsii.Number(1))

	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), &map[string]any{
		"SSESpecification": map[string]any{
			"SSEEnabled": true,
		},
	})

	envCapture := assertions.NewCapture(nil)
	template.HasResourceProperties(jsii.String("AWS::Lambda::Function"), map[string]any{
		"Environment": envCapture,
		"Handler":     "hitcounter.handler",
	})
	expectedEnv := &map[string]any{
		"Variables": map[string]any{
			"DOWNSTREAM_FUNCTION_NAME": map[string]any{
				"Ref": "TestFunction22AD90FC",
			},
			"HITS_TABLE_NAME": map[string]any{
				"Ref": "MyTestConstructHits24A357F0",
			},
		},
	}
	if !cmp.Equal(envCapture.AsObject(), expectedEnv) {
		t.Error(expectedEnv, envCapture.AsObject())
	}
}
