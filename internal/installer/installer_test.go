package installer

import (
	"bytes"
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/codebymaribel/eva-sdd/internal/manifest"
	"github.com/codebymaribel/eva-sdd/templates"
)

var update = flag.Bool("update", false, "rewrite golden files in testdata/golden")

// goldenCases cover single stack, single agent and the full monorepo combo.
var goldenCases = map[string]Config{
	"rn-claude": {
		Project: "demo", EvaVersion: "test",
		Stacks: []manifest.StackRef{{ID: "react-native", Path: "."}},
		Agents: []string{"claude"},
	},
	"next-opencode": {
		Project: "demo", EvaVersion: "test",
		Stacks: []manifest.StackRef{{ID: "nextjs", Path: "."}},
		Agents: []string{"opencode"},
	},
	"monorepo-all": {
		Project: "demo", EvaVersion: "test",
		Stacks: []manifest.StackRef{{ID: "react-native", Path: "apps/mobile"}, {ID: "nextjs", Path: "apps/web"}},
		Agents: []string{"claude", "opencode"},
	},
}

func TestGolden(t *testing.T) {
	for name, cfg := range goldenCases {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			actions, err := Plan(root, templates.FS, cfg, Options{})
			if err != nil {
				t.Fatal(err)
			}
			for _, a := range actions {
				if a.Op != OpCreate {
					t.Errorf("%s: op = %s on empty dir, want create", a.Path, a.Op)
				}
			}
			if err := Apply(root, actions); err != nil {
				t.Fatal(err)
			}
			golden := filepath.Join("testdata", "golden", name)
			if *update {
				if err := os.RemoveAll(golden); err != nil {
					t.Fatal(err)
				}
				copyTree(t, root, golden)
				return
			}
			compareTrees(t, golden, root)
		})
	}
}

func TestPlanIsIdempotent(t *testing.T) {
	root := t.TempDir()
	cfg := goldenCases["monorepo-all"]
	mustApply(t, root, cfg, Options{})

	again, err := Plan(root, templates.FS, cfg, Options{})
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range again {
		if a.Op != OpUnchanged {
			t.Errorf("%s: op = %s on second run, want unchanged", a.Path, a.Op)
		}
	}
}

func TestPlanKeepsLocallyEditedFilesUnlessForced(t *testing.T) {
	root := t.TempDir()
	cfg := goldenCases["rn-claude"]
	mustApply(t, root, cfg, Options{})

	constitution := filepath.Join(root, ".eva", "constitution.md")
	writeFile(t, constitution, "my constitution\n")

	actions, err := Plan(root, templates.FS, cfg, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if a := find(t, actions, ".eva/constitution.md"); a.Op != OpSkip {
		t.Errorf("op = %s, want skip", a.Op)
	}
	mustApplyActions(t, root, actions)
	if got := readFile(t, constitution); got != "my constitution\n" {
		t.Errorf("constitution overwritten without --force: %q", got)
	}

	forced, err := Plan(root, templates.FS, cfg, Options{Force: true})
	if err != nil {
		t.Fatal(err)
	}
	if a := find(t, forced, ".eva/constitution.md"); a.Op != OpUpdate {
		t.Errorf("op = %s, want update with --force", a.Op)
	}
}

func TestPlanMergesIntoExistingAgentsFile(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "AGENTS.md"), "# Team rules\n\nUse pnpm.\n")

	actions, err := Plan(root, templates.FS, goldenCases["rn-claude"], Options{})
	if err != nil {
		t.Fatal(err)
	}
	a := find(t, actions, "AGENTS.md")
	if a.Op != OpUpdate {
		t.Errorf("op = %s, want update", a.Op)
	}
	got := string(a.Content)
	if !strings.HasPrefix(got, "# Team rules\n\nUse pnpm.\n\n"+StartMarker) {
		t.Errorf("user content not preserved at the top:\n%s", got)
	}
	if !strings.Contains(got, "React Native (Expo) → `.`") {
		t.Errorf("stack line missing:\n%s", got)
	}
}

