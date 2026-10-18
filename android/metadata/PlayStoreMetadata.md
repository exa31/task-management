# Google Play Console Metadata & Store Listing Guide

Dokumentasi lengkap metadata untuk registrasi, listing toko, dan rilis aplikasi **Task Hub** di **Google Play Console**.

---

## 1. App Details & Identifiers (Detail Aplikasi)

| Field | Nilai / Konfigurasi | Catatan |
|---|---|---|
| **App Name (Title)** | `Task Hub - Tasks & AI MCP` | Maksimal 30 karakter (Google Play Store limit) |
| **Package Name (Application ID)** | `cloud.eka_dev.task_hub` | Sesuai `android/app/build.gradle.kts` |
| **Default Language** | `English (United States) – en-US` | Bahasa utama toko |
| **Additional Languages** | `Indonesian – id` | Lokalisasi toko bahasa Indonesia |
| **App Category** | `Productivity` (Produktivitas) | Kategori utama di Google Play |
| **Store Tags** | `Productivity`, `Task management`, `Notes`, `Calendar`, `Tools` | Maksimal 5 tag |
| **App Type** | `App` (bukan Game) | |
| **Pricing** | `Free` (Gratis tanpa iklan) | |

---

## 2. Store Listing (Deskripsi Toko)

### A. English (United States) – `en-US`

#### Short Description (Maksimal 80 Karakter)
```text
Smart task management, project tracker, notes, and AI MCP integration.
```
*(Karakter: 70 / 80)*

#### Full Description (Maksimal 4.000 Karakter)
```text
Task Hub is an all-in-one productivity workspace designed for developers, professionals, and students. Organize tasks, manage projects, write notes, schedule reminders, and connect AI coding assistants seamlessly using Model Context Protocol (MCP).

Key Features:

1. INTELLIGENT TASK MANAGEMENT
• Organize your tasks with intuitive statuses: Backlog, To Do, In Progress, Review, and Done.
• Pin important tasks to Today's Focus for maximum daily productivity.
• Breakdown complex goals into actionable subtasks with real-time completion tracking.
• Categorize tasks by priority: Urgent, High, Medium, and Low.
• Search and filter tasks instantly with fast cursor-based pagination.

2. PROJECT TRACKING & METRICS
• Group tasks into color-coded projects.
• Visual progress bars and completion percentages for every project.
• Keep track of project milestones and deliverables in one unified dashboard.

3. CALENDAR & SCHEDULE VIEW
• Monthly calendar view showing scheduled tasks and deadlines.
• Filter tasks by selected dates with interactive day chips.
• Stay ahead of upcoming due dates and meetings effortlessly.

4. PRODUCTIVE NOTES & INSTANT CONVERSION
• Capture quick thoughts, meeting notes, and brainstorming ideas.
• Seamlessly convert any note into a full actionable task with a single tap.
• Pin essential notes for immediate access at the top of your list.

5. SMART REAL-TIME REMINDERS
• Set scheduled reminders with customizable alarms.
• Quick actions: Snooze (+15 minutes) or dismiss directly from the reminder card.
• Real-time synchronization powered by Server-Sent Events (SSE).

6. AI MODEL CONTEXT PROTOCOL (MCP) INTEGRATION
• Built specifically for modern developer workflows.
• Generate and manage personal access tokens for AI tools (Cursor, Claude Desktop, Antigravity IDE).
• Empower your AI agents to read, create, and update tasks autonomously through MCP servers.

7. NATIVE MATERIAL 3 EXPERIENCE
• Built with modern Jetpack Compose and Material Design 3.
• Full Dark Mode and Light Mode support with smooth theme transitions.
• Fast, responsive, and privacy-focused.

8. SECURE CLOUD SYNC & AUTHENTICATION
• Cloud backup and real-time cross-platform synchronization (iOS, Android, Web).
• Safe and fast sign-in with Google Credential Manager.

Take control of your workflow today with Task Hub!
```

---

### B. Indonesian – `id`

#### Short Description (Maksimal 80 Karakter)
```text
Manajemen tugas pintar, pelacak proyek, catatan, dan integrasi AI MCP.
```
*(Karakter: 70 / 80)*

