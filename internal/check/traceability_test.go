package check

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseTasks(t *testing.T) {
	body := "## Phase 1\n\n" +
		"- [x] T001 [feat] [US-1] Add the form — FR-1, FR-2\n" +
		"  - Files: `src/Form.tsx`, `src/useForm.ts`, `src/Form.test.tsx`\n" +
		"  - Tests: `src/Form.test.tsx` — FR-1, FR-2\n" +
		"  - Done when: both tests pass\n" +
		"- [ ] T002 [P] [improve] Extract the hook\n" +
		"  - Files: `src/useForm.ts`\n" +
		"  - Tests: existing suites green\n" +
		"  - Done when: no behaviour change\n" +
		"<!-- - [ ] T009 [feat] commented out -->\n" +
		"## Implementation notes\n\n- T001 red: FR-1 test failed before the form existed.\n"
	got := ParseTasks(body)
	want := []Task{
		{ID: "T001", Type: "feat", Done: true, FRs: []string{"FR-1", "FR-2"},
			Files:    []string{"src/Form.tsx", "src/useForm.ts", "src/Form.test.tsx"},
			HasTests: true, TestFiles: []string{"src/Form.test.tsx"}, TestFRs: []string{"FR-1", "FR-2"},
			DoneWhen: "both tests pass"},
		{ID: "T002", Type: "improve", Done: false,
			Files:    []string{"src/useForm.ts"},
			HasTests: true, DoneWhen: "no behaviour change"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseTasks:\n got %+v\nwant %+v", got, want)
	}
}

func TestIsTestFile(t *testing.T) {
	for p, want := range map[string]bool{
		"src/Form.test.tsx":            true,
		"app/login/page.spec.ts":       true,
		"src/__tests__/Form.tsx":       true,
		"e2e/login.e2e.ts":             true,
		"e2e/flows/login.yaml":         true,
		"internal/check/check_test.go": true,
		"src/Form.tsx":                 false,
		"src/testing/helpers.ts":       false,
		"src/contest/Contest.tsx":      false,
	} {
		if got := IsTestFile(p); got != want {
			t.Errorf("IsTestFile(%q) = %v, want %v", p, got, want)
		}
	}
}

// task renders one task block for the rule tests.
func task(head, files, tests, done string) string {
	s := "- [ ] " + head + "\n"
	if files != "" {
		s += "  - Files: " + files + "\n"
	}
	if tests != "" {
		s += "  - Tests: " + tests + "\n"
	}
	if done != "" {
		s += "  - Done when: " + done + "\n"
	}
	return s
}

func TestTaskRules(t *testing.T) {
	const (
		files = "`src/a.ts`"
		tests = "`src/a.test.ts` — FR-1"
		done  = "the FR-1 test passes"
	)
	cases := map[string]struct {
		body string
		want []string
	}{
		"valid feat, fix and improve": {
			task("T001 [feat] [US-1] Add — FR-1", files, tests, done) +
				task("T002 [fix] Handle empty — FR-1", files, tests, done) +
				task("T003 [improve] Rename", files, "existing suites green", done),
			nil,
		},
		"missing type":    {task("T001 [US-1] Add — FR-1", files, tests, done), []string{"EVA080"}},
		"unknown type":    {task("T001 [chore] Add — FR-1", files, tests, done), []string{"EVA080"}},
		"unknown FR":      {task("T001 [feat] Add — FR-1, FR-9", files, tests, done), []string{"EVA081"}},
		"FR without task": {task("T001 [improve] Rename", files, "existing suites green", done), []string{"EVA082"}},
		"feat without FR": {
			task("T001 [feat] Add", files, "`src/a.test.ts`", done) + task("T002 [fix] Fix — FR-1", files, tests, done),
			[]string{"EVA083"},
		},
		"missing done when":         {task("T001 [feat] Add — FR-1", files, tests, ""), []string{"EVA084"}},
		"missing files":             {task("T001 [feat] Add — FR-1", "", tests, done), []string{"EVA084"}},
		"tests without a test file": {task("T001 [feat] Add — FR-1", files, "unit — FR-1", done), []string{"EVA085"}},
		"tests without FR":          {task("T001 [feat] Add — FR-1", files, "`src/a.test.ts`", done), []string{"EVA085"}},
		"missing tests":             {task("T001 [feat] Add — FR-1", files, "", done), []string{"EVA085"}},
		"four code files": {
			task("T001 [feat] Add — FR-1", "`a.ts`, `b.ts`, `c.ts`, `d.ts`, `a.test.ts`, `b.test.ts`", tests, done),
			[]string{"EVA086"},
		},
		"duplicate id": {
			task("T001 [feat] Add — FR-1", files, tests, done) + task("T001 [feat] Again — FR-1", files, tests, done),
			[]string{"EVA087"},
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			p := newProject(t)
			p.doc("001-a", "spec.md", "approved", specBody)
			p.doc("001-a", "plan.md", "approved", planBody)
			p.doc("001-a", "tasks.md", "approved", c.body)
			expectCodes(t, run(t, p, Options{}), c.want...)
		})
	}
}

