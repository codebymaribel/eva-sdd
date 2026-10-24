package prompt

import (
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
)

var opts = []Option{{"react-native", "React Native"}, {"nextjs", "Next.js"}}

func TestMultiSelectParsesNumbersInOptionOrder(t *testing.T) {
	p := NewLine(strings.NewReader("2, 1\n"), io.Discard)
	got, err := p.MultiSelect("Stacks?", opts, nil)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"react-native", "nextjs"}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMultiSelectEnterUsesDefaults(t *testing.T) {
	p := NewLine(strings.NewReader("\n"), io.Discard)
	got, err := p.MultiSelect("Stacks?", opts, []string{"nextjs"})
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"nextjs"}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMultiSelectRetriesOnInvalidInput(t *testing.T) {
	var out strings.Builder
	p := NewLine(strings.NewReader("\n9\nabc\n"), &out)
	_, err := p.MultiSelect("Stacks?", opts, nil)
	if !errors.Is(err, ErrAborted) {
		t.Fatalf("err = %v, want ErrAborted after 3 bad answers", err)
	}
	if !strings.Contains(out.String(), "select at least one option") ||
		!strings.Contains(out.String(), `"9" is not a number between 1 and 2`) {
		t.Errorf("missing validation messages:\n%s", out.String())
	}
}

func TestMultiSelectRecoversAfterInvalidInput(t *testing.T) {
	p := NewLine(strings.NewReader("0\n1\n"), io.Discard)
	got, err := p.MultiSelect("Stacks?", opts, nil)
	if err != nil || !reflect.DeepEqual(got, []string{"react-native"}) {
		t.Errorf("got %v, %v", got, err)
	}
}

func TestMultiSelectEOF(t *testing.T) {
	p := NewLine(strings.NewReader(""), io.Discard)
	if _, err := p.MultiSelect("Stacks?", opts, nil); !errors.Is(err, ErrAborted) {
		t.Errorf("err = %v, want ErrAborted", err)
	}
}

func TestInputDefaultAndValue(t *testing.T) {
	p := NewLine(strings.NewReader("\napps/web\n"), io.Discard)
	if got, _ := p.Input("Path", "apps/mobile"); got != "apps/mobile" {
		t.Errorf("default not used: %q", got)
	}
	if got, _ := p.Input("Path", "apps/mobile"); got != "apps/web" {
		t.Errorf("value not used: %q", got)
	}
}

func TestInputLastLineWithoutNewline(t *testing.T) {
	p := NewLine(strings.NewReader("demo"), io.Discard)
	if got, err := p.Input("Name", ""); err != nil || got != "demo" {
		t.Errorf("got %q, %v", got, err)
	}
}

func TestConfirm(t *testing.T) {
	cases := []struct {
		in   string
		def  bool
		want bool
	}{
		{"\n", true, true},
		{"\n", false, false},
		{"y\n", false, true},
		{"Sí\n", false, true},
		{"no\n", true, false},
		{"maybe\nn\n", true, false},
	}
	for _, c := range cases {
		p := NewLine(strings.NewReader(c.in), io.Discard)
		got, err := p.Confirm("Apply?", c.def)
		if err != nil || got != c.want {
			t.Errorf("Confirm(%q, %v) = %v, %v; want %v", c.in, c.def, got, err, c.want)
		}
	}
}
