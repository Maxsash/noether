// Package upkeep decides when recurring jobs such as a code review are due again.
package upkeep

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Cadence says how often a job should run: every so many days, or sooner once enough
// commits have landed since the last run.
type Cadence struct {
	Every   string `toml:"every" json:"every"`
	Commits int    `toml:"commits" json:"commits,omitempty"`
}

const (
	Never = "never"
	Due   = "due"
	OK    = "ok"
)

type Run struct {
	Date    time.Time
	Summary string
}

type Status struct {
	Kind         string    `json:"kind"`
	Cadence      Cadence   `json:"cadence"`
	Last         time.Time `json:"last,omitzero"`
	Summary      string    `json:"summary,omitempty"`
	DaysSince    int       `json:"daysSince"`
	CommitsSince int       `json:"commitsSince"`
	State        string    `json:"state"`
	Reason       string    `json:"reason,omitempty"`
}

// Days accepts a count of days ("30d") or weeks ("2w").
func (c Cadence) Days() (int, error) {
	unit := map[string]int{"d": 1, "w": 7}
	s := strings.TrimSpace(c.Every)
	if len(s) < 2 {
		return 0, fmt.Errorf("every = %q: want a number of days or weeks, like 30d or 2w", c.Every)
	}
	n, err := strconv.Atoi(s[:len(s)-1])
	multiplier, known := unit[s[len(s)-1:]]
	if err != nil || !known || n <= 0 {
		return 0, fmt.Errorf("every = %q: want a number of days or weeks, like 30d or 2w", c.Every)
	}
	return n * multiplier, nil
}

// Assess reports whether a job is due. last is nil when the job has never been logged.
func Assess(kind string, c Cadence, last *Run, commitsSince int, now time.Time) Status {
	s := Status{Kind: kind, Cadence: c, State: OK}
	if last == nil {
		s.State, s.Reason = Never, "never logged"
		return s
	}
	s.Last, s.Summary, s.CommitsSince = last.Date, last.Summary, commitsSince
	s.DaysSince = daysBetween(last.Date, now)
	days, _ := c.Days()
	switch {
	case s.DaysSince >= days:
		s.State, s.Reason = Due, fmt.Sprintf("%d days since the last run", s.DaysSince)
	case c.Commits > 0 && commitsSince >= c.Commits:
		s.State, s.Reason = Due, fmt.Sprintf("%d commits since the last run", commitsSince)
	}
	return s
}

func daysBetween(from, to time.Time) int {
	noon := func(t time.Time) time.Time {
		y, m, d := t.Date()
		return time.Date(y, m, d, 12, 0, 0, 0, time.UTC)
	}
	return int(noon(to).Sub(noon(from)).Hours() / 24)
}
