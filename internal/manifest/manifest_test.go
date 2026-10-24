package manifest

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadMissingReturnsErrNotFound(t *testing.T) {
	_, err := Load(t.TempDir())
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestMarshalLoadRoundTrip(t *testing.T) {
	root := t.TempDir()
	in := Manifest{
		EvaVersion: "v0.1.0",
		Project:    "demo",
		Stacks:     []StackRef{{ID: "react-native", Path: "apps/mobile"}},
		Agents:     []string{"claude"},
	}
	data, err := in.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if data[len(data)-1] != '\n' {
		t.Error("marshalled manifest should end with a newline")
	}
	writeFile(t, root, data)

	got, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	in.SchemaVersion = SchemaVersion
	if !reflect.DeepEqual(*got, in) {
		t.Errorf("got %+v, want %+v", *got, in)
	}
	if p, ok := got.StackPath("react-native"); !ok || p != "apps/mobile" {
		t.Errorf("StackPath = %q, %v", p, ok)
	}
	if !got.HasAgent("claude") || got.HasAgent("opencode") {
		t.Error("HasAgent mismatch")
	}
}

func TestLoadRejectsNewerSchema(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, []byte(`{"schemaVersion": 99}`))
	if _, err := Load(root); err == nil {
		t.Error("expected error for newer schema")
	}
}

func TestLoadRejectsInvalidJSON(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, []byte(`{nope`))
	if _, err := Load(root); err == nil || errors.Is(err, ErrNotFound) {
		t.Errorf("expected parse error, got %v", err)
	}
}

func writeFile(t *testing.T, root string, data []byte) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(Path))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, data, 0o644); err != nil {
		t.Fatal(err)
	}
}
