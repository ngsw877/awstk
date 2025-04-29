#!/bin/bash

# --- ヘルプメッセージ ---
usage() {
  echo "Usage: $0 -d <aurora-cluster-identifier> [-P <aws-profile>]" >&2
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
AURORA_CLUSTER_ID=""
PROFILE_OPT=""

# --- オプション解析 ---
while getopts "d:P:" opt; do
  case $opt in
    d) AURORA_CLUSTER_ID="${OPTARG}" ;;
    P) AWS_PROFILE="${OPTARG}"; PROFILE_OPT="--profile ${AWS_PROFILE}" ;;
    *) usage ;;
  esac
done

# --- 必須パラメータチェック ---
if [ -z "$AURORA_CLUSTER_ID" ]; then
  echo "エラー: Aurora DBクラスター識別子は必須です。" >&2
  usage
fi

# --- Aurora DBクラスターの起動 ---
echo "Aurora DBクラスター (${AURORA_CLUSTER_ID}) を起動します..."
if ! aws rds start-db-cluster \
    ${PROFILE_OPT} \
    --db-cluster-identifier "${AURORA_CLUSTER_ID}"; then
  echo "❌ Aurora DBクラスターの起動に失敗しました。" >&2
  exit 1
fi
echo "✅ Aurora DBクラスターの起動を開始しました。起動完了まで数分かかる場合があります。"
exit 0 