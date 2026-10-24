package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/codebymaribel/eva-sdd/internal/catalog"
	"github.com/codebymaribel/eva-sdd/internal/installer"
	"github.com/codebymaribel/eva-sdd/internal/manifest"
	"github.com/codebymaribel/eva-sdd/internal/prompt"
	"github.com/codebymaribel/eva-sdd/internal/textdiff"
	"github.com/codebymaribel/eva-sdd/templates"
)

// initFlags are the parsed flags of `eva init`.
type initFlags struct {
	dir, stacks, agents, paths, name string
	dryRun, force, yes, prune, diff  bool
}

func (a *App) runInit(args []string) int {
	f, code, ok := a.parseInitFlags(args)
	if !ok {
		return code
	}

	root, err := filepath.Abs(f.dir)
	if err == nil {
		err = requireDir(root)
	}
	if err != nil {
		fmt.Fprintf(a.Stderr, "eva init: %v\n", err)
		return ExitUsage
	}

	existing, err := manifest.Load(root)
	if err != nil && !errors.Is(err, manifest.ErrNotFound) {
		fmt.Fprintf(a.Stderr, "eva init: %v\n", err)
		return ExitFail
	}

	cfg, err := a.resolveConfig(root, f, existing)
	if err != nil {
		if errors.Is(err, prompt.ErrAborted) {
			fmt.Fprintln(a.Stderr, "\nCancelled.")
			return ExitFail
		}
		fmt.Fprintf(a.Stderr, "eva init: %v\n", err)
		return ExitUsage
	}

	actions, err := installer.Plan(root, templates.FS, cfg, installer.Options{Force: f.force, Prune: f.prune})
	if err != nil {
		fmt.Fprintf(a.Stderr, "eva init: %v\n", err)
		return ExitFail
	}
	a.printPlan(root, cfg, actions, f.diff)

	if f.dryRun {
		fmt.Fprintln(a.Stdout, "\nDry run: nothing was written.")
		return ExitOK
	}
	if a.Interactive && !f.yes {
		apply, err := a.Prompter.Confirm("\nApply these changes?", true)
		if err != nil || !apply {
			fmt.Fprintln(a.Stdout, "Cancelled, nothing was written.")
			return ExitFail
		}
	}
	if err := installer.Apply(root, actions); err != nil {
		fmt.Fprintf(a.Stderr, "eva init: %v\n", err)
		return ExitFail
	}
	a.printNextSteps(cfg, existing == nil)
	return ExitOK
}

