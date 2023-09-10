.DEFAULT_GOAL := default

.PHONY: default
default: test

.PHONY: gen
gen:
	go generate -v ./...

.PHONY: tools
tools:
	go install -v go.uber.org/mock/mockgen@latest
	go install -v github.com/golangci/golangci-lint/cmd/golangci-lint@v1.53.3

.PHONY: lint
lint: #gen
	golangci-lint run --timeout=10m --fix ./...

.PHONY: test
test: tools lint
	go test  ./... #-functional=true -func=true
