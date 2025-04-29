#!/bin/bash

# スクリプトの場所を取得（他のスクリプトを呼び出すために使用）
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )" || {
  echo "ディレクトリ移動失敗" >&2
  exit 1
}

# --- ヘルプメッセージ ---
usage() {
  echo "Usage: $0 -c <ecs-cluster-name> -s <ecs-service-name> -d <aurora-cluster-identifier> [-P <aws-profile>]" >&2
  echo "  -c : ECSクラスター名" >&2
  echo "  -s : ECSサービス名" >&2
  echo "  -d : Aurora DBクラスター識別子" >&2
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
PROFILE_OPT=""
AWS_PROFILE_PARAM=""

# --- オプション解析 ---
while getopts "c:s:d:P:" opt; do
  case $opt in
    c) ECS_CLUSTER_NAME="${OPTARG}" ;;
    s) ECS_SERVICE_NAME="${OPTARG}" ;;
    d) AURORA_CLUSTER_ID="${OPTARG}" ;;
    P) AWS_PROFILE="${OPTARG}"; AWS_PROFILE_PARAM="-P ${AWS_PROFILE}" ;;
    *) usage ;;
  esac
done

# --- 必須パラメータチェック ---
if [ -z "$ECS_CLUSTER_NAME" ] || [ -z "$ECS_SERVICE_NAME" ] || [ -z "$AURORA_CLUSTER_ID" ]; then
  echo "エラー: ECSクラスター名、ECSサービス名、Aurora DBクラスター識別子は必須です。" >&2
  usage
fi

# --- ECSサービスの停止 ---
echo "=== ECSサービスの停止処理を開始します ==="
if ! "${SCRIPT_DIR}/../ecs/stop-tasks.sh" -c "${ECS_CLUSTER_NAME}" -s "${ECS_SERVICE_NAME}" ${AWS_PROFILE_PARAM}; then
  echo "⚠️ ECSサービスの停止に失敗しましたが、処理を続行します。"
fi

# --- Aurora DBクラスターの停止 ---
echo -e "\n=== Aurora DBクラスターの停止処理を開始します ==="
if ! "${SCRIPT_DIR}/../aurora/stop-cluster.sh" -d "${AURORA_CLUSTER_ID}" ${AWS_PROFILE_PARAM}; then
  echo "⚠️ Aurora DBクラスターの停止に失敗しました。"
fi

echo -e "\n🎉 全ての停止処理が完了しました！"
exit 0 