# Panduan Lengkap GraphQL untuk Pemula (Task Management System)

Selamat datang di GraphQL! Panduan ini dirancang khusus untuk Anda yang baru pertama kali menggunakan GraphQL di proyek Task Management ini.

---

## 1. Apa Itu GraphQL dan Mengapa Kita Menggunakannya?

### Perbedaan Mendasar: REST API vs GraphQL

| Fitur | REST API Tradisional | GraphQL |
| :--- | :--- | :--- |
| **Endpoint** | Banyak endpoint (`/api/v1/tasks`, `/api/v1/reminders`, `/api/v1/activities`) | Cukup **1 endpoint** tunggal (`/api/v1/graphql`) |
| **Jumlah Request** | Untuk kalender: butuh 3x HTTP request terpisah untuk tugas, pengingat, dan riwayat | **1x HTTP request** bisa mengambil semua data sekaligus |
| **Bentuk Data** | Server menentukan struktur JSON, seringkali ada data tak terpakai (*over-fetching*) | Klien meminta field yang diinginkan secara presisi (*exact fetching*) |
| **Uji Coba & Dokumentasi** | Membutuhkan Postman atau Swagger terpisah | Dilengkapi **Interactive Playground** langsung di browser |

### Contoh Nyata pada Fitur Kalender Kita:
Di halaman kalender, kita perlu menampilkan:
1. Daftar **Tugas** yang punya tenggat waktu (*due date*).
2. Daftar **Pengingat** (*reminders*) beserta jam alarmnya.
3. Daftar **Riwayat Aktivitas** (*what was done / work log*).

Dengan REST, aplikasi web/mobile harus menembak 3 endpoint berbeda secara berurutan. Dengan GraphQL, kita cukup mengirim satu query JSON ke `/api/v1/graphql` dan menerima ketiga data tersebut dalam 1 respon instan!

---

## 2. Cara Kerja GraphQL di Backend Kita (Go Fiber)

Endpoint GraphQL aplikasi ini terletak di:
```text
http://localhost:8080/api/v1/graphql
```

- **Jika dibuka lewat Browser (GET)**: Menampilkan **GraphQL Playground** (seperti Postman interaktif langsung di browser untuk mencoba query dan melihat dokumentasi schema secara visual).
- **Jika dipanggil oleh Aplikasi (POST)**: Menerima payload JSON berformat:
  ```json
  {
    "query": "query { me { name email } }",
    "variables": {}
  }
  ```

---

## 3. Struktur Query GraphQL

Format GraphQL sangat mirip dengan struktur objek JSON, tetapi hanya berisi nama properti (*keys*) tanpa nilai. Server akan mengembalikan JSON yang bentuknya persis seperti query Anda.

### A. Contoh Query Sederhana (Data Profil Diri)
```graphql
query GetMyProfile {
  me {
    id
    name
    email
  }
}
```
**Respon dari Server:**
```json
{
  "data": {
    "me": {
      "id": "048974f4-7e5b-4267-b44a-5a99c2c7937b",
      "name": "Eka",
      "email": "eka@example.com"
    }
  }
}
```

---

### B. Contoh Query Kalender Terpadu (Tugas + Pengingat + Riwayat)
Ini adalah query utama yang digunakan oleh halaman Kalender:

```graphql
query GetCalendarData($from: String, $to: String) {
  # 1. Ambil Tugas dalam rentang tanggal
  tasks(from: $from, to: $to) {
    id
    title
    due_date
    status
    priority
    project_name
    project_color
    subtasks_count
    subtasks_completed
  }

  # 2. Ambil Pengingat/Alarm
  reminders(from: $from, to: $to) {
    id
    title
    remind_at
    status
  }

  # 3. Ambil Log Aktivitas (Apa yang telah dikerjakan)
  activities(from: $from, to: $to) {
    id
    task_id
    task_title
    activity_type
    description
    created_at
  }
}
```

Dengan Variables (Parameter JSON):
```json
{
  "from": "2026-10-01T00:00:00Z",
  "to": "2026-10-31T23:59:59Z"
}
```

**Respon Tunggal:**
```json
{
  "data": {
    "tasks": [
      {
        "id": 10,
        "title": "Review PR iOS Localization",
        "due_date": "2026-10-15T18:00:00Z",
        "status": "todo",
        "priority": "high",
        "project_name": "Task-management",
        "project_color": "#3b82f6"
      }
    ],
    "reminders": [
      {
        "id": 5,
        "title": "Daily Standup Meeting",
        "remind_at": "2026-10-15T09:00:00Z",
        "status": "pending"
      }
    ],
    "activities": [
      {
        "id": 102,
        "task_id": 10,
        "task_title": "Review PR iOS Localization",
        "activity_type": "progress_logged",
        "description": "Menambahkan dokumentasi GraphQL",
        "created_at": "2026-10-13T13:25:00Z"
      }
    ]
  }
}
```

