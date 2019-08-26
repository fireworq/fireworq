BIN=fireworq
SHELL=/bin/bash
BUILD_OUTPUT=.
TEST_OUTPUT=.
GO=go
PRERELEASE=SNAPSHOT
BUILD=$$(git describe --always)
TEST_ARGS=-parallel 1 -timeout 60s
export GO111MODULE=on

all: build

test: build test_deps
	{ ${GO} test ${TEST_ARGS} -race -v ./...; echo $$? > status.tmp; } | tee >(go-junit-report > ${TEST_OUTPUT}/junit_output.xml)
	exit $$(cat status.tmp)

cover: build test_deps
	TEST_ARGS="${TEST_ARGS}" script/cover ${TEST_OUTPUT}/profile.cov
	${GO} tool cover -html=${TEST_OUTPUT}/profile.cov -o ${TEST_OUTPUT}/coverage.html
	gocover-cobertura < ${TEST_OUTPUT}/profile.cov > ${TEST_OUTPUT}/coverage.xml

build: deps generate
	${GO} build -race -ldflags "-X main.Build=$(BUILD) -X main.Prerelease=DEBUG" -o ${BUILD_OUTPUT}/$(BIN) .
	GOOS= GOARCH= ${GO} run script/gendoc/gendoc.go config > doc/config.md

release: clean deps credits generate
	CGO_ENABLED=0 ${GO} build -ldflags "-X main.Build=$(BUILD) -X main.Prerelease=$(PRERELEASE)" -o ${BUILD_OUTPUT}/$(BIN) .

credits:
	GOOS= GOARCH= ${GO} run script/genauthors/genauthors.go > AUTHORS
	GO111MODULE=off GOOS= GOARCH= ${GO} get github.com/Songmu/gocredits/cmd/gocredits
	go mod tidy # not `go get` to get all the dependencies regardress of OS, architecture and build tags
	gocredits -w .

generate:
	touch AUTHORS
	touch CREDITS
	GOOS= GOARCH= ${GO} generate -x ./...

deps:
	GO111MODULE=off GOOS= GOARCH= ${GO} get github.com/jessevdk/go-assets-builder
	GO111MODULE=off GOOS= GOARCH= ${GO} get github.com/golang/mock/mockgen

test_deps:
	GO111MODULE=off ${GO} get github.com/jpillora/go-tcp-proxy/cmd/tcp-proxy
	GO111MODULE=off ${GO} get github.com/jstemmer/go-junit-report
	GO111MODULE=off ${GO} get golang.org/x/tools/cmd/cover
	GO111MODULE=off ${GO} get github.com/wadey/gocovmerge
	GO111MODULE=off ${GO} get github.com/t-yuki/gocover-cobertura

lint: generate
	GO111MODULE=off ${GO} get golang.org/x/lint/golint
	${GO} vet ./...
	for d in $$(${GO} list ./...); do \
	  golint --set_exit_status "$$d" || exit $$? ; \
	done
	for f in $$(${GO} list -f '{{$$p := .}}{{range $$f := .GoFiles}}{{$$p.Dir}}/{{$$f}} {{end}} {{range $$f := .TestGoFiles}}{{$$p.Dir}}/{{$$f}} {{end}}' ./... | xargs); do \
	  [ $$(basename "$$f") = 'assets.go' ] && continue ; \
	  test -z "$$(gofmt -d -s "$$f" | tee /dev/stderr)" || exit $$? ; \
	done

clean:
	find . -name assets.go -delete -or -name 'mock_*.go' -delete
	rm -f assets.go
	rm -f junit_output.xml profile.cov coverage.html coverage.xml
	rm -f $(BIN)
	${GO} clean

.PHONY: all test cover build release credits generate deps test_deps lint clean

