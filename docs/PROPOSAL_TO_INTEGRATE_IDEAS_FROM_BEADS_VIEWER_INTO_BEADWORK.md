# Proposal: Beadwork as a Git-Native Swarm Control Plane

**Status:** Expanded draft proposal  
**Audience:** beadwork CLI maintainers, pi-beadwork-extension maintainers, agent-workflow designers  
**Source research:** `/tmp/beads_viewer-research`, current `beadwork` repo, and `@solvedbydev/pi-beadwork-extension` architecture  
**Constraint:** Planning document only. It intentionally does not implement code or create tickets.

---

## 1. Executive summary

The first proposal was directionally correct but too small: it treated `beads_viewer` mainly as a graph triage engine to adapt into `bw triage` and `bw plan`. That is useful, but it is not the real opportunity.

The more ambitious opportunity is this:

> **Beadwork can become a git-native control plane for agent swarms: a local, durable, self-improving work graph that fuses issues, dependencies, recent commits, file ownership, labels/domains, worker state, validation results, and prompt context into one operational system.**

`beads_viewer` is valuable because it proves that agents should not parse raw task lists. They need deterministic, graph-aware, low-context, machine-readable decision support. But beadwork has something `beads_viewer` does not: mutation authority, a typed intent log, prompt injection via `bw prime`, claim-time briefs via `bw start`, and pi-managed workers with worktrees, validation, review, and landing.

That combination enables a product category beyond both tools:

- not a standalone viewer,
- not a local issue tracker,
- not merely a task graph,
- but an **agent operating layer** that decides what should happen next, assigns safe parallel work, briefs workers, predicts conflicts, validates plans, learns from outcomes, and preserves all of it in git.

The proposed north star:

> **Every agent session starts with a graph-aware control-plane brief; every worker is launched from an explainable execution plan; every close/land event updates causal project memory; every future recommendation improves because beadwork knows what work changed which code.**

The first implementation should still be incremental, but the design should point toward a disruptive system:

1. **Swarm Intelligence Core** — `bw triage`, `bw plan`, `bw partition`, `bw validate-plan`.
2. **Causal Work Memory** — typed event stream from beadwork intents plus issue↔commit↔file correlation.
3. **Context Compiler** — role-specific `bw prime --for planner|worker|reviewer|orchestrator` and `bw brief` packs.
4. **Conflict-Aware Orchestration** — pi `/bw run` schedules by dependency tracks and predicted file overlap.
5. **Adaptive Feedback Loop** — recommendation outcomes, worker results, remediation, validation failures, and operator feedback feed future scoring.
6. **Graph Health / Drift / Alerts** — proactive warnings when the work graph becomes unsafe or low-throughput.
7. **Recipes and Query Layer** — saved operational views like `high-impact`, `recently-drifted`, `safe-parallel`, `release-cut`, and `needs-human`.

The result would make beadwork more compelling than a `beads_rust` + `beads_viewer` setup for the user’s real workflow: many high-priority workers, recent-commit-aware planning, graph setup, and Jeffery Emanuel-style agent flywheels — but native to beadwork’s git branch model and pi’s delegation machinery.

---

## 2. Product thesis

### 2.1 The shallow interpretation

A shallow reading says:

> “`beads_viewer` has graph ranking. Add graph ranking to beadwork.”

That is worth doing, but it is not enough. It would reproduce a subset of `bv --robot-triage` and still leave beadwork as a CLI that answers isolated questions.

### 2.2 The deeper interpretation

The deeper reading is:

> **AI-agent work management is not a queue problem. It is a control problem.**

A control system needs to know:

- the current work graph,
- the desired outcome,
- the active workers,
- the dependency constraints,
- the code areas likely to be touched,
- the recent changes,
- the validation state,
- the landing order,
- and the feedback from previous attempts.

Then it needs to decide:

- what to do next,
- what can happen in parallel,
- what must not happen yet,
- where a human should intervene,
- what context each worker needs,
- how to update the plan when reality changes.

`beads_viewer` contributes the graph and robot-output instincts. Beadwork can turn those instincts into a full control loop because beadwork owns the durable event stream and pi owns the worker lifecycle.

### 2.3 The bold bet

The bold bet is that beadwork can become the **local-first equivalent of an agent engineering manager**:

- It ranks work.
- It decomposes and validates plans.
- It schedules parallel workers.
- It predicts conflicts.
- It creates compact worker briefs.
- It watches graph drift.
- It remembers why decisions were made.
- It learns which recommendations actually produced successful landed code.

All without a SaaS backend, database server, external API key, or centralized task service.

---

## 3. Why beadwork is uniquely positioned

### 3.1 Beadwork already has a typed intent log

`beads_viewer` reconstructs history from `.beads/beads.jsonl` diffs. That is clever, but beadwork has a cleaner primitive: each mutation commits an intent to the `beadwork` branch.

Examples include create, update, start, close, reopen, link, unlink, label, defer, undefer, comment, attach, config, and sync replay intents.

This means beadwork can eventually expose an event stream like:

```json
{
  "event_type": "issue_started",
  "issue_id": "bw-42",
  "actor": "Alice / claude-opus",
  "commit": "abc1234",
  "timestamp": "2026-04-29T05:30:00Z",
  "intent": "start bw-42 assignee=Alice"
}
```

That stream can power:

- lifecycle history,
- velocity,
- stale WIP detection,
- author/agent attribution,
- close-to-unblock analysis,
- worker outcome summaries,
- correlation with code commits,
- compaction-resistant handoffs.

This is more structurally powerful than `beads_viewer` because the event model is not reverse-engineered from JSONL line diffs; it is beadwork’s native history.

### 3.2 `bw prime` is not a help command; it is a control-plane injection point

