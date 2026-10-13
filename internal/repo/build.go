package repo

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/maxsash/noether/internal/git"
	"github.com/maxsash/noether/internal/guard"
	"github.com/maxsash/noether/internal/record"
	"github.com/maxsash/noether/internal/upkeep"
)

const historyLimit = 40

const parallelBuilds = 8

var errDenied = errors.New("denied by the deny list")

// BuildAll reports on every source at once, most recently active first.
func BuildAll(sources []Source, g guard.Guard, now time.Time) []Report {
	reports := make([]Report, len(sources))
	slots := make(chan struct{}, parallelBuilds)
	var wg sync.WaitGroup
	for i, src := range sources {
		wg.Go(func() {
			slots <- struct{}{}
			defer func() { <-slots }()
			reports[i] = Build(src, g, now)
		})
	}
	wg.Wait()
	SortByActivity(reports)
	return reports
}

// Build reads one repository's git state and record. Nothing it does writes to the repository.
func Build(src Source, g guard.Guard, now time.Time) Report {
	b := &builder{repo: git.Repo{Dir: src.Path}, files: files{root: src.Path, guard: g}}
	r := Report{ID: src.ID, Name: filepath.Base(src.Path), Path: src.Path}

	docs := b.files.markdown()
	settings, err := resolveSettings(src, b.files, docs)
	b.note(err)
	facts := b.gitFacts()
	r.Branch, r.Changes, r.Worktrees = facts.branch, facts.changes, facts.extraTrees

	updated := b.updatedTimes(facts, union(docs, recordFiles(settings)))
	r.Docs = b.docs(docs, updated)
	r.State, r.StateText = b.state(settings.State, updated)
	r.Tasks = b.tasks(settings.Tasks)
	entries := b.log(settings.Log)
	r.Upkeep = b.upkeep(settings, entries, facts.hasCommits, now)
	r.History = history(facts.commits, entries, historyLimit)

	statePath := ""
	if r.State != nil {
		statePath = r.State.Path
	}
	oldest, newest, newestElsewhere := activity(facts.commits, facts.modified, statePath, now)
	r.LastActivity = newest
	r.Attention = attention(signals{
		now:          now,
		hasRemote:    facts.hasRemote,
		unpushed:     facts.unpushed,
		changes:      len(r.Changes),
		oldestChange: oldest,
		extraTrees:   r.Worktrees,
		state:        r.State,
		stateLines:   settings.StateLines,
		latestWork:   newestElsewhere,
		upkeep:       r.Upkeep,
		problems:     b.problems,
	})
	return r
}

// builder gathers one report, collecting problems rather than stopping at the first.
type builder struct {
	repo     git.Repo
	files    files
	problems []string
}

func (b *builder) note(err error) {
	if err != nil {
		b.problems = append(b.problems, err.Error())
	}
}

type gitFacts struct {
	branch     string
	hasRemote  bool
	unpushed   int
	hasCommits bool
	commits    []git.Commit
	changes    []git.Change
	extraTrees []git.Worktree
	modified   map[string]time.Time
}

func (b *builder) gitFacts() gitFacts {
	var f gitFacts
	var err error
	f.branch = b.repo.Branch()
	f.hasRemote, err = b.repo.HasRemote()
	b.note(err)
	f.unpushed, _ = b.repo.Unpushed()
	if f.hasCommits = b.repo.HasCommits(); f.hasCommits {
		f.commits, err = b.repo.Commits(historyLimit)
		b.note(err)
	}
	f.changes, err = b.repo.Changes()
	b.note(err)
	trees, err := b.repo.Worktrees()
	b.note(err)
	if len(trees) > 1 {
		f.extraTrees = trees[1:]
	}
	f.modified = b.files.modifiedTimes(f.changes)
	return f
}

// updatedTimes says when each file last changed: on disk if it has uncommitted edits,
// otherwise in its latest commit, so a fresh clone does not make every file look new.
func (b *builder) updatedTimes(facts gitFacts, paths []string) func(string) time.Time {
	touched := map[string]time.Time{}
	if facts.hasCommits {
		var err error
		touched, err = b.repo.LastTouched(paths)
		b.note(err)
	}
	return func(rel string) time.Time {
		if t, ok := facts.modified[rel]; ok {
			return t
		}
		if t, ok := touched[rel]; ok {
			return t
		}
		return b.files.modTime(rel)
	}
}

func (b *builder) docs(paths []string, updated func(string) time.Time) []Doc {
	var docs []Doc
	for _, rel := range paths {
		if text, err := b.files.read(rel); err == nil {
			docs = append(docs, Doc{Path: rel, Lines: countLines(text), Updated: updated(rel)})
		}
	}
	return docs
}

func (b *builder) state(rel string, updated func(string) time.Time) (*Doc, string) {
	if rel == "" {
		return nil, ""
	}
	text, err := b.files.read(rel)
	if err != nil {
		b.note(missing(rel, err))
		return nil, ""
	}
	return &Doc{Path: rel, Lines: countLines(text), Updated: updated(rel)}, text
}

