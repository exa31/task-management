import Foundation

/// Type-safe compile-time localized string keys (mirrors web frontend schema)
public enum L10n {
    public enum Nav {
        public static var tasks: String { "nav.tasks".localized }
        public static var projects: String { "nav.projects".localized }
        public static var notes: String { "nav.notes".localized }
        public static var calendar: String { "nav.calendar".localized }
        public static var reminders: String { "nav.reminders".localized }
        public static var settings: String { "nav.settings".localized }
    }

    public enum Calendar {
        public static var title: String { "calendar.title".localized }
        public static var subtitle: String { "calendar.subtitle".localized }
        public static var today: String { "calendar.today".localized }
        public static var all: String { "calendar.all".localized }
        public static var tasks: String { "calendar.tasks".localized }
        public static var reminders: String { "calendar.reminders".localized }
        public static var activities: String { "calendar.activities".localized }
        public static var noItems: String { "calendar.noItems".localized }
        public static var noTasks: String { "calendar.noTasks".localized }
        public static var noReminders: String { "calendar.noReminders".localized }
        public static var noActivities: String { "calendar.noActivities".localized }
        public static var workLog: String { "calendar.workLog".localized }
        public static var tasksDue: String { "calendar.tasksDue".localized }
        public static var scheduledReminders: String { "calendar.scheduledReminders".localized }
        public static var graphqlPowered: String { "calendar.graphqlPowered".localized }
        public static var snooze: String { "calendar.snooze".localized }
        public static var snooze1h: String { "calendar.snooze1h".localized }
        public static var dismiss: String { "calendar.dismiss".localized }
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
        public static var sessionExpired: String { "auth.error.sessionExpired".localized }
        public static var windowNotFound: String { "auth.error.windowNotFound".localized }
        public static var googleTokenFailed: String { "auth.error.googleTokenFailed".localized }
        public static var appleTokenFailed: String { "auth.error.appleTokenFailed".localized }
    }

