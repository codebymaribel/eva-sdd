package check

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// ErrNoFrontMatter is returned when a document does not start with "---".
var ErrNoFrontMatter = errors.New("missing front-matter (the file must start with ---)")

// ParseFrontMatter splits a Markdown document into its front-matter fields and
// body. It supports the flat `key: value` subset used by SDD documents,
// strips `# comments` and surrounding quotes.
func ParseFrontMatter(data []byte) (map[string]string, string, error) {
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	if !strings.HasPrefix(text, "---\n") {
		return nil, text, ErrNoFrontMatter
	}
	rest := "\n" + text[len("---\n"):] // leading \n so an empty header still matches
	end := strings.Index(rest, "\n---")
	if end == -1 {
		return nil, text, errors.New("front-matter is not closed with ---")
	}
	header, body := rest[:end], rest[end+len("\n---"):]
	body = strings.TrimPrefix(body, "\n")

	fields := map[string]string{}
	for i, line := range strings.Split(header, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		key, value, ok := strings.Cut(trimmed, ":")
		if !ok || strings.TrimSpace(key) == "" {
			return nil, body, fmt.Errorf("front-matter line %d: expected `key: value`, got %q", i+1, line)
		}
		fields[strings.TrimSpace(key)] = cleanValue(value)
	}
	return fields, body, nil
}

// cleanValue trims spaces, an inline " # comment" and matching quotes.
func cleanValue(v string) string {
	v = strings.TrimSpace(v)
	if strings.HasPrefix(v, "#") {
		return ""
	}
	if i := strings.Index(v, " #"); i >= 0 {
		v = strings.TrimSpace(v[:i])
	}
	if len(v) >= 2 && (v[0] == '"' || v[0] == '\'') && v[len(v)-1] == v[0] {
		v = v[1 : len(v)-1]
	}
	return v
}

var (
	fencedCode = regexp.MustCompile("(?s)```.*?```")
	inlineCode = regexp.MustCompile("`[^`\n]*`")
	needsClar  = regexp.MustCompile(`\[NEEDS CLARIFICATION`)
	taskLine   = regexp.MustCompile(`(?m)^\s*[-*] \[( |x|X)\] `)
)

var todoMarker = regexp.MustCompile(`\bTODO\b`)

// stripCode removes fenced and inline code so that markers quoted in
// instructions (written in backticks) are not counted.
func stripCode(body string) string {
	return inlineCode.ReplaceAllString(fencedCode.ReplaceAllString(body, ""), "")
}

// CountClarifications counts [NEEDS CLARIFICATION …] markers outside code.
func CountClarifications(body string) int {
	return len(needsClar.FindAllString(stripCode(body), -1))
}

// CountTODOs counts TODO placeholders outside code.
func CountTODOs(body string) int {
	return len(todoMarker.FindAllString(stripCode(body), -1))
}

// CountTasks returns how many checklist items are done and in total.
func CountTasks(body string) (done, total int) {
	for _, m := range taskLine.FindAllStringSubmatch(body, -1) {
		total++
		if m[1] != " " {
			done++
		}
	}
	return done, total
}
