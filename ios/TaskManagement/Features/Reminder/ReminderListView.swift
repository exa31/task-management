import SwiftUI
import UserNotifications

@Observable
@MainActor
public final class ReminderViewModel {
    public var reminders: [Reminder] = []
    public var activeAlert: String?
    public var isLoading = false
    public var isCreating = false
    public var isUpdating = false
    public var isShowingCreateSheet = false
    public var errorMessage: String?

    private let apiClient = APIClient.shared
    private let sseClient = SSEClient.shared

    public init() {
        Task {
            await loadReminders()
            await startSSEListener()
        }
    }

    @MainActor
    public func loadReminders() async {
        isLoading = true
        do {
            self.reminders = try await apiClient.getReminders()
            AppLogger.info("Reminders loaded successfully", category: .reminder, metadata: ["count": reminders.count])
            syncAllLocalAlarms()
        } catch {
            AppLogger.error("Failed to load reminders", error: error, category: .reminder)
            self.errorMessage = error.localizedDescription
        }
        self.isLoading = false
    }

    @MainActor
    public func searchTasks(query: String = "") async -> [TaskItem] {
        do {
            let trimmed = query.trimmingCharacters(in: .whitespacesAndNewlines)
            return try await apiClient.getTasks(
                search: trimmed.isEmpty ? nil : trimmed,
                limit: 15
            )
        } catch {
            return []
        }
    }

    @MainActor
    public func startSSEListener() async {
        for await event in sseClient.events() {
            if event.type == "reminder:triggered" {
                self.activeAlert = "\(L10n.Reminders.liveAlert): \(event.data)"
                await loadReminders()
            }
        }
    }

    @MainActor
    public func createReminder(title: String, description: String?, remindAt: Date, taskId: Int64?) async -> Bool {
        let trimmedTitle = title.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmedTitle.isEmpty else {
            self.errorMessage = "\(L10n.Reminders.fieldTitle) required"
            return false
        }
        isCreating = true
        errorMessage = nil

        let remindAtStr = DateTimeHelper.toISOString(remindAt)

        do {
            let created = try await apiClient.createReminder(
                title: trimmedTitle,
                description: description?.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty == true ? nil : description,
                remindAt: remindAtStr,
                taskId: taskId
            )
            AppLogger.success("Reminder created successfully", category: .reminder, metadata: ["title": trimmedTitle])
            await loadReminders()
            scheduleLocalAlarm(for: created)
            isCreating = false
            return true
        } catch {
            AppLogger.error("Failed to create reminder", error: error, category: .reminder)
            self.errorMessage = error.localizedDescription
            isCreating = false
            return false
        }
    }

    @MainActor
    public func updateReminder(
        id: Int64,
        title: String,
        description: String?,
        remindAt: Date,
        status: String?,
        taskId: Int64?,
        clearTaskId: Bool
    ) async -> Bool {
        let trimmedTitle = title.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmedTitle.isEmpty else {
            self.errorMessage = "\(L10n.Reminders.fieldTitle) required"
            return false
        }
        isUpdating = true
        errorMessage = nil

        let remindAtStr = DateTimeHelper.toISOString(remindAt)

        let req = UpdateReminderRequest(
            title: trimmedTitle,
            description: description?.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty == true ? nil : description,
            remindAt: remindAtStr,
            status: status,
            taskId: taskId,
            clearTaskId: clearTaskId
        )

        do {
            let updated = try await apiClient.updateReminder(id: id, req: req)
            AppLogger.success("Reminder updated successfully", category: .reminder, metadata: ["id": id, "title": trimmedTitle])
            if let idx = reminders.firstIndex(where: { $0.id == id }) {
                reminders[idx] = updated
            } else {
                await loadReminders()
            }
            scheduleLocalAlarm(for: updated)
            isUpdating = false
            return true
        } catch {
            AppLogger.error("Failed to update reminder", error: error, category: .reminder, metadata: ["id": id])
            self.errorMessage = error.localizedDescription
            isUpdating = false
            return false
        }
    }

    @MainActor
    public func toggleAlarmStatus(reminder: Reminder, isEnabled: Bool) {
        let newStatus = isEnabled ? "pending" : "dismissed"
        let req = UpdateReminderRequest(
            title: reminder.title,
            description: reminder.description.isEmpty ? nil : reminder.description,
            remindAt: reminder.remindAt,
            status: newStatus,
            taskId: reminder.taskId,
            clearTaskId: false
        )
        Task {
            do {
                let updated = try await apiClient.updateReminder(id: reminder.id, req: req)
                if let idx = reminders.firstIndex(where: { $0.id == reminder.id }) {
                    reminders[idx] = updated
                }
                scheduleLocalAlarm(for: updated)
                #if canImport(UIKit)
                UIImpactFeedbackGenerator(style: .medium).impactOccurred()
                #endif
            } catch {
                AppLogger.error("Failed to toggle alarm status", error: error, category: .reminder)
                self.errorMessage = error.localizedDescription
            }
        }
    }

    @MainActor
    public func snooze(id: Int64) {
        Task {
            do {
                let res = try await apiClient.snoozeReminder(id: id, minutes: 15)
                AppLogger.success("Reminder snoozed for 15 minutes", category: .reminder, metadata: ["id": id])
                await loadReminders()
                scheduleLocalAlarm(for: res)
            } catch {
                AppLogger.error("Failed to snooze reminder", error: error, category: .reminder, metadata: ["id": id])
                self.errorMessage = error.localizedDescription
            }
        }
    }

    @MainActor
    public func dismiss(id: Int64) {
        Task {
            do {
                let res = try await apiClient.dismissReminder(id: id)
                AppLogger.success("Reminder dismissed", category: .reminder, metadata: ["id": id])
                await loadReminders()
                scheduleLocalAlarm(for: res)
            } catch {
                AppLogger.error("Failed to dismiss reminder", error: error, category: .reminder, metadata: ["id": id])
                self.errorMessage = error.localizedDescription
            }
        }
    }

