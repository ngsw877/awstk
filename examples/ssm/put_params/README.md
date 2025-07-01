# SSM put-params コマンドのサンプル

このディレクトリには、`awstk ssm put-params` コマンドで使用できるサンプルファイルが含まれています。

## ファイル説明

- `params.csv` - CSV形式のパラメータ定義ファイル
- `params.json` - JSON形式のパラメータ定義ファイル

## 使用例

```bash
# CSV形式で登録
awstk ssm put-params examples/ssm/put_params/params.csv

# JSON形式で登録（プレフィックス付き）
awstk ssm put-params examples/ssm/put_params/params.json --prefix /production/

# ドライラン（実際には登録しない）
awstk ssm put-params examples/ssm/put_params/params.csv --dry-run
```

## パラメータ型

- `String` - 通常の文字列
- `SecureString` - 暗号化される機密情報
- `StringList` - カンマ区切りのリスト