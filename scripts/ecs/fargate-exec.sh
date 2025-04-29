#!/bin/bash

# ヘルプメッセージを表示
usage() {
  echo "使用方法:"
  echo "  $0 [-P <aws-profile>] [-S <stack-name> | -c <cluster-name> -s <service-name>] [-t <container-name>]"
  echo ""
  echo "オプション:"
  echo "  -P : AWS プロファイル名（任意）"
  echo "  -S : CloudFormation スタック名（任意）"
  echo "  -c : ECS クラスター名（-S が指定されていない場合に必須）"
  echo "  -s : ECS サービス名（-S が指定されていない場合に必須）"
  echo "  -t : コンテナ名（デフォルト: app）"
  echo "  -h : このヘルプメッセージを表示"
  echo ""
  echo "例:"
  echo "  $0 -P myprofile -S my-stack"
  echo "  $0 -P myprofile -c my-cluster -s my-service"
  exit 1
}

# スクリプトのディレクトリを取得
SCRIPT_DIR=$(dirname "$0")
# ヘルパースクリプトを読み込む
source "$SCRIPT_DIR/_get-ecs-info.sh"

# パラメータを初期化
PROFILE=""
STACK_NAME=""
CLUSTER_NAME=""
SERVICE_NAME=""
CONTAINER_NAME="app"

# オプション引数を処理
while getopts "P:S:c:s:t:h" opt; do
  case $opt in
    P) PROFILE="$OPTARG" ;;
    S) STACK_NAME="$OPTARG" ;;
    c) CLUSTER_NAME="$OPTARG" ;;
    s) SERVICE_NAME="$OPTARG" ;;
    t) CONTAINER_NAME="$OPTARG" ;;
    h) usage ;;
    *) usage ;;
  esac
done

# PROFILEが未指定かつAWS_PROFILEがセットされている場合、PROFILEにAWS_PROFILEを使う
if [ -z "$PROFILE" ] && [ -n "$AWS_PROFILE" ]; then
  PROFILE="$AWS_PROFILE"
  echo "🔍 環境変数 AWS_PROFILE の値 '$PROFILE' を使用します"
fi

# プロファイルがどちらもセットされていなければエラー
if [ -z "$PROFILE" ]; then
  echo "❌ エラー: プロファイルが指定されていません。-PオプションまたはAWS_PROFILE環境変数を指定してね！" >&2
  exit 1
fi

# スタック名が指定されている場合、クラスターとサービスを自動検出
if [ -n "$STACK_NAME" ]; then
  # 共通関数を使ってスタックからクラスター名とサービス名を取得
  result=($(get_ecs_from_stack "$STACK_NAME" "$PROFILE"))
  CLUSTER_NAME=${result[0]}
  SERVICE_NAME=${result[1]}
  
  echo "🔍 検出されたクラスター: $CLUSTER_NAME"
  echo "🔍 検出されたサービス: $SERVICE_NAME"
  
elif [ -z "$CLUSTER_NAME" ] || [ -z "$SERVICE_NAME" ]; then
  echo "❌ エラー: スタック名が指定されていない場合は、クラスター名 (-c) とサービス名 (-s) が必須です" >&2
  usage
fi

# 共通関数を使って実行中のタスクを取得し、TASK_IDに代入
TASK_ID=$(get_running_task "$CLUSTER_NAME" "$SERVICE_NAME" "$PROFILE")

echo "🔍 コンテナ '$CONTAINER_NAME' に接続しています..."

# タスクにexecコマンドを実行
aws ecs execute-command \
  --region ap-northeast-1 \
  --cluster "$CLUSTER_NAME" \
  --task "$TASK_ID" \
  --container "$CONTAINER_NAME" \
  --interactive \
  --command "/bin/bash" \
  --profile "$PROFILE"