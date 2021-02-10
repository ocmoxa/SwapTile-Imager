CMD:=cmd/imager/main.go
GOLANGCI_LINT_VER:=v1.36.0

run:
	go run $(CMD)
.PHONY: run

build:
	go build -o ./bin/imager $(CMD)
.PHONY: build

test:
	go test ./...
.PHONY: test

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
