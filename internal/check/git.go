package check

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Git is the subset of git that the approval gate needs. It is an interface
// so tests can use a fake repository.
type Git interface {
	// IsRepo reports whether the directory is inside a git work tree.
	IsRepo() bool
	// CurrentBranch returns the branch name, or "" when HEAD is detached.
	CurrentBranch() (string, error)
	// RefExists reports whether ref resolves to a commit.
	RefExists(ref string) bool
	// ChangedFiles lists files changed since the merge-base with base,
	// including uncommitted and untracked files.
	ChangedFiles(base string) ([]string, error)
}

// ExecGit runs the git binary in Dir.
type ExecGit struct {
	Dir string
}

func (g ExecGit) run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.Dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %v: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(stdout.String()), nil
}

// IsRepo implements Git.
func (g ExecGit) IsRepo() bool {
	out, err := g.run("rev-parse", "--is-inside-work-tree")
	return err == nil && out == "true"
}

// CurrentBranch implements Git.
func (g ExecGit) CurrentBranch() (string, error) {
	out, err := g.run("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	if out == "HEAD" {
		return "", nil
	}
	return out, nil
}

// RefExists implements Git.
func (g ExecGit) RefExists(ref string) bool {
	_, err := g.run("rev-parse", "--verify", "--quiet", ref+"^{commit}")
	return err == nil
}

// ChangedFiles implements Git.
func (g ExecGit) ChangedFiles(base string) ([]string, error) {
	mergeBase, err := g.run("merge-base", base, "HEAD")
	if err != nil {
		return nil, err
	}
	diff, err := g.run("diff", "--name-only", mergeBase)
	if err != nil {
		return nil, err
	}
	untracked, err := g.run("ls-files", "--others", "--exclude-standard")
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var files []string
	for _, f := range strings.Split(diff+"\n"+untracked, "\n") {
		if f = strings.TrimSpace(f); f != "" && !seen[f] {
			seen[f] = true
			files = append(files, f)
		}
	}
	return files, nil
}
