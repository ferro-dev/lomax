# lomax developer tasks. Linux-oriented; assumes go, golangci-lint, and a
# POSIX shell on PATH.
.PHONY: build test vet fmt lint hooks check build-plugins test-plugins check-plugins check-all

build:
	go build ./...

test:
	go test -race ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

lint:
	golangci-lint run

# Enable the repository's git hooks (runs the CI gates on commit).
hooks:
	git config core.hooksPath .githooks
	@echo "git hooks enabled (core.hooksPath=.githooks)"

# Run everything CI runs, minus the distro matrix.
check: fmt vet lint test

# plugins/<name> are each their own Go module (multi-module monorepo — see
# docs/music-cli-plan.md section 16); root's targets above never touch them.
build-plugins:
	@for d in plugins/*/; do (cd "$$d" && go build ./...) || exit 1; done

test-plugins:
	@for d in plugins/*/; do (cd "$$d" && go test ./...) || exit 1; done

check-plugins:
	@for d in plugins/*/; do (cd "$$d" && go build ./... && go vet ./... && gofmt -l . && go test ./...) || exit 1; done

# Root module + every plugin module.
check-all: check check-plugins
