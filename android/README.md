# 📱 Task Management - Android Native App

Aplikasi Android murni (*Pure Native*) berbasis **Kotlin 2.x**, **Jetpack Compose**, dan arsitektur modular (*Feature-Driven*).

## 🚀 Fitur Utama
- **100% Native Jetpack Compose + Material 3 Design System** (warna & tema konsisten).
- **Full OAuth Sign-In**: Google 1-Tap (Credential Manager) & Sign in with Apple.
- **Tasks & Checklist**: Manajemen tugas, subtask checkbox, progress tracking, status & prioritas.
- **Proyek & Catatan**: Pengelompokan tugas per proyek dan catatan ide.
- **Real-Time Server-Sent Events (SSE)**: Streaming event live via `okhttp3-sse` dari `/api/v1/realtime/sse/events`.
- **Push Notification Ready**: Terhubung ke endpoint FCM backend (`POST /api/v1/reminders/device-token`).

## 🚀 Identitas & Konfigurasi
- **Application ID / Package:** `cloud.eka_dev.task_hub` *(Catatan: Android OS mewajibkan underscore `_` karena tanda strip `-` adalah operator minus di JVM).*
- **Backend Production/Cloud:** `https://be-task-hub.eka-dev.cloud/api/v1/`
- **Real-Time SSE:** `https://be-task-hub.eka-dev.cloud/api/v1/realtime/sse/events`

## 📂 Struktur Modular
- `core/model/`: Domain data classes & DTOs serializable.
- `core/network/`: Retrofit, OkHttp, AuthInterceptor (auto-inject Bearer token), SSEClient.
- `core/datastore/`: Jetpack DataStore (TokenManager & sesi user).
- `core/designsystem/`: Color tokens, Theme, Reusable Components (AppButton, TaskItemCard, Badges).
- `feature/auth/`: Login Screen & ViewModel.
- `feature/task/`: Task List Screen, Search, Filters, Add Task Dialog.
- `feature/project/`: Proyek & Modul Screen.
- `feature/note/`: Catatan & Ide Screen.
- `feature/reminder/`: Live SSE Alert Banner & Pengingat Screen.
- `feature/profile/`: Profil & Pengaturan Akun.

## 🛠️ Cara Menjalankan
1. Buka folder `android/` di **Android Studio (Ladybug / Meerkat / terbaru)**.
2. Pastikan Go backend berjalan di port `8080`.
   - Android Emulator otomatis memetakan `localhost` komputer ke IP: `http://10.0.2.2:8080/api/v1/`.
3. Klik tombol **Run** ▶️ di Android Studio.
