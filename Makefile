# lomax developer tasks. Linux-oriented; assumes go, golangci-lint, and a
# POSIX shell on PATH.
.PHONY: build test vet fmt lint hooks check

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
