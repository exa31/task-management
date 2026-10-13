import SwiftUI

/// Supported Application Languages
public enum AppLanguage: String, CaseIterable, Identifiable, Codable {
    case en = "en"
    case id = "id"

    public var id: String { rawValue }

    public var displayName: String {
        switch self {
        case .en: return "English"
        case .id: return "Bahasa Indonesia"
        }
    }

    public var flag: String {
        switch self {
        case .en: return "🇬🇧"
        case .id: return "🇮🇩"
        }
    }
}

/// Structured, Observable Internationalization (I18n) Engine for TaskHub
/// Provides centralized, reactive dictionary-based translations with zero manual ternary checks in views.
@Observable
public final class I18n {
    public static let shared = I18n()

    public var currentLanguage: AppLanguage {
        didSet {
            UserDefaults.standard.set(currentLanguage.rawValue, forKey: "app_language")
        }
    }

    private init() {
        let saved = UserDefaults.standard.string(forKey: "app_language") ?? "en"
        self.currentLanguage = AppLanguage(rawValue: saved) ?? .en
    }

    /// Change language and notify all observing SwiftUI views
    public func setLanguage(_ lang: AppLanguage) {
        withAnimation(.easeInOut(duration: 0.2)) {
            self.currentLanguage = lang
        }
    }

    /// Toggle between English and Indonesian
    public func toggleLanguage() {
        setLanguage(currentLanguage == .en ? .id : .en)
    }

    /// Lookup translation key in the active dictionary
    public func t(_ key: String, default defaultVal: String? = nil) -> String {
        if let val = Translations.data[currentLanguage]?[key] {
            return val
        }
        // Fallback to English dictionary if key is missing in active locale
        if currentLanguage != .en, let fallback = Translations.data[.en]?[key] {
            return fallback
        }
        return defaultVal ?? key
    }

    /// Formatted lookup with arguments (e.g. "%lld tasks")
    public func t(_ key: String, _ args: CVarArg...) -> String {
        let format = t(key)
        return String(format: format, arguments: args)
    }
}

// MARK: - String Extension for Clean Syntax
public extension String {
    /// Shorthand lookup on string keys, e.g. `"nav.tasks".localized`
    var localized: String {
        I18n.shared.t(self)
    }

    /// Formatted lookup, e.g. `"tasks.subtasks.count".localized(done, total)`
    func localized(with args: CVarArg...) -> String {
        let format = I18n.shared.t(self)
        return String(format: format, arguments: args)
    }
}

