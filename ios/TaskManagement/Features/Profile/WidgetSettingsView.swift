import SwiftUI

/// BCA-Style Widget Configuration Screen
/// Allows users to choose which active non-completed tasks (both recurring & regular)
/// appear on their Home Screen widgets, with zero selection limit.
public struct WidgetSettingsView: View {
    @Environment(\.dismiss) private var dismiss
    @State private var tasks: [TaskItem] = []
    @State private var isLoading = false
    @State private var errorMessage: String?
    @State private var searchQuery: String = ""

    // Stored selection sets from App Group
    @State private var selectedRecurringIds: Set<Int64> = []
    @State private var selectedCustomIds: Set<Int64> = []
    @State private var isAutoSelectAll: Bool = true
    @State private var excludedTaskIds: Set<Int64> = []

    private let widgetManager = SharedWidgetDataManager.shared
    private let apiClient = APIClient.shared

    public init() {}

    private var activeTasks: [TaskItem] {
        tasks.filter { $0.status != "done" }
    }

    private var filteredRecurringTasks: [TaskItem] {
        let list = activeTasks.filter { $0.isRecurring == true || $0.recurringTaskId != nil }
        guard !searchQuery.trimmingCharacters(in: .whitespaces).isEmpty else { return list }
        return list.filter {
            $0.title.localizedCaseInsensitiveContains(searchQuery) ||
            ($0.projectName?.localizedCaseInsensitiveContains(searchQuery) ?? false)
        }
    }

    private var filteredCustomTasks: [TaskItem] {
        let list = activeTasks.filter { $0.isRecurring != true && $0.recurringTaskId == nil }
        guard !searchQuery.trimmingCharacters(in: .whitespaces).isEmpty else { return list }
        return list.filter {
            $0.title.localizedCaseInsensitiveContains(searchQuery) ||
            ($0.projectName?.localizedCaseInsensitiveContains(searchQuery) ?? false)
        }
    }

    public var body: some View {
        ScrollView {
            VStack(spacing: 16) {
                // Auto-Select Master Switch
                autoSelectMasterCardView

                // Info Tip Card
                tipCardView

                // Search Bar
                searchBarView

                if isLoading && tasks.isEmpty {
                    loadingView
                } else if let error = errorMessage, tasks.isEmpty {
                    errorView(error)
                } else if activeTasks.isEmpty {
                    emptyStateView
                } else {
                    // Section 1: Recurring Tasks
                    recurringSectionView

                    // Section 2: Standard Tasks
                    customSectionView
                }
            }
            .padding(.horizontal, 16)
            .padding(.vertical, 14)
        }
        .background(Color(.systemGroupedBackground))
        .navigationTitle(L10n.Settings.widgetSettings)
        .navigationBarTitleDisplayMode(.inline)
        .task {
            loadSelections()
            await fetchTasks()
        }
        .refreshable {
            await fetchTasks()
        }
    }

