---
description: Mandatory iOS internationalization (L10n / I18n) and locale usage for all SwiftUI text, alerts, buttons, dialogs, and error messages
globs: ["ios/TaskManagement/**/*.swift"]
---

# iOS Internationalization (i18n) & Localized Message Rules

All user-facing text, button labels, placeholders, dialogs, alert sheets, and error messages in the iOS native application must strictly use the type-safe localization system (`L10n.*` / `I18n.shared`). Hardcoding raw strings in SwiftUI Views, ViewModels, or alerts is strictly prohibited.

---

## 1. Absolute Prohibition of Hardcoded Display Strings

- ❌ **Forbidden**:
  ```swift
  // Bad: hardcoded Indonesian or English text in SwiftUI views or ViewModels
  Text("Cari tugas, tag, atau deskripsi…")
  Button("Semua Proyek") { ... }
  Button("Lanjutkan dengan Google") { ... }
  self.errorMessage = "Sesi telah berakhir atau tidak valid. Silakan login kembali."
  self.errorMessage = "Tidak dapat menemukan jendela aplikasi untuk Google Sign-In"
  ```

- ✅ **Required**:
  ```swift
  // Good: localized through type-safe L10n keys
  TextField(L10n.Tasks.searchPlaceholder, text: $viewModel.searchQuery)
  Button(L10n.Tasks.allProjects) { ... }
  Text(L10n.Auth.signInGoogle)
  self.errorMessage = L10n.Auth.sessionExpired
  self.errorMessage = L10n.Auth.windowNotFound
  ```

---

## 2. Type-Safe Localization Architecture (`L10n` & `I18n`)

The iOS app uses an Observable reactive internationalization engine:
- **`I18n.shared`** (`Core/Localization/I18n.swift`): Central observable language manager supporting English (`.en`) and Indonesian (`.id`). Automatically updates the entire SwiftUI hierarchy when the language is toggled.
- **`L10n`** (`Core/Localization/L10n.swift`): Type-safe static enum providing compile-time verified access to translation keys (e.g. `L10n.Nav.tasks`, `L10n.Auth.title`, `L10n.Reminders.emptyTitle`).
- **`.localized` String Extension**: Shorthand syntax on key strings (e.g. `"tasks.title".localized` or `"reminders.snoozedCount".localized(with: count)`).

---

## 3. Strict Synchronization of English (`en`) and Indonesian (`id`)

1. **Both Locales Must Be Updated Together**:
   Whenever adding a new user-facing key:
   - Add the key to `Translations.data[.en]` in `Core/Localization/I18n.swift`
   - Add the key to `Translations.data[.id]` in `Core/Localization/I18n.swift`
   - Expose the type-safe static property in `Core/Localization/L10n.swift` under the appropriate domain enum (`Nav`, `Auth`, `Tasks`, `Projects`, `Notes`, `Reminders`, `Settings`, `General`).

2. **Parameter Interpolation**:
   When injecting numbers or dynamic variables, use format specifiers like `%lld` or `%@`:
   - Dictionary:
     ```swift
     // .en
     "reminders.snoozedCount": "Snoozed %lldx"
     // .id
     "reminders.snoozedCount": "Ditunda %lldx"
     ```
   - Usage in `L10n`:
     ```swift
     public static func snoozed(count: Int) -> String {
         "reminders.snoozedCount".localized(with: count)
     }
     ```

---

## 4. Error Handling & User-Facing Error Messages

- **Separation of Diagnostic Logs vs User Messages**:
  - Internal technical details are logged to **`AppLogger.error`** in **English**.
  - User-facing `errorMessage` displayed in banners or alerts must ALWAYS be a localized string from `L10n.*`.
  ```swift
  do {
      try await apiClient.deleteTask(id: id)
  } catch {
      AppLogger.error("Failed to delete task", error: error, category: .task, metadata: ["task_id": id])
      self.errorMessage = L10n.Tasks.deleteFailed
  }
  ```

---

## 5. Scope of Localized Content

The following elements **MUST** always be localized:
- Screen titles, navigation bars, and headers
- Button labels, menu item titles, and action sheets
- Form labels, textfield placeholders, and prompt hints
- Alert titles, confirmation dialog messages, and banner warnings
- Empty state titles and descriptive guidance text
- Segmented control and filter pill labels
- Status tags and badge text
