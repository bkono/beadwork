# Inverted Analysis: What Beadwork Can Do That beads_viewer Structurally Cannot

## The Question

Not "what features can we port?" but: **what becomes possible when your work tracker is git-native with a replayable intent log and lives inside the coding agent's runtime?** What can we build that beads_viewer could never achieve even with unlimited engineering effort, because its architectural primitives don't support it?

---

## beads_viewer's Primitives (The Ceiling)

beads_viewer reads `.beads/beads.jsonl` — a flat export file produced by `br` or `bd export`. From this it builds a graph, computes metrics, and renders views. That's genuinely impressive engineering. But its fundamental constraints are:

1. **Read-only.** It cannot modify issue state. It observes but never acts.
2. **Point-in-time.** The JSONL is a snapshot. There is no history, no event stream, no "how did we get here."
3. **External.** It is a separate binary the agent must be instructed to invoke. It has no knowledge of what the agent is doing, thinking, or about to do.
4. **Stateless between invocations.** Each `bv --robot-*` call starts from scratch. There is no session, no continuity, no memory of what it told the agent last time.
5. **No execution model.** It has no concept of workers, worktrees, concurrent agents, or in-flight changes.
6. **Reconstructed provenance.** To correlate git commits with issues, `bv` parses `git log`, matches author emails to claimed windows, scans file paths for title keywords, and assigns confidence scores. It's heuristic archaeology.
7. **No transactional guarantees.** Two agents writing to `.beads/beads.jsonl` simultaneously will corrupt it. There is no CAS, no merge, no replay.
8. **Export pipeline dependency.** `bv` can't see an issue until `br` exports it. There's an inherent lag and a brittle coupling.

These aren't bugs. They're architectural consequences of being a viewer of someone else's data format.

---

## Beadwork's Primitives (The Foundation)

Beadwork stores issues as JSON files on an orphan `beadwork` git branch. Every mutation is a structured commit message — the intent log. The pi extension lives inside the coding agent's process with lifecycle hooks, prompt injection, and worker orchestration.

The primitives that matter:

1. **The intent log IS the event stream.** Every `create`, `start`, `close`, `link`, `comment`, `attach`, `defer`, `label`, `delete` is a git commit with a machine-parseable message. Not reconstructed. Not heuristic. Authoritative.
2. **TreeFS provides CAS-guarded atomic mutations.** Concurrent writers get a `conflict: ref has moved` error, not silent corruption. The overlay model means mutations are composed in memory and materialized atomically.
3. **Intent replay resolves conflicts by re-executing.** When two agents diverge, `bw sync` doesn't merge text — it replays structured intents against the winner's state. This is event sourcing, not file merging.
4. **The pi extension has `before_agent_start`, `turn_end`, `session_start`, and `session_shutdown` hooks.** It can inject information before the agent thinks, observe what the agent did after it acts, and clean up when the session ends.
5. **`bw prime` injects live context into every agent session.** The agent doesn't need to remember to check status — status is pushed to it.
6. **Workers run in isolated git worktrees with tracked lifecycle.** The system knows which agents are working on which tickets, in which directories, with what validation status.
7. **The extension can gate actions.** It sits between the agent's intent and execution. It can prevent, warn, or augment any operation.
8. **Zero external dependency.** Clone the repo, you have the full issue database and its complete operational history. No export, no sidecar, no daemon.

---

## The Structural Advantages: What Only These Primitives Enable

### 1. Provenance Without Archaeology

**beads_viewer's approach:** Parse `git log`. For each commit, try to match it to an issue by:
- Scanning commit messages for issue IDs (high confidence)
- Matching author email to claimed-issue windows (medium confidence)
- Comparing file paths to bead title keywords (low confidence)
- Assigning confidence scores from 0.20 to 0.85

This is ~2,500 lines of `pkg/correlation/` code for something that is inherently probabilistic. The temporal correlator explicitly says it's "lower-confidence" and caps certainty at 0.85.

**Beadwork's reality:** The intent log already records every state transition with full fidelity. When an agent runs `bw start bw-42`, the commit message is literally `start bw-42 assignee="agent-1"`. When they close it: `close bw-42`. The causal chain is the commit history itself.

