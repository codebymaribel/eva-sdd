package installer

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// LockPath is where eva records the hash of every file it installed.
const LockPath = ".eva/eva-lock.json"

// lockNotice is the first field of the lock, so a reader knows not to edit it.
const lockNotice = "Written by `eva init`. Do not edit. Commit it."

// Lock records what eva wrote, so a later `eva init` can tell files the user
// edited (hash differs) from files it may safely update (hash matches).
type Lock struct {
	Generated  string            `json:"_generated"`
	EvaVersion string            `json:"evaVersion"`
	Files      map[string]string `json:"files"`
}

// LoadLock reads the lock under root. It returns nil, nil when there is none
// (projects installed before locks existed, or a fresh project).
func LoadLock(root string) (*Lock, error) {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(LockPath)))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var l Lock
	if err := json.Unmarshal(data, &l); err != nil {
		return nil, fmt.Errorf("parse %s: %w (delete it and run `eva init --force` to rebuild it)", LockPath, err)
	}
	if l.Files == nil {
		l.Files = map[string]string{}
	}
	return &l, nil
}

// Marshal returns the lock as indented JSON with sorted file paths, the
// notice first and a trailing newline.
func (l Lock) Marshal() ([]byte, error) {
	l.Generated = lockNotice
	if l.Files == nil {
		l.Files = map[string]string{}
	}
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

// HashContent returns "sha256:<hex>" of content with CRLF normalised to LF,
// so a checkout that converts line endings does not look like an edit.
func HashContent(content []byte) string {
	sum := sha256.Sum256(normalize(content))
	return "sha256:" + hex.EncodeToString(sum[:])
}

// normalize converts CRLF line endings to LF.
func normalize(b []byte) []byte {
	return bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n"))
}
