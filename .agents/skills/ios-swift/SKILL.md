---
name: ios-swift
description: >-
  Architectural guidelines, modern SwiftUI (iOS 17+), Observation framework, AppLogger, I18n/L10n localization, APIClient network patterns, and quality gates for the iOS native app. Use when building or modifying iOS components, views, view models, or services.
---

# iOS Swift Application Skill Guide

This skill provides architectural guidelines, best practices, and standard workflows for maintaining and developing the **Task Hub iOS Application** (`ios/TaskManagement`).

---

## 1. Stack Overview

- **Platform**: iOS 17.0+ (deployment target)
- **Language**: Swift 5.10 / Swift 6 (Strict Concurrency aware)
- **UI Framework**: SwiftUI (Modern declarative hierarchy)
- **State Management**: Observation framework (`@Observable`, `@Bindable`, `@State`)
- **Main Thread Safety**: `@MainActor` annotated ViewModels
- **Networking**: `URLSession` async/await with `APIClient` singleton & token auto-refresh
- **Real-Time Streaming**: Server-Sent Events via `SSEClient` (`AsyncThrowingStream`)
- **Logging**: Unified `AppLogger` (`os.Logger` + styled console)
- **Localization**: Reactive `I18n` engine with type-safe `L10n` accessors (English & Indonesian)
- **Project Structure**: Xcode Project (`TaskManagement.xcodeproj`) + Swift Package Manager (`Package.swift`)

---

## 2. Directory Architecture

```text
ios/TaskManagement/
├── Core/
│   ├── Config/          # AppConfig & environment endpoints
│   ├── DesignSystem/    # AppColors, glassCardStyle, BrandLogoView, Components
│   ├── Localization/    # I18n.swift, L10n.swift (Translations dictionary)
│   ├── Logger/          # AppLogger.swift (LogLevel, LogCategory)
│   ├── Models/          # Codable models & API request/response payloads
│   ├── Network/         # APIClient.swift, SSEClient.swift, NetworkLogger.swift
│   ├── Notification/    # PushNotificationManager.swift
│   └── Security/        # KeychainManager.swift
├── Features/
│   ├── Auth/            # AuthView, AuthViewModel
│   ├── Note/            # NoteListView, NoteViewModel
│   ├── Profile/         # ProfileView, ProfileViewModel
│   ├── Project/         # ProjectListView, ProjectViewModel
│   ├── Reminder/        # ReminderListView, ReminderViewModel
│   └── Task/            # TaskListView, TaskDetailSheet, TaskViewModel
├── Resources/           # Assets.xcassets, Localizable.xcstrings
├── Info.plist           # Bundle configurations, OAuth schemes, background modes
└── TaskManagementApp.swift # Application entry point & AppDelegate
```

---

## 3. Core Architectural Rules

### Rule 1: Always Use `AppLogger` with English Messages
Never use `print(...)`. All diagnostic log strings must be written in English:
```swift
AppLogger.debug("Fetching page \(page)", category: .task)
AppLogger.info("User session verified", category: .auth)
AppLogger.success("Task updated", category: .task, metadata: ["id": id])
AppLogger.warn("Token expired, attempting refresh", category: .auth)
AppLogger.error("Failed to load tasks", error: error, category: .task)
```

### Rule 2: Mandatory Localization via `L10n`
Never hardcode user-facing strings:
```swift
// ❌ Bad:
Text("Tugas Baru")
// ✅ Good:
Text(L10n.Tasks.createTask)
```
Always define new keys in both `.en` and `.id` in `Core/Localization/I18n.swift` and expose them through `Core/Localization/L10n.swift`.

### Rule 3: Main Thread ViewModels (`@MainActor` & `@Observable`)
All ViewModels must be declared as:
```swift
@MainActor
@Observable
public final class MyFeatureViewModel {
    public var items: [MyItem] = []
    public var isLoading = false
    public var errorMessage: String?
    
    private let apiClient = APIClient.shared
    // ...
}
```

### Rule 4: Design System & Styling
- Use curated semantic colors from `AppColors` (`AppColors.primary`, `AppColors.indigo`, `AppColors.violet`, `AppColors.priorityHigh`, etc.).
- Use modern iOS styling modifiers like `.glassCardStyle()` for cards and sheet containers.
- Avoid raw generic colors (`Color.blue`, `Color.red`).

---

## 4. Quality Gates & Validation

Before committing any iOS changes, verify with both compiler tools:

```bash
# 1. Swift Package Manager build
swift build --package-path ios --triple arm64-apple-ios17.0-simulator

# 2. Full Xcode Simulator Build
xcodebuild -project ios/TaskManagement.xcodeproj -scheme TaskManagement -destination 'generic/platform=iOS Simulator' CODE_SIGNING_ALLOWED=NO build
```
Both commands must exit with code 0 and 0 warnings.
