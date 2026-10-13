# To do

## Owner decisions

- [ ] Push the first commit to `Maxsash/noether` (created, public and empty); steps are in
      the handoff.
- [ ] Choose upkeep jobs and cadences, then enable `[defaults.upkeep]` in
      `~/.config/noether/config.toml`.

## Adopt the record

- [ ] web: add the record section to `AGENTS.md`, move open follow-ups into a task file,
      start `docs/log.md`, slim the handoff to state only. Its handoff still lists the
      "test workbench" as discussion only; that idea now lives here.
- [ ] family-tree: add a state file; tasks stay in `Docs/WorkPlan.md`.
- [ ] tenant-manager, then wedding, then the rest.

## Noether

- [ ] `noether init`: print the record section for a repository with its detected files.
- [ ] Hide chosen repositories (dormant ones) without denying them.
- [ ] Let a repository switch off an upkeep job its defaults give it.
- [ ] Upkeep playbooks as Agent Skills, each ending with its log line.
- [ ] Release builds and Homebrew tap, shipping the TOML and goldmark licence notices with
      binaries as their MIT terms require; CI running vet, tests and staticcheck.
- [ ] Decide the minimum Go version (`go.mod` says 1.27.2).
- [ ] Later: catalogue of checks (performance, contrast, fonts, legal) and a page of
      Claude plugins kept only after a trial.
