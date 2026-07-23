---
name: ticket-implementation
description: Use when implementing a work ticket from .opencode/tickets/ — branch/worktree setup, commit style, running builds and tests, addressing PR review feedback, and opening the pull request with gh.
---

# Ticket Implementation

Turn one ticket into one clean pull request.

## 1. Set up

- Read the ticket file completely, including `## Review feedback` (present on respawns) and `## Technical notes`.
- Base work on current `main`:
  - Sequential run (main checkout): `git fetch origin && git checkout -b <branch> origin/main`
  - Parallel run (orchestrator assigned a worktree): `git fetch origin && git worktree add <worktree-path> -b <branch> origin/main`, then work **only inside that worktree**.
- `<branch>` is the ticket's frontmatter `branch` value. Never commit on `main`.

## 2. Implement

- Stay inside the ticket's **In scope**. Respect **Out of scope** literally.
- Follow existing project conventions (read neighboring code first). Match the stack in `PROJECT.md`: Go backend, SvelteKit + Tailwind + shadcn-svelte frontend, SQLite.
- Small, coherent commits with imperative messages (`Add receipt schema migrations`, `Wire upload endpoint to vision pipeline`).
- If review feedback exists: address every finding, or justify the exception in the PR body.

## 3. Verify

Before opening the PR, run the relevant checks and make them pass:

- Go: `go build ./... && go test ./...` (and `go vet ./...` when touching Go)
- Frontend: the project's lint/build commands (e.g. `npm run check`, `npm run build`)
- Every acceptance criterion: verify it concretely (run it, query it, or test it — not by inspection).

## 4. Open the PR

```
gh pr create --title "P<N>-T<NN>: <ticket title>" --body <body>
```

PR body template:

```markdown
## Ticket
P<N>-T<NN> — <title> (`.opencode/tickets/phase-<N>/<file>`)

## What changed
- <bullet per logical change>

## Acceptance criteria verification
- [x] <criterion> — <how it was verified: command/test/output>

## Review feedback addressed
<Omit on first submission. On respawn: one bullet per finding → how it was addressed, or why not.>

## Out of scope / follow-ups
<Anything discovered but deliberately not done. Omit if empty.>
```

## 5. Report back

Final message to the orchestrator with: branch name, PR URL, implementation summary (bullets), exact verification commands + results, out-of-scope observations. Do not edit the ticket's status — the orchestrator owns state.

## Never

- No merges, no force-push, no rebasing onto anything but `origin/main`, no edits to other tickets or ticket statuses.
- No unrequested refactors of code outside the ticket's blast radius.
