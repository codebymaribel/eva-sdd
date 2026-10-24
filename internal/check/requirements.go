package check

import (
	"regexp"
	"strings"
)

// Requirement is one functional requirement (FR-n) defined in a spec.
type Requirement struct {
	ID   string // FR-1
	Text string // the EARS statement, without the id and leading tags
}

var (
	// A requirement is defined by a list item that starts with its id:
	//   - **FR-1** (US-1) When <trigger>, the <system> shall <response>.
	requirementLine = regexp.MustCompile(`(?m)^[ \t]*(?:[-*+]|\d+\.)[ \t]+(?:\*\*)?(FR-\d+)(?:\*\*)?[ \t]*:?(?:\*\*)?[ \t]*(.*)$`)
	// Leading tags such as (US-1), [US-2], [P1] and separators before the statement.
	leadingTags = regexp.MustCompile(`^(?:(?:\([^)]*\)|\[[^\]]*\])[ \t]*|[—–:-][ \t]*)+`)

	// EARS (Easy Approach to Requirements Syntax) patterns:
	//   ubiquitous  The <system> shall <response>.
	//   event       When <trigger>, the <system> shall <response>.
	//   state       While <state>, the <system> shall <response>.
	//   unwanted    If <condition>, then the <system> shall <response>.
	//   optional    Where <feature>, the <system> shall <response>.
	// Complex requirements combine the clauses (While …, when …, the … shall …).
	earsStart    = regexp.MustCompile(`(?i)^(the|when|while|if|where)\b`)
	earsShall    = regexp.MustCompile(`(?i)\bshall\s+\S`)
	earsIfThen   = regexp.MustCompile(`(?i),\s*then\b`)
	listItemLine = regexp.MustCompile(`(?m)^[ \t]*(?:[-*+]|\d+\.)[ \t]+\S`)
	htmlComment  = regexp.MustCompile(`(?s)<!--.*?-->`)
	headingLine  = regexp.MustCompile(`(?m)^(#{1,6})[ \t]+(.+?)[ \t]*#*[ \t]*$`)
	criteriaName = regexp.MustCompile(`(?i)^(?:\d+\.\s*)?completion criteria$`)
	frName       = regexp.MustCompile(`(?i)^(?:\d+\.\s*)?functional requirements\b.*$`)
)

// section returns the text of the first level-2 or level-3 heading whose title
// matches name, up to the next heading of the same or a higher level. Fenced
// code and HTML comments are removed first.
func section(body string, name *regexp.Regexp) (text string, found bool) {
	clean := htmlComment.ReplaceAllString(fencedCode.ReplaceAllString(body, ""), "")
	headings := headingLine.FindAllStringSubmatchIndex(clean, -1)
	for i, h := range headings {
		level := h[3] - h[2]
		if level < 2 || level > 3 || !name.MatchString(clean[h[4]:h[5]]) {
			continue
		}
		end := len(clean)
		for _, next := range headings[i+1:] {
			if next[3]-next[2] <= level {
				end = next[0]
				break
			}
		}
		return clean[h[1]:end], true
	}
	return "", false
}

// ParseRequirements returns the FR-n requirements defined in the spec's
// "Functional requirements" section, in order. FR ids mentioned elsewhere
// (completion criteria, open questions) are references, not definitions, and
// requirements quoted in code blocks or HTML comments are ignored.
func ParseRequirements(body string) []Requirement {
	text, _ := section(body, frName)
	var out []Requirement
	for _, m := range requirementLine.FindAllStringSubmatch(text, -1) {
		text := strings.TrimSpace(leadingTags.ReplaceAllString(strings.TrimSpace(m[2]), ""))
		out = append(out, Requirement{ID: m[1], Text: text})
	}
	return out
}

// IsEARS reports whether a requirement statement follows an EARS pattern:
// it opens with The/When/While/If/Where, contains "shall <response>", and an
// If clause is followed by ", then".
func IsEARS(text string) bool {
	text = strings.TrimSpace(text)
	m := earsStart.FindStringSubmatch(text)
	if m == nil || !earsShall.MatchString(text) {
		return false
	}
	if strings.EqualFold(m[1], "if") && !earsIfThen.MatchString(text) {
		return false
	}
	return true
}

// CompletionCriteria finds the "Completion criteria" section (a level-2 or
// level-3 heading, optionally numbered) and counts its list items, ignoring
// HTML comments and code blocks. The section ends at the next heading of the
// same or a higher level.
func CompletionCriteria(body string) (found bool, items int) {
	text, found := section(body, criteriaName)
	return found, len(listItemLine.FindAllString(text, -1))
}

// checkSpecContent applies the content rules of an approved (or done) spec:
// at least one requirement, unique ids, EARS statements and completion criteria.
func checkSpecContent(spec *Doc, r *Report) {
	reqs := ParseRequirements(spec.Body)
	if len(reqs) == 0 {
		r.add(SevError, "EVA026", spec.Path, "approved spec defines no functional requirements (list items starting with FR-n under a \"Functional requirements\" heading)")
	}
	seen := map[string]bool{}
	for _, req := range reqs {
		if seen[req.ID] {
			r.add(SevError, "EVA027", spec.Path, "%s is defined more than once; requirement ids must be unique", req.ID)
		}
		seen[req.ID] = true
		if !IsEARS(req.Text) {
			r.add(SevError, "EVA028", spec.Path, "%s is not in EARS form (The/When/While/If…then/Where …, the <system> shall <response>): %q", req.ID, req.Text)
		}
	}
	switch found, n := CompletionCriteria(spec.Body); {
	case !found:
		r.add(SevError, "EVA029", spec.Path, "approved spec has no \"Completion criteria\" section")
	case n == 0:
		r.add(SevError, "EVA029", spec.Path, "the \"Completion criteria\" section has no list items")
	}
}
