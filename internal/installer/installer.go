// Package installer turns a Config into a list of file actions (Plan) and
// writes them to disk (Apply). Planning is side-effect free, which makes
// --dry-run and golden tests trivial.
package installer

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/codebymaribel/eva-sdd/internal/catalog"
	"github.com/codebymaribel/eva-sdd/internal/manifest"
)

// Config is the final, resolved selection to install.
type Config struct {
	Project    string
	EvaVersion string
	Stacks     []manifest.StackRef // ids must exist in catalog.Stacks
	Agents     []string            // ids must exist in catalog.Agents
}

// Options tweak how existing files are treated.
type Options struct {
	// Force overwrites regular files that differ from the template, edited or not.
	Force bool
	// Prune deletes files that eva installed, no longer ships and nobody edited.
	Prune bool
}

// Op is the kind of change planned for a file.
type Op string

const (
	OpCreate    Op = "create"    // file does not exist yet
	OpUpdate    Op = "update"    // managed block/manifest changed, or forced overwrite
	OpUnchanged Op = "unchanged" // already up to date
	OpSkip      Op = "skip"      // exists with local changes (or obsolete); kept as is
	OpDelete    Op = "delete"    // obsolete and untouched; removed with --prune
)

// Action is one planned file change. Path is slash-separated and relative to
// the project root.
type Action struct {
	Path    string
	Op      Op
	Content []byte // what eva would write (nil for obsolete files)
	Current []byte // the file on disk, when it exists (used by --diff)
	Note    string

	diff bool // whether --diff shows Current → Content
}

// Diffable reports whether --diff shows this action: an existing file eva
// updates, or keeps although the template changed (or eva has no record of
// it). Files the user edited while the template stayed the same are not
// shown: the diff would only repeat their own edits.
func (a Action) Diffable() bool {
	return a.diff
}

// Template file name suffixes (see templates/embed.go).
const (
	suffixTemplate = ".tmpl"
	suffixBlock    = ".block"
)

// source is a template file mapped to its destination.
type source struct {
	src     string // path inside the template FS
	dest    string // path relative to the project root
	render  bool
	managed bool
}

// StackData is exposed to templates.
type StackData struct {
	ID, Name, Path string
}

// AgentData is exposed to templates.
type AgentData struct {
	ID, Name string
}

// Data is the root object passed to every *.tmpl file.
type Data struct {
	Project    string
	EvaVersion string
	Stacks     []StackData
	Agents     []AgentData
}

// Plan computes the actions needed to install cfg into root using the
// templates in tfs. It reads root but never writes to it.
func Plan(root string, tfs fs.FS, cfg Config, opts Options) ([]Action, error) {
	data, err := buildData(cfg)
	if err != nil {
		return nil, err
	}
	sources, err := collectSources(tfs, cfg)
	if err != nil {
		return nil, err
	}

	lock, err := LoadLock(root)
	if err != nil {
		return nil, err
	}
	installed := map[string]string{}
	if lock != nil {
		installed = lock.Files
	}
	next := Lock{EvaVersion: cfg.EvaVersion, Files: map[string]string{}}
	shipped := map[string]bool{}

	funcs := template.FuncMap{"stackPath": stackPathFunc(data)}
	actions := make([]Action, 0, len(sources)+2)
	for _, s := range sources {
		shipped[s.dest] = true
		raw, err := fs.ReadFile(tfs, s.src)
		if err != nil {
			return nil, err
		}
		content := raw
		if s.render {
			if content, err = render(s.src, raw, data, funcs); err != nil {
				return nil, err
			}
		}
		existing, exists, err := readExisting(root, s.dest)
		if err != nil {
			return nil, err
		}
		recorded, hasRecord := installed[s.dest]
		act, keep, err := planFile(s, content, existing, exists, recorded, hasRecord, opts)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", s.dest, err)
		}
		if keep != "" {
			next.Files[s.dest] = keep
		}
		actions = append(actions, act)
	}

	obsolete, err := planObsolete(root, installed, shipped, opts, next.Files)
	if err != nil {
		return nil, err
	}
	actions = append(actions, obsolete...)

	mAct, err := planManifest(root, cfg)
	if err != nil {
		return nil, err
	}
	lAct, err := planLock(root, next)
	if err != nil {
		return nil, err
	}
	actions = append(actions, mAct, lAct)

	sort.Slice(actions, func(i, j int) bool { return actions[i].Path < actions[j].Path })
	return actions, nil
}