`bw prime` already gives dynamic, live workflow context to agents. This is a huge differentiator.

Most tools make agents ask the right command. Beadwork can instead make the right context appear at session start:

- top graph risks,
- highest-impact ready work,
- active workers,
- stale claims,
- recent project movement,
- unlanded worktrees,
- blocked P0/P1 issues,
- changed recommendation since last session,
- exact next commands.

A standalone viewer can show intelligence. `bw prime` can **shape agent behavior before the first tool call**.

### 3.3 `bw start` is a natural worker brief compiler

`bw start` currently claims work and prints useful issue context. With graph intelligence, it can become a just-in-time work packet:

- why this issue is selected,
- where it sits in the dependency graph,
- what it unblocks,
- what files are likely relevant,
- which prior commits touched related issues,
- active workers to avoid,
- known validation expectations,
- previous failed attempts,
- close/landing checklist.

This turns `start` into a **context compiler**, not merely a status transition.

### 3.4 Pi already has the worker runtime

The pi extension already supports delegated workers, ticket worktrees, tmux supervision, validation, review, deferred landing, cleanup, and bounded epic runs.

The missing piece is not orchestration mechanics. The missing piece is **orchestration intelligence**.

`/bw run` can become much more than “launch up to N ready tickets”:

- launch across independent tracks,
- avoid file-overlap conflicts,
- prioritize critical path blockers,
- run speculative downstream work safely,
- delay landing based on graph risk,
- rebase/refresh workers proactively,
- emit operator notices when the graph changes.

---

## 4. What beads_viewer contributes beyond simple triage

The source material in `/tmp/beads_viewer-research` contains several ideas that are more important than the TUI.

### 4.1 Robot-first contracts

`beads_viewer` repeatedly emphasizes agent-safe robot commands, including `--robot-triage`, `--robot-next`, `--robot-plan`, `--robot-insights`, `--robot-alerts`, `--robot-history`, `--robot-diff`, `--robot-suggest`, `--robot-graph`, label health/flow/attention, file relations, and more.

The key insight is not the exact flags. The insight is that agents need **protocols**, not prose:

- versioned output,
- `data_hash`,
- metric status,
- generated timestamps,
- command hints,
- usage hints,
- stable schemas,
- partial-data warnings.

Beadwork should adopt this as a contract style across new machine-facing commands.

### 4.2 Graph metrics and progressive analysis

`beads_viewer` uses a two-phase model:

- fast topology and degree metrics first,
- slower centrality/cycle metrics with timeout/status later.

For beadwork, the lesson is architectural:

- fast commands must stay fast,
- analysis status must be explicit,
- missing metrics must be represented as `skipped`, `timeout`, or `approx`,
- agents must be able to degrade gracefully.

### 4.3 Execution tracks

`bv --robot-plan` groups work into tracks. This is directly relevant to pi workers.

In beadwork, tracks should be more powerful because they can combine:

- dependency components,
- parent/epic scope,
- labels/domains,
- likely file overlap,
- current worker claims,
- landing state.

### 4.4 Label overlays

The `labels-view-feature-plan.md` treats labels as graph overlays rather than filters. That is a major idea.

In beadwork, labels can become **domains**:

- `frontend`, `backend`, `database`, `docs`, `testing`,
- each with health, flow, velocity, staleness, active workers, blocked downstream impact,
- each usable for worker routing.

This allows pi to answer questions like:

- “Which domain is starving the release?”
- “Which label needs the next specialist worker?”
- “Which label is producing blockers for the others?”

### 4.5 History and code correlation

The `bead-history-feature-plan.md` identifies a powerful correlation loop:

- task lifecycle event → git commit → code files → future task/file relevance.

Beadwork can do this better because issue history already lives in git, and pi worker landing can annotate the exact branch/head/validation result.

This unlocks:

- `bw history <issue>` as causal timeline,
- `bw file-beads <path>`,
- `bw related <issue>` based on co-changed files,
- orphan commit detection,
- conflict prediction before worker launch,
- better worker briefs.

### 4.6 Agent brief bundles

`beads_viewer` includes the idea of agent briefs. Beadwork should elevate this into a first-class primitive:

```bash
bw brief --for planner
bw brief --for worker bw-42
bw brief --for reviewer bw-42
bw brief --for orchestrator --scope bw-epic
```

The brief is not just issue text. It is a token-budgeted bundle of graph, history, file, worker, and validation context.

### 4.7 Alerts and drift

`beads_viewer` has robot alerts, drift, diff, baselines, and proactive warnings.

Beadwork’s git-native model makes this especially compelling:

```bash
bw drift --since HEAD~20 --json
bw alerts --scope bw-epic --json
bw diff-graph --since yesterday --json
```

This can tell an agent what changed since the last session or since an epic plan was adopted.

### 4.8 Recipes and saved operational views

`beads_viewer` has recipes like actionable, high-impact, stale, blocked, bottlenecks, triage.

Beadwork should support repo-native recipes:

```yaml
# .beadwork/recipes/release-cut.yaml, or stored on the beadwork branch
name: release-cut
scope: bw-release
filters:
  status: [open, in_progress]
  labels: [release]
ranking:
  - overdue
  - blocks
  - priority
warnings:
  - cycles
  - stale-in-progress
  - orphan-commits
```

Then:

```bash
bw triage --recipe release-cut --json
bw plan --recipe safe-parallel --workers 6 --json
```

This turns workflows into reusable, inspectable operational policy.

### 4.9 Semantic and hybrid search

The semantic-search plan is intentionally optional, but the hybrid ranking model matters:

- text relevance,
- graph impact,
- priority,
- recency,
- status,
- labels,
- code correlation.

Beadwork should eventually add local-first `bw search` that works across title, description, comments, history, changed files, and related issues.

