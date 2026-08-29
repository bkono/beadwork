{{/* See docs/prompts/start.md */}}
{{ if eq .Type "epic" }}{{ bw "show" .ID "--only" "children" }}

## STARTING THE WORK
{{ if eq .WorkflowReview "pr" }}  - If this repo expects PR review, push the current branch and open or update a draft PR (`{{ .ID }}: <title>`) to preserve progress.
{{ end }}  - Work through this epic's children — use `bw ready` to find the next unblocked child.
  - Keep coordination notes on the epic or child issues with `bw comment`.

## LANDING THE WORK
{{ if eq .WorkflowReview "pr" }}  - Push the current branch and convert the PR to ready for review when the epic is complete.
{{ end }}  - Close the epic (`bw close {{ .ID }}`) when all children are done.
  - `bw sync`.
{{ end }}{{ if or (eq .Type "task") (eq .Type "bug") }}{{ bw "show" .ID "--only" "blockedby,unblocks" }}
{{ if eq .WorkflowReview "pr" }}
## STARTING THE WORK
  - Work on the current branch. If this will take multiple sessions, push it and open or update a draft PR early.
{{ end }}
## LANDING THE WORK
  Land this ticket before starting the next one:
  - Commit only the changes for this ticket, referencing {{ .ID }}.
{{ if eq .WorkflowReview "pr" }}  - Push the current branch and open a PR referencing {{ .ID }}. If a draft PR already exists, convert it to ready for review.
{{ end }}  - Close the ticket (`bw close {{ .ID }}`); it will tell you if work is newly unblocked.
  - `bw sync`.
{{ end }}