func TestPlanAddingStackUpdatesBlockAndManifest(t *testing.T) {
	root := t.TempDir()
	mustApply(t, root, goldenCases["rn-claude"], Options{})

	cfg := Config{
		Project: "demo", EvaVersion: "test",
		Stacks: []manifest.StackRef{{ID: "react-native", Path: "."}, {ID: "nextjs", Path: "web"}},
		Agents: []string{"claude"},
	}
	actions, err := Plan(root, templates.FS, cfg, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if a := find(t, actions, ".eva/stacks/nextjs.md"); a.Op != OpCreate {
		t.Errorf("nextjs stack op = %s, want create", a.Op)
	}
	if a := find(t, actions, "AGENTS.md"); a.Op != OpUpdate {
		t.Errorf("AGENTS.md op = %s, want update", a.Op)
	}
	if a := find(t, actions, manifest.Path); a.Op != OpUpdate || !bytes.Contains(a.Content, []byte(`"path": "web"`)) {
		t.Errorf("manifest not updated: %s %s", a.Op, a.Content)
	}
}

func TestPlanRejectsInvalidConfig(t *testing.T) {
	cases := map[string]Config{
		"no stacks":     {Agents: []string{"claude"}},
		"no agents":     {Stacks: []manifest.StackRef{{ID: "nextjs"}}},
		"unknown stack": {Stacks: []manifest.StackRef{{ID: "flutter"}}, Agents: []string{"claude"}},
		"unknown agent": {Stacks: []manifest.StackRef{{ID: "nextjs"}}, Agents: []string{"cursor"}},
		"escaping path": {Stacks: []manifest.StackRef{{ID: "nextjs", Path: "../other"}}, Agents: []string{"claude"}},
	}
	for name, cfg := range cases {
		if _, err := Plan(t.TempDir(), templates.FS, cfg, Options{}); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}

func TestPlanDetectsTemplateCollisions(t *testing.T) {
	tfs := fstest.MapFS{
		"core/README.md":             {Data: []byte("a")},
		"stacks/nextjs/README.md":    {Data: []byte("b")},
		"agents/claude/.claude/x.md": {Data: []byte("c")},
	}
	cfg := Config{Stacks: []manifest.StackRef{{ID: "nextjs", Path: "."}}, Agents: []string{"claude"}}
	if _, err := Plan(t.TempDir(), tfs, cfg, Options{}); err == nil || !strings.Contains(err.Error(), "both write README.md") {
		t.Errorf("expected collision error, got %v", err)
	}
}

func TestToSourceSuffixes(t *testing.T) {
	cases := []struct {
		in             string
		dest           string
		render, manage bool
	}{
		{"core/AGENTS.md.block.tmpl", "AGENTS.md", true, true},
		{"agents/claude/CLAUDE.md.block", "CLAUDE.md", false, true},
		{"core/.eva/constitution.md.tmpl", ".eva/constitution.md", true, false},
		{"core/specs/README.md", "specs/README.md", false, false},
	}
	for _, c := range cases {
		root := strings.Join(strings.SplitN(c.in, "/", 3)[:2], "/")
		if strings.HasPrefix(c.in, "core/") {
			root = "core"
		}
		s := toSource(root, c.in)
		if s.dest != c.dest || s.render != c.render || s.managed != c.manage {
			t.Errorf("toSource(%q) = %+v", c.in, s)
		}
	}
}

// --- helpers ---

func mustApply(t *testing.T, root string, cfg Config, opts Options) {
	t.Helper()
	actions, err := Plan(root, templates.FS, cfg, opts)
	if err != nil {
		t.Fatal(err)
	}
	mustApplyActions(t, root, actions)
}

func mustApplyActions(t *testing.T, root string, actions []Action) {
	t.Helper()
	if err := Apply(root, actions); err != nil {
		t.Fatal(err)
	}
}

func find(t *testing.T, actions []Action, p string) Action {
	t.Helper()
	for _, a := range actions {
		if a.Path == p {
			return a
		}
	}
	t.Fatalf("no action for %s", p)
	return Action{}
}

func writeFile(t *testing.T, p, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func listTree(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		b, err := os.ReadFile(p)
		out[filepath.ToSlash(rel)] = string(b)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func copyTree(t *testing.T, from, to string) {
	t.Helper()
	for rel, content := range listTree(t, from) {
		writeFile(t, filepath.Join(to, filepath.FromSlash(rel)), content)
	}
}

func compareTrees(t *testing.T, want, got string) {
	t.Helper()
	w, g := listTree(t, want), listTree(t, got)
	for p, wc := range w {
		gc, ok := g[p]
		switch {
		case !ok:
			t.Errorf("missing file %s", p)
		case gc != wc:
			t.Errorf("%s differs from golden (run `make golden` if the change is intended)", p)
		}
	}
	for p := range g {
		if _, ok := w[p]; !ok {
			t.Errorf("unexpected file %s (run `make golden` if intended)", p)
		}
	}
}
