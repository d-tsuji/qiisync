.PHONY: all build test lint clean deps devel-deps

BIN := qiisync
BUILD_LDFLAGS := "-s -w"
GOBIN ?= $(shell go env GOPATH)/bin
export GO111MODULE=on

all: clean build

deps:
	go mod tidy

devel-deps: deps
	GO111MODULE=off go get -u \
	  golang.org/x/lint/golint

build:
	go build -ldflags=$(BUILD_LDFLAGS) -o $(BIN) ./cmd/qiisync

test: deps
	go test -v -count=1 ./...

test-cover: deps
	go test -v -count=1 ./... -cover -coverprofile=c.out
	go tool cover -html=c.out -o coverage.html

lint: devel-deps
	go vet ./...
	$(GOBIN)/golint -set_exit_status ./...

clean:
	rm -rf $(BIN)
	go clean

################################
#### For E2E Testing
################################
.PHONY: pull pull-only post update
pull: clean build
	rm -rf testdata/output/pull
	./$(BIN) pull

pull-only: clean build
	./$(BIN) pull

post: clean build
	 ./$(BIN) post ./testdata/qiita/post/test_article.md

update: clean build
	./$(BIN) update ./testdata/output/pull/20200424/hoge.md
