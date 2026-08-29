# Proposal: Graph Intelligence, Speculative Execution, and Self-Aware Work Management

**Author:** AI assistant (investigating per user request)
**Date:** 2026-04-28
**Status:** Draft v2 — radically revised
**Source material:** beads_viewer v0.13.0 at `/tmp/beads_viewer-research`

---

## The Thesis

beads_viewer is a sophisticated graph-aware triage engine. It computes 9+ graph metrics, builds impact networks from git history, detects drift, and produces ranked recommendations. It's impressive engineering.

But it's fundamentally a **read-only observer**. It reads a JSONL file, analyzes it, and presents results. It doesn't change how work is planned, assigned, coordinated, or executed. It's diagnostic, not operational.

Beadwork + pi-beadwork-extension can absorb bv's best ideas and go dramatically further because of three architectural properties no other tool has simultaneously:

1. **Structured intent log** — Every mutation is a replayable, parseable event in a git commit message. Not raw diffs. Not database rows. Semantic events: `create bw-123`, `start bw-123`, `close bw-123`, `link bw-123 blocks bw-456`. This is a first-class event stream that exists nowhere else.

2. **Isolated execution environments** — Workers run in separate git worktrees on their own branches. Launching, killing, and discarding a speculative worker costs nothing.

3. **Automatic context injection** — `bw prime` renders live intelligence into every agent session. The agent doesn't need to call any diagnostic command or remember any graph tool exists. Intelligence arrives for free.

These three properties enable capabilities that are genuinely novel — not "bv features ported to Go" but new categories of behavior that didn't exist before.

---

## Part 1: The Intent Log as First-Class Event Stream

### What bv does
bv's `correlation/` package (~2500 lines) parses raw `git log` output, extracts lifecycle events (created, claimed, closed, reopened) from diffs to `.beads/beads.jsonl`, matches commits to issues via explicit references, temporal windows, and co-committed files, then reconstructs causal chains with blocked periods, gap analysis, and recommendations.

### Why beadwork makes this trivially better

Beadwork's commit messages on the `beadwork` branch ARE the event stream. Every `store.Create()`, `store.Start()`, `store.Close()`, `store.Comment()` writes a structured intent line:

```
create bw-abc "Implement OAuth2 flow" priority=1 type=task
start bw-abc
link bw-abc blocks bw-def
comment bw-abc "auth module needs refactoring first"
close bw-abc
```

This eliminates bv's entire git-log parsing pipeline. No regex extraction from diffs. No temporal heuristics. No confidence scoring. The event stream is exact, typed, and already exists.

### What this enables (that bv can't do)

#### 1a. Zero-Parse Causal Chains
Walk the intent log chronologically for any issue ID. Instantly produce:
- Full lifecycle timeline (create → start → comment → link → close)
- Exact blocked durations (when a `link X blocks Y` appears and Y is open, to when X closes)
- Gap analysis (time between events → detect stalled work)
- Session boundaries (commits from different sessions are distinguishable by timestamp gaps and potentially by `bw sync` intents)

This is `bw history --causal <id>` — a command that produces the full causal story of an issue, with no git log parsing.

#### 1b. Intent-Derived Velocity (No Separate Analytics)
Count `close` intents in time windows. Calculate:
- Issues closed per day/week/month
- Average time from `create` to `close` 
- Average time from `start` to `close` (active work time)
- Create-to-start latency (how long things sit before someone claims them)
- Weekly trend lines

This feeds directly into triage scoring and `bw prime` context. No separate analytics tool needed.

#### 1c. Pattern Detection Across Sessions
The intent log spans every session that ever touched the issue store. Mine it for:
- **Thrashing detection**: Issue started, then stopped, then started again. Multiple times. Flag as risky.
- **Scope creep signals**: Epic gets new children added after work has started on existing children.
- **Stall patterns**: Issues that go from `in_progress` back to `open` frequently.
- **Session handoff gaps**: Long gaps between events on the same issue suggest a session compacted and context was lost.

Surface these as warnings in `bw prime`: *"⚠️ bw-123 has been started/stopped 3 times across 3 sessions. Consider breaking it down or adding a detailed handoff comment."*