    public enum Tasks {
        public static var title: String { "tasks.title".localized }
        public static var newTask: String { "tasks.new".localized }
        public static var searchPlaceholder: String { "tasks.searchPlaceholder".localized }
        public static var allProjects: String { "tasks.allProjects".localized }
        public static var allStatuses: String { "tasks.allStatuses".localized }
        public static var allPriorities: String { "tasks.allPriorities".localized }
        public static var focusOnly: String { "tasks.focusOnly".localized }
        public static var resetFilters: String { "tasks.resetFilters".localized }
        public static var resetAllFilters: String { "tasks.resetAllFilters".localized }
        public static var emptyTitle: String { "tasks.emptyTitle".localized }
        public static var emptyDesc: String { "tasks.emptyDesc".localized }
        public static var noTasksFilterTitle: String { "tasks.noTasksFilterTitle".localized }
        public static var noTasksFilterDesc: String { "tasks.noTasksFilterDesc".localized }
        public static var createTask: String { "tasks.createTask".localized }
        public static var saveFailed: String { "tasks.saveFailed".localized }
        public static var noSubtasks: String { "tasks.noSubtasks".localized }
        public static var lastProgressNote: String { "tasks.lastProgressNote".localized }
        public static func subtasksCount(done: Int, total: Int) -> String {
            "tasks.subtasksCount".localized(with: done, total)
        }
        public static var detailTitle: String { "tasks.detailTitle".localized }
        public static var edit: String { "tasks.edit".localized }
        public static var editTitle: String { "tasks.editTitle".localized }
        public static var createTitle: String { "tasks.createTitle".localized }
        public static var sectionInfo: String { "tasks.sectionInfo".localized }
        public static var fieldTitle: String { "tasks.fieldTitle".localized }
        public static var fieldDescription: String { "tasks.fieldDescription".localized }
        public static var sectionStatusPriority: String { "tasks.sectionStatusPriority".localized }
        public static var statusLabel: String { "tasks.statusLabel".localized }
        public static var priorityLabel: String { "tasks.priorityLabel".localized }
        public static var sectionProject: String { "tasks.sectionProject".localized }
        public static var selectProject: String { "tasks.selectProject".localized }
        public static var noProject: String { "tasks.noProject".localized }
        public static var sectionDueDate: String { "tasks.sectionDueDate".localized }
        public static var setDueDate: String { "tasks.setDueDate".localized }
        public static var dateTime: String { "tasks.dateTime".localized }
        public static var sectionFocus: String { "tasks.sectionFocus".localized }
        public static var focusToday: String { "tasks.focusToday".localized }
        public static func activeCount(count: Int) -> String { "tasks.activeCount".localized(with: count) }
        public static var sectionSubtasks: String { "tasks.sectionSubtasks".localized }
        public static var addSubtaskPlaceholder: String { "tasks.addSubtaskPlaceholder".localized }
        public static var sectionTags: String { "tasks.sectionTags".localized }
        public static var tagsPlaceholder: String { "tasks.tagsPlaceholder".localized }
        public static var statusHeader: String { "tasks.statusHeader".localized }
        public static var markDone: String { "tasks.markDone".localized }
        public static var markUndone: String { "tasks.markUndone".localized }
        public static var focusToggle: String { "tasks.focusToggle".localized }
        public static var unfocusToggle: String { "tasks.unfocusToggle".localized }
        public static var tabDetails: String { "tasks.tabDetails".localized }
        public static var tabActivities: String { "tasks.tabActivities".localized }
        public static var postNotePlaceholder: String { "tasks.postNotePlaceholder".localized }
        public static var send: String { "tasks.send".localized }
        public static var loadingLogs: String { "tasks.loadingLogs".localized }
        public static var noLogsTitle: String { "tasks.noLogsTitle".localized }
        public static var noLogsDesc: String { "tasks.noLogsDesc".localized }
        public static var activityCreated: String { "tasks.activityCreated".localized }
        public static var activityUpdated: String { "tasks.activityUpdated".localized }
        public static var activityProgress: String { "tasks.activityProgress".localized }
        public static var activitySubtaskCreated: String { "tasks.activitySubtaskCreated".localized }
        public static var activitySubtaskUpdated: String { "tasks.activitySubtaskUpdated".localized }
        public static var activitySubtaskDeleted: String { "tasks.activitySubtaskDeleted".localized }
        public static var activityFocusChanged: String { "tasks.activityFocusChanged".localized }
        public static var activityDefault: String { "tasks.activityDefault".localized }
        public static var relativeJustNow: String { "tasks.relativeJustNow".localized }
        public static func relativeMinutesAgo(count: Int) -> String { "tasks.relativeMinutesAgo".localized(with: count) }
        public static func relativeHoursAgo(count: Int) -> String { "tasks.relativeHoursAgo".localized(with: count) }
        public static var sectionRecurrence: String { "tasks.sectionRecurrence".localized }
        public static var repeatAuto: String { "tasks.repeatAuto".localized }
        public static var repeatPattern: String { "tasks.repeatPattern".localized }
        public static var patternDaily: String { "tasks.patternDaily".localized }
        public static var patternWeekdays: String { "tasks.patternWeekdays".localized }
        public static var patternWeekly: String { "tasks.patternWeekly".localized }
        public static var patternMonthly: String { "tasks.patternMonthly".localized }
        public static func repeatInterval(interval: Int, unit: String) -> String {
            "tasks.repeatInterval".localized(with: interval, unit)
        }
        public static var unitDays: String { "tasks.unitDays".localized }
        public static var unitWeeks: String { "tasks.unitWeeks".localized }
        public static var unitMonths: String { "tasks.unitMonths".localized }
    }

    public enum Projects {
        public static var title: String { "projects.title".localized }
        public static var newProject: String { "projects.new".localized }
        public static var emptyTitle: String { "projects.emptyTitle".localized }
        public static var emptyDesc: String { "projects.emptyDesc".localized }
        public static var emptyGuidance: String { "projects.emptyGuidance".localized }
        public static var createFirst: String { "projects.createFirst".localized }
        public static var noMatchingProjects: String { "projects.noMatchingProjects".localized }
        public static func noProjectsMatchingDesc(query: String) -> String {
            "projects.noProjectsMatchingDesc".localized(with: query)
        }
        public static func deleteConfirmation(name: String) -> String {
            "projects.deleteConfirmation".localized(with: name)
        }
        public static var deleteWarningNotice: String { "projects.deleteWarningNotice".localized }
        public static var totalProjects: String { "projects.totalProjects".localized }
        public static var sectionInfo: String { "projects.sectionInfo".localized }
        public static var fieldName: String { "projects.fieldName".localized }
        public static var fieldDesc: String { "projects.fieldDesc".localized }
        public static var sectionColor: String { "projects.sectionColor".localized }
        public static var colorHint: String { "projects.colorHint".localized }
        public static var createTitle: String { "projects.createTitle".localized }
        public static var editTitle: String { "projects.editTitle".localized }
        public static var metricTotal: String { "projects.metricTotal".localized }
        public static var metricTotalSub: String { "projects.metricTotalSub".localized }
        public static var metricTasks: String { "projects.metricTasks".localized }
        public static func metricTasksActive(count: Int) -> String { "projects.metricTasksActive".localized(with: count) }
        public static var metricDone: String { "projects.metricDone".localized }
        public static var metricDoneSub: String { "projects.metricDoneSub".localized }
        public static var metricRate: String { "projects.metricRate".localized }
        public static var metricRateSub: String { "projects.metricRateSub".localized }
        public static var noTasksCategory: String { "projects.noTasksCategory".localized }
        public static var addTaskButton: String { "projects.addTaskButton".localized }
    }