    // MARK: - Auto-Select Master Card
    private var autoSelectMasterCardView: some View {
        VStack(spacing: 12) {
            HStack(alignment: .center, spacing: 12) {
                // Icon
                ZStack {
                    Circle()
                        .fill(
                            LinearGradient(
                                colors: isAutoSelectAll ? [
                                    Color(red: 16/255, green: 185/255, blue: 129/255),
                                    Color(red: 6/255, green: 182/255, blue: 212/255)
                                ] : [
                                    Color.secondary.opacity(0.2),
                                    Color.secondary.opacity(0.1)
                                ],
                                startPoint: .topLeading,
                                endPoint: .bottomTrailing
                            )
                        )
                        .frame(width: 40, height: 40)

                    Image(systemName: isAutoSelectAll ? "wand.and.stars" : "slider.horizontal.3")
                        .font(.system(size: 18, weight: .semibold))
                        .foregroundColor(isAutoSelectAll ? .white : .secondary)
                }

                VStack(alignment: .leading, spacing: 3) {
                    HStack(spacing: 6) {
                        Text(L10n.Settings.widgetAutoSelectAll)
                            .font(.system(size: 14, weight: .bold))
                            .foregroundColor(.primary)

                        Text(isAutoSelectAll ? L10n.Settings.widgetAutoSelectBadge : L10n.Settings.widgetManualBadge)
                            .font(.system(size: 9, weight: .black))
                            .foregroundColor(isAutoSelectAll ? Color(red: 16/255, green: 185/255, blue: 129/255) : .secondary)
                            .padding(.horizontal, 6)
                            .padding(.vertical, 2)
                            .background((isAutoSelectAll ? Color(red: 16/255, green: 185/255, blue: 129/255) : Color.secondary).opacity(0.12))
                            .clipShape(Capsule())
                    }

                    Text(L10n.Settings.widgetAutoSelectAllDesc)
                        .font(.system(size: 11))
                        .foregroundColor(.secondary)
                        .fixedSize(horizontal: false, vertical: true)
                }

                Spacer()

                Toggle("", isOn: Binding(
                    get: { isAutoSelectAll },
                    set: { newVal in
                        handleAutoSelectToggle(newVal)
                    }
                ))
                .labelsHidden()
                .tint(Color(red: 16/255, green: 185/255, blue: 129/255))
            }

            if isAutoSelectAll {
                HStack(spacing: 6) {
                    Image(systemName: "checkmark.circle.badge.questionmark")
                        .font(.system(size: 11, weight: .semibold))
                        .foregroundColor(Color(red: 16/255, green: 185/255, blue: 129/255))
                    Text(L10n.Settings.widgetAutoSelectHint)
                        .font(.system(size: 11, weight: .medium))
                        .foregroundColor(Color(red: 16/255, green: 185/255, blue: 129/255))
                    Spacer()
                }
                .padding(.horizontal, 10)
                .padding(.vertical, 6)
                .background(Color(red: 16/255, green: 185/255, blue: 129/255).opacity(0.08))
                .clipShape(RoundedRectangle(cornerRadius: 8, style: .continuous))
            }
        }
        .padding(14)
        .background(Color(.secondarySystemGroupedBackground))
        .clipShape(RoundedRectangle(cornerRadius: 16, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: 16, style: .continuous)
                .stroke(isAutoSelectAll ? Color(red: 16/255, green: 185/255, blue: 129/255).opacity(0.3) : Color.primary.opacity(0.06), lineWidth: 1.5)
        )
    }

    // MARK: - Tip Banner
    private var tipCardView: some View {
        HStack(alignment: .top, spacing: 10) {
            Image(systemName: "info.circle.fill")
                .font(.system(size: 16))
                .foregroundColor(AppColors.primary)
                .padding(.top, 1)

            VStack(alignment: .leading, spacing: 3) {
                Text(L10n.Settings.widgetSettingsDesc)
                    .font(.system(size: 13, weight: .semibold))
                    .foregroundColor(.primary)

                Text(L10n.Settings.widgetTip)
                    .font(.system(size: 11))
                    .foregroundColor(.secondary)
                    .fixedSize(horizontal: false, vertical: true)
            }
        }
        .padding(12)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(AppColors.primary.opacity(0.08))
        .clipShape(RoundedRectangle(cornerRadius: 14, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: 14, style: .continuous)
                .stroke(AppColors.primary.opacity(0.18), lineWidth: 1)
        )
    }

