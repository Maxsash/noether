// Package repo gathers everything noether shows about one repository.
package repo

import (
	"cmp"
	"slices"
	"time"

	"github.com/maxsash/noether/internal/git"
	"github.com/maxsash/noether/internal/record"
	"github.com/maxsash/noether/internal/upkeep"
)

// Source is a repository to report on, with the settings its owner's config gives it.
type Source struct {
	ID       string
	Path     string
	Settings record.Settings
}

type Report struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Path         string          `json:"path"`
	Branch       string          `json:"branch"`
	LastActivity time.Time       `json:"lastActivity,omitzero"`
	Changes      []git.Change    `json:"changes"`
	Worktrees    []git.Worktree  `json:"worktrees"`
	State        *Doc            `json:"state,omitempty"`
	StateText    string          `json:"-"`
	Tasks        []record.Task   `json:"tasks"`
	Upkeep       []upkeep.Status `json:"upkeep"`
	History      []Event         `json:"history"`
	Docs         []Doc           `json:"docs"`
	Attention    []Notice        `json:"attention"`
}

type Doc struct {
	Path    string    `json:"path"`
	Lines   int       `json:"lines"`
	Updated time.Time `json:"updated,omitzero"`
}

type Event struct {
	When  time.Time `json:"when"`
	Kind  string    `json:"kind"`
	Title string    `json:"title"`
	Ref   string    `json:"ref"`
}

const (
	CommitEvent = "commit"
	UpkeepEvent = "upkeep"
)

type Summary struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	Branch       string    `json:"branch"`
	LastActivity time.Time `json:"lastActivity,omitzero"`
	Warnings     int       `json:"warnings"`
	OpenTasks    int       `json:"openTasks"`
	Due          int       `json:"due"`
}

func (r Report) Summary() Summary {
	s := Summary{ID: r.ID, Name: r.Name, Path: r.Path, Branch: r.Branch, LastActivity: r.LastActivity}
	for _, n := range r.Attention {
		if n.Level == Warn {
			s.Warnings++
		}
	}
	for _, t := range r.Tasks {
		if !t.Done {
			s.OpenTasks++
		}
	}
	for _, job := range r.Upkeep {
		if job.State != upkeep.OK {
			s.Due++
		}
	}
	return s
}

// history interleaves commits and upkeep runs, newest first.
func history(commits []git.Commit, entries []record.Entry, limit int) []Event {
	events := make([]Event, 0, len(commits)+len(entries))
	for _, c := range commits {
		events = append(events, Event{When: c.When, Kind: CommitEvent, Title: c.Subject, Ref: c.Hash})
	}
	for _, e := range entries {
		events = append(events, Event{When: e.Date, Kind: UpkeepEvent, Title: e.Summary, Ref: e.Kind})
	}
	slices.SortStableFunc(events, func(a, b Event) int { return b.When.Compare(a.When) })
	return events[:min(limit, len(events))]
}

// SortByActivity puts the most recently worked-on repositories first.
func SortByActivity(reports []Report) {
	slices.SortStableFunc(reports, func(a, b Report) int {
		return cmp.Or(b.LastActivity.Compare(a.LastActivity), cmp.Compare(a.ID, b.ID))
	})
}
