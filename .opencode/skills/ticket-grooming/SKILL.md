---
name: ticket-grooming
description: Use when breaking down a project phase, blueprint section, or PROJECT.md phase into small implementation tickets / work units in .opencode/tickets/. Covers ticket sizing rules, dependency analysis, ordering, and the ticket file template.
---

# Ticket Grooming

Distill one phase of `PROJECT.md` into the smallest set of tickets that fully delivers the phase goal.

## Sizing rules

- **One ticket = one PR.** A single focused change a reviewer can evaluate in minutes.
- **Independently verifiable.** Each ticket has acceptance criteria that can be tested right after it merges, without waiting for later tickets.
- **Vertical slices over horizontal layers.** Prefer "schema + query layer for entity X" over "all schemas, then all queries" when the phase allows it.
- **~100–400 changed lines** is the sweet spot. If a ticket drafts bigger, split it. If two tickets are trivially small and tightly coupled, merge them.
- **Explicit seams.** A ticket that introduces an interface/boundary used by later tickets comes first and defines the contract in its acceptance criteria.

## Dependency analysis

- For every ticket, set `depends_on` to the minimal list of ticket ids that must merge first.
- Tickets with disjoint `depends_on` closures are parallelizable — the orchestrator uses this, so be strict and accurate.
- Order tickets so the phase stays mergeable: each merge leaves `main` working (builds + tests pass).

## Carry-overs

If the orchestrator gave you a previous architecture review: its `## Carry-over for next phase` items are mandatory inputs. Either fold each into a ticket's scope/acceptance criteria or explicitly note in your report why it was deferred.

## Procedure

1. Read the full phase text and the phase goal.
2. Read the carry-over section of the previous arch review (if provided).
3. List candidate work units; apply the sizing rules; determine dependencies.
4. Write one file per ticket at `.opencode/tickets/phase-<N>/<NN>-<short-slug>.md` (`NN` = zero-padded execution order).
5. Final message: numbered ticket list (id, title, depends_on) + one sizing rationale line each.

## Ticket template

```markdown
---
id: P<N>-T<NN>
phase: <N>
title: <imperative title, <= 60 chars>
status: pending
depends_on: []
branch: feat/p<n>-t<nn>-<short-slug>
pr: null
source: "PROJECT.md → Phase <N> → <Task ref>"
---

## Objective

<1–3 sentences: what this delivers and why it matters to the phase goal.>

## Scope

**In scope:**
- <concrete deliverable>

**Out of scope:**
- <explicit exclusion, especially tempting adjacent work>

## Acceptance criteria

- [ ] <testable statement — a reviewer can verify each one mechanically>
- [ ] Build passes and relevant tests pass

## Technical notes

<Relevant constraints, file/module pointers, contracts to honor, carry-over items from arch review. Leave empty if none.>

## Review feedback

<Empty at creation. Orchestrator appends PR-review findings here.>
```

## Quality bar for your output

- A stranger could implement each ticket without reading the phase text.
- No ticket requires "and then also" — that is two tickets.
- Acceptance criteria are observable behavior, not aspirations ("endpoint returns X for Y", not "works well").
