package repo

import (
	"fmt"
	"time"

	"github.com/maxsash/noether/internal/git"
	"github.com/maxsash/noether/internal/upkeep"
)

type Notice struct {
	Level string `json:"level"`
	Text  string `json:"text"`
}

const (
	Warn = "warn"
	Info = "info"
)

// staleAfter lets work and its state update land in separate commits without a warning.
const staleAfter = time.Hour

const oldChanges = 7 * 24 * time.Hour

const defaultStateLines = 100

type signals struct {
	now          time.Time
	hasRemote    bool
	unpushed     int
	changes      int
	oldestChange time.Time
	extraTrees   []git.Worktree
	state        *Doc
	stateLines   int
	latestWork   time.Time
	upkeep       []upkeep.Status
	problems     []string
}

func attention(s signals) []Notice {
	var notices []Notice
	add := func(level, format string, args ...any) {
		notices = append(notices, Notice{Level: level, Text: fmt.Sprintf(format, args...)})
	}
	for _, problem := range s.problems {
		add(Warn, "%s.", problem)
	}
	if !s.hasRemote {
		add(Warn, "No git remote: this repository exists only on this disk.")
	}
	if s.unpushed > 0 {
		add(Warn, "%s not pushed.", plural(s.unpushed, "commit"))
	}
	if s.changes > 0 {
		age := s.now.Sub(s.oldestChange)
		level := Info
		if age >= oldChanges {
			level = Warn
		}
		add(level, "%s uncommitted; the oldest change is %s old.", plural(s.changes, "file"), duration(age))
	}
	for _, tree := range s.extraTrees {
		add(Warn, "Extra worktree at %s.", tree.Path)
	}
	switch {
	case s.state == nil:
		add(Info, "No state file, so nothing says where this repository stands.")
	case s.latestWork.Sub(s.state.Updated) > staleAfter:
		add(Warn, "Work changed after %s was last updated.", s.state.Path)
	}
	if s.state != nil && s.state.Lines > budget(s.stateLines) {
		add(Warn, "%s is %d lines; its budget is %d.", s.state.Path, s.state.Lines, budget(s.stateLines))
	}
	for _, job := range s.upkeep {
		if job.State == upkeep.Due {
			add(Warn, "%s is due: %s.", job.Kind, job.Reason)
		}
	}
	return notices
}

func budget(lines int) int {
	if lines > 0 {
		return lines
	}
	return defaultStateLines
}

func plural(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return fmt.Sprintf("%d %ss", n, noun)
}

func duration(d time.Duration) string {
	switch {
	case d < time.Hour:
		return plural(int(d.Minutes()), "minute")
	case d < 48*time.Hour:
		return plural(int(d.Hours()), "hour")
	default:
		return plural(int(d.Hours()/24), "day")
	}
}
