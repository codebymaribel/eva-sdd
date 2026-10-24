// Package cli wires eva's subcommands to the installer and checker. It only
// uses the standard library so `go install` needs no third-party modules.
package cli

import (
	"fmt"
	"io"

	"github.com/codebymaribel/eva-sdd/internal/prompt"
)

// Exit codes.
const (
	ExitOK    = 0
	ExitFail  = 1 // command ran and found problems (e.g. check errors, cancelled)
	ExitUsage = 2 // invalid flags or arguments
)

// App holds the I/O of one CLI invocation, so tests can drive it.
type App struct {
	Stdin          io.Reader
	Stdout, Stderr io.Writer
	Version        string
	// Interactive enables prompts; main sets it when stdin is a terminal.
	Interactive bool
	// Prompter asks questions in interactive mode (defaults to a line prompter).
	Prompter prompt.Prompter
	// Getenv reads environment variables (os.Getenv in production).
	Getenv func(string) string
}

// New creates an App with a line-based prompter on stdin/stdout.
func New(stdin io.Reader, stdout, stderr io.Writer, version string) *App {
	return &App{
		Stdin:    stdin,
		Stdout:   stdout,
		Stderr:   stderr,
		Version:  version,
		Prompter: prompt.NewLine(stdin, stdout),
		Getenv:   func(string) string { return "" },
	}
}

// Run executes the command line (without the program name) and returns the
// process exit code.
func (a *App) Run(args []string) int {
	if len(args) == 0 {
		a.usage(a.Stdout)
		return ExitOK
	}
	switch args[0] {
	case "init":
		return a.runInit(args[1:])
	case "check":
		return a.runCheck(args[1:])
	case "version", "--version", "-v":
		fmt.Fprintf(a.Stdout, "eva %s\n", a.Version)
		return ExitOK
	case "help", "--help", "-h":
		a.usage(a.Stdout)
		return ExitOK
	default:
		fmt.Fprintf(a.Stderr, "eva: unknown command %q\n\n", args[0])
		a.usage(a.Stderr)
		return ExitUsage
	}
}

func (a *App) usage(w io.Writer) {
	fmt.Fprint(w, `eva — Spec-Driven Development workflow for web & mobile projects

Usage:
  eva init  [flags]   install or update the SDD workflow in a project
  eva check [flags]   validate specs and the approval gate (local or CI)
  eva version         print the version

Run "eva <command> -h" for the flags of a command.

Workflow (inside your agent):
  /eva-constitution → /eva-specify → /eva-clarify → /eva-plan → /eva-tasks → /eva-implement → /eva-verify
  A developer approves spec, plan and tasks before code is written.
`)
}
