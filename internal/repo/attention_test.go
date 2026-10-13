package repo

import (
	"strings"
	"testing"
	"time"

	"github.com/maxsash/noether/internal/git"
	"github.com/maxsash/noether/internal/upkeep"
)

func TestAttention(t *testing.T) {
	now := time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC)
	healthy := signals{
		now:        now,
		hasRemote:  true,
		state:      &Doc{Path: "docs/handoff.md", Lines: 80, Updated: now.Add(-2 * time.Hour)},
		latestWork: now.Add(-90 * time.Minute),
	}
	if got := attention(healthy); len(got) != 0 {
		t.Fatalf("a healthy repository needs attention: %+v", got)
	}

	cases := []struct {
		name  string
		edit  func(*signals)
		level string
		text  string
	}{
		{"no remote", func(s *signals) { s.hasRemote = false }, Warn, "exists only on this disk"},
		{"unpushed", func(s *signals) { s.unpushed = 3 }, Warn, "3 commits not pushed"},
		{"recent changes", func(s *signals) { s.changes, s.oldestChange = 2, now.Add(-3*time.Hour) }, Info, "2 files uncommitted; the oldest change is 3 hours old"},
		{"old changes", func(s *signals) { s.changes, s.oldestChange = 1, now.AddDate(0, 0, -9) }, Warn, "1 file uncommitted; the oldest change is 9 days old"},
		{"worktree", func(s *signals) { s.extraTrees = []git.Worktree{{Path: "/tmp/review"}} }, Warn, "Extra worktree at /tmp/review"},
		{"no state", func(s *signals) { s.state = nil }, Info, "No state file"},
		{"stale state", func(s *signals) { s.latestWork = now }, Warn, "Work changed after docs/handoff.md"},
		{"long state", func(s *signals) { s.state.Lines = 101 }, Warn, "101 lines; its budget is 100"},
		{"own budget", func(s *signals) { s.state.Lines, s.stateLines = 101, 150 }, "", ""},
		{"due job", func(s *signals) {
			s.upkeep = []upkeep.Status{{Kind: "review", State: upkeep.Due, Reason: "31 days since the last run"}}
		}, Warn, "review is due: 31 days since the last run"},
		{"never run", func(s *signals) { s.upkeep = []upkeep.Status{{Kind: "review", State: upkeep.Never}} }, "", ""},
		{"record problem", func(s *signals) { s.problems = []string{"docs/todo.md is missing"} }, Warn, "docs/todo.md is missing"},
	}
	for _, c := range cases {
		s := healthy
		state := *healthy.state
		s.state = &state
		c.edit(&s)
		got := attention(s)
		if c.text == "" {
			if len(got) != 0 {
				t.Errorf("%s: unexpected notices %+v", c.name, got)
			}
			continue
		}
		if len(got) != 1 || got[0].Level != c.level || !strings.Contains(got[0].Text, c.text) {
			t.Errorf("%s: got %+v, want one %s notice containing %q", c.name, got, c.level, c.text)
		}
	}
}
