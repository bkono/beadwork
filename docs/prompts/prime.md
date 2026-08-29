# prime.md

Design requirements for the prime prompt (`prompts/prime.md`).

For end goals, prompt architecture, and experimentation methodology,
see [`prompts.md`](prompts.md).

## Design requirements

1. **Combine principle with procedure.** Principles ("context loss is certain")
   set disposition but don't drive behavior at the point of action. Procedures
   (numbered workflow steps) drive behavior but feel arbitrary without
   rationale. Effective prompts pair a brief *why* with a concrete *what*.
   Tested: principle alone fails, procedure alone fails, both together work.

2. **Be additive, not overriding.** Agents arrive with built-in planning,
   formatting, and task management. Fighting these instincts requires heavy
   rhetorical force and still doesn't translate to interactive behavior.
   Instead, work *with* the agent's natural patterns and add a durability
   step: "plan however you want, then materialize multi-step work as tickets
   before executing." The prompt should augment the agent's flow, not replace
   it.

3. **Make current-branch work the default.** The user has already chosen a
   checkout and branch. Agents should work there by default, including on
   `main`, unless the user asks for isolation or repo config requires a PR
   path. Branches, PRs, and worktrees are delivery modes, not prerequisites
   for using beadwork.

4. **Let delivery intent select ceremony.** Blanket rules ("every change gets
   a ticket" or "every task gets a worktree") activate stochastically — agents
   override them based on their own cost/benefit heuristic. Prime should teach
   a small menu: quick fix, tracked task on the current branch, multi-step /
   swarm epic, or optional branch/PR/worktree mode when requested/configured.

5. **Use numbered lists for workflow.** When the workflow is a numbered
   checklist, agents walk through every step. When compressed to prose,
   individual steps get skipped. Keep orientation, claiming, working, landing,
   and syncing as explicit steps.

6. **Handle dirty state as ownership, not a hard worktree gate.** A dirty
   checkout may contain user work, another agent's progress, or the current
   agent's own edits. The prompt should make the efficient action be:
   identify ownership, preserve unclear changes, and ask before touching them.
   It should not imply that any dirty state requires abandoning the current
   branch or creating a new worktree.

7. **Teach same-branch swarm coordination.** Multiple agents may collaborate on
   one branch when the user wants that. Prime should push them toward child
   tickets, non-overlapping scopes, and `bw comment` breadcrumbs. Isolation via
   separate branches/worktrees remains available when requested or when edits
   would collide.

8. **Stay compact.** Shorter is not just cheaper — it's more effective. Less
   noise means less competition for attention. Keep implementation details and
   setup instructions out of prime.

9. **Keep live state near the action loop.** The ready queue and WIP list make
   the prompt immediately actionable. They should be easy to find and close to
   the workflow they feed.

10. **Adapt to project configuration at the point of action.** Per-task
    conditionals (PR review, etc.) live in start.md and render when the agent
    claims a ticket. Prime teaches the full mental model, including that those
    modes are optional unless selected.

11. **Be the canonical reference.** AGENTS.md is deliberately minimal — just
    a pointer to `bw prime`. This prompt is the single source of truth for how
    to use beadwork in this project.

12. **Land the work.** Prime establishes the principle (unfinished bookkeeping
    is invisible progress); `bw start` delivers concrete steps via start.md.
    The numbered workflow reinforces landing as part of the work: scoped
    commit, close, sync.

13. **Teach delegation concisely.** When orchestrating sub-agents, the
    orchestrator must include workflow steps in the handoff. The key
    information is the sequence (start → scoped work → comment → commit →
    close) and the principle (they don't inherit your context).

14. **Don't fight — augment.** The prime prompt competes with agents' built-in
    instructions (system prompts, plan mode templates). Overriding these
    requires escalating rhetorical force and still often wins format compliance
    rather than behavior. The additive approach sidesteps the conflict: let the
    agent plan in whatever format it wants, then add durable state and landing
    steps.
