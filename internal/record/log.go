package record

import (
	"regexp"
	"strings"
	"time"
)

// Entry is one line of an upkeep log: "- 2026-10-13 review: three fixes, all checks pass".
type Entry struct {
	Date    time.Time `json:"date"`
	Kind    string    `json:"kind"`
	Summary string    `json:"summary"`
	Line    int       `json:"line"`
}

var entryLine = regexp.MustCompile(`^\s*[-*+]\s+(\d{4}-\d{2}-\d{2})\s+([a-z0-9][a-z0-9-]*):\s*(.*)$`)

// Log reads upkeep entries in the order they appear; lines that are not entries are ignored.
func Log(markdown string) []Entry {
	var entries []Entry
	for _, l := range scan(markdown) {
		if l.fenced {
			continue
		}
		m := entryLine.FindStringSubmatch(l.text)
		if m == nil {
			continue
		}
		date, err := time.ParseInLocation(time.DateOnly, m[1], time.Local)
		if err != nil {
			continue
		}
		entries = append(entries, Entry{Date: date, Kind: m[2], Summary: strings.TrimSpace(m[3]), Line: l.number})
	}
	return entries
}

// Latest returns the most recent entry of each kind.
func Latest(entries []Entry) map[string]Entry {
	latest := map[string]Entry{}
	for _, e := range entries {
		if current, ok := latest[e.Kind]; !ok || e.Date.After(current.Date) {
			latest[e.Kind] = e
		}
	}
	return latest
}
