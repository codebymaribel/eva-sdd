// Package check validates SDD documents and enforces the approval gate:
// no feature code without an approved spec, plan and tasks.
package check

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/codebymaribel/eva-sdd/internal/manifest"
)

// Severity of a finding. Errors make `eva check` exit non-zero.
type Severity string

const (
	SevError   Severity = "error"
	SevWarning Severity = "warning"
)

// Finding is one problem found by the checker.
type Finding struct {
	Severity Severity
	Code     string
	Path     string
	Message  string
}

// Doc is a parsed SDD document (constitution, spec, plan, tasks or verification).
type Doc struct {
	Path     string
	Fields   map[string]string
	Body     string
	ParseErr error
}

// Status returns the front-matter status ("" if unparsable).
func (d *Doc) Status() string {
	if d == nil || d.Fields == nil {
		return ""
	}
	return d.Fields["status"]
}

// Approved reports whether the document is approved or done.
func (d *Doc) Approved() bool {
	s := d.Status()
	return s == "approved" || s == "done"
}

// Feature is a specs/NNN-slug directory.
type Feature struct {
	Dir               string // slash path relative to root, e.g. specs/004-login
	ID                string // 004-login
	Number            string // 004
	Spec, Plan, Tasks *Doc   // nil when the file does not exist
	Clarify           *Doc   // nil when the file does not exist
	Verification      *Doc   // nil when the file does not exist
}

// Options configure a check run.
type Options struct {
	Base   string // base ref for the diff; auto-detected when empty
	Branch string // branch name; auto-detected when empty
	Strict bool   // treat warnings about untracked code changes as errors
	NoGit  bool   // skip the branch gate
	Git    Git    // defaults to ExecGit in the project root
	Getenv func(string) string
}

// ConstitutionPath is the constitution location relative to the project root.
const ConstitutionPath = ".eva/constitution.md"

// Report is the result of a check run.
type Report struct {
	Constitution *Doc // nil when the file does not exist
	Features     []Feature
	Branch       string
	Base         string
	GateChecked  bool
	CodeChanges  int
	Findings     []Finding
}

// Count returns the number of findings with the given severity.
func (r Report) Count(s Severity) int {
	n := 0
	for _, f := range r.Findings {
		if f.Severity == s {
			n++
		}
	}
	return n
}

// Valid statuses for SDD documents.
var validStatuses = map[string]bool{"draft": true, "in-review": true, "approved": true, "done": true}

var (
	featureDir    = regexp.MustCompile(`^(\d{3,})-[a-z0-9]+(?:-[a-z0-9]+)*$`)
	branchFeature = regexp.MustCompile(`(?:^|/)(\d{3,})-[a-z0-9]`)
)

// Run checks the project at root.
func Run(root string, opts Options) (Report, error) {
	var r Report
	if _, err := manifest.Load(root); err != nil {
		if errors.Is(err, manifest.ErrNotFound) {
			r.add(SevError, "EVA001", manifest.Path, "not an eva project: run `eva init` first")
			return r, nil
		}
		return r, err
	}

	features, err := scanSpecs(root, &r)
	if err != nil {
		return r, err
	}
	r.Features = features
	if r.Constitution, err = loadDoc(root, ConstitutionPath); err != nil {
		return r, err
	}
	checkConstitution(r.Constitution, len(features), &r)
	for _, f := range features {
		validateFeature(f, &r)
	}
	checkDuplicates(features, &r)

	if !opts.NoGit {
		if err := checkGate(root, opts, &r); err != nil {
			return r, err
		}
	}

	sort.SliceStable(r.Findings, func(i, j int) bool {
		a, b := r.Findings[i], r.Findings[j]
		if a.Path != b.Path {
			return a.Path < b.Path
		}
		return a.Code < b.Code
	})
	return r, nil
}

func (r *Report) add(sev Severity, code, p, format string, args ...any) {
	r.Findings = append(r.Findings, Finding{Severity: sev, Code: code, Path: p, Message: fmt.Sprintf(format, args...)})
}

