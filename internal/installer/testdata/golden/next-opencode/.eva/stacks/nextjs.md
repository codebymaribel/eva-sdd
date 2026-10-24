# Stack: Next.js (App Router)

App root: `.`. All paths below are relative to it.

> Agents: read this file before planning, implementing or verifying any feature with `stack: nextjs`. Plans cite checklist rules by id; verification reports each applicable one.
> Developers: the baseline below is opinionated — change it here (and add an ADR) rather than per feature.

## 1. Baseline stack

| Concern | Choice |
|---------|--------|
| Framework | Next.js (latest), App Router, `src/` directory |
| Language | TypeScript `strict` |
| Rendering | React Server Components by default; `'use client'` only at the leaves that need interactivity |
| Mutations | Server Actions (authenticated and validated) or Route Handlers for external clients |
| Server data | Fetched in Server Components; deduplicated with `React.cache()` |
| Client data | SWR for client-side fetching that cannot happen on the server |
| Client state | Local state first; URL search params for shareable state; a store only with an ADR |
| Validation | Zod, shared between client and server |
| Styling | Tailwind CSS |
| Images & fonts | `next/image`, `next/font` |
| Unit / component tests | Vitest + React Testing Library |
| End-to-end tests | Playwright (`e2e/`) |

## 2. Commands

Adjust to the project's `package.json` and keep this table accurate — agents run these.

| Purpose | Command |
|---------|---------|
| Typecheck | `npx tsc --noEmit` |
| Lint | `npx eslint .` |
| Unit tests | `npx vitest run` |
| Single test file | `npx vitest run path/to/file.test.tsx` |
| E2E | `npx playwright test` |
| Build (checks route output and bundle sizes) | `npx next build` |
| Dev server | `npx next dev` |

## 3. Project structure

Feature-first. Route files are thin; logic lives in features.

```
src/
  app/                         # routing only: page, layout, loading, error, not-found, route handlers
    (marketing)/page.tsx
    (app)/dashboard/page.tsx
    api/<name>/route.ts
  features/<feature>/
    components/                # Server Components by default (PascalCase.tsx) + tests
    components/client/         # 'use client' components, kept small
    actions.ts                 # 'use server' — Server Actions (auth + Zod validation)
    queries.ts                 # import 'server-only' — data access, wrapped in React.cache()
    schemas.ts                 # Zod schemas shared by client and server
    types.ts
  components/ui/               # design-system primitives
  lib/                         # db/api clients (server-only), auth, utils
e2e/                           # Playwright specs: <feature>.spec.ts
```

Rules:

- `queries.ts` and anything touching secrets import `'server-only'`.
- A feature may import from `components/ui`, `lib` and its own folder. Importing another feature's file is allowed only if the plan's file map lists it.
- No barrel `index.ts` files (`bundle-barrel-imports`); import the file that defines the symbol.
- Each route's rendering and caching strategy (static, dynamic, revalidated, cached data) is stated in the plan.

## 4. Performance budgets

Defaults — adjust per project and keep them measurable (field data at p75, or Lighthouse on a throttled mobile profile).

| Metric | Budget |
|--------|--------|
| Largest Contentful Paint (LCP) | ≤ 2.5 s |
| Interaction to Next Paint (INP) | ≤ 200 ms |
| Cumulative Layout Shift (CLS) | ≤ 0.1 |
| First-load JS per route | ≤ 150 kB gzipped (check `next build` output) |
| Server response (TTFB) for dynamic routes | ≤ 600 ms |

## 5. Performance checklist

Based on Vercel's React best practices (`vercel-labs/agent-skills` → `react-best-practices`). Cite rules by id in plans and verification.

### Critical — eliminating waterfalls

- [ ] `async-parallel` — Independent awaits run with `Promise.all()`.
- [ ] `async-defer-await` — `await` moved into the branch that actually needs the value.
- [ ] `async-cheap-condition-before-await` — Cheap synchronous checks happen before remote awaits.
- [ ] `async-api-routes` — Route handlers start promises early and await late.
- [ ] `async-suspense-boundaries` — Slow data streams behind `<Suspense>` instead of blocking the page.

### Critical — bundle size

