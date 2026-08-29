# Current-branch-first agent workflow

This document captures the workflow review around unwinding Beadwork's old
"pushy worktrees only" assumption. The desired model is not "never use
worktrees". It is: **work in the checkout the user gave you by default, and use
branches, PRs, or worktrees only when the user, repo config, or collision risk
calls for them.**

## Desired model

Beadwork has two separate git concerns that should not be conflated:

1. **Issue storage** stays on the dedicated `beadwork` branch. This is the
   product's durable state model and remains correct.
2. **Code changes** happen in the user's current checkout by default. That
   checkout might be `main`, a feature branch, a linked worktree, a clone on
   another machine, or a branch shared by several agents.

The default agent posture should be:

- Use the current branch/check-out the user put you in.
- Do not create a branch, PR, or worktree unless asked, configured, or needed
  for isolation with user agreement.
- Treat dirty state as an ownership/preservation problem, not as an automatic
  worktree gate.
- Same-branch swarms are valid: split work into epics/child tickets, claim
  separate children or non-overlapping file scopes, and coordinate with
  `bw comment`.
- Landing is part of doing the work: commit scoped changes, close the relevant
  ticket, and `bw sync` so durable state reaches other agents/machines.

## Why this changed

The previous prompt model overfit to a branch/PR/worktree workflow. That was
reasonable for isolated delivery, but it fought the real workflow this repo is
trying to support:

- multiple machines,
- multiple active branches,
- agent swarms collaborating on one epic,
- and sometimes simply working on `main` because that is the branch the user
  chose.

The important invariant is not "one ticket, one worktree." The invariant is:
**do not lose or overwrite work you do not own, and leave durable breadcrumbs.**

## Surfaces reviewed

### Runtime prompts

#### `prompts/prime.md`

This is the most important surface because `bw prime` is the canonical session
bootloader for agents. It previously encoded the strongest worktree-first
language: dirty checkout warnings, branch/PR delivery as the clean path, and
"one ticket, one worktree" style guidance.

It now says:

- current checkout is the default place to work,
- `main` is allowed if that is where the user placed the agent,
- quick fixes may not need tickets,
- tracked tasks use `bw start` and work on the current branch,
- multi-step/swarm work uses epics and child tickets,
- branch/PR/worktree are optional delivery modes,
- dirty state requires ownership checks before editing,
- and sub-agent handoffs must include workflow steps because sub-agents do not
  inherit context.

#### `prompts/start.md`

This prompt is the point-of-action briefing for `bw start`, and may be the only
Beadwork context a delegated agent sees. It previously leaked worktree-specific
PR wording such as pushing "the branch for this worktree."

It now defaults to current-branch work. PR hints remain conditional on
`workflow.review=pr`, and those hints say to push/open/update the current branch
rather than assuming a worktree.

Tasks/bugs still get explicit landing guidance:

- commit only ticket-scoped changes,
- reference the issue ID,
- close the ticket,
- and run `bw sync`.

Epics remain organizational containers whose actual work should happen through
children.

### Prompt theory docs

The files under `docs/prompts/` matter because they are the theory and
maintenance notes that would otherwise pull the prompts back toward the old
model.

Reviewed and aligned:

- `docs/prompts/prompts.md`
- `docs/prompts/prime.md`
- `docs/prompts/start.md`

They now describe current-branch work as the baseline, with branch/PR/worktree
isolation as an optional mode. They also keep the validation philosophy: prompt
changes should be tested with real agent behavior, not just rendered and read.

### Agent onboarding prompts

`prompts/agents.md` and `prompts/onboard.md` are intentionally minimal
bootloaders. They should continue pointing agents at `bw prime` instead of
copying workflow policy inline. That prevents drift.

No major rewrite was needed there as long as `bw prime` remains the source of
truth.

### Public docs and older history

`README.md`, `AGENTS.md`, `docs/design.md`, and `docs/migration.md` mostly
describe the durable issue-storage model: issues live on the `beadwork` branch,
and the user's working tree/index are not used for Beadwork's internal state.
That remains true and should not be rewritten as if the `beadwork` branch is the
same thing as the branch used for code changes.

`CHANGELOG.md` contains older release notes that accurately describe older
prompt behavior. Those historical entries should stay. A new `Unreleased`
section records the current-branch-first shift.

## Code surfaces reviewed

### `cmd/bw/prime.go`

`bw prime` gathers repo prefix, overdue/deferral information, dirty state, and
`GitContext`, then renders `prompts/prime.md`.

The runtime code did not force worktrees. The old behavior was primarily in the
rendered prompt text.

### `cmd/bw/start.go`

`bw start` reads repo config and passes `workflow.review` into the start prompt.
This already made PR guidance config-driven rather than hardcoded.

The fix was to make the prompt text respect that model: no PR/worktree language
by default, conditional PR hints only when configured.

### `internal/repo/repo.go`

One real code issue surfaced: `GetGitContext().IsWorktree` could misdetect when
called from a subdirectory inside a linked worktree, because it checked only the
current directory's `.git` entry.

The implementation now resolves the git top-level first and checks `.git` there,
so linked worktrees are detected correctly from nested directories.

### `cmd/bw/emit.go` and `bw ready`

`bw ready` also prints git context. That is still useful: agents should know
where they are before choosing work. The important distinction is that showing
"branch" or "worktree" is informational, not a policy requiring worktree use.

## Tests added or updated

Regression coverage now protects the new workflow assumptions:

- `cmd/bw/prime_test.go`
  - asserts the prime output includes the current-branch default,
  - rejects old worktree-first phrases such as "one ticket, one worktree" and
    "This is the only way to land cleanly."

- `cmd/bw/start_test.go`
  - asserts default `bw start` output does not mention opening a PR,
  - asserts default output does not mention worktrees,
  - keeps PR-mode behavior under `workflow.review=pr`.

- `internal/repo/git_context_test.go`
  - covers normal main-working-tree git context,
  - creates a real linked worktree,
  - calls `repo.FindRepoAt` from a nested subdirectory,
  - verifies `GetGitContext()` reports the linked worktree and branch correctly.

The full suite passes with `go test ./...`.

## Remaining optional follow-ups

These are not required for the current behavior shift, but are worth considering
if the workflow model grows:

1. **Add `workflow.mode` only if needed.**
   Today `workflow.review=pr` is enough to gate PR hints. If Beadwork needs a
   stronger repo-level policy later, consider a config like
   `workflow.mode=current|pr|worktree`. Do not add it just to restate the
   default.

2. **Rename template data eventually.**
   `WorktreeDirty` is now semantically "checkout dirty." Renaming it would make
   the code match the new language, but it is internal template data and not a
   behavioral blocker.

3. **Add behavioral prompt trials.**
   The prompt docs recommend testing with real interactive agents. The unit
   tests protect text regressions, but live trials are still useful for checking
   whether agents actually avoid creating branches/worktrees by default.

4. **Public docs clarification.**
   If users keep confusing the `beadwork` branch with their code branch, add a
   short README note: Beadwork stores issue data on its own branch, while agents
   normally edit code in the current checkout.

## Bottom line

The repo no longer needs to teach "worktree first" as the default agent model.
The durable core stays git-native on the `beadwork` branch, while code work is
current-checkout-first. Worktrees, branches, and PRs remain supported delivery
and isolation tools, but they are no longer prerequisites for tracked work or
agent collaboration.