    @MainActor
    public func deleteReminder(id: Int64) {
        Task {
            do {
                try await apiClient.deleteReminder(id: id)
                AppLogger.success("Reminder deleted", category: .reminder, metadata: ["id": id])
                UNUserNotificationCenter.current().removePendingNotificationRequests(withIdentifiers: ["alarm_reminder_\(id)"])
                await loadReminders()
            } catch {
                AppLogger.error("Failed to delete reminder", error: error, category: .reminder, metadata: ["id": id])
                self.errorMessage = error.localizedDescription
            }
        }
    }

    public func scheduleLocalAlarm(for reminder: Reminder) {
        let center = UNUserNotificationCenter.current()
        let identifier = "alarm_reminder_\(reminder.id)"

        if reminder.status.lowercased() != "pending" {
            center.removePendingNotificationRequests(withIdentifiers: [identifier])
            return
        }

        guard let targetDate = DateTimeHelper.parseISO8601(reminder.remindAt),
              targetDate > Date() else {
            center.removePendingNotificationRequests(withIdentifiers: [identifier])
            return
        }

        var calendar = Calendar.current
        calendar.timeZone = DateTimeHelper.userTimeZone
        let comps = calendar.dateComponents([.year, .month, .day, .hour, .minute], from: targetDate)
        let trigger = UNCalendarNotificationTrigger(dateMatching: comps, repeats: false)

        let content = UNMutableNotificationContent()
        content.title = reminder.title
        content.body = reminder.description.isEmpty ? "Waktunya alarm & pengingat tugas Anda." : reminder.description
        content.categoryIdentifier = "ALARM_REMINDER"
        content.interruptionLevel = .timeSensitive

        let soundName = SoundManager.shared.selectedSound.fileName
        if soundName != "default" {
            content.sound = UNNotificationSound(named: UNNotificationSoundName(soundName))
        } else {
            content.sound = .default
        }

        content.userInfo = [
            "reminder_id": NSNumber(value: reminder.id),
            "title": reminder.title,
            "task_title": reminder.taskTitle ?? "",
            "body": reminder.description
        ]

        let request = UNNotificationRequest(identifier: identifier, content: content, trigger: trigger)
        center.add(request) { error in
            if let error = error {
                AppLogger.warn("Failed to schedule local alarm request: \(error.localizedDescription)", category: .push)
            } else {
                AppLogger.info("Scheduled local alarm request for reminder #\(reminder.id)", category: .push)
            }
        }
    }

    public func syncAllLocalAlarms() {
        for r in reminders {
            scheduleLocalAlarm(for: r)
        }
    }

    public func registerDeviceToken(_ token: String) {
        Task {
            try? await apiClient.registerDeviceToken(token: token)
        }
    }
}

public struct ReminderListView: View {
    @Bindable var viewModel: ReminderViewModel
    @State private var filterStatus: String = "all" // all, pending, sent, dismissed
    @State private var selectedReminderForEdit: Reminder? = nil

    public init(viewModel: ReminderViewModel) {
        self.viewModel = viewModel
    }

    private var filteredReminders: [Reminder] {
        if filterStatus == "all" {
            return viewModel.reminders
        }
        return viewModel.reminders.filter { $0.status.lowercased() == filterStatus }
    }

    private var nextUpcomingAlarm: Reminder? {
        let now = Date()
        return viewModel.reminders
            .filter { $0.status.lowercased() == "pending" }
            .compactMap { rem -> (Reminder, Date)? in
                guard let d = DateTimeHelper.parseISO8601(rem.remindAt), d > now else { return nil }
                return (rem, d)
            }
            .min(by: { $0.1 < $1.1 })?.0
    }

