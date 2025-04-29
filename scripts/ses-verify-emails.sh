#!/bin/bash

# $AWS_PROFILEがセットされているかチェック
if [ -z "$AWS_PROFILE" ]; then
  echo "エラー: AWS_PROFILE環境変数が設定されていません。"
  echo "以下のようにAWS CLIのプロファイルを設定してください。"
  echo "  export AWS_PROFILE=your-profile-name"
  exit 1
fi

echo "現在のAWSプロファイルは${AWS_PROFILE}です。"

# プロファイルに間違いがないか確認
read -p "このプロファイルで正しいですか？ [y/N]: " confirm
if [[ "$confirm" != [yY] ]]; then
  echo "処理を中断します。AWSプロファイルを確認してください。"
  exit 1
fi

# チェック：引数がない場合は使い方を表示
if [ "$#" -eq 0 ]; then
  echo "使い方: $0 email1@example.com email2@example.com ..."
  exit 1
fi

# 引数として渡されたメールアドレスを出力
echo "以下のメールアドレスの検証を開始します:"
for email in "$@"; do
  echo " - $email"
done

# 失敗したメールアドレスを格納する配列
failed_emails=()

# 各引数（メールアドレス）に対して SES の登録を実行
for email in "$@"; do
  echo "検証中: $email"
  if ! aws ses verify-email-identity --email-address "$email"; then
    echo "$emailの検証に失敗しました。"
    failed_emails+=("$email")
  fi
done

echo "メールアドレスの検証処理が完了しました。"

# 検証したメールアドレスの総数と失敗数を出力
total_emails=$#
failed_count=${#failed_emails[@]}
echo "検証したメールアドレスの総数: $total_emails"
echo "検証に失敗したメールアドレスの数: $failed_count"

# 失敗したメールアドレスがあれば表示
if [ $failed_count -gt 0 ]; then
  echo "以下のメールアドレスの検証に失敗しました："
  for failed in "${failed_emails[@]}"; do
    echo " - $failed"
  done
  echo "失敗したメールアドレスを確認してください。"
else
  echo "すべてのメールアドレスの検証が成功しました。"
fi

echo "AWSコンソールまたは受信した検証メールでステータスを確認してください。"