This matters because agents often need to ask:

> “What prior work is relevant to auth callbacks?”

The answer should not require reading the entire issue list.

### 4.10 Agent-friendliness as a measurable product quality

The agent-friendliness report rates `beads_viewer` highly because it has structured outputs, usage hints, robot commands, actionable commands, and documentation.

Beadwork should adopt that bar and surpass it:

- `bw schema` for every robot contract,
- `bw robot-docs` for machine-readable command discovery,
- `--format json|toon|markdown`,
- embedded usage hints in JSON,
- stable examples in golden tests,
- prompt-safe compact summaries.

---

## 5. The proposed system: seven disruptive pillars

### Pillar 1 — Swarm Intelligence Core

#### What it is

A read-only analysis layer over beadwork issues, dependencies, labels, parents, due/defer state, history, worker state, and later code correlation.

Initial commands:

```bash
bw triage --json
bw next --json
bw plan --json
bw partition --agents 5 --json
bw validate-plan --scope bw-epic --json
bw alerts --json
```

#### Why it is disruptive

Most issue trackers answer “what exists?” Beadwork should answer:

- what matters now,
- what is safe to parallelize,
- what blocks the most value,
- what plan smells wrong,
- what should the next worker do,
- when should the operator intervene.

This turns a task tracker into an execution optimizer.

#### Beadwork-native design

Start with explicit blocking dependencies only for scheduling. Treat parent/child as scope and structure. Treat labels as domains. Treat issue status as worker availability. Later add file and history dimensions.

Core DTO shape:

```json
{
  "schema_version": "bw.triage.v1",
  "generated_at": "2026-04-29T06:00:00Z",
  "repo": { "root": "/repo", "branch": "main" },
  "scope": { "id": "bw-epic", "title": "Release" },
  "data_hash": "sha256:...",
  "as_of": { "beadwork_ref": "refs/heads/beadwork", "commit": "abc1234" },
  "status": {
    "topology": "computed",
    "cycles": "computed",
    "centrality": "skipped",
    "history": "partial",
    "file_correlation": "skipped"
  },
  "quick_ref": {
    "open": 42,
    "ready": 9,
    "blocked": 12,
    "in_progress": 4,
    "tracks": 5,
    "cycles": 0
  },
  "recommendations": [],
  "quick_wins": [],
  "blockers_to_clear": [],
  "tracks": [],
  "alerts": [],
  "commands": {},
  "usage_hints": []
}
```

### Pillar 2 — Causal Work Memory

#### What it is

A beadwork-native event and correlation layer:

- issue lifecycle events from beadwork branch commits,
- code commits correlated by issue IDs, co-commits, temporal author windows, and pi landing metadata,
- file/index summaries showing which issues touched which code,
- confidence-scored explanations.

Potential commands:

```bash
bw events --json
bw history bw-42 --causal --json
bw correlate --issue bw-42 --json
bw correlate --commit abc1234 --json
bw file-beads internal/issue/store.go --json
bw orphans --json
```

#### Why it is disruptive

Agents lose context. Git does not. Beadwork can convert git history into durable project memory.

With causal memory, an agent can ask:

- “What code did this issue change?”
- “Which issue explains this commit?”
- “What previous attempt failed validation?”
- “Which files does this epic usually touch?”
- “Which active workers are likely to collide?”

This goes beyond `beads_viewer`: beadwork has exact issue intent commits and pi has exact worker landing metadata.

#### Data model sketch

```go
type WorkEvent struct {
    EventID     string `json:"event_id"`
    Type        string `json:"type"`
    IssueID     string `json:"issue_id,omitempty"`
    Actor       string `json:"actor,omitempty"`
    Commit      string `json:"commit"`
    Timestamp   string `json:"timestamp"`
    Intent      string `json:"intent"`
    Source      string `json:"source"` // beadwork-intent | pi-worker | correlation
}

type CodeCorrelation struct {
    IssueID     string   `json:"issue_id"`
    Commit      string   `json:"commit"`
    Files       []string `json:"files"`
    Method      string   `json:"method"` // explicit-id | co-commit | landing | temporal-author | path-hint
    Confidence  float64  `json:"confidence"`
    Reason      string   `json:"reason"`
}
```

### Pillar 3 — Context Compiler

#### What it is

A set of role-specific, token-budgeted context products built on the same analysis model:

```bash
bw prime --for planner
bw prime --for worker --scope bw-epic
bw start bw-42 --brief graph,history,files
bw brief --for worker bw-42
bw brief --for reviewer bw-42
bw brief --for orchestrator --scope bw-epic
```

The pi extension uses these when launching workers and reviewers.

#### Why it is disruptive

Agents waste turns rediscovering context. A context compiler gives each agent exactly the context it needs for its role.

A planner needs:

- graph health,
- open questions,
- missing dependencies,
- high-risk labels,
- suggested decomposition.

A worker needs:

- issue details,
- relevant prior comments,
- likely files,
- blockers/dependents,
- validation commands,
- close criteria.

A reviewer needs:

- ticket intent,
- changed files,
- correlated prior issues,
- validation output,
- scope boundary.

An orchestrator needs:

- worker states,
- track occupancy,
- landing risks,
- next launch candidates,
- attention items.

#### Example worker brief

```markdown
# Worker brief: bw-42 Fix auth callback race

## Why this was selected
- Rank #1 in scope bw-auth; ready; P1; unblocks bw-51 and bw-52.
- On critical path for release track `auth-flow`.

## Graph context
- Direct dependents: bw-51, bw-52
- Transitive downstream: 5 open issues
- Sibling ready work: bw-44, bw-45
- Avoid overlap: worker bw-39 is active in `internal/auth/session.go`

## Code memory
- Related prior issue bw-17 touched `internal/auth/callback.go` and `cmd/bw/start.go`.
- Commit abc1234 closed bw-17 with similar validation failure: race under concurrent callback.

## Expected validation
- go test ./...
- targeted: go test ./internal/auth -run Callback

## Close guidance
- On close, verify bw-51 becomes ready.
- Suggested follow-up: `bw ready --ranked --scope bw-auth`.
```