    public var body: some View {
        AdaptiveTabNavigationStack {
            VStack(spacing: 0) {
                // Next Upcoming Alarm Banner
                if let nextAlarm = nextUpcomingAlarm {
                    HStack(spacing: 12) {
                        ZStack {
                            Circle()
                                .fill(Color.orange.opacity(0.18))
                                .frame(width: 38, height: 38)
                            Image(systemName: "alarm.waves.left.and.right.fill")
                                .font(.system(size: 16, weight: .bold))
                                .foregroundColor(.orange)
                        }

                        VStack(alignment: .leading, spacing: 2) {
                            Text(L10n.Alarm.nearest)
                                .font(.system(size: 11, weight: .bold))
                                .foregroundColor(.orange)
                            Text("\(nextAlarm.title) • \(DateTimeHelper.formatTime(nextAlarm.remindAt)) (\(DateTimeHelper.formatRelativeAlarmDate(nextAlarm.remindAt)))")
                                .font(.system(size: 13, weight: .semibold))
                                .foregroundColor(.primary)
                                .lineLimit(1)
                        }

                        Spacer()
                    }
                    .padding(12)
                    .background(Color.orange.opacity(0.08))
                    .overlay(
                        RoundedRectangle(cornerRadius: 14)
                            .stroke(Color.orange.opacity(0.25), lineWidth: 1)
                    )
                    .clipShape(RoundedRectangle(cornerRadius: 14))
                    .padding(.horizontal, 16)
                    .padding(.top, 10)
                }

                // Live Alert Banner (SSE Triggered)
                if let alert = viewModel.activeAlert {
                    HStack(spacing: 12) {
                        ZStack {
                            Circle()
                                .fill(Color.orange.opacity(0.2))
                                .frame(width: 36, height: 36)
                            Image(systemName: "bell.badge.fill")
                                .font(.system(size: 16))
                                .foregroundColor(.orange)
                        }

                        VStack(alignment: .leading, spacing: 2) {
                            Text(L10n.Reminders.liveAlert)
                                .font(.system(size: 12, weight: .bold))
                                .foregroundColor(.orange)
                            Text(alert)
                                .font(.system(size: 13, weight: .medium))
                                .lineLimit(2)
                        }

                        Spacer()

                        Button(action: { viewModel.activeAlert = nil }) {
                            Image(systemName: "xmark.circle.fill")
                                .font(.system(size: 18))
                                .foregroundColor(.secondary)
                        }
                        .buttonStyle(.plain)
                    }
                    .padding(14)
                    .background(Color.orange.opacity(0.1))
                    .overlay(
                        RoundedRectangle(cornerRadius: 14)
                            .stroke(Color.orange.opacity(0.3), lineWidth: 1)
                    )
                    .clipShape(RoundedRectangle(cornerRadius: 14))
                    .padding(.horizontal, 16)
                    .padding(.top, 10)
                }

                // Filter Tabs Bar
                ScrollView(.horizontal, showsIndicators: false) {
                    HStack(spacing: 8) {
                        filterChip(title: "\(L10n.Reminders.filterAll) (\(viewModel.reminders.count))", value: "all")
                        filterChip(title: L10n.Reminders.filterPending, value: "pending")
                        filterChip(title: L10n.Reminders.filterSent, value: "sent")
                        filterChip(title: L10n.Reminders.filterDismissed, value: "dismissed")
                    }
                    .padding(.horizontal, 16)
                    .padding(.vertical, 10)
                }

                if viewModel.isLoading && viewModel.reminders.isEmpty {
                    TaskListSkeleton(count: 4)
                        .padding(.horizontal, 16)
                } else if filteredReminders.isEmpty {
                    Spacer()
                    VStack(spacing: 14) {
                        ZStack {
                            Circle()
                                .fill(AppColors.primary.opacity(0.12))
                                .frame(width: 72, height: 72)
                            Image(systemName: "bell.slash.fill")
                                .font(.system(size: 32))
                                .foregroundColor(AppColors.primary)
                        }
                        Text(L10n.Reminders.emptyTitle)
                            .font(.system(size: 17, weight: .bold))
                        Text(L10n.Reminders.emptyDesc)
                            .font(.system(size: 13))
                            .foregroundColor(.secondary)
                            .multilineTextAlignment(.center)
                            .padding(.horizontal, 32)

                        Button(action: { viewModel.isShowingCreateSheet = true }) {
                            HStack(spacing: 6) {
                                Image(systemName: "plus")
                                Text(L10n.Reminders.createFirst)
                            }
                            .font(.system(size: 14, weight: .semibold))
                            .padding(.horizontal, 18)
                            .padding(.vertical, 10)
                            .background(AppColors.brandGradient)
                            .foregroundColor(.white)
                            .clipShape(Capsule())
                        }
                        .padding(.top, 8)
                    }
                    Spacer()
                } else {
                    List {
                        ForEach(filteredReminders) { reminder in
                            reminderCard(reminder)
                                .contentShape(Rectangle())
                                .onTapGesture {
                                    selectedReminderForEdit = reminder
                                }
                                .swipeActions(edge: .trailing) {
                                    Button(role: .destructive) {
                                        viewModel.deleteReminder(id: reminder.id)
                                    } label: {
                                        Label(L10n.General.delete, systemImage: "trash")
                                    }

                                    Button {
                                        selectedReminderForEdit = reminder
                                    } label: {
                                        Label(L10n.General.edit, systemImage: "pencil")
                                    }
                                    .tint(AppColors.primary)
                                }
                                .swipeActions(edge: .leading) {
                                    if reminder.status != "dismissed" {
                                        Button {
                                            viewModel.dismiss(id: reminder.id)
                                        } label: {
                                            Label(L10n.Reminders.done, systemImage: "checkmark")
                                        }
                                        .tint(AppColors.statusDone)
                                    }
                                }
                                .contextMenu {
                                    Button {
                                        selectedReminderForEdit = reminder
                                    } label: {
                                        Label(L10n.Reminders.editReminder, systemImage: "pencil")
                                    }

                                    if reminder.status != "dismissed" {
                                        Button {
                                            viewModel.snooze(id: reminder.id)
                                        } label: {
                                            Label(L10n.Reminders.snooze, systemImage: "clock.arrow.circlepath")
                                        }

                                        Button {
                                            viewModel.dismiss(id: reminder.id)
                                        } label: {
                                            Label(L10n.Reminders.done, systemImage: "checkmark")
                                        }
                                    }

                                    Divider()

                                    Button(role: .destructive) {
                                        viewModel.deleteReminder(id: reminder.id)
                                    } label: {
                                        Label(L10n.General.delete, systemImage: "trash")
                                    }
                                }
                                .listRowSeparator(.hidden)
                                .listRowBackground(Color.clear)
                                .listRowInsets(EdgeInsets(top: 6, leading: 16, bottom: 6, trailing: 16))
                        }
                    }
                    .listStyle(.plain)
                    .refreshable {
                        await viewModel.loadReminders()
                    }
                }
            }
            .navigationTitle(L10n.Reminders.title)
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Button(action: { viewModel.isShowingCreateSheet = true }) {
                        HStack(spacing: 4) {
                            Image(systemName: "plus")
                                .font(.system(size: 12, weight: .bold))
                            Text(L10n.Reminders.add)
                                .font(.system(size: 13, weight: .semibold))
                        }
                        .padding(.horizontal, 10)
                        .padding(.vertical, 5)
                        .background(AppColors.primary.opacity(0.12))
                        .foregroundColor(AppColors.primary)
                        .clipShape(Capsule())
                    }
                }
            }
            .sheet(isPresented: $viewModel.isShowingCreateSheet) {
                CreateReminderSheet(viewModel: viewModel)
            }
            .sheet(item: $selectedReminderForEdit) { reminder in
                EditReminderSheet(reminder: reminder, viewModel: viewModel)
            }
        }
    }

    private func filterChip(title: String, value: String) -> some View {
        Button(action: { filterStatus = value }) {
            Text(title)
                .font(.system(size: 12, weight: filterStatus == value ? .bold : .medium))
                .padding(.horizontal, 12)
                .padding(.vertical, 6)
                .background(filterStatus == value ? AppColors.primary : Color(.secondarySystemGroupedBackground))
                .foregroundColor(filterStatus == value ? .white : .secondary)
                .clipShape(Capsule())
                .overlay(
                    Capsule()
                        .stroke(Color.primary.opacity(filterStatus == value ? 0 : 0.08), lineWidth: 1)
                )
        }
        .buttonStyle(.plain)
    }

    private func reminderCard(_ reminder: Reminder) -> some View {
        let isDismissed = reminder.status.lowercased() == "dismissed"

        return VStack(alignment: .leading, spacing: 10) {
            HStack(alignment: .center) {
                VStack(alignment: .leading, spacing: 4) {
                    // Big digital alarm time & Date badge
                    HStack(alignment: .firstTextBaseline, spacing: 8) {
                        Text(DateTimeHelper.formatTime(reminder.remindAt))
                            .font(.system(size: 38, weight: .bold, design: .rounded))
                            .foregroundColor(isDismissed ? .secondary.opacity(0.4) : .primary)

                        Text(DateTimeHelper.formatRelativeAlarmDate(reminder.remindAt))
                            .font(.system(size: 13, weight: .semibold))
                            .foregroundColor(isDismissed ? .secondary.opacity(0.4) : AppColors.primary)
                    }

                    // Alarm Title & Label
                    HStack(spacing: 6) {
                        Text(reminder.title)
                            .font(.system(size: 15, weight: .medium))
                            .foregroundColor(isDismissed ? .secondary.opacity(0.6) : .primary)
                            .lineLimit(1)

                        if let snooze = reminder.snoozeCount, snooze > 0 {
                            Text(L10n.Reminders.snoozed(count: snooze))
                                .font(.system(size: 10, weight: .bold))
                                .padding(.horizontal, 6)
                                .padding(.vertical, 2)
                                .background(Color.orange.opacity(0.15))
                                .foregroundColor(.orange)
                                .clipShape(Capsule())
                        }
                    }

                    if !reminder.description.isEmpty {
                        Text(reminder.description)
                            .font(.system(size: 12))
                            .foregroundColor(.secondary.opacity(0.8))
                            .lineLimit(2)
                    }

                    // Tags: Linked Task & Sound Info
                    HStack(spacing: 8) {
                        if let taskTitle = reminder.taskTitle {
                            HStack(spacing: 4) {
                                Image(systemName: "link")
                                    .font(.system(size: 9))
                                Text(taskTitle)
                                    .font(.system(size: 11, weight: .medium))
                            }
                            .foregroundColor(AppColors.violet)
                        }

                        HStack(spacing: 4) {
                            Image(systemName: "bell.fill")
                                .font(.system(size: 9))
                            Text(SoundManager.shared.selectedSound.name)
                                .font(.system(size: 11))
                        }
                        .foregroundColor(.secondary)
                    }
                }

                Spacer()

                // On/Off Toggle Switch (like Apple Clock app)
                Toggle("", isOn: Binding(
                    get: { reminder.status.lowercased() != "dismissed" },
                    set: { isEnabled in
                        viewModel.toggleAlarmStatus(reminder: reminder, isEnabled: isEnabled)
                    }
                ))
                .labelsHidden()
                .tint(AppColors.primary)
            }

            Divider()
                .padding(.vertical, 2)

            // Bottom actions: Snooze & Dismiss / Edit
            HStack(spacing: 12) {
                if !isDismissed {
                    Button(action: { viewModel.snooze(id: reminder.id) }) {
                        HStack(spacing: 6) {
                            Image(systemName: "clock.arrow.circlepath")
                            Text(L10n.Reminders.snooze)
                        }
                        .font(.system(size: 12, weight: .semibold))
                        .padding(.horizontal, 12)
                        .padding(.vertical, 6)
                        .background(Color(.secondarySystemGroupedBackground))
                        .overlay(
                            Capsule().stroke(Color.secondary.opacity(0.2), lineWidth: 1)
                        )
                        .clipShape(Capsule())
                    }
                    .buttonStyle(.plain)

                    Button(action: { viewModel.dismiss(id: reminder.id) }) {
                        HStack(spacing: 6) {
                            Image(systemName: "checkmark")
                            Text(L10n.Reminders.done)
                        }
                        .font(.system(size: 12, weight: .semibold))
                        .padding(.horizontal, 14)
                        .padding(.vertical, 6)
                        .background(AppColors.statusDone)
                        .foregroundColor(.white)
                        .clipShape(Capsule())
                    }
                    .buttonStyle(.plain)
                } else {
                    HStack(spacing: 4) {
                        Image(systemName: "moon.fill")
                            .font(.system(size: 11))
                            .foregroundColor(.secondary)
                        Text(L10n.Alarm.disabled)
                            .font(.system(size: 12, weight: .medium))
                            .foregroundColor(.secondary)
                    }
                }

                Spacer()

                Button(action: { selectedReminderForEdit = reminder }) {
                    HStack(spacing: 4) {
                        Image(systemName: "pencil")
                        Text(L10n.General.edit)
                    }
                    .font(.system(size: 11, weight: .semibold))
                    .foregroundColor(AppColors.primary)
                    .padding(.horizontal, 8)
                    .padding(.vertical, 5)
                    .background(AppColors.primary.opacity(0.08))
                    .clipShape(Capsule())
                }
                .buttonStyle(.plain)
            }
        }
        .padding(14)
        .background(Color(.secondarySystemGroupedBackground))
        .clipShape(RoundedRectangle(cornerRadius: 18, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: 18, style: .continuous)
                .stroke(isDismissed ? Color.clear : AppColors.primary.opacity(0.12), lineWidth: 1)
        )
    }

    private func statusBadge(_ status: String) -> some View {
        let text: String
        let color: Color
        switch status.lowercased() {
        case "pending":
            text = L10n.Reminders.filterPending
            color = .orange
        case "sent":
            text = L10n.Reminders.filterSent
            color = AppColors.violet
        case "dismissed":
            text = L10n.Reminders.done
            color = AppColors.statusDone
        default:
            text = status.capitalized
            color = .secondary
        }

        return Text(text)
            .font(.system(size: 9, weight: .bold))
            .padding(.horizontal, 6)
            .padding(.vertical, 2)
            .background(color.opacity(0.12))
            .foregroundColor(color)
            .clipShape(Capsule())
    }
}