---

## 4. Cara Menggunakan GraphQL di Frontend (Web / React)

Di frontend kita tidak perlu menginstall library berat seperti Apollo Client. Kita cukup menggunakan helper ringan berbasis Axios yang sudah otomatis menyertakan token JWT:

### Implementasi Helper (`src/api/graphql.ts`):
```typescript
import { apiClient } from './client'

export const fetchGraphQL = async <T>(
  query: string,
  variables?: Record<string, unknown>
): Promise<T> => {
  const response = await apiClient.post('/graphql', {
    query,
    variables
  })

  // Tangani error jika ada
  if (response.data?.errors?.length) {
    throw new Error(response.data.errors[0].message)
  }

  return response.data?.data as T
}
```

### Pemakaian di Komponen React / Hook:
```typescript
import { useQuery } from '@tanstack/react-query'
import { fetchGraphQL } from '@/api/graphql'

const CALENDAR_QUERY = `
  query GetCalendar($from: String, $to: String) {
    tasks(from: $from, to: $to) {
      id
      title
      due_date
      status
      project_color
    }
    reminders(from: $from, to: $to) {
      id
      title
      remind_at
      status
    }
    activities(from: $from, to: $to) {
      id
      task_title
      description
      created_at
    }
  }
`

export function useCalendarData(from: string, to: string) {
  return useQuery({
    queryKey: ['calendar', from, to],
    queryFn: () => fetchGraphQL(CALENDAR_QUERY, { from, to })
  })
}
```

---

## 5. Cara Menggunakan GraphQL di Mobile (iOS / SwiftUI)

Di iOS, pemanggilan GraphQL sama mudahnya seperti pemanggilan REST API melalui method `POST` dengan body JSON berisi `query` dan `variables`:

```swift
// Struktur Request
struct GraphQLRequest: Codable {
    let query: String
    let variables: [String: String]?
}

// Struktur Respon
struct CalendarGraphQLData: Codable {
    let tasks: [TaskItem]
    let reminders: [Reminder]
    let activities: [TaskActivity]
}

struct GraphQLResponse<T: Codable>: Codable {
    let data: T?
    let errors: [GraphQLError]?
}

// Pemanggilan via APIClient
let query = """
query GetCalendar($from: String, $to: String) {
    tasks(from: $from, to: $to) { id title due_date status priority }
    reminders(from: $from, to: $to) { id title remind_at status }
    activities(from: $from, to: $to) { id task_title description created_at }
}
"""

let body = try JSONEncoder().encode(GraphQLRequest(query: query, variables: ["from": fromStr, "to": toStr]))
let res: GraphQLResponse<CalendarGraphQLData> = try await APIClient.shared.request(
    endpoint: "graphql",
    method: "POST",
    body: body
)
```

---

## 6. Cara Mencoba & Debugging (GraphQL Playground)

1. Pastikan backend Go sedang berjalan:
   ```bash
   cd backend && go run ./cmd/api/main.go
   ```
2. Buka browser dan kunjungi:
   ```text
   http://localhost:8080/api/v1/graphql
   ```
3. Di panel kiri, ketik query Anda:
   ```graphql
   query {
     projects {
       id
       name
       color
     }
   }
   ```
4. Di panel kiri bawah (**HTTP HEADERS**), masukkan Authorization header jika diperlukan:
   ```json
   {
     "Authorization": "Bearer TOKEN_ANDA_DISINI"
   }
   ```
5. Klik tombol **Play (▶)** di tengah. Hasilnya akan langsung muncul di panel kanan!
6. Anda juga bisa mengklik tab **DOCS** atau **SCHEMA** di sisi kanan playground untuk melihat semua tipe data dan fungsi query yang tersedia secara otomatis.

---

## Ringkasan Konsep Kunci
- **Query**: Operasi untuk membaca data (mirip HTTP `GET`).
- **Mutation**: Operasi untuk membuat/mengubah/menghapus data (mirip HTTP `POST`/`PUT`/`DELETE`).
- **Field**: Properti data spesifik yang Anda minta (misal: `id`, `title`, `due_date`).
- **Variables**: Nilai dinamis yang dikirim bersama query (misal: filter tanggal `$from` dan `$to`).
