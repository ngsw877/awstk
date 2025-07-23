# PHONYターゲット一覧
.PHONY: docs
.PHONY: sync-ai-docs
.PHONY: precommit-enable
.PHONY: precommit-disable
.PHONY: precommit-status

# ドキュメント生成
docs:
	@go run scripts/gen-docs/main.go

# AI用ドキュメント同期
sync-ai-docs:
	@go run scripts/sync-ai-docs/main.go

# Pre-commit管理
precommit-enable:
	@go run scripts/precommit/main.go enable

precommit-disable:
	@go run scripts/precommit/main.go disable

precommit-status:
	@go run scripts/precommit/main.go status