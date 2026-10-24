package check

import (
	"reflect"
	"testing"
)

func TestParseRequirements(t *testing.T) {
	body := "## 3. User stories\n\n- FR-7 is only mentioned here.\n\n" +
		"## 4. Functional requirements\n\nIntro mentions FR-8 in prose.\n\n" +
		"### US-1\n\n" +
		"- **FR-1** (US-1) When the user taps Save, the app shall store the draft.\n" +
		"- **FR-2:** The app shall keep drafts for 30 days.\n" +
		"* **FR-3** [US-2] [P2] — If the network fails, then the app shall retry once.\n" +
		"  - **FR-4**   While offline, the app shall show a banner.\n" +
		"```\n- **FR-9** inside a code block\n```\n" +
		"<!-- - **FR-10** The app shall be commented out. -->\n" +
		"## 7. Completion criteria\n\n- FR-1 has a passing test.\n"
	got := ParseRequirements(body)
	want := []Requirement{
		{ID: "FR-1", Text: "When the user taps Save, the app shall store the draft."},
		{ID: "FR-2", Text: "The app shall keep drafts for 30 days."},
		{ID: "FR-3", Text: "If the network fails, then the app shall retry once."},
		{ID: "FR-4", Text: "While offline, the app shall show a banner."},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseRequirements:\n got %+v\nwant %+v", got, want)
	}
}

func TestIsEARS(t *testing.T) {
	for text, want := range map[string]bool{
		"The app shall encrypt tokens at rest.":                                    true,  // ubiquitous
		"When the user taps Sign in, the app shall validate both fields.":          true,  // event-driven
		"While the device is offline, the app shall queue changes.":                true,  // state-driven
		"If the password is wrong, then the app shall show \"Wrong password\".":    true,  // unwanted behaviour
		"Where biometrics are available, the app shall offer Face ID.":             true,  // optional feature
		"While offline, when the user taps Sync, the app shall queue the request.": true,  // complex
		"when the user taps Save, the app shall store the draft.":                  true,  // case-insensitive keyword
		"If the password is wrong, the app shall show an error.":                   false, // If without then
		"The app should be fast.":                                                  false, // no shall
		"Show an error when the password is empty.":                                false, // no EARS keyword
		"Whenever the user logs in, the app shall greet them.":                     false, // not an EARS keyword
		"The app shall": false, // no response
	} {
		if got := IsEARS(text); got != want {
			t.Errorf("IsEARS(%q) = %v, want %v", text, got, want)
		}
	}
}

func TestCompletionCriteria(t *testing.T) {
	cases := map[string]struct {
		body         string
		found        bool
		wantCriteria int
	}{
		"missing":          {"## 1. Summary\n\nText\n", false, 0},
		"only a comment":   {"## 7. Completion criteria\n\n<!-- list them -->\n\n## 8. Out of scope\n- x\n", true, 0},
		"only prose":       {"## Completion criteria\n\nSee the requirements.\n", true, 0},
		"two items":        {"## 7. Completion criteria\n\n- FR-1 and FR-2 pass.\n- [ ] Works offline.\n\n## 8. Out of scope\n- x\n", true, 2},
		"items in code":    {"## Completion criteria\n```\n- not an item\n```\n", true, 0},
		"subsection items": {"## Completion criteria\n### Mobile\n1. Android 10+ supported.\n", true, 1},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			found, n := CompletionCriteria(c.body)
			if found != c.found || n != c.wantCriteria {
				t.Errorf("got (%v, %d), want (%v, %d)", found, n, c.found, c.wantCriteria)
			}
		})
	}
}

func TestSpecContentRules(t *testing.T) {
	const (
		reqs     = "## 4. Functional requirements\n\n"
		criteria = "## 7. Completion criteria\n\n- Done.\n"
	)
	cases := map[string]struct {
		status string
		body   string
		want   []string
	}{
		"approved spec without requirements": {"approved", criteria, []string{"EVA026"}},
		"requirements outside their section": {"approved",
			"## 3. User stories\n\n- **FR-1** The app shall a.\n" + criteria, []string{"EVA026"}},
		"duplicate requirement ids": {"approved",
			reqs + "- **FR-1** The app shall a.\n- **FR-1** The app shall b.\n" + criteria, []string{"EVA027"}},
		"requirement not in EARS": {"approved",
			reqs + "- **FR-1** The app shall a.\n- **FR-2** Show an error.\n" + criteria, []string{"EVA028"}},
		"missing completion criteria": {"approved", reqs + "- **FR-1** The app shall a.\n", []string{"EVA029"}},
		"empty completion criteria": {"approved",
			reqs + "- **FR-1** The app shall a.\n## 7. Completion criteria\n\n<!-- TBD -->\n", []string{"EVA029"}},
		"draft specs are not checked yet":     {"draft", "Rough notes.\n", nil},
		"in-review specs are not checked yet": {"in-review", "- **FR-1** Show an error.\n", nil},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			p := newProject(t)
			p.doc("001-a", "spec.md", c.status, c.body)
			expectCodes(t, run(t, p, Options{}), c.want...)
		})
	}
}