// MARK: - Translation Dictionaries
public enum Translations {
    public static let data: [AppLanguage: [String: String]] = [
        .en: [
            // Nav
            "nav.tasks": "Tasks",
            "nav.projects": "Projects",
            "nav.notes": "Notes",
            "nav.calendar": "Calendar",
            "nav.reminders": "Reminders",
            "nav.settings": "Settings",

            // Calendar & Work Log
            "calendar.title": "Calendar & Work Log",
            "calendar.subtitle": "Tasks, reminders, and activity history",
            "calendar.today": "Today",
            "calendar.all": "All",
            "calendar.tasks": "Tasks",
            "calendar.reminders": "Reminders",
            "calendar.activities": "Work Log",
            "calendar.noItems": "No items for this date",
            "calendar.noTasks": "No tasks due on this date",
            "calendar.noReminders": "No reminders scheduled",
            "calendar.noActivities": "No activity logged on this date",
            "calendar.workLog": "Work Log & Activities",
            "calendar.tasksDue": "Tasks Due",
            "calendar.scheduledReminders": "Reminders",
            "calendar.graphqlPowered": "GraphQL Single Query",
            "calendar.snooze": "Snooze",
            "calendar.snooze1h": "+1h",
            "calendar.dismiss": "Dismiss",

            // Auth
            "auth.title": "Task Hub",
            "auth.subtitle": "Smart productivity synchronized in real-time.",
            "auth.sync.title": "Real-time Cloud Sync",
            "auth.sync.desc": "Instant synchronization across all your devices",
            "auth.projects.title": "Project Management",
            "auth.projects.desc": "Organize workspace modules, tasks & checklists",
            "auth.reminders.title": "Smart Reminders",
            "auth.reminders.desc": "Custom deadlines with automated push alerts",
            "auth.signInApple": "Continue with Apple",
            "auth.signInGoogle": "Continue with Google",
            "auth.secure": "Secure End-to-End Cloud Workspace",
            "auth.signOut": "Sign Out",
            "auth.error.sessionExpired": "Your session has expired or is invalid. Please sign in again.",
            "auth.error.windowNotFound": "Could not find active application window for authentication.",
            "auth.error.googleTokenFailed": "Failed to verify Google authentication token.",
            "auth.error.appleTokenFailed": "Failed to extract Apple ID Token.",

            // Tasks
            "tasks.title": "Tasks",
            "tasks.new": "New Task",
            "tasks.searchPlaceholder": "Search tasks, tags, or description...",
            "tasks.allProjects": "All Projects",
            "tasks.allStatuses": "All Statuses",
            "tasks.allPriorities": "All Priorities",
            "tasks.focusOnly": "Focus Only",
            "tasks.resetFilters": "Reset Filters",
            "tasks.resetAllFilters": "Reset All Filters",
            "tasks.emptyTitle": "No Tasks Found",
            "tasks.emptyDesc": "Create your first task or adjust active filters.",
            "tasks.noTasksFilterTitle": "No Matching Tasks",
            "tasks.noTasksFilterDesc": "No tasks match the active filters or search query.",
            "tasks.createTask": "Create Task",
            "tasks.saveFailed": "Failed to save task.",
            "tasks.noSubtasks": "No subtasks checklist.",
            "tasks.lastProgressNote": "LAST PROGRESS NOTE",
            "tasks.subtasksCount": "%lld of %lld subtasks",
            "tasks.detailTitle": "Task Details",
            "tasks.edit": "Edit",
            "tasks.editTitle": "Edit Task",
            "tasks.createTitle": "Create New Task",
            "tasks.sectionInfo": "Task Information",
            "tasks.fieldTitle": "Task Title *",
            "tasks.fieldDescription": "Description (Optional)",
            "tasks.sectionStatusPriority": "Status & Priority",
            "tasks.statusLabel": "Status",
            "tasks.priorityLabel": "Priority",
            "tasks.sectionProject": "Project",
            "tasks.selectProject": "Select Project",
            "tasks.noProject": "No Project",
            "tasks.sectionDueDate": "Due Date",
            "tasks.setDueDate": "Set Due Date",
            "tasks.dateTime": "Date & Time",
            "tasks.sectionFocus": "Focus",
            "tasks.focusToday": "Today's Focus",
            "tasks.activeCount": "%lld active",
            "tasks.sectionSubtasks": "Subtasks Checklist",
            "tasks.addSubtaskPlaceholder": "Add new subtask...",
            "tasks.sectionTags": "Tags / Labels",
            "tasks.tagsPlaceholder": "Comma separated (e.g. bug, api, mobile)",
            "tasks.statusHeader": "TASK STATUS",
            "tasks.markDone": "Mark as Done",
            "tasks.markUndone": "Mark as Incomplete",
            "tasks.focusToggle": "Focus",
            "tasks.unfocusToggle": "Unfocus",
            "tasks.tabDetails": "Details",
            "tasks.tabActivities": "Activities & Logs",
            "tasks.postNotePlaceholder": "Write progress note or log update...",
            "tasks.send": "Send",
            "tasks.loadingLogs": "Loading log history...",
            "tasks.noLogsTitle": "No Log History Yet",
            "tasks.noLogsDesc": "All status changes, subtasks, and progress notes will be recorded here chronologically.",
            "tasks.activityCreated": "Task Created",
            "tasks.activityUpdated": "Task Updated",
            "tasks.activityProgress": "Progress Note",
            "tasks.activitySubtaskCreated": "Subtask Added",
            "tasks.activitySubtaskUpdated": "Subtask Updated",
            "tasks.activitySubtaskDeleted": "Subtask Deleted",
            "tasks.activityFocusChanged": "Focus Toggled",
            "tasks.activityDefault": "Log Note",
            "tasks.relativeJustNow": "Just now",

            // Projects
            "projects.title": "Projects",
            "projects.new": "New Project",
            "projects.emptyTitle": "No Projects Found",
            "projects.emptyDesc": "Create a project module to organize tasks.",
            "projects.emptyGuidance": "Group tasks into projects to track progress easily.",
            "projects.createFirst": "Create First Project",
            "projects.noMatchingProjects": "No Projects Found",
            "projects.noProjectsMatchingDesc": "No projects match the keyword \"%@\".",
            "projects.deleteConfirmation": "Delete Project \"%@\"",
            "projects.deleteWarningNotice": "Tasks linked to this project will not be deleted, but will be detached from this project.",
            "projects.totalProjects": "Total Projects",
            "projects.sectionInfo": "Project Information",
            "projects.fieldName": "Project Name *",
            "projects.fieldDesc": "Description (Optional)",
            "projects.sectionColor": "Color Selection",
            "projects.colorHint": "Choose an accent color for this module/project:",
            "projects.createTitle": "Create New Project",
            "projects.editTitle": "Edit Project",
            "projects.metricTotal": "Total Projects",
            "projects.metricTotalSub": "Active Projects",
            "projects.metricTasks": "Project Tasks",
            "projects.metricTasksActive": "%lld active",
            "projects.metricDone": "Completed",
            "projects.metricDoneSub": "Tasks finished",
            "projects.metricRate": "Completion Rate",
            "projects.metricRateSub": "Progress ratio",
            "projects.noTasksCategory": "No tasks in this category yet",
            "projects.addTaskButton": "Add Task in this Project",

            // Notes
            "notes.title": "Notes",
            "notes.new": "New Note",
            "notes.emptyTitle": "No Notes Found",
            "notes.emptyDesc": "Write memos, scratchpads, and ideas here.",
            "notes.loading": "Loading notes...",
            "notes.sectionInfo": "Note Information",
            "notes.fieldTitle": "Note Title *",
            "notes.fieldType": "Type",
            "notes.typeActionable": "Actionable (Convertible to Task)",
            "notes.typeKnowledge": "Knowledge (Insights)",
            "notes.fieldContent": "Markdown Content",
            "notes.pinToTop": "Pin to Top",
            "notes.searchPlaceholder": "Search notes or tags...",
            "notes.filterAll": "All",
            "notes.pin": "Pin Note",
            "notes.unpin": "Unpin Note",
            "notes.edit": "Edit Note",
            "notes.delete": "Delete Note",
            "notes.deleteConfirmTitle": "Delete Note?",
            "notes.deleteConfirmMessage": "This action cannot be undone.",
            "notes.convertToTask": "Convert to Task",
            "notes.convertConfirmTitle": "Convert Note to Task?",
            "notes.convertConfirmMessage": "A new task will automatically be created from this note in your task list.",
            "notes.convertedSuccess": "Note successfully converted into a Task!",

            // Reminders
            "reminders.title": "Reminders & Alerts",
            "reminders.new": "New Reminder",
            "reminders.emptyTitle": "No Reminders Found",
            "reminders.emptyDesc": "Stay ahead of schedule by adding deadlines and alerts.",
            "reminders.createFirst": "Create First Reminder",
            "reminders.add": "Add",
            "reminders.filter.all": "All",
            "reminders.filter.pending": "Pending",
            "reminders.filter.sent": "Triggered",
            "reminders.filter.dismissed": "Completed",
            "reminders.field.title": "Reminder title",
            "reminders.field.notes": "Notes / Description (optional)",
            "reminders.section.info": "Reminder Information",
            "reminders.section.schedule": "Schedule",
            "reminders.section.association": "Task Association (Optional)",
            "reminders.remindAt": "Remind At",
            "reminders.presets": "Quick Presets",
            "reminders.tomorrowMorning": "Tomorrow 9 AM",
            "reminders.taskNone": "None (General Reminder)",
            "reminders.selectTask": "Select Task",
            "reminders.linkedTask": "Linked Task",
            "reminders.searchTasksPlaceholder": "Search tasks by title...",
            "reminders.noTasksFound": "No matching tasks found",
            "reminders.unlinkTask": "Unlink Task",
            "reminders.snooze": "Snooze 15m",
            "reminders.done": "Done",
            "reminders.snoozedCount": "Snoozed %lldx",
            "reminders.completedNotice": "Reminder Completed",
            "reminders.liveAlert": "Live Alert",

            // Settings
            "settings.title": "Settings & Profile",
            "settings.profileInfo": "Profile Information",
            "settings.changeAvatar": "Change Profile Photo",
            "settings.fullName": "Full Name",
            "settings.emailAddress": "Email Address",
            "settings.saveProfile": "Save Changes",
            "settings.profileUpdated": "Profile updated successfully!",
            "settings.avatarUpdated": "Profile photo updated successfully!",
            "settings.preferencesSaved": "Notification preferences saved successfully!",
            "settings.accountDisconnected": "%@ account disconnected successfully.",
            "settings.appleConnected": "Apple ID connected successfully.",
            "settings.deleteConfirmPrompt": "Type \"DELETE\" to confirm account deletion.",
            "settings.deletePermanentNotice": "This action cannot be undone. All tasks, projects, notes, and records will be deleted forever.",
            "settings.linkedAccounts": "Linked Accounts",
            "settings.connected": "Connected",
            "settings.disconnect": "Disconnect",
            "settings.connect": "Connect",
            "settings.notifications": "Notifications & Alerts",
            "settings.timezone": "Display & Reminder Timezone",
            "settings.threshold": "Due Date Warning Alert Threshold",
            "settings.thresholdHint": "TaskHub will send email alerts and push notifications when a task approaches its deadline.",
            "settings.emailNotifications": "Email Deadline Alerts",
            "settings.emailNotificationsDesc": "Send automated email reminders when tasks approach overdue.",
            "settings.savePreferences": "Save Notification Preferences",
            "settings.appearance": "Appearance & Language",
            "settings.themeMode": "Theme Mode",
            "settings.theme.system": "System",
            "settings.theme.light": "Light",
            "settings.theme.dark": "Dark",
            "settings.language": "Language",
            "settings.appVersion": "App Version",
            "settings.dangerZone": "Danger Zone",
            "settings.deleteAccount": "Delete Account Permanently",
            "settings.deleteWarning": "Once deleted, all tasks, notes, checklists, and projects will be permanently destroyed.",
            "settings.mcpIntegration": "AI & MCP Integration",
            "settings.mcpIntegrationDesc": "Model Context Protocol & Personal Access Tokens for AI coding agents",
            "settings.signOut": "Sign Out",
            "settings.signOutDialogTitle": "Confirm Sign Out",
            "settings.signOutConfirmButton": "Sign Out",
            "settings.signOutDialogMessage": "You will be signed out of this session and will need to log in again to access your data.",
            "settings.disconnectDialogTitle": "Disconnect Account?",
            "settings.disconnectConfirmButton": "Disconnect %@",
            "settings.disconnectDialogMessage": "Are you sure you want to disconnect %@ account?",
            "settings.linkedAccountsDesc": "Connect your accounts for easy multi-platform sign-in.",
            "settings.notConnected": "Not Connected",
            "settings.notificationsDesc": "Set reminder timezone and warning tolerance before tasks are due.",
            "settings.threshold15m": "15 minutes before due date",
            "settings.threshold1h": "1 hour before due date",
            "settings.threshold3h": "3 hours before due date",
            "settings.threshold12h": "12 hours before due date",
            "settings.threshold24h": "24 hours (1 day) before due date",
            "settings.threshold48h": "48 hours (2 days) before due date",
            "settings.deleteConfirmTitle": "Confirm Deletion",

            // MCP (Model Context Protocol)
            "mcp.title": "Model Context Protocol (MCP)",
            "mcp.subtitle": "Connect AI coding agents to your Task Hub workspace",
            "mcp.endpoint": "MCP Server Endpoint",
            "mcp.endpointCopied": "MCP endpoint copied to clipboard",
            "mcp.tokensTitle": "Personal Access Tokens",
            "mcp.tokensSubtitle": "Generate dedicated tokens for AI agents (Claude, Cursor, Antigravity IDE)",
            "mcp.generateToken": "Generate New Token",
            "mcp.tokenNamePlaceholder": "Token name (e.g. Cursor Mac, Antigravity IDE)",
            "mcp.tokenCreatedTitle": "Token Generated Successfully",
            "mcp.tokenCreatedWarning": "Make sure to copy your Personal Access Token now. You will not be able to see it again!",
            "mcp.copyToken": "Copy Token",
            "mcp.tokenCopied": "Token copied to clipboard",
            "mcp.revoke": "Revoke",
            "mcp.revokeConfirmTitle": "Revoke Token?",
            "mcp.revokeConfirmMessage": "Are you sure you want to revoke '%@'? Any AI agent or client using this token will immediately lose access.",
            "mcp.noTokens": "No access tokens generated yet.",
            "mcp.lastUsed": "Last used: %@",
            "mcp.neverUsed": "Never used",
            "mcp.createdOn": "Created: %@",
            "mcp.configSnippets": "Quick Setup Snippets",
            "mcp.copyConfig": "Copy Config",
            "mcp.configCopied": "Configuration copied to clipboard",
            "mcp.antigravityConfig": "Antigravity IDE / VS Code",
            "mcp.cursorConfig": "Cursor / Windsurf",
            "mcp.claudeConfig": "Claude Desktop",

            // General
            "general.create": "Create",
            "general.save": "Save",
            "general.cancel": "Cancel",
            "general.delete": "Delete",
            "general.close": "Close",
            "general.done": "Done",
            "general.back": "Back"
        ],
        .id: [
            // Nav
            "nav.tasks": "Tugas",
            "nav.projects": "Proyek",
            "nav.notes": "Catatan",
            "nav.calendar": "Kalender",
            "nav.reminders": "Pengingat",
            "nav.settings": "Pengaturan",

            // Calendar & Work Log
            "calendar.title": "Kalender & Riwayat",
            "calendar.subtitle": "Tugas, pengingat, dan riwayat aktivitas",
            "calendar.today": "Hari Ini",
            "calendar.all": "Semua",
            "calendar.tasks": "Tugas",
            "calendar.reminders": "Pengingat",
            "calendar.activities": "Riwayat",
            "calendar.noItems": "Tidak ada agenda di tanggal ini",
            "calendar.noTasks": "Tidak ada tugas jatuh tempo",
            "calendar.noReminders": "Tidak ada pengingat terjadwal",
            "calendar.noActivities": "Belum ada riwayat aktivitas",
            "calendar.workLog": "Riwayat Aktivitas & Kerja",
            "calendar.tasksDue": "Tugas Jatuh Tempo",
            "calendar.scheduledReminders": "Pengingat",
            "calendar.graphqlPowered": "Single Query GraphQL",
            "calendar.snooze": "Tunda",
            "calendar.snooze1h": "+1j",
            "calendar.dismiss": "Tutup",

            // Auth
            "auth.title": "Task Hub",
            "auth.subtitle": "Produktivitas cerdas tersinkronisasi real-time.",
            "auth.sync.title": "Sinkronisasi Real-time",
            "auth.sync.desc": "Sinkronisasi instan di seluruh perangkat Anda",
            "auth.projects.title": "Manajemen Proyek",
            "auth.projects.desc": "Kelola modul kerja, tugas & checklist",
            "auth.reminders.title": "Pengingat Pintar",
            "auth.reminders.desc": "Notifikasi deadline & alert otomatis",
            "auth.signInApple": "Masuk dengan Apple",
            "auth.signInGoogle": "Lanjutkan dengan Google",
            "auth.secure": "Workspace Cloud Aman & Terenkripsi",
            "auth.signOut": "Keluar (Sign Out)",
            "auth.error.sessionExpired": "Sesi telah berakhir atau tidak valid. Silakan masuk kembali.",
            "auth.error.windowNotFound": "Tidak dapat menemukan jendela aplikasi aktif untuk autentikasi.",
            "auth.error.googleTokenFailed": "Gagal memverifikasi token autentikasi Google.",
            "auth.error.appleTokenFailed": "Gagal mengekstrak Apple ID Token.",

            // Tasks
            "tasks.title": "Tugas",
            "tasks.new": "Tugas Baru",
            "tasks.searchPlaceholder": "Cari tugas, tag, atau deskripsi…",
            "tasks.allProjects": "Semua Proyek",
            "tasks.allStatuses": "Semua Status",
            "tasks.allPriorities": "Semua Prioritas",
            "tasks.focusOnly": "Hanya Fokus",
            "tasks.resetFilters": "Reset Filter",
            "tasks.resetAllFilters": "Reset Semua Filter",
            "tasks.emptyTitle": "Tidak Ada Tugas",
            "tasks.emptyDesc": "Buat tugas baru atau sesuaikan filter yang aktif.",
            "tasks.noTasksFilterTitle": "Tidak Ada Tugas yang Sesuai",
            "tasks.noTasksFilterDesc": "Tidak ada tugas yang sesuai dengan kriteria filter atau pencarian saat ini.",
            "tasks.createTask": "Buat Tugas",
            "tasks.saveFailed": "Gagal menyimpan tugas.",
            "tasks.noSubtasks": "Tidak ada subtask checklist.",
            "tasks.lastProgressNote": "CATATAN PROGRES TERAKHIR",
            "tasks.subtasksCount": "%lld dari %lld subtugas",
            "tasks.detailTitle": "Detail Tugas",
            "tasks.edit": "Edit",
            "tasks.editTitle": "Edit Tugas",
            "tasks.createTitle": "Buat Tugas Baru",
            "tasks.sectionInfo": "Informasi Tugas",
            "tasks.fieldTitle": "Judul Tugas *",
            "tasks.fieldDescription": "Deskripsi (Opsional)",
            "tasks.sectionStatusPriority": "Status & Prioritas",
            "tasks.statusLabel": "Status",
            "tasks.priorityLabel": "Prioritas",
            "tasks.sectionProject": "Proyek",
            "tasks.selectProject": "Pilih Proyek",
            "tasks.noProject": "Tanpa Proyek",
            "tasks.sectionDueDate": "Tenggat Waktu",
            "tasks.setDueDate": "Tetapkan Batas Waktu",
            "tasks.dateTime": "Tanggal & Jam",
            "tasks.sectionFocus": "Fokus",
            "tasks.focusToday": "Fokus Hari Ini",
            "tasks.activeCount": "%lld aktif",
            "tasks.sectionSubtasks": "Checklist Subtask",
            "tasks.addSubtaskPlaceholder": "Tambah subtask baru…",
            "tasks.sectionTags": "Tags / Label",
            "tasks.tagsPlaceholder": "Pisahkan koma (contoh: bug, api, mobile)",
            "tasks.statusHeader": "STATUS TUGAS",
            "tasks.markDone": "Tandai Selesai",
            "tasks.markUndone": "Tandai Belum Selesai",
            "tasks.focusToggle": "Fokus",
            "tasks.unfocusToggle": "Batal Fokus",
            "tasks.tabDetails": "Detail",
            "tasks.tabActivities": "Aktivitas & Log",
            "tasks.postNotePlaceholder": "Tulis catatan progres atau update log...",
            "tasks.send": "Kirim",
            "tasks.loadingLogs": "Memuat riwayat log...",
            "tasks.noLogsTitle": "Belum Ada Riwayat Log",
            "tasks.noLogsDesc": "Semua perubahan status, subtask, dan catatan progres akan tercatat di sini secara kronologis.",
            "tasks.activityCreated": "Tugas Dibuat",
            "tasks.activityUpdated": "Tugas Diperbarui",
            "tasks.activityProgress": "Catatan Progres",
            "tasks.activitySubtaskCreated": "Subtask Ditambahkan",
            "tasks.activitySubtaskUpdated": "Subtask Diperbarui",
            "tasks.activitySubtaskDeleted": "Subtask Dihapus",
            "tasks.activityFocusChanged": "Fokus Diubah",
            "tasks.activityDefault": "Catatan Log",
            "tasks.relativeJustNow": "Baru saja",

            // Projects
            "projects.title": "Proyek",
            "projects.new": "Proyek Baru",
            "projects.emptyTitle": "Belum Ada Proyek",
            "projects.emptyDesc": "Buat modul proyek untuk mengorganisir tugas-tugas Anda.",
            "projects.emptyGuidance": "Kelompokkan tugas ke dalam proyek untuk memudahkan pelacakan progres.",
            "projects.createFirst": "Buat Proyek Pertama",
            "projects.noMatchingProjects": "Tidak Ada Hasil",
            "projects.noProjectsMatchingDesc": "Tidak ada proyek yang sesuai dengan kata kunci \"%@\".",
            "projects.deleteConfirmation": "Hapus Proyek \"%@\"",
            "projects.deleteWarningNotice": "Tugas yang terhubung dengan proyek ini tidak akan dihapus, tetapi relasi proyek akan dilepas.",
            "projects.totalProjects": "Total Proyek",
            "projects.sectionInfo": "Informasi Proyek",
            "projects.fieldName": "Nama Proyek *",
            "projects.fieldDesc": "Deskripsi (Opsional)",
            "projects.sectionColor": "Pilihan Warna",
            "projects.colorHint": "Pilih warna aksen untuk modul/proyek ini:",
            "projects.createTitle": "Buat Proyek Baru",
            "projects.editTitle": "Edit Proyek",
            "projects.metricTotal": "Total Proyek",
            "projects.metricTotalSub": "Proyek Aktif",
            "projects.metricTasks": "Tugas Proyek",
            "projects.metricTasksActive": "%lld berjalan",
            "projects.metricDone": "Selesai",
            "projects.metricDoneSub": "Tugas tuntas",
            "projects.metricRate": "Tingkat Selesai",
            "projects.metricRateSub": "Rasio progres",
            "projects.noTasksCategory": "Belum ada tugas di kategori ini",
            "projects.addTaskButton": "Tambah Tugas di Proyek Ini",

            // Notes
            "notes.title": "Catatan",
            "notes.new": "Catatan Baru",
            "notes.emptyTitle": "Belum Ada Catatan",
            "notes.emptyDesc": "Tulis memo, ide, dan catatan penting di sini.",
            "notes.loading": "Memuat catatan…",
            "notes.sectionInfo": "Informasi Catatan",
            "notes.fieldTitle": "Judul Catatan *",
            "notes.fieldType": "Tipe",
            "notes.typeActionable": "Actionable (Bisa jadi tugas)",
            "notes.typeKnowledge": "Knowledge (Wawasan)",
            "notes.fieldContent": "Isi Catatan Markdown",
            "notes.pinToTop": "Pin ke Atas",
            "notes.searchPlaceholder": "Cari catatan atau tag…",
            "notes.filterAll": "Semua",
            "notes.pin": "Pin Catatan",
            "notes.unpin": "Lepas Pin",
            "notes.edit": "Edit Catatan",
            "notes.delete": "Hapus Catatan",
            "notes.deleteConfirmTitle": "Hapus Catatan?",
            "notes.deleteConfirmMessage": "Tindakan ini tidak dapat dibatalkan.",
            "notes.convertToTask": "Konversi ke Tugas",
            "notes.convertConfirmTitle": "Konversi Catatan ke Tugas?",
            "notes.convertConfirmMessage": "Catatan ini akan otomatis dibuatkan tugas baru di daftar tugas Anda.",
            "notes.convertedSuccess": "Catatan berhasil dikonversi menjadi Tugas!",

            // Reminders
            "reminders.title": "Pengingat & Alert",
            "reminders.new": "Pengingat Baru",
            "reminders.emptyTitle": "Tidak Ada Pengingat",
            "reminders.emptyDesc": "Semua tugas & pengingat Anda sudah tertangani dengan baik.",
            "reminders.createFirst": "Tambah Pengingat Baru",
            "reminders.add": "Tambah",
            "reminders.filter.all": "Semua",
            "reminders.filter.pending": "Menunggu",
            "reminders.filter.sent": "Terkirim",
            "reminders.filter.dismissed": "Selesai",
            "reminders.field.title": "Judul pengingat",
            "reminders.field.notes": "Catatan / Deskripsi (opsional)",
            "reminders.section.info": "Informasi Pengingat",
            "reminders.section.schedule": "Jadwal Waktu",
            "reminders.section.association": "Asosiasi Tugas (Opsional)",
            "reminders.remindAt": "Waktu Pengingat",
            "reminders.presets": "Pilihan Cepat",
            "reminders.tomorrowMorning": "Besok 09:00",
            "reminders.taskNone": "Tidak ada (Pengingat Umum)",
            "reminders.selectTask": "Pilih Tugas",
            "reminders.linkedTask": "Tugas Terhubung",
            "reminders.searchTasksPlaceholder": "Cari tugas berdasarkan judul...",
            "reminders.noTasksFound": "Tidak ada tugas yang cocok",
            "reminders.unlinkTask": "Hapus Tautan",
            "reminders.snooze": "Tunda 15 mnt",
            "reminders.done": "Selesai",
            "reminders.snoozedCount": "Ditunda %lldx",
            "reminders.completedNotice": "Pengingat Telah Selesai",
            "reminders.liveAlert": "Notifikasi Real-time",

            // Settings
            "settings.title": "Pengaturan & Profil",
            "settings.profileInfo": "Informasi Profil",
            "settings.changeAvatar": "Ganti Foto Profil",
            "settings.fullName": "Nama Lengkap",
            "settings.emailAddress": "Alamat Email",
            "settings.saveProfile": "Simpan Perubahan",
            "settings.profileUpdated": "Profil berhasil diperbarui!",
            "settings.avatarUpdated": "Foto profil berhasil diperbarui!",
            "settings.preferencesSaved": "Preferensi notifikasi berhasil disimpan!",
            "settings.accountDisconnected": "Akun %@ berhasil diputuskan.",
            "settings.appleConnected": "Apple ID berhasil dihubungkan.",
            "settings.deleteConfirmPrompt": "Ketik \"DELETE\" untuk mengonfirmasi penghapusan akun.",
            "settings.deletePermanentNotice": "Tindakan ini tidak dapat dibatalkan. Semua data tugas, proyek, dan catatan Anda akan dihapus selamanya.",
            "settings.linkedAccounts": "Akun Terhubung",
            "settings.connected": "Terhubung",
            "settings.disconnect": "Putuskan",
            "settings.connect": "Hubungkan",
            "settings.notifications": "Preferensi Notifikasi",
            "settings.timezone": "Zona Waktu Tampilan & Pengingat",
            "settings.threshold": "Batas Peringatan Batas Waktu",
            "settings.thresholdHint": "TaskHub akan mengirim notifikasi email dan pemberitahuan saat tugas mendekati batas tenggat waktu.",
            "settings.emailNotifications": "Kirim Email Peringatan",
            "settings.emailNotificationsDesc": "Kirim email otomatis saat tugas mendekati batas tenggat waktu.",
            "settings.savePreferences": "Simpan Preferensi Notifikasi",
            "settings.appearance": "Tampilan & Bahasa",
            "settings.themeMode": "Mode Tema",
            "settings.theme.system": "Sistem",
            "settings.theme.light": "Terang",
            "settings.theme.dark": "Gelap",
            "settings.language": "Bahasa",
            "settings.appVersion": "Versi Aplikasi",
            "settings.dangerZone": "Zona Berbahaya (Danger Zone)",
            "settings.deleteAccount": "Hapus Akun Permanen",
            "settings.deleteWarning": "Tindakan ini tidak dapat dibatalkan. Seluruh tugas, catatan, dan data Anda akan dihapus selamanya.",
            "settings.mcpIntegration": "Integrasi AI & MCP",
            "settings.mcpIntegrationDesc": "Model Context Protocol & Personal Access Token untuk AI coding agent",
            "settings.signOut": "Keluar dari Akun",
            "settings.signOutDialogTitle": "Konfirmasi Keluar",
            "settings.signOutConfirmButton": "Keluar dari Akun",
            "settings.signOutDialogMessage": "Anda akan keluar dari sesi ini dan harus login kembali untuk mengakses data.",
            "settings.disconnectDialogTitle": "Putuskan Akun?",
            "settings.disconnectConfirmButton": "Putuskan %@",
            "settings.disconnectDialogMessage": "Apakah Anda yakin ingin memutuskan hubungan akun %@?",
            "settings.linkedAccountsDesc": "Hubungkan akun Anda untuk kemudahan masuk multi-platform.",
            "settings.notConnected": "Belum Terhubung",
            "settings.notificationsDesc": "Atur zona waktu reminder dan toleransi peringatan sebelum tugas jatuh tempo.",
            "settings.threshold15m": "15 Menit sebelum jatuh tempo",
            "settings.threshold1h": "1 Jam sebelum jatuh tempo",
            "settings.threshold3h": "3 Jam sebelum jatuh tempo",
            "settings.threshold12h": "12 Jam sebelum jatuh tempo",
            "settings.threshold24h": "24 Jam (1 hari) sebelum jatuh tempo",
            "settings.threshold48h": "48 Jam (2 hari) sebelum jatuh tempo",
            "settings.deleteConfirmTitle": "Konfirmasi Penghapusan",

            // MCP (Model Context Protocol)
            "mcp.title": "Model Context Protocol (MCP)",
            "mcp.subtitle": "Hubungkan AI coding agent ke ruang kerja Task Hub Anda",
            "mcp.endpoint": "Endpoint Server MCP",
            "mcp.endpointCopied": "Endpoint MCP disalin ke papan klip",
            "mcp.tokensTitle": "Token Akses Personal",
            "mcp.tokensSubtitle": "Buat token khusus untuk AI agent (Claude, Cursor, Antigravity IDE)",
            "mcp.generateToken": "Buat Token Baru",
            "mcp.tokenNamePlaceholder": "Nama token (cth. Cursor Mac, Antigravity IDE)",
            "mcp.tokenCreatedTitle": "Token Berhasil Dibuat",
            "mcp.tokenCreatedWarning": "Pastikan Anda menyalin Personal Access Token Anda sekarang. Anda tidak akan bisa melihatnya lagi!",
            "mcp.copyToken": "Salin Token",
            "mcp.tokenCopied": "Token disalin ke papan klip",
            "mcp.revoke": "Cabut",
            "mcp.revokeConfirmTitle": "Cabut Token?",
            "mcp.revokeConfirmMessage": "Apakah Anda yakin ingin mencabut '%@'? AI agent atau klien yang menggunakan token ini akan langsung kehilangan akses.",
            "mcp.noTokens": "Belum ada token akses yang dibuat.",
            "mcp.lastUsed": "Terakhir digunakan: %@",
            "mcp.neverUsed": "Belum pernah digunakan",
            "mcp.createdOn": "Dibuat: %@",
            "mcp.configSnippets": "Snippet Konfigurasi Cepat",
            "mcp.copyConfig": "Salin Konfigurasi",
            "mcp.configCopied": "Konfigurasi disalin ke papan klip",
            "mcp.antigravityConfig": "Antigravity IDE / VS Code",
            "mcp.cursorConfig": "Cursor / Windsurf",
            "mcp.claudeConfig": "Claude Desktop",

            // General
            "general.create": "Simpan",
            "general.save": "Simpan",
            "general.cancel": "Batal",
            "general.delete": "Hapus",
            "general.close": "Tutup",
            "general.done": "Selesai",
            "general.back": "Kembali"
        ]
    ]
}
