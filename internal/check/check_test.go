package check

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// --- fixtures ---

type project struct {
	t    *testing.T
	root string
}

func newProject(t *testing.T) *project {
	t.Helper()
	p := &project{t: t, root: t.TempDir()}
	p.write(".eva/config.json", `{"schemaVersion":1,"project":"demo","stacks":[{"id":"nextjs","path":"."}],"agents":["claude"]}`)
	p.constitution("approved", "# Constitution\n")
	return p
}

// testDate is the date used for every front-matter date in fixtures.
const testDate = "2026-01-01"

// specBody is the smallest spec body that passes the content rules of an
// approved spec: one EARS requirement and a completion criterion.
const specBody = "## 4. Functional requirements\n\n" +
	"- **FR-1** (US-1) When the user submits the form, the app shall show a confirmation.\n\n" +
	"## 7. Completion criteria\n\n- FR-1 has a passing test.\n"

// planBody is the smallest plan body that traces every requirement of specBody.
const planBody = "## 2. Requirements traceability\n\n| FR | Plan sections | Files |\n|----|---------------|-------|\n| FR-1 | §4 | `src/form.tsx` |\n"

// tasks returns a valid tasks.md body with one feat task per element, citing
// FR-1; true means the task is ticked.
func tasks(ticked ...bool) string {
	var b strings.Builder
	for i, done := range ticked {
		box := " "
		if done {
			box = "x"
		}
		fmt.Fprintf(&b, "- [%s] T%03d [feat] [US-1] Show the confirmation — FR-1\n", box, i+1)
		b.WriteString("  - Files: `src/form.tsx`\n  - Tests: `src/form.test.tsx` — FR-1\n  - Done when: the FR-1 test passes\n")
	}
	return b.String()
}

// verificationBody is a passing verification of specBody.
const verificationBody = "## 2. Requirements traceability\n\n| FR | Test(s) | Result |\n|----|---------|--------|\n| FR-1 | `src/form.test.tsx` › \"FR-1: …\" | ✅ |\n\n## 6. Verdict\n\n**Verdict:** PASS\n"

// front renders a front-matter block from key/value pairs, in order.
func front(kv ...string) string {
	var b strings.Builder
	b.WriteString("---\n")
	for i := 0; i+1 < len(kv); i += 2 {
		b.WriteString(kv[i] + ": " + kv[i+1] + "\n")
	}
	b.WriteString("---\n")
	return b.String()
}

// lifecycle returns the approval and closing fields a human fills for status.
func lifecycle(status string) []string {
	by, at, closed := "", "", ""
	if status == "approved" || status == "done" {
		by, at = "Maribel", testDate
	}
	if status == "done" {
		closed = testDate
	}
	return []string{"approved_by", by, "approved_at", at, "closed_at", closed}
}

// constitution writes .eva/constitution.md with consistent dates.
func (p *project) constitution(status, body string) {
	p.t.Helper()
	kv := append([]string{"type", "constitution", "status", status, "version", "1", "created", testDate}, lifecycle(status)...)
	p.write(ConstitutionPath, front(kv...)+body)
}

func (p *project) remove(rel string) {
	p.t.Helper()
	if err := os.Remove(filepath.Join(p.root, filepath.FromSlash(rel))); err != nil {
		p.t.Fatal(err)
	}
}

func (p *project) write(rel, content string) {
	p.t.Helper()
	full := filepath.Join(p.root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		p.t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		p.t.Fatal(err)
	}
}

// doc writes an SDD document whose type, dates and approval fields match status.
// An approved or done spec also gets a clarify.md without findings, because
// eva requires one before a spec can be approved, and a done spec gets a
// passing verification.md.
func (p *project) doc(feature, name, status, body string) {
	p.t.Helper()
	kv := append([]string{"id", feature, "type", strings.TrimSuffix(name, ".md"), "status", status, "created", testDate}, lifecycle(status)...)
	p.write("specs/"+feature+"/"+name, front(kv...)+body)
	if name == "spec.md" && (status == "approved" || status == "done") {
		p.clarify(feature, "")
	}
	if name == "spec.md" && status == "done" {
		p.verification(feature, verificationBody)
	}
}

// verification writes specs/<feature>/verification.md with the given body.
func (p *project) verification(feature, body string) {
	p.t.Helper()
	p.write("specs/"+feature+"/verification.md", front("id", feature, "type", "verification", "created", testDate, "closed_at", testDate)+body)
}

// clarify writes specs/<feature>/clarify.md with the given findings.
func (p *project) clarify(feature, findings string) {
	p.t.Helper()
	p.write("specs/"+feature+"/clarify.md", front("id", feature, "type", "clarify", "created", testDate, "closed_at", "")+"# Clarify\n\n## Findings\n\n"+findings)
}

// approvedFeature writes spec, plan and tasks, all approved.
func (p *project) approvedFeature(id string) {
	p.doc(id, "spec.md", "approved", specBody)
	p.doc(id, "plan.md", "approved", planBody)
	p.doc(id, "tasks.md", "approved", tasks(true, false))
}

