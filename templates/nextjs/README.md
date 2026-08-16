# Next.js Starter

A basic Next.js App Router project with TypeScript, Tailwind CSS v4, shadcn/ui, TanStack Query, Axios, and Zustand.

## Run locally

```bash
npm run dev
```

Or with pnpm:

```bash
pnpm run dev
```

Open [http://localhost:3000](http://localhost:3000), then edit `src/app/page.tsx`.

## Data fetching and state

- `src/providers/query-provider.tsx` configures TanStack Query for the App Router.
- `src/lib/api.ts` exports an Axios client using `NEXT_PUBLIC_API_URL` or `/api` by default.
- `src/stores/use-counter-store.ts` provides a small Zustand store.
- `src/components/starter-panel.tsx` demonstrates both integrations against `GET /api/status`.

## Add shadcn/ui components

```bash
npx shadcn@latest add input
```

Or with pnpm:

```bash
pnpm dlx shadcn@latest add input
```
