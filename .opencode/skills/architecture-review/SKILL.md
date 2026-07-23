---
name: architecture-review
description: Use when performing the post-phase architecture review of the codebase — evaluating structure, boundaries, data model, and tech choices against PROJECT.md goals, and writing the report with carry-over recommendations to .opencode/reviews/.
---

# Architecture Review

After a phase fully merges, assess architectural health — not line-level code quality (PR review already did that), but whether the system is evolving toward the blueprint's goals on sound footing.

## Evaluation axes

1. **Blueprint alignment** — does what exists match the phase goal and the overall PROJECT.md trajectory? Any drift that later phases will pay for?
2. **Module boundaries** — are concerns separated (ingestion / persistence / API / UI)? Are dependencies one-directional, or is coupling creeping in?
3. **Data model fit** — does the schema support what later phases need (price history queries, cross-marketplace linking, unit normalization)? Migrations manageable?
4. **Contracts** — are API shapes, internal interfaces, and JSON schemas stable and documented enough for the next phase's tickets to build on?
5. **Cross-cutting concerns** — configuration, error handling strategy, logging, test strategy. Consistent or ad-hoc per ticket?
6. **Technical debt** — shortcuts taken under ticket scope pressure. Classify by interest rate: what compounds vs. what's inert?
7. **Prior carry-overs** — read earlier reports in `.opencode/reviews/`: were previous carry-over items addressed or silently dropped?

## Method

- Read the phase goal in `PROJECT.md`, all phase tickets, and the phase's merged diffs (`git log`/`git show` on the merge commits).
- Read the actual resulting code structure — judge what exists, not what was planned.
- Every finding cites concrete files/modules. No vague "could be cleaner".

## Report — write to `.opencode/reviews/phase-<N>-architecture.md`

```markdown
---
phase: <N>
date: <YYYY-MM-DD>
verdict: sound | sound-with-concerns | needs-remediation
---

# Phase <N> Architecture Review

## Summary
<3–5 sentences: what was built, overall verdict, the single most important observation.>

## Findings
### 1. <title> [severity: concern | risk | debt]
**Observation:** <what exists, with file/module references>
**Impact:** <what it costs later phases if left alone>
**Recommendation:** <concrete action>

### 2. ...

## What went well
<Bullets — patterns worth keeping as conventions.>

## Carry-over for next phase
- [ ] <mandatory input for the next ticket-grooming run: specific, actionable, sized to fit inside a ticket>

## Prior carry-over disposition
<For each carry-over from the previous report: addressed / partially / dropped (why). Omit for phase 1.>
```

## Rules for carry-overs

- Each must be small enough to fold into a future ticket's scope — otherwise split it.
- `needs-remediation` verdict requires at least one carry-over marked as blocking the next phase's grooming.
- You write the report file and a summary as your final message (verdict, top findings, carry-over list). You never change implementation code.