### Pillar 4 — Conflict-Aware Orchestration

#### What it is

A smarter pi `/bw run` that schedules by graph track, file-risk, worker state, and landing risk.

Potential flow:

1. `bw plan --scope <epic> --workers N --json` returns tracks and candidates.
2. Pi asks `bw conflict-risk --candidates ... --json` or consumes risk embedded in plan.
3. Pi launches workers across tracks, avoiding likely file overlap.
4. Worker handoffs include track/file context.
5. Completion triggers impact-aware landing order.
6. Active workers receive refresh/rebase/remediation notices when graph/code state changes.

#### Why it is disruptive

Most multi-agent systems parallelize by optimism: launch N agents and hope merges work.

Beadwork can parallelize by **graph and code topology**:

- dependency independence,
- label/domain separation,
- likely file separation,
- worktree divergence,
- previous co-change clusters,
- current worker reservations.

This can materially increase throughput while reducing merge conflicts.

#### New commands/tools

```bash
bw partition --scope bw-epic --agents 6 --json
bw conflict-risk bw-42 bw-43 bw-44 --json
bw workers --json                 # optionally in CLI or pi-only tool
```

Pi tools:

- `beadwork_plan`
- `beadwork_partition`
- `beadwork_conflict_risk`
- `beadwork_worker_brief`
- `beadwork_graph_alerts`

### Pillar 5 — Speculative Execution

#### What it is

Allow pi to launch **speculative workers** for blocked downstream tasks when their blocker is already active and likely to land soon.

Example:

- bw-10 is active and blocks bw-11.
- bw-11 is mostly independent except for an API shape from bw-10.
- Pi launches bw-11 in a speculative worktree based on bw-10’s branch or expected interface.
- bw-11 cannot land until bw-10 lands.
- If bw-10 changes direction, bw-11 is refreshed, remediated, or discarded.

#### Why it is disruptive

This is branch prediction for software work.

Dependency graphs often serialize work unnecessarily because blocked tasks cannot start until blockers are closed. With worktrees and explicit landing gates, beadwork/pi can safely do controlled speculation.

This should be an advanced, opt-in mode:

```bash
/bw run bw-epic --workers 6 --speculative 2
```

Safety rules:

- speculative workers never auto-land,
- must declare assumption points,
- must rebase after blocker lands,
- must be killed if blocker fails validation or changes scope,
- must be clearly marked in `/bw workers`.

This is high ambition, not MVP, but it is exactly the kind of innovation pi worktrees make possible.

### Pillar 6 — Plan CI and Graph Repair

#### What it is

Treat plans as artifacts that can fail validation.

Commands:

```bash
bw validate-plan --scope bw-epic
bw dep cycles --json
bw suggest --type dependency --json
bw repair-plan --preview --scope bw-epic
```

Pi `/bw adopt` should run plan validation after converting markdown into beadwork issues.

Checks:

- dependency cycles,
- orphan high-priority tasks,
- P0/P1 blocked by low-priority work,
- giant “god” epics with no internal edges,
- tasks with vague acceptance criteria,
- labels with unhealthy flow,
- over-deep critical path,
- duplicated titles/descriptions,
- stale in-progress work,
- recently changed files with no matching issue,
- closed issues whose dependents did not become ready as expected.

#### Why it is disruptive

Agents often create task graphs that look plausible but execute poorly. Plan CI would catch broken decomposition before a swarm burns hours.

This makes beadwork the place where agent-generated plans become operationally trustworthy.

### Pillar 7 — Adaptive Feedback Loop

#### What it is

Record which recommendations were accepted, ignored, successful, failed validation, required remediation, conflicted on landing, or generated follow-up work.

Potential commands:

```bash
bw feedback accept --recommendation rec-123 --issue bw-42
bw feedback ignore --recommendation rec-123 --reason "wrong domain"
bw outcomes --json
bw score explain bw-42
```

Pi can write feedback automatically:

- worker launched from recommendation,
- worker closed ticket,
- validation passed/failed,
- review requested changes,
- landing conflicted,
- remediation attempts used,
- operator manually landed or canceled.

#### Why it is disruptive

Ranking systems usually stay static. Beadwork can learn locally from its own execution history while staying transparent and git-native.

This does **not** require opaque ML. Start with explainable outcome statistics:

- tasks in label `testing` often validate quickly,
- tasks touching `internal/repo` frequently conflict,
- worker model X has high remediation rate on docs,
- P2 quick wins often unblock no one and should not outrank P1 blockers,
- issues with no acceptance criteria have higher reopen rate.

Then feed those signals into future triage as visible score parts.

---

## 6. New command surface: from CLI to agent protocol

### 6.1 Command tiers

#### Tier A — MVP graph intelligence

```bash
bw triage [--scope ID] [--top N] [--include-blocked] [--json]
bw plan [--scope ID] [--workers N] [--json]
bw ready --ranked [--json]
bw dep cycles [--json]
```

#### Tier B — control-plane operations

```bash
bw partition --scope ID --agents N --json
bw validate-plan --scope ID --json
bw alerts [--scope ID] [--json]
bw drift --since REF --json
bw brief --for worker|planner|reviewer|orchestrator [ID] [--json|--markdown]
```

#### Tier C — causal memory and code graph

```bash
bw events --json
bw history ID --causal --json
bw correlate --issue ID --json
bw correlate --commit SHA --json
bw file-beads PATH --json
bw related ID --json
bw orphans --json
```

