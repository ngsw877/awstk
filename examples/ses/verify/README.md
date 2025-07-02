# SES Verify サンプル

このディレクトリには `awstk ses verify` コマンドで使用するサンプルファイルが含まれています。

## ファイル形式

`ses verify` コマンドは、1行に1つのメールアドレスを記載したテキストファイルを読み込みます。

- 空行は無視されます
- `#` で始まる行はコメントとして扱われ、無視されます
- メールアドレスは `@` を含む必要があります
- 重複するメールアドレスは自動的に除去されます（大文字小文字を正規化）

## 使用例

```bash
# 基本的な使用方法
awstk ses verify -f emails.txt

# AWS プロファイルとリージョンを指定
awstk ses verify -f emails.txt -P my-profile -R us-east-1
```

## 出力例

```
✅ 検証成功: 5件
  - dev-team@example.com
  - admin@example.org
  - test1@example.net
  - support@example.org
  - noreply@example.net

❌ 検証失敗: 1件
  - test2@example.invalid
```

## 注意事項

- SES の検証リクエストは実際のメールアドレスに確認メールを送信します
- 存在しないメールアドレスや無効なドメインは検証に失敗します
- 検証が成功したメールアドレスは、SES で送信可能になります
- AWS の SES サービス制限（送信制限、サンドボックス制限など）にご注意ください

## ファイルの準備

メールアドレス一覧は以下の方法で準備できます：

```bash
# CSV ファイルからメールアドレスを抽出
cut -d',' -f2 users.csv > emails.txt

# 既存の検証済みアドレス一覧を取得
aws ses list-verified-email-addresses --query 'VerifiedEmailAddresses' --output text > verified-emails.txt
```