**What this enables:**
- **Perfect causal timelines.** Not "we think this commit relates to this issue with 0.65 confidence" but "here is the exact sequence of operations, in order, with timestamps."
- **State transition auditing.** Did an issue go through `open → in_progress → closed` or did it thrash through `open → in_progress → open → in_progress → closed`? The intent log tells you, with zero parsing ambiguity.
- **Cross-session continuity verification.** Agent A started bw-42 in session 1 and the session crashed. Did agent B in session 2 pick it up or did it get lost? Walk the intent log: if there's no `start` or `close` after agent A's last `comment`, it was dropped.
- **Pattern detection on the management process itself.** Not "what happened to the code" but "what happened to how we managed the work." Which issues get reopened most? Which epics have their children re-scoped? Where do agents get stuck in thrash loops? beads_viewer literally cannot observe management process patterns because it only sees current state.

### 2. The Agent Doesn't Ask — It Knows

**beads_viewer's approach:** The agent must be told to run `bv --robot-triage`. It must parse the stdout JSON. It must decide what to do with the information. If it forgets to call `bv`, or calls it with wrong flags, or ignores the output, there's no safety net. The `bv` binary has zero visibility into what the agent subsequently does with its recommendations.

**Beadwork's reality:** The pi extension's `before_agent_start` hook fires before every LLM call. It reads session state, calls `bw prime`, and injects the result directly into the system prompt. The agent literally cannot start thinking without seeing the current work state.

**What this enables:**
- **Ambient operational intelligence.** Graph insights, drift alerts, cycle warnings, velocity trends — all injected into the agent's context before it forms its first thought. No explicit invocation required. No flag to remember. No output to parse.
- **Contextual augmentation.** Because the extension knows the session scope (which epic, which ticket), it can tailor injected context to exactly what's relevant. An agent scoped to epic bw-10 sees bw-10's subtree, not the whole project. beads_viewer has no concept of "the current session's focus."
- **Progressive disclosure.** `bw prime` already includes in-progress work, ready queue, expired deferrals, and overdue items. Each new capability (graph rank, drift, cycles) slots into a template that's already rendered per-session. beads_viewer would need the agent to make N separate `--robot-*` calls and mentally synthesize the results.
- **Invisible upgrades.** When beadwork adds cycle detection to `bw prime`, every agent session immediately benefits. No workflow change, no new command to learn, no prompt engineering. beads_viewer's equivalent would require updating every `AGENTS.md` / `CLAUDE.md` to tell agents about the new `--robot-*` flag.

### 3. Prevention, Not Detection

**beads_viewer's approach:** `bv --robot-graph` tells you there are cycles. `bv --robot-triage` tells you what's blocked. Both are after the fact. The bad state already exists. The agent already created the cycle or started the blocked task. bv can only report the damage.

**Beadwork's reality:** The extension sits inside the agent's tool execution pipeline. It can intercept operations before they commit.

**What this enables:**
- **Cycle prevention at creation time.** When the agent calls `beadwork_add_dependency`, the extension can compute whether this creates a cycle before the intent is committed. "Adding bw-42 blocks bw-17 would create a cycle: bw-17 → bw-23 → bw-42 → bw-17. Refusing." beads_viewer can only tell you the cycle exists after `br` creates it and `bv` analyzes the export.
- **Blocked-work guardrails.** When the agent calls `beadwork_start_issue`, the extension can check `blocked_by` and warn: "bw-42 is blocked by bw-17 (still open). Starting it anyway?" This is impossible for an external viewer.
- **Scope enforcement.** When scoped to an epic, the extension can prevent the agent from starting work on unrelated tickets. "bw-99 is not a child of epic bw-10. Did you mean to change scope?"
- **Structural validation as a gate.** Before `/bw adopt` converts a markdown plan into tickets, the extension can validate: no cycles, no orphans, no dangling references, critical path is reasonable. beads_viewer can validate post-hoc; beadwork can refuse to create the bad state in the first place.

### 4. Conflict-Free Concurrent Work via Intent Replay

**beads_viewer's approach:** If two agents modify `.beads/beads.jsonl` simultaneously, you get a git merge conflict on a JSON Lines file. Resolution is manual. There is no structured merge strategy for interleaved JSON records.

**Beadwork's reality:** When two agents working on the same repo both run `bw sync` and their local branches have diverged:
1. TreeFS attempts a 3-way file-level merge. Because issues are separate files and status/blocks/labels are directory-based indexes, most concurrent changes affect different paths and merge cleanly.
2. If the tree merge has conflicts, `bw sync` falls back to intent replay: reset to the remote tip and re-execute each local intent (`create`, `close`, `link`, etc.) against the winner's state.