#### Tier D — domains, recipes, search, and feedback

```bash
bw label health --json
bw label flow --json
bw label attention --json
bw recipe list
bw triage --recipe high-impact --json
bw search "auth callback race" --json
bw feedback accept|ignore ...
bw outcomes --json
```

### 6.2 Standard robot envelope

Every new machine-facing command should use a shared envelope:

```json
{
  "schema_version": "bw.<command>.v1",
  "generated_at": "2026-04-29T06:00:00Z",
  "repo": {
    "root": "/repo",
    "worktree": false,
    "branch": "main",
    "dirty": false
  },
  "beadwork": {
    "ref": "refs/heads/beadwork",
    "commit": "abc1234",
    "prefix": "bw"
  },
  "data_hash": "sha256:...",
  "status": {},
  "warnings": [],
  "data": {},
  "commands": {},
  "usage_hints": [],
  "errors": []
}
```

This is one of the most important architecture decisions. It turns beadwork into a stable local protocol for agents.

### 6.3 Formats

Support should be staged:

1. `--json` for canonical output.
2. `--format json` as explicit synonym.
3. `--schema` or `bw schema <command>` for JSON Schema.
4. Optional `--format toon` later for token efficiency.
5. Markdown/human output stays concise.

---

## 7. Graph and scoring model

### 7.1 Graph layers

Beadwork should not treat “the graph” as one thing. It should define layers:

1. **Scheduling graph** — blocking dependencies only.
2. **Scope graph** — parent/child hierarchy.
3. **Domain graph** — labels and cross-label flows.
4. **Code graph** — issue↔commit↔file correlations.
5. **Worker graph** — active workers, worktrees, leases, landing states.
6. **Time graph** — events, drift, velocity, stale WIP.
7. **Semantic graph** — optional relatedness via text/search/embeddings later.

The MVP uses layers 1–3. The disruptive product emerges when layers 4–6 are added.

### 7.2 Score parts

A recommendation should never be a black box. Suggested score parts:

- priority,
- readiness,
- downstream unlocks,
- transitive blocker depth,
- critical path position,
- overdue urgency,
- stale WIP pressure,
- domain health / label attention,
- recent movement relevance,
- file conflict risk,
- worker availability,
- confidence level,
- historical outcome adjustment.

Example:

```json
{
  "id": "bw-42",
  "score": 91.4,
  "action": "start",
  "confidence": 0.86,
  "score_parts": [
    { "name": "priority", "value": 22, "reason": "P1 issue" },
    { "name": "readiness", "value": 20, "reason": "no unresolved blockers" },
    { "name": "unblocks", "value": 18, "reason": "unblocks 3 direct / 7 transitive issues" },
    { "name": "domain_health", "value": 10, "reason": "database label attention is critical" },
    { "name": "conflict_risk", "value": -4, "reason": "possible overlap with worker bw-39" }
  ],
  "reasons": [
    "ready P1 work",
    "highest downstream unlock in scope",
    "on critical path for release",
    "database domain is currently the release bottleneck"
  ]
}
```

### 7.3 Ranking modes

Different operators need different policies:

- `actionable` — ready work only.
- `unblock` — blocked items ranked by blocker to clear.
- `critical-path` — longest dependency chain first.
- `safe-parallel` — minimize file/track overlap.
- `recent` — emphasize recent commits and active context.
- `human-needed` — ambiguous/stale/high-risk issues.
- `review` — items awaiting review/landing.

This is where recipes become powerful.

---

## 8. Pi-beadwork-extension: from dashboard to operator console

### 8.1 New model tools

Expose the control-plane surface:

- `beadwork_triage`
- `beadwork_plan`
- `beadwork_partition`
- `beadwork_alerts`
- `beadwork_brief`
- `beadwork_history`
- `beadwork_file_beads`
- `beadwork_conflict_risk`
- `beadwork_label_health`
- `beadwork_drift`

These should call `bw` structured outputs through the existing adapter layer.

### 8.2 Track-aware bounded runs

`/bw run <epic>` should consume `bw plan --json` and schedule by:

1. graph track,
2. priority and unblock value,
3. label/domain health,
4. active worker count per track,
5. predicted file overlap,
6. landing/validation risk,
7. operator-specified mode.

Modes:

```bash
/bw run bw-epic --mode safe          # avoid overlap, conservative
/bw run bw-epic --mode throughput    # maximize workers with acceptable risk
/bw run bw-epic --mode unblock       # focus highest-impact blockers
/bw run bw-epic --mode speculative   # allow controlled downstream speculation
```

### 8.3 Worker handoff as compiled context

`launchTicketWorker` already writes handoff files. Replace ad hoc context with `bw brief --for worker <ticket>` output.

The handoff should include:

- selection reason,
- graph rank and track,
- scope boundary,
- relevant comments/attachments,
- likely files and file-risk warnings,
- active nearby workers,
- validation commands,
- close checklist,
- landing policy,
- what to comment back if blocked.

### 8.4 Operator dashboard surfaces

Do **not** port the `beads_viewer` TUI. Instead add compact operator surfaces:

- **Control line:** ready / blocked / tracks / cycles / active workers / held workers.
- **Top next:** the current highest-value next actions.
- **Track board:** each execution track with active/ready/blocked/held state.
- **Domain health:** label chips showing attention and velocity.
- **Conflict map:** active workers and predicted overlapping files.
- **Drift alerts:** what changed since last turn/session.
- **Landing queue:** completed workers ordered by impact/risk.

### 8.5 Proactive notices

The extension already de-duplicates worker notices. Add graph-level notices when material changes occur:

