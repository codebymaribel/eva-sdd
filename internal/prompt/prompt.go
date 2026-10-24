// Package prompt implements the interactive questions of `eva init` using
// only the standard library. The Prompter interface allows swapping in a
// richer TUI (e.g. charmbracelet/huh) later without touching the CLI.
package prompt

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ErrAborted is returned when input ends (Ctrl-D) or retries are exhausted.
var ErrAborted = errors.New("aborted")

// maxAttempts bounds how many invalid answers are tolerated per question.
const maxAttempts = 3

// Option is a selectable choice.
type Option struct {
	Value string
	Label string
}

// Prompter asks the user questions.
type Prompter interface {
	// MultiSelect returns the chosen values (at least one), in option order.
	MultiSelect(question string, options []Option, defaults []string) ([]string, error)
	// Input returns free text, or def when the answer is empty.
	Input(question, def string) (string, error)
	// Confirm returns a yes/no answer, or def when the answer is empty.
	Confirm(question string, def bool) (bool, error)
}

// Line is a line-based Prompter reading from in and writing to out.
type Line struct {
	in  *bufio.Reader
	out io.Writer
}

// NewLine creates a line-based prompter.
func NewLine(in io.Reader, out io.Writer) *Line {
	return &Line{in: bufio.NewReader(in), out: out}
}

// MultiSelect prints a numbered list and accepts "1,2", "1 2" or Enter for
// the defaults (marked with *).
func (l *Line) MultiSelect(question string, options []Option, defaults []string) ([]string, error) {
	isDefault := map[string]bool{}
	for _, d := range defaults {
		isDefault[d] = true
	}
	fmt.Fprintf(l.out, "\n%s\n", question)
	for i, o := range options {
		mark := " "
		if isDefault[o.Value] {
			mark = "*"
		}
		fmt.Fprintf(l.out, "  %s %d) %s\n", mark, i+1, o.Label)
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		hint := "numbers separated by commas"
		if len(defaults) > 0 {
			hint += ", Enter for *"
		}
		fmt.Fprintf(l.out, "Select (%s): ", hint)
		answer, err := l.readLine()
		if err != nil {
			return nil, err
		}
		if answer == "" && len(defaults) > 0 {
			return orderedValues(options, isDefault), nil
		}
		chosen, err := parseSelection(answer, len(options))
		if err != nil {
			fmt.Fprintf(l.out, "  %v\n", err)
			continue
		}
		picked := map[string]bool{}
		for _, idx := range chosen {
			picked[options[idx].Value] = true
		}
		return orderedValues(options, picked), nil
	}
	return nil, ErrAborted
}

// Input asks for free text.
func (l *Line) Input(question, def string) (string, error) {
	if def != "" {
		fmt.Fprintf(l.out, "%s [%s]: ", question, def)
	} else {
		fmt.Fprintf(l.out, "%s: ", question)
	}
	answer, err := l.readLine()
	if err != nil {
		return "", err
	}
	if answer == "" {
		return def, nil
	}
	return answer, nil
}

// Confirm asks a yes/no question.
func (l *Line) Confirm(question string, def bool) (bool, error) {
	choices := "y/N"
	if def {
		choices = "Y/n"
	}
	for attempt := 0; attempt < maxAttempts; attempt++ {
		fmt.Fprintf(l.out, "%s [%s]: ", question, choices)
		answer, err := l.readLine()
		if err != nil {
			return false, err
		}
		switch strings.ToLower(answer) {
		case "":
			return def, nil
		case "y", "yes", "s", "si", "sí":
			return true, nil
		case "n", "no":
			return false, nil
		}
		fmt.Fprintln(l.out, "  please answer y or n")
	}
	return false, ErrAborted
}

// readLine returns the next trimmed line; EOF with no data is ErrAborted.
func (l *Line) readLine() (string, error) {
	s, err := l.in.ReadString('\n')
	if err != nil && !(errors.Is(err, io.EOF) && s != "") {
		if errors.Is(err, io.EOF) {
			return "", ErrAborted
		}
		return "", err
	}
	return strings.TrimSpace(s), nil
}

// parseSelection converts "1, 3" into zero-based indexes, validating range.
func parseSelection(answer string, n int) ([]int, error) {
	fields := strings.FieldsFunc(answer, func(r rune) bool { return r == ',' || r == ' ' })
	if len(fields) == 0 {
		return nil, errors.New("select at least one option")
	}
	out := make([]int, 0, len(fields))
	for _, f := range fields {
		i, err := strconv.Atoi(f)
		if err != nil || i < 1 || i > n {
			return nil, fmt.Errorf("%q is not a number between 1 and %d", f, n)
		}
		out = append(out, i-1)
	}
	return out, nil
}

func orderedValues(options []Option, set map[string]bool) []string {
	var out []string
	for _, o := range options {
		if set[o.Value] {
			out = append(out, o.Value)
		}
	}
	return out
}