#### Full Description (Maksimal 4.000 Karakter)
```text
Task Hub adalah ruang kerja produktivitas lengkap yang dirancang untuk pengembang, profesional, dan pelajar. Kelola tugas, pantau proyek, catat ide, pasang pengingat, dan hubungkan asisten AI secara langsung menggunakan Model Context Protocol (MCP).

Fitur Unggulan:

1. MANAJEMEN TUGAS PINTAR
• Atur alur kerja dengan status fleksibel: Backlog, Todo, Sedang Dikerjakan, Review, dan Selesai.
• Sematkan tugas utama ke Fokus Hari Ini agar produktivitas harian tetap terarah.
• Uraikan pekerjaan besar menjadi subtugas terstruktur dengan pelacak progres instan.
• Atur prioritas tugas: Urgent, Tinggi, Sedang, dan Rendah.
• Pencarian dan filter cepat didukung cursor pagination hemat memori.

2. PELACAK PROYEK & METRIK
• Kelompokkan tugas ke dalam proyek dengan label warna yang terorganisir.
• Pantau indikator progres visual dan persentase penyelesaian setiap proyek.
• Monitor target dan pencapaian tim dalam satu dasbor terpadu.

3. TAMPILAN KALENDER & AGENDA
• Tampilan kalender bulanan yang menampilkan jadwal tugas dan tenggat waktu.
• Filter tugas berdasarkan tanggal yang dipilih dengan navigasi hari interaktif.
• Jangan lewatkan deadline atau jadwal penting harian.

4. CATATAN PRODUKTIF & KONVERSI INSTAN
• Tulis ide spontan, ringkasan rapat, atau rencana kerja.
• Ubah catatan menjadi tugas aksi nyata hanya dengan satu ketukan tombol.
• Sematkan catatan penting agar selalu tampil di posisi teratas.

5. PENGINGAT REAL-TIME CERDAS
• Pasang pengingat terjadwal dengan alarm notifikasi tepat waktu.
• Aksi cepat: Tunda 15 menit (*snooze*) atau tandai selesai langsung dari kartu pengingat.
• Sinkronisasi data real-time berbasis Server-Sent Events (SSE).

6. INTEGRASI AI MODEL CONTEXT PROTOCOL (MCP)
• Dirancang khusus untuk workflow developer masa kini.
• Buat dan kelola personal access token untuk AI agent (Cursor, Claude Desktop, Antigravity IDE).
• Beri kemampuan bagi AI agent Anda untuk membaca, membuat, dan memperbarui tugas otomatis melalui MCP.

7. PENGALAMAN NATIVE MATERIAL 3
• Dibangun dengan Jetpack Compose modern dan standar Material Design 3.
• Dukungan penuh Mode Gelap (Dark Mode) dan Mode Terang dengan transisi mulus.
• Cepat, ringan, responsif, dan mengutamakan privasi.

8. SINKRONISASI CLOUD AMAN & MULTIPLATFORM
• Backup cloud dan sinkronisasi real-time antar perangkat (Android, iOS, Web).
• Masuk praktis dan aman dengan Google Credential Manager.

Tingkatkan produktivitas kerja harian Anda bersama Task Hub!
```

---

## 3. URLs & Developer Contact Information

| Field | Nilai |
|---|---|
| **Support Email** | `support@eka-dev.cloud` |
| **Support Website** | `https://be-task-hub.eka-dev.cloud/support` |
| **Privacy Policy URL** | `https://be-task-hub.eka-dev.cloud/privacy` |
| **Marketing / Homepage URL** | `https://be-task-hub.eka-dev.cloud` |
| **Developer Name** | `Eka Dev Cloud` |

---

## 4. Graphic Assets & Media Specifications (Spesifikasi Aset Grafis)

| Aset | Dimensi | Format | Persyaratan Google Play |
|---|---|---|---|
| **App Icon** | `512 x 512 px` | PNG 32-bit (dengan alpha) | Maksimal 1 MB. Format flat tanpa bayangan luar (store menambahkan mask squircles otomatis). |
| **Feature Graphic** | `1024 x 500 px` | JPEG atau PNG 24-bit (tanpa alpha) | Maksimal 15 MB. Visual banner utama di bagian atas listing Play Store. |
| **Phone Screenshots** | Min. `1080 x 1920 px` (9:16) atau `1920 x 1080 px` (16:9) | JPEG atau PNG 24-bit | Minimal 2 screenshot, maksimal 8 screenshot per bahasa. Rekomendasi 5 screenshot: Tasks, Projects, Calendar, Notes, MCP Integration. |
| **7-inch Tablet Screenshots** | Min. `1200 x 1920 px` | JPEG atau PNG 24-bit | Minimal 1 screenshot (rekomendasi 4-5) jika mendukung tablet. |
| **10-inch Tablet Screenshots** | Min. `1600 x 2560 px` | JPEG atau PNG 24-bit | Minimal 1 screenshot jika mendukung tablet besar. |

---

## 5. Google Play Data Safety Declaration (Deklarasi Keamanan Data)

Sesuai kuesioner **Google Play Data Safety Form**:

### A. Data Collection & Sharing Summary
- **Data shared with third parties?** `No` (Tidak ada data pengguna yang dibagikan atau dijual ke pihak ketiga).
- **Data encrypted in transit?** `Yes` (Semua transfer data menggunakan protokol HTTPS / TLS 1.3 terenkripsi).
- **Can users request data deletion?** `Yes` (Pengguna dapat mengajukan penghapusan akun dan data melalui menu profil atau email support).

### B. Data Types Collected

