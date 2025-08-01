# ssm Commands

This document describes all `ssm` related commands.

## Table of Contents

- [awstk ssm](#awstk-ssm)
- [awstk ssm delete-params](#awstk-ssm-delete-params)
- [awstk ssm put-params](#awstk-ssm-put-params)
- [awstk ssm session](#awstk-ssm-session)

---

## awstk ssm

SSM関連の操作を行うコマンド群

### Synopsis

AWS SSMセッションマネージャーを利用したEC2インスタンスへの接続やParameter Storeの操作を行うCLIコマンド群です。

### Options

```
  -h, --help   help for ssm
```

### Options inherited from parent commands

```
  -P, --profile string   AWSプロファイル
  -R, --region string    AWSリージョン (default "ap-northeast-1")
```

### SEE ALSO

* [awstk](README.md)	 - AWS リソース管理用 CLI ツール
* [awstk ssm delete-params](ssm.md#awstk-ssm-delete-params)	 - ファイルからParameter Storeを一括削除
* [awstk ssm put-params](ssm.md#awstk-ssm-put-params)	 - ファイルからParameter Storeに一括登録
* [awstk ssm session](ssm.md#awstk-ssm-session)	 - EC2インスタンスにSSMで接続する

###### Auto generated by spf13/cobra on 1-Aug-2025

---

## awstk ssm delete-params

ファイルからParameter Storeを一括削除

### Synopsis

テキストファイルに記載されたパラメータ名のリストから、AWS Systems Manager Parameter Storeのパラメータを一括削除します。

ファイル形式:
  - 1行に1つのパラメータ名を記載
  - 空行と#で始まるコメント行は無視されます

例:
  awstk ssm delete-params params.txt
  awstk ssm delete-params params.txt --force
  awstk ssm delete-params params.txt --dry-run
  awstk ssm delete-params params.txt --prefix /myapp/  # 削除対象パラメータ名に/myapp/を付加


```
awstk ssm delete-params <file> [flags]
```

### Options

```
  -d, --dry-run         実際には削除せず、削除対象を確認
  -f, --force           確認プロンプトをスキップ
  -h, --help            help for delete-params
  -p, --prefix string   パラメータ名のプレフィックス
```

### Options inherited from parent commands

```
  -P, --profile string   AWSプロファイル
  -R, --region string    AWSリージョン (default "ap-northeast-1")
```

### SEE ALSO

* [awstk ssm](ssm.md)	 - SSM関連の操作を行うコマンド群

###### Auto generated by spf13/cobra on 1-Aug-2025

---

## awstk ssm put-params

ファイルからParameter Storeに一括登録

### Synopsis

CSV/JSONファイルからAWS Systems Manager Parameter Storeにパラメータを一括登録します。

対応ファイル形式:
  - CSV (.csv): name,value,type,description の形式
  - JSON (.json): {"parameters": [{"name": "...", "value": "...", "type": "...", "description": "..."}]}

例:
  awstk ssm put-params params.csv
  awstk ssm put-params params.json --prefix /myapp/
  awstk ssm put-params params.csv --dry-run


```
awstk ssm put-params <file> [flags]
```

### Options

```
  -d, --dry-run         実際には登録せず、登録内容を確認
  -h, --help            help for put-params
  -p, --prefix string   パラメータ名のプレフィックス
```

### Options inherited from parent commands

```
  -P, --profile string   AWSプロファイル
  -R, --region string    AWSリージョン (default "ap-northeast-1")
```

### SEE ALSO

* [awstk ssm](ssm.md)	 - SSM関連の操作を行うコマンド群

###### Auto generated by spf13/cobra on 1-Aug-2025

---

## awstk ssm session

EC2インスタンスにSSMで接続する

### Synopsis

指定したEC2インスタンスIDにSSMセッションで接続します。

例:
  awstk ssm session -i <ec2-instance-id> [-P <aws-profile>]
  awstk ssm session [-P <aws-profile>]  # インスタンス一覧から選択


```
awstk ssm session [flags]
```

### Options

```
  -h, --help                 help for session
  -i, --instance-id string   EC2インスタンスID（省略時は一覧から選択）
```

### Options inherited from parent commands

```
  -P, --profile string   AWSプロファイル
  -R, --region string    AWSリージョン (default "ap-northeast-1")
```

### SEE ALSO

* [awstk ssm](ssm.md)	 - SSM関連の操作を行うコマンド群

###### Auto generated by spf13/cobra on 1-Aug-2025

---

