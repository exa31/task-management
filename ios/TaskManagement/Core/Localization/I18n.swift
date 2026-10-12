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
            "nav.reminders": "Reminders",
            "nav.settings": "Settings",

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
            "nav.reminders": "Pengingat",
            "nav.settings": "Pengaturan",

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
