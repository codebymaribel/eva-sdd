package check

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseClarifications(t *testing.T) {
	body := "## Findings\n\n" +
		"<!-- - [ ] **C-9** [ambiguity] commented-out example -->\n" +
		"- [ ] **C-1** [ambiguity] (FR-2) What does \"recent\" mean?\n" +
		"  - Answer:\n" +
		"- [x] **C-2** [edge-case] (§5) What happens offline?\n" +
		"  - Answer: Queue the change and retry.\n" +
		"  - Applied: FR-4 added.\n" +
		"* [X] C-3: [constitution] Conflicts with principle 7?\n" +
		"  - **Answer:** No, keep it.\n" +
		"```\n- [ ] **C-8** [ambiguity] inside code\n```\n" +
		"## Notes\n\n- Answer: not part of C-3.\n"
	got := ParseClarifications(body)
	want := []Clarification{
		{ID: "C-1", Category: "ambiguity", Resolved: false, Answer: ""},
		{ID: "C-2", Category: "edge-case", Resolved: true, Answer: "Queue the change and retry."},
		{ID: "C-3", Category: "constitution", Resolved: true, Answer: "No, keep it."},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseClarifications:\n got %+v\nwant %+v", got, want)
	}
}

func TestClarifyRules(t *testing.T) {
	const (
		open     = "- [ ] **C-1** [ambiguity] (FR-1) Which form?\n  - Answer:\n"
		answered = "- [x] **C-1** [ambiguity] (FR-1) Which form?\n  - Answer: The sign-up form.\n"
	)
	cases := map[string]struct {
		status string // spec status
		setup  func(p *project)
		want   []string
	}{
		"approved spec with resolved findings": {"approved", func(p *project) { p.clarify("001-a", answered) }, nil},
		"approved spec without clarify.md": {"approved", func(p *project) {
			if err := os.Remove(filepath.Join(p.root, "specs/001-a/clarify.md")); err != nil {
				p.t.Fatal(err)
			}
		}, []string{"EVA070"}},
		"approved spec with open findings":  {"approved", func(p *project) { p.clarify("001-a", open) }, []string{"EVA071"}},
		"in-review spec with open findings": {"in-review", func(p *project) { p.clarify("001-a", open) }, nil},
		"in-review spec without clarify.md": {"in-review", func(p *project) {}, nil},
		"duplicate finding id": {"in-review", func(p *project) {
			p.clarify("001-a", open+"- [ ] **C-1** [edge-case] Offline?\n")
		}, []string{"EVA072"}},
		"unknown category": {"in-review", func(p *project) {
			p.clarify("001-a", "- [ ] **C-1** [typo] Wording?\n")
		}, []string{"EVA073"}},
		"missing category": {"in-review", func(p *project) {
			p.clarify("001-a", "- [ ] **C-1** Wording?\n")
		}, []string{"EVA073"}},
		"resolved without answer": {"in-review", func(p *project) {
			p.clarify("001-a", "- [x] **C-1** [ambiguity] Which form?\n  - Answer:\n")
		}, []string{"EVA074"}},
		"no front-matter": {"in-review", func(p *project) {
			p.write("specs/001-a/clarify.md", "# Clarify\n")
		}, []string{"EVA021"}},
		"wrong id": {"in-review", func(p *project) {
			p.write("specs/001-a/clarify.md", front("id", "002-b", "type", "clarify", "created", testDate))
		}, []string{"EVA022"}},
		"wrong type": {"in-review", func(p *project) {
			p.write("specs/001-a/clarify.md", front("id", "001-a", "type", "spec", "created", testDate))
		}, []string{"EVA060"}},
		"without created": {"in-review", func(p *project) {
			p.write("specs/001-a/clarify.md", front("id", "001-a", "type", "clarify"))
		}, []string{"EVA062"}},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			p := newProject(t)
			p.doc("001-a", "spec.md", c.status, specBody)
			c.setup(p)
			expectCodes(t, run(t, p, Options{}), c.want...)
		})
	}
}

func TestOpenFindingsBlockPlan(t *testing.T) {
	p := newProject(t)
	p.doc("001-a", "spec.md", "approved", specBody)
	p.doc("001-a", "plan.md", "draft", "")
	p.clarify("001-a", "- [ ] **C-1** [contradiction] FR-1 vs FR-2?\n")
	r := run(t, p, Options{})
	expectCodes(t, r, "EVA071")
	for _, f := range r.Findings {
		if f.Code == "EVA071" && f.Path != "specs/001-a/clarify.md" {
			t.Errorf("EVA071 path = %q, want the clarify file", f.Path)
		}
	}
}
