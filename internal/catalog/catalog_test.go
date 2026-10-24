package catalog

import (
	"reflect"
	"testing"
)

func TestFindStackByIDAndAlias(t *testing.T) {
	for _, in := range []string{"react-native", "RN", " expo "} {
		s, ok := FindStack(in)
		if !ok || s.ID != "react-native" {
			t.Errorf("FindStack(%q) = %q, %v; want react-native", in, s.ID, ok)
		}
	}
	if _, ok := FindStack("angular"); ok {
		t.Error("FindStack(angular) should not match")
	}
}

func TestFindAgent(t *testing.T) {
	a, ok := FindAgent("Claude-Code")
	if !ok || a.ID != "claude" {
		t.Errorf("FindAgent(Claude-Code) = %q, %v", a.ID, ok)
	}
}

func TestParseStacksDedupesAndKeepsCatalogOrder(t *testing.T) {
	got, err := ParseStacks("next, rn ,nextjs,")
	if err != nil {
		t.Fatal(err)
	}
	ids := []string{}
	for _, s := range got {
		ids = append(ids, s.ID)
	}
	if want := []string{"react-native", "nextjs"}; !reflect.DeepEqual(ids, want) {
		t.Errorf("got %v, want %v", ids, want)
	}
}

func TestParseStacksUnknown(t *testing.T) {
	if _, err := ParseStacks("rn,flutter"); err == nil {
		t.Error("expected error for unknown stack")
	}
}

func TestParseAgents(t *testing.T) {
	got, err := ParseAgents("opencode,claude")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].ID != "claude" || got[1].ID != "opencode" {
		t.Errorf("unexpected agents: %+v", got)
	}
	if _, err := ParseAgents("cursor"); err == nil {
		t.Error("expected error for unsupported agent")
	}
}

func TestParseEmpty(t *testing.T) {
	s, err := ParseStacks("")
	if err != nil || len(s) != 0 {
		t.Errorf("ParseStacks(\"\") = %v, %v", s, err)
	}
}
