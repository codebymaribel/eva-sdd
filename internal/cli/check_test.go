package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeSpec writes a spec that passes every document rule for its status.
func writeSpec(t *testing.T, dir, feature, status string) {
	t.Helper()
	by, at := "", ""
	if status == "approved" {
		by, at = "dev", "2026-01-02"
	}
	p := filepath.Join(dir, "specs", feature, "spec.md")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\nid: " + feature + "\ntype: spec\nstatus: " + status + "\ncreated: 2026-01-01\napproved_by: " + by + "\napproved_at: " + at + "\nclosed_at:\n---\n" +
		"## 4. Functional requirements\n\n- **FR-1** When the user signs in, the app shall show the home screen.\n\n" +
		"## 7. Completion criteria\n\n- FR-1 has a passing test.\n"
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if status == "approved" { // an approved spec needs a clarify.md without open findings
		clarify := "---\nid: " + feature + "\ntype: clarify\ncreated: 2026-01-01\nclosed_at:\n---\n# Clarify\n\nNo findings.\n"
		if err := os.WriteFile(filepath.Join(filepath.Dir(p), "clarify.md"), []byte(clarify), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func initProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if code, _, e := runApp(t, false, "", "init", "--dir", dir, "--stack", "next", "--agent", "claude"); code != ExitOK {
		t.Fatal(e)
	}
	return dir
}

// approveConstitution simulates a developer approving the generated constitution.
func approveConstitution(t *testing.T, dir string) {
	t.Helper()
	p := filepath.Join(dir, ".eva", "constitution.md")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	s := strings.Replace(string(b), "status: draft", "status: approved", 1)
	s = strings.Replace(s, "approved_by:", "approved_by: dev", 1)
	s = strings.Replace(s, "\ncreated:", "\ncreated: 2026-01-01", 1)
	s = strings.Replace(s, "approved_at:", "approved_at: 2026-01-01", 1)
	s = strings.ReplaceAll(s, "TODO", "done") // the developer filled every placeholder
	if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCheckFreshProjectWarnsAboutConstitution(t *testing.T) {
	dir := initProject(t)
	code, out, errOut := runApp(t, false, "", "check", "--dir", dir, "--no-git")
	if code != ExitOK || !strings.Contains(out, "WARN  EVA053") {
		t.Errorf("exit %d\nstdout %s\nstderr %s", code, out, errOut)
	}
}

func TestCheckPassesWithApprovedConstitution(t *testing.T) {
	dir := initProject(t)
	approveConstitution(t, dir)
	code, out, errOut := runApp(t, false, "", "check", "--dir", dir, "--no-git")
	if code != ExitOK || !strings.Contains(out, "OK") {
		t.Errorf("exit %d\nstdout %s\nstderr %s", code, out, errOut)
	}
}

func TestCheckFailsOnErrors(t *testing.T) {
	dir := initProject(t)
	approveConstitution(t, dir)
	writeSpec(t, dir, "001-login", "approved") // approved spec…
	planPath := filepath.Join(dir, "specs", "001-login", "plan.md")
	if err := os.WriteFile(planPath, []byte("---\nid: 001-login\ntype: plan\nstatus: approved\ncreated: 2026-01-03\napproved_at: 2026-01-03\n---\n## 2. Requirements traceability\n\n| FR-1 | §4 |\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, out, _ := runApp(t, false, "", "check", "--dir", dir, "--no-git")
	if code != ExitFail {
		t.Errorf("exit %d, want %d", code, ExitFail)
	}
	if !strings.Contains(out, "ERROR EVA024 specs/001-login/plan.md") || !strings.Contains(out, "1 error(s)") {
		t.Errorf("unexpected output:\n%s", out)
	}
}

func TestCheckWarningsDoNotFail(t *testing.T) {
	dir := initProject(t)
	// Not a git repository → EVA040 warning only.
	code, out, _ := runApp(t, false, "", "check", "--dir", dir)
	if code != ExitOK || !strings.Contains(out, "WARN  EVA040") {
		t.Errorf("exit %d\n%s", code, out)
	}
}

func TestCheckWithoutInit(t *testing.T) {
	code, out, _ := runApp(t, false, "", "check", "--dir", t.TempDir(), "--no-git")
	if code != ExitFail || !strings.Contains(out, "EVA001") {
		t.Errorf("exit %d\n%s", code, out)
	}
}

func TestCheckUsageErrors(t *testing.T) {
	if code, _, _ := runApp(t, false, "", "check", "extra"); code != ExitUsage {
		t.Errorf("extra arg exit %d", code)
	}
	if code, _, _ := runApp(t, false, "", "check", "--dir", filepath.Join(t.TempDir(), "nope")); code != ExitUsage {
		t.Errorf("missing dir exit %d", code)
	}
	if code, _, _ := runApp(t, false, "", "check", "-h"); code != ExitOK {
		t.Errorf("help exit %d", code)
	}
}
