CMD:=cmd/imager/main.go
GOLANGCI_LINT_VER:=v1.36.0

export GOBIN:=$(PWD)/bin

run:
	go run $(CMD)
.PHONY: run

build:
	go build -o ./bin/imager $(CMD)
.PHONY: build

test: test.unit test.integration
.PHONY: test

test.coverage:
	go test -tags=integration,!integration -covermode=count -coverprofile=coverage.out ./...
.PHONY: test.coverage

test.unit:
	go test ./...
.PHONY: test.unit

test.integration:
	go test -tags=integration ./...
.PHONY: test.integration

proto:
	protoc --go_out=internal/pkg/imager/improto proto/improto.proto
.PHONY: proto

vendor:
	go mod tidy
	go mod vendor
.PHONY: vendor

lint:
	./bin/golangci-lint run
.PHONY: lint

prepare:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
		| sh -s -- -b ./bin $(GOLANGCI_LINT_VER)
.PHONY: prepare
