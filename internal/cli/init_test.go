package cli

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"encoding/json"
	"github.com/codebymaribel/eva-sdd/internal/installer"
	"github.com/codebymaribel/eva-sdd/internal/manifest"
)

// runApp runs eva with the given stdin and returns exit code, stdout, stderr.
func runApp(t *testing.T, interactive bool, stdin string, args ...string) (int, string, string) {
	t.Helper()
	var out, errOut strings.Builder
	app := New(strings.NewReader(stdin), &out, &errOut, "test")
	app.Interactive = interactive
	code := app.Run(args)
	return code, out.String(), errOut.String()
}

func loadManifest(t *testing.T, dir string) *manifest.Manifest {
	t.Helper()
	m, err := manifest.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestInitWithFlagsInstallsFiles(t *testing.T) {
	dir := t.TempDir()
	code, out, errOut := runApp(t, false, "", "init", "--dir", dir, "--stack", "rn", "--agent", "claude,opencode", "--name", "shop")
	if code != ExitOK {
		t.Fatalf("exit %d, stderr: %s", code, errOut)
	}
	for _, f := range []string{"AGENTS.md", "CLAUDE.md", ".claude/commands/eva-specify.md", ".opencode/commands/eva-plan.md", ".eva/stacks/react-native.md", "specs/README.md"} {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Errorf("missing %s", f)
		}
	}
	m := loadManifest(t, dir)
	if m.Project != "shop" || !reflect.DeepEqual(m.Stacks, []manifest.StackRef{{ID: "react-native", Path: "."}}) {
		t.Errorf("unexpected manifest: %+v", m)
	}
	if !strings.Contains(out, "Next steps:") {
		t.Errorf("missing next steps:\n%s", out)
	}
}

