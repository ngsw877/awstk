#!/bin/bash

# --- ヘルプメッセージ ---
usage() {
  echo "使用方法: $0 [-r <リージョン>] [-P <AWSプロファイル>]" >&2
  echo "  -r : AWSリージョンを指定（デフォルト: AWS_REGION環境変数またはap-northeast-1）" >&2
  echo "  -P : AWSプロファイルを指定（デフォルト: AWS_PROFILE環境変数）" >&2
  echo "  -h : このヘルプメッセージを表示" >&2
  echo "" >&2
  echo "現在アクティブなCloudFormationスタックの一覧を表示します。" >&2
  echo "他のスクリプトでスタック名を指定する際に参照できます。" >&2
  exit 1
}

# --- 変数初期化 ---
REGION="${AWS_REGION:-ap-northeast-1}"
PROFILE="${AWS_PROFILE:-}"

# --- オプション解析 ---
while getopts "r:P:h" opt; do
  case $opt in
    r) REGION="$OPTARG" ;;
    P) PROFILE="$OPTARG" ;;
    h) usage ;;
    *) usage ;;
  esac
done

# --- プロファイル指定があればオプションに追加 ---
PROFILE_OPT=""
if [ -n "$PROFILE" ]; then
  PROFILE_OPT="--profile $PROFILE"
fi

# --- アクティブなスタックのステータスフィルター ---
ACTIVE_STATUSES="CREATE_COMPLETE UPDATE_COMPLETE UPDATE_ROLLBACK_COMPLETE ROLLBACK_COMPLETE IMPORT_COMPLETE"

# --- 表示ヘッダー ---
echo "🔍 CloudFormationスタック一覧" >&2
echo "  リージョン: $REGION" >&2
if [ -n "$PROFILE" ]; then
  echo "  プロファイル: $PROFILE" >&2
fi
echo "" >&2

# --- スタック一覧取得（1行ずつ表示） ---
aws cloudformation list-stacks \
  $PROFILE_OPT \
  --region $REGION \
  --stack-status-filter $ACTIVE_STATUSES \
  --query "StackSummaries[].StackName" \
  --output text | tr '\t' '\n'

# 終了コードの確認
exit_code=${PIPESTATUS[0]}
if [ $exit_code -ne 0 ]; then
  echo "❌ スタック一覧の取得に失敗しました。終了コード: $exit_code" >&2
  exit $exit_code
fi
