---
description: Mandatory frontend internationalization (i18n) and locale usage for all React UI text, toasts, and dialogs
globs: ["frontend/src/**/*.tsx", "frontend/src/**/*.ts"]
---

# Frontend Internationalization (i18n) & Locale Rules

All user-facing text in the React frontend application must strictly use the internationalization system (`useTranslation()`). Hardcoded strings in JSX, toasts, modals, and dialogs are strictly prohibited.

---

## 1. Absolute Prohibition of Hardcoded Display Strings

- ❌ **Forbidden**:
  ```tsx
  // Bad: hardcoded Indonesian or English text in JSX
  <h1>Pengaturan Akun & Modul</h1>
  <button>Simpan Perubahan</button>
  <input placeholder="Ketik DELETE untuk konfirmasi" />
  toast.success('Akun Google berhasil ditautkan.')
  confirmDialog({ header: 'Hapus Modul', message: 'Hapus project ini?' })
  ```

- ✅ **Required**:
  ```tsx
  // Good: localized through useTranslation hook
  const { t } = useTranslation()

  <h1>{t('settings.title')}</h1>
  <button>{t('settings.profile.saveChanges')}</button>
  <input placeholder={t('settings.dangerZone.typeDeletePrompt')} />
  toast.success(t('settings.oauth.googleLinkedSuccess'))
  confirmDialog({
    header: t('settings.projects.deleteHeader'),
    message: t('settings.projects.deleteMessage', { name: proj.name }),
    acceptLabel: t('settings.projects.deleteAccept'),
    rejectLabel: t('settings.projects.deleteReject'),
  })
  ```

---

## 2. Using the `useTranslation` Hook

Import `useTranslation` directly from `@/i18n`:

```tsx
import { useTranslation } from '@/i18n'

export const MyComponent: React.FC = () => {
  const { t, locale, setLocale, toggleLocale } = useTranslation()
  // ...
}
```

### Parameter Interpolation
When injecting dynamic values (names, numbers, identifiers), use named parameters with `{param}` in the locale definition:

- **Locale dictionary**:
  ```ts
  // en
  tasksConnected: '{count} tasks connected',
  deleteMessage: 'Delete project "{name}"? Linked tasks will not be deleted.',

  // id
  tasksConnected: '{count} task terhubung',
  deleteMessage: 'Hapus project "{name}"? Task yang terhubung tidak akan dihapus.',
  ```

- **Component usage**:
  ```tsx
  <span>{t('settings.projects.tasksConnected', { count: proj.task_count ?? 0 })}</span>
  <p>{t('settings.projects.deleteMessage', { name: proj.name })}</p>
  ```

---

## 3. Strict Synchronization of English (`en`) and Indonesian (`id`)

1. **Both Locales Must Be Updated Together**:
   Whenever a new feature, section, or key is introduced, define it in both:
   - `frontend/src/i18n/locales/en/<feature>.ts`
   - `frontend/src/i18n/locales/id/<feature>.ts`
2. **Type Safety via `TranslationSchema`**:
   The Indonesian locale enforces `TranslationSchema = typeof en`. Any missing key or type mismatch in `id` will fail `npm run build` and TypeScript check.
3. **Register in Index**:
   Export the dictionary in `frontend/src/i18n/locales/en/index.ts` and `frontend/src/i18n/locales/id/index.ts`.

---

## 4. Error Handling & API Error Localization

When displaying API errors in catch blocks:
- Use localized fallback messages rather than hardcoded error fallbacks:
  ```tsx
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : t('settings.oauth.unbindError'))
  }
  ```
- Backend errors with custom codes map to `t('errors.<CODE>')`. Ensure corresponding error codes are defined in both `en/errors.ts` and `id/errors.ts`.

---

## 5. Scope of Localized Content

The following elements **MUST** always be localized:
- Page titles, headings, and descriptions
- Button, link, and action labels
- Form labels, placeholder texts, and helper texts
- Dialog / modal headers, body copy, and button labels (`confirmDialog`, custom modals)
- Toast notifications (`toast.success`, `toast.error`, `toast.info`, `toast.warn`)
- Empty states, loading labels, and badge descriptions
- Aria attributes and tooltip text
