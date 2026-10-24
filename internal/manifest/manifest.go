// Package manifest reads and writes .eva/config.json, the record of what eva
// installed in a project (stacks, their paths and agents).
package manifest

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Path is the manifest location relative to the project root.
const Path = ".eva/config.json"

// SchemaVersion is bumped when the manifest format changes incompatibly.
const SchemaVersion = 1

// StackRef is an installed stack and the app root it applies to.
type StackRef struct {
	ID   string `json:"id"`
	Path string `json:"path"`
}

// Manifest describes an eva installation.
type Manifest struct {
	SchemaVersion int        `json:"schemaVersion"`
	EvaVersion    string     `json:"evaVersion"`
	Project       string     `json:"project"`
	Stacks        []StackRef `json:"stacks"`
	Agents        []string   `json:"agents"`
}

// ErrNotFound is returned by Load when the project has no manifest.
var ErrNotFound = errors.New("eva manifest not found")

// Load reads the manifest under root. It returns ErrNotFound when the file
// does not exist.
func Load(root string) (*Manifest, error) {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(Path)))
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse %s: %w", Path, err)
	}
	if m.SchemaVersion > SchemaVersion {
		return nil, fmt.Errorf("%s was written by a newer eva (schema %d); upgrade eva", Path, m.SchemaVersion)
	}
	return &m, nil
}

// Marshal returns the manifest as indented JSON with a trailing newline.
func (m Manifest) Marshal() ([]byte, error) {
	m.SchemaVersion = SchemaVersion
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

// StackPath returns the path recorded for a stack id.
func (m Manifest) StackPath(id string) (string, bool) {
	for _, s := range m.Stacks {
		if s.ID == id {
			return s.Path, true
		}
	}
	return "", false
}

// HasAgent reports whether the agent id is installed.
func (m Manifest) HasAgent(id string) bool {
	for _, a := range m.Agents {
		if a == id {
			return true
		}
	}
	return false
}
