package check

import (
	"path"
	"regexp"
	"strings"
)

// Task is one checklist item of tasks.md with its sub-fields.
type Task struct {
	ID        string   // T001
	Type      string   // feat | fix | improve, "" when missing or unknown
	Done      bool     // the checkbox is ticked
	FRs       []string // requirements cited in the title and the "Tests:" line, in order, unique
	Files     []string // backticked paths of the "Files:" line
	HasTests  bool     // a "Tests:" line exists
	TestFiles []string // backticked test paths of the "Tests:" line
	TestFRs   []string // requirements cited in the "Tests:" line
	DoneWhen  string   // text of the "Done when:" line
}

// Code files allowed per task: more than soft is a warning, more than hard an error.
const (
	maxCodeFilesSoft = 3
	maxCodeFilesHard = 5
)

var (
	taskTypes   = map[string]bool{"feat": true, "fix": true, "improve": true}
	taskItem    = regexp.MustCompile(`(?m)^[ \t]*[-*][ \t]+\[( |x|X)\][ \t]+(T\d{3,})\b(.*)$`)
	taskTag     = regexp.MustCompile(`\[([^\]]+)\]`)
	taskField   = regexp.MustCompile(`(?mi)^[ \t]*[-*][ \t]+(files|tests|done when):[ \t]*(.*)$`)
	backticked  = regexp.MustCompile("`([^`]+)`")
	frRef       = regexp.MustCompile(`\bFR-\d+\b`)
	traceName   = regexp.MustCompile(`(?i)^(?:\d+\.\s*)?requirements traceability$`)
	verdictLine = regexp.MustCompile(`(?mi)^[ \t]*(?:[-*][ \t]+)?(?:\*\*)?verdict:?(?:\*\*)?:?[ \t]*(.*)$`)
	failedRow   = regexp.MustCompile(`❌|\bFAIL(?:ED)?\b`)
	testPath    = regexp.MustCompile(`(?i)(\.(test|spec|e2e)\.[a-z0-9]+$|_test\.[a-z0-9]+$|(^|/)(__tests__|e2e)/)`)
)

// IsTestFile reports whether a path is a test by name or location
// (*.test.*, *.spec.*, *.e2e.*, *_test.*, __tests__/, e2e/).
func IsTestFile(p string) bool {
	return testPath.MatchString(p)
}

// isCodeFile reports whether a task file counts towards the code-file limit:
// anything that is not a test, a Markdown document or a snapshot.
func isCodeFile(p string) bool {
	switch strings.ToLower(path.Ext(p)) {
	case ".md", ".mdx", ".snap":
		return false
	}
	return !IsTestFile(p)
}

// ParseTasks returns the T-numbered tasks of a tasks.md body. A task's block
// runs until the next task or heading; checklist items in code blocks or HTML
// comments are ignored.
func ParseTasks(body string) []Task {
	clean := htmlComment.ReplaceAllString(fencedCode.ReplaceAllString(body, ""), "")
	locs := taskItem.FindAllStringSubmatchIndex(clean, -1)
	var out []Task
	for i, l := range locs {
		end := len(clean)
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		block := clean[l[1]:end]
		if h := headingLine.FindStringIndex(block); h != nil {
			block = block[:h[0]]
		}
		head := clean[l[6]:l[7]]
		t := Task{ID: clean[l[4]:l[5]], Done: clean[l[2]:l[3]] != " "}
		for _, m := range taskTag.FindAllStringSubmatch(head, -1) {
			if tag := strings.ToLower(strings.TrimSpace(m[1])); taskTypes[tag] && t.Type == "" {
				t.Type = tag
			}
		}
		for _, f := range taskField.FindAllStringSubmatch(block, -1) {
			value := strings.TrimSpace(f[2])
			switch strings.ToLower(f[1]) {
			case "files":
				t.Files = backtickedPaths(value)
			case "tests":
				t.HasTests = true
				for _, p := range backtickedPaths(value) {
					if IsTestFile(p) {
						t.TestFiles = append(t.TestFiles, p)
					}
				}
				t.TestFRs = uniqueFRs(value)
			case "done when":
				t.DoneWhen = value
			}
		}
		t.FRs = uniqueFRs(head + " " + strings.Join(t.TestFRs, " "))
		out = append(out, t)
	}
	return out
}

func backtickedPaths(s string) []string {
	var out []string
	for _, m := range backticked.FindAllStringSubmatch(s, -1) {
		out = append(out, strings.TrimSpace(m[1]))
	}
	return out
}