#### 1d. Cross-Session Continuity Auditing
When an agent picks up work via `bw start`, the system can check: "The last event on this issue was 2 days ago by a different session. Here's the last comment left by the previous agent: ..."

This turns the intent log into an automatic context-transfer mechanism. No explicit handoff protocol needed — the intent log IS the handoff.

---

## Part 2: Graph Intelligence — But Operational, Not Diagnostic

### What bv does well
bv computes PageRank, betweenness, HITS, critical path, eigenvector, k-core, cycle detection, and topological sort using gonum's graph library. It produces ranked recommendations, quick wins, blockers-to-clear, parallel execution tracks, and what-if analysis.

### What to take

The graph algorithms are genuinely valuable. The two-phase async model is smart. The composite triage scoring formula works. Take the algorithms, not the architecture.

**New package: `internal/graph/`**

```
internal/graph/
├── graph.go         // Build directed graph from []*issue.Issue using gonum
├── metrics.go       // PageRank, betweenness, critical path, cycles, topo sort
├── triage.go        // Composite scoring, recommendations, quick wins, blockers
├── tracks.go        // Parallel execution tracks via connected components
├── whatif.go         // "What if I close X?" impact analysis
├── drift.go         // Baseline comparison and drift detection
├── velocity.go      // Intent-log-derived velocity metrics
└── graph_test.go
```

Key design:
- Input: `[]*issue.Issue` from `store.List()` — same type already used everywhere
- Two-phase: instant Phase 1 (degree, topo, density, cycles), background Phase 2 (PageRank, betweenness, HITS) — same pattern as bv but without async complexity since CLI commands can just block
- Cache by tree SHA (TreeFS operates on a specific git ref, so tree SHA is a natural cache key)
- `gonum/graph` for algorithms — battle-tested, pure Go, same lib bv uses

### Where it gets radical: the graph is OPERATIONAL

bv computes metrics and presents them. Beadwork can use the same metrics to **change behavior**:

#### 2a. `bw ready --ranked` — Impact-Sorted Work Queue

Current `bw ready` shows unblocked issues grouped by parent. With graph intelligence:

```
$ bw ready --ranked

  Available work (ranked by impact):

  1. bw-abc  P1  "Fix auth token refresh"          (unblocks 5 downstream)  ⚡ quick win
  2. bw-def  P0  "Database migration framework"     (critical path: depth 4)
  3. bw-ghi  P2  "Add rate limiting middleware"      (betweenness: 0.84, bottleneck)
  
  ⚠️ 1 dependency cycle detected. Run `bw dep cycles` to fix.
  📊 Velocity: 3.2 issues/week (↑ from 2.8 last week)
```

The default `bw ready` stays unchanged. `--ranked` opts into graph computation.

#### 2b. `bw triage` — The Mega-Command

```bash
bw triage              # Human-readable: top picks, blockers, health
bw triage --json       # Machine-readable for tools
bw triage --plan       # Parallel execution tracks
bw triage --whatif ID  # "If I close this, what happens?"
bw triage --drift      # Compare to last baseline
```

JSON output includes everything an agent needs in one call:
- `quick_ref`: counts + top 3 ranked picks
- `recommendations`: ranked by composite score
- `quick_wins`: low-effort high-impact items
- `blockers_to_clear`: items that unblock the most downstream work
- `execution_tracks`: independent parallel work streams
- `project_health`: status distributions, graph density, velocity, cycles
- `drift`: changes since last baseline (new cycles, PageRank shifts, velocity changes)

#### 2c. `bw dep cycles` and `bw dep graph`

Cycle detection is safety-critical — circular dependencies make `bw ready` incorrect:

```bash
bw dep cycles            # List all circular dependency paths
bw dep graph --mermaid   # Mermaid diagram
bw dep graph --dot       # Graphviz DOT
bw dep graph --json      # Nodes + edges + metrics
```

#### 2d. Enhanced `bw prime` — The Killer Feature

`bw prime` already renders live context. Inject graph intelligence:

