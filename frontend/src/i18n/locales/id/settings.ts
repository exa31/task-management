export const settings = {
  title: 'Pengaturan Akun & Modul',
  subtitle: 'Kelola profil, tautan OAuth, modul project task, dan preferensi akun Anda.',

  // Profile Section
  profile: {
    title: 'Profil Pengguna',
    changePhoto: 'Ganti',
    uploadCropWebp: 'Unggah & Auto Crop WebP',
    fullName: 'Nama Lengkap',
    email: 'Alamat Email (Akun Utama)',
    saveChanges: 'Simpan Perubahan',
    saving: 'Menyimpan...',
    updateSuccess: 'Profil berhasil diperbarui.',
    updateError: 'Gagal memperbarui profil.'
  },

  // OAuth Bindings Section
  oauth: {
    title: 'Tautan Akun OAuth',
    subtitle:
      'Tautkan akun Google atau Apple Anda untuk mempermudah login 1-klik di berbagai perangkat.',
    google: 'Google',
    apple: 'Apple ID',
    connected: 'Terhubung',
    notLinked: 'Belum ditautkan',
    disconnect: 'Putuskan Tautan',
    linkGoogle: 'Tautkan Google',
    linkApple: 'Tautkan Apple',
    googleSdkNotReady: 'Google SDK belum siap. Periksa koneksi internet Anda.',
    googleClientNotConfigured: 'Google Client ID belum dikonfigurasi.',
    googleAuthCancelled: 'Otorisasi Google dibatalkan atau gagal.',
    googleLinkedSuccess: 'Akun Google berhasil ditautkan.',
    googleLinkError: 'Gagal menautkan Google.',
    googleInitError: 'Gagal membuka Google.',
    appleSdkNotReady: 'Apple Sign-In SDK belum siap. Periksa koneksi internet Anda.',
    appleLinkedSuccess: 'Akun Apple berhasil ditautkan.',
    appleLinkError: 'Gagal menautkan Apple.',
    unbindSuccess: 'Tautan akun {provider} berhasil dibatalkan.',
    unbindError: 'Gagal membatalkan tautan.'
  },

  // Project Task Modules Section
  projects: {
    title: 'Modul / Project Task',
    subtitle: 'Kelompokkan task ke dalam modul/project spesifik (misal: ftracker, e-commerce).',
    addProject: 'Tambah Project',
    loading: 'Memuat project...',
    empty: 'Belum ada modul project. Buat project baru untuk mengelompokkan task Anda!',
    tasksConnected: '{count} task terhubung',
    editModalTitle: 'Edit Modul Project',
    createModalTitle: 'Tambah Modul Project',
    nameLabel: 'Nama Project',
    namePlaceholder: 'Misal: ftracker, e-commerce',
    descLabel: 'Deskripsi (Opsional)',
    descPlaceholder: 'Keterangan singkat project...',
    colorLabel: 'Warna Identitas',
    cancel: 'Batal',
    save: 'Simpan',
    create: 'Buat Project',
    deleteHeader: 'Hapus Modul Project',
    deleteMessage: 'Hapus project "{name}"? Task yang terhubung tidak akan dihapus.',
    deleteAccept: 'Hapus',
    deleteReject: 'Batal'
  },

  // MCP AI Link Section
  mcp: {
    title: 'Model Context Protocol (MCP) & AI',
    subtitle:
      'Buat Personal Access Token untuk menghubungkan AI agent (Antigravity, Claude, Cline) langsung dengan task dan catatan Anda.',
    button: 'Buka Pengaturan MCP'
  },

  // Notification & Due Date Warning Preferences
  notifications: {
    title: 'Preferensi Notifikasi & Peringatan Deadline',
    subtitle: 'Atur zona waktu, batas waktu peringatan sebelum deadline tiba, dan pengingat email otomatis.',
    timezoneLabel: 'Zona Waktu Tampilan & Pengingat',
    timezoneHint: 'Seluruh notifikasi tenggat waktu dan format email akan disesuaikan dengan zona waktu ini.',
    thresholdLabel: 'Batas Peringatan Menjelang Deadline',
    thresholdHint: 'TaskHub akan mengirim notifikasi email (Resend), realtime (SSE), dan mobile push (FCM) saat tugas mendekati batas tenggat waktu.',
    threshold15m: '15 Menit sebelum deadline',
    threshold1h: '1 Jam sebelum deadline',
    threshold3h: '3 Jam sebelum deadline',
    threshold12h: '12 Jam sebelum deadline',
    threshold24h: '24 Jam (1 Hari) sebelum deadline',
    threshold48h: '48 Jam (2 Hari) sebelum deadline',
    emailNotificationsLabel: 'Notifikasi Email Deadline (via Resend)',
    emailNotificationsDesc: 'Kirim email peringatan ke akun Anda saat tugas mencapai ambang batas peringatan.',
    savePreferences: 'Simpan Preferensi Notifikasi',
    saving: 'Menyimpan preferensi...',
    savedSuccess: 'Preferensi notifikasi berhasil disimpan.',
    savedError: 'Gagal memperbarui preferensi notifikasi.'
  },

  // Danger Zone Section
  dangerZone: {
    title: 'Zona Bahaya',
    subtitle:
      'Tindakan ini tidak dapat dibatalkan. Menghapus akun akan membersihkan seluruh data secara permanen, termasuk semua task, subtask, notes, file avatar di penyimpanan, dan sesi aktif.',
    deleteAccountBtn: 'Hapus Akun Permanen',
    modalTitle: 'Hapus Akun Permanen?',
    modalDesc:
      'Tindakan ini akan menghapus seluruh data akun secara permanen dari basis data, termasuk semua task, project, notes, dan foto profil di penyimpanan, serta mencabut seluruh sesi aktif. Tindakan ini tidak dapat dibatalkan.',
    typeDeletePrompt: 'Ketik DELETE untuk melanjutkan:',
    errorConfirmation: 'Harap ketik DELETE untuk mengonfirmasi.',
    cancelBtn: 'Batalkan',
    confirmDeleteBtn: 'Saya Mengerti, Hapus Total',
    deleting: 'Menghapus total...',
    deletedToast: 'Akun Anda telah dihapus secara permanen.',
    deleteErrorToast: 'Gagal menghapus akun.'
  },

  // Avatar Crop Modal
  avatarModal: {
    title: 'Foto Profil',
    subtitle: 'Auto-crop 1:1 & konversi otomatis ke WebP',
    uploadPrompt: 'Pilih atau seret foto ke sini',
    supportedFormats: 'PNG, JPG, JPEG, WEBP (maks. 10MB)',
    changePhoto: 'Ganti Foto',
    resetPosition: 'Reset Posisi',
    cancel: 'Batal',
    save: 'Simpan Avatar',
    saving: 'Mengompres & Menyimpan...',
    invalidFileType: 'Harap pilih file gambar (JPG, PNG, atau WEBP).',
    fileTooLarge: 'Ukuran file maksimal 10MB.',
    convertError: 'Gagal mengonversi gambar ke format WebP.',
    uploadError: 'Gagal mengunggah avatar.',
    uploadSuccess: 'Foto profil berhasil diperbarui.'
  }
}
