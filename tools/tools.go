//go:build tools
// +build tools

// The tools package tracks packages used by go:generate, see https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
// For stringer, however, I need to run
//
//	go install golang.org/x/tools/cmd/stringer@latest
package tools

import (
	_ "github.com/maxbrunsfeld/counterfeiter/v6"
	_ "golang.org/x/tools/cmd/stringer"
)
