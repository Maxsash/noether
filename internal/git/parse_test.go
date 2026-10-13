package git

import (
	"reflect"
	"testing"
	"time"
)

func TestParseLog(t *testing.T) {
	out := "e47fea9\x1f2026-10-13T09:15:00+05:30\x1fdocs: note \x1f in a subject\n" +
		"broken row\n" +
		"3c65ccb\x1fnot a date\x1fskipped\n"
	got := parseLog(out)
	if len(got) != 1 || got[0].Hash != "e47fea9" || got[0].Subject != "docs: note \x1f in a subject" {
		t.Fatalf("parseLog = %+v", got)
	}
	if want := time.Date(2026, 10, 13, 3, 45, 0, 0, time.UTC); !got[0].When.Equal(want) {
		t.Errorf("when = %v, want %v", got[0].When, want)
	}
}

func TestParseStatus(t *testing.T) {
	out := " M docs/handoff.md\x00?? notes with spaces.md\x00R  new.go\x00old.go\x00D  gone.go\x00"
	want := []Change{
		{Path: "docs/handoff.md", Status: "M"},
		{Path: "notes with spaces.md", Status: "??"},
		{Path: "new.go", Status: "R"},
		{Path: "gone.go", Status: "D"},
	}
	if got := parseStatus(out); !reflect.DeepEqual(got, want) {
		t.Errorf("parseStatus = %+v, want %+v", got, want)
	}
}

func TestParseWorktrees(t *testing.T) {
	out := "worktree /repo\nHEAD abc\nbranch refs/heads/main\n\n" +
		"worktree /tmp/review\nHEAD abc\ndetached\n\n"
	want := []Worktree{{Path: "/repo", Branch: "main"}, {Path: "/tmp/review"}}
	if got := parseWorktrees(out); !reflect.DeepEqual(got, want) {
		t.Errorf("parseWorktrees = %+v, want %+v", got, want)
	}
}

func TestParseLastTouched(t *testing.T) {
	out := "\x1e2026-10-13T10:00:00Z\n\ndocs/handoff.md\nREADME.md\n" +
		"\x1e2026-10-01T10:00:00Z\n\ndocs/handoff.md\ndocs/todo.md\n"
	got := parseLastTouched(out)
	newest := time.Date(2026, 10, 13, 10, 0, 0, 0, time.UTC)
	older := time.Date(2026, 10, 1, 10, 0, 0, 0, time.UTC)
	want := map[string]time.Time{"docs/handoff.md": newest, "README.md": newest, "docs/todo.md": older}
	if len(got) != len(want) {
		t.Fatalf("parseLastTouched = %v", got)
	}
	for path, when := range want {
		if !got[path].Equal(when) {
			t.Errorf("%s = %v, want %v", path, got[path], when)
		}
	}
}