```markdown
{{ if .HasCycles }}
> ⚠️ **{{ .CycleCount }} dependency cycles detected.** Run `bw dep cycles` to fix before proceeding.
{{ end }}
{{ if .TopBlocker }}
> 🎯 **Highest-impact blocker:** {{ .TopBlocker.ID }} "{{ .TopBlocker.Title }}" — completing this unblocks {{ .TopBlocker.UnblockCount }} items
{{ end }}
{{ if .QuickWin }}
> ⚡ **Quick win available:** {{ .QuickWin.ID }} "{{ .QuickWin.Title }}" ({{ .QuickWin.Reason }})
{{ end }}
{{ if .VelocityTrend }}
> 📊 Velocity: {{ .VelocityTrend }} issues/week ({{ .VelocityDelta }})
{{ end }}
{{ if .DriftWarnings }}
{{ range .DriftWarnings }}
> 🔴 {{ . }}
{{ end }}
{{ end }}

## Currently available work (ranked by impact):
{{ bw "ready" "--ranked" "--no-context" }}
```

Every agent session — every fresh context window after compaction — starts with graph-ranked intelligence. The agent doesn't know graph theory exists. It just sees: "here's the most impactful thing to work on, here's why, here's what's broken."

**This is the ultimate cognitive offloading.** Not a tool the agent has to remember to call. Not a diagnostic it has to interpret. Pre-digested, prioritized, actionable intelligence injected before the first turn.

---

## Part 3: Code-Aware Work Coordination (The Conflict Gate)

### What bv does
bv's `correlation/file_index.go` builds a reverse index from file paths to issues. `network.go` constructs an impact network from shared-commit edges, shared-file edges, and dependency edges, then discovers clusters via connected components. `ImpactAnalysis()` computes risk scores for a set of files by finding all issues that touched them.

### Why this is transformative for beadwork

Beadwork workers run in isolated worktrees. The pi-extension orchestrator launches them independently. Right now, it's **blind to code-level conflicts** — two workers can absolutely stomp on each other's files, leading to merge conflicts during the landing pipeline.

The radical move: **turn file-issue correlation from a diagnostic into a live coordination gate.**

#### 3a. The File-Issue Reverse Index

Build from the working branch's git history (not the beadwork branch). When a commit references issue ID `bw-abc` in its message, map all changed files to that issue. Additionally, use temporal correlation (same author, during active window) and co-change patterns to fill gaps.

Store this as a lightweight cache (JSON file or in-memory, invalidated on new commits).

**New CLI:**
```bash
bw files bw-abc              # Files touched by this issue's commits
bw files --hotspots          # Files touched by the most issues (conflict zones)
bw files --impact auth.go    # Issues that care about this file
bw files --cochange auth.go  # Files that frequently change alongside auth.go
```

**New pi-extension tool:** `beadwork_file_impact`
```typescript
// Input: list of file paths
// Output: affected issues, risk level, warnings about active work
// Agent calls this before making changes to understand the blast radius
```

#### 3b. The Conflict Gate for Worker Orchestration

**Before launching a worker**, the orchestrator:

1. Predicts which files the worker will touch by looking at:
   - Files changed by sibling issues under the same epic
   - Files associated with the ticket's labels/keywords
   - Co-change patterns from historical data
2. Checks the prediction against files currently being modified by running workers
3. If overlap is detected:
   - **Low overlap**: Launch anyway, but flag both workers to coordinate
   - **High overlap**: Serialize the launches — don't start the second until the first lands
   - **Critical overlap**: Flag for human decision

**After a worker lands**, update the file map with actual touched files, refining future predictions.

This turns merge-conflict recovery (remediation relaunches, rebase failures, human intervention) into **merge-conflict prevention**. The system knows, before any code is written, that two tasks will step on each other.

#### 3c. File Hotspot Dashboard

The pi-extension dashboard's Workers tab gains hotspot awareness:
- Files with active work from multiple issues shown in red
- Cluster visualization: "These 3 tasks all touch `internal/auth/` — they form a serial chain"
- Warning when launching a worker into a contested area

---

## Part 4: Speculative Execution — Branch Prediction for Agent Work

### Why this is possible

Three beadwork/pi-extension properties make speculative execution viable:
1. **Worktrees are cheap** — `git worktree add` is nearly free. Creating and discarding worktrees costs nothing.
2. **Workers are isolated** — Each worker has its own branch, its own worktree, its own process. No shared state.
3. **Landing is gated** — Workers don't merge automatically. The orchestrator validates, rebases, and merges-back. This is already a checkpoint where speculative work can be discarded.

