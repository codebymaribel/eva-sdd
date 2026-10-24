// Package catalog lists the stacks and agents eva knows how to install.
package catalog

import (
	"fmt"
	"strings"
)

// Stack is a technology overlay (rules, structure, performance checklist).
type Stack struct {
	ID          string   // canonical id, also the template folder name
	Name        string   // human-readable name
	Aliases     []string // accepted spellings on the command line
	DefaultPath string   // suggested app root when several stacks share a repo
}

// Agent is a coding agent whose shortcut commands eva can generate.
type Agent struct {
	ID      string
	Name    string
	Aliases []string
}

// Stacks is the ordered list of supported stacks.
var Stacks = []Stack{
	{ID: "react-native", Name: "React Native (Expo)", Aliases: []string{"rn", "expo"}, DefaultPath: "apps/mobile"},
	{ID: "nextjs", Name: "Next.js (App Router)", Aliases: []string{"next", "next.js"}, DefaultPath: "apps/web"},
}

// Agents is the ordered list of supported agents.
var Agents = []Agent{
	{ID: "claude", Name: "Claude Code", Aliases: []string{"claude-code"}},
	{ID: "opencode", Name: "OpenCode", Aliases: []string{"open-code"}},
}

// FindStack resolves an id or alias (case-insensitive) to a Stack.
func FindStack(name string) (Stack, bool) {
	n := normalize(name)
	for _, s := range Stacks {
		if n == s.ID || contains(s.Aliases, n) {
			return s, true
		}
	}
	return Stack{}, false
}

// FindAgent resolves an id or alias (case-insensitive) to an Agent.
func FindAgent(name string) (Agent, bool) {
	n := normalize(name)
	for _, a := range Agents {
		if n == a.ID || contains(a.Aliases, n) {
			return a, true
		}
	}
	return Agent{}, false
}

// ParseStacks parses a comma-separated list ("rn,next") into stacks,
// removing duplicates and returning them in catalog order.
func ParseStacks(csv string) ([]Stack, error) {
	seen := map[string]bool{}
	for _, part := range splitCSV(csv) {
		s, ok := FindStack(part)
		if !ok {
			return nil, fmt.Errorf("unknown stack %q (available: %s)", part, strings.Join(StackIDs(), ", "))
		}
		seen[s.ID] = true
	}
	var out []Stack
	for _, s := range Stacks {
		if seen[s.ID] {
			out = append(out, s)
		}
	}
	return out, nil
}

// ParseAgents parses a comma-separated list ("claude,opencode") into agents,
// removing duplicates and returning them in catalog order.
func ParseAgents(csv string) ([]Agent, error) {
	seen := map[string]bool{}
	for _, part := range splitCSV(csv) {
		a, ok := FindAgent(part)
		if !ok {
			return nil, fmt.Errorf("unknown agent %q (available: %s)", part, strings.Join(AgentIDs(), ", "))
		}
		seen[a.ID] = true
	}
	var out []Agent
	for _, a := range Agents {
		if seen[a.ID] {
			out = append(out, a)
		}
	}
	return out, nil
}

// StackIDs returns the canonical ids of all stacks.
func StackIDs() []string {
	ids := make([]string, len(Stacks))
	for i, s := range Stacks {
		ids[i] = s.ID
	}
	return ids
}

// AgentIDs returns the canonical ids of all agents.
func AgentIDs() []string {
	ids := make([]string, len(Agents))
	for i, a := range Agents {
		ids[i] = a.ID
	}
	return ids
}

func splitCSV(csv string) []string {
	var out []string
	for _, p := range strings.Split(csv, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func normalize(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}
