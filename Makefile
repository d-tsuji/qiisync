.PHONY: all build test lint clean deps devel-deps installer

BIN := qiisync
BUILD_LDFLAGS := "-s -w"
GOBIN ?= $(shell go env GOPATH)/bin
export GO111MODULE=on

repo_name := d-tsuji/qiisync
current_dir := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

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

installer:
	sh -c '\
      tmpdir=$$(mktemp -d); \
      cd $$tmpdir; \
      # Build from source, because "go get github.com/goreleaser/godownloader" will result in an error; \
      git clone --depth 1 https://github.com/goreleaser/godownloader && cd godownloader; \
      go run . -f -r ${repo_name} -o ${current_dir}/install.sh; \
      rm -rf $$tmpdir'

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
