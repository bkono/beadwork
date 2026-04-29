# Proposal: Integrate Graph Intelligence from beads_viewer into Beadwork

**Author:** AI assistant (investigating per user request)
**Date:** 2026-04-28
**Status:** Draft for review
**Source material:** beads_viewer v0.13.0 at `/tmp/beads_viewer-research`

---

## Executive Summary

[beads_viewer](https://github.com/Dicklesworthstone/beads_viewer) (`bv`) is Jeffery Emanuel's graph-aware triage engine for the Beads ecosystem. It reads `.beads/beads.jsonl` and computes 9+ graph-theoretic metrics (PageRank, betweenness, HITS, critical path, eigenvector, k-core, cycles, density, topo sort) to produce ranked triage recommendations, parallel execution plans, what-if impact analysis, risk signals, and interactive HTML visualizations.

The problem: **bv becomes inaccessible outside Jeffery's single-repo, beads_rust/bd + bv toolchain.** If you work across machines, in worktrees, with multiple repos, or without the external `bv` binary installed, the graph intelligence vanishes. The beads-workflow skill depends on `bv --robot-*` commands that simply aren't there in our workflow.

The opportunity: **beadwork already owns the issue store.** Issues are JSON files on a git branch, loaded into an in-memory TreeFS. Beadwork's `bw prime` already renders dynamic context (ready queue, WIP, git state, expired deferrals) into every agent session. The pi-beadwork-extension already has 22 tools, a TUI dashboard, and a full worker orchestration pipeline.

This means beadwork can deliver graph intelligence *natively*, without an external binary, without a separate JSONL file, and without network access to a viewer service. Every `bw prime`, every `bw ready`, every delegated worker handoff can include graph-derived insights that survive compaction, session boundaries, and agent crashes.

This proposal identifies the highest-value ideas from beads_viewer and reimagines them as first-class beadwork + pi-extension primitives.

---

## What beads_viewer Does Well (Ideas Worth Taking)

### 1. Graph-Theoretic Issue Ranking

bv's core insight: **a flat priority list is insufficient when issues have dependencies.** A P2 task that unblocks 8 downstream items is more impactful than a P0 task that unblocks nothing. bv computes:

| Metric | What It Reveals | Why It Matters |
|--------|----------------|----------------|
| **PageRank** | Recursive dependency importance | Foundational blockers |
| **Betweenness** | Shortest-path traffic / bottlenecks | Gatekeepers between work streams |
| **HITS** | Hub/Authority duality | Distinguishes epics from utilities |
| **Critical Path** | Longest dependency chain | Keystones with zero slack |
| **Eigenvector** | Influence via important neighbors | Strategic dependencies |
| **Degree** | Direct in/out connections | Immediate blockers/blocked |
| **Cycles** | Circular dependencies | Structural errors in the plan |
| **Topo Sort** | Valid execution order | Work queue foundation |
| **k-Core** | Structural cohesion | Tightly coupled clusters |

bv implements this with gonum's graph library, a compact adjacency-list graph, and a two-phase async analysis pipeline (instant Phase 1 for degree/topo/density, async Phase 2 with 500ms timeout for heavier metrics).

### 2. Composite Triage Scoring

bv's `--robot-triage` is a "mega-command" that returns everything an agent needs in one call:
- **quick_ref**: at-a-glance counts + top 3 picks
- **recommendations**: ranked by composite score (PageRank × betweenness × blocker ratio × staleness × priority × urgency × risk)
- **quick_wins**: low-effort high-impact items (log-scaled unblock impact × simplicity × priority)
- **blockers_to_clear**: items that unblock the most downstream work
- **project_health**: status/type/priority distributions, graph density, velocity
- **commands**: copy-paste shell commands for next steps

The scoring formula blends 8 weighted factors into a single composite score, then applies triage-specific boosts (unblock boost, quick-win boost) on top.

### 3. What-If Impact Analysis

For any issue, bv computes: "If I complete this, what happens?" → direct unblocks, transitive cascade, critical-path depth reduction, estimated days saved, parallelization gain. This lets agents make informed decisions about work ordering.

### 4. Parallel Execution Tracks

bv uses union-find on the dependency graph to discover connected components, then builds independent execution tracks that can be worked in parallel. Each track gets its own priority-sorted item list and unblocks analysis.

### 5. Risk Signals

Per-issue risk assessment: fan variance (dependency structure stability), activity churn (comment/edit frequency), cross-repo risk, status risk. Combined into a composite 0-1 risk score.

### 6. Label Health & Attention

Per-label health scoring (velocity, staleness, blocked count → healthy/warning/critical), cross-label dependency flow matrix, attention-ranked labels by `(pagerank × staleness × block_impact) / velocity`.

### 7. Interactive HTML Graph Visualization

Self-contained HTML files with force-directed graph rendering, node sizing by metric, status coloring, hover details, path finding, and mini-map. No server required.

### 8. Cognitive Offloading Philosophy

The foundational design principle: LLMs are bad at graph traversal. Don't force agents to parse raw issue data and hallucinate dependency analysis. Give them pre-computed, deterministic graph intelligence.

---

## What Makes Beadwork Different (Our Unique Strengths)

These strengths mean we can go *further* than bv ever could:

| Beadwork Strength | Why It Matters |
|---|---|
| **Issues ARE the store** | No export step, no JSONL intermediary. Graph analysis can read directly from TreeFS. |
| **`bw prime` renders live context** | Graph insights can be injected into *every* agent session automatically. No agent needs to remember to call `bv --robot-triage`. |
| **pi-extension has 22 tools** | Graph insights can power tool responses, not just CLI output. The LLM can ask "what should I work on?" and get a graph-ranked answer. |
| **Worker orchestration** | Delegated workers get handoff prompts. Graph insights can inform which tickets to delegate and in what order. |
| **Intent replay** | Commit messages are replayable events. We can compute velocity, churn, and history correlation from the intent log itself. |
| **Zero working-tree impact** | All computation happens in git's object database. Graph analysis adds zero overhead to the user's checkout. |
| **CAS safety** | TreeFS refuses stale commits. Graph cache can be invalidated by ref movement. |
| **`bw ready` already computes actionability** | We already walk the dependency graph to find unblocked issues. Extending to full graph metrics is incremental. |
| **Deferred/due date scheduling** | bv doesn't have this. Our scheduling primitives feed naturally into urgency scoring. |

---

## Proposed Integration: Three Tiers

### Tier 1: Graph Intelligence in `bw` CLI (Go — beadwork core)

These changes live in the beadwork Go binary and require no external dependencies.

#### 1a. `internal/graph/` — Lightweight Graph Analysis Package

A new package that builds a directed graph from the issue store and computes metrics. **Not a port of bv's 3000-line graph.go.** Instead, a focused implementation that leverages beadwork's existing data model.

```
internal/graph/
├── analyzer.go       // Build graph from []*issue.Issue, compute metrics
├── metrics.go        // PageRank, betweenness, critical path, cycles
├── triage.go         // Composite scoring, recommendations, quick wins
├── plan.go           // Parallel execution tracks via connected components
├── whatif.go         // "What if I close this?" impact analysis
└── graph_test.go
```

**Key design decisions:**
- Use `gonum/graph` like bv does — it's battle-tested and pure Go.
- Two-phase computation: instant Phase 1 (degree, topo, density, cycles) and optional Phase 2 (PageRank, betweenness, eigenvector). Phase 2 runs only when explicitly requested via `--insights` or `--json` flags, keeping `bw ready` fast.
- Cache graph stats by a hash of the issue tree. Since TreeFS operates on a specific git ref, the cache key is naturally the tree SHA.
- The analyzer takes `[]*issue.Issue` — the same type `store.List()` already returns.

#### 1b. Enhanced `bw ready` — Graph-Ranked Output

`bw ready` currently shows unblocked issues grouped by parent. Add:

- **`bw ready --ranked`**: Sort by composite impact score instead of creation order. Show unblock count next to each issue.
- **`bw ready --insights`**: Show top blockers-to-clear, quick wins, and project health summary alongside the ready queue.
- **`bw ready --json`**: Include graph metrics in JSON output (for pi-extension tools to consume).

The default `bw ready` remains fast and unchanged. The `--ranked` and `--insights` flags opt in to graph computation.

#### 1c. `bw triage` — New Command

A dedicated triage command modeled on bv's `--robot-triage`:

```bash
bw triage              # Human-readable: top picks, blockers to clear, health
bw triage --json       # Machine-readable: full triage result for tools
bw triage --plan       # Parallel execution tracks
bw triage --whatif ID  # Impact analysis for specific issue
```

Output structure (JSON mode):
```json
{
  "quick_ref": {
    "open": 12, "actionable": 8, "blocked": 4, "in_progress": 2,
    "top_picks": [
      { "id": "bw-abc", "title": "...", "score": 0.85, "unblocks": 3, "reasons": [...] }
    ]
  },
  "recommendations": [...],
  "quick_wins": [...],
  "blockers_to_clear": [...],
  "execution_tracks": [...],
  "project_health": {
    "counts": {...},
    "graph": { "density": 0.04, "has_cycles": false, "bottleneck_id": "bw-xyz" },
    "velocity": { "closed_7d": 5, "closed_30d": 18, "avg_days_to_close": 3.2 }
  }
}
```

#### 1d. `bw dep cycles` — Cycle Detection

Beadwork tracks dependencies but doesn't detect cycles. Add:

```bash
bw dep cycles          # List all circular dependency paths
bw dep graph --mermaid # Export dependency graph as Mermaid diagram
bw dep graph --dot     # Export as Graphviz DOT
bw dep graph --json    # Export as JSON (nodes + edges)
```

Cycle detection is critical because circular dependencies make `bw ready` incorrect — an issue might appear "unblocked" when it's actually part of a cycle where nothing can proceed.

#### 1e. Enhanced `bw prime` — Graph Context in Every Session

This is the killer feature. `bw prime` already renders dynamic context. Add graph intelligence:

```markdown
## Work In Progress
{{ bw "list" "--status" "in_progress" }}

## Graph Insights
{{ if .HasCycles }}
> ⚠️ **{{ .CycleCount }} dependency cycles detected.** Run `bw dep cycles` to fix.
{{ end }}
{{ if .TopBlocker }}
> 🎯 **Top blocker to clear:** {{ .TopBlocker.ID }} "{{ .TopBlocker.Title }}" — unblocks {{ .TopBlocker.UnblockCount }} items
{{ end }}
{{ if .QuickWin }}
> ⚡ **Quick win:** {{ .QuickWin.ID }} "{{ .QuickWin.Title }}" — {{ .QuickWin.Reason }}
{{ end }}

## Currently available work (ranked by impact):
{{ bw "ready" "--ranked" "--no-context" }}
```

This means every agent session — every fresh context window — starts with graph-derived intelligence. No agent needs to know about graph analysis. No agent needs to run a separate command. The prime template does the cognitive offloading automatically.

#### 1f. Velocity from Intent Log

Beadwork's commit messages are structured intents (`close bw-abc`, `create bw-xyz`, etc.). We can compute velocity metrics directly from the intent log without any external tooling:

- Issues closed in last 7/30 days
- Average time from create to close
- Weekly closure trend

This feeds into the triage scoring and prime template context.

---

### Tier 2: Graph Intelligence in pi-beadwork-extension (TypeScript)

These changes live in the pi extension and surface graph intelligence through tools, the TUI dashboard, and worker orchestration.

#### 2a. New Tools: `beadwork_triage` and `beadwork_insights`

Register new pi tools that call `bw triage --json` and surface results:

```typescript
// beadwork_triage tool
// Returns: top picks, recommendations, quick wins, blockers to clear
// The LLM can call this to decide what to work on next

// beadwork_insights tool  
// Returns: graph metrics, project health, execution tracks
// The LLM can call this for strategic planning

// beadwork_whatif tool
// Input: issue ID
// Returns: impact analysis (what unblocks if this is completed)
```

These tools replace the need for an external `bv` binary. The LLM gets graph intelligence through the same tool interface it already uses for `beadwork_ready`, `beadwork_show`, etc.

#### 2b. Enhanced `bw prime` Injection

The extension's `before_agent_start` hook already injects beadwork context into the system prompt. Enhance it to include triage results when the mode is "engaged":

```typescript
// In activation.ts, before_agent_start hook:
if (mode === 'engaged' && hasOpenIssues) {
  const triage = await adapter.triage({ json: true });
  appendix += formatTriageForPrompt(triage);
}
```

This gives every LLM turn access to graph-ranked work prioritization without the agent needing to call any tools.

#### 2c. Dashboard Enhancement: Graph Tab

The existing 4-tab TUI dashboard (Issues/Workers/Run/Scope) gets a 5th tab: **Graph**.

The Graph tab shows:
- Dependency graph as ASCII art (like bv's graph view, using box-drawing characters)
- Top bottlenecks highlighted
- Cycle warnings
- Execution tracks visualization

This replaces the need for `bv`'s interactive TUI entirely, within the pi overlay the user already knows.

#### 2d. Smarter Worker Orchestration

The orchestrator currently delegates tickets in the order they appear. With graph intelligence:

1. **`/bw run`** ranks epic children by impact score, delegating highest-impact first.
2. **Worker handoff prompts** include what-if context: "Completing this ticket unblocks 3 others: [list]."
3. **Parallel track awareness**: The orchestrator can launch workers for independent tracks simultaneously, knowing they won't conflict.
4. **Bottleneck detection**: If a worker gets stuck on a high-betweenness ticket, the orchestrator can flag it for attention.

#### 2e. Plan Adoption with Graph Validation

`/bw adopt` converts markdown plans into epic/task graphs. With graph analysis:

1. After adoption, automatically run cycle detection and warn if the plan has circular dependencies.
2. Show the critical path through the adopted plan.
3. Suggest dependency additions the user might have missed (e.g., "Task X mentions Task Y's output but has no dependency link").

---

### Tier 3: Advanced Features (Future)

These are ideas from bv that would be valuable but aren't urgent. They're listed here for completeness.

#### 3a. Interactive HTML Export

```bash
bw export --html graph.html    # Self-contained HTML with force-directed graph
bw export --mermaid plan.md    # Mermaid diagram of dependency graph
bw export --markdown status.md # Markdown status report with embedded diagrams
```

Self-contained HTML files using a force-directed layout (like bv's `--export-graph`). Useful for sharing project status with stakeholders who don't have `bw` installed.

#### 3b. Time-Travel Diffing

```bash
bw diff --since HEAD~10   # What changed in the last 10 commits
bw diff --since 2026-04-01 # Changes since a date
```

Since beadwork issues live on a git branch, we can diff any two points in history. Show: new issues, closed issues, modified issues, cycles introduced/resolved.

#### 3c. Commit-to-Issue Correlation

Walk the working branch's commit history and correlate commits with issues (by ID reference in commit messages). Surface:
- Which commits touched which issues
- Orphan commits (code changes with no associated issue)
- Issue activity timeline

#### 3d. Label/Epic Health Dashboard

Per-epic and per-label health metrics: velocity, staleness, blocked count, risk score. Useful for multi-epic projects where different work streams need independent health monitoring.

#### 3e. Sprint/Milestone Tracking

```bash
bw milestone create "v1.0" --due 2026-05-15
bw milestone add bw-abc bw-def bw-ghi
bw milestone burndown
```

Time-boxed work tracking with burndown projections, scope change detection, and at-risk item flagging.

#### 3f. Semantic Search

Hybrid search combining text relevance with graph metrics. "Find the most impactful open issue related to authentication" returns issues ranked by both textual match and graph importance.

---

## Implementation Priority

| Phase | What | Where | Why First |
|-------|------|-------|-----------|
| **Phase 1** | Graph analysis package + `bw triage` + enhanced `bw ready --ranked` | beadwork Go CLI | Foundation for everything else |
| **Phase 2** | Cycle detection (`bw dep cycles`) + graph export (`bw dep graph --mermaid`) | beadwork Go CLI | Safety-critical (cycles break `bw ready`) + low effort |
| **Phase 3** | Enhanced `bw prime` template with graph context | beadwork Go CLI | Massive leverage — every session gets smarter for free |
| **Phase 4** | `beadwork_triage` / `beadwork_insights` / `beadwork_whatif` tools | pi extension | Expose graph intelligence to LLM tool calls |
| **Phase 5** | Smarter worker orchestration (impact-ranked delegation) | pi extension | Makes `/bw run` dramatically more effective |
| **Phase 6** | Dashboard Graph tab + plan adoption validation | pi extension | Visual feedback + plan quality |
| **Phase 7** | HTML export, time-travel, correlation, sprints | both | Nice-to-have features |

---

## What We Explicitly Don't Take

| bv Feature | Why Not |
|---|---|
| **Interactive Bubble Tea TUI** | pi-extension already has a TUI dashboard. No need for a second TUI. |
| **JSONL file format** | Beadwork's git-native storage is strictly superior. No intermediary needed. |
| **External binary dependency** | The whole point is eliminating external dependencies. Graph analysis is embedded. |
| **beads_rust / bd CLI compatibility** | We're beadwork. Different tool, different ecosystem. |
| **MCP Agent Mail integration** | Beadwork uses pi's worker orchestration, not Agent Mail. |
| **Background daemon / file watching** | Beadwork is invoked explicitly. No daemon architecture. |
| **Cass integration** | Session search is orthogonal to issue tracking. |
| **FTS5 SQLite search index** | Beadwork issues are small enough for in-memory search. No need for a database. |
| **Homebrew/Scoop packaging** | Beadwork is `go install`. Keep it simple. |
| **WASM graph renderer** | Overengineered for our use case. Simple Mermaid/DOT export suffices. |

---

## The Core Thesis

bv's innovation is **cognitive offloading** — giving agents pre-computed graph intelligence instead of making them hallucinate dependency analysis. But bv implements this as an external viewer reading a flat file.

Beadwork can do this *better* because:

1. **The graph intelligence lives inside the issue store itself.** No export, no sync, no external binary. When you call `bw ready`, the graph analysis is right there.

2. **`bw prime` delivers graph insights to every agent session automatically.** This is the ultimate cognitive offloading — the agent doesn't even need to know graph analysis exists. It just sees "here's your ranked work queue, here are the bottlenecks, here are the quick wins."

3. **pi-extension tools expose graph intelligence to LLM reasoning.** When the LLM calls `beadwork_triage`, it gets the same quality of triage recommendations that bv provides, but through the tool interface it already knows.

4. **Worker orchestration uses graph intelligence to make better decisions.** Delegate highest-impact work first. Launch independent tracks in parallel. Detect when bottlenecks are stuck.

5. **Everything survives compaction.** The graph metrics are computed from durable state (git-stored issues), not from ephemeral context. New session, same intelligence.

The result: beadwork becomes a **self-aware work management system** that doesn't just track what you asked for — it understands the hidden structure of your project and surfaces the right work at the right time to the right agent.

---

## Appendix: Key Files Investigated in beads_viewer

| File | What I Learned |
|---|---|
| `pkg/model/types.go` | Issue model: ID, title, description, design, acceptance_criteria, notes, status, priority, type, dependencies, comments, labels, sprints, forecasts |
| `pkg/analysis/graph.go` (2888 lines) | Core graph engine: compact adjacency-list implementation, two-phase async analysis, gonum integration, PageRank/betweenness/HITS/eigenvector/k-core/articulation/slack computation |
| `pkg/analysis/triage.go` (1720 lines) | Unified triage engine: composite scoring, recommendations, quick wins, blockers to clear, project health, velocity, staleness, track/label grouping |
| `pkg/analysis/plan.go` (339 lines) | Execution planning: connected components via union-find, parallel tracks, unblocks analysis |
| `pkg/analysis/insights.go` (167 lines) | Insights generation: top items from each metric, orphan detection |
| `pkg/analysis/whatif.go` (266 lines) | What-if analysis: priority explanations, enhanced recommendations, top what-if deltas |
| `pkg/analysis/risk.go` (357 lines) | Risk signals: fan variance, activity churn, cross-repo risk, status risk, composite scoring |
| `pkg/analysis/config.go` | Size-based analysis configuration: automatic algorithm selection based on graph size |
| `pkg/correlation/` | Bead-to-commit correlation: git log walking, explicit/temporal/co-commit strategies, orphan detection |
| `pkg/export/` | Export engines: Markdown, Mermaid, HTML/WASM interactive graph, static site, SQLite |
| `README.md` (3969 lines) | Comprehensive documentation covering all features, algorithms, agent integration, TUI views |
| `AGENTS.md` | Agent guidelines, toolchain, workflows, bv/br/cass integration patterns |
| `SKILL.md` | Skill file for agent consumption: robot commands, metrics reference |