- new cycle introduced,
- P0/P1 becomes blocked,
- a worker close unlocks a high-impact track,
- a track has ready work but no active worker,
- active workers are likely to collide,
- worker output invalidates a speculative downstream worker,
- plan drift exceeds threshold,
- stale in-progress issue crosses threshold.

---

## 9. Recent-commit workflow: first-class design target

The user’s actual workflow is not just “rank my backlog.” It is closer to:

1. inspect recent repo movement,
2. infer what the graph should become,
3. set up high-leverage work items,
4. launch multiple agents,
5. keep agents aligned through dynamic prompts,
6. land safely,
7. repeat after compaction or session handoff.

Beadwork should explicitly support that.

### 9.1 Recent movement commands

```bash
bw recent --json
bw drift --since HEAD~20 --json
bw triage --recent HEAD~20..HEAD --json
bw suggest --from-recent --json
```

Potential output:

- commits mentioning beadwork IDs,
- commits with no issue reference,
- files changed without matching open/closed issues,
- issues whose likely files changed recently,
- stale issues in recently active areas,
- suggested dependencies from code movement,
- suggested labels from changed paths.

### 9.2 Git-native graph repair

Examples:

```text
Suggestion: commit 9abc123 changed internal/repo/sync.go but no beadwork issue references it.
Action: create follow-up issue or correlate with bw-51.

Suggestion: bw-74 is open and mentions sync replay; recent commits changed sync replay files.
Action: inspect whether bw-74 is already done or should depend on bw-51.

Suggestion: three active workers are likely to touch cmd/bw/command.go.
Action: serialize them or launch only one in that conflict cluster.
```

### 9.3 Why this beats standalone beads_viewer for this workflow

`beads_viewer` can rank `.beads/beads.jsonl`. Beadwork can rank **work in relation to the actual git repo and active pi workers**.

That is the user’s gap: not merely “what does the graph say,” but “what should the swarm do next given recent commits, worker state, and graph health?”

---

## 10. Data model and storage

### 10.1 Keep issue schema stable initially

Do not overload the issue JSON schema for every new analysis field. Most intelligence should be derived.

Initial derived data can be computed on demand from:

- issue files,
- block marker files,
- parent relationships,
- comments,
- beadwork branch history,
- git log of main branch,
- pi worker registry.

### 10.2 Add optional event/feedback artifacts later

If derived computation becomes expensive or feedback needs persistence, add explicit artifacts on the beadwork branch:

```text
events/
  <event-id>.json
correlations/
  issue/<id>.json
  commit/<sha>.json
feedback/
  recommendations/<id>.json
snapshots/
  graph/<timestamp>.json
recipes/
  <name>.yaml
```

These should be append-friendly and merge-friendly.

### 10.3 Worker state bridge

The pi extension currently stores worker runtime state outside the beadwork branch. That is appropriate for volatile runtime, but selected worker outcomes should become durable beadwork context:

- worker launched,
- model/provider,
- branch/worktree,
- validation result,
- review result,
- landing result,
- files changed summary,
- remediation count,
- final status.

This can be stored as typed comments initially to avoid schema churn:

```text
worker-result: worker=abc ticket=bw-42 status=landed validation=passed head=1234 files=7 remediation=0
```

Later it can become structured event metadata.

---

## 11. Recipes: operationalizing Jeffery-style workflows

Recipes should be first-class because they let users encode how they actually run agent swarms.

Examples:

### 11.1 `safe-parallel`

Goal: maximize concurrency with low merge risk.

Policy:

- one worker per dependency track,
- avoid same predicted file cluster,
- avoid in-progress labels with low health,
- prefer ready P1/P2 tasks,
- exclude tasks with vague acceptance criteria.

### 11.2 `critical-unblock`

Goal: clear the highest-impact blockers.

Policy:

- include blocked issues only as “blocker-to-clear” recommendations,
- score by transitive downstream count and priority of dependents,
- highlight one blocker per critical path.

### 11.3 `recent-commit-followup`

Goal: convert recent movement into graph updates.

Policy:

- find orphan commits,
- find open issues whose files changed,
- suggest dependency/close/comment actions,
- propose follow-up tasks where code changed without issue coverage.

### 11.4 `release-cut`

Goal: prepare a release branch.

Policy:

- focus release-labeled work,
- warn on cycles and P0 blockers,
- include validation/landing status,
- require no stale in-progress issues,
- show forecast/risk.

Recipes can be stored in repo config, user config, or on the beadwork branch. They should be visible to both CLI and pi dashboard.

---

## 12. Implementation architecture

### 12.1 New Go packages

Suggested package boundaries:

```text
internal/graph/
  graph.go          // scheduling graph from issues + blocks
  scope.go          // parent/subtree handling
  metrics.go        // degree, topo, depth, connected components, cycles
  score.go          // explainable score parts
  triage.go         // recommendation assembly
  plan.go           // execution tracks and partitioning
  labels.go         // label/domain health and flow
  alerts.go         // graph health warnings
  envelope.go       // robot output envelope helpers

internal/events/
  intents.go        // parse beadwork branch commit intents into typed events
  history.go        // issue lifecycle timelines
  snapshots.go      // optional graph snapshots later

internal/correlation/
  commits.go        // explicit ID / co-commit / temporal correlation
  files.go          // issue↔file indexes
  confidence.go     // transparent confidence scoring

internal/brief/
  compiler.go       // role-specific context assembly
  budget.go         // token/line budget enforcement

internal/recipe/
  recipe.go         // saved operational policies
  eval.go           // apply filters/scoring modes
```

The first PR does **not** need all of these. But naming the architecture now prevents `internal/graph` from becoming a dumping ground.

### 12.2 Keep analysis read-only

All analysis packages should be read-only by default. Mutations should remain in existing issue/store methods and commands.

