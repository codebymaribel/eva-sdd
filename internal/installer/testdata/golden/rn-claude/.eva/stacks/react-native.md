# Stack: React Native (Expo)

App root: `.`. All paths below are relative to it.

> Agents: read this file before planning, implementing or verifying any feature with `stack: react-native`. Plans cite checklist rules by id; verification reports each applicable one.
> Developers: the baseline below is opinionated — change it here (and add an ADR) rather than per feature.

## 1. Baseline stack

| Concern | Choice |
|---------|--------|
| Runtime | Expo (latest SDK), Continuous Native Generation, development builds — no Expo Go-only APIs |
| Language | TypeScript `strict` |
| Navigation | Expo Router (file-based) with native stack and native tabs |
| Server state | TanStack Query |
| Client state | Local component state first; Zustand for shared client state (selectors only) |
| Forms & validation | React Hook Form + Zod |
| Lists | FlashList or LegendList (never `ScrollView` + `map` for lists) |
| Animation & gestures | Reanimated + Gesture Handler |
| Images | `expo-image` |
| Styling | `StyleSheet.create` or NativeWind (pick one per project, ADR) |
| Secure storage | `expo-secure-store` for tokens; MMKV for fast non-sensitive storage |
| Unit / component tests | Jest (`jest-expo`) + React Native Testing Library |
| End-to-end tests | Maestro (`.maestro/` flows) |
| Compiler | React Compiler enabled; manual memoization only where the compiler cannot help |

## 2. Commands

Adjust to the project's `package.json` and keep this table accurate — agents run these.

| Purpose | Command |
|---------|---------|
| Typecheck | `npx tsc --noEmit` |
| Lint | `npx expo lint` |
| Unit tests | `npx jest` |
| Single test file | `npx jest path/to/file.test.tsx` |
| E2E | `maestro test .maestro/` |
| Dev server | `npx expo start` |

## 3. Project structure

Feature-first. Routes are thin; logic lives in features.

```
app/                         # Expo Router routes only — screens compose feature components
  _layout.tsx
  (tabs)/_layout.tsx
  (tabs)/index.tsx
src/
  features/<feature>/
    components/              # feature UI (PascalCase.tsx) + tests
    hooks/                   # useThing.ts + tests
    api/                     # query/mutation hooks and fetchers for this feature
    store/                   # Zustand slice, only if shared client state is needed
    schemas.ts               # Zod schemas
    types.ts
  ui/                        # design-system primitives (Button, ButtonText, Text, …)
  lib/                       # api client, storage, analytics, utils
  theme/                     # tokens: colours, spacing, typography
assets/
.maestro/                    # e2e flows: <feature>-<journey>.yaml
```

Rules:

- A feature may import from `src/ui`, `src/lib`, `src/theme` and its own folder. Importing another feature's file is allowed only if the plan's file map lists it; frequent cross-feature imports mean the code belongs in `src/lib` or `src/ui`.
- No barrel `index.ts` files; import the file that defines the symbol.
- Native dependencies are installed in the app package (`monorepo-native-deps-in-app`) with one version across the repo (`monorepo-single-dependency-versions`).

## 4. Performance budgets

Defaults — adjust per project and keep them measurable.

| Metric | Budget |
|--------|--------|
| Scrolling and animations | 60 fps on a mid-range Android device, no JS-thread frame drops during scroll |
| Cold start to interactive | ≤ 2 s on a mid-range Android device (release build) |
| Screen transition | Native navigator; no blocking work in the first render of a new screen |
| JS bundle | No new dependency > 50 kB minified without a note in the plan |
| Re-renders | Typing or scrolling does not re-render unrelated components |

## 5. Performance checklist

Based on Vercel's React Native skills (`vercel-labs/agent-skills` → `react-native-skills`). Cite rules by id in plans and verification.

### Critical — lists

