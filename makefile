CMD:=cmd/imager/main.go
GOLANGCI_LINT_VER:=v1.36.0

run:
	go run $(CMD)
.PHONY: run

build:
	go build -o ./bin/imager $(CMD)
.PHONY: build

test: test.unit test.integration
.PHONY: test

test.unit:
	go test ./...
.PHONY: test.unit

test.integration:
	go test -tags=integration ./...
ifeq ($(TEST_IMAGE_REDIS),)
	@echo "\033[0;33mEnvironment variable TEST_IMAGE_REDIS not set. Some tests were skipped.\033[0m"
	@echo "Try: export TEST_IMAGE_REDIS=redis://localhost:6379"
endif
.PHONY: test.integration

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