    public enum Notes {
        public static var title: String { "notes.title".localized }
        public static var newNote: String { "notes.new".localized }
        public static var emptyTitle: String { "notes.emptyTitle".localized }
        public static var emptyDesc: String { "notes.emptyDesc".localized }
        public static var loading: String { "notes.loading".localized }
        public static var sectionInfo: String { "notes.sectionInfo".localized }
        public static var fieldTitle: String { "notes.fieldTitle".localized }
        public static var fieldType: String { "notes.fieldType".localized }
        public static var typeActionable: String { "notes.typeActionable".localized }
        public static var typeKnowledge: String { "notes.typeKnowledge".localized }
        public static var fieldContent: String { "notes.fieldContent".localized }
        public static var pinToTop: String { "notes.pinToTop".localized }
        public static var searchPlaceholder: String { "notes.searchPlaceholder".localized }
        public static var filterAll: String { "notes.filterAll".localized }
        public static var pin: String { "notes.pin".localized }
        public static var unpin: String { "notes.unpin".localized }
        public static var editNote: String { "notes.edit".localized }
        public static var deleteNote: String { "notes.delete".localized }
        public static var deleteConfirmTitle: String { "notes.deleteConfirmTitle".localized }
        public static var deleteConfirmMessage: String { "notes.deleteConfirmMessage".localized }
        public static var convertToTask: String { "notes.convertToTask".localized }
        public static var convertConfirmTitle: String { "notes.convertConfirmTitle".localized }
        public static var convertConfirmMessage: String { "notes.convertConfirmMessage".localized }
        public static var convertedSuccess: String { "notes.convertedSuccess".localized }
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
        public static var profileUpdated: String { "settings.profileUpdated".localized }
        public static var avatarUpdated: String { "settings.avatarUpdated".localized }
        public static var preferencesSaved: String { "settings.preferencesSaved".localized }
        public static func accountDisconnected(provider: String) -> String {
            "settings.accountDisconnected".localized(with: provider)
        }
        public static var appleConnected: String { "settings.appleConnected".localized }
        public static var deleteConfirmPrompt: String { "settings.deleteConfirmPrompt".localized }
        public static var deletePermanentNotice: String { "settings.deletePermanentNotice".localized }
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
        public static var mcpIntegration: String { "settings.mcpIntegration".localized }
        public static var mcpIntegrationDesc: String { "settings.mcpIntegrationDesc".localized }
        public static var signOut: String { "settings.signOut".localized }
        public static var signOutDialogTitle: String { "settings.signOutDialogTitle".localized }
        public static var signOutConfirmButton: String { "settings.signOutConfirmButton".localized }
        public static var signOutDialogMessage: String { "settings.signOutDialogMessage".localized }
        public static var disconnectDialogTitle: String { "settings.disconnectDialogTitle".localized }
        public static func disconnectConfirmButton(provider: String) -> String { "settings.disconnectConfirmButton".localized(with: provider) }
        public static func disconnectDialogMessage(provider: String) -> String { "settings.disconnectDialogMessage".localized(with: provider) }
        public static var linkedAccountsDesc: String { "settings.linkedAccountsDesc".localized }
        public static var notConnected: String { "settings.notConnected".localized }
        public static var notificationsDesc: String { "settings.notificationsDesc".localized }
        public static var threshold15m: String { "settings.threshold15m".localized }
        public static var threshold1h: String { "settings.threshold1h".localized }
        public static var threshold3h: String { "settings.threshold3h".localized }
        public static var threshold12h: String { "settings.threshold12h".localized }
        public static var threshold24h: String { "settings.threshold24h".localized }
        public static var threshold48h: String { "settings.threshold48h".localized }
        public static var deleteConfirmTitle: String { "settings.deleteConfirmTitle".localized }
    }

