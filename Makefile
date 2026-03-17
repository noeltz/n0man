.PHONY: test lint vet build all

test:
	go test ./... -v -race -cover

lint:
	golangci-lint run ./...

vet:
	go vet ./...

build:
	go build -o n0man ./cmd/n0man

all: test lint vet build
