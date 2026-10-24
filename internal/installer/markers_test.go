package installer

import (
	"errors"
	"strings"
	"testing"
)

func TestMergeBlockIntoEmptyFile(t *testing.T) {
	got, err := MergeBlock(nil, []byte("hello\n"))
	if err != nil {
		t.Fatal(err)
	}
	want := StartMarker + "\nhello\n" + EndMarker + "\n"
	if string(got) != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestMergeBlockAppendsAfterUserContent(t *testing.T) {
	got, err := MergeBlock([]byte("# My rules\n\nkeep me\n\n\n"), []byte("eva rules"))
	if err != nil {
		t.Fatal(err)
	}
	want := "# My rules\n\nkeep me\n\n" + StartMarker + "\neva rules\n" + EndMarker + "\n"
	if string(got) != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestMergeBlockReplacesExistingBlockAndKeepsSurroundings(t *testing.T) {
	existing := "before\n" + StartMarker + "\nold\nstuff\n" + EndMarker + "\nafter\n"
	got, err := MergeBlock([]byte(existing), []byte("new"))
	if err != nil {
		t.Fatal(err)
	}
	want := "before\n" + StartMarker + "\nnew\n" + EndMarker + "\nafter\n"
	if string(got) != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestMergeBlockAcceptsOlderStartMarkerText(t *testing.T) {
	existing := "<!-- eva:start old wording -->\nold\n" + EndMarker + "\n"
	got, err := MergeBlock([]byte(existing), []byte("new"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(got), "old") {
		t.Errorf("old block not replaced: %q", got)
	}
}

func TestMergeBlockIsIdempotent(t *testing.T) {
	once, _ := MergeBlock([]byte("user\n"), []byte("eva"))
	twice, err := MergeBlock(once, []byte("eva"))
	if err != nil {
		t.Fatal(err)
	}
	if string(once) != string(twice) {
		t.Errorf("not idempotent:\n%q\n%q", once, twice)
	}
}

func TestMergeBlockRejectsCorruptMarkers(t *testing.T) {
	cases := map[string]string{
		"start only":    StartMarker + "\nx\n",
		"end only":      "x\n" + EndMarker + "\n",
		"end first":     EndMarker + "\n" + StartMarker + "\n",
		"double start":  StartMarker + "\n" + StartMarker + "\n" + EndMarker + "\n",
		"double blocks": StartMarker + "\n" + EndMarker + "\n" + EndMarker + "\n",
	}
	for name, in := range cases {
		if _, err := MergeBlock([]byte(in), []byte("x")); !errors.Is(err, ErrCorruptBlock) {
			t.Errorf("%s: err = %v, want ErrCorruptBlock", name, err)
		}
	}
}
