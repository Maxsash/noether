package repo

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/maxsash/noether/internal/guard"
	"github.com/maxsash/noether/internal/upkeep"
)

func gitIn(t *testing.T, dir string, env []string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir, "-c", "user.name=test", "-c", "user.email=test@example.com", "-c", "commit.gpgsign=false"}, args...)...)
	cmd.Env = append(os.Environ(), env...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func write(t *testing.T, dir, rel, content string) {
	t.Helper()
	path := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func commitAt(t *testing.T, dir, when, message string) {
	t.Helper()
	gitIn(t, dir, nil, "add", "-A")
	gitIn(t, dir, []string{"GIT_AUTHOR_DATE=" + when, "GIT_COMMITTER_DATE=" + when}, "commit", "-q", "-m", message)
}

func TestBuild(t *testing.T) {
	dir := t.TempDir()
	gitIn(t, dir, nil, "init", "-q", "-b", "main")
	write(t, dir, "AGENTS.md", "# Agents\n\n```noether\n"+
		"state = \"docs/handoff.md\"\ntasks = [\"docs/todo.md\"]\nlog = \"docs/log.md\"\n"+
		"[upkeep]\nreview = { every = \"30d\" }\ntesting = { every = \"7d\" }\n```\n")
	write(t, dir, "docs/handoff.md", "# Handoff\n\n## Now\nBuilding.\n")
	write(t, dir, "docs/todo.md", "- [ ] first\n- [x] second\n")
	write(t, dir, "docs/log.md", "- 2026-10-01 review: two fixes\n")
	commitAt(t, dir, "2026-10-03T10:00:00Z", "docs: start the record")
	write(t, dir, "main.go", "package main\n")
	commitAt(t, dir, "2026-10-05T10:00:00Z", "feat: add main")
	write(t, dir, "notes.txt", "draft\n")

	g, err := guard.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC)
	r := Build(Source{ID: "demo", Path: dir}, g, now)

	if r.Branch != "main" || r.State == nil || r.State.Path != "docs/handoff.md" || r.State.Lines != 4 {
		t.Fatalf("branch %q, state %+v", r.Branch, r.State)
	}
	if len(r.Tasks) != 2 || r.Summary().OpenTasks != 1 {
		t.Errorf("tasks = %+v", r.Tasks)
	}
	jobs := map[string]upkeep.Status{}
	for _, job := range r.Upkeep {
		jobs[job.Kind] = job
	}
	if jobs["review"].State != upkeep.OK || jobs["review"].CommitsSince != 2 || jobs["testing"].State != upkeep.Never {
		t.Errorf("upkeep = %+v", r.Upkeep)
	}
	if r.History[0].Title != "feat: add main" || !slices.ContainsFunc(r.History, func(e Event) bool { return e.Kind == UpkeepEvent }) {
		t.Errorf("history = %+v", r.History)
	}
	var docs []string
	for _, d := range r.Docs {
		docs = append(docs, d.Path)
	}
	if want := []string{"AGENTS.md", "docs/handoff.md", "docs/log.md", "docs/todo.md"}; !slices.Equal(docs, want) {
		t.Errorf("docs = %v, want %v", docs, want)
	}
	text := noticeText(r.Attention)
	for _, want := range []string{"No git remote", "Work changed after docs/handoff.md", "1 file uncommitted"} {
		if !strings.Contains(text, want) {
			t.Errorf("attention lacks %q:\n%s", want, text)
		}
	}
}

func TestBuildReportsRecordProblems(t *testing.T) {
	dir := t.TempDir()
	gitIn(t, dir, nil, "init", "-q", "-b", "main")
	write(t, dir, "AGENTS.md", "```noether\ntasks = [\"docs/missing.md\"]\n[upkeep]\nreview = { every = \"soon\" }\n```\n")
	g, _ := guard.New(nil)
	r := Build(Source{ID: "empty", Path: dir}, g, time.Now())
	if text := noticeText(r.Attention); !strings.Contains(text, "noether block") {
		t.Errorf("a broken block went unreported:\n%s", text)
	}
	write(t, dir, "AGENTS.md", "```noether\ntasks = [\"docs/missing.md\"]\n```\n")
	r = Build(Source{ID: "empty", Path: dir}, g, time.Now())
	if text := noticeText(r.Attention); !strings.Contains(text, "docs/missing.md is named in the record but does not exist") {
		t.Errorf("a missing task file went unreported:\n%s", text)
	}
	if len(r.History) != 0 {
		t.Errorf("a repository without commits has history %+v", r.History)
	}
	if text := noticeText(r.Attention); strings.Contains(text, "fatal") {
		t.Errorf("a repository without commits reported a git error:\n%s", text)
	}
}

func noticeText(notices []Notice) string {
	var lines []string
	for _, n := range notices {
		lines = append(lines, n.Level+": "+n.Text)
	}
	return strings.Join(lines, "\n")
}
