package secretsmanager

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsSecretsManager "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// DeleteSecret deletes a secret immediately and without a recovery window.
// It depends on the concrete client type from AWS SDK.
func DeleteSecret(client *awsSecretsManager.Client, secretId string) error {
	input := &awsSecretsManager.DeleteSecretInput{
		SecretId:                   aws.String(secretId),
		ForceDeleteWithoutRecovery: aws.Bool(true), // 復旧期間なしで即時削除
	}

	// Use context.Background() as this is a background task not tied to a specific request.
	_, err := client.DeleteSecret(context.Background(), input)
	if err != nil {
		return fmt.Errorf("failed to delete secret %s: %w", secretId, err)
	}

	return nil
}
