# Proposal: Graph-Native Agent Throughput for Beadwork

**Status:** Draft proposal  
**Audience:** beadwork CLI maintainers, pi-beadwork-extension maintainers, agent-workflow designers  
**Source research:** `/tmp/beads_viewer-research`, current `beadwork` repo, and `@solvedbydev/pi-beadwork-extension` architecture  
**Constraint:** This is a planning document only. It intentionally does not implement code or create tickets.

---

## 1. Executive summary

Beadwork already has the most important substrate for durable AI-agent work: issues live in git, the CLI owns mutation, `bw prime` injects live workflow context, and `bw start` can brief an agent at the moment it claims work. The pi extension adds the missing operator layer: scoped sessions, delegated workers, worker supervision, validation, review, and landing.

The main gap is not storage or basic issue operations. The gap is **throughput intelligence**: agents need a graph-aware answer to:

- What should I work on next?
- Which blocker should be cleared first to maximize downstream unlocks?
- Which tasks can run in parallel without stepping on each other?
- Is the project graph healthy enough to delegate to a swarm?
- What context should a worker receive so it does not waste a turn rediscovering the graph?

`beads_viewer` demonstrates this missing layer well. Its strongest idea is not the Bubble Tea UI; it is its **robot-first graph triage contract**: deterministic JSON/TOON outputs, explainable recommendations, graph health, cycle alerts, parallel execution tracks, command suggestions, history/diff views, and agent-safe entry points like `bv --robot-triage`, `--robot-next`, and `--robot-plan`.

This proposal adapts those ideas into beadwork’s native architecture instead of copying `beads_viewer` wholesale. The recommended direction is:

> Turn beadwork into a git-native graph intelligence engine for agent swarms, where `bw prime`, `bw start`, `bw ready`, and pi worker orchestration all consume the same deterministic triage and planning model.

The first valuable slice is deliberately small:

1. Add an internal graph/triage package over beadwork issues and blocking edges.
2. Add `bw triage --json` and `bw plan --json`.
3. Add `bw ready --ranked` as a human-friendly view over the same scorer.
4. Inject compact graph intelligence into `bw prime` and `bw start`.
5. Teach `/bw run` in the pi extension to launch workers by execution track, not just by flat ready order.

This preserves what the user likes about beadwork—git-native state, dynamic prompts, `prime`, `start`, delegated workers—while regaining the high-quality ranking and graph setup that become inaccessible when moving away from the `.beads/beads.jsonl` + `beads_viewer` ecosystem.

---

## 2. Normalized intent

The desired outcome is not “port beads_viewer.” The desired outcome is to make beadwork support the same kind of high-throughput Jeffery Emanuel / agent-flywheel workflow without requiring a separate local single-repo beads stack.

More specifically:

- Keep beadwork as the source of truth.
- Keep the CLI mutation model: agents mutate through `bw`, not by editing JSON files directly.
- Preserve the pi extension’s worker lifecycle and landing machinery.
- Add deterministic, machine-readable graph intelligence as a first-class beadwork capability.
- Use graph intelligence to improve ranking, task graph setup, worker assignment, prompt context, and supervision.
- Make the result feel native to beadwork, not like a compatibility shim around `bv`.

The core product bet:

> Beadwork should become the agent-native work graph, while pi-beadwork-extension becomes the graph-aware operator console and swarm launcher.

---

## 3. Success criteria

### Agent workflow success

- An agent can run one command and receive a deterministic ranked recommendation with reasons.
- `bw prime` tells an agent about graph health, top blockers, and the highest-leverage next moves before the agent acts.
- `bw start <id>` produces a richer work brief including why the issue matters, what it unlocks, active blockers/dependents, parent/subtree context, and suggested validation.
- `/bw run <epic>` launches independent worker tracks instead of blindly consuming the first N ready tickets.
- A swarm can keep high-priority workers busy while avoiding dependency conflicts.

### Technical success

- The first implementation is small enough to land safely inside the Go CLI.
- No dependency on `.beads/beads.jsonl`, `br`, `bd`, or the `bv` binary.
- Graph output is deterministic and stable under tests.
- JSON output is schema-friendly and explicitly versioned.
- Human output remains concise; verbose intelligence is opt-in or template-driven.
- Existing command behavior remains backward-compatible.

### Product success