func run(t *testing.T, p *project, opts Options) Report {
	t.Helper()
	if opts.Git == nil && !opts.NoGit {
		opts.NoGit = true
	}
	r, err := Run(p.root, opts)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func codes(r Report) map[string]Severity {
	out := map[string]Severity{}
	for _, f := range r.Findings {
		out[f.Code] = f.Severity
	}
	return out
}

func expectCodes(t *testing.T, r Report, want ...string) {
	t.Helper()
	got := codes(r)
	for _, c := range want {
		if _, ok := got[c]; !ok {
			t.Errorf("missing finding %s; got %+v", c, r.Findings)
		}
	}
	if len(got) != len(want) {
		t.Errorf("got findings %+v, want only %v", r.Findings, want)
	}
}

// fakeGit is an in-memory Git.
type fakeGit struct {
	repo    bool
	branch  string
	refs    map[string]bool
	changed []string
}

func (g fakeGit) IsRepo() bool                          { return g.repo }
func (g fakeGit) CurrentBranch() (string, error)        { return g.branch, nil }
func (g fakeGit) RefExists(ref string) bool             { return g.refs[ref] }
func (g fakeGit) ChangedFiles(string) ([]string, error) { return g.changed, nil }

func repo(branch string, changed ...string) fakeGit {
	return fakeGit{repo: true, branch: branch, refs: map[string]bool{"main": true}, changed: changed}
}

// --- document rules ---

func TestMissingManifest(t *testing.T) {
	r, err := Run(t.TempDir(), Options{NoGit: true})
	if err != nil {
		t.Fatal(err)
	}
	expectCodes(t, r, "EVA001")
}

func TestCleanProjectHasNoFindings(t *testing.T) {
	p := newProject(t)
	p.approvedFeature("001-login")
	p.doc("002-profile", "spec.md", "in-review", "Needs [NEEDS CLARIFICATION: avatar size?]\n")
	r := run(t, p, Options{})
	expectCodes(t, r)
	if len(r.Features) != 2 {
		t.Errorf("features = %d, want 2", len(r.Features))
	}
}

func TestDocumentRules(t *testing.T) {
	cases := map[string]struct {
		setup func(p *project)
		want  []string
	}{
		"missing spec": {
			func(p *project) { p.write("specs/001-a/notes.md", "x") },
			[]string{"EVA020"},
		},
		"no front-matter": {
			func(p *project) { p.write("specs/001-a/spec.md", "# Spec\n") },
			[]string{"EVA021"},
		},
		"id mismatch": {
			func(p *project) {
				p.write("specs/001-a/spec.md", front("id", "001-b", "type", "spec", "status", "draft"))
			},
			[]string{"EVA022"},
		},
		"invalid status": {
			func(p *project) {
				p.write("specs/001-a/spec.md", front("id", "001-a", "type", "spec", "status", "ready"))
			},
			[]string{"EVA023"},
		},
		"approved without approver": {
			func(p *project) {
				p.write("specs/001-a/spec.md", front("id", "001-a", "type", "spec", "status", "approved", "created", testDate, "approved_by", "", "approved_at", testDate)+specBody)
				p.clarify("001-a", "")
			},
			[]string{"EVA024"},
		},
		"approved spec with open questions": {
			func(p *project) {
				p.doc("001-a", "spec.md", "approved", specBody+"SSO? [NEEDS CLARIFICATION: which IdP]\n")
			},
			[]string{"EVA025"},
		},
		"plan before spec approval": {
			func(p *project) {
				p.doc("001-a", "spec.md", "in-review", "")
				p.doc("001-a", "plan.md", "draft", "")
			},
			[]string{"EVA030"},
		},
		"tasks before plan approval": {
			func(p *project) {
				p.doc("001-a", "spec.md", "approved", specBody)
				p.doc("001-a", "plan.md", "in-review", planBody)
				p.doc("001-a", "tasks.md", "draft", "- [ ] T001\n")
			},
			[]string{"EVA031"},
		},
		"ticked tasks before approval": {
			func(p *project) {
				p.doc("001-a", "spec.md", "approved", specBody)
				p.doc("001-a", "plan.md", "approved", planBody)
				p.doc("001-a", "tasks.md", "in-review", tasks(true))
			},
			[]string{"EVA032"},
		},
		"done with open tasks": {
			func(p *project) {
				p.doc("001-a", "spec.md", "done", specBody)
				p.doc("001-a", "plan.md", "approved", planBody)
				p.doc("001-a", "tasks.md", "approved", tasks(true, false))
			},
			[]string{"EVA033"},
		},
		"bad folder name": {
			func(p *project) { p.write("specs/Login Feature/spec.md", "x") },
			[]string{"EVA010"},
		},
		"duplicate numbers": {
			func(p *project) {
				p.doc("004-login", "spec.md", "draft", "")
				p.doc("004-signup", "spec.md", "draft", "")
			},
			[]string{"EVA011"},
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			p := newProject(t)
			c.setup(p)
			expectCodes(t, run(t, p, Options{}), c.want...)
		})
	}
}