// scanSpecs reads every feature directory under specs/.
func scanSpecs(root string, r *Report) ([]Feature, error) {
	entries, err := os.ReadDir(filepath.Join(root, "specs"))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var features []Feature
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := path.Join("specs", e.Name())
		m := featureDir.FindStringSubmatch(e.Name())
		if m == nil {
			r.add(SevWarning, "EVA010", dir, "folder name should be NNN-kebab-slug (e.g. 004-user-login); it is ignored")
			continue
		}
		f := Feature{Dir: dir, ID: e.Name(), Number: m[1]}
		for name, target := range map[string]**Doc{"spec.md": &f.Spec, "plan.md": &f.Plan, "tasks.md": &f.Tasks, "clarify.md": &f.Clarify, "verification.md": &f.Verification} {
			doc, err := loadDoc(root, path.Join(dir, name))
			if err != nil {
				return nil, err
			}
			*target = doc
		}
		features = append(features, f)
	}
	return features, nil
}

// loadDoc parses a document; it returns nil when the file does not exist.
func loadDoc(root, rel string) (*Doc, error) {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	fields, body, perr := ParseFrontMatter(data)
	return &Doc{Path: rel, Fields: fields, Body: body, ParseErr: perr}, nil
}

// validateFeature applies document and phase-order rules to one feature.
func validateFeature(f Feature, r *Report) {
	if f.Spec == nil {
		r.add(SevError, "EVA020", f.Dir, "missing spec.md")
	}
	for _, d := range []*Doc{f.Spec, f.Plan, f.Tasks} {
		if d != nil {
			validateDoc(f, d, r)
		}
	}
	open := 0
	if f.Clarify != nil {
		open = validateClarify(f, f.Clarify, r)
	}
	checkClarifyGate(f, open, r)
	if f.Verification != nil {
		validateVerification(f, f.Verification, r)
	}
	if f.Plan != nil {
		checkPlanTraceability(f, r)
	}
	if f.Tasks != nil {
		checkTasks(f, r)
	}
	checkVerification(f, r)

	if f.Spec.Approved() && f.Spec.ParseErr == nil {
		if n := CountClarifications(f.Spec.Body); n > 0 {
			r.add(SevError, "EVA025", f.Spec.Path, "approved spec still has %d [NEEDS CLARIFICATION] marker(s)", n)
		}
		checkSpecContent(f.Spec, r)
	}
	if f.Plan != nil && !f.Spec.Approved() {
		r.add(SevError, "EVA030", f.Plan.Path, "plan.md exists but spec.md is not approved (status: %s)", statusOrNone(f.Spec))
	}
	if f.Tasks != nil && !f.Plan.Approved() {
		r.add(SevError, "EVA031", f.Tasks.Path, "tasks.md exists but plan.md is not approved (status: %s)", statusOrNone(f.Plan))
	}
	done, total := 0, 0
	if f.Tasks != nil {
		done, total = CountTasks(f.Tasks.Body)
		if done > 0 && !f.Tasks.Approved() {
			r.add(SevError, "EVA032", f.Tasks.Path, "%d task(s) ticked but tasks.md is not approved (status: %s)", done, statusOrNone(f.Tasks))
		}
	}
	if f.Spec.Status() == "done" && (f.Tasks == nil || total == 0 || done < total) {
		r.add(SevError, "EVA033", f.Spec.Path, "spec is done but tasks are incomplete (%d/%d ticked)", done, total)
	}
}