- Beadwork becomes compelling for users who like `beads_viewer` ranking but want git-native branch-backed issue storage.
- The pi extension gains a reasoned orchestration policy instead of a queue-only policy.
- The design highlights existing beadwork strengths instead of replacing them.

---

## 4. Non-goals

This proposal does **not** recommend:

- Porting the `beads_viewer` TUI.
- Requiring `.beads/beads.jsonl` as beadwork’s storage format.
- Depending on `beads_viewer`, `br`, or `bd` at runtime.
- Reimplementing every graph metric in `beads_viewer` immediately.
- Adding a background daemon to beadwork CLI.
- Replacing pi-beadwork-extension worker supervision with a separate orchestration system.
- Building semantic search, embeddings, WASM graph rendering, sprint capacity planning, or full history analytics in the MVP.
- Automatically mutating task priorities based on scores.

The goal is to extract the durable idea: **robot-first graph-aware triage for agents**.

---

## 5. Repo context: current beadwork strengths

### 5.1 Git-native issue store

Beadwork stores issue state on a dedicated `beadwork` branch, using a `treefs` abstraction that reads and writes git trees with compare-and-swap semantics. This gives beadwork several advantages for graph intelligence:

- The issue graph is already versioned.
- Sync is already a git operation.
- History and branch context are naturally available.
- Agents can coordinate without an external server.
- The CLI can compute graph intelligence from local state with no network dependency.

Relevant architecture:

- `cmd/bw/` owns CLI command dispatch and output modes.
- `internal/issue/` owns persisted issue state, listing, dependencies, ready/blocked calculations, comments, labels, parent/child relations, due/defer logic, and status transitions.
- `internal/repo/` owns git context, sync, config, branch setup, and linked-worktree handling.
- `internal/treefs/` provides the git-tree-backed filesystem abstraction.
- `prompts/` contains dynamic prompt templates including `prime` and `start`.

### 5.2 Existing graph primitives

Beadwork already has enough graph structure for a useful MVP:

- Issues have `BlockedBy` and `Blocks` fields.
- Blocking edges are backed by marker files under `blocks/`.
- `Store.LoadEdges()` can load forward and reverse dependency maps.
- `Store.Ready()` and `Store.Blocked()` already compute actionable work.
- `Link()` rejects self-blocking, dependency cycles, and child-to-ancestor blocking.
- Parent/child issue trees are already modeled.
- Scoped readiness exists via `ReadyScoped()`.
- Closed-blocker and hidden-blocker display helpers already exist.

This means graph intelligence can start as a thin analytical layer over existing issue data rather than a storage rewrite.

### 5.3 Dynamic prompt surfaces

Beadwork’s prompt surfaces are unusually important:

- `bw prime` already renders live context for agents.
- `bw start` already creates a work brief at claim time.
- The command registry can be invoked from template helpers.
- The pi extension can inject prime content into active sessions.

This is a major advantage over a standalone viewer. Graph intelligence should not live only in a separate command; it should feed the prompts that shape agent behavior.

### 5.4 Pi extension worker orchestration

The pi-beadwork extension already supports:

- Session engagement and scoping.
- Delegated tmux workers.
- Per-ticket worktrees.
- Worker runtime state.
- Validation/review/landing flows.
- Deferred landing.
- Bounded epic runs.
- Worker inspection and status dashboards.

The missing input is a graph-aware scheduler. `/bw run` currently has the machinery to launch and supervise workers; it needs a better model for choosing which workers to launch together.

---

## 6. Source research: what beads_viewer contributes

The highest-value `beads_viewer` ideas are agent-facing rather than UI-facing.

### 6.1 Robot-first command surface

`beads_viewer` strongly separates the human TUI from safe agent output. Its agent workflow centers on robot commands such as:

- `--robot-triage`
- `--robot-next`
- `--robot-plan`
- `--robot-insights`
- `--robot-graph`
- `--robot-alerts`
- `--robot-suggest`
- `--robot-history`
- `--robot-diff`
- label health / flow / attention commands

The lesson for beadwork: provide explicit command contracts for agents instead of expecting agents to infer ranking from human list output.

### 6.2 Deterministic recommendation output

The `beads_viewer` triage contract includes:

- metadata and versioning
- data hash
- graph status
- quick reference counts
- recommendations
- quick wins
- blockers to clear
- project health
- alerts
- suggested commands
- grouped recommendations by track or label

