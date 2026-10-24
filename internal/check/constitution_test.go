package check

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConstitutionRules(t *testing.T) {
	cases := map[string]struct {
		setup func(p *project)
		want  []string
	}{
		"approved and complete": {
			setup: func(p *project) {},
		},
		"missing file": {
			setup: func(p *project) {
				if err := os.Remove(filepath.Join(p.root, ConstitutionPath)); err != nil {
					p.t.Fatal(err)
				}
			},
			want: []string{"EVA050"},
		},
		"no front-matter": {
			setup: func(p *project) { p.write(ConstitutionPath, "# Constitution\n") },
			want:  []string{"EVA051"},
		},
		"invalid status": {
			setup: func(p *project) { p.constitution("done", "") },
			want:  []string{"EVA051"},
		},
		"approved without approver": {
			setup: func(p *project) {
				p.write(ConstitutionPath, front("type", "constitution", "status", "approved", "created", testDate, "approved_by", "", "approved_at", testDate))
			},
			want: []string{"EVA052"},
		},
		"approved with TODOs left": {
			setup: func(p *project) {
				p.constitution("approved", "Replace every `TODO` below.\n- Users: TODO\n- Coverage: TODO: adjust\n")
			},
			want: []string{"EVA054"},
		},
		"draft without specs is a warning": {
			setup: func(p *project) { p.constitution("draft", "") },
			want:  []string{"EVA053"},
		},
		"in-review with specs is an error": {
			setup: func(p *project) {
				p.constitution("in-review", "")
				p.doc("001-login", "spec.md", "draft", "")
			},
			want: []string{"EVA053"},
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

func TestConstitutionPendingSeverity(t *testing.T) {
	p := newProject(t)
	p.constitution("draft", "")
	if sev := codes(run(t, p, Options{}))["EVA053"]; sev != SevWarning {
		t.Errorf("no specs: severity = %s, want warning", sev)
	}
	p.doc("001-login", "spec.md", "draft", "")
	if sev := codes(run(t, p, Options{}))["EVA053"]; sev != SevError {
		t.Errorf("with specs: severity = %s, want error", sev)
	}
}

func TestGateRequiresApprovedConstitution(t *testing.T) {
	p := newProject(t)
	p.approvedFeature("004-login")
	p.constitution("in-review", "")
	r := run(t, p, Options{Git: repo("feat/004-login", "src/login.ts")})

	var gate *Finding
	for i := range r.Findings {
		if r.Findings[i].Code == "EVA042" {
			gate = &r.Findings[i]
		}
	}
	if gate == nil || !strings.Contains(gate.Message, "constitution (in-review)") {
		t.Fatalf("EVA042 should list the constitution as pending, got %+v", r.Findings)
	}
}

func TestCountTODOsIgnoresCode(t *testing.T) {
	body := "Replace `TODO` markers.\n```\nTODO in code\n```\nUsers: TODO\nTODOS is not a marker\n"
	if got := CountTODOs(body); got != 1 {
		t.Errorf("got %d, want 1", got)
	}
}
