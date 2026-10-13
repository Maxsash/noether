// Package git reads facts about a repository by running the git command line.
package git

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Repo struct {
	Dir string
}

// run never takes git's optional locks, so reading a repository cannot get in the way of
// someone committing to it at the same moment.
func (r Repo) run(args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", r.Dir, "-c", "core.quotePath=false"}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_OPTIONAL_LOCKS=0", "LC_ALL=C")
	out, err := cmd.Output()
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			return "", fmt.Errorf("git %s: %s", args[0], strings.TrimSpace(string(exit.Stderr)))
		}
		return "", err
	}
	return string(out), nil
}

func (r Repo) HasCommits() bool {
	_, err := r.run("rev-parse", "--verify", "--quiet", "HEAD")
	return err == nil
}

// Branch is empty when HEAD is detached.
func (r Repo) Branch() string {
	out, err := r.run("symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

func (r Repo) HasRemote() (bool, error) {
	out, err := r.run("remote")
	return strings.TrimSpace(out) != "", err
}

// Unpushed is false when the branch tracks no upstream.
func (r Repo) Unpushed() (int, bool) {
	out, err := r.run("rev-list", "--count", "@{upstream}..HEAD")
	if err != nil {
		return 0, false
	}
	n, err := strconv.Atoi(strings.TrimSpace(out))
	return n, err == nil
}

func (r Repo) Commits(limit int) ([]Commit, error) {
	out, err := r.run("log", "-n", strconv.Itoa(limit), logFormat)
	return parseLog(out), err
}

func (r Repo) Changes() ([]Change, error) {
	out, err := r.run("status", "--porcelain=v1", "-z")
	return parseStatus(out), err
}

func (r Repo) Worktrees() ([]Worktree, error) {
	out, err := r.run("worktree", "list", "--porcelain")
	return parseWorktrees(out), err
}

func (r Repo) CommitsAfter(t time.Time) (int, error) {
	out, err := r.run("rev-list", "--count", "--after="+t.Format(time.RFC3339), "HEAD")
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(out))
}

// LastTouched gives the date of the latest commit that changed each of the paths.
func (r Repo) LastTouched(paths []string) (map[string]time.Time, error) {
	if len(paths) == 0 {
		return map[string]time.Time{}, nil
	}
	out, err := r.run(append([]string{"log", lastTouchedFormat, "--name-only", "--"}, paths...)...)
	return parseLastTouched(out), err
}
