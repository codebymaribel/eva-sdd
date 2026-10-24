// Command eva installs and enforces a Spec-Driven Development (SDD) workflow
// in web and mobile projects.
package main

import (
	"os"
	"runtime/debug"

	"github.com/codebymaribel/eva-sdd/internal/cli"
)

// version is set at build time with -ldflags "-X main.version=v0.1.0".
// When empty, the module version recorded by `go install` is used.
var version = ""

func main() {
	app := cli.New(os.Stdin, os.Stdout, os.Stderr, resolveVersion())
	app.Interactive = isTerminal(os.Stdin) && os.Getenv("CI") == ""
	app.Getenv = os.Getenv
	os.Exit(app.Run(os.Args[1:]))
}

// resolveVersion returns the ldflags version, the `go install` module
// version, or "dev" for local builds.
func resolveVersion() string {
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}
	return "dev"
}

// isTerminal reports whether f looks like an interactive terminal without
// external dependencies: a character device that is not the null device
// (CI runners and `< /dev/null` also report a character device).
func isTerminal(f *os.File) bool {
	stat, err := f.Stat()
	if err != nil || stat.Mode()&os.ModeCharDevice == 0 {
		return false
	}
	if null, err := os.Stat(os.DevNull); err == nil && os.SameFile(stat, null) {
		return false
	}
	return true
}
