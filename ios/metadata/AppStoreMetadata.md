# App Store Connect Metadata & Submission Guide

Dokumentasi lengkap metadata untuk registrasi dan rilis aplikasi **Task Hub** di **App Store Connect**.

---

## 1. General Information (Informasi Umum)

| Field | Value | Catatan |
|---|---|---|
| **App Name** | `Task Hub - Tasks & AI MCP` | Maksimal 30 karakter |
| **Subtitle (EN)** | `Smart Productivity & Sync` | Maksimal 30 karakter |
| **Subtitle (ID)** | `Produktivitas Pintar & Sync` | Maksimal 30 karakter |
| **Primary Category** | `Productivity` (Produktivitas) | |
| **Secondary Category** | `Business` (Bisnis) | |
| **Bundle ID** | `cloud.eka-dev.task-hub` | Sesuai `project.pbxproj` & provisioning profile |
| **SKU** | `TASKHUB_IOS_001` | Identifier unik internal di App Store Connect |
| **Default Language** | `English (U.S.)` | Tersedia lokalisasi Bahasa Indonesia (`id-ID`) |
| **Copyright** | `© 2026 Eka Dev Cloud. All rights reserved.` | |

---

## 2. URLs & Links (Tautan Regulasi & Bantuan)

| Field | URL |
|---|---|
| **Privacy Policy URL** | `https://be-task-hub.eka-dev.cloud/privacy` |
| **Support URL** | `https://be-task-hub.eka-dev.cloud/support` |
| **Marketing URL** | `https://be-task-hub.eka-dev.cloud` |

---

## 3. Keywords (Kata Kunci Pencarian)

- **English (U.S.)**:
  ```text
  tasks,productivity,notes,reminders,task management,projects,ai,mcp,task hub,model context protocol,kanban
  ```
- **Bahasa Indonesia**:
  ```text
  tugas,produktivitas,catatan,pengingat,manajemen tugas,proyek,ai,mcp,task hub,agenda harian
  ```

---

## 4. Promotional Text & Release Notes (Versi 1.0.0)

### Promotional Text
> **EN**: Supercharge your daily workflow with real-time synchronized task management and AI Model Context Protocol integration.  
> **ID**: Tingkatkan produktivitas harian Anda dengan manajemen tugas real-time dan integrasi AI Model Context Protocol.

### What's New in This Version (Release Notes)
> **EN**:
> Welcome to the initial release of Task Hub!
> • Real-time task, subtask, and project tracking
> • Markdown notes and smart reminder alerts
> • Model Context Protocol (MCP) Personal Access Token management for AI assistants (Cursor, Claude, Antigravity IDE)
> • Sign in with Apple and Google authentication
> • Native iOS 17+ experience with full Dark Mode support
>
> **ID**:
> Selamat datang di rilis perdana Task Hub!
> • Pengelolaan tugas, subtugas, dan proyek tersinkronisasi real-time
> • Catatan produktif dan pengingat pintar
> • Integrasi Model Context Protocol (MCP) untuk AI coding agent (Cursor, Claude, Antigravity IDE)
> • Masuk aman dengan Apple ID dan Google
> • Tampilan native iOS 17+ dengan dukungan tema Gelap & Terang

---

## 5. Age Rating & Content Declaration

- **Age Rating**: 4+ (Tidak ada konten dewasa, kekerasan, atau perjudian)
- **Data Collection (App Privacy Details)**:
  - **User Content**: Tasks, Notes, Projects (Linked to User, for App Functionality)
  - **Contact Info**: Name, Email address (Linked to User, for Account Authentication)
  - **Identifiers**: User ID (Linked to User, for App Functionality)
  - **Data Used to Track You**: None (Tidak ada data tracking pihak ketiga / iklan)

---

## 6. App Review Information (Pedoman Review Apple)

Sediakan akun pengujian untuk tim App Reviewer Apple:

- **Sign-in Required**: Yes
- **Sign-in Method**: Sign in with Apple / Test Account
- **Notes for Reviewer**:
  ```text
  Task Hub is a task and workspace productivity application with optional Model Context Protocol (MCP) integrations for developer assistants.
  Testers can log in easily with their Apple ID (Sandbox / Sandbox tester) or use standard Google Sign-In.
  ```

---

## 7. Export Compliance (Enkripsi)

Aplikasi menggunakan HTTPS / TLS standar untuk komunikasi jaringan.
- Di `ios/TaskManagement/Info.plist`:
  ```xml
  <key>ITSAppUsesNonExemptEncryption</key>
  <false/>
  ```
  *(Dengan konfigurasi ini, TestFlight dan App Store tidak akan memunculkan dialog "Missing Compliance" saat build diunggah).*