**What this enables:**
- **True multi-agent concurrency.** Agent A creates bw-42 and agent B closes bw-17 simultaneously. Their intents touch different files. TreeFS merges them. If they somehow conflict (both try to close the same issue), replay picks the winner and re-applies the loser's other intents. No manual intervention. No corrupt state.
- **Attachment recovery.** The `replayAttach` function can recover blob OIDs from the pre-reset commit tree. Even when the ref is force-reset to the remote tip, the old blobs survive in the ODB. beads_viewer has no concept of attachments, let alone attachment recovery across divergent histories.
- **Sync as a first-class operation.** `bw sync` is a single command that handles fetch, merge, replay, and push. The agent doesn't think about git operations on the beadwork branch. beads_viewer requires `br` to handle sync of `.beads/` files, with no event-sourcing fallback.

### 5. The Execution Model: Workers as First-Class Entities

**beads_viewer's approach:** `bv` has no concept of execution. It knows about issues. It can recommend what to work on next. But it has zero visibility into:
- Whether any agent is currently working on a recommendation
- Whether two agents are about to modify the same files
- Whether a worker succeeded or failed
- Whether landed changes need validation

**Beadwork's reality:** The pi extension maintains a full worker registry with ~60 fields per worker, tracking:
- tmux session/window/pane
- git worktree path and branch
- ticket assignment and status
- validation/review/remediation/landing/cleanup pipeline stages
- timestamps for every lifecycle transition
- error details and remediation attempts

**What this enables:**
- **File conflict prediction.** Before launching worker B on bw-43, check what files worker A (on bw-42) has modified in its worktree. If there's overlap, serialize the launches or warn. beads_viewer can tell you two issues are related; it cannot tell you two agents are about to edit the same file.
- **Blast-radius awareness at landing time.** Before merging worker A's changes back to main, check if workers B and C are working on affected files. Notify them to rebase proactively rather than discovering conflicts after the fact.
- **Speculative execution.** Because worktrees are cheap and landing is gated, you can start a worker on bw-43 (which is blocked by bw-42) speculatively, assuming bw-42 will succeed. If bw-42 fails, kill the speculative worker. If it succeeds, the downstream work is already partially done. This is CPU branch prediction for project management. Only possible when you have isolated execution environments and a landing gate.
- **Automatic remediation.** When a worker's validation fails, the orchestrator can relaunch it in the same worktree with the failure context. When a reviewer requests changes, remediation runs automatically. beads_viewer has no concept of "try again."

### 6. Comments as Durable Agent Memory

**beads_viewer's approach:** Comments exist in the JSONL data model but are static text attached to issues. They're display data. There's no mechanism for an agent in session N to read a comment written by an agent in session N-1 and understand it as a handoff instruction.

**Beadwork's reality:** `bw comment` creates a commit with intent `comment bw-42 "text"`. The comment persists on the beadwork branch and survives compaction, session boundaries, and agent restarts. `bw prime` already renders in-progress work, and `bw show` displays comments.

**What this enables:**
- **Typed comment protocols.** Comments with prefixes like `decision:`, `blocker:`, `handoff:`, `note:` that `bw prime` can parse and surface contextually. When agent B starts bw-43 (which was blocked by bw-42), prime can inject: "Note from bw-42: decided to use JWT instead of session cookies. See commit abc123." beads_viewer's comments are opaque strings with no semantic layer.
- **Cross-issue context propagation.** A `decision:` comment on a blocker should be visible when working on anything it unblocks. A `handoff:` comment on a parent epic should be visible when starting any child. The intent log knows the dependency graph; comments can flow along it. beads_viewer can't inject context into an agent's session because it doesn't have an agent session.
- **Automatic worker result recording.** When a worker completes (success or failure), the extension auto-comments on the ticket with the outcome, validation result, and any error details. This creates a permanent, queryable record of execution history on the issue itself. Not in a log file. On the issue.

### 7. The Closed Loop: Issue State Drives Code Execution

**beads_viewer's approach:** Strictly one-directional. Issues exist → bv analyzes them → agent reads analysis → agent does something. The "does something" part is outside bv's world entirely. There is no feedback loop.

**Beadwork's reality:** Issue state and code execution are in a closed loop:
- `bw start bw-42` → worker launches in worktree → worker writes code → worker runs `bw close bw-42` → extension detects closure → validation pipeline runs → changes land on main → `bw sync` pushes
- `before_agent_start` reads issue state → agent acts → `turn_end` observes results → background supervisor detects worker exit → landing pipeline fires

