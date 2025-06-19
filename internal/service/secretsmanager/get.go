package secretsmanager

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// GetSecretValues Secrets Managerからシークレット値を取得してMapで返す
func GetSecretValues(secretsClient *secretsmanager.Client, secretName string) (map[string]interface{}, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := secretsClient.GetSecretValue(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("シークレット取得に失敗: %w", err)
	}

	// シークレット値をJSONとしてパース
	var secretMap map[string]interface{}
	err = json.Unmarshal([]byte(*result.SecretString), &secretMap)
	if err != nil {
		return nil, fmt.Errorf("シークレットのJSON解析に失敗: %w", err)
	}

	return secretMap, nil
}
