package installer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/codebymaribel/eva-sdd/internal/manifest"
)

// upgradeCfg is the configuration used with the small template trees below.
var upgradeCfg = Config{
	Project: "demo", EvaVersion: "test",
	Stacks: []manifest.StackRef{{ID: "nextjs", Path: "."}},
	Agents: []string{"claude"},
}

// templatesV returns a minimal template tree; files maps core/ paths to content.
func templatesV(files map[string]string) fstest.MapFS {
	tfs := fstest.MapFS{
		"stacks/nextjs/.eva/stacks/nextjs.md":        {Data: []byte("next rules\n")},
		"agents/claude/.claude/commands/eva-plan.md": {Data: []byte("plan shortcut\n")},
	}
	for p, c := range files {
		tfs["core/"+p] = &fstest.MapFile{Data: []byte(c)}
	}
	return tfs
}

func plan(t *testing.T, root string, tfs fstest.MapFS, opts Options) []Action {
	t.Helper()
	actions, err := Plan(root, tfs, upgradeCfg, opts)
	if err != nil {
		t.Fatal(err)
	}
	return actions
}

func install(t *testing.T, root string, tfs fstest.MapFS, opts Options) []Action {
	t.Helper()
	actions := plan(t, root, tfs, opts)
	mustApplyActions(t, root, actions)
	return actions
}

func TestUntouchedFileIsUpdatedWhenTemplateChanges(t *testing.T) {
	root := t.TempDir()
	install(t, root, templatesV(map[string]string{".eva/commands/plan.md": "v1\n"}), Options{})

	v2 := templatesV(map[string]string{".eva/commands/plan.md": "v2\n"})
	a := find(t, plan(t, root, v2, Options{}), ".eva/commands/plan.md")
	if a.Op != OpUpdate || !strings.Contains(a.Note, "template updated") {
		t.Fatalf("got %s (%s), want update with a 'template updated' note", a.Op, a.Note)
	}
	install(t, root, v2, Options{})
	if got := readFile(t, filepath.Join(root, ".eva/commands/plan.md")); got != "v2\n" {
		t.Errorf("file = %q, want v2", got)
	}
	l, _ := LoadLock(root)
	if l.Files[".eva/commands/plan.md"] != HashContent([]byte("v2\n")) {
		t.Error("lock must record the new hash after the update")
	}
}

func TestEditedFileIsKeptWhenTemplateChanges(t *testing.T) {
	root := t.TempDir()
	install(t, root, templatesV(map[string]string{".eva/commands/plan.md": "v1\n"}), Options{})
	writeFile(t, filepath.Join(root, ".eva/commands/plan.md"), "mine\n")

	v2 := templatesV(map[string]string{".eva/commands/plan.md": "v2\n"})
	actions := install(t, root, v2, Options{})
	a := find(t, actions, ".eva/commands/plan.md")
	if a.Op != OpSkip || !strings.Contains(a.Note, "template changed") || !strings.Contains(a.Note, "--diff") {
		t.Fatalf("got %s (%s), want skip mentioning the template change and --diff", a.Op, a.Note)
	}
	if !a.Diffable() {
		t.Error("an edited file whose template changed must be shown by --diff")
	}
	if string(a.Current) != "mine\n" || string(a.Content) != "v2\n" {
		t.Errorf("skip must carry the current file and the new template for --diff: %q / %q", a.Current, a.Content)
	}
	if got := readFile(t, filepath.Join(root, ".eva/commands/plan.md")); got != "mine\n" {
		t.Errorf("edited file overwritten: %q", got)
	}
	l, _ := LoadLock(root)
	if l.Files[".eva/commands/plan.md"] != HashContent([]byte("v1\n")) {
		t.Error("a skipped file keeps the hash of what eva installed")
	}

	install(t, root, v2, Options{Force: true})
	if got := readFile(t, filepath.Join(root, ".eva/commands/plan.md")); got != "v2\n" {
		t.Errorf("--force must overwrite: %q", got)
	}
}

func TestEditedFileWithSameTemplateIsKept(t *testing.T) {
	root := t.TempDir()
	v1 := templatesV(map[string]string{".eva/constitution.md": "TODO\n"})
	install(t, root, v1, Options{})
	writeFile(t, filepath.Join(root, ".eva/constitution.md"), "my rules\n")

	a := find(t, plan(t, root, v1, Options{}), ".eva/constitution.md")
	if a.Op != OpSkip || a.Note != "edited locally, kept" {
		t.Errorf("got %s (%q)", a.Op, a.Note)
	}
	if a.Diffable() {
		t.Error("--diff must not show files the user edited when the template did not change")
	}
}

func TestCRLFCheckoutCountsAsUntouched(t *testing.T) {
	root := t.TempDir()
	install(t, root, templatesV(map[string]string{".eva/commands/plan.md": "line 1\nline 2\n"}), Options{})
	writeFile(t, filepath.Join(root, ".eva/commands/plan.md"), "line 1\r\nline 2\r\n")

	if a := find(t, plan(t, root, templatesV(map[string]string{".eva/commands/plan.md": "line 1\nline 2\n"}), Options{}), ".eva/commands/plan.md"); a.Op != OpUnchanged {
		t.Errorf("same template: got %s, want unchanged", a.Op)
	}
	if a := find(t, plan(t, root, templatesV(map[string]string{".eva/commands/plan.md": "v2\n"}), Options{}), ".eva/commands/plan.md"); a.Op != OpUpdate {
		t.Errorf("new template: got %s (%s), want update", a.Op, a.Note)
	}
}

