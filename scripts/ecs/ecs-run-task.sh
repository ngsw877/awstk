#!/bin/bash
set -e

# 引数の解析
if [ $# -lt 2 ]; then
  echo "使用方法: $0 <クラスター名> <コンテナ名> [コマンド]"
  echo "例: $0 Hoge-Ecs-Cluster app 'echo hello'"
  exit 1
fi

CLUSTER_NAME=$1
CONTAINER_NAME=$2
COMMAND=${3:-'echo "デフォルトコマンドを実行しています"; sleep 5; echo "完了しました"'}

echo "🔍 サービス情報を取得中..."

# クラスターから唯一のサービスを取得
SERVICE_NAME=$(aws ecs list-services \
  --cluster "$CLUSTER_NAME" \
  --query 'serviceArns[0]' \
  --output text | awk -F'/' '{print $NF}')

if [ -z "$SERVICE_NAME" ]; then
  echo "❌ クラスター '$CLUSTER_NAME' にサービスが見つかりません"
  exit 1
fi

echo "🚀 サービス名: $SERVICE_NAME"

# サービスからタスク定義を取得
TASK_DEFINITION=$(aws ecs describe-services \
  --cluster "$CLUSTER_NAME" \
  --services "$SERVICE_NAME" \
  --query "services[0].taskDefinition" \
  --output text)

echo "📋 タスク定義: $TASK_DEFINITION"
echo "🧩 コンテナ名: $CONTAINER_NAME"

# サービスからネットワーク設定を取得
NETWORK_CONFIGURATION=$(aws ecs describe-services \
  --cluster "$CLUSTER_NAME" \
  --services "$SERVICE_NAME" \
  --query "services[0].networkConfiguration.awsvpcConfiguration" \
  --output json)

# ネットワーク設定からサブネットとセキュリティグループを抽出
SUBNETS=$(echo "$NETWORK_CONFIGURATION" | jq -r '.subnets | join(",")')
SECURITY_GROUPS=$(echo "$NETWORK_CONFIGURATION" | jq -r '.securityGroups | join(",")')
ASSIGN_PUBLIC_IP=$(echo "$NETWORK_CONFIGURATION" | jq -r '.assignPublicIp')

echo "🌐 サブネット: $SUBNETS"
echo "🔒 セキュリティグループ: $SECURITY_GROUPS"
echo "🌍 パブリックIP割り当て: $ASSIGN_PUBLIC_IP"

# コマンド内のダブルクォートをエスケープ
ESCAPED_COMMAND=$(echo "$COMMAND" | sed 's/"/\\"/g')

# JSONオーバーライドを作成
OVERRIDES=$(cat <<EOF
{
  "containerOverrides": [
    {
      "name": "$CONTAINER_NAME",
      "command": ["sh", "-c", "$ESCAPED_COMMAND"]
    }
  ]
}
EOF
)

# run-taskを実行
echo "🚀 タスクを実行中..."
TASK_ARN=$(aws ecs run-task \
  --cluster "$CLUSTER_NAME" \
  --launch-type FARGATE \
  --task-definition "$TASK_DEFINITION" \
  --network-configuration "awsvpcConfiguration={subnets=[$SUBNETS],securityGroups=[$SECURITY_GROUPS],assignPublicIp=$ASSIGN_PUBLIC_IP}" \
  --overrides "$OVERRIDES" \
  --query 'tasks[0].taskArn' \
  --output text)

echo "✅ タスクが開始されました: $TASK_ARN"

# タスクの完了を待機
echo "⏳ タスクの完了を待機中..."
aws ecs wait tasks-stopped \
  --cluster "$CLUSTER_NAME" \
  --tasks "$TASK_ARN"

# タスクの詳細情報を取得
TASK_DETAILS=$(aws ecs describe-tasks \
  --cluster "$CLUSTER_NAME" \
  --tasks "$TASK_ARN")

# 指定したコンテナの終了コードを取得
EXIT_CODE=$(echo "$TASK_DETAILS" | jq -r --arg CONTAINER_NAME "$CONTAINER_NAME" '.tasks[0].containers[] | select(.name == $CONTAINER_NAME) | .exitCode')

if [ -z "$EXIT_CODE" ] || [ "$EXIT_CODE" = "null" ]; then
  echo "❌ 指定したコンテナ '$CONTAINER_NAME' の終了コードが取得できませんでした"
  
  # すべてのコンテナの状態を表示
  echo "📊 タスク内のすべてのコンテナ:"
  echo "$TASK_DETAILS" | jq -r '.tasks[0].containers[] | "  - \(.name): 終了コード \(.exitCode // "不明"), 状態 \(.lastStatus)"'
else
  echo "🏁 コンテナ '$CONTAINER_NAME' が完了しました。終了コード: $EXIT_CODE"
fi

echo "🎉 完了"
