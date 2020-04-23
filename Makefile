.PHONY: all build test lint clean

BIN := qiisync
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

################################
#### For E2E Testing
################################
pull: clean build
	rm -rf testdata/qiita/pull
	./$(BIN) pull

pull-only: clean build
	./$(BIN) pull

post: clean build
	 ./$(BIN) post ./testdata/qiita/post/test_article.md

update: clean build
	./$(BIN) update ./testdata/output/pull/20200423/a.md
