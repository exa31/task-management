---
name: frontend-add-feature
description: >-
  Step-by-step procedure to scaffold a new domain feature in the React frontend template using Feature-Sliced architecture, Zustand stores, and PrimeReact components. Use when creating new UI features or views.
---

# Scaffolding a New React Feature

This skill guides you through implementing a new domain feature in the **React + Vite** frontend template following Feature-Sliced Architecture.

---

## Feature-Sliced Directory Workflow

When adding a feature `<feature>` (e.g. `k8s`, `users`, `products`):

```text
src/
├── types/<feature>.ts           # TypeScript domain types & API response DTOs
├── api/<feature>.ts             # Axios / HTTP client API methods
├── stores/<feature>Store.ts     # Zustand state store
├── features/<feature>/          # Domain-specific UI components
│   ├── <Feature>List.tsx        # List / Data table view component
│   └── index.ts                 # Barrel export of feature components
└── pages/<feature>/             # Route-level page wrapper
    └── <Feature>Page.tsx
```

---

## Step-by-Step Implementation

### Step 1: Define Types (`src/types/<feature>.ts`)
```typescript
export interface <Entity> {
  id: string | number;
  name: string;
  status: 'active' | 'inactive' | 'pending';
  createdAt: string;
}

export interface Create<Entity>Payload {
  name: string;
  status: string;
}

export interface <Entity>FilterParams {
  status?: string;
  limit?: number;
  offset?: number;
}
```
*Always re-export in `src/types/index.ts`:*
```typescript
export * from './<feature>';
```

### Step 2: Implement API Client (`src/api/<feature>.ts`)
```typescript
import { apiClient } from './client';
import type { <Entity>, Create<Entity>Payload, <Entity>FilterParams } from '@/types';

export const <feature>Api = {
  getAll: (params?: <Entity>FilterParams) =>
    apiClient.get<<Entity>[]>('/api/v1/<feature>', { params }),

  getById: (id: string | number) =>
    apiClient.get<<Entity>>(`/api/v1/<feature>/${id}`),

  create: (payload: Create<Entity>Payload) =>
    apiClient.post<<Entity>>('/api/v1/<feature>', payload),

  delete: (id: string | number) =>
    apiClient.delete(`/api/v1/<feature>/${id}`),
};
```
*Always re-export in `src/api/index.ts`:*
```typescript
export * from './<feature>';
```

### Step 3: Implement Zustand Store (`src/stores/<feature>Store.ts`)
```typescript
import { create } from 'zustand';
import { <feature>Api } from '@/api';
import type { <Entity>, Create<Entity>Payload } from '@/types';
import { logger } from '@/utils';

interface <Feature>State {
  items: <Entity>[];
  loading: boolean;
  error: string | null;
  fetchAll: () => Promise<void>;
  createItem: (payload: Create<Entity>Payload) => Promise<<Entity>>;
}

export const use<Feature>Store = create<<Feature>State>((set) => ({
  items: [],
  loading: false,
  error: null,

  fetchAll: async () => {
    set({ loading: true, error: null });
    try {
      const response = await <feature>Api.getAll();
      set({ items: response.data, loading: false });
    } catch (err: any) {
      logger.error('Failed to fetch <feature> items', err);
      set({ error: err.message || 'Failed to fetch items', loading: false });
    }
  },

  createItem: async (payload) => {
    set({ loading: true });
    try {
      const response = await <feature>Api.create(payload);
      set((state) => ({ items: [...state.items, response.data], loading: false }));
      return response.data;
    } catch (err: any) {
      set({ loading: false });
      throw err;
    }
  },
}));
```
*Always re-export in `src/stores/index.ts`:*
```typescript
export * from './<feature>Store';
```

### Step 4: Create Feature Component (`src/features/<feature>/<Feature>List.tsx`)
```tsx
import React, { useEffect } from 'react';
import { DataTable } from 'primereact/datatable';
import { Column } from 'primereact/column';
import { Tag } from 'primereact/tag';
import { use<Feature>Store } from '@/stores';
import type { <Entity> } from '@/types';

export const <Feature>List: React.FC = () => {
  const { items, loading, fetchAll } = use<Feature>Store();

  useEffect(() => {
    fetchAll();
  }, [fetchAll]);

  const statusTemplate = (rowData: <Entity>) => (
    <Tag
      value={rowData.status}
      severity={rowData.status === 'active' ? 'success' : 'warning'}
    />
  );

  return (
    <div className="rounded-xl border border-border bg-card p-4 shadow-xs">
      <DataTable value={items} loading={loading} stripedRows responsiveLayout="scroll">
        <Column field="name" header="Name" sortable />
        <Column field="status" header="Status" body={statusTemplate} />
        <Column field="createdAt" header="Created At" />
      </DataTable>
    </div>
  );
};
```

### Step 5: Barrel Export (`src/features/<feature>/index.ts`)
```typescript
export * from './<Feature>List';
```

### Step 6: Create Page View & Register Route
Create `src/pages/<feature>/<Feature>Page.tsx`:
```tsx
import React from 'react';
import { <Feature>List } from '@/features/<feature>';

export const <Feature>Page: React.FC = () => {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-foreground"><Feature> Management</h1>
        <p className="text-sm text-muted-foreground">Manage and view all <feature> instances.</p>
      </div>
      <<Feature>List />
    </div>
  );
};
```

---

## Quality Gate Checklist

Before finishing:
1. `npm run lint` passes with 0 errors.
2. `npm run build` (`tsc -b && vite build`) compiles with 0 errors.
3. Path aliases strictly use `@/` (no `../../`).
4. All barrel export `index.ts` files are updated.