The lesson for beadwork: `bw triage --json` should be self-describing, deterministic, and directly actionable.

### 6.3 Graph analysis with progressive depth

`beads_viewer` uses a two-phase model:

- Phase 1: immediate topology, degree counts, density, simple graph structure.
- Phase 2: heavier graph metrics such as PageRank, betweenness, HITS, eigenvector, cycle detection, and critical path, with status values like computed/approx/timeout/skipped.

The lesson for beadwork: do not block the MVP on advanced centrality. Start with cheap deterministic metrics and design the JSON shape so heavier metrics can be added later.

### 6.4 Explainable scoring

`beads_viewer` impact scoring is not just a single number. It includes component breakdowns and reasons: priority, blocker ratio, staleness, urgency, risk, time-to-impact, and downstream unlocks.

The lesson for beadwork: every recommendation should explain why it is recommended. This is essential for agent trust and operator override.

### 6.5 Parallel execution tracks

`beads_viewer --robot-plan` groups actionable issues into execution tracks using connected components over dependency relations. Tracks make parallel work safer by separating independent streams.

The lesson for beadwork: `/bw run` should consume tracks so multiple workers do useful independent work instead of racing through a flat queue.

### 6.6 Graph health and alerts

`beads_viewer` treats cycles, stale work, blocked high-priority items, and priority misalignment as first-class signals.

The lesson for beadwork: graph health should appear in `bw prime` and `/bw status`, because an agent should know when the plan is structurally unhealthy before it starts implementation.

---

## 7. Adaptation map

| beads_viewer idea | Beadwork-native adaptation | Why this fits beadwork |
|---|---|---|
| `bv --robot-triage` | `bw triage --json` and human `bw triage` | Agent-safe, deterministic recommendation contract |
| `bv --robot-next` | `bw next --json` or `bw triage --top 1 --json` | Single top-pick workflow for agents |
| `bv --robot-plan` | `bw plan --json` | Feed `/bw run` and worker orchestration |
| Graph metrics | `internal/graph` over `issue.Store` | Reuses beadwork issue/dependency model |
| Track grouping | Execution tracks scoped by parent/epic | Directly improves bounded epic runs |
| Suggested commands | Output beadwork commands (`bw start`, `bw show`, `bw dep`) | Keeps CLI as mutation authority |
| Graph health | Prime/status/start prompt sections | Uses beadwork’s dynamic prompt advantage |
| Label health | Later `bw labels insights` or triage grouping | Useful for routing specialists/workers |
| Time travel/diff | Later git-native graph diff using treefs/repo history | Beadwork’s branch model makes this natural |
| Viewer dashboards | Pi extension graph/status panels | Avoids porting TUI while preserving operator value |

---

## 8. Proposed architecture

### 8.1 New internal package: `internal/graph`

Add a small analytical package that depends on `internal/issue` types but does not mutate the store.

Proposed files:

```text
internal/graph/
  graph.go        // graph model and construction from issues + edges
  metrics.go      // cheap deterministic metrics
  score.go        // ranking and explanations
  triage.go       // triage result assembly
  plan.go         // execution track generation
  alerts.go       // cycle/staleness/blocked-priority alerts
  json.go         // versioned output DTOs if needed
  *_test.go
```

The package should be intentionally boring at first. It should compute:

- node count
- edge count
- ready count
- blocked count
- in-progress count
- out-degree / in-degree
- downstream unblock count
- blocker depth
- cycle detection
- connected components
- parent/subtree-aware track grouping
- stale/overdue/deferred summary using existing issue helpers

Advanced centrality can be added later behind explicit status fields.

### 8.2 Graph construction

Inputs:

- all issues from `Store.List(Filter{All...})` or equivalent internal loading
- `Store.LoadEdges()` for blocking dependencies
- parent/child relationships from issue fields
- current time from `Store.Now()` for due/defer/staleness logic
- optional scope ID

Important rule:

> The initial graph should only treat blocking dependencies as scheduling dependencies.

Parent/child relations are hierarchy/context, not blocking edges. They can shape scopes and tracks, but they should not be mixed into central scheduling metrics unless explicitly modeled.

### 8.3 Core DTOs

Sketch, not final schema:

```go
type TriageResult struct {
    SchemaVersion string          `json:"schema_version"`
    GeneratedAt   string          `json:"generated_at"`
    Scope         *ScopeSummary   `json:"scope,omitempty"`
    DataHash      string          `json:"data_hash"`
    Status        AnalysisStatus  `json:"status"`
    QuickRef      QuickRef        `json:"quick_ref"`
    Recommendations []Recommendation `json:"recommendations"`
    QuickWins     []Recommendation `json:"quick_wins"`
    BlockersToClear []BlockerRecommendation `json:"blockers_to_clear"`
    ProjectHealth ProjectHealth   `json:"project_health"`
    Alerts        []Alert         `json:"alerts,omitempty"`
    Commands      CommandHints    `json:"commands"`
}
```

```go
type Recommendation struct {
    ID          string       `json:"id"`
    Title       string       `json:"title"`
    Priority    int          `json:"priority"`
    Type        string       `json:"type"`
    Status      string       `json:"status"`
    Score       float64      `json:"score"`
    ScoreParts  []ScorePart  `json:"score_parts"`
    Reasons     []string     `json:"reasons"`
    Unblocks    []string     `json:"unblocks,omitempty"`
    BlockedBy   []string     `json:"blocked_by,omitempty"`
    Parent      string       `json:"parent,omitempty"`
    Labels      []string     `json:"labels,omitempty"`
    Commands    []string     `json:"commands,omitempty"`
}
```

```go
type ExecutionPlan struct {
    SchemaVersion string           `json:"schema_version"`
    GeneratedAt   string           `json:"generated_at"`
    Scope         *ScopeSummary    `json:"scope,omitempty"`
    Summary       PlanSummary      `json:"summary"`
    Tracks        []ExecutionTrack `json:"tracks"`
    Alerts        []Alert          `json:"alerts,omitempty"`
}

type ExecutionTrack struct {
    ID        string           `json:"id"`
    RootIDs   []string         `json:"root_ids"`
    Reason    string           `json:"reason"`
    Items     []PlanItem       `json:"items"`
    Unblocks  []string         `json:"unblocks,omitempty"`
}
```

### 8.4 Data hash

Use a deterministic hash over graph-relevant state:

- issue IDs
- status
- priority
- type
- parent
- labels
- due/defer timestamps
- blocked_by / blocks
- updated_at

This gives agents and tests a cheap way to detect whether triage results correspond to the graph they inspected.

### 8.5 Analysis status

Borrow the `beads_viewer` status idea but simplify:

```json
"status": {
  "topology": "computed",
  "cycles": "computed",
  "centrality": "skipped",
  "reason": "centrality metrics are not enabled in this version"
}
```

This leaves room for future advanced metrics without pretending they exist in the MVP.

---

## 9. Scoring model

The MVP scoring model should be understandable and deterministic.

Suggested initial score components:

1. **Priority weight** — P0/P1 work should rank highly.
2. **Actionability** — ready issues outrank blocked issues for next-work recommendations.
3. **Downstream unlocks** — issues that unblock more open work get a boost.
4. **Blocker depth** — clearing deep blockers gets a boost.
5. **Overdue urgency** — overdue work gets a boost using existing due-date semantics.
6. **Staleness** — stale open work gets a smaller boost or alert, not necessarily top rank.
7. **Scope fit** — issues inside the current epic/scope outrank unrelated work when scoped.
8. **In-progress penalty** — avoid assigning work already claimed unless explicitly requested.

Example explanation:

```text
bw-42 is recommended because it is ready, priority 1, unblocks 3 open issues, and is on the critical path for epic bw-10.
```

The score should not be magical. Agent users need reasons more than mathematical purity.

### 9.1 Avoid overfitting the score

Do not implement a complex centrality-weighted ranking before the product has real usage data. Instead:

- Make score parts visible.
- Make weights easy to tune later.
- Keep deterministic tie-breaking by priority, score, created time, then ID.
- Add tests that lock expected ordering for representative graphs.

---

## 10. CLI changes

### 10.1 `bw triage`

Purpose: answer “what matters now?”

Examples:

```bash
bw triage
bw triage --json
bw triage --top 5 --json
bw triage --scope bw-123 --json
bw triage --include-blocked --json
```

Human output should be compact:

```text
Graph health: 18 open, 6 ready, 4 blocked, 0 cycles

Top recommendations:
1. bw-17  P1  ready  score 91  Unblocks 3 issues
   Why: ready; high priority; clears downstream work for bw-8
   Next: bw start bw-17

2. bw-22  P2  ready  score 78  Quick win; no dependents
   Why: ready; small leaf task; likely safe parallel worker
   Next: bw start bw-22

Blockers to clear:
- bw-11 blocks bw-15, bw-18, bw-19
```

JSON output is the canonical agent contract.

### 10.2 `bw next`

Optional convenience alias for the top triage item:

```bash
bw next --json
bw next --scope bw-123
```

This may be implemented as `bw triage --top 1` internally, or deferred if command surface growth is a concern.

### 10.3 `bw plan`

Purpose: answer “what can run in parallel?”

Examples:

```bash
bw plan --json
bw plan --scope bw-123 --json
bw plan --workers 4
```

Output should group ready/actionable issues into execution tracks:

```json
{
  "summary": {
    "ready": 6,
    "tracks": 3,
    "recommended_workers": 3
  },
  "tracks": [
    {
      "id": "track-1",
      "reason": "Authentication epic dependency component",
      "items": [
        { "id": "bw-17", "score": 91, "unblocks": ["bw-31", "bw-32"] }
      ]
    }
  ]
}
```

### 10.4 `bw ready --ranked`

Keep `bw ready` familiar, but add ranking as an opt-in mode:

```bash
bw ready --ranked
bw ready --ranked --json
```

This prevents breaking users who expect the current grouped ready view.

### 10.5 `bw dep cycles`

Cycle detection deserves a direct command because cycles are structural execution hazards.

```bash
bw dep cycles
bw dep cycles --json
```

If cycles exist, `bw prime` should mention them prominently.

### 10.6 `bw dep graph`

Later, expose graph export:

```bash
bw dep graph --json
bw dep graph --mermaid
bw dep graph --dot
bw dep graph --scope bw-123 --depth 2
```

This is useful for docs, dashboards, and future pi visualization, but should not block the MVP.

---

## 11. Prompt/template integration

This is where beadwork can become more valuable than a standalone viewer.

### 11.1 `bw prime`

Add a compact “Graph intelligence” section:

```markdown
## Graph intelligence

- Ready: 6; blocked: 4; in progress: 2; cycles: 0
- Top next: bw-17 — P1, ready, unblocks 3 issues
- Highest leverage blocker: bw-11 — blocks 3 open issues
- Parallel tracks available: 3
- Suggested command: bw triage --json
```

If unhealthy:

```markdown
## Graph warnings

- Dependency cycle detected: bw-12 -> bw-19 -> bw-12
- P0 issue bw-8 is blocked by bw-11
- 5 open issues are stale by more than 14 days
```

The content must be short. `prime` should guide the agent, not flood context.

### 11.2 `bw start`

Add a “Why this issue matters” work brief after claiming:

```markdown
## Graph context

- Rank: #1 of 6 ready issues in current scope
- Reason: ready; P1; unblocks bw-31 and bw-32
- Downstream: 2 direct, 5 transitive
- Related track: track-1 Authentication dependency stream
- Watch out: closing this should newly unblock bw-31
```

Also add post-close hints:

```text
After closing, run: bw ready --ranked --scope bw-10
```

### 11.3 `bw close`

Beadwork already has `NewlyUnblocked`. Use this more aggressively:

```text
Closed bw-17.
Newly unblocked:
- bw-31 P1 Finish auth callback
- bw-32 P2 Add auth regression test
Suggested next: bw start bw-31
```

This closes the agent flywheel: work completion immediately feeds next-work selection.

---

## 12. Pi extension integration

### 12.1 New model tools

Expose graph intelligence to agents:

- `beadwork_triage`
- `beadwork_next`
- `beadwork_plan`
- `beadwork_graph_health`
- optionally `beadwork_dep_cycles`

These tools should shell out through the existing adapter, parse JSON, and return typed objects.

### 12.2 `/bw run` track-aware scheduling

Current bounded epic orchestration can become much smarter without changing its worker lifecycle.

Proposed behavior:

1. Call `bw plan --scope <epic> --json`.
2. Select at most one active item per execution track per cycle unless configured otherwise.
3. Prefer high-score items that unblock downstream work.
4. Avoid launching multiple workers in the same dependency component unless they are independent ready leaves and worker capacity remains.
5. Stop early on graph health hazards such as dependency cycles in scope.

