package installer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHashContentNormalisesLineEndings(t *testing.T) {
	lf := HashContent([]byte("a\nb\n"))
	if crlf := HashContent([]byte("a\r\nb\r\n")); crlf != lf {
		t.Errorf("CRLF hash %s != LF hash %s", crlf, lf)
	}
	if !strings.HasPrefix(lf, "sha256:") || len(lf) != len("sha256:")+64 {
		t.Errorf("unexpected hash format %q", lf)
	}
	if HashContent([]byte("a\nb")) == lf {
		t.Error("different content must hash differently")
	}
}

func TestLoadLock(t *testing.T) {
	root := t.TempDir()
	if l, err := LoadLock(root); err != nil || l != nil {
		t.Fatalf("missing lock: got %v, %v; want nil, nil", l, err)
	}

	writeFile(t, filepath.Join(root, LockPath), `{"files":{"a.md":"sha256:x"}}`)
	l, err := LoadLock(root)
	if err != nil || l.Files["a.md"] != "sha256:x" {
		t.Fatalf("got %+v, %v", l, err)
	}

	writeFile(t, filepath.Join(root, LockPath), `{not json`)
	if _, err := LoadLock(root); err == nil {
		t.Error("invalid lock should be an error")
	}
}

func TestLockMarshalIsStableAndSaysGenerated(t *testing.T) {
	l := Lock{EvaVersion: "1.1.0", Files: map[string]string{"b.md": "sha256:2", "a.md": "sha256:1"}}
	out, err := l.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	if !strings.HasPrefix(s, "{\n  \"_generated\": \"Written by `eva init`. Do not edit. Commit it.\",\n") {
		t.Errorf("lock must start with the _generated notice:\n%s", s)
	}
	if strings.Index(s, `"a.md"`) > strings.Index(s, `"b.md"`) {
		t.Errorf("files must be sorted:\n%s", s)
	}
	if !strings.HasSuffix(s, "}\n") {
		t.Error("lock must end with a newline")
	}
}

func TestFreshInstallWritesLockAndGitattributes(t *testing.T) {
	root := t.TempDir()
	mustApply(t, root, goldenCases["rn-claude"], Options{})

	l, err := LoadLock(root)
	if err != nil || l == nil {
		t.Fatalf("lock not written: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(root, ".eva", "constitution.md"))
	if err != nil {
		t.Fatal(err)
	}
	if got := l.Files[".eva/constitution.md"]; got != HashContent(b) {
		t.Errorf("constitution hash = %q, want %q", got, HashContent(b))
	}
	for _, p := range []string{"AGENTS.md", "CLAUDE.md", ".gitattributes", ".eva/config.json", LockPath} {
		if _, ok := l.Files[p]; ok {
			t.Errorf("%s is managed by eva and must not be in the lock", p)
		}
	}
	ga := readFile(t, filepath.Join(root, ".gitattributes"))
	if !strings.Contains(ga, "# eva:start") || !strings.Contains(ga, ".eva/eva-lock.json linguist-generated=true") || !strings.Contains(ga, "# eva:end") {
		t.Errorf("unexpected .gitattributes:\n%s", ga)
	}
}
