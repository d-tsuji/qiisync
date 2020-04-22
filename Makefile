BIN := qsync
BUILD_LDFLAGS := "-s -w"
GOBIN ?= $(shell go env GOPATH)/bin
export GO111MODULE=on

.PHONY: all
all: clean build

.PHONY: build
build:
	go build -ldflags=$(BUILD_LDFLAGS) -o $(BIN)

.PHONY: test
test:
	go test -v -count=1 ./...

.PHONY: lint
lint:
	go get golang.org/x/lint/golint
	go vet ./...
	$(GOBIN)/golint -set_exit_status ./...

.PHONY: clean
clean:
	rm -rf $(BIN)
	go clean

#### Local Test
.PHONY: pull
pull: clean build
	rm -rf testdata/output/pull
	$(BIN) pull

.PHONY: post
post: build
	$(BIN) post --path ./testdata/output/post/test_article.md --title first_article_2 --tag Go:1.14 --private true

.PHONY: pull-only
pull-only: clean build
	$(BIN) pull

.PHONY: upload
upload: build
	$(BIN) upload ./testdata/output/post/test_article_posted.md
