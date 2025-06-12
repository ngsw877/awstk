package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	awslambda "github.com/aws/aws-sdk-go-v2/service/lambda"
)

var (
	dynamoClient *dynamodb.Client
	lambdaClient *awslambda.Client
)

func init() {
	// AWS設定をロード（Lambda実行ロールが自動的に使われる）
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("AWS設定の読み込みに失敗: %v", err)
	}

	// クライアントを初期化（関数の再利用時に効率的）
	dynamoClient = dynamodb.NewFromConfig(cfg)
	lambdaClient = awslambda.NewFromConfig(cfg)
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// リクエストをログ出力
	requestJSON, _ := json.MarshalIndent(request, "", "  ")
	fmt.Printf("request: %s\n", string(requestJSON))

	// DynamoDBのhitsを更新
	tableName := os.Getenv("HITS_TABLE_NAME")
	_, err := dynamoClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"path": &types.AttributeValueMemberS{Value: request.Path},
		},
		UpdateExpression: aws.String("ADD hits :incr"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":incr": &types.AttributeValueMemberN{Value: "1"},
		},
	})
	if err != nil {
		return events.APIGatewayProxyResponse{}, fmt.Errorf("DynamoDB更新に失敗: %w", err)
	}

	// 下流のLambda関数を呼び出し
	downstreamFunctionName := os.Getenv("DOWNSTREAM_FUNCTION_NAME")
	payload, _ := json.Marshal(request)

	invokeResult, err := lambdaClient.Invoke(ctx, &awslambda.InvokeInput{
		FunctionName: aws.String(downstreamFunctionName),
		Payload:      payload,
	})
	if err != nil {
		return events.APIGatewayProxyResponse{}, fmt.Errorf("下流Lambda呼び出しに失敗: %w", err)
	}

	fmt.Printf("downstream response: %s\n", string(invokeResult.Payload))

	// 下流Lambda関数のレスポンスをパース
	var downstreamResponse events.APIGatewayProxyResponse
	err = json.Unmarshal(invokeResult.Payload, &downstreamResponse)
	if err != nil {
		return events.APIGatewayProxyResponse{}, fmt.Errorf("下流レスポンスのパースに失敗: %w", err)
	}

	return downstreamResponse, nil
}

func main() {
	lambda.Start(handler)
}