- [ ] `list-performance-virtualize` — Any list uses FlashList/LegendList, even short ones.
- [ ] `list-performance-item-memo` — List items are memoized components receiving primitive props.
- [ ] `list-performance-callbacks` — Callbacks passed to items are stable references.
- [ ] `list-performance-inline-objects` — No inline style objects or object literals in item props.
- [ ] `list-performance-function-references` — Functions are defined outside render, not recreated per item.
- [ ] `list-performance-images` — Images in lists are sized thumbnails via `expo-image` with caching.
- [ ] `list-performance-item-expensive` — Expensive work (formatting, sorting) happens before the list, not inside items.
- [ ] `list-performance-item-types` — Heterogeneous lists declare item types.

### High — animation, scroll, navigation, UI

- [ ] `animation-gpu-properties` — Animate only `transform` and `opacity`.
- [ ] `animation-derived-value` — Computed animation values use `useDerivedValue`.
- [ ] `animation-gesture-detector-press` — Animated press feedback uses `Gesture.Tap`.
- [ ] `scroll-position-no-state` — Scroll position never lives in `useState`; use shared values or refs.
- [ ] `state-ground-truth` — State stores facts (`pressed`, `progress`), visuals are derived.
- [ ] `navigation-native-navigators` — Native stack / native tabs only, no JS navigators.
- [ ] `ui-expo-image` — All images use `expo-image`.
- [ ] `ui-pressable` — `Pressable`, not `TouchableOpacity`.
- [ ] `ui-safe-area-scroll` / `ui-scrollview-content-inset` — Safe areas handled via content insets, not wrapper padding.
- [ ] `ui-native-modals` / `ui-menus` — Native modals and context menus when available.
- [ ] `ui-measure-views` — Layout read with `onLayout`, not `measure()`.
- [ ] `ui-styling` — `StyleSheet.create` or NativeWind; no ad-hoc inline styles on hot paths.

### Medium — state and rendering

- [ ] `react-state-minimize` — Components subscribe to the smallest slice of state (selectors, derived booleans).
- [ ] `react-state-dispatcher` — Pass dispatchers/stable callbacks down instead of many handlers.
- [ ] `react-state-fallback` — `undefined` initial state + `??` fallback instead of syncing props into state.
- [ ] `react-compiler-destructure-functions` / `react-compiler-reanimated-shared-values` — Code is React Compiler friendly.
- [ ] `rendering-text-in-text-component` — All strings render inside `<Text>`.
- [ ] `rendering-no-falsy-and` — No `cond && <X/>` with possibly falsy numbers/strings; use ternaries.
- [ ] `design-system-compound-components` — Components with text use compound parts (`Button` + `ButtonText`).

### Low — configuration

- [ ] `fonts-config-plugin` — Custom fonts loaded via config plugin, not at runtime.
- [ ] `imports-design-system-folder` — Design-system imports come from `src/ui`.
- [ ] `js-hoist-intl` — `Intl.*` formatters are created once at module level.

### Data and scalability

- [ ] Server data goes through TanStack Query with explicit query keys; no duplicated fetches across screens.
- [ ] Lists backed by APIs are paginated or infinite; no unbounded fetches.
- [ ] Heavy screens and rarely used features load lazily.
- [ ] No work on the JS thread > 16 ms during interactions (use `InteractionManager` or move work off the critical path).

## 6. Testing conventions

- Component tests with React Native Testing Library: query by role, label or text (`getByRole`, `getByLabelText`), not by test id unless unavoidable.
- Mock the network at the API client boundary; never hit real services.
- Use fake timers for time-based behaviour.
- Test names start with the requirement: `it('FR-3: disables submit while saving', …)`.
- Critical journeys have a Maestro flow in `.maestro/` named `<feature>-<journey>.yaml`.

## 7. Accessibility

- Every touchable has `accessibilityRole` and an `accessibilityLabel` when the content is not descriptive text.
- Support dynamic type (no fixed heights that clip text); minimum touch target 44×44 pt.
- Respect "Reduce Motion" for non-essential animations.
