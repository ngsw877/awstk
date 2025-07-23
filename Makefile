# Pre-commit管理
.PHONY: precommit-enable
precommit-enable:
	@go run scripts/precommit.go enable

.PHONY: precommit-disable
precommit-disable:
	@go run scripts/precommit.go disable

.PHONY: precommit-status
precommit-status:
	@go run scripts/precommit.go status