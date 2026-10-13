# Handoff — start here

Updated 13 October 2026. Read [AGENTS.md](../AGENTS.md), then
[architecture.md](architecture.md). Pending work is in [todo.md](todo.md).

## Current state

New repository, nothing committed, MIT licensed (Maxsash Studio). `Maxsash/noether` exists
on GitHub, public and empty; no remote added locally yet. Named Noether on 13 October (it was
"chartroom" while unnamed); the module path is `github.com/maxsash/noether`.

- `noether serve` shows a local page: every git repository under the configured roots,
  each with what needs attention, its state file, open and done tasks, upkeep jobs,
  history (commits and upkeep runs) and docs with lengths and dates.
- `noether status` prints the same summary in the terminal; `--json` for scripts.
- [record.md](record.md) defines the record and holds the section other repositories paste
  into their `AGENTS.md`. This repository follows it.
- The owner's config is at `~/.config/noether/config.toml`, outside this repository:
  roots `~/WithAIAssistant/development`, the workspace's forbidden paths as the deny list,
  family-tree's tasks mapped to `Docs/WorkPlan.md`. Upkeep defaults are commented out until
  the owner chooses jobs and cadences.
- Go 1.27.2 was installed with Homebrew for this project.

## Next

1. Owner: push the first commit (below), then choose upkeep cadences.
2. Adopt the record in web and family-tree. These are changes inside those repositories and
   need the owner's go-ahead there.
3. `noether init`, and hiding dormant repositories.

## Publishing (owner)

The GitHub repository is created, public and empty. In this folder: `git add -A`, commit
with the message below, `git remote add origin https://github.com/Maxsash/noether.git`,
and `git push -u origin main`. Afterwards GitHub should show "MIT license" beside the
repository; then update the "no remote" row in `../../AGENTS.md`.

## Verified (13 October 2026)

- `gofmt -l .` empty; `go vet`, staticcheck (no findings) and `go test ./...` pass. Unit
  tests cover the guard, record readers, upkeep, git output parsers and attention rules;
  integration tests build reports from temporary git repositories and exercise the HTTP
  API, including refusing other hosts, removing raw HTML and `javascript:` links, and
  editor links.
- Run against the owner's 12 repositories in about 1.5 s. web, family-tree and yash-web
  reports were checked against git by hand.
- Headless Chrome captures inspected: desktop light and dark, the phone list, and the
  narrow detail at 500 px, which is headless Chrome's minimum width. They found three bugs,
  all fixed: an empty problems list sent as `null` broke the page, a stray "null" in the
  list, and narrow-screen overflow. Wrapped task lines are now joined, and a repository
  without commits no longer shows a git error.
- Not verified: a true 390 px layout, Windows or Linux, editor links clicked in a browser,
  keyboard-only use in a real browser, a screen reader.

First run's real findings, reported to the owner and not acted on: web has 6 unpushed
commits and an extra worktree at `/private/tmp/maxsash-atmosphere-review` from another
session; yash-web has 1 unpushed commit and a change uncommitted for 153 days;
BirthdayAnniversaryWidgets has 8 files uncommitted for 142 days.

## Open items

- web's handoff still says the "test workbench" remains discussion only. It was not edited
  because another session was working in web; see [todo.md](todo.md).
- Nothing left in the scratch folder; the review server was stopped.

## Suggested commit

```
feat: show every repository's record on one local page

Noether reads git state and each repository's plain-text record (state,
tasks, upkeep log) and shows them on a loopback-only page, with a terminal
summary. Read-only, no network, and a deny list checked before every read.
MIT licensed.

Verified: gofmt, go vet, staticcheck and go test ./... (unit tests and
temporary-repository integration tests); run against 12 local
repositories; light, dark and narrow captures inspected.
```
