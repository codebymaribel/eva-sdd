package check

import "testing"

func TestParseDate(t *testing.T) {
	for in, ok := range map[string]bool{
		"2026-10-23":           true,
		"2024-02-29":           true, // leap day
		"2026-02-30":           false,
		"23/10/2026":           false,
		"2026-1-5":             false,
		"2026-10-23T10:00:00Z": false,
		"YYYY-MM-DD":           false,
	} {
		if _, err := parseDate(in); (err == nil) != ok {
			t.Errorf("parseDate(%q) err = %v, want ok=%v", in, err, ok)
		}
	}
}

func TestDocumentMetadataRules(t *testing.T) {
	const spec = "specs/001-a/spec.md"
	cases := map[string]struct {
		setup func(p *project)
		want  []string
	}{
		"missing type": {
			func(p *project) { p.write(spec, front("id", "001-a", "status", "draft")) },
			[]string{"EVA060"},
		},
		"type of another document": {
			func(p *project) {
				p.doc("001-a", "spec.md", "approved", specBody)
				p.write("specs/001-a/plan.md", front("id", "001-a", "type", "spec", "status", "draft", "created", testDate))
			},
			[]string{"EVA060"},
		},
		"created is not YYYY-MM-DD": {
			func(p *project) {
				p.write(spec, front("id", "001-a", "type", "spec", "status", "draft", "created", "23/10/2026"))
			},
			[]string{"EVA061"},
		},
		"created is not a calendar date": {
			func(p *project) {
				p.write(spec, front("id", "001-a", "type", "spec", "status", "draft", "created", "2026-02-30"))
			},
			[]string{"EVA061"},
		},
		"draft without created is allowed": {
			func(p *project) {
				p.write(spec, front("id", "001-a", "type", "spec", "status", "draft", "created", ""))
			},
			nil,
		},
		"in-review without created": {
			func(p *project) { p.write(spec, front("id", "001-a", "type", "spec", "status", "in-review")) },
			[]string{"EVA062"},
		},
		"approved without approved_at": {
			func(p *project) {
				p.write(spec, front("id", "001-a", "type", "spec", "status", "approved", "created", testDate, "approved_by", "Maribel")+specBody)
			},
			[]string{"EVA063"},
		},
		"approved_at on an unapproved document": {
			func(p *project) {
				p.write(spec, front("id", "001-a", "type", "spec", "status", "in-review", "created", testDate, "approved_at", testDate))
			},
			[]string{"EVA063"},
		},
		"done without closed_at": {
			func(p *project) {
				p.write(spec, front("id", "001-a", "type", "spec", "status", "done", "created", testDate,
					"approved_by", "Maribel", "approved_at", testDate)+specBody)
				p.doc("001-a", "plan.md", "approved", planBody)
				p.doc("001-a", "tasks.md", "approved", tasks(true))
				p.verification("001-a", verificationBody)
			},
			[]string{"EVA064"},
		},
		"closed_at before done": {
			func(p *project) {
				p.write(spec, front("id", "001-a", "type", "spec", "status", "approved", "created", testDate,
					"approved_by", "Maribel", "approved_at", testDate, "closed_at", testDate)+specBody)
			},
			[]string{"EVA064"},
		},
		"approved before created": {
			func(p *project) {
				p.write(spec, front("id", "001-a", "type", "spec", "status", "approved", "created", "2026-03-10",
					"approved_by", "Maribel", "approved_at", "2026-03-09")+specBody)
			},
			[]string{"EVA065"},
		},
		"closed before approved": {
			func(p *project) {
				p.write(spec, front("id", "001-a", "type", "spec", "status", "done", "created", "2026-03-01",
					"approved_by", "Maribel", "approved_at", "2026-03-10", "closed_at", "2026-03-05")+specBody)
				p.doc("001-a", "plan.md", "approved", planBody)
				p.doc("001-a", "tasks.md", "approved", tasks(true))
				p.verification("001-a", verificationBody)
			},
			[]string{"EVA065"},
		},
		"same day is in order": {
			func(p *project) {
				p.doc("001-a", "spec.md", "done", specBody)
				p.doc("001-a", "plan.md", "approved", planBody)
				p.doc("001-a", "tasks.md", "approved", tasks(true))
				p.verification("001-a", verificationBody)
			},
			nil,
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			p := newProject(t)
			p.clarify("001-a", "") // approved specs need a clarify.md
			c.setup(p)
			expectCodes(t, run(t, p, Options{}), c.want...)
		})
	}
}

func TestVerificationMetadataRules(t *testing.T) {
	const path = "specs/001-a/verification.md"
	cases := map[string]struct {
		fm   string
		want []string
	}{
		"valid open verification": {
			front("id", "001-a", "type", "verification", "created", testDate, "closed_at", ""),
			nil,
		},
		"valid closed verification": {
			front("id", "001-a", "type", "verification", "created", "2026-03-01", "closed_at", "2026-03-02"),
			nil,
		},
		"no front-matter": {"# Verification\n", []string{"EVA021"}},
		"wrong id": {
			front("id", "002-b", "type", "verification", "created", testDate),
			[]string{"EVA022"},
		},
		"wrong type": {
			front("id", "001-a", "type", "spec", "created", testDate),
			[]string{"EVA060"},
		},
		"without created": {
			front("id", "001-a", "type", "verification"),
			[]string{"EVA062"},
		},
		"bad closed_at": {
			front("id", "001-a", "type", "verification", "created", testDate, "closed_at", "soon"),
			[]string{"EVA061"},
		},
		"closed before created": {
			front("id", "001-a", "type", "verification", "created", "2026-03-02", "closed_at", "2026-03-01"),
			[]string{"EVA065"},
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			p := newProject(t)
			p.approvedFeature("001-a")
			p.write(path, c.fm)
			expectCodes(t, run(t, p, Options{}), c.want...)
		})
	}
}

func TestConstitutionMetadataRules(t *testing.T) {
	cases := map[string]struct {
		fm   string
		want []string
	}{
		"draft from eva init has no dates yet": {
			front("type", "constitution", "status", "draft", "version", "1", "created", "", "approved_by", "", "approved_at", "", "closed_at", ""),
			[]string{"EVA053"},
		},
		"wrong type": {
			front("type", "spec", "status", "approved", "created", testDate, "approved_by", "Maribel", "approved_at", testDate),
			[]string{"EVA060"},
		},
		"approved without approved_at": {
			front("type", "constitution", "status", "approved", "created", testDate, "approved_by", "Maribel"),
			[]string{"EVA063"},
		},
		"closed_at is never set": {
			front("type", "constitution", "status", "approved", "created", testDate, "approved_by", "Maribel", "approved_at", testDate, "closed_at", testDate),
			[]string{"EVA064"},
		},
		"bad date": {
			front("type", "constitution", "status", "in-review", "created", "2026/01/01"),
			[]string{"EVA053", "EVA061"},
		},
		"approved before created": {
			front("type", "constitution", "status", "approved", "created", "2026-05-02", "approved_by", "Maribel", "approved_at", "2026-05-01"),
			[]string{"EVA065"},
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			p := newProject(t)
			p.write(ConstitutionPath, c.fm)
			expectCodes(t, run(t, p, Options{}), c.want...)
		})
	}
}