### How it works

The graph shows a clear critical path. Task A (in progress) blocks tasks B, C, and D. Traditional orchestration waits for A to finish before launching B, C, D.

Speculative execution:

1. **Identify speculation candidates**: Tasks blocked only by the currently in-progress task, where the in-progress task has high confidence of success (no error signals, making progress, worker still running cleanly).
2. **Launch speculative workers**: Start workers on B, C, D with a special flag indicating they're speculative. The worker's handoff prompt says: "This task is blocked by bw-A. Assume bw-A will succeed and its changes will be available. Work on your task accordingly."
3. **On success**: When A lands, rebase B/C/D onto the new state. If they still pass validation, they're essentially pre-done. Massive time savings.
4. **On failure**: When A fails or gets stuck, kill speculative workers B, C, D. Discard their worktrees. No harm done.

### Impact

For an epic with a serial critical path of depth 5, traditional execution takes 5× the time of a single task. Speculative execution can reduce this to approximately 2× (the critical task plus validation/rebase overhead), because downstream work was done in parallel on the assumption the critical path would succeed.

This is a genuinely novel capability. No work management tool does this because no other tool has both graph intelligence AND cheap isolated execution environments.

### Safety

- Speculative workers are clearly marked in the registry
- They don't land until their blocker resolves
- They can be killed at any time with zero side effects
- The orchestrator caps the number of speculative workers (configurable)
- Speculative work is discarded on blocker failure — no partial states

---

## Part 5: Drift Detection as Continuous Intelligence

### What bv does
bv's `drift/` package compares current graph metrics against a saved baseline. It detects: new cycles, density growth, node/edge count changes, blocked increases, PageRank shifts, staleness, blocking cascades, velocity drops. Results are alerts with severity levels (critical/warning/info).

bv's `baseline/` package saves/loads metric snapshots as JSON files in `.bv/baseline.json`.

### What beadwork can do differently

bv's drift detection is **manual** — you save a baseline, then explicitly run a drift check. Beadwork can make drift detection **continuous and automatic** because:

- Every `bw sync` updates the beadwork branch ref
- `bw prime` runs at session start
- The tree SHA changes when any issue changes

#### 5a. Automatic Baseline Snapshots

After every `bw sync`, automatically snapshot the current graph metrics and store them as a `.bwconfig` entry or a special file on the beadwork branch. No manual `bw baseline save`.

#### 5b. `bw prime` Drift Injection

When `bw prime` renders, compare current metrics against the last snapshot:

```markdown
{{ if .DriftAlerts }}
## ⚠️ Changes Since Last Session
{{ range .DriftAlerts }}
- {{ .Icon }} {{ .Message }}
{{ end }}
{{ end }}
```

Example output:
```
## ⚠️ Changes Since Last Session
- 🔴 2 new dependency cycles introduced (bw-abc → bw-def → bw-abc)
- 🟡 Blocked issues increased from 3 to 7 (4 new blocked items)
- 🟡 PageRank leader changed: bw-ghi replaced bw-jkl as top bottleneck
- 🔵 6 issues closed since last session (velocity: 4.2/week, ↑ 31%)
```

The agent sees, immediately, what changed while it was away. No command to run. No diagnostic to remember. The drift report is part of the ambient context.

#### 5c. `bw triage --drift` for Deep Analysis

When the agent needs more detail:
```bash
bw triage --drift     # Full drift report with all alerts
bw triage --drift --since HEAD~5  # Drift since a specific point
```

---

## Part 6: Comments as Structured Agent Memory

### The problem

AI agents lose context constantly. Compaction, session boundaries, crashes. The most important context — design decisions, discovered blockers, implementation notes — evaporates.

### The solution

