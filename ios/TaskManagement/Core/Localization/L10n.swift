import Foundation

/// Type-safe compile-time localized string keys (mirrors web frontend schema)
public enum L10n {
    public enum Nav {
        public static var tasks: String { "nav.tasks".localized }
        public static var projects: String { "nav.projects".localized }
        public static var notes: String { "nav.notes".localized }
        public static var reminders: String { "nav.reminders".localized }
        public static var settings: String { "nav.settings".localized }
    }

    public enum Auth {
        public static var title: String { "auth.title".localized }
        public static var subtitle: String { "auth.subtitle".localized }
        public static var syncTitle: String { "auth.sync.title".localized }
        public static var syncDesc: String { "auth.sync.desc".localized }
        public static var projectsTitle: String { "auth.projects.title".localized }
        public static var projectsDesc: String { "auth.projects.desc".localized }
        public static var remindersTitle: String { "auth.reminders.title".localized }
        public static var remindersDesc: String { "auth.reminders.desc".localized }
        public static var signInApple: String { "auth.signInApple".localized }
        public static var signInGoogle: String { "auth.signInGoogle".localized }
        public static var secure: String { "auth.secure".localized }
        public static var signOut: String { "auth.signOut".localized }
    }

    public enum Reminders {
        public static var title: String { "reminders.title".localized }
        public static var newReminder: String { "reminders.new".localized }
        public static var emptyTitle: String { "reminders.emptyTitle".localized }
        public static var emptyDesc: String { "reminders.emptyDesc".localized }
        public static var createFirst: String { "reminders.createFirst".localized }
        public static var add: String { "reminders.add".localized }
        public static var filterAll: String { "reminders.filter.all".localized }
        public static var filterPending: String { "reminders.filter.pending".localized }
        public static var filterSent: String { "reminders.filter.sent".localized }
        public static var filterDismissed: String { "reminders.filter.dismissed".localized }
        public static var fieldTitle: String { "reminders.field.title".localized }
        public static var fieldNotes: String { "reminders.field.notes".localized }
        public static var sectionInfo: String { "reminders.section.info".localized }
        public static var sectionSchedule: String { "reminders.section.schedule".localized }
        public static var sectionAssociation: String { "reminders.section.association".localized }
        public static var remindAt: String { "reminders.remindAt".localized }
        public static var presets: String { "reminders.presets".localized }
        public static var tomorrowMorning: String { "reminders.tomorrowMorning".localized }
        public static var taskNone: String { "reminders.taskNone".localized }
        public static var snooze: String { "reminders.snooze".localized }
        public static var done: String { "reminders.done".localized }
        public static func snoozed(count: Int) -> String {
            "reminders.snoozedCount".localized(with: count)
        }
        public static var selectTask: String { "reminders.selectTask".localized }
        public static var linkedTask: String { "reminders.linkedTask".localized }
        public static var searchTasksPlaceholder: String { "reminders.searchTasksPlaceholder".localized }
        public static var noTasksFound: String { "reminders.noTasksFound".localized }
        public static var unlinkTask: String { "reminders.unlinkTask".localized }
        public static var completedNotice: String { "reminders.completedNotice".localized }
        public static var liveAlert: String { "reminders.liveAlert".localized }
    }

    public enum Settings {
        public static var title: String { "settings.title".localized }
        public static var profileInfo: String { "settings.profileInfo".localized }
        public static var changeAvatar: String { "settings.changeAvatar".localized }
        public static var fullName: String { "settings.fullName".localized }
        public static var emailAddress: String { "settings.emailAddress".localized }
        public static var saveProfile: String { "settings.saveProfile".localized }
        public static var linkedAccounts: String { "settings.linkedAccounts".localized }
        public static var connected: String { "settings.connected".localized }
        public static var disconnect: String { "settings.disconnect".localized }
        public static var connect: String { "settings.connect".localized }
        public static var notifications: String { "settings.notifications".localized }
        public static var timezone: String { "settings.timezone".localized }
        public static var threshold: String { "settings.threshold".localized }
        public static var thresholdHint: String { "settings.thresholdHint".localized }
        public static var emailNotifications: String { "settings.emailNotifications".localized }
        public static var emailNotificationsDesc: String { "settings.emailNotificationsDesc".localized }
        public static var savePreferences: String { "settings.savePreferences".localized }
        public static var appearance: String { "settings.appearance".localized }
        public static var themeMode: String { "settings.themeMode".localized }
        public static var themeSystem: String { "settings.theme.system".localized }
        public static var themeLight: String { "settings.theme.light".localized }
        public static var themeDark: String { "settings.theme.dark".localized }
        public static var language: String { "settings.language".localized }
        public static var appVersion: String { "settings.appVersion".localized }
        public static var dangerZone: String { "settings.dangerZone".localized }
        public static var deleteAccount: String { "settings.deleteAccount".localized }
        public static var deleteWarning: String { "settings.deleteWarning".localized }
    }

    public enum General {
        public static var create: String { "general.create".localized }
        public static var save: String { "general.save".localized }
        public static var cancel: String { "general.cancel".localized }
        public static var delete: String { "general.delete".localized }
        public static var close: String { "general.close".localized }
        public static var done: String { "general.done".localized }
        public static var back: String { "general.back".localized }
    }
}
