//go:build tools
// +build tools

package tools

import (
	_ "github.com/99designs/gqlgen"
	_ "golang.org/x/vuln/cmd/govulncheck"
	_ "honnef.co/go/tools/cmd/staticcheck"
)
