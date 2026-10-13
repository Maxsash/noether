# Architecture

One Go binary. No database: repositories are the only source of truth, read afresh on
every request.

```
config ──► discover ──► repo.Build ──► server (page and JSON) / status (terminal)
             │              │
           guard      git · record · upkeep · attention
```

| Package            | Responsibility                                                                  |
| ------------------ | ------------------------------------------------------------------------------- |
| `cmd/noether`    | Commands and flags: `serve`, `status`; falls back to the current folder         |
| `internal/config`  | The owner's config file: roots, depth, deny list, editor links, record settings |
| `internal/guard`   | Which paths may be opened; secret files are always denied                       |
| `internal/discover`| Finds repositories under the roots and gives each a short unique name           |
| `internal/git`     | Runs read-only git commands (`git.go`) and parses their output (`parse.go`)     |
| `internal/record`  | Reads the record: settings block, conventional names, tasks, log entries        |
| `internal/upkeep`  | Pure: when an upkeep job is due                                                 |
| `internal/repo`    | Builds one report per repository; `attention.go` decides what needs attention   |
| `internal/server`  | Local HTTP server, markdown rendering, editor links, the embedded page          |

## Flow of one request

1. `discover.Sources` walks the roots, skipping hidden, dependency and denied folders, and
   pairs each repository with the settings the config gives it.
2. `repo.Build` reads git facts, resolves the record's files (conventional names, then the
   config, then the repository's own `noether` block in `AGENTS.md`), reads them through
   the guard, assesses upkeep, and derives notices. Problems become notices, never failures.
3. `repo.BuildAll` builds up to eight repositories at once, most recently active first.
4. The server adds rendered HTML and editor links. `/api/repos` sends summaries,
   `/api/repos/{id}` one full report; the page is `internal/server/assets`, embedded.

## Timing rules

- A file's "updated" time is its modification time while it has uncommitted edits,
  otherwise the date of its latest commit, so a fresh clone does not look newly edited.
- Work is "after the state file" when the newest commit or uncommitted change outside it
  is more than an hour newer than the state file.
- An upkeep log line carries a date only, so commits are counted from the next day.

## Page

Plain HTML, CSS and JavaScript, no build step. Hash routes (`#/<id>`) select a repository;
the list reloads on Refresh and when the tab becomes visible again. Light and dark follow
the system. Below 761 px the list and the detail take turns.
