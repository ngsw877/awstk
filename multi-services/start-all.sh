#!/bin/bash

# スクリプトの場所を取得（他のスクリプトを呼び出すために使用）
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )" || {
  echo "ディレクトリ移動失敗" >&2
  exit 1
}

# --- ヘルプメッセージ ---
usage() {
  echo "Usage: $0 -c <ecs-cluster-name> -s <ecs-service-name> -d <aurora-cluster-identifier> [-m <min-capacity>] [-M <max-capacity>] [-P <aws-profile>]" >&2
  echo "  -c : ECSクラスター名" >&2
  echo "  -s : ECSサービス名" >&2
  echo "  -d : Aurora DBクラスター識別子" >&2
  echo "  -m : 最小キャパシティ (デフォルト: 1)" >&2
  echo "  -M : 最大キャパシティ (デフォルト: 2)" >&2
  echo "  -P : AWSプロファイル (任意)" >&2
  exit 1
}

# --- 引数が1つも与えられなかった場合はusage関数を実行して終了 ---
if [ $# -eq 0 ]; then
  echo "エラー: 引数が指定されていません。" >&2
  usage
fi

# --- 変数初期化 ---
ECS_CLUSTER_NAME=""
ECS_SERVICE_NAME=""
AURORA_CLUSTER_ID=""
MIN_CAPACITY="1"
MAX_CAPACITY="2"
PROFILE_OPT=""
AWS_PROFILE_PARAM=""
MIN_CAPACITY_PARAM=""
MAX_CAPACITY_PARAM=""

# --- オプション解析 ---
while getopts "c:s:d:m:M:P:" opt; do
  case $opt in
    c) ECS_CLUSTER_NAME="${OPTARG}" ;;
    s) ECS_SERVICE_NAME="${OPTARG}" ;;
    d) AURORA_CLUSTER_ID="${OPTARG}" ;;
    m) MIN_CAPACITY="${OPTARG}"; MIN_CAPACITY_PARAM="-m ${MIN_CAPACITY}" ;;
    M) MAX_CAPACITY="${OPTARG}"; MAX_CAPACITY_PARAM="-M ${MAX_CAPACITY}" ;;
    P) AWS_PROFILE="${OPTARG}"; AWS_PROFILE_PARAM="-P ${AWS_PROFILE}" ;;
    *) usage ;;
  esac
done

# --- 必須パラメータチェック ---
if [ -z "$ECS_CLUSTER_NAME" ] || [ -z "$ECS_SERVICE_NAME" ] || [ -z "$AURORA_CLUSTER_ID" ]; then
  echo "エラー: ECSクラスター名、ECSサービス名、Aurora DBクラスター識別子は必須です。" >&2
  usage
fi

# --- Aurora DBクラスターの起動 ---
echo "=== Aurora DBクラスターの起動処理を開始します ==="
if ! "${SCRIPT_DIR}/../aurora/start-cluster.sh" -d "${AURORA_CLUSTER_ID}" ${AWS_PROFILE_PARAM}; then
  echo "⚠️ Aurora DBクラスターの起動に失敗しましたが、処理を続行します。"
fi

# --- Auroraの起動完了を待機 ---
echo -e "\n⏳ Aurora DBクラスターの起動完了を待っています..."
while true; do
  status=$(aws rds describe-db-clusters --db-cluster-identifier "${AURORA_CLUSTER_ID}" ${AWS_PROFILE_PARAM} --query 'DBClusters[0].Status' --output text 2>/dev/null)
  if [ "$status" = "available" ]; then
    echo "✅ Aurora DBクラスターが起動しました！"
    break
  fi
  echo "  Aurora起動待ち...（現在の状態: $status）"
  sleep 30
  # Ctrl+Cで抜けられるようにする
  trap 'echo "\n⏹️ 待機を中断しました。"; exit 1' INT
done

# --- ECSサービスの起動 ---
echo -e "\n=== ECSサービスの起動処理を開始します ==="
if ! "${SCRIPT_DIR}/../ecs/start-tasks.sh" -c "${ECS_CLUSTER_NAME}" -s "${ECS_SERVICE_NAME}" ${MIN_CAPACITY_PARAM} ${MAX_CAPACITY_PARAM} ${AWS_PROFILE_PARAM}; then
  echo "⚠️ ECSサービスの起動に失敗しました。"
fi

echo -e "\n🎉 全ての起動処理が完了しました！"
echo "ℹ️ Aurora DBクラスターの起動完了まで数分かかる場合があります。"
exit 0 