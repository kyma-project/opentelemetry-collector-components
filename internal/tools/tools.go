//go:build tools

package tools

// This file follows the recommendation at
// https://go.dev/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
// on how to pin tooling dependencies to a go.mod file.
// This ensures that all systems use the same version of tools in addition to regular dependencies.
import (
	_ "github.com/golangci/golangci-lint/v2/cmd/golangci-lint"
	_ "github.com/vladopajic/go-test-coverage/v2"
	_ "go.opentelemetry.io/build-tools/crosslink"
	_ "go.opentelemetry.io/collector/cmd/builder"
	_ "go.opentelemetry.io/collector/cmd/mdatagen"
	_ "golang.org/x/tools/cmd/goimports"
)