// uniqueFRs returns the FR ids cited in s, in order of first appearance.
func uniqueFRs(s string) []string {
	var out []string
	seen := map[string]bool{}
	for _, id := range frRef.FindAllString(s, -1) {
		if !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	return out
}

// contentSeverity is the severity of content rules for a document: errors
// once it is approved (or done), warnings while in review so the agent sees
// them before asking for review, nothing for drafts.
func contentSeverity(d *Doc) (Severity, bool) {
	switch d.Status() {
	case "approved", "done":
		return SevError, true
	case "in-review":
		return SevWarning, true
	}
	return "", false
}

// specRequirements returns the FR ids defined by the spec.
func specRequirements(spec *Doc) map[string]bool {
	ids := map[string]bool{}
	if spec == nil || spec.ParseErr != nil {
		return ids
	}
	for _, r := range ParseRequirements(spec.Body) {
		ids[r.ID] = true
	}
	return ids
}

// checkPlanTraceability requires the plan to cite only existing requirements
// and to trace every requirement in its "Requirements traceability" section.
func checkPlanTraceability(f Feature, r *Report) {
	plan := f.Plan
	sev, ok := contentSeverity(plan)
	if !ok || plan.ParseErr != nil {
		return
	}
	reqs := specRequirements(f.Spec)
	for _, id := range uniqueFRs(stripCode(plan.Body)) {
		if !reqs[id] {
			r.add(sev, "EVA081", plan.Path, "%s is cited but the spec does not define it", id)
		}
	}
	section, found := section(plan.Body, traceName)
	if !found {
		r.add(sev, "EVA088", plan.Path, "no \"Requirements traceability\" section mapping each FR to the plan")
		return
	}
	traced := map[string]bool{}
	for _, id := range uniqueFRs(section) {
		traced[id] = true
	}
	for _, id := range sortedIDs(reqs) {
		if !traced[id] {
			r.add(sev, "EVA088", plan.Path, "%s is not in the \"Requirements traceability\" section", id)
		}
	}
}

// checkTasks applies the task rules: a type, existing requirements, every
// requirement covered, Files / Tests / Done when, size and unique ids.
func checkTasks(f Feature, r *Report) {
	doc := f.Tasks
	sev, ok := contentSeverity(doc)
	if !ok || doc.ParseErr != nil {
		return
	}
	reqs := specRequirements(f.Spec)
	covered := map[string]bool{}
	seen := map[string]bool{}
	for _, t := range ParseTasks(doc.Body) {
		if seen[t.ID] {
			r.add(sev, "EVA087", doc.Path, "%s is used by more than one task", t.ID)
		}
		seen[t.ID] = true
		if t.Type == "" {
			r.add(sev, "EVA080", doc.Path, "%s has no type: tag it [feat], [fix] or [improve]", t.ID)
		}
		for _, id := range t.FRs {
			covered[id] = true
			if !reqs[id] {
				r.add(sev, "EVA081", doc.Path, "%s cites %s but the spec does not define it", t.ID, id)
			}
		}
		behaviour := t.Type == "feat" || t.Type == "fix"
		if behaviour && len(t.FRs) == 0 {
			r.add(sev, "EVA083", doc.Path, "%s is [%s] but cites no requirement (FR-n)", t.ID, t.Type)
		}
		if len(t.Files) == 0 {
			r.add(sev, "EVA084", doc.Path, "%s has no \"Files:\" line with backticked paths", t.ID)
		}
		if t.DoneWhen == "" {
			r.add(sev, "EVA084", doc.Path, "%s has no \"Done when:\" condition", t.ID)
		}
		if behaviour {
			switch {
			case !t.HasTests || len(t.TestFiles) == 0:
				r.add(sev, "EVA085", doc.Path, "%s needs a \"Tests:\" line with the test file(s) written first (e.g. `src/x.test.ts`)", t.ID)
			case len(t.FRs) > 0 && len(t.TestFRs) == 0:
				r.add(sev, "EVA085", doc.Path, "%s: the \"Tests:\" line must say which requirement each test verifies (FR-n)", t.ID)
			}
		}
		code := 0
		for _, p := range t.Files {
			if isCodeFile(p) {
				code++
			}
		}
		switch {
		case code > maxCodeFilesHard:
			r.add(sev, "EVA086", doc.Path, "%s touches %d code files (max %d): split it into smaller tasks", t.ID, code, maxCodeFilesHard)
		case code > maxCodeFilesSoft:
			r.add(SevWarning, "EVA086", doc.Path, "%s touches %d code files; aim for %d or fewer", t.ID, code, maxCodeFilesSoft)
		}
	}
	for _, id := range sortedIDs(reqs) {
		if !covered[id] {
			r.add(sev, "EVA082", doc.Path, "%s has no task", id)
		}
	}
}

// checkVerification is the closing gate: a done spec needs a verification
// that checks every requirement and ends with "Verdict: PASS".
func checkVerification(f Feature, r *Report) {
	if f.Spec == nil || f.Spec.Status() != "done" {
		return
	}
	v := f.Verification
	if v == nil {
		r.add(SevError, "EVA090", f.Spec.Path, "spec is done but verification.md is missing: run /eva-verify")
		return
	}
	if v.ParseErr != nil {
		return // reported as EVA021
	}
	clean := htmlComment.ReplaceAllString(fencedCode.ReplaceAllString(v.Body, ""), "")
	table, _ := section(v.Body, traceName)
	rows := map[string]string{}
	for _, line := range strings.Split(table, "\n") {
		for _, id := range uniqueFRs(line) {
			rows[id] += line
		}
	}
	for _, id := range sortedIDs(specRequirements(f.Spec)) {
		row, ok := rows[id]
		switch {
		case !ok:
			r.add(SevError, "EVA091", v.Path, "%s is not verified in the \"Requirements traceability\" table", id)
		case failedRow.MatchString(row):
			r.add(SevError, "EVA093", v.Path, "%s failed verification but the spec is done", id)
		}
	}
	m := verdictLine.FindStringSubmatch(clean)
	verdict := ""
	if m != nil {
		verdict = strings.TrimSpace(strings.Trim(strings.TrimSpace(m[1]), "*"))
	}
	pass := strings.HasPrefix(verdict, "PASS") && !strings.HasPrefix(strings.TrimSpace(strings.TrimPrefix(verdict, "PASS")), "|")
	if !pass {
		if verdict == "" {
			verdict = "missing"
		}
		r.add(SevError, "EVA092", v.Path, "spec is done but the verdict is not PASS (%s)", verdict)
	}
}

// sortedIDs returns FR ids in numeric order (FR-2 before FR-10).
func sortedIDs(ids map[string]bool) []string {
	out := make([]string, 0, len(ids))
	for id := range ids {
		out = append(out, id)
	}
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && frLess(out[j], out[j-1]); j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

func frLess(a, b string) bool {
	if len(a) != len(b) {
		return len(a) < len(b)
	}
	return a < b
}