func (b *builder) tasks(paths []string) []record.Task {
	var tasks []record.Task
	for _, rel := range paths {
		text, err := b.files.read(rel)
		if err != nil {
			b.note(missing(rel, err))
			continue
		}
		tasks = append(tasks, record.Tasks(rel, text)...)
	}
	return tasks
}

func (b *builder) log(rel string) []record.Entry {
	if rel == "" {
		return nil
	}
	text, err := b.files.read(rel)
	if err != nil {
		b.note(missing(rel, err))
		return nil
	}
	return record.Log(text)
}

// upkeep counts commits from the day after each run, since a log entry carries only a date.
func (b *builder) upkeep(settings record.Settings, entries []record.Entry, hasCommits bool, now time.Time) []upkeep.Status {
	latest := record.Latest(entries)
	var jobs []upkeep.Status
	for _, kind := range settings.Kinds() {
		var last *upkeep.Run
		commitsSince := 0
		if e, ok := latest[kind]; ok {
			last = &upkeep.Run{Date: e.Date, Summary: e.Summary}
			if hasCommits {
				var err error
				commitsSince, err = b.repo.CommitsAfter(e.Date.AddDate(0, 0, 1))
				b.note(err)
			}
		}
		jobs = append(jobs, upkeep.Assess(kind, settings.Upkeep[kind], last, commitsSince, now))
	}
	return jobs
}

// activity finds the age of the oldest uncommitted change, the latest work of any kind, and
// the latest work outside the state file, which is what the state file should describe.
func activity(commits []git.Commit, modified map[string]time.Time, statePath string, now time.Time) (oldest, newest, newestElsewhere time.Time) {
	if len(commits) > 0 {
		newest, newestElsewhere = commits[0].When, commits[0].When
	}
	oldest = now
	for rel, t := range modified {
		oldest, newest = minTime(oldest, t), maxTime(newest, t)
		if rel != statePath {
			newestElsewhere = maxTime(newestElsewhere, t)
		}
	}
	return oldest, newest, newestElsewhere
}

func resolveSettings(src Source, files files, docs []string) (record.Settings, error) {
	settings := record.Detect(docs).Merge(src.Settings)
	text, err := files.read("AGENTS.md")
	if err != nil {
		return settings, nil
	}
	block, ok := record.Block(text)
	if !ok {
		return settings, nil
	}
	own, err := record.ParseSettings(block)
	if err != nil {
		return settings, fmt.Errorf("AGENTS.md noether block: %w", err)
	}
	return settings.Merge(own), nil
}

func recordFiles(s record.Settings) []string {
	var named []string
	for _, rel := range append([]string{s.State, s.Log}, s.Tasks...) {
		if rel != "" {
			named = append(named, rel)
		}
	}
	return named
}

func missing(rel string, err error) error {
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("%s is named in the record but does not exist", rel)
	}
	return fmt.Errorf("%s: %w", rel, err)
}

// files opens paths inside one repository, refusing any the guard denies.
type files struct {
	root  string
	guard guard.Guard
}

func (f files) abs(rel string) string {
	return filepath.Join(f.root, filepath.FromSlash(rel))
}

func (f files) read(rel string) (string, error) {
	if !f.guard.Allows(f.abs(rel)) {
		return "", errDenied
	}
	data, err := os.ReadFile(f.abs(rel))
	return string(data), err
}

func (f files) modTime(rel string) time.Time {
	info, err := os.Lstat(f.abs(rel))
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}

// markdown lists the markdown files at the top level and in a docs folder of any case.
func (f files) markdown() []string {
	var found []string
	var docsDir string
	entries, _ := os.ReadDir(f.root)
	for _, e := range entries {
		switch {
		case e.IsDir() && strings.EqualFold(e.Name(), "docs"):
			docsDir = e.Name()
		case !e.IsDir() && isMarkdown(e.Name()):
			found = append(found, e.Name())
		}
	}
	if docsDir != "" {
		entries, _ := os.ReadDir(f.abs(docsDir))
		for _, e := range entries {
			if !e.IsDir() && isMarkdown(e.Name()) {
				found = append(found, path.Join(docsDir, e.Name()))
			}
		}
	}
	return slices.DeleteFunc(found, func(rel string) bool { return !f.guard.Allows(f.abs(rel)) })
}

func (f files) modifiedTimes(changes []git.Change) map[string]time.Time {
	times := map[string]time.Time{}
	for _, c := range changes {
		if t := f.modTime(strings.TrimSuffix(c.Path, "/")); !t.IsZero() {
			times[c.Path] = t
		}
	}
	return times
}

func isMarkdown(name string) bool {
	return strings.EqualFold(path.Ext(name), ".md")
}

func countLines(text string) int {
	if text == "" {
		return 0
	}
	return strings.Count(strings.TrimSuffix(text, "\n"), "\n") + 1
}

func union(a, b []string) []string {
	all := slices.Concat(a, b)
	slices.Sort(all)
	return slices.Compact(all)
}

func minTime(a, b time.Time) time.Time {
	if b.Before(a) {
		return b
	}
	return a
}

func maxTime(a, b time.Time) time.Time {
	if b.After(a) {
		return b
	}
	return a
}