Beadwork comments already survive compaction (they're in the git-stored issue JSON). Make them a first-class knowledge system:

#### 6a. Typed Comments

Support structured comment prefixes that `bw prime` and tools can parse:

```bash
bw comment bw-abc "decision: Using JWT for auth tokens because OAuth2 is overkill for internal services"
bw comment bw-abc "blocker: Need database migration before this can proceed"
bw comment bw-abc "note: The auth module at internal/auth/ needs refactoring - don't extend the existing interface"
bw comment bw-abc "handoff: Completed steps 1-3. Step 4 (middleware integration) is ready to start. The key insight is that rate limiting must happen before auth validation."
```

#### 6b. Cross-Issue Comment Surfacing in `bw prime`

When an agent starts a session and `bw prime` renders the ready queue, for each actionable issue:
- Check its blockers' comments for `handoff:` and `note:` entries
- Check its parent epic's comments for `decision:` entries
- Surface the most recent relevant comments

```markdown
## Currently available work:
1. bw-def "Add rate limiting" (P1)
   > 💬 Note from previous session on blocker bw-abc: "The auth module at internal/auth/ needs refactoring"
   > 💬 Handoff from bw-abc: "Completed steps 1-3. Step 4 is ready."
```

This creates **durable inter-session context** without any explicit handoff protocol. The agent writes comments as it works. Future agents (or the same agent after compaction) see them automatically.

#### 6c. Worker Handoff Comments

When a pi-extension worker finishes (regardless of success/failure), automatically append a comment summarizing what happened:

```bash
bw comment bw-abc "worker-result: Worker completed successfully. 3 commits, 4 files changed. Validation passed. Key changes: added JWT middleware to internal/auth/middleware.go, updated config schema."
```

Or on failure:
```bash
bw comment bw-abc "worker-result: Worker failed validation. Test failures in auth_test.go (3 failures). Remediation attempted but tests still fail. Manual review needed."
```

This means the next agent to pick up the work knows exactly what happened, even across completely separate sessions.

---

## Part 7: Plan Validation — CI for Work Plans

### What bv does
bv runs `br dep cycles` to find circular dependencies. That's it for validation.

### What beadwork should do

When a plan is created (via `bw create` with parent/dependency links, or via pi-extension's `/bw adopt`), validate the structural integrity of the resulting graph:

```bash
bw validate                    # Validate entire issue graph
bw validate --epic bw-parent   # Validate a specific epic's subtree
```

Checks:
1. **Cycle detection**: Circular dependencies that make `bw ready` incorrect
2. **Orphan detection**: Open issues with no dependencies and no parent — probably should be linked somewhere
3. **Dangling dependencies**: Dependencies on issues that don't exist (deleted or typo'd)
4. **Critical path sanity**: If the critical path is >10 sequential tasks deep, warn about serial bottleneck
5. **Overcommitment detection**: If >N tasks are assigned to the same entity and they're all in_progress
6. **Missing links**: Two issues that reference the same files/concepts in their titles but have no dependency relationship — suggest adding one
7. **Parallel opportunity**: Parts of a serial chain that could be parallelized

`/bw adopt` in the pi-extension should run validation automatically after converting a markdown plan to issues. Block adoption if cycles are detected. Warn about structural issues.

**This turns plan validation from a manual review into an automated structural analysis that catches errors before any work begins.**

---

## Part 8: The Recursive Learning Loop

### The idea

The intent log is a complete history of every issue lifecycle across all sessions. Over time, this accumulates patterns that can be mined:

#### 8a. Historical Duration Estimation

When a new issue is created, look at closed issues with similar properties (same type, same label, similar title keywords, same parent epic) and estimate:
- Expected time to completion
- Likelihood of getting blocked
- Common blockers in this area

Surface in `bw show`:
```
bw-xyz "Add OAuth2 flow" (P1, task)
  Estimate: ~3 days based on 7 similar closed issues
  Risk: 60% chance of blocking (4/7 similar issues were blocked at least once)
  Common blocker: database schema changes
```

#### 8b. Epic Decomposition Patterns

When an agent creates a new epic, suggest decomposition based on historical epics:
- "Similar epics had 5-8 child tasks"  
- "Consider including a testing task — 80% of similar epics needed one"
- "Similar epics had a critical path of depth 3"

#### 8c. Velocity-Adjusted Scheduling

When `bw triage --plan` generates execution tracks, use historical velocity to estimate completion dates:
- "Track A: 3 tasks, estimated 4 days at current velocity"
- "Track B: 5 tasks, estimated 7 days (includes 2 tasks in a historically slow area)"

---

## Part 9: Impact-Aware Landing

### The problem

When worker A lands (merges its changes), workers B and C are still running. Worker A changed `internal/auth/middleware.go`. Worker B is also modifying that file. When B tries to land, it hits a merge conflict.

The current pi-extension handles this with remediation (rebase + re-validate). But remediation is expensive — the worker might need to re-run tests, or the merge conflict might be non-trivial.

### Impact-Aware Landing

When worker A's changes are ready to land:

1. **Compute the blast radius**: Which files did A change?
2. **Check other active workers**: Are any of them likely touching the same files? (Use the file-issue reverse index + predictions from part 3)
3. **If overlap exists**:
   - **Preventive rebase**: Before landing A, signal overlapping workers to rebase onto A's branch. The worker incorporates A's changes while still running, before there's a conflict.
   - **Priority merge**: If A and B both modify the same file, land the higher-impact one first (graph ranking), then rebase the other.
   - **Conflict warning**: If preventive rebase isn't possible (worker already finished), flag the overlap before attempting merge-back, so the orchestrator can plan remediation.

This reduces merge-conflict remediation from "rebuild and re-validate from scratch" to "incremental rebase during execution." Workers are never surprised by changes to their files.

---

## Part 10: The File-Aware Agent — "What Am I About to Break?"

### A new agent tool: `beadwork_file_impact`

This is one of the most practically useful ideas, borrowed from bv's `ImpactAnalysis()` but made operational:

```typescript
// beadwork_file_impact tool
// Input: { files: ["internal/auth/middleware.go", "internal/config/config.go"] }
// Output: {
//   risk_level: "high",
//   risk_score: 0.72,
//   affected_issues: [
//     { id: "bw-abc", title: "Auth refactoring", status: "in_progress", overlap_files: [...] },
//     { id: "bw-def", title: "Config migration", status: "open", overlap_files: [...] }
//   ],
//   warnings: ["Active work in progress on internal/auth/ — coordinate before changing"],
//   cochange_files: ["internal/auth/token.go", "internal/auth/middleware_test.go"],
//   summary: "⚠️ 2 active issues touch these files. Coordinate before modifying."
// }
```

The agent calls this before making changes to understand the blast radius. The tool combines:
- File-to-issue reverse index (from working branch history)
- Co-change analysis (files that frequently change together)
- Active work detection (issues in `in_progress` status)
- Risk scoring (in-progress overlap = high risk, open overlap = medium, closed overlap = low)

**This is not a diagnostic. It's a safety net.** The agent asks "is it safe to change auth.go?" and gets an actionable answer.

---

## Part 11: Recipes — Saved Views

bv's recipe system is simple but useful: save filter/sort/display/export configurations as YAML for reuse. Port this as:

```bash
bw recipe save "backend-triage" --status open --label backend --sort ranked
bw recipe run "backend-triage"
bw recipe list
```

Recipes store: issue filters (status, priority, label, type), sort order (ranked, priority, created, updated), display options (columns, grouping), and export format. They're stored on the beadwork branch alongside issues.

This is useful for teams where different people care about different subsets of work.

---

## Part 12: Interactive HTML Export

Port bv's `--export-graph` as:

```bash
bw export --html graph.html
```

Self-contained HTML file with:
- Force-directed graph layout (embed `force-graph.min.js` or equivalent)
- Node sizing by metric (PageRank, betweenness, etc.)
- Status coloring (open=blue, in_progress=yellow, blocked=red, closed=gray)
- Click-to-expand issue details
- Path finding between issues
- Metric overlays (toggle between PageRank, betweenness, critical path views)
- Embedded triage results
- Search/filter

Useful for sharing project status with stakeholders who don't have `bw` installed. Also useful as a debugging artifact: "here's what the project graph looked like when things went wrong."

---

## Implementation: What to Build First

The radical ideas above aren't equally hard or equally valuable. Here's the build order:

### Phase 1: Foundation (beadwork Go CLI)
**Graph package + enhanced `bw ready` + `bw triage`**

This is the foundation everything else builds on. Without `internal/graph/`, nothing else works.

- `internal/graph/graph.go` — build directed graph from `[]*issue.Issue`
- `internal/graph/metrics.go` — PageRank, betweenness, critical path, cycles, topo sort
- `internal/graph/triage.go` — composite scoring, recommendations, quick wins, blockers
- `internal/graph/tracks.go` — parallel execution tracks
- `internal/graph/whatif.go` — what-if impact analysis
- `bw ready --ranked` — impact-sorted output
- `bw triage` command with `--json`, `--plan`, `--whatif`
- `bw dep cycles` and `bw dep graph --mermaid`

**Estimated scope:** ~2000 lines of Go + tests. Most algorithms exist in bv and gonum.

### Phase 2: `bw prime` Intelligence (beadwork Go CLI)
**Graph context + velocity + drift in every session**

This is the highest-leverage change. Every agent session gets smarter for free.

- Enhanced `bw prime` template with graph insights, cycle warnings, quick wins, velocity
- `internal/graph/velocity.go` — intent-log-derived velocity
- `internal/graph/drift.go` — baseline snapshots + comparison
- Automatic baseline on `bw sync`
- Drift alerts in `bw prime`

**Estimated scope:** ~800 lines of Go.

### Phase 3: Validation + Comments (beadwork Go CLI)
**Plan validation + structured comments**

Safety-critical and high-value.

- `bw validate` — cycle detection, orphans, dangling deps, critical path warnings
- Typed comment parsing (decision/blocker/note/handoff prefixes)
- Cross-issue comment surfacing in `bw prime` and `bw show`
- `/bw adopt` gains automatic validation

**Estimated scope:** ~600 lines of Go.

### Phase 4: pi-Extension Tools (TypeScript)
**Expose graph intelligence to LLM tool calls**

- `beadwork_triage` tool (calls `bw triage --json`)
- `beadwork_insights` tool (calls `bw triage --json` with metric details)
- `beadwork_whatif` tool (calls `bw triage --whatif <id>`)
- Enhanced system-prompt injection with triage results
- Graph tab in TUI dashboard

**Estimated scope:** ~500 lines of TypeScript.

### Phase 5: File Correlation + Conflict Gate (beadwork Go CLI + pi-extension)
**Code-aware coordination**

This is the radical one — preventing worker conflicts before they happen.

- `internal/graph/files.go` — file-to-issue reverse index from working branch history
- `bw files` command (`--hotspots`, `--impact`, `--cochange`)
- `beadwork_file_impact` pi-extension tool
- Orchestrator conflict gate: check file overlap before launching workers
- Worker handoff auto-comments

**Estimated scope:** ~1500 lines of Go + ~400 lines of TypeScript.

### Phase 6: Speculative Execution (pi-extension)
**Branch prediction for agent work**

This requires all previous phases to be solid. It's the most novel feature but also the most complex.

- Speculative worker launch in orchestrator
- Speculative flag in worker registry
- Kill-on-blocker-failure logic
- Rebase-on-blocker-success logic
- Caps and safety limits

**Estimated scope:** ~600 lines of TypeScript.

### Phase 7: Learning Loop + Estimation (beadwork Go CLI)
**Historical pattern mining**

- Duration estimation from similar closed issues
- Velocity-adjusted scheduling
- Epic decomposition suggestions
- Thrashing detection

**Estimated scope:** ~800 lines of Go.

### Phase 8: Polish
**HTML export, recipes, impact-aware landing**

- `bw export --html`
- `bw recipe` commands
- Impact-aware landing in orchestrator
- Preventive rebase signaling

---

## What We Explicitly Don't Take

| bv Feature | Why Not |
|---|---|
| **Interactive Bubble Tea TUI** | pi-extension already has a TUI dashboard |
| **JSONL file format** | Beadwork's git-native storage is superior |
| **External binary dependency** | Graph analysis is embedded in `bw` |
| **beads_rust / bd CLI compatibility** | Different ecosystem |
| **MCP Agent Mail integration** | Beadwork uses pi's worker orchestration |
| **Background daemon / file watching** | Beadwork is invoked explicitly |
| **Cass / FTS5 / SQLite search** | Issue sets are small enough for in-memory |
| **WASM graph renderer** | Mermaid/DOT export + simple HTML suffices |
| **Homebrew/Scoop packaging** | `go install` keeps it simple |
| **bv's git-log parsing pipeline** | The intent log eliminates the need entirely |
| **Multi-repo workspace support** | Beadwork is single-repo by design |

---

## The Vision

Today, a work management tool is a database you query. You create issues, you query status, you update fields. The tool is passive.

With these changes, beadwork becomes something different: **a work management system that understands the hidden structure of your project and actively guides agents toward the highest-impact work.**

- An agent starts a session → `bw prime` tells it exactly what to work on and why, with graph-ranked priorities, cycle warnings, and drift alerts
- An agent considers a task → `bw triage --whatif` tells it exactly what completing that task will unblock
- An agent starts coding → `beadwork_file_impact` tells it what other work touches the same files
- Workers launch → the conflict gate prevents merge conflicts before they happen
- A bottleneck is in progress → speculative workers pre-compute downstream tasks
- Work lands → impact-aware landing rebases affected workers proactively
- An agent crashes → the next agent reads handoff comments and picks up exactly where the last one stopped
- Patterns accumulate → the system learns to estimate durations and suggest decompositions

The tool isn't passive anymore. It's a self-aware coordination layer that makes every agent — and every session — smarter than the last.

---

## Appendix: Key Files Investigated in beads_viewer

| File / Package | Lines | What I Learned |
|---|---|---|
| `pkg/model/types.go` | ~350 | Issue model with design/acceptance criteria/notes, 10 status types, extensible issue types, typed dependencies, sprint/forecast/burndown structures |
| `pkg/analysis/graph.go` | 2888 | Core graph engine: gonum adjacency list, two-phase async pipeline, PageRank/betweenness/HITS/eigenvector/k-core/articulation/slack, with per-algorithm timeouts and size-adaptive configuration |
| `pkg/analysis/triage.go` | 1720 | Unified triage engine: composite 8-factor scoring, recommendations, quick wins, blockers-to-clear, project health, velocity (7d/30d windows), staleness detection, track/label grouping |
| `pkg/analysis/plan.go` | 339 | Execution planning: union-find connected components for parallel tracks, unblocks analysis, deterministic ordering |
| `pkg/analysis/insights.go` | 167 | Insights generation: top items per metric, orphan detection, stats aggregation |
| `pkg/analysis/whatif.go` | 266 | What-if analysis: priority explanations with top-3 weighted reasons, enhanced recommendations with impact deltas, transitive unblock computation |
| `pkg/analysis/risk.go` | 357 | Risk signals: fan variance (CV of neighbor degrees), activity churn (comments/day + recency), cross-repo risk, status-sensitive risk, weighted composite |
| `pkg/correlation/correlator.go` | 328 | History report orchestrator: lifecycle event extraction, co-commit extraction, bead history assembly, commit index, stats computation |
| `pkg/correlation/causality.go` | 451 | Causal chain analysis: event timeline reconstruction, blocked-period detection, critical path, gap analysis, average time between events, actionable recommendations |
| `pkg/correlation/temporal.go` | 333 | Temporal correlation: author+window matching with dynamic confidence scoring (active beads, window duration, path-hint matching), extractPathHints from titles |
| `pkg/correlation/file_index.go` | 790 | File-to-bead reverse index, file lookup (exact/prefix/glob), co-change matrix (file pairs that change together), hotspot detection, impact analysis with risk scoring |
| `pkg/correlation/network.go` | 815 | Impact network: shared-commit edges, shared-file edges, dependency edges, cluster detection via connected components, cluster labeling from common file paths, subnetwork extraction |
| `pkg/drift/drift.go` | 619 | Drift detection: 13 alert types (new cycles, density, graph size, blocked increase, PageRank changes, staleness, velocity drop, blocking cascade, abandoned claim, potential duplicate), configurable thresholds, CI exit codes |
| `pkg/baseline/baseline.go` | 231 | Metric snapshots: graph stats + top-N metrics + cycles, versioned schema, git context (SHA/branch/message), save/load/summary |
| `pkg/recipe/types.go` | 119 | Reusable views: filter/sort/view/export configs, relative time parsing |
| `pkg/export/graph_interactive.go` | 340 | Self-contained HTML export: force-directed layout, embedded JS (force-graph.min.js + marked.min.js), full bead data including metrics + commits + dependency edges |