func TestDoneSpecWithAllTasksTickedIsValid(t *testing.T) {
	p := newProject(t)
	p.doc("001-a", "spec.md", "done", specBody)
	p.doc("001-a", "plan.md", "approved", planBody)
	p.doc("001-a", "tasks.md", "approved", tasks(true, true))
	expectCodes(t, run(t, p, Options{}))
}

// --- branch gate ---

func TestGate(t *testing.T) {
	cases := map[string]struct {
		setup  func(p *project)
		git    fakeGit
		opts   Options
		want   []string
		strict bool
	}{
		"approved feature with code": {
			setup: func(p *project) { p.approvedFeature("004-login") },
			git:   repo("feat/004-login", "src/app/login/page.tsx"),
		},
		"plan not approved with code": {
			setup: func(p *project) {
				p.doc("004-login", "spec.md", "approved", specBody)
				p.doc("004-login", "plan.md", "in-review", planBody)
			},
			git:  repo("feat/004-login", "src/app/login/page.tsx"),
			want: []string{"EVA042"},
		},
		"only spec changes are allowed before approval": {
			setup: func(p *project) { p.doc("004-login", "spec.md", "draft", "") },
			git:   repo("feat/004-login", "specs/004-login/spec.md", "AGENTS.md", "docs/adr/0001-x.md"),
		},
		"branch refers to missing feature with code": {
			setup: func(p *project) {},
			git:   repo("feat/009-search", "src/search.ts"),
			want:  []string{"EVA041"},
		},
		"code on branch without feature id": {
			setup: func(p *project) {},
			git:   repo("fix-typo", "src/a.ts"),
			want:  []string{"EVA043"},
		},
		"code on branch without feature id (strict)": {
			setup:  func(p *project) {},
			git:    repo("fix-typo", "src/a.ts"),
			want:   []string{"EVA043"},
			strict: true,
		},
		"on base branch": {
			setup: func(p *project) {},
			git:   repo("main", "src/a.ts"),
		},
		"not a repo": {
			setup: func(p *project) {},
			git:   fakeGit{},
			want:  []string{"EVA040"},
		},
		"no base ref": {
			setup: func(p *project) {},
			git:   fakeGit{repo: true, branch: "feat/001-a"},
			want:  []string{"EVA044"},
		},
		"detached head": {
			setup: func(p *project) {},
			git:   fakeGit{repo: true, refs: map[string]bool{"main": true}},
			want:  []string{"EVA045"},
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			p := newProject(t)
			c.setup(p)
			opts := c.opts
			opts.Git = c.git
			opts.Strict = c.strict
			r := run(t, p, opts)
			expectCodes(t, r, c.want...)
			if c.strict && codes(r)["EVA043"] != SevError {
				t.Error("EVA043 should be an error in strict mode")
			}
			if !c.strict && len(c.want) == 1 && c.want[0] == "EVA043" && codes(r)["EVA043"] != SevWarning {
				t.Error("EVA043 should be a warning by default")
			}
		})
	}
}

func TestGateUsesGitHubEnvironment(t *testing.T) {
	p := newProject(t)
	p.doc("007-cart", "spec.md", "in-review", "")
	g := fakeGit{repo: true, branch: "", refs: map[string]bool{"origin/develop": true}, changed: []string{"src/cart.ts"}}
	env := map[string]string{"GITHUB_HEAD_REF": "feat/007-cart", "GITHUB_BASE_REF": "develop"}
	r := run(t, p, Options{Git: g, Getenv: func(k string) string { return env[k] }})
	expectCodes(t, r, "EVA042")
	if r.Branch != "feat/007-cart" || r.Base != "origin/develop" || !r.GateChecked || r.CodeChanges != 1 {
		t.Errorf("unexpected report: %+v", r)
	}
}

func TestIsCodePath(t *testing.T) {
	for p, want := range map[string]bool{
		"src/app/page.tsx":         true,
		"package.json":             true,
		"apps/mobile/README.md":    true,
		"README.md":                false,
		"specs/001-a/spec.md":      false,
		".eva/constitution.md":     false,
		"docs/adr/0001.md":         false,
		".claude/commands/x.md":    false,
		".opencode/commands/x.md":  false,
		".github/workflows/ci.yml": true,
	} {
		if got := isCodePath(p); got != want {
			t.Errorf("isCodePath(%q) = %v, want %v", p, got, want)
		}
	}
}

// TestExecGitIntegration exercises the real git binary end to end.
func TestExecGitIntegration(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	p := newProject(t)
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = p.root
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("init", "-q", "-b", "main")
	git("add", ".")
	git("commit", "-q", "-m", "init")
	git("switch", "-q", "-c", "feat/001-login")
	p.doc("001-login", "spec.md", "approved", specBody)
	p.write("src/login.ts", "export {}\n") // untracked code change

	r, err := Run(p.root, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if r.Branch != "feat/001-login" || r.Base != "main" || r.CodeChanges != 1 {
		t.Errorf("unexpected report: branch=%q base=%q code=%d", r.Branch, r.Base, r.CodeChanges)
	}
	expectCodes(t, r, "EVA042")
	if !strings.Contains(r.Findings[0].Message, "plan.md (missing)") {
		t.Errorf("message = %q", r.Findings[0].Message)
	}
}
