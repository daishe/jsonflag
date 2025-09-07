
GO_SRCS := $(shell find . -type f -name '*.go' -print)
GO_MODULE_FILES := go.mod go.sum
DEPS := $(GO_SRCS) $(GO_MODULE_FILES)

.PHONY: all
all: lint test

.PHONY: clean
clean: tools-clean
	rm -rf dependencies

dependencies: go.mod go.sum
	go mod download
	touch dependencies

.PHONY: test
test: dependencies
	go test -count 1 -timeout 5m ./...

.PHONY: lint
lint: bin/golangci-lint
	bin/golangci-lint config verify
	bin/golangci-lint run

.PHONY: tools
tools: bin/golangci-lint

.PHONY: tools-clean
tools-clean:
	rm -rf bin
	rm -rf tools/dependencies

tools/dependencies: tools/go.mod tools/go.sum tools/tools.go
	cd tools && go mod download
	touch tools/dependencies

bin/golangci-lint: tools/dependencies
	cd tools && go build -o ../bin/golangci-lint github.com/golangci/golangci-lint/v2/cmd/golangci-lint