Commands like `bw suggest` and `bw repair-plan` should preview by default and require explicit apply flags later.

### 12.3 Performance posture

Adopt `beads_viewer`’s performance lesson:

- phase 1: cheap deterministic metrics,
- phase 2: optional expensive metrics,
- explicit metric status,
- no slow analysis in default `bw ready`,
- golden tests for output stability,
- benchmark large graphs.

### 12.4 Output determinism

Every robot output must be deterministic:

- stable map iteration,
- stable tie-breaking,
- stable timestamps in tests via `BW_CLOCK`,
- stable `data_hash`,
- explicit schema versions.

---

## 13. Implementation sequence

### Phase 0 — Contract and examples

Deliverables:

- JSON envelope standard.
- Example outputs for `triage`, `plan`, `brief`, `alerts`.
- Golden fixtures for chain, diamond, independent tracks, cycle, scoped epic, label-flow.
- Decision on schema naming.

Acceptance:

- Maintainers can review the protocol before code spreads.
- Examples are good enough for pi adapter tests.

### Phase 1 — Graph MVP

Deliverables:

- `internal/graph` with scheduling graph, counts, downstream unlocks, blocker depth, cycles, connected components.
- `bw triage --json` and compact human output.
- `bw ready --ranked` optional view.

Acceptance:

- Ready ranking excludes blocked work by default.
- Blockers-to-clear recommendations exist.
- All scoring is explainable.
- Tests cover cycles, diamonds, parent scopes, due/defer interactions.

### Phase 2 — Execution planning and partitioning

Deliverables:

- `bw plan --json`.
- Track grouping.
- Recommended worker count.
- `bw partition --agents N --json`.

Acceptance:

- Tracks are stable and deterministic.
- Plan can feed pi without additional inference.
- Dry-run/human output shows why workers were chosen.

### Phase 3 — Prompt and brief compiler

Deliverables:

- Compact graph intelligence in `bw prime`.
- Graph-aware `bw start` context.
- `bw brief --for worker|planner|reviewer|orchestrator`.

Acceptance:

- Prime stays compact.
- Worker brief includes selection reason and graph context.
- Briefs have token/line budgets.

### Phase 4 — Pi adapter and track-aware run

Deliverables:

- TypeScript adapter methods for triage/plan/brief.
- `beadwork_triage`, `beadwork_plan`, `beadwork_brief` tools.
- `/bw run` consumes plan tracks.
- Worker handoffs use compiled briefs.

Acceptance:

- Existing worker lifecycle remains intact.
- Run summary reports tracks.
- Dashboard shows graph health and track occupancy.

### Phase 5 — Alerts, drift, and plan CI

Deliverables:

- `bw alerts --json`.
- `bw drift --since REF --json`.
- `bw validate-plan --scope ID --json`.
- `/bw adopt` runs validation after plan adoption.

Acceptance:

- Cycles and stale high-priority blockers are surfaced before execution.
- Plan smells are warnings unless structurally unsafe.

### Phase 6 — Causal memory and code correlation

Deliverables:

- `internal/events` over beadwork branch intents.
- `bw history ID --causal --json`.
- Initial explicit-ID and pi-landing correlation.
- `bw file-beads PATH --json`.

Acceptance:

- High-confidence correlations explain method and confidence.
- Worker briefs can include related prior issues/files.

### Phase 7 — Conflict-aware scheduling

Deliverables:

- file overlap risk from correlation and active worker state.
- `bw conflict-risk --candidates ... --json`.
- Pi scheduling avoids high-risk combinations.
- Landing queue orders by impact/risk.

Acceptance:

- Conservative mode demonstrably avoids same-file worker collisions.
- Operator can override.

### Phase 8 — Domains, recipes, feedback

Deliverables:

- `bw label health|flow|attention --json`.
- Recipe storage/evaluation.
- Feedback/outcomes events.
- Score explanations include historical outcome factors.

Acceptance:

- Saved recipes produce repeatable plans.
- Recommendation feedback is visible and auditable.

### Phase 9 — Speculative execution

Deliverables:

- Speculative worker mode in pi.
- Downstream assumptions recorded in handoff.
- Landing gates prevent speculative work from merging early.
- Remediation/kill path when blocker diverges.

Acceptance:

- Opt-in only.
- Clear dashboard state.
- No speculative worker can auto-land before prerequisite closure.

---

## 14. Testing strategy

### 14.1 Go unit tests

Graph:

- empty graph,
- single issue,
- chain,
- diamond,
- multiple independent components,
- cycle,
- parent-scoped subtree,
- closed blocker unlock,
- overdue/deferred logic,
- label flow,
- deterministic data hash.

Events/correlation:

- parse intent commits,
- lifecycle timeline,
- explicit ID matching,
- co-commit matching where feasible,
- confidence scoring,
- orphan commit detection.

Briefs:

- budget enforcement,
- role-specific sections,
- stable ordering,
- no giant JSON blobs in markdown.

### 14.2 CLI golden tests

Golden outputs for:

- `bw triage`,
- `bw triage --json`,
- `bw plan --json`,
- `bw partition --agents 3 --json`,
- `bw alerts --json`,
- `bw brief --for worker`,
- `bw history --causal --json`.

Use `BW_CLOCK` for deterministic time.

### 14.3 Pi extension tests

- adapter parsing,
- tool return shapes,
- inactive repo behavior,
- track selection,
- worker handoff content,
- dashboard graph status,
- conflict-risk scheduling,
- deferred/speculative states later.

### 14.4 End-to-end scenarios

1. Adopt an epic plan.
2. Validate graph.
3. Triage ranks a blocker.
4. Plan creates three safe tracks.
5. Pi launches three workers.
6. One worker closes and lands.
7. Newly unblocked work rises in ranking.
8. A second worker is predicted to conflict and is deferred.
9. Prime in a new session reports what changed.