// MARK: - Sound Option Row Helper View
private struct SoundOptionRowView: View {
    let sound: SoundOption
    let isSelected: Bool
    let onSelect: () -> Void

    var body: some View {
        HStack {
            VStack(alignment: .leading, spacing: 2) {
                Text(sound.name)
                    .font(.system(size: 14, weight: .medium))
                Text(sound.fileName)
                    .font(.system(size: 11))
                    .foregroundColor(.secondary)
            }
            Spacer()
            if isSelected {
                Image(systemName: "checkmark")
                    .foregroundColor(AppColors.primary)
                    .font(.system(size: 14, weight: .bold))
            }
            Button(action: {
                SoundManager.shared.playPreview(sound: sound)
            }) {
                Image(systemName: "play.circle.fill")
                    .font(.system(size: 22))
                    .foregroundColor(AppColors.primary)
            }
            .buttonStyle(.plain)
        }
        .contentShape(Rectangle())
        .onTapGesture(perform: onSelect)
    }
}

// MARK: - Create Reminder Sheet Modal
public struct CreateReminderSheet: View {
    @Bindable var viewModel: ReminderViewModel
    @Environment(\.dismiss) private var dismiss

    @State private var title: String = ""
    @State private var description: String = ""
    @State private var remindAt: Date = Date().addingTimeInterval(3600)
    @State private var selectedTaskId: Int64? = nil
    @State private var selectedTaskTitle: String? = nil
    @State private var isShowingTaskPicker = false
    @State private var selectedSoundId: String = SoundManager.shared.selectedSoundId

