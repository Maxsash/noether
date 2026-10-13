package git

import (
	"strings"
	"time"
)

type Commit struct {
	Hash    string    `json:"hash"`
	When    time.Time `json:"when"`
	Subject string    `json:"subject"`
}

type Change struct {
	Path   string `json:"path"`
	Status string `json:"status"`
}

type Worktree struct {
	Path   string `json:"path"`
	Branch string `json:"branch,omitempty"`
}

const (
	field  = "\x1f"
	record = "\x1e"
)

// logFormat pairs with parseLog.
const logFormat = "--format=%h" + field + "%cI" + field + "%s"

func parseLog(out string) []Commit {
	var commits []Commit
	for _, row := range strings.Split(strings.TrimSpace(out), "\n") {
		parts := strings.SplitN(row, field, 3)
		if len(parts) != 3 {
			continue
		}
		when, err := time.Parse(time.RFC3339, parts[1])
		if err != nil {
			continue
		}
		commits = append(commits, Commit{Hash: parts[0], When: when, Subject: parts[2]})
	}
	return commits
}

// parseStatus reads `git status --porcelain=v1 -z`, where a rename is followed by its old path.
func parseStatus(out string) []Change {
	var changes []Change
	entries := strings.Split(out, "\x00")
	for i := 0; i < len(entries); i++ {
		entry := entries[i]
		if len(entry) < 4 {
			continue
		}
		code := entry[:2]
		changes = append(changes, Change{Path: entry[3:], Status: strings.TrimSpace(code)})
		if code[0] == 'R' || code[0] == 'C' {
			i++
		}
	}
	return changes
}

// parseWorktrees reads `git worktree list --porcelain`. The first entry is the main worktree.
func parseWorktrees(out string) []Worktree {
	var trees []Worktree
	for _, block := range strings.Split(strings.TrimSpace(out), "\n\n") {
		var tree Worktree
		for _, row := range strings.Split(block, "\n") {
			key, value, _ := strings.Cut(row, " ")
			switch key {
			case "worktree":
				tree.Path = value
			case "branch":
				tree.Branch = strings.TrimPrefix(value, "refs/heads/")
			}
		}
		if tree.Path != "" {
			trees = append(trees, tree)
		}
	}
	return trees
}

// lastTouchedFormat pairs with parseLastTouched.
const lastTouchedFormat = "--format=" + record + "%cI"

// parseLastTouched reads `git log --name-only` output, newest first, into the date each
// path last changed.
func parseLastTouched(out string) map[string]time.Time {
	touched := map[string]time.Time{}
	for _, commit := range strings.Split(out, record) {
		rows := strings.Split(strings.TrimSpace(commit), "\n")
		when, err := time.Parse(time.RFC3339, rows[0])
		if err != nil {
			continue
		}
		for _, path := range rows[1:] {
			if path = strings.TrimSpace(path); path != "" {
				if _, seen := touched[path]; !seen {
					touched[path] = when
				}
			}
		}
	}
	return touched
}
