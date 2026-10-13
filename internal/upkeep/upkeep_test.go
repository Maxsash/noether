package upkeep

import (
	"testing"
	"time"
)

func TestDays(t *testing.T) {
	for every, want := range map[string]int{"30d": 30, "2w": 14, " 1d ": 1} {
		if got, err := (Cadence{Every: every}).Days(); err != nil || got != want {
			t.Errorf("Days(%q) = %d, %v; want %d", every, got, err, want)
		}
	}
	for _, bad := range []string{"", "d", "0d", "-3d", "30", "1m", "monthly"} {
		if _, err := (Cadence{Every: bad}).Days(); err == nil {
			t.Errorf("Days(%q) accepted", bad)
		}
	}
}

func TestAssess(t *testing.T) {
	now := time.Date(2026, 10, 13, 9, 0, 0, 0, time.Local)
	run := func(daysAgo int) *Run { return &Run{Date: now.AddDate(0, 0, -daysAgo), Summary: "done"} }
	cadence := Cadence{Every: "30d", Commits: 40}
	cases := []struct {
		name    string
		last    *Run
		commits int
		state   string
	}{
		{"never logged", nil, 0, Never},
		{"recent", run(3), 5, OK},
		{"old enough", run(30), 0, Due},
		{"busy enough", run(3), 40, Due},
	}
	for _, c := range cases {
		if got := Assess("review", cadence, c.last, c.commits, now); got.State != c.state {
			t.Errorf("%s: state = %s, want %s (%+v)", c.name, got.State, c.state, got)
		}
	}
	if got := Assess("review", Cadence{Every: "30d"}, run(3), 500, now); got.State != OK {
		t.Errorf("commit threshold applied without being set: %+v", got)
	}
}

func TestDaysBetweenIgnoresClockChanges(t *testing.T) {
	berlin, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Skip("no time zone data")
	}
	before := time.Date(2026, 10, 24, 23, 30, 0, 0, berlin)
	after := time.Date(2026, 10, 26, 0, 15, 0, 0, berlin)
	if got := daysBetween(before, after); got != 2 {
		t.Errorf("daysBetween across a clock change = %d, want 2", got)
	}
}