// validateDoc checks front-matter fields common to spec.md, plan.md and tasks.md.
func validateDoc(f Feature, d *Doc, r *Report) {
	if d.ParseErr != nil {
		r.add(SevError, "EVA021", d.Path, "invalid front-matter: %v", d.ParseErr)
		return
	}
	if id := d.Fields["id"]; id != f.ID {
		r.add(SevError, "EVA022", d.Path, "front-matter id is %q, expected %q (the folder name)", id, f.ID)
	}
	status := d.Status()
	if !validStatuses[status] {
		r.add(SevError, "EVA023", d.Path, "invalid status %q (want draft, in-review, approved or done)", status)
		return
	}
	if d.Approved() && strings.TrimSpace(d.Fields["approved_by"]) == "" {
		r.add(SevError, "EVA024", d.Path, "status is %s but approved_by is empty (only a human approves)", status)
	}
	checkMeta(d, docType(d.Path), lifecycleRules{hasStatus: true, canClose: true}, r)
}

// validateVerification checks verification.md, which has no status of its
// own: the developer closes the feature by marking the spec done.
func validateVerification(f Feature, d *Doc, r *Report) {
	if d.ParseErr != nil {
		r.add(SevError, "EVA021", d.Path, "invalid front-matter: %v", d.ParseErr)
		return
	}
	if id := d.Fields["id"]; id != f.ID {
		r.add(SevError, "EVA022", d.Path, "front-matter id is %q, expected %q (the folder name)", id, f.ID)
	}
	checkMeta(d, "verification", lifecycleRules{hasStatus: false, canClose: true}, r)
}

// docType is the expected front-matter type of a feature document: its file
// name without the extension (spec, plan, tasks, verification).
func docType(p string) string {
	return strings.TrimSuffix(path.Base(p), path.Ext(p))
}

// checkConstitution validates .eva/constitution.md. Specs may only exist once
// the constitution is approved; with no specs yet, a pending constitution is a
// warning so a freshly initialised project still passes.
func checkConstitution(c *Doc, features int, r *Report) {
	if c == nil {
		r.add(SevError, "EVA050", ConstitutionPath, "missing constitution: run `eva init` to restore it")
		return
	}
	if c.ParseErr != nil {
		r.add(SevError, "EVA051", ConstitutionPath, "invalid front-matter: %v", c.ParseErr)
		return
	}
	status := c.Status()
	if status != "draft" && status != "in-review" && status != "approved" {
		r.add(SevError, "EVA051", ConstitutionPath, "invalid status %q (want draft, in-review or approved)", status)
		return
	}
	checkMeta(c, "constitution", lifecycleRules{hasStatus: true, canClose: false}, r)
	if status == "approved" {
		if strings.TrimSpace(c.Fields["approved_by"]) == "" {
			r.add(SevError, "EVA052", ConstitutionPath, "status is approved but approved_by is empty (only a human approves)")
		}
		if n := CountTODOs(c.Body); n > 0 {
			r.add(SevError, "EVA054", ConstitutionPath, "approved constitution still has %d TODO marker(s)", n)
		}
		return
	}
	if features > 0 {
		r.add(SevError, "EVA053", ConstitutionPath, "%d feature spec(s) exist but the constitution is not approved (status: %s)", features, status)
		return
	}
	r.add(SevWarning, "EVA053", ConstitutionPath, "constitution is not approved yet (status: %s): run /eva-constitution and approve it before writing specs", status)
}

// checkDuplicates reports feature numbers used by more than one folder.
func checkDuplicates(features []Feature, r *Report) {
	byNumber := map[string][]string{}
	for _, f := range features {
		byNumber[f.Number] = append(byNumber[f.Number], f.ID)
	}
	for number, ids := range byNumber {
		if len(ids) > 1 {
			sort.Strings(ids)
			r.add(SevError, "EVA011", "specs/", "feature number %s is used by %s; renumber one of them", number, strings.Join(ids, ", "))
		}
	}
}

