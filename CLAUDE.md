# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## プロジェクト概要

**awstk** は Go 言語 + Cobra で実装した AWS リソース操作用 CLI です。
S3 / ECR / ECS / CloudFormation などをコマンドラインから一括管理・クリーンアップできることを目的としています。

### 技術スタック
- Go **1.24+**
- AWS SDK for Go v2
- Cobra CLI フレームワーク
- CDK for Go (検証用テンプレートは `demo-infra/` に同梱)

## ディレクトリ構成と責務

```
.
├── cmd/                    # CLI コマンド定義 (サービス別)
├── internal/
│   ├── aws/               # AWS 設定・共通クライアント
│   ├── cli/               # AWS CLI 実行処理
│   └── service/           # AWS SDK 操作ロジック
│       ├── cleanup/       # 横断クリーンアップ機能
│       ├── s3/            # S3 関連操作
│       ├── ecr/           # ECR 関連操作
│       ├── ecs/           # ECS 関連操作
│       ├── cfn/           # CloudFormation 関連操作
│       └── (その他サービス)/
├── demo-infra/            # CDK テンプレート
└── main.go               # エントリーポイント
```

### 責務分離の原則

| 層 | 責務 | 配置例 |
|---|------|--------|
| **CLI層** | コマンド定義・フラグ処理 | `cmd/<service>.go` |
| **サービス層** | AWS SDK操作・ビジネスロジック | `internal/service/<service>/` |
| **AWS層** | AWS設定・クライアント管理 | `internal/aws/` |
| **CLI実行層** | AWS CLI プロセス実行 | `internal/cli/` |

## 開発コマンド

```bash
# ビルド
go build -o awstk .

# テスト実行
go test ./...

# 特定パッケージのテスト
go test ./internal/service/s3

# フォーマット
go fmt ./...

# 静的解析
go vet ./...

# 実行例
./awstk ec2 ls -P your-profile -R ap-northeast-1 -S stack-name
```

## AI アシスタントへの共通指示

### 1. 変更の粒度とワークフロー

- **機能追加 / リファクタは 1 コミット 1 単位** を基本とするが、
  **設計方針の変更など大規模改修が必要な場合はまとめて変更しても構わない**。
- **大規模変更を行う前に必ず行うこと**
  1. 最新のディレクトリ構成をツリー形式で提示する
  2. どのファイルを変更・追加・削除するか、簡潔なプランを提示してユーザーの確認を得る

### 2. コミットメッセージ提案

- **AI はコミットメッセージを"提案"するだけであり、実際のコミット操作は行わない**
- メッセージは Conventional Commits の prefix を付け、日本語で要点を簡潔にまとめる
  - 例: `feat: S3 バケット削除コマンドを追加`
- Scope（`(cli)` や `(iac)` など）は必要に応じて付与してよいが必須ではない

## コードスタイル & 横断ルール

### 命名規則

- **パッケージ名** はすべて小文字・単数形（例: `service`, `aws`, `config`）
- **変数・関数・定数**
  - 非公開なら *camelCase*（先頭小文字）
  - 公開（エクスポート）する場合は *UpperCamelCase*（先頭大文字）
- **構造体・インターフェース** は UpperCamelCase（型名として読みやすくする）
- **略語の大文字化**
  - `AWS` → `Aws`, `HTTP` → `Http`, `ID` → `Id` など
  - 例: `AwsClient`, `HttpRequest`, `UserId`

### 可視性ポリシー (公開範囲)

- まず **private（小文字始まり）** で定義すること
- 他パッケージから参照される必要が生じた場合のみ public（大文字始まり）に昇格
- パブリック関数・構造体には必ず GoDoc コメントを付ける

### エラーハンドリング

- 下位層では `fmt.Errorf("%w", err)` でエラーをラップして伝播
  - 上位で `errors.Is/As` による判定が可能
- 上位層（CLIなど）でユーザ向けメッセージへ整形
- Sentinel error を使う場合は `errors.New` で定義し、比較は `errors.Is`

### コメント・ドキュメント (GoDoc)

- パブリック API には GoDoc 形式のコメントを付与
  - 1 行目に概要、2 行目以降に詳細
- コメントは「なぜ」を中心に書き、コードで「何をするか」を示す

### contextの使い方

- AWS SDK for Go v2 のメソッド呼び出し時は、`context.TODO()` ではなく `context.Background()` を使用すること。
  - 理由: `TODO()`は本来「未実装」や「後で置き換える」用途のため、実運用コードでは`Background()`を使う。

## 設計方針（要点）

1. **CLI 層とロジック層の分離**: `cmd/` → `internal/service/` の一方向依存
2. **サービス別パッケージ化**: 各 AWS サービスを独立したパッケージで管理
3. **型定義の分離**: 複雑な構造体は `types.go` に集約
4. **責務の明確化**: AWS SDK操作と AWS CLI実行処理を分離
5. **エラーハンドリング**: 下位層でラップ、上位層でユーザー向けに整形
6. **CDK独立性**: `demo-infra/` はアプリ本体と依存関係なし

## 主要機能例

- **横断クリーンアップ**: `cleanup all` でS3/ECRを一括削除
- **サービス別クリーンアップ**: `s3 cleanup`, `ecr cleanup` で個別削除
- **S3操作**: `s3 ls` (ツリー表示), `s3 gunzip` (.gz一括処理)
- **ECS操作**: Fargate コンテナへのシェル接続、サービス再起動
- **デモインフラ**: CDK テンプレート (`awstk-lab`, `cdk-workshop`) 自動デプロイ

## 基本的なファイル配置

- **CLI コマンド**: `cmd/<service>.go`
- **AWS SDK 操作**: `internal/service/s3/`, `internal/service/ecr/` など
- **AWS CLI 実行**: `internal/cli/`
- **AWS 設定**: `internal/aws/`
- **CDK テンプレート**: `demo-infra/`

## 共通フラグ

- `-P, --profile`: AWS プロファイル（未指定時は AWS_PROFILE 環境変数を使用）
- `-R, --region`: AWS リージョン
- `-S, --stack`: CloudFormation スタック名でリソースをフィルタリング