---

## 15. Failure modes and mitigations

### Ranking feels wrong

Mitigations:

- expose score parts,
- include confidence,
- support recipes,
- allow operator override,
- collect feedback without auto-mutating priority.

### The system becomes too clever

Mitigations:

- phase features carefully,
- keep conservative defaults,
- require opt-in for speculation,
- make every automation explainable,
- preserve current simple commands.

### Prompt context gets noisy

Mitigations:

- strict budgeted briefs,
- role-specific context,
- compact prime summary,
- detailed JSON available only via explicit commands.

### Graph analysis slows normal use

Mitigations:

- no expensive analysis in default commands,
- status fields for skipped metrics,
- optional cache/snapshots later,
- performance tests on large synthetic graphs.

### File correlation is inaccurate

Mitigations:

- confidence scores,
- method explanations,
- high-confidence methods first,
- feedback confirm/reject path,
- never block work solely on low-confidence correlation.

### Multi-agent orchestration causes deadlocks

Mitigations:

- leases expire,
- operator override,
- clear worker states,
- avoid hard locks initially,
- use advisory reservations before enforcement.

---

## 16. Security and safety

- Graph analysis should not execute arbitrary shell commands.
- Correlation should use controlled `git` invocations with bounded output.
- Command hints are not auto-executed.
- Speculative workers must not auto-land.
- File paths in briefs should be repo-relative and sanitized.
- No issue text should leave the machine unless a user explicitly enables hosted semantic providers later.
- Pi should continue to route mutations through `bw` and existing landing validation.

---

## 17. Recommended first PR boundary

Even with the ambitious vision, the first PR should be narrow:

1. Shared robot envelope type/helper.
2. `internal/graph` MVP.
3. `bw triage --json`.
4. Compact human `bw triage`.
5. Tests for deterministic scoring and graph fixtures.

Do **not** start with pi changes, semantic search, file correlation, or speculative execution.

But design the output contract so those future features can fit without breaking schemas.

---

## 18. Recommended first “wow” demo

A compelling demo should show more than ranking:

```bash
bw triage --json
```

Shows top blocker and why.

```bash
bw plan --scope bw-epic --workers 4 --json
```

Shows four independent tracks.

```bash
/bw run bw-epic --workers 4 --dry-run
```

Shows which tickets would launch and why.

```bash
bw start bw-42
```

Shows a worker brief with graph context and expected downstream unlock.

```bash
bw close bw-42
```

Shows newly unblocked work and next command.

```bash
bw prime
```

In a new session, summarizes graph health, recent change, and top next move.

This creates the visible flywheel:

> triage → plan → delegate → brief → close → unblock → reprime.

That is the seed of the larger control plane.

---

## 19. Open design questions

1. Should robot envelopes be introduced globally or only for new commands?
2. Should recipes live in repo config, user config, or the beadwork branch?
3. Should pi worker outcomes be durable typed comments or separate event files?
4. How much graph intelligence should `prime` include by default?
5. Should file correlation start with explicit issue IDs only before co-commit parsing?
6. Should `bw partition` be separate from `bw plan`, or a mode of `plan`?
7. What is the minimum safe design for advisory file reservations?
8. How should speculative execution represent assumptions and invalidation?

Recommended defaults:

- Use robot envelopes for new commands first.
- Store recipes on the beadwork branch eventually, but allow config files initially.
- Start worker outcomes as typed comments to avoid schema churn.
- Keep `prime` compact with a pointer to `bw triage --json`.
- Start correlation with explicit IDs and pi landing metadata.
- Make `partition` a separate command once track planning exists.
- Start with advisory conflict warnings, not hard locks.
- Defer speculative execution until conflict-aware scheduling is mature.

---

## 20. Source idea map

| Source idea | Beadwork-native adaptation | Why it matters |
|---|---|---|
| `bv --robot-triage` | `bw triage --json` with robot envelope | Stable agent contract |
| `bv --robot-plan` | `bw plan` / `bw partition` | Feeds pi workers |
| Two-phase metrics | metric status fields and cheap MVP | Keeps CLI fast |
| Agent brief export | `bw brief --for ...` | Compaction-resistant context |
| History correlation | `bw history --causal`, `bw file-beads` | Connects work to code |
| Label health/flow | `bw label health|flow|attention` | Domains as schedulable overlays |
| Alerts/drift | `bw alerts`, `bw drift` | Proactive graph hygiene |
| Recipes | repo-native operational policies | Encodes Jeffery-style workflows |
| Semantic search | optional local-first hybrid search | Find related work cheaply |
| Usage hints/schema | `bw schema`, usage hints in JSON | Better zero-shot agent use |
| Agent swarm protocol | leases, conflict risk, partitioning | Safe parallelism |
| Feedback accept/ignore | recommendation outcome memory | Local adaptive learning |

---

## 21. Final recommendation

Do not merely add graph ranking to beadwork.

Use `beads_viewer` as proof that agent-facing graph intelligence is valuable, then build the more powerful beadwork-native version:

> **A git-native, prompt-aware, worker-aware, history-aware swarm control plane.**

The MVP should be disciplined: `bw triage`, `bw plan`, ranked ready, compact prime/start integration. But the architecture should explicitly anticipate causal memory, conflict-aware scheduling, context compilation, plan CI, drift alerts, recipes, and adaptive feedback.

This is the radically innovative synthesis:

- `beads_viewer` tells agents what the graph means.
- Beadwork remembers the graph as durable git-native work state.
- Pi turns the graph into supervised parallel execution.
- `prime` and `start` compile the graph into agent behavior.
- Worker outcomes feed back into the graph.

That flywheel is the product.
