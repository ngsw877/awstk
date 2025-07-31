package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Event はCloudFormationカスタムリソースのイベント構造体
// CDK Provider frameworkから呼び出される際のリクエスト情報を含む
type Event struct {
	RequestType           string                 `json:"RequestType"`
	RequestId             string                 `json:"RequestId"`
	StackId               string                 `json:"StackId"`
	LogicalResourceId     string                 `json:"LogicalResourceId"`
	PhysicalResourceId    string                 `json:"PhysicalResourceId,omitempty"`
	ResourceType          string                 `json:"ResourceType"`
	ResourceProperties    map[string]interface{} `json:"ResourceProperties"`
	OldResourceProperties map[string]interface{} `json:"OldResourceProperties,omitempty"`
}

// Response はCloudFormationカスタムリソースのレスポンス構造体
// 処理結果をCDK Provider frameworkに返却する
type Response struct {
	PhysicalResourceId string                 `json:"PhysicalResourceId"`
	Data               map[string]interface{} `json:"Data,omitempty"`
}

// handler はLambda関数のメインエントリーポイント
// CloudFormationカスタムリソースのライフサイクルイベント（Create/Update/Delete）を処理し、
// S3バケットにテストデータを作成する
func handler(ctx context.Context, event Event) (Response, error) {
	log.Printf("Received event: %+v", event)

	// PhysicalResourceIdの取得または生成
	// CloudFormationがリソースを一意に識別するために必要
	physicalResourceId := event.PhysicalResourceId
	if physicalResourceId == "" {
		// 初回作成時は新しいIDを生成（バケット名とパターンを組み合わせて一意性を確保）
		bucketName := event.ResourceProperties["BucketName"].(string)
		pattern := event.ResourceProperties["Pattern"].(string)
		physicalResourceId = fmt.Sprintf("S3DataCreator-%s-%s", bucketName, pattern)
	}

	if event.RequestType != "Create" {
		// Create以外（Update/Delete）は何もしない
		// - Update: データの重複作成を防ぐため処理をスキップ
		// - Delete: S3バケット削除時に自動的にオブジェクトも削除されるため処理不要
		log.Printf("%s request received, no action needed", event.RequestType)
		return Response{
			PhysicalResourceId: physicalResourceId,
			Data: map[string]interface{}{
				"Status": "Success",
				"Message": fmt.Sprintf("No action required for %s", event.RequestType),
			},
		}, nil
	}

	// ========== 以下、Create時のみ実行される処理 ==========
	
	// CDKから渡されたプロパティを取得
	bucketName := event.ResourceProperties["BucketName"].(string)
	pattern := event.ResourceProperties["Pattern"].(string)

	// ObjectCountの型変換処理
	// Provider frameworkが数値を文字列として渡すことがあるため、両方のケースに対応
	var objectCount int
	switch v := event.ResourceProperties["ObjectCount"].(type) {
	case float64:
		objectCount = int(v)
	case string:
		_, err := fmt.Sscanf(v, "%d", &objectCount)
		if err != nil {
			objectCount = 10 // パース失敗時はデフォルト値
		}
	default:
		objectCount = 10 // デフォルト値
	}

	// AWS SDKクライアントの初期化
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return Response{PhysicalResourceId: physicalResourceId}, fmt.Errorf("failed to load config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	// パターンに応じて異なるテストデータを作成
	switch pattern {
	case "normal":
		// 通常のオブジェクト（指定された個数のファイルを作成）
		err = createNormalObjects(ctx, client, bucketName, objectCount)
	case "versioned":
		// バージョニングテスト用（同一キーで複数バージョンを作成）
		err = createVersionedObjects(ctx, client, bucketName)
	case "deleted-markers":
		// 削除マーカーテスト用（オブジェクト作成後に削除）
		err = createDeletedMarkers(ctx, client, bucketName)
	default:
		err = fmt.Errorf("unknown pattern: %s", pattern)
	}

	if err != nil {
		log.Printf("Error creating data: %v", err)
		return Response{PhysicalResourceId: physicalResourceId}, err
	}

	// 成功レスポンスを作成して返却
	log.Printf("Successfully handled %s request for pattern %s, returning physicalResourceID: %s", event.RequestType, pattern, physicalResourceId)
	return Response{
		PhysicalResourceId: physicalResourceId,
		Data: map[string]interface{}{
			"Status": "Success",
			"Message": fmt.Sprintf("Created %s data in bucket %s", pattern, bucketName),
		},
	}, nil
}

// createNormalObjects は指定された個数の通常オブジェクトをS3バケットに作成する
// awstk s3 cleanupコマンドの基本動作テスト用
func createNormalObjects(ctx context.Context, client *s3.Client, bucketName string, count int) error {
	log.Printf("Creating %d objects in bucket %s", count, bucketName)
	for i := 0; i < count; i++ {
		key := fmt.Sprintf("data/object-%04d.txt", i)
		body := fmt.Sprintf("This is demo object number %d", i)

		_, err := client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: &bucketName,
			Key:    &key,
			Body:   strings.NewReader(body),
		})
		if err != nil {
			return fmt.Errorf("failed to put object %s: %w", key, err)
		}

		// 進捗ログ（100個ごと）
		if (i+1)%100 == 0 {
			log.Printf("Created %d/%d objects", i+1, count)
		}
	}
	log.Printf("Successfully created %d objects", count)
	return nil
}

// createVersionedObjects はバージョニングが有効なバケットでのテスト用データを作成する
// 同一キーに対して複数回アップロードすることで、複数バージョンを生成
func createVersionedObjects(ctx context.Context, client *s3.Client, bucketName string) error {
	log.Printf("Creating versioned objects in bucket %s", bucketName)
	// 同じキーで3バージョン作成（バージョニング機能のテスト用）
	key := "versioned/document.txt"
	for v := 1; v <= 3; v++ {
		body := fmt.Sprintf("This is version %d of the document", v)
		_, err := client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: &bucketName,
			Key:    &key,
			Body:   strings.NewReader(body),
		})
		if err != nil {
			return fmt.Errorf("failed to put version %d: %w", v, err)
		}
		log.Printf("Created version %d", v)
	}
	log.Printf("Successfully created 3 versions")
	return nil
}

// createDeletedMarkers は削除マーカーのテスト用データを作成する
// バージョニングが有効なバケットでオブジェクトを削除すると、
// 実際には削除マーカーが作成される（完全削除ではない）
func createDeletedMarkers(ctx context.Context, client *s3.Client, bucketName string) error {
	log.Printf("Creating deleted markers in bucket %s", bucketName)
	// オブジェクトを作成してから削除することで削除マーカーを生成
	for i := 0; i < 3; i++ {
		key := fmt.Sprintf("deleted/file-%d.txt", i)
		body := fmt.Sprintf("This file %d will be deleted", i)

		_, err := client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: &bucketName,
			Key:    &key,
			Body:   strings.NewReader(body),
		})
		if err != nil {
			return fmt.Errorf("failed to put object %s: %w", key, err)
		}
		log.Printf("Created object %s", key)

		// 削除（削除マーカーが作成される）
		_, err = client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: &bucketName,
			Key:    &key,
		})
		if err != nil {
			return fmt.Errorf("failed to delete object %s: %w", key, err)
		}
		log.Printf("Deleted object %s (created delete marker)", key)
	}
	log.Printf("Successfully created 3 delete markers")
	return nil
}

func main() {
	lambda.Start(handler)
}