    // MARK: - Search Bar
    private var searchBarView: some View {
        HStack(spacing: 8) {
            Image(systemName: "magnifyingglass")
                .font(.system(size: 14))
                .foregroundColor(.secondary)

            TextField(L10n.Settings.widgetSearchPlaceholder, text: $searchQuery)
                .font(.system(size: 14))
                .textFieldStyle(.plain)

            if !searchQuery.isEmpty {
                Button(action: { searchQuery = "" }) {
                    Image(systemName: "xmark.circle.fill")
                        .font(.system(size: 13))
                        .foregroundColor(.secondary)
                }
            }
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 9)
        .background(Color(.secondarySystemGroupedBackground))
        .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))
    }

    // MARK: - Section 1: Recurring Tasks
    private var recurringSectionView: some View {
        VStack(alignment: .leading, spacing: 12) {
            // Header
            HStack(alignment: .center) {
                VStack(alignment: .leading, spacing: 2) {
                    HStack(spacing: 6) {
                        Image(systemName: "repeat.circle.fill")
                            .foregroundColor(Color(red: 56/255, green: 189/255, blue: 248/255))
                            .font(.system(size: 15))
                        Text(L10n.Settings.widgetRecurringSection)
                            .font(.system(size: 15, weight: .bold))
                            .foregroundColor(.primary)
                    }

                    Text(L10n.Settings.widgetRecurringSectionDesc)
                        .font(.system(size: 11))
                        .foregroundColor(.secondary)
                }

                Spacer()

                // Selection Counter Badge
                let count = filteredRecurringTasks.filter { selectedRecurringIds.contains($0.id) }.count
                Text(L10n.Settings.widgetSelectedCount(count: count))
                    .font(.system(size: 11, weight: .semibold))
                    .foregroundColor(Color(red: 56/255, green: 189/255, blue: 248/255))
                    .padding(.horizontal, 8)
                    .padding(.vertical, 3)
                    .background(Color(red: 56/255, green: 189/255, blue: 248/255).opacity(0.12))
                    .clipShape(Capsule())
            }

            // Quick Actions: Select All / Deselect All
            if !filteredRecurringTasks.isEmpty {
                HStack {
                    Button(action: selectAllRecurring) {
                        Text(L10n.Settings.widgetSelectAll)
                            .font(.system(size: 11, weight: .semibold))
                            .foregroundColor(AppColors.primary)
                    }

                    Text("•")
                        .font(.system(size: 11))
                        .foregroundColor(.secondary)

                    Button(action: deselectAllRecurring) {
                        Text(L10n.Settings.widgetDeselectAll)
                            .font(.system(size: 11, weight: .semibold))
                            .foregroundColor(.secondary)
                    }

                    Spacer()
                }
                .padding(.top, -2)
            }

            // Task List
            if filteredRecurringTasks.isEmpty {
                Text(L10n.Settings.widgetEmptyTasks)
                    .font(.system(size: 12))
                    .foregroundColor(.secondary)
                    .padding(.vertical, 10)
            } else {
                VStack(spacing: 8) {
                    ForEach(filteredRecurringTasks) { task in
                        taskRow(
                            task: task,
                            isSelected: selectedRecurringIds.contains(task.id),
                            accentColor: Color(red: 56/255, green: 189/255, blue: 248/255),
                            onTap: { toggleRecurring(task) }
                        )
                    }
                }
            }
        }
        .padding(14)
        .background(Color(.secondarySystemGroupedBackground))
        .clipShape(RoundedRectangle(cornerRadius: 16, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: 16, style: .continuous)
                .stroke(Color.primary.opacity(0.06), lineWidth: 1)
        )
    }

    // MARK: - Section 2: Custom / Standard Tasks
    private var customSectionView: some View {
        VStack(alignment: .leading, spacing: 12) {
            // Header
            HStack(alignment: .center) {
                VStack(alignment: .leading, spacing: 2) {
                    HStack(spacing: 6) {
                        Image(systemName: "checklist.checked")
                            .foregroundColor(Color(red: 168/255, green: 85/255, blue: 247/255))
                            .font(.system(size: 15))
                        Text(L10n.Settings.widgetCustomSection)
                            .font(.system(size: 15, weight: .bold))
                            .foregroundColor(.primary)
                    }

                    Text(L10n.Settings.widgetCustomSectionDesc)
                        .font(.system(size: 11))
                        .foregroundColor(.secondary)
                }

                Spacer()

                // Selection Counter Badge
                let count = filteredCustomTasks.filter { selectedCustomIds.contains($0.id) }.count
                Text(L10n.Settings.widgetSelectedCount(count: count))
                    .font(.system(size: 11, weight: .semibold))
                    .foregroundColor(Color(red: 168/255, green: 85/255, blue: 247/255))
                    .padding(.horizontal, 8)
                    .padding(.vertical, 3)
                    .background(Color(red: 168/255, green: 85/255, blue: 247/255).opacity(0.12))
                    .clipShape(Capsule())
            }

            // Quick Actions: Select All / Deselect All
            if !filteredCustomTasks.isEmpty {
                HStack {
                    Button(action: selectAllCustom) {
                        Text(L10n.Settings.widgetSelectAll)
                            .font(.system(size: 11, weight: .semibold))
                            .foregroundColor(AppColors.primary)
                    }

                    Text("•")
                        .font(.system(size: 11))
                        .foregroundColor(.secondary)

                    Button(action: deselectAllCustom) {
                        Text(L10n.Settings.widgetDeselectAll)
                            .font(.system(size: 11, weight: .semibold))
                            .foregroundColor(.secondary)
                    }

                    Spacer()
                }
                .padding(.top, -2)
            }

            // Task List
            if filteredCustomTasks.isEmpty {
                Text(L10n.Settings.widgetEmptyTasks)
                    .font(.system(size: 12))
                    .foregroundColor(.secondary)
                    .padding(.vertical, 10)
            } else {
                VStack(spacing: 8) {
                    ForEach(filteredCustomTasks) { task in
                        taskRow(
                            task: task,
                            isSelected: selectedCustomIds.contains(task.id),
                            accentColor: Color(red: 168/255, green: 85/255, blue: 247/255),
                            onTap: { toggleCustom(task) }
                        )
                    }
                }
            }
        }
        .padding(14)
        .background(Color(.secondarySystemGroupedBackground))
        .clipShape(RoundedRectangle(cornerRadius: 16, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: 16, style: .continuous)
                .stroke(Color.primary.opacity(0.06), lineWidth: 1)
        )
    }

    // MARK: - Task Row
    @ViewBuilder
    private func taskRow(task: TaskItem, isSelected: Bool, accentColor: Color, onTap: @escaping () -> Void) -> some View {
        Button(action: {
            UIImpactFeedbackGenerator(style: .light).impactOccurred()
            onTap()
        }) {
            HStack(spacing: 12) {
                // Checkbox
                Image(systemName: isSelected ? "checkmark.circle.fill" : "circle")
                    .font(.system(size: 20, weight: .semibold))
                    .foregroundColor(isSelected ? accentColor : .secondary.opacity(0.5))

                // Task Information
                VStack(alignment: .leading, spacing: 3) {
                    Text(task.title)
                        .font(.system(size: 14, weight: .medium))
                        .foregroundColor(.primary)
                        .lineLimit(2)
                        .multilineTextAlignment(.leading)

                    HStack(spacing: 6) {
                        if let projectName = task.projectName, !projectName.isEmpty {
                            Text(projectName)
                                .font(.system(size: 10, weight: .medium))
                                .foregroundColor(.secondary)
                                .padding(.horizontal, 5)
                                .padding(.vertical, 1.5)
                                .background(Color(.tertiarySystemGroupedBackground))
                                .clipShape(RoundedRectangle(cornerRadius: 4))
                        }

                        if let count = task.subtasksCount, count > 0 {
                            let completed = task.subtasksCompleted ?? 0
                            HStack(spacing: 2) {
                                Image(systemName: "checklist")
                                    .font(.system(size: 8))
                                Text("\(completed)/\(count)")
                                    .font(.system(size: 10, weight: .semibold, design: .monospaced))
                            }
                            .foregroundColor(completed == count ? AppColors.statusDone : .secondary)
                        }

                        if !task.priority.isEmpty {
                            Text(task.priority.capitalized)
                                .font(.system(size: 9, weight: .bold))
                                .foregroundColor(priorityBadgeColor(task.priority))
                        }
                    }
                }

                Spacer()
            }
            .padding(10)
            .background(isSelected ? accentColor.opacity(0.08) : Color(.tertiarySystemGroupedBackground).opacity(0.5))
            .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))
            .overlay(
                RoundedRectangle(cornerRadius: 10, style: .continuous)
                    .stroke(isSelected ? accentColor.opacity(0.3) : Color.clear, lineWidth: 1)
            )
        }
        .buttonStyle(.plain)
    }

    // MARK: - State Helpers
    private var loadingView: some View {
        VStack(spacing: 10) {
            ProgressView()
                .scaleEffect(1.1)
            Text(L10n.General.loading)
                .font(.system(size: 13))
                .foregroundColor(.secondary)
        }
        .frame(maxWidth: .infinity)
        .padding(.vertical, 40)
    }

    @ViewBuilder
    private func errorView(_ msg: String) -> some View {
        VStack(spacing: 8) {
            Image(systemName: "exclamationmark.triangle.fill")
                .font(.system(size: 26))
                .foregroundColor(AppColors.priorityUrgent)
            Text(msg)
                .font(.system(size: 13))
                .foregroundColor(.secondary)
                .multilineTextAlignment(.center)
            Button(L10n.General.retry) {
                Task { await fetchTasks() }
            }
            .font(.system(size: 13, weight: .semibold))
            .foregroundColor(AppColors.primary)
        }
        .frame(maxWidth: .infinity)
        .padding(.vertical, 30)
    }

    private var emptyStateView: some View {
        VStack(spacing: 8) {
            Image(systemName: "tray.fill")
                .font(.system(size: 28))
                .foregroundColor(.secondary.opacity(0.5))
            Text(L10n.Settings.widgetEmptyTasks)
                .font(.system(size: 14, weight: .semibold))
                .foregroundColor(.primary)
            Text(L10n.Settings.widgetEmptyDesc)
                .font(.system(size: 12))
                .foregroundColor(.secondary)
                .multilineTextAlignment(.center)
        }
        .frame(maxWidth: .infinity)
        .padding(.vertical, 40)
    }

    // MARK: - Data Management & Toggle Actions
    private func loadSelections() {
        isAutoSelectAll = widgetManager.isAutoSelectAllEnabled()
        excludedTaskIds = widgetManager.getExcludedTaskIds()
        selectedRecurringIds = widgetManager.getSelectedRecurringTaskIds()
        selectedCustomIds = widgetManager.getSelectedCustomTaskIds()
    }

    @MainActor
    private func fetchTasks() async {
        isLoading = true
        errorMessage = nil
        do {
            let fetched = try await apiClient.getTasks(limit: 100)
            // Filter strictly: only active tasks (never done)
            self.tasks = fetched.filter { $0.status != "done" }
            let activeIds = Set(self.tasks.map { $0.id })

            // Clean up excluded tasks against current active tasks
            excludedTaskIds = excludedTaskIds.intersection(activeIds)
            widgetManager.setExcludedTaskIds(excludedTaskIds)

            if isAutoSelectAll {
                // Auto-select * : select all active tasks not explicitly excluded
                let recActive = Set(self.tasks.filter { ($0.isRecurring == true || $0.recurringTaskId != nil) && !excludedTaskIds.contains($0.id) }.map { $0.id })
                let cusActive = Set(self.tasks.filter { ($0.isRecurring != true && $0.recurringTaskId == nil) && !excludedTaskIds.contains($0.id) }.map { $0.id })
                selectedRecurringIds = recActive
                selectedCustomIds = cusActive
            } else {
                selectedRecurringIds = selectedRecurringIds.intersection(activeIds)
                selectedCustomIds = selectedCustomIds.intersection(activeIds)
            }

            widgetManager.setSelectedRecurringTaskIds(selectedRecurringIds)
            widgetManager.setSelectedCustomTaskIds(selectedCustomIds)
            // Sync widget bundle
            widgetManager.syncWithTaskList(self.tasks)
        } catch {
            self.errorMessage = error.localizedDescription
            AppLogger.error("Failed to fetch tasks for widget settings", error: error, category: .task)
        }
        isLoading = false
    }

    private func handleAutoSelectToggle(_ enabled: Bool) {
        UIImpactFeedbackGenerator(style: .medium).impactOccurred()
        isAutoSelectAll = enabled
        widgetManager.setAutoSelectAllEnabled(enabled)
        if enabled {
            excludedTaskIds.removeAll()
            widgetManager.setExcludedTaskIds([])
            let recActive = Set(tasks.filter { ($0.isRecurring == true || $0.recurringTaskId != nil) && $0.status != "done" }.map { $0.id })
            let cusActive = Set(tasks.filter { ($0.isRecurring != true && $0.recurringTaskId == nil) && $0.status != "done" }.map { $0.id })
            selectedRecurringIds = recActive
            selectedCustomIds = cusActive
            widgetManager.setSelectedRecurringTaskIds(recActive)
            widgetManager.setSelectedCustomTaskIds(cusActive)
        }
        widgetManager.syncWithTaskList(tasks)
    }

    private func toggleRecurring(_ task: TaskItem) {
        if isAutoSelectAll {
            if excludedTaskIds.contains(task.id) {
                excludedTaskIds.remove(task.id)
                selectedRecurringIds.insert(task.id)
            } else {
                excludedTaskIds.insert(task.id)
                selectedRecurringIds.remove(task.id)
            }
            widgetManager.setExcludedTaskIds(excludedTaskIds)
            widgetManager.setSelectedRecurringTaskIds(selectedRecurringIds)
        } else {
            if selectedRecurringIds.contains(task.id) {
                selectedRecurringIds.remove(task.id)
            } else {
                selectedRecurringIds.insert(task.id)
            }
            widgetManager.setSelectedRecurringTaskIds(selectedRecurringIds)
        }
        widgetManager.syncWithTaskList(tasks)
    }

    private func toggleCustom(_ task: TaskItem) {
        if isAutoSelectAll {
            if excludedTaskIds.contains(task.id) {
                excludedTaskIds.remove(task.id)
                selectedCustomIds.insert(task.id)
            } else {
                excludedTaskIds.insert(task.id)
                selectedCustomIds.remove(task.id)
            }
            widgetManager.setExcludedTaskIds(excludedTaskIds)
            widgetManager.setSelectedCustomTaskIds(selectedCustomIds)
        } else {
            if selectedCustomIds.contains(task.id) {
                selectedCustomIds.remove(task.id)
            } else {
                selectedCustomIds.insert(task.id)
            }
            widgetManager.setSelectedCustomTaskIds(selectedCustomIds)
        }
        widgetManager.syncWithTaskList(tasks)
    }

    private func selectAllRecurring() {
        for t in filteredRecurringTasks {
            selectedRecurringIds.insert(t.id)
            excludedTaskIds.remove(t.id)
        }
        widgetManager.setExcludedTaskIds(excludedTaskIds)
        widgetManager.setSelectedRecurringTaskIds(selectedRecurringIds)
        widgetManager.syncWithTaskList(tasks)
    }

    private func deselectAllRecurring() {
        for t in filteredRecurringTasks {
            selectedRecurringIds.remove(t.id)
            if isAutoSelectAll {
                excludedTaskIds.insert(t.id)
            }
        }
        widgetManager.setExcludedTaskIds(excludedTaskIds)
        widgetManager.setSelectedRecurringTaskIds(selectedRecurringIds)
        widgetManager.syncWithTaskList(tasks)
    }

    private func selectAllCustom() {
        for t in filteredCustomTasks {
            selectedCustomIds.insert(t.id)
            excludedTaskIds.remove(t.id)
        }
        widgetManager.setExcludedTaskIds(excludedTaskIds)
        widgetManager.setSelectedCustomTaskIds(selectedCustomIds)
        widgetManager.syncWithTaskList(tasks)
    }

    private func deselectAllCustom() {
        for t in filteredCustomTasks {
            selectedCustomIds.remove(t.id)
            if isAutoSelectAll {
                excludedTaskIds.insert(t.id)
            }
        }
        widgetManager.setExcludedTaskIds(excludedTaskIds)
        widgetManager.setSelectedCustomTaskIds(selectedCustomIds)
        widgetManager.syncWithTaskList(tasks)
    }

    private func priorityBadgeColor(_ priority: String) -> Color {
        switch priority.lowercased() {
        case "urgent": return AppColors.priorityUrgent
        case "high": return AppColors.priorityHigh
        case "medium": return AppColors.primary
        default: return .secondary
        }
    }
}