- [ ] `bundle-barrel-imports` — Direct imports, no barrel files.
- [ ] `bundle-dynamic-imports` — Heavy, below-the-fold or interaction-only components use `next/dynamic`.
- [ ] `bundle-defer-third-party` — Analytics/logging load after hydration.
- [ ] `bundle-conditional` — Feature-gated modules load only when the feature is active.
- [ ] `bundle-preload` — Likely next interactions preload on hover/focus.
- [ ] `bundle-analyzable-paths` — Imports and file paths are statically analyzable.

### High — server

- [ ] `server-auth-actions` — Every Server Action authenticates and authorises like an API route.
- [ ] `server-cache-react` — Per-request deduplication with `React.cache()`.
- [ ] `server-cache-lru` — Cross-request caching is explicit (LRU or framework cache) where data allows.
- [ ] `server-serialization` / `server-dedup-props` — Only the fields the client needs cross the RSC boundary, once.
- [ ] `server-parallel-fetching` / `server-parallel-nested-fetching` — Components are structured so fetches run in parallel.
- [ ] `server-no-shared-module-state` — No mutable module-level state holding request data.
- [ ] `server-hoist-static-io` — Static I/O (fonts, logos, config) hoisted to module level.
- [ ] `server-after-nonblocking` — Logging/analytics on the server use `after()`.

### Medium-high — client data

- [ ] `client-swr-dedup` — Client fetching goes through SWR (deduplicated, cached).
- [ ] `client-event-listeners` / `client-passive-event-listeners` — Global listeners are deduplicated; scroll/touch listeners are passive.
- [ ] `client-localstorage-schema` — `localStorage` data is versioned and minimal.

### Medium — re-renders and rendering

- [ ] `rerender-no-inline-components` — No components defined inside components.
- [ ] `rerender-derived-state-no-effect` — Derived values computed during render, not synced with effects.
- [ ] `rerender-defer-reads` / `rerender-derived-state` — Subscribe to derived booleans; don't subscribe to state used only in callbacks.
- [ ] `rerender-move-effect-to-event` — Interaction logic lives in event handlers, not effects.
- [ ] `rerender-transitions` / `rerender-use-deferred-value` — Non-urgent updates use transitions; expensive renders are deferred.
- [ ] `rerender-functional-setstate` / `rerender-lazy-state-init` — Stable callbacks and lazy initial state.
- [ ] `rendering-conditional-render` — Ternaries instead of `&&` for conditional rendering.
- [ ] `rendering-content-visibility` — Long lists/sections use `content-visibility` or virtualization.
- [ ] `rendering-hydration-no-flicker` / `rendering-hydration-suppress-warning` — Client-only data does not cause hydration flicker.
- [ ] `rendering-resource-hints` / `rendering-script-defer-async` — Critical resources preloaded; scripts deferred.

### Low — JavaScript and advanced

- [ ] `js-set-map-lookups` / `js-index-maps` — Repeated lookups use `Set`/`Map`.
- [ ] `js-combine-iterations` / `js-early-exit` — Avoid redundant passes over large data.
- [ ] `js-hoist-regexp` — RegExps created outside loops.
- [ ] `advanced-init-once` — App-wide initialisation runs once per load.

### Web fundamentals

- [ ] Images use `next/image` with explicit `sizes`; the LCP image has `priority`.
- [ ] Fonts use `next/font`; no layout shift from late-loading fonts or ads/embeds.
- [ ] Metadata (`generateMetadata`) for every public route.

## 6. Testing conventions

- Component tests with React Testing Library: query by role/label/text, not by test id unless unavoidable.
- Server Actions and `queries.ts` are tested as plain async functions with the data layer mocked.
- Mock the network (MSW or the API client boundary); never hit real services.
- Test names start with the requirement: `it('FR-2: shows an inline error for invalid email', …)`.
- Critical journeys have a Playwright spec in `e2e/<feature>.spec.ts`.

## 7. Accessibility

- Semantic HTML first; ARIA only when no native element fits.
- Every form control has a visible label; errors are announced (`aria-live` or `aria-describedby`).
- Full keyboard navigation with visible focus; colour contrast WCAG AA.
