# Price Tracker — Frontend

SvelteKit application with Tailwind CSS and [shadcn-svelte](https://shadcn-svelte.com) components.

## Prerequisites

- Node.js >= 20
- npm

## Getting started

```bash
cd frontend
npm install
npm run dev
```

The dev server starts at `http://localhost:5173`.

## Build

```bash
npm run build
npm run preview
```

## Dev proxy

Requests to `/api/*` are proxied to the backend at `http://localhost:8080`. Start the backend alongside the frontend for full-stack development.

## Adding shadcn-svelte components

Components are managed via the shadcn-svelte CLI:

```bash
npx shadcn-svelte add <component>
```

Example:

```bash
npx shadcn-svelte add button
```

This installs the component source to `src/lib/components/ui/` so you can edit it directly.

## Project structure

```
src/
├── lib/
│   ├── components/ui/   # shadcn-svelte UI components
│   └── utils.ts         # cn() utility (clsx + tailwind-merge)
├── routes/
│   ├── +layout.svelte   # Root layout
│   ├── +page.svelte     # Landing page
│   └── layout.css       # Tailwind + CSS variables
└── app.html             # HTML shell
```

## Stack

- [SvelteKit](https://svelte.dev/docs/kit) (Svelte 5 with runes)
- [Tailwind CSS v4](https://tailwindcss.com)
- [shadcn-svelte](https://shadcn-svelte.com)
- TypeScript
- Vite
