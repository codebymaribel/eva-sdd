package check

import (
	"errors"
	"testing"
)

func TestParseFrontMatter(t *testing.T) {
	doc := "---\r\nid: 001-login\r\nstatus: approved # only humans\r\ntitle: \"Login: email\"\r\napproved_by:\r\nissue: # optional\r\n# comment line\r\n---\r\n# Body\r\n"
	fields, body, err := ParseFrontMatter([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"id": "001-login", "status": "approved", "title": "Login: email", "approved_by": "", "issue": ""}
	for k, v := range want {
		if fields[k] != v {
			t.Errorf("%s = %q, want %q", k, fields[k], v)
		}
	}
	if body != "# Body\n" {
		t.Errorf("body = %q", body)
	}
}

func TestParseFrontMatterErrors(t *testing.T) {
	if _, _, err := ParseFrontMatter([]byte("# no front matter")); !errors.Is(err, ErrNoFrontMatter) {
		t.Errorf("err = %v, want ErrNoFrontMatter", err)
	}
	if _, _, err := ParseFrontMatter([]byte("---\nid: x\n")); err == nil {
		t.Error("expected error for unclosed front-matter")
	}
	if _, _, err := ParseFrontMatter([]byte("---\nnot a field\n---\n")); err == nil {
		t.Error("expected error for line without colon")
	}
}

func TestCountClarificationsIgnoresCode(t *testing.T) {
	body := "Use `[NEEDS CLARIFICATION: question]` for doubts.\n" +
		"```md\n[NEEDS CLARIFICATION: example]\n```\n" +
		"Users can log in with [NEEDS CLARIFICATION: SSO?] and [NEEDS CLARIFICATION: 2FA?].\n"
	if got := CountClarifications(body); got != 2 {
		t.Errorf("got %d, want 2", got)
	}
}

func TestCountTasks(t *testing.T) {
	body := "- [x] T001 a\n- [ ] T002 b\n  - [X] nested\n* [ ] T003\n- [] not a task\ntext - [x] inline\n"
	done, total := CountTasks(body)
	if done != 2 || total != 4 {
		t.Errorf("done=%d total=%d, want 2/4", done, total)
	}
}