| Kategori Data | Tipe Data | Tujuan Pengumpulan | Dibagikan? | Wajib / Opsional |
|---|---|---|---|---|
| **Personal Info** | Name (Nama) | App functionality & Account management | Tidak | Wajib |
| **Personal Info** | Email address | Account authentication & notifications | Tidak | Wajib |
| **Personal Info** | User IDs | Account management | Tidak | Wajib |
| **User Content** | Tasks & Notes content | App functionality (Core feature) | Tidak | Wajib |
| **Diagnostics** | Crash logs & error diagnostics | App functionality & Performance monitoring | Tidak | Opsional |

---

## 6. Content Rating (IARC Questionnaire)

Isian kuesioner **International Age Rating Coalition (IARC)** di Google Play Console:

| Pertanyaan Kuesioner | Jawaban |
|---|---|
| **Category** | `Utility, Productivity, Communication, or Other` |
| **Violence** | `No` (Tidak ada adegan kekerasan) |
| **Sexuality / Nudity** | `No` (Tidak ada konten seksual) |
| **Language** | `No` (Tidak ada bahasa kasar / vulgar) |
| **Controlled Substances** | `No` (Tidak mempromosikan alkohol/rokok/narkoba) |
| **Promotion of Gambling** | `No` |
| **User Interaction** | `No` (Aplikasi produktivitas personal, tidak ada open chat publik) |
| **Physical Location Sharing** | `No` (Tidak membagikan lokasi GPS fisik) |
| **Digital Purchases** | `No` (Tidak ada in-app purchases saat ini) |

> **Hasil Prediksi Rating**:
> - **PEGI**: `PEGI 3`
> - **ESRB**: `Everyone (E)`
> - **Google Play Indonesia**: `3+`

---

## 7. Target Audience & Content (Audiens Target)

- **Target Age Groups**: `18 and older`, `13-17 years old`.
- **Children (Under 13)**: `No`.  
  *(Memilih 'No' untuk anak-anak di bawah 13 tahun mencegah kewajiban kepatuhan Google Play Designed for Families Policy yang kompleks).*
- **Appeal to Children**: `No`.

---

## 8. App Review Credentials & Test Account (Panduan Penguji Google)

Sediakan akun uji untuk tim Google Play App Review:

```text
Sign-in required: Yes
Testing Account Method: Google Sign-In or Email Credential
Notes for Reviewer:
Task Hub is a workspace productivity application featuring real-time task management, markdown notes, calendar scheduling, and Model Context Protocol (MCP) integrations for AI coding agents.
Reviewers can freely sign in using Google Sign-In or test accounts provided. All network communications are encrypted via HTTPS.
```

---

## 9. Release Notes (What's New in Version 1.0.0)

Gunakan catatan rilis ini saat membuat rilis di tab **Production / Closed Testing**:

### English (`en-US`):
```text
Welcome to the official launch of Task Hub on Android!
• Real-time task, subtask, and project tracking
• Monthly calendar schedule view with date filtering
• Markdown note taking with instant convert-to-task feature
• Smart reminder alerts with snooze & dismiss actions
• AI Model Context Protocol (MCP) Personal Access Token management
• Google Credential Manager integration for quick and secure sign-in
• Modern Material 3 UI with full Dark Mode and multi-language support (English & Indonesian)
```

### Indonesian (`id`):
```text
Selamat datang di rilis perdana Task Hub untuk Android!
• Pengelolaan tugas, subtugas, dan proyek real-time
• Tampilan kalender bulanan dengan filter tanggal interaktif
• Catatan ide dengan fitur konversi instan menjadi tugas
• Pengingat terjadwal dengan aksi tunda dan selesaikan
• Integrasi AI Model Context Protocol (MCP) Personal Access Token
• Masuk aman dan cepat dengan Google Credential Manager
• Tampilan modern Material 3 dengan Mode Gelap dan dukungan bahasa Indonesia & Inggris
```

---

## 10. Checklist Sebelum Rilis (Pre-Launch Checklist)

- [ ] **Application ID terverifikasi**: `cloud.eka_dev.task_hub`
- [ ] **Android App Bundle (AAB)** di-generate dengan build type `release`
- [ ] **Keystore signing** terkonfigurasi (Play Feature Delivery / Play App Signing aktif)
- [ ] **App Icon (512x512)** dan **Feature Graphic (1024x500)** terunggah di Main Store Listing
- [ ] **Minimal 2 Phone Screenshots** (rekomendasi 5 screenshot ukuran 1080x1920) terunggah
- [ ] **Privacy Policy URL** aktif dan dapat diakses publik (`https://be-task-hub.eka-dev.cloud/privacy`)
- [ ] **Data Safety Form** telah diisi sesuai panduan Section 5
- [ ] **Target SDK** memenuhi standar Google Play (Target SDK 35 untuk rilis 2026/2027)
