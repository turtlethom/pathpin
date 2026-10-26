// Command pathpin is a bookmark-first interactive directory navigator.
package main

import "github.com/turtlethom/pathpin/v2/cmd"

// version is set when GoReleaser builds a release: -ldflags "-X main.version=2.0.0".
// It lives in main so that flag doesn't name the module path, which changes
// with the major version. Empty in other builds; cmd falls back to Go's build info.
// It's pathpin's one package-level variable, because -X can only set those;
// nothing assigns it while pathpin runs.
var version string

// main only hands off to cmd: package main can't be imported by tests, so
// everything worth testing lives in cmd and internal/.
func main() {
	cmd.Execute(version)
}
