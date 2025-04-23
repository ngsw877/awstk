#!/bin/bash

# CloudFormationスタックからECSクラスターとサービス情報を取得する
get_ecs_from_stack() {
  local stack_name="$1"
  local profile="$2"
  
  if [ -z "$stack_name" ]; then
    echo "❌ エラー: スタック名が指定されていません" >&2
    exit 1
  fi
  
  if [ -z "$profile" ]; then
    profile="$AWS_PROFILE"
  fi
  if [ -z "$profile" ]; then
    echo "❌ エラー: プロファイルが指定されていません。引数またはAWS_PROFILE環境変数を指定してください" >&2
    exit 1
  fi
  
  # スタックからクラスター名を取得
  echo "🔍 スタック '$stack_name' からECSクラスターを検索中..." >&2
  local cluster_names=$(aws cloudformation describe-stack-resources \
    --stack-name "$stack_name" \
    --profile "$profile" \
    --query "StackResources[?ResourceType=='AWS::ECS::Cluster'].PhysicalResourceId" \
    --output text)
  
  # 改行で分割してクラスター名の配列を作成
  local IFS=$'\n'
  local cluster_array=()
  read -r -a cluster_array <<< "$cluster_names"
  
  # クラスターが見つからない場合はエラー
  if [ ${#cluster_array[@]} -eq 0 ] || [ -z "${cluster_array[0]}" ]; then
    echo "❌ エラー: スタック '$stack_name' からECSクラスターを検出できませんでした" >&2
    exit 1
  fi
  
  # 複数のクラスターがある場合は警告を表示
  if [ ${#cluster_array[@]} -gt 1 ]; then
    echo "⚠️ 警告: スタック '$stack_name' に複数のECSクラスターが見つかりました。最初のクラスターを使用します:" >&2
    for (( i=0; i<${#cluster_array[@]}; i++ )); do
      if [ $i -eq 0 ]; then
        echo " * ${cluster_array[$i]} (使用するクラスター)" >&2
      else
        echo " * ${cluster_array[$i]}" >&2
      fi
    done
  fi
  
  # 最初のクラスターを使用
  local cluster_name="${cluster_array[0]}"
  
  # スタックからサービス名を取得
  echo "🔍 スタック '$stack_name' からECSサービスを検索中..." >&2
  local service_resources=$(aws cloudformation describe-stack-resources \
    --stack-name "$stack_name" \
    --profile "$profile" \
    --query "StackResources[?ResourceType=='AWS::ECS::Service'].PhysicalResourceId" \
    --output text)
  
  # 改行で分割してサービス名の配列を作成
  local service_array=()
  read -r -a service_array <<< "$service_resources"
  
  # サービスが見つからない場合はエラー
  if [ ${#service_array[@]} -eq 0 ] || [ -z "${service_array[0]}" ]; then
    echo "❌ エラー: スタック '$stack_name' からECSサービスを検出できませんでした" >&2
    exit 1
  fi
  
  # サービス名を抽出（形式: arn:aws:ecs:REGION:ACCOUNT:service/CLUSTER/SERVICE_NAME）
  local service_name=$(echo "${service_array[0]}" | awk -F'/' '{print $NF}')
  
  # 複数のサービスがある場合は警告を表示
  if [ ${#service_array[@]} -gt 1 ]; then
    echo "⚠️ 警告: スタック '$stack_name' に複数のECSサービスが見つかりました。最初のサービスを使用します:" >&2
    for (( i=0; i<${#service_array[@]}; i++ )); do
      local service=$(echo "${service_array[$i]}" | awk -F'/' '{print $NF}')
      if [ $i -eq 0 ]; then
        echo " * $service (使用するサービス)" >&2
      else
        echo " * $service" >&2
      fi
    done
  fi
  
  # 配列で返す
  echo "$cluster_name $service_name"
}

# ECSサービスからタスク定義とネットワーク設定を取得する
get_service_details() {
  local cluster_name="$1"
  local service_name="$2"
  local profile="$3"
  
  if [ -z "$cluster_name" ] || [ -z "$service_name" ]; then
    echo "❌ エラー: クラスター名とサービス名は必須です" >&2
    exit 1
  fi
  if [ -z "$profile" ]; then
    echo "❌ エラー: プロファイルが指定されていません" >&2
    exit 1
  fi
  
  echo "🔍 サービス '$service_name' の詳細を取得中..."
  
  # サービスからタスク定義を取得
  local task_definition=$(aws ecs describe-services \
    --cluster "$cluster_name" \
    --services "$service_name" \
    --profile "$profile" \
    --query "services[0].taskDefinition" \
    --output text)
  
  if [ -z "$task_definition" ] || [ "$task_definition" == "None" ]; then
    echo "❌ エラー: タスク定義の取得に失敗しました" >&2
    exit 1
  fi
  
  # サービスからネットワーク設定を取得
  local network_configuration=$(aws ecs describe-services \
    --cluster "$cluster_name" \
    --services "$service_name" \
    --profile "$profile" \
    --query "services[0].networkConfiguration.awsvpcConfiguration" \
    --output json)
  
  # 結果を出力
  echo "TASK_DEFINITION=$task_definition"
  echo "NETWORK_CONFIGURATION='$network_configuration'"
  echo "✅ サービス詳細情報の取得が完了しました"
}

# サービス名が指定されていない場合、クラスターから最初のサービスを取得
get_first_service() {
  local cluster_name="$1"
  local profile="$2"
  
  if [ -z "$cluster_name" ]; then
    echo "❌ エラー: クラスター名は必須です" >&2
    exit 1
  fi
  if [ -z "$profile" ]; then
    echo "❌ エラー: プロファイルが指定されていません" >&2
    exit 1
  fi
  
  echo "🔍 クラスター '$cluster_name' からサービスを検索中..."
  
  # クラスターから唯一のサービスを取得
  local service_name=$(aws ecs list-services \
    --cluster "$cluster_name" \
    --profile "$profile" \
    --query 'serviceArns[0]' \
    --output text | awk -F'/' '{print $NF}')
  
  if [ -z "$service_name" ] || [ "$service_name" == "None" ]; then
    echo "❌ エラー: クラスター '$cluster_name' にサービスが見つかりません" >&2
    exit 1
  fi
  
  echo "SERVICE_NAME=$service_name"
  echo "✅ サービス '$service_name' を検出しました"
}

# タスクIDを取得する
get_running_task() {
  local cluster_name="$1"
  local service_name="$2"
  local profile="$3"
  
  if [ -z "$cluster_name" ] || [ -z "$service_name" ]; then
    echo "❌ エラー: クラスター名とサービス名は必須です" >&2
    exit 1
  fi
  if [ -z "$profile" ]; then
    echo "❌ エラー: プロファイルが指定されていません" >&2
    exit 1
  fi
  
  echo "🔍 実行中のタスクを検索中..." >&2
  
  # タスクIDを取得
  local task_id=$(aws ecs list-tasks \
    --cluster "$cluster_name" \
    --service-name "$service_name" \
    --profile "$profile" \
    --query 'taskArns[0]' \
    --output text)
  
  if [ -z "$task_id" ] || [ "$task_id" == "None" ]; then
    echo "❌ エラー: クラスター '$cluster_name' のサービス '$service_name' で実行中のタスクが見つかりませんでした" >&2
    exit 1
  fi
  
  echo "✅ 実行中のタスク '$task_id' を検出しました" >&2
  echo -n "$task_id"
}

# もし直接実行された場合はヘルプメッセージを表示
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  cat <<EOF
使用方法: 
  このスクリプトは、他のスクリプトからソースとして読み込むために設計されています。
  
  例: 
  source $(basename "${BASH_SOURCE[0]}") 
  
提供される関数:
  get_ecs_from_stack <stack-name> [aws-profile]
    - CloudFormationスタックからECSクラスターとサービス情報を取得
  
  get_service_details <cluster-name> <service-name> [aws-profile]
    - ECSサービスからタスク定義とネットワーク設定を取得
  
  get_first_service <cluster-name> [aws-profile]
    - クラスターから最初のサービスを取得
  
  get_running_task <cluster-name> <service-name> [aws-profile]
    - サービスの実行中タスクIDを取得
EOF
fi 