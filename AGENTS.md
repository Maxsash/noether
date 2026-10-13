# Noether — guide for agents

A Go command that reads the record and git state of many repositories and shows them on a
local page. Read [docs/handoff.md](docs/handoff.md) first, then
[docs/architecture.md](docs/architecture.md) and [docs/decisions.md](docs/decisions.md).

## What must stay true

- **Read-only.** Noether never writes to, commits in or changes a repository it reports
  on. Git commands are read-only and run with `GIT_OPTIONAL_LOCKS=0`.
- **No AI and no network to show status.** Every fact comes from files and git.
- **Local only.** The server binds to loopback and refuses other `Host` headers.
- **Every file read goes through the guard.** Secret files and denied paths are never opened.
- **Not tied to any repository.** Conventions live in `record.Detect` and the config,
  never in code that names a particular repository.

## Documentation is king

Assume the session can end at any moment. Anything that exists only in the conversation is
lost.

- [docs/handoff.md](docs/handoff.md) is the start-here file. After every step, before
  reporting back, rewrite it: current state, what is next, and what was and was not
  verified. Keep it within 100 lines.
- Pending work lives in [docs/todo.md](docs/todo.md), decisions in
  [docs/decisions.md](docs/decisions.md), upkeep runs in [docs/log.md](docs/log.md).
  Write each fact once. Delete what stops being true instead of appending.

## Clean up immediately

Delete what a step created as soon as it is no longer needed: build output, scratch
scripts, screenshots, browser profiles, logs. Stop servers you started. Anything that must
stay goes in the handoff with when to delete it.

## Git

Never commit or push unless the owner asks in the current session. Leave work uncommitted
for review and suggest a conventional commit message in the handoff.

## Code

This repository is public. Write code that is a pleasure to read.

- No comments unless the reason is not visible in the code. A comment explains why.
- No repetition, one responsibility per package. Pure logic (record and git parsers,
  upkeep, attention) stays apart from file, process and HTTP code, and is tested directly.
- Delete dead code. No new dependency without a reason in decisions.md.
- The page is plain HTML, CSS and JavaScript with no build step. Text is set as text; only
  markdown the server rendered with raw HTML removed is set as HTML.

## Verify

```sh
gofmt -l .                     # prints nothing
go vet ./...
go test ./...
go run honnef.co/go/tools/cmd/staticcheck@latest ./...
```

For page changes, run `go run ./cmd/noether serve` and look at a wide and a narrow
window in light and dark. Headless Chrome lays pages out at least 500 px wide whatever
window size it is given.

## Record

```noether
state = "docs/handoff.md"
tasks = ["docs/todo.md"]
log = "docs/log.md"

[upkeep]
review = { every = "30d", commits = 40 }
testing = { every = "30d", commits = 40 }
```