func TestTaskRuleSeverity(t *testing.T) {
	sixFiles := task("T001 [feat] Add — FR-1", "`a.ts`, `b.ts`, `c.ts`, `d.ts`, `e.ts`, `f.ts`", "`a.test.ts` — FR-1", "done")
	fourFiles := task("T001 [feat] Add — FR-1", "`a.ts`, `b.ts`, `c.ts`, `d.ts`", "`a.test.ts` — FR-1", "done")
	noType := task("T001 Add — FR-1", "`a.ts`", "`a.test.ts` — FR-1", "done")
	cases := []struct {
		status, body string
		want         map[string]Severity // nil: no findings
	}{
		{"approved", sixFiles, map[string]Severity{"EVA086": SevError}},
		{"approved", fourFiles, map[string]Severity{"EVA086": SevWarning}},
		{"in-review", sixFiles, map[string]Severity{"EVA086": SevWarning}},
		{"approved", noType, map[string]Severity{"EVA080": SevError}},
		{"in-review", noType, map[string]Severity{"EVA080": SevWarning}},
		{"draft", noType + sixFiles, nil},
	}
	for _, c := range cases {
		p := newProject(t)
		p.doc("001-a", "spec.md", "approved", specBody)
		p.doc("001-a", "plan.md", "approved", planBody)
		p.doc("001-a", "tasks.md", c.status, c.body)
		got := codes(run(t, p, Options{}))
		if len(c.want) == 0 && len(got) > 0 {
			t.Errorf("%s: want no findings, got %v", c.status, got)
		}
		for code, sev := range c.want {
			if got[code] != sev {
				t.Errorf("%s %s: severity = %q, want %q (all: %v)", c.status, code, got[code], sev, got)
			}
		}
	}
}

func TestPlanRules(t *testing.T) {
	cases := map[string]struct {
		status, body string
		want         []string
	}{
		"traces every FR":         {"approved", planBody, nil},
		"no traceability section": {"approved", "## 1. Approach\n\nFR-1 is handled by the form.\n", []string{"EVA088"}},
		"FR missing from traceability": {"approved",
			"## 2. Requirements traceability\n\n| FR | Plan |\n|----|----|\n| — | nothing yet |\n", []string{"EVA088"}},
		"cites an unknown FR":    {"approved", planBody + "\n## 3. Architecture\n\nAlso covers FR-4.\n", []string{"EVA081"}},
		"in-review is a warning": {"in-review", "## 1. Approach\n", []string{"EVA088"}},
		"draft is not checked":   {"draft", "## 1. Approach\n\nFR-9\n", nil},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			p := newProject(t)
			p.doc("001-a", "spec.md", "approved", specBody)
			p.doc("001-a", "plan.md", c.status, c.body)
			r := run(t, p, Options{})
			expectCodes(t, r, c.want...)
			if c.status == "in-review" && len(c.want) > 0 && codes(r)[c.want[0]] != SevWarning {
				t.Errorf("in-review findings should be warnings, got %+v", r.Findings)
			}
		})
	}
}

func TestVerificationGate(t *testing.T) {
	const table = "## 2. Requirements traceability\n\n| FR | Test(s) | Result |\n|----|----|----|\n"
	cases := map[string]struct {
		body string // "" removes verification.md
		want []string
	}{
		"passing verification": {verificationBody, nil},
		"missing verification": {"", []string{"EVA090"}},
		"FR not verified":      {table + "| — | none | ✅ |\n\n**Verdict:** PASS\n", []string{"EVA091"}},
		"FR failed":            {table + "| FR-1 | `a.test.ts` | ❌ |\n\n**Verdict:** PASS\n", []string{"EVA093"}},
		"verdict fail":         {table + "| FR-1 | `a.test.ts` | ✅ |\n\n**Verdict:** FAIL — perf budget\n", []string{"EVA092"}},
		"no verdict":           {table + "| FR-1 | `a.test.ts` | ✅ |\n", []string{"EVA092"}},
		"placeholder verdict":  {table + "| FR-1 | `a.test.ts` | ✅ |\n\n**Verdict:** PASS | FAIL\n", []string{"EVA092"}},
		"plain verdict line":   {table + "| FR-1 | `a.test.ts` | PASS |\n\nVerdict: PASS\n", nil},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			p := newProject(t)
			p.doc("001-a", "spec.md", "done", specBody)
			p.doc("001-a", "plan.md", "approved", planBody)
			p.doc("001-a", "tasks.md", "approved", tasks(true))
			if c.body == "" {
				p.remove("specs/001-a/verification.md")
			} else {
				p.verification("001-a", c.body)
			}
			expectCodes(t, run(t, p, Options{}), c.want...)
		})
	}
}

func TestVerificationNotCheckedBeforeDone(t *testing.T) {
	p := newProject(t)
	p.approvedFeature("001-a")
	p.verification("001-a", "**Verdict:** FAIL\n")
	r := run(t, p, Options{})
	for _, f := range r.Findings {
		if strings.HasPrefix(f.Code, "EVA09") {
			t.Errorf("verification content is only checked once the spec is done: %+v", f)
		}
	}
}