**What this enables:**
- **Issue state as an execution trigger.** Closing a ticket doesn't just update a JSON field — it triggers a multi-stage landing pipeline: verify, validate, rebase, review, merge-back, cleanup. In beads_viewer, closing an issue is a data change. In beadwork, it's an operational event.
- **Validation as a contract.** "This ticket is not done until lint passes, tests pass, and the reviewer approves" — enforced automatically, not as a convention. beads_viewer can recommend that you validate; beadwork's orchestrator refuses to land without it.
- **Operational feedback loops.** If validation fails repeatedly for issues of type X, that's a signal. If workers on epic Y consistently need remediation, that's a signal. If issues labeled "auth" take 3x longer than issues labeled "ui", that's a signal. All derivable from the intent log + worker registry. beads_viewer has no execution data to learn from.

### 8. Distribution is Solved by Construction

**beads_viewer's approach:** Install `bv` binary (Go binary, needs Homebrew/Scoop/Nix or manual install). Ensure `br` or `bd` is also installed for the export pipeline. Run `br` to export `.beads/beads.jsonl`. Then run `bv`. If the JSONL is stale, the analysis is stale. The README documents this pipeline in detail because it's a real operational concern.

**Beadwork's reality:** `git clone` gives you the full issue database. `bw` is a single binary. There is no export pipeline, no intermediate format, no daemon, no external service. The issues ARE the git data.

**What this enables:**
- **New-contributor zero-setup.** Clone the repo. Run `bw ready`. Start working. No `br init`, no `bv` install, no JSONL export. The onboarding friction is literally zero beyond having the `bw` binary.
- **Offline-first by construction.** Everything works without network. `bw sync` is just `git push`. beads_viewer works offline too, but only if the JSONL was recently exported. Beadwork's data is always current because there is no export step.
- **Auditability.** `git log beadwork` shows every management decision ever made, in chronological order, with structured messages. This is a complete, immutable audit trail that any git tool can read. beads_viewer's JSONL is a point-in-time snapshot with no history.

### 9. The Self-Referential Property

Here's the deepest structural advantage, and it's the one beads_viewer could never replicate:

**Beadwork can analyze its own operational history to improve its own behavior.**

The intent log records not just what issues exist, but how the management process unfolded over time. Every time an agent:
- Creates an issue → records a planning decision
- Starts an issue → records a resource allocation
- Adds a dependency → records a structural insight
- Comments → records a reasoning step
- Closes → records completion
- Reopens → records a failure or scope change

This is metadata about the AI development process itself. Over time, the patterns in this log reveal:
- How long issues of various types/labels/sizes actually take
- Which dependency structures lead to smooth execution vs. thrashing
- Which epic decomposition patterns produce parallelizable tracks
- How often issues get reopened (quality signal)
- How many remediation cycles workers typically need (complexity signal)
- What the typical velocity is, and whether it's accelerating or decelerating

beads_viewer can compute graph metrics on the current issue set. Beadwork can compute metrics on the entire management trajectory — past, present, and extrapolated future. And because `bw prime` injects this automatically, the learning loop closes without any human or agent intervention.

**This is the difference between a dashboard and an operating system.**

---

## Summary Table

| Capability | beads_viewer | beadwork + pi extension |
|---|---|---|
| Issue provenance | Heuristic reconstruction (0.20-0.85 confidence) | Exact intent log (1.0 confidence) |
| Context delivery | Agent must call `bv --robot-*` | Auto-injected via `before_agent_start` |
| Cycle prevention | Detects after creation | Blocks at creation time |
| Concurrent writes | JSONL merge conflict | CAS + intent replay |
| Worker awareness | None | Full lifecycle tracking (~60 fields) |
| File conflict detection | None | Worktree diff comparison |
| Speculative execution | Impossible (no execution model) | Native (cheap worktrees + landing gate) |
| Landing validation | None | Multi-stage pipeline (lint/test/review) |
| Comment semantics | Opaque text | Typed protocols with cross-issue propagation |
| Feedback loop | One-directional (analyze → recommend) | Closed (state → execute → validate → state) |
| Distribution | Binary + export pipeline + intermediate format | `git clone` + single binary |
| Self-improvement | N/A (no execution data) | Intent log + worker data → velocity/patterns |
| Management process analysis | Current state only | Full operational trajectory |
| Offline capability | Requires recent JSONL export | Always current (no export step) |

---

## The Punchline

beads_viewer is a very good graph-analysis engine that happens to read issue data.

Beadwork is an operational system where the issue data, the execution model, the agent runtime, and the analysis engine are the same thing.

The graph analysis ideas from `bv` are valuable and should be absorbed. But the most interesting things beadwork can build are the ones that only make sense when you have all four of those properties simultaneously — and those are the things beads_viewer could never build even if it tried.
