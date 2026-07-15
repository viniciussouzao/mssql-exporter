// Package version holds build metadata injected via -ldflags at build time
// (see Makefile and Dockerfile). Defaults apply to `go run`/`go test` and any
// build that doesn't pass the ldflags.
package version

import (
	"fmt"
	"runtime"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// String formats build metadata for the --version flag and log lines.
func String() string {
	return fmt.Sprintf("mssql-exporter %s (commit: %s, built: %s, %s)", Version, Commit, Date, runtime.Version())
}
