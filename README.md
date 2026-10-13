# Noether

See where every repository stands: what changed, what is left to do, and which upkeep is
due. Noether reads plain files and git. It needs no AI, no account and no network, and
nothing it reads leaves your machine.

Working with AI agents, small things slip: a review that never happened, a handoff that
stopped matching the code, a worktree left by an interrupted session, a commit never
pushed. Each session sees only its own step. Noether shows all of it, for every
repository, on one page.

The name is Emmy Noether's, the mathematician of invariants: what stays true while
everything else changes. Keeping each repository's record true while its code moves is
this tool's whole job.

## How it works

- **The record.** Each repository keeps three plain-text parts: its state, its tasks and an
  upkeep log. See [docs/record.md](docs/record.md), which also has the section to paste into
  a repository's `AGENTS.md` so agents keep the record as they work.
- **The page.** `noether serve` reads every git repository under the folders you name and
  shows them on a page only your machine can open.
- **No setup needed to start.** Any git repository shows its history, uncommitted and
  unpushed work, leftover worktrees and docs. The record adds state, tasks and upkeep.

## Try it

From a clone, with Go installed:

```sh
go run ./cmd/noether serve --open   # the page, at http://127.0.0.1:4747
go run ./cmd/noether status         # a summary in the terminal; --json for scripts
```

Without a config file, noether reads the repositories under the current folder.

## Config

`~/.config/noether/config.toml`:

```toml
roots = ["~/code"]                  # folders searched for git repositories
depth = 4                           # how many folders down to look
deny = ["~/code/private", "**/notes"]   # never opened; .env files and keys always are
editor = "vscode://file{path}:{line}"   # links that open files; "none" turns them off

[defaults]                          # record settings every repository starts with
state_lines = 100

[defaults.upkeep]
review = { every = "30d", commits = 40 }

[[repo]]                            # one repository's settings, kept out of the repository
path = "~/code/app"
tasks = ["Docs/WorkPlan.md"]
```

## Guarantees

- Reads only. Noether runs read-only git commands, without git's optional locks, and
  never writes to a repository.
- Local only. The page listens on 127.0.0.1 and refuses requests addressed to any other
  host.
- Denied paths are never opened. Markdown is shown with raw HTML removed.

## Status

Early. MIT licensed; see [LICENSE](LICENSE).
