// Package textdiff renders line-based unified diffs of small text files (the
// templates eva installs). The standard library has no diff package.
package textdiff

import (
	"fmt"
	"strings"
)

// context is the number of unchanged lines shown around each change.
const context = 3

type opKind byte

const (
	same opKind = ' '
	del  opKind = '-'
	add  opKind = '+'
)

type op struct {
	kind       opKind
	line       string
	aIdx, bIdx int // 0-based line index in a / b before this op
}

// Unified returns a unified diff from a to b with 3 lines of context, or ""
// when they only differ in line endings or a missing final newline.
func Unified(aName, bName, a, b string) string {
	al, bl := splitLines(a), splitLines(b)
	ops := diffLines(al, bl)

	changed := false
	for _, o := range ops {
		if o.kind != same {
			changed = true
			break
		}
	}
	if !changed {
		return ""
	}

	var out strings.Builder
	fmt.Fprintf(&out, "--- %s\n+++ %s\n", aName, bName)
	for _, h := range hunks(ops) {
		writeHunk(&out, ops[h[0]:h[1]])
	}
	return out.String()
}

func splitLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.TrimSuffix(s, "\n")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

// diffLines computes an edit script with a longest-common-subsequence table.
// Templates are a few hundred lines at most, so O(n·m) is fine.
func diffLines(a, b []string) []op {
	n, m := len(a), len(b)
	lcs := make([][]int, n+1)
	for i := range lcs {
		lcs[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if a[i] == b[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else {
				lcs[i][j] = max(lcs[i+1][j], lcs[i][j+1])
			}
		}
	}
	var ops []op
	i, j := 0, 0
	for i < n || j < m {
		switch {
		case i < n && j < m && a[i] == b[j]:
			ops = append(ops, op{same, a[i], i, j})
			i++
			j++
		case j < m && (i == n || lcs[i][j+1] >= lcs[i+1][j]):
			ops = append(ops, op{add, b[j], i, j})
			j++
		default:
			ops = append(ops, op{del, a[i], i, j})
			i++
		}
	}
	return reorder(ops)
}

// reorder puts deletions before additions inside each run of changes, the
// conventional order in unified diffs.
func reorder(ops []op) []op {
	out := make([]op, 0, len(ops))
	for i := 0; i < len(ops); {
		if ops[i].kind == same {
			out = append(out, ops[i])
			i++
			continue
		}
		j := i
		for j < len(ops) && ops[j].kind != same {
			j++
		}
		for _, o := range ops[i:j] {
			if o.kind == del {
				out = append(out, o)
			}
		}
		for _, o := range ops[i:j] {
			if o.kind == add {
				out = append(out, o)
			}
		}
		i = j
	}
	return out
}

// hunks returns [start, end) ranges of ops: each change with its context,
// merging changes whose context overlaps or touches.
func hunks(ops []op) [][2]int {
	var out [][2]int
	for i, o := range ops {
		if o.kind == same {
			continue
		}
		start, end := max(0, i-context), min(len(ops), i+1+context)
		if len(out) > 0 && start <= out[len(out)-1][1] {
			out[len(out)-1][1] = max(out[len(out)-1][1], end)
			continue
		}
		out = append(out, [2]int{start, end})
	}
	return out
}

func writeHunk(out *strings.Builder, ops []op) {
	aStart, bStart := ops[0].aIdx, ops[0].bIdx
	aLen, bLen := 0, 0
	for _, o := range ops {
		if o.kind != add {
			aLen++
		}
		if o.kind != del {
			bLen++
		}
	}
	fmt.Fprintf(out, "@@ -%s +%s @@\n", rangeOf(aStart, aLen), rangeOf(bStart, bLen))
	for _, o := range ops {
		out.WriteByte(byte(o.kind))
		out.WriteString(o.line)
		out.WriteByte('\n')
	}
}

// rangeOf formats a hunk range: 1-based start and length ("0,0" style for
// empty ranges, as GNU diff does).
func rangeOf(start, length int) string {
	if length == 0 {
		return fmt.Sprintf("%d,0", start)
	}
	return fmt.Sprintf("%d,%d", start+1, length)
}
