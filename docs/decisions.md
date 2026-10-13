# Decisions

Newest first.

## 2026-10-13 — MIT licence, held by Maxsash Studio

The aim is for other developers to use and improve Noether, so the most widely understood
permissive licence fits; both dependencies are MIT too. Apache-2.0's patent grant matters
once companies with patents contribute, and MIT code can move under it later; GPL would
guard against closed forks at the cost of users and contributors. Maxsash Studio holds the
copyright, as in tenant-manager's licence.

## 2026-10-13 — Named Noether

After Emmy Noether, the mathematician of invariants: the tool keeps each repository's record
true while its code changes. Chosen by the owner over sea names and other mathematical
ones. Other projects share the name (a deep-learning framework, a robot path planner), none
a developer tool of this kind; the Homebrew name is free.

## 2026-10-13 — Build a viewer, reuse everything else

Research found most of the surrounding pieces already exist: scheduled agent runs (GitHub
Agentic Workflows, Claude Code routines, Codex automations), ready prompts (githubnext/
agentics), and checkers (Unlighthouse, MegaLinter, licence and secret scanners). Nothing
shows a person's own repositories' records, history, tasks and upkeep in one place, or
compares what the docs say with the machine's actual state. Noether is that viewer.
Upkeep playbooks and a checks catalogue may follow; a scheduler or checking engine will not.

## 2026-10-13 — Plain files and git are the only inputs

The status must not depend on AI. Records are plain markdown with three machine-readable
shapes (free text, checkboxes, dated lines). Noether never writes to a repository; agents
and people do, following the section in [record.md](record.md).

## 2026-10-13 — Settings live in AGENTS.md, with fallbacks

A repository's settings sit in a `noether` code block in its `AGENTS.md`, the file every
major agent already reads, so the instructions and the settings stay together and no new
file is added. Repositories that cannot carry it get a `[[repo]]` entry in the owner's
config; repositories with neither still work through conventional names.

## 2026-10-13 — Lessons taken from the maxsash web repository

That repository's practice is the starting point, not the standard. Its handoff mixed state,
verification results and owner reports, and the same report sat in three docs; its docs
outgrew their own size rule. The record therefore separates state, tasks and log, asks for
one home per fact, and noether measures doc length instead of trusting it.

## 2026-10-13 — Go, two dependencies, no front-end build

One binary installs on any platform without a runtime, which suits a tool for developers on
every stack. `BurntSushi/toml` reads settings; `yuin/goldmark` renders markdown with raw
HTML removed. The page is plain HTML, CSS and JavaScript embedded in the binary.

## 2026-10-13 — Local only, read only

The server binds to loopback and checks the `Host` header against DNS rebinding. Git runs
without optional locks so reading never blocks someone committing. A deny list, which
always includes secret files, is checked before every read.