// parseInitFlags returns the flags, or an exit code when parsing stops the run.
func (a *App) parseInitFlags(args []string) (initFlags, int, bool) {
	var f initFlags
	fs := flag.NewFlagSet("eva init", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	fs.StringVar(&f.dir, "dir", ".", "project directory")
	fs.StringVar(&f.stacks, "stack", "", "comma-separated stacks: "+strings.Join(catalog.StackIDs(), ", ")+" (aliases: rn, next)")
	fs.StringVar(&f.agents, "agent", "", "comma-separated agents: "+strings.Join(catalog.AgentIDs(), ", "))
	fs.StringVar(&f.paths, "path", "", `app root per stack, e.g. "rn=apps/mobile,next=apps/web" (default "." for one stack)`)
	fs.StringVar(&f.name, "name", "", "project name (default: directory name)")
	fs.BoolVar(&f.dryRun, "dry-run", false, "show what would change without writing")
	fs.BoolVar(&f.force, "force", false, "overwrite files that differ from the templates, even if you edited them (managed blocks are always merged)")
	fs.BoolVar(&f.prune, "prune", false, "delete files eva installed earlier, no longer ships and nobody edited")
	fs.BoolVar(&f.diff, "diff", false, "show how each edited or updated file differs from the new template")
	fs.BoolVar(&f.yes, "yes", false, "do not ask for confirmation")
	fs.Usage = func() {
		fmt.Fprint(a.Stderr, "Usage: eva init [flags]\n\nInstalls the SDD workflow. Re-run it to add stacks/agents or to update templates.\nWithout --stack/--agent in a terminal, an interactive wizard asks for them.\n\nFlags:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return f, ExitOK, false
		}
		return f, ExitUsage, false
	}
	if fs.NArg() > 0 {
		fmt.Fprintf(a.Stderr, "eva init: unexpected argument %q\n", fs.Arg(0))
		return f, ExitUsage, false
	}
	return f, 0, true
}

// resolveConfig merges flags, the existing manifest and interactive answers
// into the final configuration. Already-installed stacks and agents are kept.
func (a *App) resolveConfig(root string, f initFlags, existing *manifest.Manifest) (installer.Config, error) {
	if existing == nil {
		existing = &manifest.Manifest{}
	}

	paths, err := parsePaths(f.paths)
	if err != nil {
		return installer.Config{}, err
	}
	cfg := installer.Config{Project: f.name, EvaVersion: a.Version}
	if cfg.Project == "" {
		cfg.Project = existing.Project
	}
	if cfg.Project == "" {
		cfg.Project = filepath.Base(root)
	}

	// Wizard order: stacks → app roots → agents.
	stackIDs, err := a.resolveStacks(f.stacks, existing)
	if err != nil {
		return installer.Config{}, err
	}
	for _, id := range stackIDs {
		p, err := a.stackPath(id, len(stackIDs), paths, existing)
		if err != nil {
			return installer.Config{}, err
		}
		cfg.Stacks = append(cfg.Stacks, manifest.StackRef{ID: id, Path: p})
	}

	if cfg.Agents, err = a.resolveAgents(f.agents, existing); err != nil {
		return installer.Config{}, err
	}
	return cfg, nil
}

// resolveStacks returns installed ∪ requested stacks in catalog order.
func (a *App) resolveStacks(flagValue string, existing *manifest.Manifest) ([]string, error) {
	var installed []string
	for _, s := range existing.Stacks {
		installed = append(installed, s.ID)
	}

	var requested []string
	switch {
	case flagValue != "":
		stacks, err := catalog.ParseStacks(flagValue)
		if err != nil {
			return nil, err
		}
		for _, s := range stacks {
			requested = append(requested, s.ID)
		}
	case a.Interactive:
		var opts []prompt.Option
		for _, s := range catalog.Stacks {
			opts = append(opts, prompt.Option{Value: s.ID, Label: s.Name})
		}
		question := "Which stacks does this project use?"
		if len(installed) > 0 {
			question += " (installed stacks are kept)"
		}
		chosen, err := a.Prompter.MultiSelect(question, opts, installed)
		if err != nil {
			return nil, err
		}
		requested = chosen
	case len(installed) == 0:
		return nil, errors.New("--stack is required when not running in a terminal")
	}
	return unionInOrder(catalog.StackIDs(), installed, requested), nil
}

// resolveAgents returns installed ∪ requested agents in catalog order.
func (a *App) resolveAgents(flagValue string, existing *manifest.Manifest) ([]string, error) {
	installed := existing.Agents

	var requested []string
	switch {
	case flagValue != "":
		agents, err := catalog.ParseAgents(flagValue)
		if err != nil {
			return nil, err
		}
		for _, ag := range agents {
			requested = append(requested, ag.ID)
		}
	case a.Interactive:
		var opts []prompt.Option
		for _, ag := range catalog.Agents {
			opts = append(opts, prompt.Option{Value: ag.ID, Label: ag.Name})
		}
		chosen, err := a.Prompter.MultiSelect("Which agents should get /eva-* shortcuts? (AGENTS.md works for any agent)", opts, installed)
		if err != nil {
			return nil, err
		}
		requested = chosen
	case len(installed) == 0:
		return nil, errors.New("--agent is required when not running in a terminal")
	}
	return unionInOrder(catalog.AgentIDs(), installed, requested), nil
}

// stackPath picks the app root: --path flag > manifest > prompt > default.
func (a *App) stackPath(id string, total int, flagPaths map[string]string, existing *manifest.Manifest) (string, error) {
	if p, ok := flagPaths[id]; ok {
		return p, nil
	}
	if p, ok := existing.StackPath(id); ok {
		return p, nil
	}
	if total == 1 {
		return ".", nil
	}
	s, _ := catalog.FindStack(id)
	if a.Interactive {
		return a.Prompter.Input(fmt.Sprintf("App root for %s", s.Name), s.DefaultPath)
	}
	return s.DefaultPath, nil
}

// parsePaths parses "rn=apps/mobile,next=apps/web" into stack id → path.
func parsePaths(value string) (map[string]string, error) {
	out := map[string]string{}
	for _, pair := range strings.Split(value, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		key, p, ok := strings.Cut(pair, "=")
		if !ok || strings.TrimSpace(p) == "" {
			return nil, fmt.Errorf("invalid --path entry %q (want stack=path)", pair)
		}
		s, found := catalog.FindStack(key)
		if !found {
			return nil, fmt.Errorf("invalid --path entry %q: unknown stack %q", pair, key)
		}
		out[s.ID] = strings.TrimSpace(p)
	}
	return out, nil
}

// unionInOrder returns the values present in any list, ordered by `order`.
func unionInOrder(order []string, lists ...[]string) []string {
	set := map[string]bool{}
	for _, l := range lists {
		for _, v := range l {
			set[v] = true
		}
	}
	var out []string
	for _, v := range order {
		if set[v] {
			out = append(out, v)
		}
	}
	return out
}

func (a *App) printPlan(root string, cfg installer.Config, actions []installer.Action, showDiff bool) {
	w := a.Stdout
	fmt.Fprintf(w, "\neva init → %s (project %q)\n", root, cfg.Project)
	var stacks []string
	for _, s := range cfg.Stacks {
		stacks = append(stacks, fmt.Sprintf("%s (%s)", s.ID, s.Path))
	}
	fmt.Fprintf(w, "  stacks: %s\n  agents: %s\n\n", strings.Join(stacks, ", "), strings.Join(cfg.Agents, ", "))

	for _, act := range actions {
		if act.Op == installer.OpUnchanged {
			continue
		}
		line := fmt.Sprintf("  %-9s %s", act.Op, act.Path)
		if act.Note != "" {
			line += "  — " + act.Note
		}
		fmt.Fprintln(w, line)
		if showDiff && act.Diffable() {
			d := textdiff.Unified(act.Path+" (yours)", act.Path+" (template)", string(act.Current), string(act.Content))
			for _, l := range strings.SplitAfter(strings.TrimSuffix(d, "\n"), "\n") {
				fmt.Fprint(w, "      ", l)
			}
			fmt.Fprintln(w)
		}
	}
	sum := installer.Summary(actions)
	fmt.Fprintf(w, "\n  %d to create, %d to update, %d unchanged, %d skipped",
		sum[installer.OpCreate], sum[installer.OpUpdate], sum[installer.OpUnchanged], sum[installer.OpSkip])
	if n := sum[installer.OpDelete]; n > 0 {
		fmt.Fprintf(w, ", %d to delete", n)
	}
	fmt.Fprintln(w)
}

func (a *App) printNextSteps(cfg installer.Config, firstInstall bool) {
	w := a.Stdout
	fmt.Fprintln(w, "\nDone.")
	if !firstInstall {
		return
	}
	fmt.Fprintln(w, "\nNext steps:")
	fmt.Fprintln(w, "  1. In your agent, run /eva-constitution, then review and approve .eva/constitution.md")
	fmt.Fprintln(w, "  2. Review .eva/stacks/*.md and commit the generated files")
	fmt.Fprintln(w, "  3. Start a feature:  /eva-specify <what you want to build>")
	fmt.Fprintln(w, "  4. Add `eva check` to CI (see the eva README)")
}

func requireDir(p string) error {
	info, err := os.Stat(p)
	if err != nil {
		return fmt.Errorf("project directory %s: %w", p, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", p)
	}
	return nil
}