    public enum Mcp {
        public static var title: String { "mcp.title".localized }
        public static var subtitle: String { "mcp.subtitle".localized }
        public static var endpoint: String { "mcp.endpoint".localized }
        public static var endpointCopied: String { "mcp.endpointCopied".localized }
        public static var tokensTitle: String { "mcp.tokensTitle".localized }
        public static var tokensSubtitle: String { "mcp.tokensSubtitle".localized }
        public static var generateToken: String { "mcp.generateToken".localized }
        public static var tokenNamePlaceholder: String { "mcp.tokenNamePlaceholder".localized }
        public static var tokenCreatedTitle: String { "mcp.tokenCreatedTitle".localized }
        public static var tokenCreatedWarning: String { "mcp.tokenCreatedWarning".localized }
        public static var copyToken: String { "mcp.copyToken".localized }
        public static var tokenCopied: String { "mcp.tokenCopied".localized }
        public static var revoke: String { "mcp.revoke".localized }
        public static var revokeConfirmTitle: String { "mcp.revokeConfirmTitle".localized }
        public static func revokeConfirmMessage(name: String) -> String {
            "mcp.revokeConfirmMessage".localized(with: name)
        }
        public static var noTokens: String { "mcp.noTokens".localized }
        public static func lastUsed(date: String) -> String {
            "mcp.lastUsed".localized(with: date)
        }
        public static var neverUsed: String { "mcp.neverUsed".localized }
        public static func createdOn(date: String) -> String {
            "mcp.createdOn".localized(with: date)
        }
        public static var configSnippets: String { "mcp.configSnippets".localized }
        public static var copyConfig: String { "mcp.copyConfig".localized }
        public static var configCopied: String { "mcp.configCopied".localized }
        public static var antigravityConfig: String { "mcp.antigravityConfig".localized }
        public static var cursorConfig: String { "mcp.cursorConfig".localized }
        public static var claudeConfig: String { "mcp.claudeConfig".localized }
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

    public enum Widgets {
        public static var allSubtasksDone: String { "widget.checklist.allSubtasksDone".localized }
        public static var focusToday: String { "widget.checklist.focusToday".localized }
        public static var noActiveSubtasks: String { "widget.checklist.noActiveSubtasks".localized }
        public static var interactiveTitle: String { "widget.checklist.interactiveTitle".localized }
        public static func completedCount(done: Int, total: Int) -> String {
            "widget.checklist.completedCount".localized(with: done, total)
        }
        public static var allDone: String { "widget.checklist.allDone".localized }

        public static var todaySummary: String { "widget.overview.todaySummary".localized }
        public static var doneBadge: String { "widget.overview.done".localized }
        public static var productivitySummary: String { "widget.overview.productivitySummary".localized }
        public static func tasksDone(completed: Int, total: Int) -> String {
            "widget.overview.tasksDone".localized(with: completed, total)
        }
        public static func tasksInFocus(count: Int) -> String {
            "widget.overview.tasksInFocus".localized(with: count)
        }
        public static var quoteAllDone: String { "widget.overview.quoteAllDone".localized }
        public static var quoteAlmost: String { "widget.overview.quoteAlmost".localized }
        public static var quoteSteady: String { "widget.overview.quoteSteady".localized }
        public static var quoteStart: String { "widget.overview.quoteStart".localized }

        public static var deadlineBadge: String { "widget.deadlines.badge".localized }
        public static var upcomingHeader: String { "widget.deadlines.approachingHeader".localized }
        public static func tasksCount(count: Int) -> String { "widget.deadlines.tasksCount".localized(with: count) }
        public static func activeTasksCount(count: Int) -> String { "widget.deadlines.activeTasksCount".localized(with: count) }
        public static var listTitle: String { "widget.deadlines.listTitle".localized }
        public static func todayAt(time: String) -> String { "widget.deadlines.todayAt".localized(with: time) }
        public static var tomorrow: String { "widget.deadlines.tomorrow".localized }
        public static var overdue: String { "widget.deadlines.overdue".localized }
        public static var noUrgent: String { "widget.deadlines.noUrgent".localized }

        public static func lockscreenTasksDone(completed: Int, total: Int) -> String {
            "widget.lockscreen.tasksDone".localized(with: completed, total)
        }
        public static func duePrefix(time: String) -> String {
            "widget.lockscreen.duePrefix".localized(with: time)
        }
        public static var lockscreenAllDone: String { "widget.lockscreen.allDone".localized }
    }
}
