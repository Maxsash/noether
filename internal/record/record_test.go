package record

import (
	"reflect"
	"testing"
	"time"
)

const agents = "# Agents\n\n" +
	"Example only:\n\n" +
	"```toml\nstate = \"wrong.md\"\n```\n\n" +
	"## Record\n\n" +
	"````noether\n" +
	"state = \"docs/handoff.md\"\n" +
	"tasks = [\"docs/todo.md\"]\n" +
	"```\n" +
	"not the end: a shorter fence cannot close a longer one\n" +
	"```\n" +
	"[upkeep]\n" +
	"review = { every = \"30d\", commits = 40 }\n" +
	"````\n\n" +
	"```noether\nstate = \"second.md\"\n```\n"

func TestBlockReadsOnlyTheFirstTaggedFence(t *testing.T) {
	body, ok := Block(agents)
	if !ok {
		t.Fatal("block not found")
	}
	want := "state = \"docs/handoff.md\"\n" +
		"tasks = [\"docs/todo.md\"]\n" +
		"```\n" +
		"not the end: a shorter fence cannot close a longer one\n" +
		"```\n" +
		"[upkeep]\n" +
		"review = { every = \"30d\", commits = 40 }"
	if body != want {
		t.Errorf("body =\n%s\nwant\n%s", body, want)
	}
	if _, ok := Block("no block here"); ok {
		t.Error("found a block in plain text")
	}
}

func TestParseSettings(t *testing.T) {
	s, err := ParseSettings("state = \"docs/handoff.md\"\n[upkeep]\nreview = { every = \"2w\", commits = 40 }\n")
	if err != nil {
		t.Fatal(err)
	}
	if s.State != "docs/handoff.md" || s.Upkeep["review"].Commits != 40 {
		t.Errorf("settings = %+v", s)
	}
	if _, err := ParseSettings("stat = \"typo.md\"\n"); err == nil {
		t.Error("unknown key accepted")
	}
	if _, err := ParseSettings("[upkeep]\nreview = { every = \"monthly\" }\n"); err == nil {
		t.Error("bad cadence accepted")
	}
}

func TestTasksJoinWrappedLines(t *testing.T) {
	md := "- [ ] Choose the licence\n      before publishing.\n  - [ ] a nested task\n" +
		"- [x] done\nnot indented, so not part of it\n\n- [ ] last\n\n    after a blank line\n"
	var got []string
	for _, task := range Tasks("todo.md", md) {
		got = append(got, task.Text)
	}
	want := []string{"Choose the licence before publishing.", "a nested task", "done", "last"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Tasks = %q, want %q", got, want)
	}
}

func TestTasks(t *testing.T) {
	md := "# Plan\n- [ ] first\n  * [x] nested done\n\n## Later\n+ [X] **bold** done\n" +
		"```\n- [ ] in a code block\n```\n- [] not a checkbox\n"
	got := Tasks("todo.md", md)
	want := []Task{
		{File: "todo.md", Line: 2, Section: "Plan", Text: "first"},
		{File: "todo.md", Line: 3, Section: "Plan", Text: "nested done", Done: true},
		{File: "todo.md", Line: 6, Section: "Later", Text: "**bold** done", Done: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Tasks =\n%+v\nwant\n%+v", got, want)
	}
}

func TestLog(t *testing.T) {
	md := "# Upkeep log\n\n```\n- 2000-01-01 review: example in a fence\n```\n" +
		"- 2026-10-13 review: three fixes\n" +
		"- 2026-09-01 review: first pass\n" +
		"- 2026-10-02 testing: 12 new checks\n" +
		"- 2026-13-40 review: impossible date\n" +
		"- someday review: no date\n"
	entries := Log(md)
	if len(entries) != 3 {
		t.Fatalf("got %d entries: %+v", len(entries), entries)
	}
	latest := Latest(entries)
	if latest["review"].Summary != "three fixes" || latest["testing"].Line != 8 {
		t.Errorf("latest = %+v", latest)
	}
	if want := time.Date(2026, 10, 13, 0, 0, 0, 0, time.Local); !latest["review"].Date.Equal(want) {
		t.Errorf("date = %v, want %v", latest["review"].Date, want)
	}
}

func TestDetect(t *testing.T) {
	files := []string{"README.md", "TODO.md", "STATUS.md", "Docs/WorkPlan.md", "docs/handoff.md", "docs/follow-ups.md", "docs/log.md"}
	got := Detect(files)
	want := Settings{
		State: "docs/handoff.md",
		Tasks: []string{"TODO.md", "docs/follow-ups.md", "Docs/WorkPlan.md"},
		Log:   "docs/log.md",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Detect = %+v, want %+v", got, want)
	}
}

func TestMerge(t *testing.T) {
	base, _ := ParseSettings("state = \"a.md\"\nstate_lines = 100\n[upkeep]\nreview = { every = \"30d\" }\ntesting = { every = \"14d\" }\n")
	over, _ := ParseSettings("tasks = [\"b.md\"]\n[upkeep]\nreview = { every = \"7d\" }\n")
	got := base.Merge(over)
	if got.State != "a.md" || got.Tasks[0] != "b.md" || got.StateLines != 100 {
		t.Errorf("merged = %+v", got)
	}
	if got.Upkeep["review"].Every != "7d" || got.Upkeep["testing"].Every != "14d" {
		t.Errorf("upkeep = %+v", got.Upkeep)
	}
	if base.Upkeep["review"].Every != "30d" {
		t.Error("merge changed its receiver's upkeep map")
	}
}
