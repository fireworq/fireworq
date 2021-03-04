BIN=fireworq
SHELL=/bin/bash
BUILD_OUTPUT=.
TEST_OUTPUT=.
GO=go
GOINSTALL=GOOS= GOARCH= ${GO} install
PRERELEASE=SNAPSHOT
BUILD=$$(git describe --always)
TEST_ARGS=-parallel 1 -timeout 60s
export GO111MODULE=on

.PHONY: all
all: build

.PHONY: test
test: build test_deps
	{ ${GO} test ${TEST_ARGS} -race -v ./...; echo $$? > status.tmp; } | tee >(go-junit-report > ${TEST_OUTPUT}/junit_output.xml)
	exit $$(cat status.tmp)

.PHONY: cover
cover: build test_deps
	TEST_ARGS="${TEST_ARGS}" script/cover ${TEST_OUTPUT}/profile.cov
	${GO} tool cover -html=${TEST_OUTPUT}/profile.cov -o ${TEST_OUTPUT}/coverage.html
	gocover-cobertura < ${TEST_OUTPUT}/profile.cov > ${TEST_OUTPUT}/coverage.xml

.PHONY: build
build: generate
	${GO} build -race -ldflags "-X main.Build=$(BUILD) -X main.Prerelease=DEBUG" -o ${BUILD_OUTPUT}/$(BIN) .
	GOOS= GOARCH= ${GO} run script/gendoc/gendoc.go config > doc/config.md

.PHONY: release
release: clean credits generate
	CGO_ENABLED=0 ${GO} build -ldflags "-X main.Build=$(BUILD) -X main.Prerelease=$(PRERELEASE)" -o ${BUILD_OUTPUT}/$(BIN) .

.PHONY: credits
credits:
	GOOS= GOARCH= ${GO} run script/genauthors/genauthors.go > AUTHORS
	${GOINSTALL} github.com/Songmu/gocredits/cmd/gocredits@v0.2.0
	${GO} mod download
	gocredits -w .

.PHONY: generate
generate: generate_deps
	touch AUTHORS
	touch CREDITS
	GOOS= GOARCH= ${GO} generate -x ./...

.PHONY: generate_deps
generate_deps:
	${GOINSTALL} github.com/jessevdk/go-assets-builder@v0.0.0-20130903091706-b8483521738f
	${GOINSTALL} github.com/golang/mock/mockgen@v1.5.0

.PHONY: test_deps
test_deps:
	${GOINSTALL} github.com/jpillora/go-tcp-proxy/cmd/tcp-proxy@v1.0.2
	${GOINSTALL} github.com/jstemmer/go-junit-report@v0.9.1
	${GOINSTALL} golang.org/x/tools/cmd/cover@v0.0.0-20200831203904-5a2aa26beb65
	${GOINSTALL} github.com/wadey/gocovmerge@v0.0.0-20160331181800-b5bfa59ec0ad
	${GOINSTALL} github.com/t-yuki/gocover-cobertura@v0.0.0-20180217150009-aaee18c8195c

.PHONY: lint
lint: generate
	${GOINSTALL} golang.org/x/lint/golint@v0.0.0-20200302205851-738671d3881b
	${GO} vet ./...
	for d in $$(${GO} list ./...); do \
	  golint --set_exit_status "$$d" || exit $$? ; \
	done
	${GOINSTALL} github.com/kyoh86/scopelint@v0.2.0
	scopelint --set-exit-status --no-test ./...
	for f in $$(${GO} list -f '{{$$p := .}}{{range $$f := .GoFiles}}{{$$p.Dir}}/{{$$f}} {{end}} {{range $$f := .TestGoFiles}}{{$$p.Dir}}/{{$$f}} {{end}}' ./... | xargs); do \
	  [ $$(basename "$$f") = 'assets.go' ] && continue ; \
	  test -z "$$(gofmt -d -s "$$f" | tee /dev/stderr)" || exit $$? ; \
	done

.PHONY: clean
clean:
	find . -name assets.go -delete -or -name 'mock_*.go' -delete
	rm -f assets.go
	rm -f junit_output.xml profile.cov coverage.html coverage.xml
	rm -f $(BIN)
	${GO} clean
