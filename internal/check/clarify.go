package check

import (
	"regexp"
	"strings"
)

// Clarification is one finding of a feature's clarify.md.
type Clarification struct {
	ID       string // C-1
	Category string // ambiguity | contradiction | edge-case | constitution
	Resolved bool   // the checkbox is ticked
	Answer   string // the developer's answer, "" when not answered yet
}

// clarifyCategories are the kinds of findings /eva-clarify may report.
var clarifyCategories = map[string]bool{"ambiguity": true, "contradiction": true, "edge-case": true, "constitution": true}

var (
	// - [ ] **C-1** [ambiguity] (FR-2) Question…
	clarifyItem = regexp.MustCompile(`(?m)^[ \t]*[-*][ \t]+\[( |x|X)\][ \t]+(?:\*\*)?(C-\d+)(?:\*\*)?[ \t]*:?(?:\*\*)?[ \t]*(.*)$`)
	categoryTag = regexp.MustCompile(`^\[([^\]]*)\]`)
	// "  - Answer: …" or "  - **Answer:** …"
	answerLine = regexp.MustCompile(`(?mi)^[ \t]*(?:[-*][ \t]+)?(?:\*\*)?answer:(?:\*\*)?[ \t]*(.*)$`)
)

// ParseClarifications returns the C-n findings of a clarify.md body, in order.
// The lines below a finding, up to the next finding or heading, belong to it;
// its answer is the first "Answer:" line among them. Findings in code blocks
// or HTML comments are ignored.
func ParseClarifications(body string) []Clarification {
	clean := htmlComment.ReplaceAllString(fencedCode.ReplaceAllString(body, ""), "")
	locs := clarifyItem.FindAllStringSubmatchIndex(clean, -1)
	out := make([]Clarification, 0, len(locs))
	for i, l := range locs {
		end := len(clean)
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		block := clean[l[1]:end]
		if h := headingLine.FindStringIndex(block); h != nil {
			block = block[:h[0]]
		}
		c := Clarification{ID: clean[l[4]:l[5]], Resolved: clean[l[2]:l[3]] != " "}
		if m := categoryTag.FindStringSubmatch(strings.TrimSpace(clean[l[6]:l[7]])); m != nil {
			c.Category = strings.ToLower(strings.TrimSpace(m[1]))
		}
		if m := answerLine.FindStringSubmatch(block); m != nil {
			c.Answer = strings.TrimSpace(m[1])
		}
		out = append(out, c)
	}
	return out
}

// validateClarify checks clarify.md and returns how many findings are still
// open. Like verification.md it has no status: the spec's approval closes it.
func validateClarify(f Feature, d *Doc, r *Report) (open int) {
	if d.ParseErr != nil {
		r.add(SevError, "EVA021", d.Path, "invalid front-matter: %v", d.ParseErr)
		return 0
	}
	if id := d.Fields["id"]; id != f.ID {
		r.add(SevError, "EVA022", d.Path, "front-matter id is %q, expected %q (the folder name)", id, f.ID)
	}
	checkMeta(d, "clarify", lifecycleRules{hasStatus: false, canClose: true}, r)

	seen := map[string]bool{}
	for _, c := range ParseClarifications(d.Body) {
		if seen[c.ID] {
			r.add(SevError, "EVA072", d.Path, "%s is used by more than one finding; finding ids must be unique", c.ID)
		}
		seen[c.ID] = true
		if !clarifyCategories[c.Category] {
			r.add(SevError, "EVA073", d.Path, "%s has category %q, want [ambiguity], [contradiction], [edge-case] or [constitution]", c.ID, c.Category)
		}
		if c.Resolved && c.Answer == "" {
			r.add(SevError, "EVA074", d.Path, "%s is ticked but has no answer; the developer answers before a finding is resolved", c.ID)
		}
		if !c.Resolved {
			open++
		}
	}
	return open
}

// checkClarifyGate requires an approved spec to have a clarify.md without open findings.
func checkClarifyGate(f Feature, open int, r *Report) {
	if !f.Spec.Approved() {
		return
	}
	switch {
	case f.Clarify == nil:
		r.add(SevError, "EVA070", f.Spec.Path, "spec is %s but clarify.md is missing: run /eva-clarify before approving", f.Spec.Status())
	case open > 0:
		r.add(SevError, "EVA071", f.Clarify.Path, "%d open finding(s) but spec.md is %s: answer them and apply them with /eva-specify", open, f.Spec.Status())
	}
}
