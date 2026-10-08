# AGENTS.md — React Template Instructions for AI Coding Agents

This document is read automatically by opencode, Claude Code, Cursor, Antigravity, and other AI tools. Follow all guidelines below **without exception** when writing or modifying code in this React template.

---

## 1. Stack Overview

- **Framework**: React 18 / 19 (`react`, `react-dom`)
- **Build Tool**: Vite (`@vitejs/plugin-react`)
- **Language**: TypeScript (Strict mode, `@/` path alias mapped to `src/`)
- **Styling**: Tailwind CSS v4 (`@tailwindcss/vite`, `tailwindcss`) + custom CSS variables (`src/index.css`)
- **UI Components**: PrimeReact (Unstyled mode / Aura Theme `@primeuix/themes/aura`) + `primeicons`
- **State Management**: Zustand (`zustand`)
- **Routing**: React Router v7 (`react-router-dom`) with `createBrowserRouter` & `RouterProvider`
- **Logging**: Structured logger via `logger` from `@/utils` (Consola)

---

## 2. Mandatory Commands & Quality Gate

- `npm run build` — `tsc -b && vite build` — **MUST pass cleanly** before finishing any task.
- `npm run lint` — `eslint . --fix` — **MUST pass with ZERO errors**.
- `npm run format` — `prettier --write src/` — Formats source code according to repo configuration.
- `npm run dev` — Launches Vite development server.

---

## 3. Architecture & Modular Layout

This project follows a domain-driven, feature-sliced modular architecture:

```text
src/
├── api/             # API client & domain endpoint functions (e.g., users.ts, client.ts)
├── assets/          # Static assets & global CSS
├── components/      # Global shared UI components
│   └── layout/      # Shared layout components (Sidebar, TopNav)
├── hooks/           # Reusable React hooks (e.g., useDarkMode)
├── features/        # Feature modules containing domain-specific components (e.g., features/users/)
├── layouts/         # Page layout wrappers (DashboardLayout, AuthLayout)
├── pages/           # Route-level components / Views
├── stores/          # Zustand global state management stores
├── types/           # TypeScript interfaces, types & DTOs
└── utils/           # Shared helper functions, formatters (cn utility), and logger
```

### Key Architectural Rules:
- **Barrel Exports**: Every major directory (`api`, `components`, `features`, `hooks`, `layouts`, `pages`, `stores`, `types`, `utils`) MUST contain an `index.ts` file exporting its public interface.
- **Import Paths**: Always use path alias `@/` for imports from `src/` (e.g., `import { logger, cn } from '@/utils'`, `import { useAuthStore } from '@/stores'`).
- **Feature Isolation**: Place domain-specific components inside `src/features/<feature>/`. Do not bloat `src/components/` with feature-specific code.

---

## 4. Code & Style Conventions

### Component & JSX Conventions
- Use functional components with explicit TypeScript interfaces for props.
- Use self-closing tags for elements without children (`<div />`, `<Button />`).
- Prefer boolean props short syntax (`<Modal isOpen />` instead of `<Modal isOpen={true} />`).
- Avoid using array indices as `key` props in list rendering unless items have no stable ID.

### TypeScript & Imports
- **Type-only imports**: Use explicit type imports: `import type { User } from '@/types'` (`@typescript-eslint/consistent-type-imports`).
- **Strict Typing**: Avoid `any`. Define proper interfaces/types in `src/types/`.
- **Import Sorting**: `eslint-plugin-simple-import-sort` is enforced. Do not fight import order; run `npm run lint` to fix automatically.
- **Unused Imports**: Unused imports/variables are disallowed unless prefixed with `_`.

### Logging & Side Effects
- **DO NOT** use `console.log()`. Use the app logger: `import { logger } from '@/utils'; logger.info(...)`.
- Keep side effects inside custom hooks or event handlers; do not perform side-effects directly in render functions.

### State Management (Zustand)
- Store files live in `src/stores/`.
- Export typed store hooks, e.g. `useAuthStore`.

---

## 5. Definition of Done

Before marking any task complete:
1. [ ] `npm run lint` passes without errors.
2. [ ] `npm run build` compiles without TypeScript or Vite errors.
3. [ ] All imports are clean, sorted, and use path alias `@/`.
4. [ ] No `console.log` statements remain in code (use `logger` from `@/utils`).
5. [ ] Barrel export `index.ts` is updated if new modules or components are added.
