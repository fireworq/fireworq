// +build tools

package main

import (
	_ "github.com/Songmu/gocredits"
	_ "github.com/golang/mock/mockgen"
	_ "github.com/jessevdk/go-assets-builder"
	_ "github.com/jpillora/go-tcp-proxy"
	_ "github.com/jstemmer/go-junit-report"
	_ "github.com/kyoh86/scopelint"
	_ "github.com/t-yuki/gocover-cobertura"
	_ "github.com/wadey/gocovmerge"
	_ "golang.org/x/lint/golint"
	_ "golang.org/x/tools/cmd/cover"
)
