{{/* See docs/prompts/prime.md */}}
{{ if .WorktreeDirty -}}
> [!WARNING]
> This checkout has uncommitted changes. Before editing, identify whether they are yours, user work, or another agent's progress. Ask before touching unclear changes.

{{ end -}}
# Beadwork

Beadwork persists plans, progress, and decisions to git so they survive. Compaction erases context.

Issues live on the `beadwork` branch. IDs: `{{ .Prefix }}-XYZ`. Status: open → in_progress → closed / deferred. Priority: P0-P4 (default P2). Epics have children (`--parent`) and deps (`bw dep add <blocker> blocks <blocked>`). `bw ready` feeds you the next unblocked step, so compaction can't erase your progress.

Due dates (`bw update <id> --due <date>`) are deadlines that do not change status. Deferred issues (`bw defer`) are hidden from `bw ready`; due issues are not. Overdue items appear in `bw list --overdue`. Date expressions: `YYYY-MM-DD`, `tomorrow`, `2 weeks`, `next monday`, `in 15 minutes`, `tomorrow at 2pm`, `3pm`, or full RFC3339.

## Where You Are

Current checkout: {{ if .Git.IsWorktree }}worktree on{{ else }}branch{{ end }} `{{ .Git.Branch }}`{{ if .Git.Dirty }} · **uncommitted changes**{{ else }} · clean{{ end }} · last commit: `{{ .Git.LastCommit }}`

## Default: Work Here

Use the current branch and checkout the user put you in. That can be `main`. Do not create a branch, PR, or worktree unless the user asks, repo config requires it, or you need isolation and the user agrees.

Pick the lightest durable workflow that fits:

- **Quick fix**: make the scoped change in the current checkout. No ticket needed unless the user asks.
- **Tracked task**: use an existing issue or create one (`bw create "Title" --description "..." -t task`), then `bw start <id>` and work on the current branch.
- **Multi-step / swarm**: create an epic with child tasks and dependencies. Multiple agents may collaborate on this same branch by claiming separate children and leaving comments.
- **Branch / PR / worktree**: optional delivery modes. Use them only when requested or configured; `bw start` will show PR-specific landing hints when this repo is configured for PR review.

## Plans Are Scratch — Tickets Survive

Plan however you want. Your plan is useful for thinking, but it lives in context and dies at compaction. Before you execute a multi-step plan, materialize it into beadwork:

1. Create an epic: `bw create "Epic title" -t epic --description "..."`
2. Create a child task for each step: `bw create "Step title" --parent <epic> --description "..."`
3. Wire dependencies: `bw dep add <blocker> blocks <blocked>`
4. Include a mermaid sequencing graph in the plan so the dependency structure is visible:
   ```mermaid
   graph LR
       1 --> 2
       1 --> 3
       2 --> 4
       3 --> 4
   ```

## Workflow

1. **Orient**: check the current branch, dirty state, WIP, and ready queue. Do not overwrite changes you do not own.
2. **Claim**: for tracked work, run `bw start <id>` before editing. For quick fixes, proceed without ceremony.
3. **Work**: keep edits scoped to the current task. Use `bw comment <id> "..."` for breadcrumbs, handoffs, and swarm coordination.
4. **Land**: commit scoped changes referencing the ticket ID when there is one, then `bw close <id>`.
5. **Sync**: run `bw sync` so the durable beadwork state reaches collaborators and future sessions.

What isn't committed, closed, and synced will pollute the next session.

## Delegation

Sub-agents do not inherit your context. Current-branch swarms are valid: give each agent a child ticket or non-overlapping file scope, and have them coordinate through comments. Use separate branches/worktrees only when the user wants isolation or concurrent edits would collide.

Include the workflow in the handoff:

```
Run `bw start <id>`. Make the scoped change on the current branch. Commit referencing <id>. Comment with handoff notes. Run `bw close <id>`.
```

Verify the work landed after the agent returns.

{{ if .ExpiredDeferrals -}}
## Reminders

The following items were deferred and are now due for attention:

{{ .ExpiredDeferrals }}

{{ end -}}
{{ if gt .OverdueCount 0 -}}
> **{{ .OverdueCount }} items are past due.** Run `bw list --overdue` for details.

{{ end -}}
## Work In Progress

{{ bw "list" "--status" "in_progress" }}

## Currently available work:

{{ bw "ready" "--no-context" }}

`bw comment <id> "..."` = breadcrumbs. `bw --help` for everything. `--json` gets you raw data.