    public init(viewModel: ReminderViewModel) {
        self.viewModel = viewModel
    }

    public var body: some View {
        NavigationStack {
            Form {
                // Section 1: Details
                Section {
                    TextField(L10n.Reminders.fieldTitle, text: $title)
                        .font(.system(size: 15))

                    TextField(L10n.Reminders.fieldNotes, text: $description, axis: .vertical)
                        .lineLimit(3...6)
                        .font(.system(size: 14))
                } header: {
                    Text(L10n.Reminders.sectionInfo)
                }

                // Section 2: Time & Date
                Section {
                    DatePicker(
                        L10n.Reminders.remindAt,
                        selection: $remindAt,
                        in: Date()...,
                        displayedComponents: [.date, .hourAndMinute]
                    )
                    .environment(\.timeZone, DateTimeHelper.userTimeZone)

                    // Quick Presets
                    VStack(alignment: .leading, spacing: 8) {
                        Text(L10n.Reminders.presets)
                            .font(.system(size: 11, weight: .medium))
                            .foregroundColor(.secondary)

                        HStack(spacing: 8) {
                            presetButton(label: "+30m", minutes: 30)
                            presetButton(label: "+1h", minutes: 60)
                            presetButton(label: "+3h", minutes: 180)
                            Button(action: applyTomorrowMorning) {
                                Text(L10n.Reminders.tomorrowMorning)
                                    .font(.system(size: 11, weight: .semibold))
                                    .padding(.horizontal, 8)
                                    .padding(.vertical, 6)
                                    .background(Color(.tertiarySystemGroupedBackground))
                                    .clipShape(Capsule())
                            }
                            .buttonStyle(.plain)
                        }
                    }
                    .padding(.vertical, 4)
                } header: {
                    Text(L10n.Reminders.sectionSchedule)
                }

                // Section: Nada Dering Alarm
                Section {
                    ForEach(SoundManager.presetSounds) { sound in
                        SoundOptionRowView(
                            sound: sound,
                            isSelected: selectedSoundId == sound.id,
                            onSelect: {
                                selectedSoundId = sound.id
                                SoundManager.shared.selectedSoundId = sound.id
                            }
                        )
                    }
                } header: {
                    Text(L10n.Alarm.ringtoneSection)
                }

                // Section 3: Link to Task (On-Demand Searchable)
                Section {
                    if let taskId = selectedTaskId {
                        HStack(spacing: 12) {
                            Image(systemName: "checkmark.circle.fill")
                                .foregroundColor(AppColors.primary)
                                .font(.system(size: 18))

                            VStack(alignment: .leading, spacing: 2) {
                                Text(selectedTaskTitle ?? "Task #\(taskId)")
                                    .font(.system(size: 14, weight: .semibold))
                                    .foregroundColor(.primary)
                                    .lineLimit(1)
                                Text(L10n.Reminders.linkedTask)
                                    .font(.system(size: 11))
                                    .foregroundColor(.secondary)
                            }

                            Spacer()

                            Button(action: {
                                selectedTaskId = nil
                                selectedTaskTitle = nil
                            }) {
                                Image(systemName: "xmark.circle.fill")
                                    .foregroundColor(.secondary)
                                    .font(.system(size: 18))
                            }
                            .buttonStyle(.plain)
                        }
                        .padding(.vertical, 4)
                    } else {
                        Button(action: {
                            isShowingTaskPicker = true
                        }) {
                            HStack {
                                Label(L10n.Reminders.selectTask, systemImage: "link")
                                    .font(.system(size: 14, weight: .medium))
                                    .foregroundColor(AppColors.primary)
                                Spacer()
                                Image(systemName: "chevron.right")
                                    .font(.system(size: 12, weight: .semibold))
                                    .foregroundColor(.secondary)
                            }
                            .padding(.vertical, 2)
                        }
                        .buttonStyle(.plain)
                    }
                } header: {
                    Text(L10n.Reminders.sectionAssociation)
                }

                if let err = viewModel.errorMessage {
                    Section {
                        Text(err)
                            .font(.system(size: 12))
                            .foregroundColor(AppColors.priorityUrgent)
                    }
                }
            }
            .navigationTitle(L10n.Reminders.newReminder)
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button(L10n.General.cancel) {
                        dismiss()
                    }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button {
                        Task {
                            let ok = await viewModel.createReminder(
                                title: title,
                                description: description.isEmpty ? nil : description,
                                remindAt: remindAt,
                                taskId: selectedTaskId
                            )
                            if ok {
                                dismiss()
                            }
                        }
                    } label: {
                        if viewModel.isCreating {
                            ProgressView()
                                .scaleEffect(0.8)
                        } else {
                            Text(L10n.General.create)
                                .fontWeight(.bold)
                        }
                    }
                    .disabled(title.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty || viewModel.isCreating)
                }
            }
            .sheet(isPresented: $isShowingTaskPicker) {
                TaskSearchPickerSheet(viewModel: viewModel) { chosenTask in
                    if let t = chosenTask {
                        selectedTaskId = t.id
                        selectedTaskTitle = t.title
                    } else {
                        selectedTaskId = nil
                        selectedTaskTitle = nil
                    }
                }
            }
        }
    }

    private func presetButton(label: String, minutes: Int) -> some View {
        Button(action: {
            remindAt = Date().addingTimeInterval(TimeInterval(minutes * 60))
        }) {
            Text(label)
                .font(.system(size: 11, weight: .semibold))
                .padding(.horizontal, 10)
                .padding(.vertical, 6)
                .background(Color(.tertiarySystemGroupedBackground))
                .clipShape(Capsule())
        }
        .buttonStyle(.plain)
    }

    private func applyTomorrowMorning() {
        var calendar = Calendar.current
        calendar.timeZone = DateTimeHelper.userTimeZone
        var comps = calendar.dateComponents([.year, .month, .day], from: Date().addingTimeInterval(86400))
        comps.hour = 9
        comps.minute = 0
        comps.second = 0
        if let d = calendar.date(from: comps) {
            remindAt = d
        }
    }
}

