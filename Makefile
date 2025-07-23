# PHONYターゲット一覧
.PHONY: docs
.PHONY: precommit-enable
.PHONY: precommit-disable
.PHONY: precommit-status

# ドキュメント生成
docs:
	@go run scripts/gen-docs.go

# Pre-commit管理
precommit-enable:
	@go run scripts/precommit.go enable

precommit-disable:
	@go run scripts/precommit.go disable

precommit-status:
	@go run scripts/precommit.go status