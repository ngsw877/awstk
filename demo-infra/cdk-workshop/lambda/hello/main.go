package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// リクエストをログ出力
	requestJSON, _ := json.MarshalIndent(request, "", "  ")
	fmt.Printf("request: %s\n", string(requestJSON))

	// レスポンスを返す
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "text/plain",
		},
		Body: fmt.Sprintf("Hello, CDK! You've hit %s\n", request.Path),
	}, nil
}

func main() {
	lambda.Start(handler)
}
