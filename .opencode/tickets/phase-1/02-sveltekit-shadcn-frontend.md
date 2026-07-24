---
id: P1-T02
phase: 1
title: Scaffold SvelteKit + Tailwind + shadcn-svelte frontend
status: done
depends_on: []
branch: feat/p1-t02-sveltekit-shadcn-frontend
pr: 6
source: "PROJECT.md → Phase 1 → Task 1.2"
---

## Objective

Bootstrap the SvelteKit frontend application with Tailwind CSS and the
`shadcn-svelte` component system, plus a Vite dev proxy to the backend API so
feature work in later phases can call `/api/*` without CORS friction.

## Scope

**In scope:**
- SvelteKit application in `frontend/` (Node project with `package.json`)
- Tailwind CSS configured and working (content paths, theme tokens)
- `shadcn-svelte` initialized: `components.json`, base theme, CSS variables for theming, utility helpers (`cn`, `utils.ts`)
- One base component added via the CLI (e.g. `Button`) to prove the pipeline
- Vite dev server proxy: requests to `/api/*` forward to the backend (default `http://localhost:8080`)
- Landing route (`/`) rendering a styled `shadcn-svelte` component to verify the stack
- `frontend/README.md` with run instructions

**Out of scope:**
- Feature pages, charts, data tables (Phase 5)
- PWA / service worker / share target (Phase 6)
- Any real API integration beyond the proxy config

## Acceptance criteria

- [ ] `cd frontend && npm install` completes
- [ ] `npm run dev` starts the SvelteKit dev server on its default port
- [ ] A `shadcn-svelte` `Button` renders with correct theme styling on `/`
- [ ] Tailwind utility classes apply (verify a `className`/`class` styled element)
- [ ] `npm run build` produces a production build with no errors
- [ ] Dev proxy configured: a request to `/api/health` is forwarded to `http://localhost:8080` (returns 200 when backend is running; 502/connection error is acceptable when backend is not running — the proxy config itself is the acceptance target)
- [ ] `components.json` present and valid; a second component can be added via `npx shadcn-svelte add` (document the command in README)

## Technical notes

- Parallelizable with P1-T01: this ticket only creates `frontend/` and does not touch the backend.
- Proxy target is the contract with P1-T01: backend listens on `PORT` (default 8080). Configure `vite.config.ts` `server.proxy` for `/api`.
- Use the current SvelteKit + Svelte 5 + Vite tooling (`npx sv create`). Confirm Svelte 5 runes compatibility for `shadcn-svelte` at implementation time and pin versions.
- Tailwind v4 vs v3: `shadcn-svelte` compatibility should drive the choice — pick the combination the CLI currently supports and note it in the README.
- Root `.gitignore` (from P1-T01) already covers `frontend/node_modules` and `frontend/build`; if not, append frontend-specific ignores here.

## Review feedback

**Verdict: APPROVED (PR #6)**

Criteria check: all 7 acceptance criteria met and independently verified (npm install, dev server, shadcn Button with bg-primary theme styling + Tailwind utilities, build, /api proxy -> :8080, components.json valid, scope clean under frontend/).

Findings (non-blocking, nit):
- [nit] frontend/.vscode/{settings.json,extensions.json} — scaffold-generated editor config committed; optional to gitignore .vscode/ in a follow-up.

Notes: Svelte 5 runes mode enabled; adapter-static SPA mode with fallback index.html — revisit if any later route needs prerendering. Proxy key is `/api` (prefix match) which correctly matches `/api/health` etc.