// MARK: - Edit Reminder Sheet Modal
public struct EditReminderSheet: View {
    let reminder: Reminder
    @Bindable var viewModel: ReminderViewModel
    @Environment(\.dismiss) private var dismiss

    @State private var title: String
    @State private var description: String
    @State private var remindAt: Date
    @State private var status: String
    @State private var selectedTaskId: Int64?
    @State private var selectedTaskTitle: String?
    @State private var isShowingTaskPicker = false
    @State private var selectedSoundId: String = SoundManager.shared.selectedSoundId
    @State private var isSubmitting: Bool = false
    @State private var errorMessage: String? = nil

    public init(reminder: Reminder, viewModel: ReminderViewModel) {
        self.reminder = reminder
        self.viewModel = viewModel

        _title = State(initialValue: reminder.title)
        _description = State(initialValue: reminder.description)
        _status = State(initialValue: reminder.status.lowercased())
        _selectedTaskId = State(initialValue: reminder.taskId)
        _selectedTaskTitle = State(initialValue: reminder.taskTitle)

        let parsed = DateTimeHelper.parseISO8601(reminder.remindAt)
        _remindAt = State(initialValue: parsed ?? Date().addingTimeInterval(3600))
    }

    public var body: some View {
        NavigationStack {
            Form {
                // Section 1: Details
                Section {
                    TextField(L10n.Reminders.fieldTitle, text: $title)
                        .font(.system(size: 15))

                    TextField(L10n.Reminders.fieldNotes, text: $description, axis: .vertical)
                        .lineLimit(3...6)
                        .font(.system(size: 14))
                } header: {
                    Text(L10n.Reminders.sectionInfo)
                }

                // Section 2: Time & Date
                Section {
                    DatePicker(
                        L10n.Reminders.remindAt,
                        selection: $remindAt,
                        displayedComponents: [.date, .hourAndMinute]
                    )
                    .environment(\.timeZone, DateTimeHelper.userTimeZone)

                    // Quick Presets
                    VStack(alignment: .leading, spacing: 8) {
                        Text(L10n.Reminders.presets)
                            .font(.system(size: 11, weight: .medium))
                            .foregroundColor(.secondary)

                        HStack(spacing: 8) {
                            presetButton(label: "+30m", minutes: 30)
                            presetButton(label: "+1h", minutes: 60)
                            presetButton(label: "+3h", minutes: 180)
                            Button(action: applyTomorrowMorning) {
                                Text(L10n.Reminders.tomorrowMorning)
                                    .font(.system(size: 11, weight: .semibold))
                                    .padding(.horizontal, 8)
                                    .padding(.vertical, 6)
                                    .background(Color(.tertiarySystemGroupedBackground))
                                    .clipShape(Capsule())
                            }
                            .buttonStyle(.plain)
                        }
                    }
                    .padding(.vertical, 4)
                } header: {
                    Text(L10n.Reminders.sectionSchedule)
                }

                // Section: Nada Dering Alarm
                Section {
                    ForEach(SoundManager.presetSounds) { sound in
                        SoundOptionRowView(
                            sound: sound,
                            isSelected: selectedSoundId == sound.id,
                            onSelect: {
                                selectedSoundId = sound.id
                                SoundManager.shared.selectedSoundId = sound.id
                            }
                        )
                    }
                } header: {
                    Text(L10n.Alarm.ringtoneSection)
                }

                // Section 3: Status
                Section {
                    Picker(L10n.Reminders.status, selection: $status) {
                        Text(L10n.Reminders.statusPending).tag("pending")
                        Text(L10n.Reminders.statusSent).tag("sent")
                        Text(L10n.Reminders.statusDismissed).tag("dismissed")
                    }
                } header: {
                    Text(L10n.Reminders.status)
                }

                // Section 4: Task Association (Optional)
                Section {
                    if let taskId = selectedTaskId {
                        HStack(spacing: 12) {
                            Image(systemName: "checkmark.circle.fill")
                                .foregroundColor(AppColors.primary)
                                .font(.system(size: 18))

                            VStack(alignment: .leading, spacing: 2) {
                                Text(selectedTaskTitle ?? "Task #\(taskId)")
                                    .font(.system(size: 14, weight: .semibold))
                                    .foregroundColor(.primary)
                                    .lineLimit(1)
                                Text(L10n.Reminders.linkedTask)
                                    .font(.system(size: 11))
                                    .foregroundColor(.secondary)
                            }

                            Spacer()

                            Button(action: {
                                selectedTaskId = nil
                                selectedTaskTitle = nil
                            }) {
                                Image(systemName: "xmark.circle.fill")
                                    .foregroundColor(.secondary)
                                    .font(.system(size: 18))
                            }
                            .buttonStyle(.plain)
                        }
                        .padding(.vertical, 4)
                    } else {
                        Button(action: {
                            isShowingTaskPicker = true
                        }) {
                            HStack {
                                Label(L10n.Reminders.selectTask, systemImage: "link")
                                    .font(.system(size: 14, weight: .medium))
                                    .foregroundColor(AppColors.primary)
                                Spacer()
                                Image(systemName: "chevron.right")
                                    .font(.system(size: 12, weight: .semibold))
                                    .foregroundColor(.secondary)
                            }
                            .padding(.vertical, 2)
                        }
                        .buttonStyle(.plain)
                    }
                } header: {
                    Text(L10n.Reminders.sectionAssociation)
                }

                if let err = errorMessage ?? viewModel.errorMessage {
                    Section {
                        Text(err)
                            .font(.system(size: 12))
                            .foregroundColor(AppColors.priorityUrgent)
                    }
                }
            }
            .navigationTitle(L10n.Reminders.editReminder)
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button(L10n.General.cancel) {
                        dismiss()
                    }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button {
                        handleSave()
                    } label: {
                        if isSubmitting || viewModel.isUpdating {
                            ProgressView()
                                .scaleEffect(0.8)
                        } else {
                            Text(L10n.General.save)
                                .fontWeight(.bold)
                        }
                    }
                    .disabled(title.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty || isSubmitting || viewModel.isUpdating)
                }
            }
            .sheet(isPresented: $isShowingTaskPicker) {
                TaskSearchPickerSheet(viewModel: viewModel) { chosenTask in
                    if let t = chosenTask {
                        selectedTaskId = t.id
                        selectedTaskTitle = t.title
                    } else {
                        selectedTaskId = nil
                        selectedTaskTitle = nil
                    }
                }
            }
        }
    }

    private func handleSave() {
        let trimmedTitle = title.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmedTitle.isEmpty else { return }

        isSubmitting = true
        errorMessage = nil

        let clearTask = (selectedTaskId == nil && reminder.taskId != nil)

        Task {
            let ok = await viewModel.updateReminder(
                id: reminder.id,
                title: trimmedTitle,
                description: description.isEmpty ? nil : description,
                remindAt: remindAt,
                status: status,
                taskId: selectedTaskId,
                clearTaskId: clearTask
            )
            await MainActor.run {
                isSubmitting = false
                if ok {
                    dismiss()
                } else {
                    errorMessage = viewModel.errorMessage
                }
            }
        }
    }

    private func presetButton(label: String, minutes: Int) -> some View {
        Button(action: {
            remindAt = Date().addingTimeInterval(TimeInterval(minutes * 60))
        }) {
            Text(label)
                .font(.system(size: 11, weight: .semibold))
                .padding(.horizontal, 10)
                .padding(.vertical, 6)
                .background(Color(.tertiarySystemGroupedBackground))
                .clipShape(Capsule())
        }
        .buttonStyle(.plain)
    }

    private func applyTomorrowMorning() {
        var calendar = Calendar.current
        calendar.timeZone = DateTimeHelper.userTimeZone
        var comps = calendar.dateComponents([.year, .month, .day], from: Date().addingTimeInterval(86400))
        comps.hour = 9
        comps.minute = 0
        comps.second = 0
        if let d = calendar.date(from: comps) {
            remindAt = d
        }
    }
}

