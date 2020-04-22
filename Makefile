.PHONY: all build test lint clean

BIN := qsync
BUILD_LDFLAGS := "-s -w"
GOBIN ?= $(shell go env GOPATH)/bin
export GO111MODULE=on

all: clean build

build:
	go build -ldflags=$(BUILD_LDFLAGS) -o $(BIN)

test:
	go test -v -count=1 ./...

test-cover:
	go test -v -cover -coverprofile=c.out
	go tool cover -html=c.out -o coverage.html

lint:
	go get golang.org/x/lint/golint
	go vet ./...
	$(GOBIN)/golint -set_exit_status ./...

clean:
	rm -rf $(BIN)
	go clean

#### Local Test
pull: clean build
	rm -rf testdata/output/pull
	$(BIN) pull

post: build
	$(BIN) post --path ./testdata/output/post/test_article.md --title first_article_2 --tag Go:1.14 --private true

pull-only: clean build
	$(BIN) pull

upload: build
	$(BIN) upload ./testdata/output/post/test_article_posted.md
