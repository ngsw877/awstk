#!/bin/bash

# タスク定義情報を取得
if [ "$#" -ne 3 ]; then
  echo "引数が不足しています。既存のタスク定義ファミリー名、コンテナ名、イメージ名を指定してください。"
  echo "使用例: $0 <既存のタスク定義ファミリー名> <コンテナ名> <イメージ名>"
  exit 1
fi

existing_task_def_family="$1"
container_name="$2"
image="$3"

task_def_file_name="../tmp/taskdef.json"

task_def=$(aws ecs describe-task-definition \
  --task-definition $existing_task_def_family \
  --query 'taskDefinition' \
  --output json)

# 取得したタスク定義情報をもとに、パラメータの上書きと整形を行い新しいタスク定義用jsonを作成する
new_task_def=$(\
  echo "${task_def}" |
  jq --arg IMAGE "${image}" --arg CONTAINER_NAME "${container_name}" \
  'del(
      .taskDefinitionArn,
      .revision,
      .status,
      .requiresAttributes,
      .compatibilities,
      .registeredAt,
      .registeredBy
    ) |
    .containerDefinitions |= map(
      select(.name != "web") |
      if .name == $CONTAINER_NAME then .image = $IMAGE else . end
    )
  ')

echo "${new_task_def}" > ${task_def_file_name}

# ユーザーに既存の更新か新規作成かを選択させる
echo "既存のタスク定義を更新しますか？新しいタスク定義を作成しますか？"
echo "1: 既存のタスク定義を更新"
echo "2: 新しいタスク定義を作成"
read -p "選択してください (1/2): " choice

if [ "$choice" -eq 1 ]; then
  # 既存のタスク定義を更新
  aws ecs register-task-definition \
    --family "${existing_task_def_family}" \
    --cli-input-json file://${task_def_file_name}
elif [ "$choice" -eq 2 ]; then
  # 新しいタスク定義を作成
  read -p "新しいタスク定義ファミリー名を入力してください: " new_task_def_family
  aws ecs register-task-definition \
    --family "${new_task_def_family}" \
    --cli-input-json file://${task_def_file_name}
else
  echo "無効な選択です。スクリプトを終了します。"
  exit 1
fi 