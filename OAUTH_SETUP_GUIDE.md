# Panduan Setup OAuth: Google Cloud Console & Apple Developer Console

Panduan ini menjelaskan langkah demi langkah cara membuat kredensial resmi di **Google Cloud Console** dan **Apple Developer Account** untuk aplikasi **TaskHub**.

---

## 1. Setup Google Cloud Console (Google Sign-In & One Tap)

### Langkah 1: Buat Proyek di Google Cloud Console
1. Buka [Google Cloud Console](https://console.cloud.google.com/).
2. Login dengan akun Google Anda.
3. Di pojok kiri atas (sebelah logo Google Cloud), klik **Select a project** > **New Project**.
4. Beri nama proyek, misalnya: `TaskHub Workspace` > klik **Create**.

### Langkah 2: Konfigurasi OAuth Consent Screen
1. Buka menu **APIs & Services** > **OAuth consent screen** (atau [klik di sini](https://console.cloud.google.com/apis/credentials/consent)).
2. Pilih User Type:
   - **External** (jika ingin bisa login dengan akun Google apa pun) > klik **Create**.
3. Isi informasi aplikasi:
   - **App name**: `TaskHub`
   - **User support email**: Pilih email Anda.
   - **Developer contact information**: Isi email Anda.
   - Klik **Save and Continue**.
4. Pada langkah **Scopes**:
   - Klik **Add or Remove Scopes**.
   - Centang 3 scope dasar: `.../auth/userinfo.email`, `.../auth/userinfo.profile`, `openid`.
   - Klik **Update** > **Save and Continue**.
5. Pada langkah **Test users** (selama status proyek masih *Testing*):
   - Klik **Add Users** lalu masukkan alamat email Google yang ingin Anda gunakan untuk pengujian login.
   - Klik **Save and Continue** > **Back to Dashboard**.

### Langkah 3: Buat OAuth 2.0 Client ID (Web Application)
1. Buka menu **APIs & Services** > **Credentials** (atau [klik di sini](https://console.cloud.google.com/apis/credentials)).
2. Klik tombol **+ CREATE CREDENTIALS** di bagian atas > pilih **OAuth client ID**.
3. Pada **Application type**, pilih **Web application**.
4. Beri nama, misalnya: `TaskHub Web Client`.
5. Di bagian **Authorized JavaScript origins**, tambahkan URL origin:
   - `http://localhost:5173` (Frontend Vite dev)
   - `http://localhost:3000` (Backend API dev)
   - `https://your-domain.com` (Domain production Anda jika sudah ada)
6. Di bagian **Authorized redirect URIs**, tambahkan:
   - `http://localhost:5173`
   - `http://localhost:5173/auth/callback`
   - `http://localhost:3000/api/v1/auth/oauth`
7. Klik **CREATE**.
8. Modal pop-up akan menampilkan:
   - **Your Client ID** (misal: `1234567890-abcdefgh123456.apps.googleusercontent.com`)
   - **Your Client Secret** (misal: `GOCSPX-xxxxxxxxxxxxxxxxxxxxxxxx`)

### Langkah 4: Masukkan Kredensial ke `.env`
- Buka `frontend/.env`:
  ```env
  VITE_GOOGLE_CLIENT_ID=1234567890-abcdefgh123456.apps.googleusercontent.com
  ```
- Buka `backend/.env`:
  ```env
  GOOGLE_CLIENT_ID=1234567890-abcdefgh123456.apps.googleusercontent.com
  GOOGLE_CLIENT_SECRET=GOCSPX-xxxxxxxxxxxxxxxxxxxxxxxx
  GOOGLE_REDIRECT_URI=postmessage
  ```

---

## 2. Cara Kerja 2 Flow Google Sign-In di Aplikasi

1. **Flow Modal Dialog (Popup `initCodeClient`)**:
   - Saat pengguna mengklik tombol **Continue with Google**, browser akan membuka dialog popup resmi Google OAuth.
   - Pengguna memilih akun Google dan memberikan persetujuan.
   - Google mengembalikan Authorization `code` ke frontend via secure postMessage.
   - Frontend mengirimkan `code` ke Backend Go (`POST /api/v1/auth/oauth`).
   - Backend menukar `code` dengan Google Token Endpoint secara aman (server-to-server) menggunakan `GOOGLE_CLIENT_SECRET`, memverifikasi profil, dan menerbitkan token JWT aplikasi.

2. **Flow 1-Klik Instan (Google One Tap)**:
   - Jika pengguna **pernah login sebelumnya di perangkat ini** (`has_logged_in_before = true`), floating prompt 1-klik Google One Tap akan otomatis muncul di sudut kanan atas layar login.
   - Pengguna cukup klik nama akunnya 1 kali tanpa membuka popup baru.
   - ID Token JWT yang diterbitkan Google langsung diverifikasi oleh Backend Go.

3. **Perilaku Saat Logout**:
   - Saat pengguna menekan tombol **Sign Out** dari dashboard:
     - Sistem memanggil `google.accounts.id.disableAutoSelect()`.
     - Flag sesi `just_logged_out` dicatat sehingga Google One Tap **tidak langsung menginterupsi pengguna kembali**.
     - Pengguna diwajibkan mengklik tombol login secara manual untuk membuka popup dialog.

---

## 3. Setup Apple Developer Console (Sign in with Apple)

> *Catatan: Apple Developer Account berbayar ($99/tahun) diperlukan untuk membuat Apple Services ID.*

### Langkah 1: Buat App ID & Services ID
1. Buka [Apple Developer Account > Certificates, Identifiers & Profiles](https://developer.apple.com/account/resources/identifiers/list).
2. Di bagian **Identifiers**, buat **App ID** primer (misal: `com.taskhub.app`) dan centang kapabilitas **Sign in with Apple**.
3. Buat identifier kedua dengan tipe **Services IDs**:
   - **Description**: `TaskHub Web Authentication`
   - **Identifier**: `com.taskhub.web.auth`
   - Centang **Sign in with Apple** lalu klik **Configure**.
   - Masukkan Primary App ID Anda.
   - Masukkan **Domains and Subdomains**: domain web Anda (misal `yourdomain.com`).
   - Masukkan **Return URLs**: `https://yourdomain.com/auth/callback` (atau tunnel https untuk development seperti ngrok).
   - Klik **Save** dan **Continue**.

### Langkah 2: Buat Private Key (.p8)
1. Buka menu **Keys** > klik **+** untuk membuat Key baru.
2. Beri nama (misal: `TaskHub Apple Auth Key`).
3. Centang **Sign in with Apple** > klik **Configure** > pilih Primary App ID.
4. Klik **Register** lalu unduh berkas `.p8` (simpan dengan aman karena hanya bisa diunduh sekali).
5. Catat:
   - **Key ID** (10 karakter, tertera di halaman Key).
   - **Team ID** (10 karakter, tertera di pojok kanan atas akun Apple Developer).

### Langkah 3: Masukkan Kredensial Apple ke `.env`
- Di `frontend/.env`:
  ```env
  VITE_APPLE_CLIENT_ID=com.taskhub.web.auth
  ```
- Di `backend/.env`:
  ```env
  APPLE_CLIENT_ID=com.taskhub.web.auth
  APPLE_TEAM_ID=XXXXXXXXXX
  APPLE_KEY_ID=YYYYYYYYYY
  ```

---

## 4. Pengujian Cepat Tanpa Menunggu Setup Console

Jika Anda belum sempat membuka console atau sedang bekerja secara offline:
1. Klik tombol **Masuk Cepat sebagai Akun Demo** di halaman login.
2. Atau jika Anda mengklik tombol **Continue with Google** sebelum Client ID diisi di `.env`, dialog bantuan pengembang otomatis muncul dengan tombol shortcut instan ke mode demo.