func TestObsoleteFiles(t *testing.T) {
	v1 := templatesV(map[string]string{".eva/commands/old.md": "old\n", ".eva/commands/kept.md": "kept\n", ".eva/commands/edited.md": "x\n"})
	v2 := templatesV(map[string]string{".eva/commands/kept.md": "kept\n"})

	root := t.TempDir()
	install(t, root, v1, Options{})
	writeFile(t, filepath.Join(root, ".eva/commands/edited.md"), "mine\n")

	actions := install(t, root, v2, Options{})
	if a := find(t, actions, ".eva/commands/old.md"); a.Op != OpSkip || !strings.Contains(a.Note, "--prune") {
		t.Errorf("untouched obsolete without --prune: got %s (%s)", a.Op, a.Note)
	}
	if a := find(t, actions, ".eva/commands/edited.md"); a.Op != OpSkip || !strings.Contains(a.Note, "edited") {
		t.Errorf("edited obsolete: got %s (%s)", a.Op, a.Note)
	}
	if _, err := os.Stat(filepath.Join(root, ".eva/commands/old.md")); err != nil {
		t.Error("obsolete file deleted without --prune")
	}
	l, _ := LoadLock(root)
	if _, ok := l.Files[".eva/commands/old.md"]; !ok {
		t.Error("untouched obsolete file stays in the lock so --prune can delete it later")
	}
	if _, ok := l.Files[".eva/commands/edited.md"]; ok {
		t.Error("an edited obsolete file is the user's now and leaves the lock")
	}

	actions = install(t, root, v2, Options{Prune: true})
	if a := find(t, actions, ".eva/commands/old.md"); a.Op != OpDelete {
		t.Errorf("with --prune: got %s, want delete", a.Op)
	}
	if _, err := os.Stat(filepath.Join(root, ".eva/commands/old.md")); !os.IsNotExist(err) {
		t.Error("--prune did not delete the untouched obsolete file")
	}
	if got := readFile(t, filepath.Join(root, ".eva/commands/edited.md")); got != "mine\n" {
		t.Error("--prune must never delete an edited file")
	}
	l, _ = LoadLock(root)
	if _, ok := l.Files[".eva/commands/old.md"]; ok {
		t.Error("deleted file must leave the lock")
	}
}

func TestObsoleteFileAlreadyGoneIsForgotten(t *testing.T) {
	root := t.TempDir()
	install(t, root, templatesV(map[string]string{".eva/commands/old.md": "old\n"}), Options{})
	if err := os.Remove(filepath.Join(root, ".eva/commands/old.md")); err != nil {
		t.Fatal(err)
	}
	actions := install(t, root, templatesV(map[string]string{".eva/commands/other.md": "other\n"}), Options{})
	for _, a := range actions {
		if a.Path == ".eva/commands/old.md" {
			t.Errorf("a missing obsolete file needs no action, got %s", a.Op)
		}
	}
	l, _ := LoadLock(root)
	if _, ok := l.Files[".eva/commands/old.md"]; ok {
		t.Error("missing obsolete file must leave the lock")
	}
}

func TestProjectsWithoutLockAreMigrated(t *testing.T) {
	root := t.TempDir()
	v1 := templatesV(map[string]string{".eva/commands/same.md": "same\n", ".eva/commands/diff.md": "v1\n"})
	install(t, root, v1, Options{})
	if err := os.Remove(filepath.Join(root, LockPath)); err != nil {
		t.Fatal(err)
	}

	v2 := templatesV(map[string]string{".eva/commands/same.md": "same\n", ".eva/commands/diff.md": "v2\n"})
	actions := install(t, root, v2, Options{})
	if a := find(t, actions, ".eva/commands/same.md"); a.Op != OpUnchanged {
		t.Errorf("identical file: got %s", a.Op)
	}
	if a := find(t, actions, ".eva/commands/diff.md"); a.Op != OpSkip || !strings.Contains(a.Note, "no install record") {
		t.Errorf("differing file without record: got %s (%s)", a.Op, a.Note)
	}
	if a := find(t, actions, LockPath); a.Op != OpCreate {
		t.Errorf("lock: got %s, want create", a.Op)
	}
	l, _ := LoadLock(root)
	if l.Files[".eva/commands/same.md"] == "" {
		t.Error("identical file must be recorded on the first run")
	}
	if _, ok := l.Files[".eva/commands/diff.md"]; ok {
		t.Error("a differing file without record must not be recorded")
	}
}

func TestGitattributesKeepsUserLines(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".gitattributes"), "*.png binary\n")
	install(t, root, templatesV(map[string]string{".gitattributes.block": ".eva/eva-lock.json linguist-generated=true\n"}), Options{})
	got := readFile(t, filepath.Join(root, ".gitattributes"))
	if !strings.HasPrefix(got, "*.png binary\n\n# eva:start") || strings.Contains(got, "<!--") {
		t.Errorf("unexpected .gitattributes:\n%s", got)
	}
}
