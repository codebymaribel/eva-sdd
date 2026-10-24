package textdiff

import "testing"

func TestUnifiedIdentical(t *testing.T) {
	if got := Unified("a", "b", "x\ny\n", "x\ny\n"); got != "" {
		t.Errorf("identical inputs must produce no diff, got %q", got)
	}
}

func TestUnifiedChange(t *testing.T) {
	a := "1\n2\n3\n4\n5\n6\n7\n8\n9\n"
	b := "1\n2\n3\n4\nfive\n6\n7\n8\n9\nten\n"
	want := "--- yours\n+++ template\n" +
		"@@ -2,8 +2,9 @@\n" +
		" 2\n 3\n 4\n-5\n+five\n 6\n 7\n 8\n 9\n+ten\n"
	if got := Unified("yours", "template", a, b); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestUnifiedSeparateHunks(t *testing.T) {
	var a, b string
	for i := 1; i <= 20; i++ {
		line := string(rune('a'+i-1)) + "\n"
		a += line
		switch i {
		case 2:
			b += "B\n"
		case 19:
			b += "S\n"
		default:
			b += line
		}
	}
	got := Unified("x", "y", a, b)
	want := "--- x\n+++ y\n" +
		"@@ -1,5 +1,5 @@\n a\n-b\n+B\n c\n d\n e\n" +
		"@@ -16,5 +16,5 @@\n p\n q\n r\n-s\n+S\n t\n"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestUnifiedNormalisesCRLFAndMissingNewline(t *testing.T) {
	if got := Unified("a", "b", "x\r\ny\r\n", "x\ny"); got != "" {
		t.Errorf("CRLF and a missing final newline are not content changes, got %q", got)
	}
}