This creates safer parallelism.

### 12.3 Worker handoff enrichment

When `launchTicketWorker` writes `handoff.txt`, include graph context:

- why this ticket was selected
- current rank and score
- blockers/dependents
- parent epic
- track ID
- what closing it may unblock
- suggested follow-up command

This uses the extension’s existing handoff machinery rather than inventing a new worker protocol.

### 12.4 Dashboard additions

Avoid porting the `beads_viewer` TUI. Instead, add small graph-native surfaces to the existing pi dashboard:

- compact graph health line in status
- “Top next” panel
- “Blockers to clear” panel
- execution tracks view for scoped epics
- cycle warning banner
- worker tab grouped by track

The dashboard should remain an operator console, not become a full graph visualization app.

### 12.5 Background notices

The extension already de-duplicates worker notices. Add graph-level notices only when state changes materially:

- new dependency cycle introduced
- P0/P1 issue becomes blocked
- worker closure unblocks high-impact task
- epic has ready tracks but no active workers
- all ready work exhausted; project is blocked

---

## 13. Recent-commit workflow adaptation

The user’s workflow preference is not just “rank my backlog.” It is closer to:

1. inspect recent repo movement,
2. infer where agent energy should go next,
3. build or adjust the task graph,
4. keep many high-priority workers busy,
5. use prompt context to keep agents aligned.

Beadwork can support this better than `beads_viewer` because beadwork already lives in git.

Future direction after MVP:

### 13.1 Commit-to-issue correlation

Add optional correlation between recent commits and issue graph:

- commits that mention issue IDs
- files changed by issue-linked commits
- issues whose likely files changed recently
- closed issues with landing metadata from pi workers
- branches/worktrees associated with tickets

Potential command:

```bash
bw recent --json
bw triage --recent HEAD~20..HEAD --json
```

### 13.2 Graph repair suggestions

Use recent commits and issue state to suggest graph adjustments:

- “This closed commit mentions bw-17, but bw-17 is still open.”
- “These 3 open tasks reference files heavily changed this week.”
- “This P1 issue has no dependencies but appears downstream of bw-12 based on recent code ownership.”

This should start as suggestions, not automatic mutation.

### 13.3 Agent flywheel loop

An ideal future loop:

```text
bw prime
  -> graph health + recent movement + top tracks
bw plan --scope epic --json
  -> parallel worker launch plan
/bw run epic
  -> track-aware workers
worker closes issue
  -> landing + validation + graph update
bw close output
  -> newly unblocked recommendations
next bw prime
  -> updated graph and recent changes
```

This is the workflow where beadwork can become more powerful than both a plain issue tracker and a standalone viewer.

---

## 14. Implementation sequence

### Phase 0 — Lock output contracts

Deliverables:

- Draft JSON schema examples for `triage` and `plan`.
- Add golden fixtures representing small dependency graphs.
- Decide command names: `triage`, `plan`, optional `next`.

Acceptance criteria:

- Maintainers agree on stable fields for MVP.
- Existing commands remain unchanged.

### Phase 1 — Internal graph MVP

Deliverables:

- `internal/graph` package.
- Build graph from issues + blocking edges.
- Compute ready/blocked/in-progress counts.
- Compute direct downstream unlocks.
- Compute blocker depth.
- Detect cycles.
- Deterministic data hash.

Acceptance criteria:

- Unit tests cover cycles, diamonds, chains, independent components, parent-scoped graphs, and no-edge graphs.
- Stable sort order is tested.

### Phase 2 — `bw triage` and `bw ready --ranked`

Deliverables:

- `bw triage` human output.
- `bw triage --json` agent output.
- `bw ready --ranked` reusing triage scoring.
- Command hints in JSON.

Acceptance criteria:

- `bw triage --json` is valid deterministic JSON.
- Top recommendation excludes blocked work by default.
- `--include-blocked` can surface blockers to clear.
- Ranking reasons are visible.

### Phase 3 — `bw plan --json`

Deliverables:

- Connected-component track grouping.
- Scope-aware plan generation.
- Recommended worker count.
- Track reasons and unblocks.

Acceptance criteria:

