import type { landing as EnLanding } from '../en/landing'

export const landing: typeof EnLanding = {
  badge: 'Workspace Generasi Baru untuk Engineer Berkecepatan Tinggi',
  heroTitle1: 'Developer Task &',
  heroTitle2: 'MCP AI Workspace',
  heroSubtitle:
    'Didesain khusus untuk developer yang bosan dengan project manager lambat. Kanban board interaktif, Markdown scratchpad instan, dan Model Context Protocol (MCP) server untuk AI agent di Cursor & Antigravity.',
  ctaGetStarted: 'Mulai Gratis',
  ctaOpenWorkspace: 'Buka Workspace',
  ctaInstantDemo: 'Coba Demo Instan',
  ctaDocs: 'Jelajahi Fitur',
  welcomeBack: 'Selamat datang kembali, {name}!',

  stats: {
    stat1Value: '100%',
    stat1Label: 'Fokus Developer',
    stat2Value: '< 10ms',
    stat2Label: 'Latensi Respon UI',
    stat3Value: 'MCP 2.0',
    stat3Label: 'Protokol AI Agent',
    stat4Value: 'OAuth 2.0',
    stat4Label: 'Google & Apple Siap',
  },

  features: {
    sectionTitle: 'Dirancang untuk Menjaga Flow State',
    sectionSubtitle: 'Semua kebutuhan produktivitas harianmu tanpa friksi atau gangguan konteks.',
    kanbanTitle: 'Developer Kanban Board',
    kanbanDesc:
      'Lacak tugas dari backlog, in progress, review, hingga selesai dengan checklist subtask, catatan progres, dan filter fokus Horizon.',
    mcpTitle: 'Integrasi Native MCP',
    mcpDesc:
      'Hubungkan AI agent di Cursor, Antigravity, atau Claude Code agar otomatis membaca, membuat, dan memperbarui status tugasmu saat coding.',
    notesTitle: 'Markdown Scratchpad',
    notesDesc:
      'Catat ide arsitektur, snippet kode, atau catatan cepat. Ubah catatan apa pun menjadi tiket tugas kanban hanya dengan 1 klik.',
    commandTitle: 'Navigasi Keyboard-First',
    commandDesc:
      'Command Palette universal (Cmd + K) memberikan akses secepat kilat ke pencarian tugas, catatan, dan tema tanpa melepas tangan dari keyboard.',
    speedTitle: 'Performa Super Cepat',
    speedDesc:
      'Dibangun dengan React 19, Tailwind CSS v4, Go Fiber, dan Redis. Respon antarmuka secepat kilat dengan sinkronisasi cloud real-time.',
    securityTitle: 'Keamanan Multi-Tenant Terisolasi',
    securityDesc:
      'Dukungan Google GIS One-Tap, Apple Sign-In dengan private relay email, validasi kriptografis token, serta isolasi penuh antar user.',
  },

  comparison: {
    title: 'Mengapa Developer Memilih TaskHub',
    subtitle: 'Tinggalkan tools project management enterprise yang berat dan lambat.',
    featureCol: 'Fitur & Kemampuan',
    traditionalCol: 'Tools Konvensional (Jira / Trello)',
    taskhubCol: 'TaskHub Workspace',
    row1Feature: 'Integrasi AI Agent',
    row1Trad: 'Plugin pihak ketiga berbayar / tidak ada',
    row1Hub: 'Native MCP Server (Cursor & Antigravity)',
    row2Feature: 'Kecepatan & Responsivitas',
    row2Trad: 'Reload halaman lambat berdetik-detik',
    row2Hub: 'Interaksi SPA instan di bawah 10ms',
    row3Feature: 'Developer Scratchpad',
    row3Trad: 'Tidak ada (harus pakai app eksternal)',
    row3Hub: 'Markdown notes terintegrasi dengan 1-klik konversi tugas',
    row4Feature: 'Pintasan Keyboard',
    row4Trad: 'Terbatas atau rumit',
    row4Hub: 'Full Cmd+K Command Palette cepat',
    row5Feature: 'Desain Minimalis Bebas Distraksi',
    row5Trad: 'Penuh form rumit dan bloat enterprise',
    row5Hub: 'Gaya dark mode Linear & Raycast yang elegan',
  },

  mcpShowcase: {
    title: 'Otomatisasi Tugas Langsung dari Editor Kode',
    subtitle:
      'Sambungkan Cursor, Antigravity, atau Claude Code ke workspace TaskHub kamu menggunakan konfigurasi MCP server native.',
    copied: 'Konfigurasi berhasil disalin ke clipboard!',
    copyButton: 'Salin Konfigurasi',
  },

  ctaBanner: {
    title: 'Siap Tingkatkan Kecepatan Engineering Tim Kamu?',
    subtitle: 'Masuk dengan Google atau Apple, atau coba langsung sandbox demo interaktif tanpa konfigurasi apa pun.',
    button: 'Mulai Gratis Sekarang',
    demoButton: 'Buka Demo Sandbox',
  },

  footer: {
    tagline: 'Developer Task & MCP AI Workspace. Dibangun untuk kecepatan software engineering modern.',
    status: 'Semua Sistem Beroperasi Normal',
    copyright: '© {year} TaskHub. Hak cipta dilindungi.',
  },
}