// Apply writes every create/update action and removes every delete action.
func Apply(root string, actions []Action) error {
	for _, a := range actions {
		dest := filepath.Join(root, filepath.FromSlash(a.Path))
		if a.Op == OpDelete {
			if err := os.Remove(dest); err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
			continue
		}
		if a.Op != OpCreate && a.Op != OpUpdate {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(dest, a.Content, 0o644); err != nil {
			return err
		}
	}
	return nil
}

// Summary counts actions by operation.
func Summary(actions []Action) map[Op]int {
	out := map[Op]int{}
	for _, a := range actions {
		out[a.Op]++
	}
	return out
}

// buildData validates cfg against the catalog and builds template data.
func buildData(cfg Config) (Data, error) {
	if len(cfg.Stacks) == 0 {
		return Data{}, errors.New("at least one stack is required")
	}
	if len(cfg.Agents) == 0 {
		return Data{}, errors.New("at least one agent is required")
	}
	d := Data{Project: cfg.Project, EvaVersion: cfg.EvaVersion}
	for _, ref := range cfg.Stacks {
		s, ok := catalog.FindStack(ref.ID)
		if !ok {
			return Data{}, fmt.Errorf("unknown stack %q", ref.ID)
		}
		p := cleanAppPath(ref.Path)
		if path.IsAbs(p) || p == ".." || strings.HasPrefix(p, "../") {
			return Data{}, fmt.Errorf("stack %s: path %q must be relative to the project root", s.ID, ref.Path)
		}
		d.Stacks = append(d.Stacks, StackData{ID: s.ID, Name: s.Name, Path: p})
	}
	for _, id := range cfg.Agents {
		a, ok := catalog.FindAgent(id)
		if !ok {
			return Data{}, fmt.Errorf("unknown agent %q", id)
		}
		d.Agents = append(d.Agents, AgentData{ID: a.ID, Name: a.Name})
	}
	return d, nil
}

// collectSources walks core/, stacks/<id>/ and agents/<id>/ and maps every
// template file to its destination. Two templates targeting the same path is
// a template bug and returns an error.
func collectSources(tfs fs.FS, cfg Config) ([]source, error) {
	roots := []string{"core"}
	for _, s := range cfg.Stacks {
		roots = append(roots, path.Join("stacks", s.ID))
	}
	for _, a := range cfg.Agents {
		roots = append(roots, path.Join("agents", a))
	}

	var out []source
	owner := map[string]string{}
	for _, root := range roots {
		err := fs.WalkDir(tfs, root, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			s := toSource(root, p)
			if prev, dup := owner[s.dest]; dup {
				return fmt.Errorf("templates %s and %s both write %s", prev, p, s.dest)
			}
			owner[s.dest] = p
			out = append(out, s)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("templates for %s: %w", root, err)
		}
	}
	return out, nil
}

// toSource strips the template root and the .tmpl/.block suffixes.
func toSource(root, p string) source {
	s := source{src: p, dest: strings.TrimPrefix(p, root+"/")}
	if strings.HasSuffix(s.dest, suffixTemplate) {
		s.render = true
		s.dest = strings.TrimSuffix(s.dest, suffixTemplate)
	}
	if strings.HasSuffix(s.dest, suffixBlock) {
		s.managed = true
		s.dest = strings.TrimSuffix(s.dest, suffixBlock)
	}
	return s
}

// render executes a template with [[ ]] delimiters so that Markdown/JSX
// braces in templates never need escaping.
func render(name string, raw []byte, data Data, funcs template.FuncMap) ([]byte, error) {
	t, err := template.New(name).Delims("[[", "]]").Option("missingkey=error").Funcs(funcs).Parse(string(raw))
	if err != nil {
		return nil, fmt.Errorf("parse template %s: %w", name, err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("render template %s: %w", name, err)
	}
	return buf.Bytes(), nil
}

// stackPathFunc returns the template func `stackPath "<id>"`.
func stackPathFunc(d Data) func(string) (string, error) {
	return func(id string) (string, error) {
		for _, s := range d.Stacks {
			if s.ID == id {
				return s.Path, nil
			}
		}
		return "", fmt.Errorf("stack %q is not selected", id)
	}
}

// planFile decides what to do with one destination file. For regular files
// it also returns the hash to keep in the lock ("" to forget the file):
//
//   - missing                          → create
//   - same as the template             → unchanged
//   - --force                          → update
//   - matches the hash eva recorded    → update (nobody edited it; the template changed)
//   - anything else                    → skip (edited locally, or no install record)
func planFile(s source, content, existing []byte, exists bool, recorded string, hasRecord bool, opts Options) (Action, string, error) {
	act := Action{Path: s.dest}
	if exists {
		act.Current = existing
	}
	if s.managed {
		merged, err := MergeBlockFor(s.dest, existing, content)
		if err != nil {
			return act, "", err
		}
		act.Content = merged
		switch {
		case !exists:
			act.Op = OpCreate
		case bytes.Equal(existing, merged):
			act.Op = OpUnchanged
		default:
			act.Op = OpUpdate
			act.Note = "eva block updated, your content is kept"
			act.diff = true
		}
		return act, "", nil
	}

	act.Content = content
	newHash := HashContent(content)
	switch {
	case !exists:
		act.Op = OpCreate
		return act, newHash, nil
	case bytes.Equal(normalize(existing), normalize(content)):
		act.Op = OpUnchanged
		return act, newHash, nil
	case opts.Force:
		act.Op = OpUpdate
		act.Note = "overwritten (--force)"
		act.diff = true
		return act, newHash, nil
	case hasRecord && HashContent(existing) == recorded:
		act.Op = OpUpdate
		act.Note = "template updated (the file was not edited)"
		act.diff = true
		return act, newHash, nil
	}

	act.Op = OpSkip
	switch {
	case !hasRecord:
		act.Note = "differs from the template and eva has no install record: kept (compare with --diff, overwrite with --force)"
		act.diff = true
	case recorded == newHash:
		act.Note = "edited locally, kept"
	default:
		act.Note = "edited locally and the template changed: kept (compare with --diff, overwrite with --force)"
		act.diff = true
	}
	return act, recorded, nil
}

// planObsolete handles files eva installed (they are in the lock) but no
// longer ships. Untouched ones are deleted with --prune and otherwise
// reported; edited ones belong to the user now and leave the lock. The kept
// hashes are added to keep.
func planObsolete(root string, installed map[string]string, shipped map[string]bool, opts Options, keep map[string]string) ([]Action, error) {
	var out []Action
	for p, recorded := range installed {
		if shipped[p] || p == LockPath || p == manifest.Path {
			continue
		}
		existing, exists, err := readExisting(root, p)
		if err != nil {
			return nil, err
		}
		if !exists {
			continue
		}
		act := Action{Path: p, Current: existing}
		switch {
		case HashContent(existing) != recorded:
			act.Op = OpSkip
			act.Note = "obsolete and edited locally: kept, eva no longer tracks it"
		case opts.Prune:
			act.Op = OpDelete
			act.Note = "obsolete (--prune)"
		default:
			act.Op = OpSkip
			act.Note = "obsolete: eva no longer installs it; delete it with --prune"
			keep[p] = recorded
		}
		out = append(out, act)
	}
	return out, nil
}

// planLock writes the hashes of the installed files; like the manifest, the
// lock is owned by eva.
func planLock(root string, l Lock) (Action, error) {
	content, err := l.Marshal()
	if err != nil {
		return Action{}, err
	}
	existing, exists, err := readExisting(root, LockPath)
	if err != nil {
		return Action{}, err
	}
	act := Action{Path: LockPath, Content: content, Op: OpCreate}
	if exists {
		act.Op = OpUpdate
		if bytes.Equal(existing, content) {
			act.Op = OpUnchanged
		}
	}
	return act, nil
}

// planManifest always reflects cfg; the manifest is owned by eva.
func planManifest(root string, cfg Config) (Action, error) {
	m := manifest.Manifest{
		EvaVersion: cfg.EvaVersion,
		Project:    cfg.Project,
		Agents:     append([]string(nil), cfg.Agents...),
	}
	for _, s := range cfg.Stacks {
		m.Stacks = append(m.Stacks, manifest.StackRef{ID: s.ID, Path: cleanAppPath(s.Path)})
	}
	content, err := m.Marshal()
	if err != nil {
		return Action{}, err
	}
	existing, exists, err := readExisting(root, manifest.Path)
	if err != nil {
		return Action{}, err
	}
	act := Action{Path: manifest.Path, Content: content, Op: OpCreate}
	if exists {
		act.Op = OpUpdate
		if bytes.Equal(existing, content) {
			act.Op = OpUnchanged
		}
	}
	return act, nil
}

func readExisting(root, rel string) ([]byte, bool, error) {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return data, true, nil
}

// cleanAppPath normalises an app root to a clean slash path ("." for root).
func cleanAppPath(p string) string {
	p = strings.TrimSpace(filepath.ToSlash(p))
	if p == "" {
		return "."
	}
	return path.Clean(p)
}
