#!/bin/bash

# awstk 開発環境セットアップスクリプト

set -e  # エラーで停止


# ユーザー確認関数（Homebrew前提）
ask_install_by_homebrew() {
    echo "   Homebrewでインストールしますか？ (y/N)"
    echo "   (Nを選択した場合は手動でインストールしてください)"
    read -r response
    [[ "$response" =~ ^[Yy]$ ]]
}

echo "🚀 awstk 開発環境をセットアップします..."
echo ""

# Go バージョン確認
echo "📌 Go バージョンを確認中..."
if ! command -v go &> /dev/null; then
    echo "❌ Go がインストールされていません（必須）"
    if ask_install_by_homebrew; then
        echo "📦 Go をインストール中..."
        brew install go
        echo "✅ Go インストール完了"
    else
        echo "   手動でインストールしてから再実行してください"
        echo "   https://go.dev/dl/"
        exit 1
    fi
    echo ""
fi

GO_VERSION=$(go version | awk '{print $3}')
echo "✅ Go バージョン: $GO_VERSION"
echo ""

# AWS CLI の確認（必須）
echo "☁️  AWS CLI を確認中..."
if ! command -v aws &> /dev/null; then
    echo "❌ AWS CLI がインストールされていません（必須）"
    echo "   公式インストーラでインストールしますか？ (y/N)"
    echo "   (Nを選択した場合は手動でインストールしてください)"
    read -r response
    if [[ "$response" =~ ^[Yy]$ ]]; then
        echo "📦 AWS CLI をインストール中..."
        # 最新版のダウンロードURL
        curl "https://awscli.amazonaws.com/AWSCLIV2.pkg" -o "AWSCLIV2.pkg"
        sudo installer -pkg AWSCLIV2.pkg -target /
        rm AWSCLIV2.pkg
        echo "✅ AWS CLI インストール完了"
    else
        echo "   手動でインストールしてから再実行してください"
        echo "   https://docs.aws.amazon.com/ja_jp/cli/latest/userguide/getting-started-install.html"
        exit 1
    fi
    echo ""
fi

AWS_VERSION=$(aws --version | awk '{print $1}')
echo "✅ AWS CLI: $AWS_VERSION"
echo ""

# 依存関係のダウンロード
echo "📦 Go モジュールをダウンロード中..."
go mod download
echo "✅ モジュールのダウンロード完了"
echo ""

# golangci-lint のインストール確認
echo "🔍 golangci-lint を確認中..."
if ! command -v golangci-lint &> /dev/null; then
    echo "⚠️  golangci-lint がインストールされていません"
    if ask_install_by_homebrew; then
        echo "📦 golangci-lint をインストール中..."
        brew install golangci-lint
        echo "✅ golangci-lint インストール完了"
    else
        echo "   スキップしました。後でインストールできます:"
        echo "   brew install golangci-lint または go install でインストール可能"
    fi
    echo ""
fi

# golangci-lintがインストールされている場合のみバージョン表示
if command -v golangci-lint &> /dev/null; then
    LINT_VERSION=$(golangci-lint --version | head -n 1)
    echo "✅ golangci-lint: $LINT_VERSION"
    echo ""
fi


# pre-commit フックの設定確認
echo "🪝 pre-commit フックを確認中..."
HOOKS_PATH=$(git config --local --get core.hooksPath 2>/dev/null || echo "")
if [ "$HOOKS_PATH" = ".githooks" ]; then
    echo "✅ pre-commit フックは有効です"
else
    echo "🔧 pre-commit フックを有効化中..."
    if make precommit-enable; then
        echo "✅ pre-commit フック有効化完了"
    else
        echo "❌ pre-commit フックの有効化に失敗しました"
    fi
fi
echo ""

# ビルド確認
echo "🔨 ビルドを実行中..."
if go build -o /tmp/awstk-test . 2>/dev/null; then
    echo "✅ ビルド成功"
    rm -f /tmp/awstk-test
else
    echo "❌ ビルドエラー"
    echo "   'go build .' を実行して詳細を確認してください"
fi
echo ""

# 開発コマンドの案内
echo "📝 利用可能な開発コマンド:"
if [ -f "Makefile" ]; then
    # Makefileからターゲットとその直前のコメントを抽出
    awk '
    /^# / && !/^# PHONY/ && !/^# PHONYターゲット/ {
        comment = substr($0, 3)  # "# "を削除
        getline  # 次の行を読む
        if ($0 ~ /^[a-zA-Z_-]+:/) {
            gsub(/:.*/, "", $1)  # ターゲット名のみ抽出
            printf "   make %-12s - %s\n", $1, comment
        }
    }
    /^[a-zA-Z_-]+:/ && comment == "" {
        gsub(/:.*/, "", $1)
        printf "   make %s\n", $1
    }
    /^[a-zA-Z_-]+:/ { comment = "" }
    ' Makefile
else
    echo "   Makefile が見つかりません"
fi
echo ""

echo "✨ セットアップ完了！"