- Independent components become separate tracks.
- A scoped epic only plans inside that scope unless configured otherwise.
- Output is stable under tests.

### Phase 4 — Prompt integration

Deliverables:

- Prime graph summary.
- Start graph brief.
- Close newly-unblocked recommendation polish.

Acceptance criteria:

- Prompt additions stay compact.
- No large JSON blobs are inserted into normal prime output.
- Agents receive direct next commands.

### Phase 5 — Pi extension tools

Deliverables:

- Adapter methods for triage/plan.
- `beadwork_triage` and `beadwork_plan` model tools.
- Status/dashboard graph summary.

Acceptance criteria:

- Tools work only when beadwork is active.
- Inactive repos fail gracefully.
- Output mirrors CLI JSON.

### Phase 6 — Track-aware `/bw run`

Deliverables:

- `/bw run` consumes `bw plan --json`.
- Worker launches are distributed across tracks.
- Worker handoff includes graph context.
- Run summary reports tracks.

Acceptance criteria:

- Bounded run launches no more than configured workers.
- It avoids launching multiple conflicting tickets from the same dependency stream unless safe.
- Existing validation/landing behavior remains unchanged.

### Phase 7 — Later graph intelligence

Possible extensions:

- graph export (`--mermaid`, `--dot`, `--json`)
- label health
- history/diff over beadwork branch
- recent commit correlation
- graph repair suggestions
- advanced centrality metrics
- dashboard mini-map

---

## 15. Testing strategy

### Go unit tests

Add focused tests for `internal/graph`:

- empty graph
- single ready issue
- simple blocker chain
- diamond dependency
- multiple independent components
- dependency cycle
- parent/child scope
- closed blocker unlock behavior
- overdue and deferred issue interactions
- deterministic hash changes only when graph-relevant state changes

### CLI golden tests

Use existing command test patterns to lock:

- `bw triage`
- `bw triage --json`
- `bw plan --json`
- `bw ready --ranked`
- `bw dep cycles`

Prefer stable fixtures with pinned `BW_CLOCK`.

### Pi extension tests

Add TypeScript tests for:

- adapter parsing of triage/plan JSON
- tool registration return shapes
- `/bw run` track selection
- worker handoff graph context
- inactive repo behavior

### Acceptance scenarios

Representative scenario:

1. Create epic with three independent tracks.
2. Add blockers and one high-priority blocked issue.
3. Run `bw triage --json`.
4. Verify blocker-to-clear recommendation.
5. Run `bw plan --scope <epic> --json`.
6. Verify three tracks.
7. Close one blocker.
8. Verify newly unblocked work rises in ranking.

---

## 16. Failure modes and mitigations

### Ranking feels wrong

Mitigation:

- expose score parts and reasons
- keep manual priority meaningful
- make ranking opt-in for `ready`
- add config later only after real usage

### Output becomes too noisy

Mitigation:

- compact human output
- detailed JSON only with `--json`
- prime/start use summaries, not full reports

### Graph metrics slow down normal commands

Mitigation:

- no graph analysis on default `ready` unless requested
- cheap MVP metrics only
- avoid advanced centrality until needed
- add explicit status fields for skipped metrics

### Parent/child hierarchy gets confused with dependencies

Mitigation:

- treat blocking edges as scheduling dependencies
- treat parent/child as scope/context
- test scoped behavior heavily

### `/bw run` becomes too clever

Mitigation:

- keep worker lifecycle unchanged
- make track-aware scheduling explain decisions
- provide dry-run/no-spawn plan view
- allow simple fallback to current ready-order behavior if needed

### Agents over-trust scores

Mitigation:

- include warnings that scores are guidance
- include reasons and graph health
- keep operator override easy

---

## 17. Security and safety considerations

- No shell execution should be added to graph analysis.
- JSON output should derive only from issue metadata already visible through `bw`.
- Command hints are strings for humans/agents, not auto-executed actions.
- Pi extension should continue to route mutations through existing `bw` adapter methods.
- Worktree landing and validation remain unchanged.
- Graph export should avoid leaking file contents unless a future correlation feature explicitly opts in.

---

## 18. Open design questions

