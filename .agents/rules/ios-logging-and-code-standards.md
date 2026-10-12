---
description: Mandatory AppLogger usage, English-only diagnostic logging, and modular architecture standards for the iOS Swift application
globs: ["ios/TaskManagement/**/*.swift"]
---

# iOS Logging & Code Architecture Standards

All Swift code in the iOS application must strictly follow modular architecture, modern Swift concurrency, clean separation of concerns, and unified diagnostic logging via `AppLogger`.

---

## 1. Absolute Prohibition of Raw `print(...)`

- ❌ **Forbidden**:
  ```swift
  // Bad: unstructured raw prints, ad-hoc emoji prefixes
  print("⚡ [SSE CONNECTED] Stream established")
  print("⚠️ Failed to register token: \(error)")
  print("User logged in")
  ```

- ✅ **Required**:
  ```swift
  // Good: structured AppLogger with log level and subsystem category
  AppLogger.success("SSE stream established successfully", category: .sse)
  AppLogger.error("Failed to register device token", error: error, category: .push)
  AppLogger.info("User authenticated successfully", category: .auth, metadata: ["provider": "google"])
  ```

---

## 2. Mandatory English-Only Logging Policy

To maintain professional, uniform diagnostic logs that align with backend services, cloud monitors, and Apple Console logs:

- **All log messages, error descriptions, and debug traces in `AppLogger` MUST be written in English.**
- ❌ **Forbidden**:
  ```swift
  AppLogger.info("Sesi pengguna telah terverifikasi", category: .auth)
  AppLogger.error("Gagal memuat daftar pengingat", category: .reminder)
  ```
- ✅ **Required**:
  ```swift
  AppLogger.info("User session verified", category: .auth)
  AppLogger.error("Failed to load reminders", error: error, category: .reminder)
  ```
- **Golden Rule**: **User-facing UI text is localized** (via `L10n.*` in English and Indonesian), while **internal developer logs are ALWAYS in English** (via `AppLogger`).

---

## 3. Log Levels & Subsystem Categories

Use the appropriate severity level:
- `AppLogger.debug(...)` : High-frequency debug events, parsed payload previews, SSE event data.
- `AppLogger.info(...)`  : Normal application lifecycle events (app launch, connecting to stream, cached token loaded).
- `AppLogger.success(...)`: Successful completion of network operations or state mutations (HTTP 200, OAuth linked, token registered).
- `AppLogger.warn(...)`   : Non-critical anomalies or recoverables (401 token refresh attempt, simulator APNs unavailability).
- `AppLogger.error(...)`  : Failures that prevented an operation (network loss, decoding errors, API failures). Always pass `error: error`.

Available Categories (`LogCategory`):
- `.app`      : General application lifecycle and configuration
- `.auth`     : Authentication, OAuth flows, JWT management, Keychain
- `.network`  : HTTP REST requests, responses, status codes, latency
- `.push`     : Apple Push Notifications, APNs device tokens, backend token registration
- `.sse`      : Realtime Server-Sent Events streaming
- `.task`     : Task CRUD, checklist mutations, focus state
- `.reminder` : Reminder alarms, snoozing, scheduling
- `.storage`  : Local caching and Keychain operations

---

## 4. Modular Feature-First Architecture

Code must be strictly partitioned into layers:

```text
ios/TaskManagement/
├── Core/
│   ├── Config/          # AppConfig, environment endpoints
│   ├── DesignSystem/    # Reusable UI components, AppColors, themes
│   ├── Localization/    # I18n engine, Translations dictionaries, L10n keys
│   ├── Logger/          # AppLogger diagnostic system
│   ├── Models/          # Codable DTOs and domain entity models
│   ├── Network/         # APIClient, SSEClient, NetworkLogger
│   ├── Notification/    # PushNotificationManager, UNUserNotificationCenter
│   └── Security/        # KeychainManager, token storage
└── Features/
    ├── Auth/            # AuthView, AuthViewModel
    ├── Note/            # NoteListView, NoteViewModel
    ├── Profile/         # ProfileView, ProfileViewModel
    ├── Project/         # ProjectListView, ProjectViewModel
    ├── Reminder/        # ReminderListView, ReminderViewModel
    └── Task/            # TaskListView, TaskDetailSheet, TaskViewModel
```

---

## 5. State Management & Concurrency (iOS 17+)

1. **Observation Framework (`@Observable`)**:
   - Always use `@Observable` for ViewModels.
   - Do NOT use legacy `@StateObject` / `ObservableObject` / `@Published` unless required for backward compatibility.
2. **Main Thread Safety (`@MainActor`)**:
   - Annotate all ViewModels with `@MainActor public final class <Feature>ViewModel`.
   - Ensures all UI mutations happen safely on the main thread without data races.
3. **Structured Concurrency (`async/await`)**:
   - Use `async/await` for asynchronous operations.
   - Run asynchronous side effects inside `Task { [weak self] in ... }` with proper error handling.

---

## 6. Error Handling Pattern

Every network or async call in a ViewModel must follow this structure:

```swift
@MainActor
public func deleteTask(id: Int64) {
    isLoading = true
    errorMessage = nil

    Task {
        do {
            try await apiClient.deleteTask(id: id)
            tasks.removeAll { $0.id == id }
            AppLogger.success("Task deleted successfully", category: .task, metadata: ["id": id])
        } catch {
            AppLogger.error("Failed to delete task", error: error, category: .task, metadata: ["id": id])
            self.errorMessage = error.localizedDescription
        }
        self.isLoading = false
    }
}
```
