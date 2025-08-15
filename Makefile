# PHONYターゲット一覧
.PHONY: setup
.PHONY: docs
.PHONY: sync-ai-docs
.PHONY: fmt
.PHONY: vet
.PHONY: lint
.PHONY: lint-fix
.PHONY: fix
.PHONY: check
.PHONY: precommit-enable
.PHONY: precommit-disable
.PHONY: precommit-status
.PHONY: sync-mcp

# 開発環境セットアップ
setup:
	@bash scripts/setup.sh

# ドキュメント生成
docs:
	go run scripts/gen-docs/main.go

# AI用ドキュメント同期
sync-ai-docs:
	go run scripts/sync-ai-docs/main.go

# MCP設定同期
sync-mcp:
	go run scripts/sync-mcp/main.go

# フォーマット
fmt:
	go fmt ./...

# 静的解析
vet:
	go vet ./...

# Lint（要golangci-lint）
lint:
	golangci-lint run

# Lint自動修正
lint-fix:
	golangci-lint run --fix

# まとめて修正（フォーマット + lint修正）
fix: fmt lint-fix
	@echo "✅ All fixes applied!"

# まとめてチェック
check: vet lint
	@echo "✅ All checks passed!"

# Pre-commit有効化
precommit-enable:
	go run scripts/precommit/main.go enable

# Pre-commit無効化
precommit-disable:
	go run scripts/precommit/main.go disable

# Pre-commit状態確認
precommit-status:
	go run scripts/precommit/main.go status