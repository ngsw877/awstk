#!/bin/bash

# --- ヘルプメッセージ ---
usage() {
  echo "Usage: $0 [-S <stack-name> | -c <ecs-cluster-name> -s <ecs-service-name>] [-m <min-capacity>] [-M <max-capacity>] [-P <aws-profile>]" >&2
  echo "  -S : CloudFormation スタック名（任意）" >&2
  echo "  -c : ECSクラスター名 (-S が指定されていない場合に必須)" >&2
  echo "  -s : ECSサービス名 (-S が指定されていない場合に必須)" >&2
  echo "  -m : 最小キャパシティ (デフォルト: 1)" >&2
  echo "  -M : 最大キャパシティ (デフォルト: 2)" >&2
  echo "  -P : AWSプロファイル (任意)" >&2
  exit 1
}

# スクリプトのディレクトリを取得
SCRIPT_DIR=$(dirname "$0")
# ヘルパースクリプトを読み込む
source "$SCRIPT_DIR/_get-ecs-info.sh"

# --- 引数が1つも与えられなかった場合はusage関数を実行して終了 ---
if [ $# -eq 0 ]; then
  echo "❌ エラー: 引数が指定されていません。" >&2
  usage
fi

# --- 変数初期化 ---
STACK_NAME=""
ECS_CLUSTER_NAME=""
ECS_SERVICE_NAME=""
MIN_CAPACITY="1"
MAX_CAPACITY="2"
PROFILE=""

# --- オプション解析 ---
while getopts "S:c:s:m:M:P:" opt; do
  case $opt in
    S) STACK_NAME="${OPTARG}" ;;
    c) ECS_CLUSTER_NAME="${OPTARG}" ;;
    s) ECS_SERVICE_NAME="${OPTARG}" ;;
    m) MIN_CAPACITY="${OPTARG}" ;;
    M) MAX_CAPACITY="${OPTARG}" ;;
    P) PROFILE="${OPTARG}" ;;
    *) usage ;;
  esac
done

# プロファイルが指定されていない場合、環境変数から取得を試みる
if [ -z "$PROFILE" ] && [ -n "$AWS_PROFILE" ]; then
  PROFILE="$AWS_PROFILE"
  echo "🔍 環境変数 AWS_PROFILE の値 '$PROFILE' を使用します"
fi

# プロファイルがどちらもセットされていなければエラー
if [ -z "$PROFILE" ]; then
  echo "❌ エラー: プロファイルが指定されていません。-PオプションまたはAWS_PROFILE環境変数を指定してね！" >&2
  exit 1
fi

# --- スタック名が指定されている場合、クラスターとサービスを自動検出 ---
if [ -n "$STACK_NAME" ]; then
  echo "🔍 CloudFormation スタック '$STACK_NAME' からリソースを検出しています..."
  # 共通関数を使ってスタックからクラスター名とサービス名を取得
  result=($(get_ecs_from_stack "$STACK_NAME" "$PROFILE"))
  ECS_CLUSTER_NAME=${result[0]}
  ECS_SERVICE_NAME=${result[1]}
  echo "🔍 検出されたクラスター: $ECS_CLUSTER_NAME"
  echo "🔍 検出されたサービス: $ECS_SERVICE_NAME"

elif [ -z "$ECS_CLUSTER_NAME" ] || [ -z "$ECS_SERVICE_NAME" ]; then
  echo "❌ エラー: スタック名が指定されていない場合は、クラスター名 (-c) とサービス名 (-s) が必須です。" >&2
  usage
fi

# --- 必須パラメータチェック ---
if [ -z "$ECS_CLUSTER_NAME" ] || [ -z "$ECS_SERVICE_NAME" ]; then
  echo "❌ エラー: ECSクラスター名とECSサービス名は必須です。" >&2
  usage
fi

# --- Fargate (ECSサービス) の起動 ---
echo "🔍 🚀 Fargate (ECSサービス: ${ECS_SERVICE_NAME}) のDesiredCountを${MIN_CAPACITY}～${MAX_CAPACITY}に設定します..."
if ! aws application-autoscaling register-scalable-target \
    --profile $PROFILE \
    --service-namespace ecs \
    --scalable-dimension ecs:service:DesiredCount \
    --resource-id "service/${ECS_CLUSTER_NAME}/${ECS_SERVICE_NAME}" \
    --min-capacity ${MIN_CAPACITY} \
    --max-capacity ${MAX_CAPACITY}; then
  echo "❌ Fargate (ECSサービス) の起動に失敗しました。" >&2
  exit 1
fi
echo "✅ Fargate (ECSサービス) のDesiredCountを設定しました。サービスが起動中です。"
exit 0 