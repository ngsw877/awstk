# SSM Delete Parameters サンプル

このディレクトリには `awstk ssm delete-params` コマンドで使用するサンプルファイルが含まれています。

## ファイル形式

`delete-params` コマンドは、1行に1つのパラメータ名を記載したテキストファイルを読み込みます。

- 空行は無視されます
- `#` で始まる行はコメントとして扱われ、無視されます
- パラメータ名は `/` で始まる必要があります

## 使用例

```bash
# ドライラン（削除対象を確認）
awstk ssm delete-params params.txt --dry-run

# 通常の削除（確認プロンプトあり）
awstk ssm delete-params params.txt

# 強制削除（確認プロンプトなし）
awstk ssm delete-params params.txt --force

# プレフィックスを付けて削除
# 例: params.txt内の "/database/host" → "/staging/database/host" として削除
awstk ssm delete-params params.txt --prefix /staging/
```

### --prefix オプションの動作

`--prefix` オプションを使用すると、ファイル内の各パラメータ名の前にプレフィックスが付加されます：

```bash
# params.txt の内容:
# /database/host
# /api/key

# コマンド実行:
awstk ssm delete-params params.txt --prefix /myapp/prod/

# 実際に削除されるパラメータ:
# /myapp/prod/database/host
# /myapp/prod/api/key
```

この機能により、同じ削除リストファイルを使って異なる環境のパラメータを管理できます。

## AWS CLIとの連携

既存のパラメータ一覧から削除対象を抽出する場合：

```bash
# 特定のプレフィックスを持つパラメータを抽出
aws ssm describe-parameters --query 'Parameters[?starts_with(Name, `/myapp/`)].Name' --output text > params.txt

# 削除実行
awstk ssm delete-params params.txt
```