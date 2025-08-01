# S3 Cleanup Demo Stack

S3 cleanupコマンドのデモ用CDKスタックです。

## デモパターン

1. **空のバケット** (`awstk-s3cleanupdemo-empty-bucket`)
   - オブジェクトが一つもないバケット

2. **通常バケット** (`awstk-s3cleanupdemo-normal-bucket`) 
   - 10個のオブジェクトを含むバケット
   - demo-dataディレクトリの内容をデプロイ

3. **ネストされたフォルダ構造** (`awstk-s3cleanupdemo-nested-bucket`)
   - deep/nested/folder/ 配下にファイルを配置

4. **バージョニング有効バケット（複数バージョン）** (`awstk-s3cleanupdemo-versioned-bucket`)
   - バージョニングが有効
   - Lambda関数により同じキーで3バージョンを作成

5. **バージョニング有効バケット（削除マーカー）** (`awstk-s3cleanupdemo-deleted-marker-bucket`)
   - バージョニングが有効
   - Lambda関数によりオブジェクトを作成後削除（削除マーカーを作成）

6. **大量オブジェクトバケット** (`awstk-s3cleanupdemo-large-bucket`)
   - 1200個のオブジェクトを含む（ページネーションテスト用）

## ファイルアップロードの仕組み

CDKでは以下の2つの方法でS3バケットにファイルを自動的にアップロードしています：

### 1. BucketDeployment（静的ファイル用）
- `demo-data/`ディレクトリ内のファイルを自動的にS3にアップロード
- CDKがデプロイ時にLambda関数を作成し、ファイルをコピー
- 通常バケットとネストされたフォルダ構造のバケットで使用

```go
awss3deployment.NewBucketDeployment(stack, jsii.String("DeployTestData"), &awss3deployment.BucketDeploymentProps{
    Sources: &[]awss3deployment.ISource{
        awss3deployment.Source_Asset(jsii.String("./demo-data")),
    },
    DestinationBucket: normalBucket,
})
```

### 2. Custom Resource + Lambda（動的データ用）
- Go言語で書かれたLambda関数（`lambda/data-creator.go`）が実行される
- CloudFormationのカスタムリソースとして、スタック作成時に自動実行
- バージョニング、削除マーカー、大量データの作成に使用

```go
awscdk.NewCustomResource(stack, jsii.String("LargeData"), &awscdk.CustomResourceProps{
    ServiceToken: provider.ServiceToken(),
    Properties: &map[string]interface{}{
        "BucketName":  largeBucket.BucketName(),
        "ObjectCount": 1200,
        "Pattern":     "normal",
    },
})
```

## デプロイ方法

```bash
# 依存関係のインストール
go mod tidy

# Lambda関数の依存関係
cd lambda && go mod tidy && cd ..

# CDKデプロイ
cdk deploy
```

## クリーンアップデモ

デプロイ後、CloudFormation出力に表示されるコマンドでクリーンアップ機能をデモ：

```bash
# 全てのデモバケットを削除（CloudFormation出力に表示）
awstk s3 cleanup --filter "awstk-s3cleanupdemo-"
```

または、個別にバケット名を指定：

```bash
# 特定のバケットのみクリーンアップ
awstk s3 cleanup --filter "awstk-s3cleanupdemo-large-bucket"
```

## 削除方法

```bash
cdk destroy
```
