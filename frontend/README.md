# React 18 SaaS Dashboard Template

A production-ready, modern SaaS dashboard template built with React and Vite.

## 🚀 Tech Stack

- **Framework**: [React 18](https://react.dev/)
- **Build Tool**: [Vite](https://vitejs.dev/)
- **Language**: [TypeScript](https://www.typescriptlang.org/)
- **Styling**: [Tailwind CSS v4](https://tailwindcss.com/)
- **UI Components**: [PrimeReact](https://primereact.org/) (Unstyled mode / Aura Theme)
- **State Management**: [Zustand](https://zustand-demo.pmnd.rs/)
- **Routing**: [React Router v6](https://reactrouter.com/)
- **Icons**: [PrimeIcons](https://primefaces.org/primeicons/)

## 📂 Modular Architecture

This template uses a domain-driven, feature-sliced architecture designed for scalability.

```text
src/
├── api/             # API clients and endpoints grouped by domain (e.g., users.ts)
├── assets/          # Static assets (images, global CSS)
├── components/      # Global shared UI components
│   └── layout/      # Layout-specific components (Sidebar, TopNav)
├── hooks/           # Reusable React hooks (e.g., useDarkMode)
├── features/        # Domain-specific components grouped by feature (e.g., users/)
├── layouts/         # Page layout wrappers (DashboardLayout, AuthLayout)
├── pages/           # Route-level React components (Views)
├── stores/          # Zustand state management stores
├── types/           # TypeScript interfaces and types
└── utils/           # Helper functions and formatters (e.g. cn() utility)
```

> **Note on Barrel Exports**: Every major directory contains an `index.ts` (Barrel Export). This allows you to import multiple items from a directory cleanly, e.g., `import { formatCurrency, cn } from '@/utils'`.

## 🛠️ Getting Started

### 1. Install Dependencies
```bash
npm install
```

### 2. Run Development Server
```bash
npm run dev
```

### 3. Build for Production
```bash
npm run build
```

## 🎨 Design System

The application uses a custom set of CSS variables (`src/index.css`) that are fully integrated with Tailwind CSS v4. It features a premium, clean aesthetic inspired by modern SaaS applications (Linear, Vercel, Stripe), complete with dark mode support.
