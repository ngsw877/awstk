package internal

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// GetSecretValues は指定したシークレット名から全ての値を取得して返す
func GetSecretValues(awsCtx AwsContext, secretName string) (map[string]interface{}, error) {
	cfg, err := LoadAwsConfig(awsCtx)
	if err != nil {
		return nil, err
	}
	client := secretsmanager.NewFromConfig(cfg)
	input := &secretsmanager.GetSecretValueInput{
		SecretId: &secretName,
	}
	result, err := client.GetSecretValue(context.Background(), input)
	if err != nil {
		return nil, err
	}

	if result.SecretString == nil {
		return nil, fmt.Errorf("SecretString is nil")
	}

	var secretMap map[string]interface{}
	if err := json.Unmarshal([]byte(*result.SecretString), &secretMap); err != nil {
		return nil, err
	}

	return secretMap, nil
}
