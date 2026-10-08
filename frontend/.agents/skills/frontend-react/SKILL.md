---
name: frontend-react
description: >-
  Architectural guidelines, component patterns, Tailwind CSS v4, PrimeReact, Zustand state management, and quality gates for the React template. Use when building or modifying React components, pages, stores, hooks, or layouts.
---

# React Frontend Template Skill Guide

This skill provides standard procedures, rules, and best practices for developing and maintaining the **React + Vite** frontend template.

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

## 2. Directory Structure & Feature-Sliced Architecture

Code must be strictly organized by layer and domain feature:

```text
src/
├── api/             # API client & domain endpoint functions (e.g., users.ts, client.ts)
├── assets/          # Static assets & global CSS
├── components/      # Global shared UI components
│   └── layout/      # Shared layout components (Sidebar, TopNav)
├── hooks/           # Reusable React hooks (e.g., useDarkMode)
├── features/        # Feature modules with domain-specific components (e.g., features/users/)
├── layouts/         # Page layout wrappers (DashboardLayout, AuthLayout)
├── pages/           # Route-level page components / Views
├── stores/          # Zustand global state management stores
├── types/           # TypeScript interfaces, types & DTOs
└── utils/           # Shared helper functions, formatters (cn utility), and logger
```

### Golden Rules:

1. **Barrel Exports (`index.ts`)**: Every major directory (`api`, `components`, `features`, `hooks`, `layouts`, `pages`, `stores`, `types`, `utils`) MUST maintain an `index.ts` file exporting its public symbols.
2. **Import Alias (`@/`)**: Always use path alias `@/` mapped to `src/` (e.g., `import { logger, cn } from '@/utils'`, `import { useAuthStore } from '@/stores'`). Never use relative traversal like `../../`.
3. **Feature Isolation**: Place domain-specific UI components inside `src/features/<feature>/`. Do not bloat `src/components/` with feature-specific code.

---

## 3. Code & Style Conventions

### Component Patterns & JSX

- Use functional components with explicit TypeScript interfaces for props.
- Use self-closing tags for elements without children (`<div />`, `<Button />`).
- Prefer boolean props short syntax (`<Modal isOpen />` instead of `<Modal isOpen={true} />`).
- Avoid using array indices as `key` props in list rendering unless items have no unique ID.

### TypeScript & Imports

- **Type-only imports**: Use explicit type imports: `import type { User } from '@/types'` (`@typescript-eslint/consistent-type-imports`).
- **Strict Typing**: Avoid `any`. Define proper interfaces/types in `src/types/`.
- **Import Sorting**: `eslint-plugin-simple-import-sort` is enforced. Run `npm run lint` to fix automatically.
- **Unused Imports**: Disallowed unless prefixed with `_`.

### Logging & Side Effects

- **DO NOT** use `console.log()`. Use the app logger:
  ```typescript
  import { logger } from '@/utils'
  logger.info('User action performed', { userId: 123 })
  ```

### State Management (Zustand)

```typescript
// src/stores/authStore.ts
import { create } from 'zustand'

interface AuthState {
  token: string | null
  setToken: (token: string | null) => void
}

export const useAuthStore = create<AuthState>((set) => ({
  token: null,
  setToken: (token) => set({ token })
}))
```

---

## 4. Quality Gate & Mandatory Commands

All changes must satisfy the quality gate before completion:

```bash
# Run dev server
npm run dev

# Check linter and auto-fix
npm run lint

# Check formatting
npm run format

# Full production build and type check (MUST PASS with 0 errors)
npm run build
```
