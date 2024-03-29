//go:build tools
// +build tools

package main

import (
	_ "github.com/client9/misspell/cmd/misspell"
	_ "github.com/fzipp/gocyclo/cmd/gocyclo"
	_ "github.com/rakyll/gotest"
	_ "golang.org/x/tools/cmd/goimports"
	_ "honnef.co/go/tools/cmd/staticcheck"

	_ "github.com/fornellas/rrb"
)
