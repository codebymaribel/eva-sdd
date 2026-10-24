package installer

import (
	"bytes"
	"errors"
	"path"
	"strings"
)

// Markers delimiting the eva-managed block inside user files such as AGENTS.md.
const (
	StartMarker = "<!-- eva:start — managed by eva, do not edit inside this block; re-run `eva init` to update -->"
	EndMarker   = "<!-- eva:end -->"
)

// ErrCorruptBlock is returned when the markers are unbalanced.
var ErrCorruptBlock = errors.New("eva block markers are unbalanced (found start without end, or end before start)")

// blockStyle is the comment syntax used for the markers of one kind of file.
type blockStyle struct {
	start, end, startPrefix string
}

var (
	// htmlBlock is used in Markdown files (AGENTS.md, CLAUDE.md).
	htmlBlock = blockStyle{StartMarker, EndMarker, "<!-- eva:start"}
	// hashBlock is used in files whose comments start with # (.gitattributes):
	// an HTML comment there would be read as a pattern.
	hashBlock = blockStyle{
		"# eva:start — managed by eva, do not edit inside this block; re-run `eva init` to update",
		"# eva:end",
		"# eva:start",
	}
)

// styleFor picks the marker syntax for a destination file.
func styleFor(dest string) blockStyle {
	switch path.Base(dest) {
	case ".gitattributes", ".gitignore":
		return hashBlock
	}
	return htmlBlock
}

// WrapBlock surrounds content with the Markdown (HTML comment) markers.
func WrapBlock(content []byte) []byte {
	return htmlBlock.wrap(content)
}

// MergeBlock inserts or replaces the eva-managed block in a Markdown file.
//
//   - existing empty           → the block alone
//   - existing without markers → existing + blank line + block (user content kept on top)
//   - existing with markers    → the block between the markers is replaced in place
func MergeBlock(existing, content []byte) ([]byte, error) {
	return htmlBlock.merge(existing, content)
}

// MergeBlockFor is MergeBlock with the marker syntax that suits dest.
func MergeBlockFor(dest string, existing, content []byte) ([]byte, error) {
	return styleFor(dest).merge(existing, content)
}

func (s blockStyle) wrap(content []byte) []byte {
	var b bytes.Buffer
	b.WriteString(s.start)
	b.WriteByte('\n')
	b.Write(bytes.TrimRight(content, "\n"))
	b.WriteByte('\n')
	b.WriteString(s.end)
	b.WriteByte('\n')
	return b.Bytes()
}

func (s blockStyle) merge(existing, content []byte) ([]byte, error) {
	block := s.wrap(content)
	if len(bytes.TrimSpace(existing)) == 0 {
		return block, nil
	}

	lines := strings.SplitAfter(string(existing), "\n")
	start, end := -1, -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, s.startPrefix):
			if start != -1 {
				return nil, ErrCorruptBlock
			}
			start = i
		case trimmed == s.end:
			if start == -1 || end != -1 {
				return nil, ErrCorruptBlock
			}
			end = i
		}
	}

	switch {
	case start == -1 && end == -1:
		out := bytes.TrimRight(existing, "\n")
		out = append(out, '\n', '\n')
		return append(out, block...), nil
	case start == -1 || end == -1:
		return nil, ErrCorruptBlock
	}

	var b strings.Builder
	for _, l := range lines[:start] {
		b.WriteString(l)
	}
	b.Write(block)
	for _, l := range lines[end+1:] {
		b.WriteString(l)
	}
	return []byte(b.String()), nil
}