1. Should the top-level command be `bw triage`, `bw next`, or both?
2. Should `bw plan` be limited to ready work, or include blocked future sequence items?
3. Should scores be configurable in repo config from day one, or hardcoded until usage stabilizes?
4. Should graph intelligence include `in_review` if beadwork formalizes that status more broadly?
5. Should `bw prime` always include graph intelligence, or only when a repo crosses a size threshold?
6. How should pi-extension display graph health without cluttering the existing dashboard?

Recommended answers for MVP:

- Add `bw triage`; defer `bw next` unless users ask.
- Make `bw plan` include ready items first, with optional blocked lookahead later.
- Hardcode weights initially, but expose score parts.
- Treat unknown statuses conservatively using existing issue status helpers.
- Always include a very compact graph line in `prime`; include warnings only when meaningful.
- Add dashboard graph health as text first, not visualization.

---

## 19. Why this should be beadwork-native

The tempting approach is to export beadwork issues to a `.beads/beads.jsonl` file and keep using `beads_viewer`. That would recover some ranking quickly, but it would make beadwork dependent on an external worldview and would not exploit beadwork’s strongest features.

A native implementation is better because:

- beadwork already owns issue mutation and sync
- beadwork already has prompt templates
- beadwork already has worktree-aware agent delegation through the pi extension
- beadwork can correlate graph state with git branch/history directly
- beadwork can brief workers at `start` and `delegate` time
- beadwork can use graph intelligence during landing and close/unblock loops

`beads_viewer` is best treated as prior art for agent-safe graph contracts, not as a runtime dependency.

---

## 20. Recommended first PR boundary

The first PR should be intentionally narrow:

1. Add `internal/graph` with graph construction, cheap metrics, cycle detection, scoring, and triage DTOs.
2. Add `bw triage --json` plus compact human output.
3. Add tests for deterministic ranking and JSON shape.
4. Do **not** modify pi extension yet.
5. Do **not** add advanced metrics yet.
6. Do **not** alter default `bw ready` ordering yet.

The second PR can add:

- `bw ready --ranked`
- prime/start template summaries

The third PR can add:

- `bw plan --json`
- pi extension adapter/tools
- track-aware `/bw run`

This sequencing keeps risk low while quickly proving whether native graph triage improves the user’s workflow.

---

## 21. Handoff notes for implementation agents

When implementing, preserve these rules:

- Graph analysis must be read-only.
- Blocking dependencies drive scheduling; parent/child drives scope.
- JSON output must be deterministic.
- Every score needs explanations.
- Prompt additions must be short.
- Existing command behavior must not regress.
- Tests should pin time with `BW_CLOCK` where due/defer/staleness matters.
- Pi extension worker lifecycle should not be rewritten; feed it better plans.

Useful starting points in beadwork:

- `cmd/bw/command.go` for command registry shape.
- `cmd/bw/ready.go` for ready output behavior.
- `cmd/bw/start.go` for claim-time brief rendering.
- `cmd/bw/prime.go` for dynamic prompt context.
- `internal/issue/graph.go` / dependency helpers for edge loading and graph guards.
- `internal/issue/list.go` for filtering, overdue, and deferral semantics.
- `prompts/prime.md` and `prompts/start.md` for prompt additions.

Useful starting points in pi-beadwork-extension:

- `src/bw.ts` for adapter methods and JSON parsing.
- `src/index.ts` command/tool registration.
- worker launch and handoff generation paths.
- bounded epic run orchestration.
- worker inspection and notice de-duplication.

Useful source ideas from `beads_viewer` research:

- `pkg/analysis/triage.go` for triage output shape and recommendation categories.
- `pkg/analysis/priority.go` for explainable score composition.
- `pkg/analysis/plan.go` for execution tracks.
- `pkg/analysis/graph.go` for phase/status separation.
- robot command registry for agent-safe command contracts.

---

## 22. Final recommendation

Build graph intelligence directly into beadwork, starting with deterministic triage and execution planning. Then use beadwork’s existing prompt and worker surfaces to make that intelligence operational:

- `bw triage` tells agents what matters.
- `bw plan` tells the orchestrator what can run in parallel.
- `bw prime` keeps every session graph-aware.
- `bw start` gives each worker the reason and context for its ticket.
- `/bw run` uses tracks to keep high-priority workers busy safely.

This is the synthesis: keep beadwork’s git-native durability and pi’s delegated-worker machinery, but add the graph-aware recommendation layer that made `beads_viewer` valuable for agent throughput.
