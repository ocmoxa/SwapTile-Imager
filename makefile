CMD:=cmd/imager/main.go
GOLANGCI_LINT_VER:=v1.36.0

TEST_PKGS := ./internal/... ./docs/...
TEST_ARGS := -race

export GOBIN:=$(PWD)/bin

run:
	go run $(CMD)
.PHONY: run

build:
	go build -o ./bin/imager $(CMD)
.PHONY: build

test: test.unit test.integration test.becnhmark
.PHONY: test

test.coverage:
	go test \
		-tags=integration,!integration \
		-covermode=atomic \
		-coverprofile=coverage.out \
		$(TEST_ARGS) \
		${TEST_PKGS}
.PHONY: test.coverage

test.unit:
	go test $(TEST_ARGS) ${TEST_PKGS}
.PHONY: test.unit

test.integration:
	go test -tags=integration $(TEST_ARGS) ${TEST_PKGS}
.PHONY: test.integration

test.becnhmark:
	go test -test.run="^$$" -test.bench=. -tags=integration,!integration ${TEST_PKGS}
.PHONY: test.becnhmark

proto:
	protoc --go_out=internal/pkg/imager/improto proto/improto.proto
.PHONY: proto

vendor:
	go mod tidy
	go mod vendor
.PHONY: vendor

lint:
	./bin/golangci-lint run ./internal/... ./docs/... ./cmd/...
.PHONY: lint

prepare:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
		| sh -s -- -b ./bin $(GOLANGCI_LINT_VER)
.PHONY: prepare
