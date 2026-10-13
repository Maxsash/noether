# The record

A repository's record is a few plain markdown files that say where it stands. People and
AI agents write them as they work; noether only reads them. No part needs AI to read:
each is free text shown as written, a checkbox, or a dated line.

## Three parts

| Part  | Holds                                                                          | Format                                                       |
| ----- | ------------------------------------------------------------------------------ | ------------------------------------------------------------ |
| State | Where things stand now: what was done, what is next, what is open or unverified | Free markdown, rewritten rather than appended, within a line budget (100 unless set) |
| Tasks | Work still to do                                                               | `- [ ] open` and `- [x] done` lines under headings           |
| Log   | When each upkeep job last ran and what it found                                | One line per run: `- 2026-10-13 review: three fixes; checks pass` |

Everything else, decisions and architecture included, is an ordinary doc. Noether lists
every markdown file at the top level and in `docs/` with its length and when it changed.

## Where noether looks

Without settings, noether picks files by name at the top level and in `docs/`, ignoring
case, hyphens and underscores:

- **State:** `handoff.md`, else `status.md`
- **Tasks:** every `todo.md`, `follow-ups.md`, `workplan.md`, `tasks.md` and `backlog.md`
- **Log:** `upkeep.md`, else `log.md`

A repository names its own files in a code block tagged `noether` in its `AGENTS.md`:

```noether
state = "docs/handoff.md"
tasks = ["docs/todo.md"]
log = "docs/log.md"
state_lines = 100

[upkeep]
review = { every = "30d", commits = 40 }
testing = { every = "2w" }
```

The same keys can sit in the owner's config, as `[defaults]` for every repository or a
`[[repo]]` entry for one that should not carry the block. The repository's own block wins,
then the `[[repo]]` entry, then the defaults, then the names above.

## Upkeep

An upkeep job is any recurring piece of care: a review, a testing pass, removing comments
that explain nothing, finding repetition, pruning docs, clearing leftovers. Names are free;
the log line and the schedule only have to agree.

A job is due once `every` days or weeks have passed since its last log line, or once
`commits` commits have landed after that day. A job with no log line shows as never run.

## What noether flags without being told

- no git remote, so the repository exists on one disk
- commits not pushed
- uncommitted files, and a warning once the oldest change is a week old
- extra worktrees, often left by an interrupted session
- work changed more than an hour after the state file last did
- a state file over its line budget
- a due upkeep job
- a record file that is named but missing, and a noether block that does not parse

## The section to add to AGENTS.md

Copy this into a repository's `AGENTS.md` and adjust the paths. It is the instruction any
agent, or person, follows to keep the record true.

````markdown
## Record

This repository keeps a record of where it stands, for people, agents and noether.
Keep it true, and delete what stops being true.

- **State, `docs/handoff.md`.** Rewrite it after every step: what was done, what is next,
  open questions, and what was and was not verified. Stay within 100 lines. It describes
  now; history belongs to git and the log.
- **Tasks, `docs/todo.md`.** Every piece of pending work is a `- [ ]` line under a
  heading. Tick it when done and delete done items once they stop being useful. Never keep
  a to-do only in the state file or in a conversation.
- **Log, `docs/log.md`.** After an upkeep job, add one line at the top:
  `- YYYY-MM-DD job: what it found and what changed`. Log a job even when it found nothing.
- **One home per fact.** A result is written once, in the log or the state, never both.
  Decisions go in `docs/decisions.md`.

```noether
state = "docs/handoff.md"
tasks = ["docs/todo.md"]
log = "docs/log.md"

[upkeep]
review = { every = "30d", commits = 40 }
testing = { every = "30d", commits = 40 }
```
````