func TestInitDryRunWritesNothing(t *testing.T) {
	dir := t.TempDir()
	code, out, _ := runApp(t, false, "", "init", "--dir", dir, "--stack", "next", "--agent", "opencode", "--dry-run")
	if code != ExitOK {
		t.Fatalf("exit %d", code)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 0 {
		t.Errorf("dry run wrote %d entries", len(entries))
	}
	if !strings.Contains(out, "create    AGENTS.md") || !strings.Contains(out, "Dry run") {
		t.Errorf("plan not printed:\n%s", out)
	}
}

func TestInitNonInteractiveRequiresFlags(t *testing.T) {
	code, _, errOut := runApp(t, false, "", "init", "--dir", t.TempDir(), "--agent", "claude")
	if code != ExitUsage || !strings.Contains(errOut, "--stack is required") {
		t.Errorf("exit %d, stderr %q", code, errOut)
	}
	code, _, errOut = runApp(t, false, "", "init", "--dir", t.TempDir(), "--stack", "rn")
	if code != ExitUsage || !strings.Contains(errOut, "--agent is required") {
		t.Errorf("exit %d, stderr %q", code, errOut)
	}
}

func TestInitInteractiveWizard(t *testing.T) {
	dir := t.TempDir()
	// stacks: 1,2 · path rn: Enter (apps/mobile) · path next: web · agents: 1 · confirm: y
	stdin := "1,2\n\nweb\n1\ny\n"
	code, out, errOut := runApp(t, true, stdin, "init", "--dir", dir)
	if code != ExitOK {
		t.Fatalf("exit %d, stderr %s\nstdout %s", code, errOut, out)
	}
	m := loadManifest(t, dir)
	want := []manifest.StackRef{{ID: "react-native", Path: "apps/mobile"}, {ID: "nextjs", Path: "web"}}
	if !reflect.DeepEqual(m.Stacks, want) || !reflect.DeepEqual(m.Agents, []string{"claude"}) {
		t.Errorf("unexpected manifest: %+v", m)
	}
	if m.Project != filepath.Base(dir) {
		t.Errorf("project = %q, want dir name", m.Project)
	}
}

func TestInitInteractiveCancel(t *testing.T) {
	dir := t.TempDir()
	code, _, _ := runApp(t, true, "1\n1\nn\n", "init", "--dir", dir)
	if code != ExitFail {
		t.Errorf("exit %d, want %d", code, ExitFail)
	}
	if _, err := os.Stat(filepath.Join(dir, "AGENTS.md")); !os.IsNotExist(err) {
		t.Error("files written after cancelling")
	}
}

func TestInitRerunKeepsInstalledAndAddsNew(t *testing.T) {
	dir := t.TempDir()
	if code, _, e := runApp(t, false, "", "init", "--dir", dir, "--stack", "rn", "--agent", "claude"); code != ExitOK {
		t.Fatal(e)
	}
	// Re-run adding Next.js and OpenCode: rn keeps ".", next gets the default path.
	if code, _, e := runApp(t, false, "", "init", "--dir", dir, "--stack", "next", "--agent", "opencode"); code != ExitOK {
		t.Fatal(e)
	}
	m := loadManifest(t, dir)
	wantStacks := []manifest.StackRef{{ID: "react-native", Path: "."}, {ID: "nextjs", Path: "apps/web"}}
	if !reflect.DeepEqual(m.Stacks, wantStacks) || !reflect.DeepEqual(m.Agents, []string{"claude", "opencode"}) {
		t.Errorf("unexpected manifest: %+v", m)
	}

	// Re-run without flags in CI mode reuses the manifest and changes nothing.
	code, out, e := runApp(t, false, "", "init", "--dir", dir)
	if code != ExitOK {
		t.Fatal(e)
	}
	if !strings.Contains(out, "0 to create, 0 to update") {
		t.Errorf("expected no changes:\n%s", out)
	}
}

func TestInitPathFlag(t *testing.T) {
	dir := t.TempDir()
	code, _, e := runApp(t, false, "", "init", "--dir", dir, "--stack", "rn,next", "--agent", "claude", "--path", "rn=mobile, next=web")
	if code != ExitOK {
		t.Fatal(e)
	}
	got := loadManifest(t, dir).Stacks
	want := []manifest.StackRef{{ID: "react-native", Path: "mobile"}, {ID: "nextjs", Path: "web"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v", got)
	}
}

func TestInitErrors(t *testing.T) {
	cases := map[string][]string{
		"unknown stack":  {"init", "--dir", t.TempDir(), "--stack", "flutter", "--agent", "claude"},
		"bad path entry": {"init", "--dir", t.TempDir(), "--stack", "rn", "--agent", "claude", "--path", "rn"},
		"missing dir":    {"init", "--dir", filepath.Join(t.TempDir(), "nope"), "--stack", "rn", "--agent", "claude"},
		"extra argument": {"init", "extra"},
		"unknown flag":   {"init", "--nope"},
	}
	for name, args := range cases {
		if code, _, _ := runApp(t, false, "", args...); code != ExitUsage {
			t.Errorf("%s: exit %d, want %d", name, code, ExitUsage)
		}
	}
}

func TestParsePaths(t *testing.T) {
	got, err := parsePaths("rn=apps/mobile, nextjs = apps/web ,")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"react-native": "apps/mobile", "nextjs": "apps/web"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v", got)
	}
	for _, bad := range []string{"rn", "rn=", "vue=web"} {
		if _, err := parsePaths(bad); err == nil {
			t.Errorf("parsePaths(%q) should fail", bad)
		}
	}
}

func TestUnionInOrder(t *testing.T) {
	got := unionInOrder([]string{"a", "b", "c"}, []string{"c"}, []string{"a", "c"})
	if !reflect.DeepEqual(got, []string{"a", "c"}) {
		t.Errorf("got %v", got)
	}
}

func TestRunTopLevel(t *testing.T) {
	if code, out, _ := runApp(t, false, "", "version"); code != ExitOK || out != "eva test\n" {
		t.Errorf("version: %d %q", code, out)
	}
	if code, out, _ := runApp(t, false, ""); code != ExitOK || !strings.Contains(out, "Usage:") {
		t.Errorf("no args: %d %q", code, out)
	}
	if code, _, _ := runApp(t, false, "", "deploy"); code != ExitUsage {
		t.Errorf("unknown command exit %d", code)
	}
}

// editLock lets a test pretend a file was installed by an older eva.
func editLock(t *testing.T, dir string, edit func(files map[string]string)) {
	t.Helper()
	p := filepath.Join(dir, ".eva", "eva-lock.json")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	var l installer.Lock
	if err := json.Unmarshal(b, &l); err != nil {
		t.Fatal(err)
	}
	edit(l.Files)
	out, err := l.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, out, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestInitPruneDeletesObsoleteUntouchedFiles(t *testing.T) {
	dir := t.TempDir()
	if code, _, e := runApp(t, false, "", "init", "--dir", dir, "--stack", "rn", "--agent", "claude"); code != ExitOK {
		t.Fatal(e)
	}
	old := filepath.Join(dir, ".claude", "commands", "sdd-plan.md")
	if err := os.WriteFile(old, []byte("old shortcut\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	editLock(t, dir, func(files map[string]string) {
		files[".claude/commands/sdd-plan.md"] = installer.HashContent([]byte("old shortcut\n"))
	})

	code, out, e := runApp(t, false, "", "init", "--dir", dir)
	if code != ExitOK || !strings.Contains(out, "skip      .claude/commands/sdd-plan.md") || !strings.Contains(out, "--prune") {
		t.Fatalf("without --prune: exit %d\n%s\n%s", code, out, e)
	}
	code, out, e = runApp(t, false, "", "init", "--dir", dir, "--prune")
	if code != ExitOK || !strings.Contains(out, "delete    .claude/commands/sdd-plan.md") || !strings.Contains(out, "1 to delete") {
		t.Fatalf("with --prune: exit %d\n%s\n%s", code, out, e)
	}
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Error("obsolete file not deleted")
	}
}

func TestInitDiffShowsTemplateChangesOfEditedFiles(t *testing.T) {
	dir := t.TempDir()
	if code, _, e := runApp(t, false, "", "init", "--dir", dir, "--stack", "rn", "--agent", "claude"); code != ExitOK {
		t.Fatal(e)
	}
	plan := filepath.Join(dir, ".eva", "commands", "plan.md")
	if err := os.WriteFile(plan, []byte("my plan instructions\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Pretend the installed template was different from today's.
	editLock(t, dir, func(files map[string]string) {
		files[".eva/commands/plan.md"] = installer.HashContent([]byte("an older template\n"))
	})

	code, out, e := runApp(t, false, "", "init", "--dir", dir, "--dry-run", "--diff")
	if code != ExitOK {
		t.Fatalf("exit %d\n%s", code, e)
	}
	for _, want := range []string{"skip      .eva/commands/plan.md", "--- .eva/commands/plan.md (yours)", "+++ .eva/commands/plan.md (template)", "-my plan instructions", "+# eva-plan"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	if got, _ := os.ReadFile(plan); string(got) != "my plan instructions\n" {
		t.Error("--dry-run --diff must not write")
	}

	// Without --diff the hint is there but no diff is printed.
	_, out, _ = runApp(t, false, "", "init", "--dir", dir, "--dry-run")
	if strings.Contains(out, "+++ ") || !strings.Contains(out, "--diff") {
		t.Errorf("unexpected output without --diff:\n%s", out)
	}
}
