package cli

import (
	"errors"
	"flag"
	"fmt"
	"path/filepath"

	"github.com/codebymaribel/eva-sdd/internal/check"
)

func (a *App) runCheck(args []string) int {
	var (
		dir  string
		opts check.Options
	)
	fs := flag.NewFlagSet("eva check", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	fs.StringVar(&dir, "dir", ".", "project directory")
	fs.StringVar(&opts.Base, "base", "", "base ref to diff against (default: $GITHUB_BASE_REF, origin/main, main, origin/master, master)")
	fs.StringVar(&opts.Branch, "branch", "", "branch name (default: $GITHUB_HEAD_REF or the current branch)")
	fs.BoolVar(&opts.Strict, "strict", false, "fail when a branch without a feature id changes code")
	fs.BoolVar(&opts.NoGit, "no-git", false, "only validate documents, skip the branch gate")
	fs.Usage = func() {
		fmt.Fprint(a.Stderr, "Usage: eva check [flags]\n\nValidates specs/ and the approval gate: code on feat/NNN-* needs spec, plan and tasks approved.\nExits 1 when errors are found.\n\nFlags:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		return ExitUsage
	}
	if fs.NArg() > 0 {
		fmt.Fprintf(a.Stderr, "eva check: unexpected argument %q\n", fs.Arg(0))
		return ExitUsage
	}

	root, err := filepath.Abs(dir)
	if err == nil {
		err = requireDir(root)
	}
	if err != nil {
		fmt.Fprintf(a.Stderr, "eva check: %v\n", err)
		return ExitUsage
	}
	opts.Getenv = a.Getenv

	report, err := check.Run(root, opts)
	if err != nil {
		fmt.Fprintf(a.Stderr, "eva check: %v\n", err)
		return ExitFail
	}
	a.printReport(report)
	if report.Count(check.SevError) > 0 {
		return ExitFail
	}
	return ExitOK
}

func (a *App) printReport(r check.Report) {
	w := a.Stdout
	fmt.Fprintf(w, "eva check — %d feature(s)", len(r.Features))
	if r.GateChecked {
		fmt.Fprintf(w, ", branch %s vs %s (%d code change(s))", r.Branch, r.Base, r.CodeChanges)
	}
	fmt.Fprintln(w)

	for _, f := range r.Findings {
		label := "ERROR"
		if f.Severity == check.SevWarning {
			label = "WARN "
		}
		fmt.Fprintf(w, "  %s %s %s: %s\n", label, f.Code, f.Path, f.Message)
	}

	errs, warns := r.Count(check.SevError), r.Count(check.SevWarning)
	if errs == 0 && warns == 0 {
		fmt.Fprintln(w, "OK")
		return
	}
	fmt.Fprintf(w, "%d error(s), %d warning(s)\n", errs, warns)
}