// MARK: - Dedicated Lightweight Task Search Picker Sheet (Limits to 15 items)
public struct TaskSearchPickerSheet: View {
    @Environment(\.dismiss) private var dismiss
    let viewModel: ReminderViewModel
    let onSelect: (TaskItem?) -> Void

    @State private var searchQuery: String = ""
    @State private var searchResults: [TaskItem] = []
    @State private var isSearching: Bool = false
    @State private var searchTask: Task<Void, Never>? = nil

    public init(viewModel: ReminderViewModel, onSelect: @escaping (TaskItem?) -> Void) {
        self.viewModel = viewModel
        self.onSelect = onSelect
    }

    public var body: some View {
        NavigationStack {
            VStack(spacing: 0) {
                // Search bar
                HStack(spacing: 8) {
                    Image(systemName: "magnifyingglass")
                        .foregroundColor(.secondary)
                        .font(.system(size: 14))

                    TextField(L10n.Reminders.searchTasksPlaceholder, text: $searchQuery)
                        .font(.system(size: 14))
                        .autocorrectionDisabled()
                        .textInputAutocapitalization(.never)
                        .onChange(of: searchQuery) { _, newValue in
                            scheduleSearch(query: newValue)
                        }

                    if !searchQuery.isEmpty {
                        Button(action: {
                            searchQuery = ""
                            scheduleSearch(query: "")
                        }) {
                            Image(systemName: "xmark.circle.fill")
                                .foregroundColor(.secondary)
                                .font(.system(size: 14))
                        }
                        .buttonStyle(.plain)
                    }
                }
                .padding(.horizontal, 12)
                .padding(.vertical, 9)
                .background(Color(.secondarySystemGroupedBackground))
                .clipShape(RoundedRectangle(cornerRadius: 10))
                .padding(.horizontal, 16)
                .padding(.vertical, 10)

                Divider()

                List {
                    // Option to select None / Clear
                    Button(action: {
                        onSelect(nil)
                        dismiss()
                    }) {
                        HStack(spacing: 12) {
                            Image(systemName: "xmark.circle")
                                .foregroundColor(.secondary)
                                .font(.system(size: 16))
                            Text(L10n.Reminders.taskNone)
                                .font(.system(size: 14))
                                .foregroundColor(.primary)
                            Spacer()
                        }
                    }

                    if isSearching && searchResults.isEmpty {
                        HStack {
                            Spacer()
                            ProgressView()
                                .scaleEffect(0.9)
                            Spacer()
                        }
                        .listRowSeparator(.hidden)
                    } else if searchResults.isEmpty && !isSearching {
                        VStack(spacing: 8) {
                            Image(systemName: "doc.text.magnifyingglass")
                                .font(.system(size: 28))
                                .foregroundColor(.secondary.opacity(0.6))
                            Text(L10n.Reminders.noTasksFound)
                                .font(.system(size: 13, weight: .medium))
                                .foregroundColor(.secondary)
                        }
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 24)
                        .listRowSeparator(.hidden)
                    } else {
                        ForEach(searchResults) { task in
                            Button(action: {
                                onSelect(task)
                                dismiss()
                            }) {
                                VStack(alignment: .leading, spacing: 6) {
                                    HStack(alignment: .top, spacing: 8) {
                                        Text(task.title)
                                            .font(.system(size: 14, weight: .semibold))
                                            .foregroundColor(.primary)
                                            .lineLimit(2)
                                        Spacer()
                                        taskStatusBadge(status: task.status)
                                    }

                                    HStack(spacing: 8) {
                                        if let pName = task.projectName, !pName.isEmpty {
                                            HStack(spacing: 4) {
                                                Image(systemName: "folder.fill")
                                                    .font(.system(size: 9))
                                                Text(pName)
                                                    .font(.system(size: 11, weight: .medium))
                                            }
                                            .foregroundColor(.secondary)
                                        }

                                        if let due = task.dueDate, !due.isEmpty {
                                            HStack(spacing: 4) {
                                                Image(systemName: "calendar")
                                                    .font(.system(size: 9))
                                                Text(formatDate(due))
                                                    .font(.system(size: 11))
                                            }
                                            .foregroundColor(.secondary)
                                        }
                                    }
                                }
                                .padding(.vertical, 4)
                            }
                        }
                    }
                }
                .listStyle(.insetGrouped)
            }
            .navigationTitle(L10n.Reminders.selectTask)
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button(L10n.General.cancel) {
                        dismiss()
                    }
                }
            }
            .task {
                await performSearch(query: "")
            }
        }
    }

    private func scheduleSearch(query: String) {
        searchTask?.cancel()
        searchTask = Task {
            try? await Task.sleep(nanoseconds: 250_000_000) // 250ms debounce
            guard !Task.isCancelled else { return }
            await performSearch(query: query)
        }
    }

    private func performSearch(query: String) async {
        isSearching = true
        let results = await viewModel.searchTasks(query: query)
        if !Task.isCancelled {
            self.searchResults = results
            self.isSearching = false
        }
    }

    private func taskStatusBadge(status: String) -> some View {
        let (label, color) = statusDisplay(status)
        return Text(label)
            .font(.system(size: 10, weight: .semibold))
            .padding(.horizontal, 6)
            .padding(.vertical, 2)
            .background(color.opacity(0.12))
            .foregroundColor(color)
            .clipShape(Capsule())
    }

    private func statusDisplay(_ status: String) -> (String, Color) {
        switch status.lowercased() {
        case "in_progress": return ("In Progress", AppColors.primary)
        case "review": return ("Review", .purple)
        case "done": return ("Done", AppColors.priorityLow)
        case "backlog": return ("Backlog", .secondary)
        default: return ("Todo", AppColors.statusTodo)
        }
    }

    private func formatDate(_ dateStr: String) -> String {
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        var date = formatter.date(from: dateStr)
        if date == nil {
            let basic = ISO8601DateFormatter()
            date = basic.date(from: dateStr)
        }
        guard let validDate = date else { return dateStr }
        let outFormatter = DateFormatter()
        outFormatter.dateStyle = .medium
        outFormatter.timeStyle = .none
        return outFormatter.string(from: validDate)
    }
}
