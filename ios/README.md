# 🍏 Task Management - iOS Native App

Aplikasi iOS murni (*Pure Native*) berbasis **Swift 6 / 5.10**, **SwiftUI (iOS 17+)**, dan arsitektur modular modern.

## 🚀 Fitur Utama
- **Modern SwiftUI 17+**: Menggunakan `@Observable` macro dan Swift Concurrency (`async/await`, `AsyncSequence`).
- **100% Native Design System**: Warna, font, card modifiers, dan status badges.
- **Full OAuth Sign-In**: Native `SignInWithAppleButton` (`AuthenticationServices`) + Google Sign-In.
- **Tasks & Checklist**: Manajemen tugas, subtask checkbox, progress tracking, status & prioritas.
- **Proyek & Catatan**: Pengelompokan tugas per proyek dan catatan ide.
- **Native SSE (Server-Sent Events)**: Streaming event live via `URLSession.shared.bytes(for:)` tanpa library pihak ketiga.
- **Push Notification Ready**: Terhubung ke endpoint FCM backend (`POST /api/v1/reminders/device-token`).

## 🚀 Identitas & Konfigurasi
- **Bundle Identifier:** `cloud.eka-dev.task-hub`
- **Backend Production/Cloud:** `https://be-task-hub.eka-dev.cloud/api/v1`
- **Real-Time SSE:** `https://be-task-hub.eka-dev.cloud/api/v1/realtime/sse/events`

## 📂 Struktur Modular
- `Core/Models/`: Model data Codable terstruktur.
- `Core/Network/`: Actor-based `APIClient` (URLSession + Auto-refresh token on 401) & `SSEClient`.
- `Core/Security/`: `KeychainManager` (penyimpanan aman token & sesi).
- `Core/DesignSystem/`: Color tokens, Theme, Reusable Components (AppButton, StatusBadge, PriorityBadge).
- `Features/Auth/`: AuthView & AuthViewModel.
- `Features/Task/`: TaskListView, Searchable, Filter Capsules, Create Task Sheet.
- `Features/Project/`: ProjectListView.
- `Features/Note/`: NoteListView.
- `Features/Reminder/`: Live SSE Alert Banner & ReminderListView.
- `Features/Profile/`: ProfileView & Sign Out.

## 🛠️ Cara Menjalankan
1. Buka folder `ios/` di **Xcode 15/16/27**.
   - Bisa dibuka langsung via Swift Package (`Package.swift`) atau dibuatkan New Xcode App Project targeting iOS 17+.
2. Pastikan Go backend berjalan di port `8080`.
   - iOS Simulator otomatis menggunakan: `http://localhost:8080/api/v1`.
3. Pilih target simulator iPhone 15/16 dan klik **Run** (Cmd + R) ▶️.
