---
name: ios-add-feature
description: >-
  Step-by-step procedure to scaffold and implement a new domain feature in the iOS Swift app using Feature-driven architecture, @Observable ViewModels, AppLogger, and L10n localization.
---

# Scaffolding a New iOS Feature Module

This skill guides you through implementing a new domain feature in the **iOS Swift** application adhering to Clean Architecture and project standards.

---

## Step 1: Define Models & DTOs

Add the entity models and request payloads to `ios/TaskManagement/Core/Models/Models.swift`:

```swift
public struct MyItem: Identifiable, Codable, Sendable {
    public let id: Int64
    public var title: String
    public var description: String?
    public var createdAt: String?

    enum CodingKeys: String, CodingKey {
        case id, title, description
        case createdAt = "created_at"
    }
}

public struct CreateItemRequest: Codable {
    public let title: String
    public let description: String?
}
```

---

## Step 2: Implement API Client Endpoints

Add async network methods to `ios/TaskManagement/Core/Network/APIClient.swift`:

```swift
public func getItems() async throws -> [MyItem] {
    return try await request(endpoint: "items")
}

public func createItem(req: CreateItemRequest) async throws -> MyItem {
    let body = try encoder.encode(req)
    return try await request(endpoint: "items", method: "POST", body: body)
}
```

---

## Step 3: Add Localization Keys (English & Indonesian)

1. In `ios/TaskManagement/Core/Localization/I18n.swift`:
   Add keys under `Translations.data[.en]` and `Translations.data[.id]`:
   ```swift
   // .en
   "myfeature.title": "My Feature",
   "myfeature.empty": "No items found",
   "myfeature.create": "Create Item",

   // .id
   "myfeature.title": "Fitur Saya",
   "myfeature.empty": "Tidak ada item ditemukan",
   "myfeature.create": "Tambah Item",
   ```

2. In `ios/TaskManagement/Core/Localization/L10n.swift`:
   Expose the type-safe static accessors:
   ```swift
   public enum MyFeature {
       public static var title: String { "myfeature.title".localized }
       public static var empty: String { "myfeature.empty".localized }
       public static var create: String { "myfeature.create".localized }
   }
   ```

---

## Step 4: Implement Feature ViewModel

Create `ios/TaskManagement/Features/<Feature>/<Feature>ViewModel.swift`:

```swift
import SwiftUI

@MainActor
@Observable
public final class MyFeatureViewModel {
    public var items: [MyItem] = []
    public var isLoading = false
    public var errorMessage: String?

    private let apiClient = APIClient.shared

    public init() {
        Task {
            await loadItems()
        }
    }

    public func loadItems() async {
        isLoading = true
        errorMessage = nil
        do {
            items = try await apiClient.getItems()
            AppLogger.info("Loaded \(items.count) items", category: .app)
        } catch {
            AppLogger.error("Failed to load items", error: error, category: .app)
            errorMessage = error.localizedDescription
        }
        isLoading = false
    }
}
```

---

## Step 5: Implement SwiftUI View

Create `ios/TaskManagement/Features/<Feature>/<Feature>View.swift`:

```swift
import SwiftUI

public struct MyFeatureView: View {
    @Bindable var viewModel: MyFeatureViewModel

    public init(viewModel: MyFeatureViewModel) {
        self.viewModel = viewModel
    }

    public var body: some View {
        NavigationStack {
            List(viewModel.items) { item in
                Text(item.title)
            }
            .navigationTitle(L10n.MyFeature.title)
        }
    }
}
```

---

## Step 6: Link to Xcode Project (`project.pbxproj`)

Add the newly created files to `ios/TaskManagement.xcodeproj/project.pbxproj`:
1. `PBXBuildFile`: Add file entry in Sources.
2. `PBXFileReference`: Add file path reference.
3. `PBXGroup`: Add file to the target's group hierarchy.
4. `PBXSourcesBuildPhase`: Add build file to compilation list.

---

## Step 7: Validate & Test

Run the build checks:
```bash
swift build --package-path ios --triple arm64-apple-ios17.0-simulator
xcodebuild -project ios/TaskManagement.xcodeproj -scheme TaskManagement -destination 'generic/platform=iOS Simulator' CODE_SIGNING_ALLOWED=NO build
```
Verify zero compiler warnings or errors.