// checkGate links the current branch to a feature and requires its documents
// to be approved when the branch contains code changes.
func checkGate(root string, opts Options, r *Report) error {
	getenv := opts.Getenv
	if getenv == nil {
		getenv = func(string) string { return "" }
	}
	git := opts.Git
	if git == nil {
		git = ExecGit{Dir: root}
	}
	if !git.IsRepo() {
		r.add(SevWarning, "EVA040", ".", "not a git repository: branch gate skipped")
		return nil
	}

	branch := firstNonEmpty(opts.Branch, getenv("GITHUB_HEAD_REF"))
	if branch == "" {
		b, err := git.CurrentBranch()
		if err != nil {
			return err
		}
		branch = b
	}
	if branch == "" {
		r.add(SevWarning, "EVA045", ".", "detached HEAD: pass --branch to check the gate")
		return nil
	}
	r.Branch = branch

	base := resolveBase(opts.Base, getenv("GITHUB_BASE_REF"), git)
	if base == "" {
		r.add(SevWarning, "EVA044", ".", "base branch not found (tried %s): pass --base to check the gate", strings.Join(baseCandidates(opts.Base, getenv("GITHUB_BASE_REF")), ", "))
		return nil
	}
	r.Base = base
	if strings.TrimPrefix(base, "origin/") == branch {
		return nil // on the base branch itself: nothing to gate
	}

	files, err := git.ChangedFiles(base)
	if err != nil {
		r.add(SevWarning, "EVA046", ".", "could not diff against %s: %v", base, err)
		return nil
	}
	var code []string
	for _, f := range files {
		if isCodePath(f) {
			code = append(code, f)
		}
	}
	r.GateChecked = true
	r.CodeChanges = len(code)

	m := branchFeature.FindStringSubmatch(branch)
	if m == nil {
		if len(code) > 0 {
			sev := SevWarning
			if opts.Strict {
				sev = SevError
			}
			r.add(sev, "EVA043", ".", "branch %q has %d code change(s) but no feature id; use feat/NNN-slug", branch, len(code))
		}
		return nil
	}

	var feature *Feature
	for i := range r.Features {
		if r.Features[i].Number == m[1] {
			feature = &r.Features[i]
			break
		}
	}
	if feature == nil {
		sev := SevWarning
		if len(code) > 0 {
			sev = SevError
		}
		r.add(sev, "EVA041", "specs/", "branch %q refers to feature %s but no specs/%s-* folder exists", branch, m[1], m[1])
		return nil
	}
	if len(code) == 0 {
		return nil
	}

	var pending []string
	if !r.Constitution.Approved() {
		pending = append(pending, fmt.Sprintf("constitution (%s)", statusOrNone(r.Constitution)))
	}
	for _, d := range []struct {
		name string
		doc  *Doc
	}{{"spec.md", feature.Spec}, {"plan.md", feature.Plan}, {"tasks.md", feature.Tasks}} {
		if !d.doc.Approved() {
			pending = append(pending, fmt.Sprintf("%s (%s)", d.name, statusOrNone(d.doc)))
		}
	}
	if len(pending) > 0 {
		r.add(SevError, "EVA042", feature.Dir, "%d code change(s) on %q but not approved yet: %s", len(code), branch, strings.Join(pending, ", "))
	}
	return nil
}

// resolveBase returns the first existing candidate base ref.
func resolveBase(flagBase, githubBase string, git Git) string {
	for _, ref := range baseCandidates(flagBase, githubBase) {
		if git.RefExists(ref) {
			return ref
		}
	}
	return ""
}

func baseCandidates(flagBase, githubBase string) []string {
	if flagBase != "" {
		return []string{flagBase}
	}
	if githubBase != "" {
		return []string{"origin/" + githubBase, githubBase}
	}
	return []string{"origin/main", "main", "origin/master", "master"}
}

// isCodePath reports whether a changed file counts as feature code. SDD
// documents, docs, agent configuration and root Markdown files do not.
func isCodePath(p string) bool {
	for _, prefix := range []string{"specs/", ".eva/", "docs/", ".claude/", ".opencode/"} {
		if strings.HasPrefix(p, prefix) {
			return false
		}
	}
	if !strings.Contains(p, "/") && strings.HasSuffix(strings.ToLower(p), ".md") {
		return false
	}
	return true
}

func statusOrNone(d *Doc) string {
	if d == nil {
		return "missing"
	}
	if s := d.Status(); s != "" {
		return s
	}
	return "none"
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
