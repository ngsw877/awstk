#!/bin/bash

# 引数チェック
if [ "$#" -ne 2 ]; then
    echo "使用方法: $0 <AWSプロファイル> <検索文字列>"
    echo "例: $0 hoge-profile hoge-stack"
    exit 1
fi

AWS_PROFILE=$1
SEARCH_STRING=$2

# プロファイルの存在確認
if ! aws configure list-profiles | grep -q "^${AWS_PROFILE}$"; then
    echo "エラー: プロファイル '${AWS_PROFILE}' が見つかりません。"
    echo "利用可能なプロファイル:"
    aws configure list-profiles
    exit 1
fi

echo "AWS Profile: $AWS_PROFILE"
echo "検索文字列: $SEARCH_STRING"
echo "削除を開始します..."

# S3バケットの削除
echo "S3バケットの削除を開始..."
buckets=$(aws s3api list-buckets --profile $AWS_PROFILE --query "Buckets[?contains(Name, '$SEARCH_STRING')].Name" --output text)

if [ -z "$buckets" ]; then
    echo "  検索文字列 '${SEARCH_STRING}' にマッチするS3バケットは見つかりませんでした。"
else
    for bucket in $buckets; do
        echo "バケット $bucket を空にして削除中..."
        
        # バージョニングの状態を確認
        versioning_status=$(aws s3api get-bucket-versioning --bucket $bucket --profile $AWS_PROFILE --query 'Status' --output text)

        if [ "$versioning_status" == "Enabled" ]; then
            echo "  バージョン管理が有効です。すべてのバージョンを削除します..."
            # バージョン管理されたオブジェクトを一括削除
            versions=$(aws s3api list-object-versions --bucket $bucket --profile $AWS_PROFILE --output json --query '{Objects: Versions[].{Key:Key,VersionId:VersionId}}')
            if [ "$versions" != "{}" ]; then
                echo "$versions" | aws s3api delete-objects --bucket $bucket --delete file:///dev/stdin --profile $AWS_PROFILE
            else
                echo "  削除するオブジェクトがありません。"
            fi

            # 削除マーカーを一括削除
            markers=$(aws s3api list-object-versions --bucket $bucket --profile $AWS_PROFILE --output json --query '{Objects: DeleteMarkers[].{Key:Key,VersionId:VersionId}}')
            if [ "$markers" != "{}" ]; then
                echo "$markers" | aws s3api delete-objects --bucket $bucket --delete file:///dev/stdin --profile $AWS_PROFILE
            else
                echo "  削除する削除マーカーがありません。"
            fi
        else
            echo "  バージョン管理は無効です。通常のオブジェクトを削除します..."
            aws s3 rm s3://$bucket --recursive --profile $AWS_PROFILE
        fi

        # バケットの削除
        echo "  バケット削除中: $bucket"
        aws s3api delete-bucket --profile $AWS_PROFILE --bucket $bucket
    done
fi

# ECRリポジトリの削除
echo "ECRリポジトリの削除を開始..."
repos=$(aws ecr describe-repositories --profile $AWS_PROFILE --query "repositories[?contains(repositoryName, '$SEARCH_STRING')].repositoryName" --output text)

if [ -z "$repos" ]; then
    echo "  検索文字列 '${SEARCH_STRING}' にマッチするECRリポジトリは見つかりませんでした。"
else
    for repo in $repos; do
        echo "リポジトリ $repo を空にして削除中..."
        
        # すべてのイメージを削除
        aws ecr list-images --profile $AWS_PROFILE --repository-name $repo --query 'imageIds[*]' --output text | while read imageId; do
            if [ ! -z "$imageId" ]; then
                echo "  イメージ削除中: $imageId"
                aws ecr batch-delete-image --profile $AWS_PROFILE --repository-name $repo --image-ids imageDigest=$imageId
            fi
        done

        # リポジトリの削除
        echo "  リポジトリ削除中: $repo"
        aws ecr delete-repository --profile $AWS_PROFILE --repository-name $repo --force
    done
fi

echo "クリーンアップ完了！"