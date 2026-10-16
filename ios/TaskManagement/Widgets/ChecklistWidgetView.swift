import SwiftUI
import AppIntents
#if canImport(WidgetKit)
import WidgetKit
#endif

// MARK: - Subtask Pagination Helper & Components
public struct SubtaskPaginationHelper {
    public static func paginate(task: TaskItem, pageSize: Int) -> (subtasks: [Subtask], page: Int, totalPages: Int, totalCount: Int) {
        let all = task.subtasks ?? []
        guard !all.isEmpty else { return ([], 0, 1, 0) }

        // Prioritize incomplete subtasks first, completed subtasks second (preserving order within each group)
        let incomplete = all.filter { !$0.isCompleted }
        let completed = all.filter { $0.isCompleted }
        let ordered = incomplete + completed

        let totalPages = max(1, Int(ceil(Double(ordered.count) / Double(pageSize))))
        let page = SharedWidgetDataManager.shared.getSubtaskPage(for: task.id, totalPages: totalPages)
        let startIndex = page * pageSize
        let endIndex = min(startIndex + pageSize, ordered.count)
        let sliced = Array(ordered[startIndex..<endIndex])

        return (sliced, page, totalPages, ordered.count)
    }
}

public struct SubtaskPaginationBar: View {
    public let taskId: Int64
    public let currentPage: Int
    public let totalPages: Int
    public let totalCount: Int

    public init(taskId: Int64, currentPage: Int, totalPages: Int, totalCount: Int) {
        self.taskId = taskId
        self.currentPage = currentPage
        self.totalPages = totalPages
        self.totalCount = totalCount
    }

    public var body: some View {
        HStack(spacing: 6) {
            Text(L10n.Widgets.subtasksCount(count: totalCount))
                .font(.system(size: 10, weight: .semibold))
                .foregroundStyle(.white.opacity(0.55))

            Spacer()

            HStack(spacing: 4) {
                Button(intent: CycleSubtaskPageIntent(taskId: taskId, delta: -1, totalPages: totalPages)) {
                    Image(systemName: "chevron.left")
                        .font(.system(size: 8, weight: .bold))
                        .foregroundStyle(currentPage > 0 ? .white : .white.opacity(0.3))
                        .frame(width: 20, height: 20)
                        .background(Color.white.opacity(currentPage > 0 ? 0.12 : 0.05))
                        .clipShape(Circle())
                }
                .buttonStyle(.plain)

                Text(L10n.Widgets.pageIndicator(page: currentPage + 1, total: totalPages))
                    .font(.system(size: 10, weight: .bold, design: .monospaced))
                    .foregroundStyle(.white.opacity(0.9))
                    .padding(.horizontal, 4)

                Button(intent: CycleSubtaskPageIntent(taskId: taskId, delta: 1, totalPages: totalPages)) {
                    Image(systemName: "chevron.right")
                        .font(.system(size: 8, weight: .bold))
                        .foregroundStyle(currentPage < totalPages - 1 ? .white : .white.opacity(0.3))
                        .frame(width: 20, height: 20)
                        .background(Color.white.opacity(currentPage < totalPages - 1 ? 0.12 : 0.05))
                        .clipShape(Circle())
                }
                .buttonStyle(.plain)
            }
            .padding(.horizontal, 6)
            .padding(.vertical, 2)
            .background(Color.white.opacity(0.06))
            .clipShape(Capsule())
        }
        .padding(.top, 2)
    }
}

public struct CompactSubtaskPill: View {
    public let taskId: Int64
    public let currentPage: Int
    public let totalPages: Int

    public init(taskId: Int64, currentPage: Int, totalPages: Int) {
        self.taskId = taskId
        self.currentPage = currentPage
        self.totalPages = totalPages
    }

    public var body: some View {
        Button(intent: CycleSubtaskPageIntent(taskId: taskId, delta: 1, totalPages: totalPages)) {
            HStack(spacing: 3) {
                Text(L10n.Widgets.pageIndicator(page: currentPage + 1, total: totalPages))
                    .font(.system(size: 9, weight: .bold, design: .monospaced))
                Image(systemName: "chevron.right")
                    .font(.system(size: 7, weight: .bold))
            }
            .foregroundStyle(.white.opacity(0.85))
            .padding(.horizontal, 6)
            .padding(.vertical, 2)
            .background(Color.white.opacity(0.12))
            .clipShape(Capsule())
        }
        .buttonStyle(.plain)
    }
}

// MARK: - 1. Dedicated Recurring Task Checklist Widget View
public struct RecurringTaskWidgetView: View {
    @Environment(\.widgetFamily) var family
    public let entry: TaskWidgetEntry

    public init(entry: TaskWidgetEntry) {
        self.entry = entry
    }

    private var allRecurringTasks: [TaskItem] {
        let combined = entry.bundle.focusTasks + entry.bundle.upcomingDeadlines
        var seen = Set<Int64>()
        var result: [TaskItem] = []
        let isAuto = SharedWidgetDataManager.shared.isAutoSelectAllEnabled()
        let selectedIds = SharedWidgetDataManager.shared.getSelectedRecurringTaskIds()
        let excluded = SharedWidgetDataManager.shared.getExcludedTaskIds()
        for t in combined {
            if t.isRecurring == true && t.status != "done" && !seen.contains(t.id) {
                let isSelected = isAuto ? !excluded.contains(t.id) : selectedIds.contains(t.id)
                if isSelected {
                    seen.insert(t.id)
                    result.append(t)
                }
            }
        }
        return result
    }

    private var activeRecurringTask: TaskItem? {
        guard !allRecurringTasks.isEmpty else { return nil }
        let idx = SharedWidgetDataManager.shared.getRecurringTaskIndex() % allRecurringTasks.count
        return allRecurringTasks[idx]
    }

    public var body: some View {
        switch family {
        case .systemSmall:
            smallView
        case .systemLarge:
            largeView
        default:
            mediumView
        }
    }

    // MARK: - Small Family

    private var smallView: some View {
        VStack(alignment: .leading, spacing: 6) {
            HStack(spacing: 4) {
                HStack(spacing: 3) {
                    Image(systemName: "repeat.circle.fill")
                        .font(.system(size: 11, weight: .bold))
                    Text(L10n.Widgets.recurringBadge)
                        .font(.system(size: 9, weight: .black))
                        .tracking(0.5)
                }
                .foregroundStyle(Color(red: 56/255, green: 189/255, blue: 248/255))
                .padding(.horizontal, 6)
                .padding(.vertical, 2)
                .background(Color(red: 56/255, green: 189/255, blue: 248/255).opacity(0.15))
                .clipShape(Capsule())

                Spacer()

                if allRecurringTasks.count > 1 {
                    Button(intent: CycleNextRecurringTaskIntent()) {
                        Image(systemName: "arrow.triangle.2.circlepath")
                            .font(.system(size: 10, weight: .bold))
                            .foregroundStyle(.white.opacity(0.7))
                            .padding(3)
                            .background(Color.white.opacity(0.1))
                            .clipShape(Circle())
                    }
                    .buttonStyle(.plain)
                }

                if let task = activeRecurringTask, let count = task.subtasksCount, count > 0 {
                    let done = task.subtasksCompleted ?? 0
                    Text("\(done)/\(count)")
                        .font(.system(size: 10, weight: .bold, design: .monospaced))
                        .foregroundStyle(done == count ? Color(red: 52/255, green: 211/255, blue: 153/255) : Color(red: 56/255, green: 189/255, blue: 248/255))
                        .padding(.horizontal, 5)
                        .padding(.vertical, 1.5)
                        .background(Color.white.opacity(0.08))
                        .clipShape(Capsule())
                }
            }

            if let task = activeRecurringTask {
                if task.status == "done" {
                    doneStateView(isCompact: true)
                } else {
                    HStack {
                        Text(task.title)
                            .font(.system(size: 12, weight: .bold))
                            .lineLimit(1)
                            .foregroundStyle(.white)

                        Spacer(minLength: 2)

                        let p = SubtaskPaginationHelper.paginate(task: task, pageSize: 2)
                        if p.totalPages > 1 {
                            CompactSubtaskPill(taskId: task.id, currentPage: p.page, totalPages: p.totalPages)
                        }
                    }

                    progressBar(for: task)

                    let p = SubtaskPaginationHelper.paginate(task: task, pageSize: 2)
                    let allChecked = !(task.subtasks ?? []).isEmpty && (task.subtasks ?? []).allSatisfy { $0.isCompleted }

                    if allChecked {
                        smallActionsBar(taskId: task.id)
                    } else if p.subtasks.isEmpty {
                        Text(L10n.Widgets.recurringNoSubtasks)
                            .font(.system(size: 10))
                            .foregroundStyle(.white.opacity(0.6))
                    } else {
                        VStack(alignment: .leading, spacing: 4) {
                            ForEach(p.subtasks) { st in
                                subtaskRow(taskId: task.id, subtask: st, isCompact: true)
                            }
                        }
                    }
                }
            } else {
                let isAuto = SharedWidgetDataManager.shared.isAutoSelectAllEnabled()
                let selectedIds = SharedWidgetDataManager.shared.getSelectedRecurringTaskIds()
                if !isAuto && selectedIds.isEmpty {
                    unselectedRecurringView(isCompact: true)
                } else if entry.bundle.completedToday > 0 {
                    doneStateView(isCompact: true)
                } else {
                    emptyRecurringView(isCompact: true)
                }
            }

            Spacer(minLength: 0)
        }
        .padding(12)
        .widgetBackground(accentColor: Color(red: 56/255, green: 189/255, blue: 248/255))
        .widgetURL(activeRecurringTask.flatMap { URL(string: "taskhub://tasks/\($0.id)") } ?? URL(string: "taskhub://settings/widget"))
    }

    // MARK: - Medium Family

    private var mediumView: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack(spacing: 8) {
                HStack(spacing: 4) {
                    Image(systemName: "repeat.circle.fill")
                        .font(.system(size: 12, weight: .bold))
                    Text(L10n.Widgets.recurringTitle)
                        .font(.system(size: 9, weight: .black))
                        .tracking(0.8)
                }
                .foregroundStyle(Color(red: 56/255, green: 189/255, blue: 248/255))
                .padding(.horizontal, 7)
                .padding(.vertical, 3)
                .background(Color(red: 56/255, green: 189/255, blue: 248/255).opacity(0.15))
                .clipShape(Capsule())

                Spacer()

                if allRecurringTasks.count > 1 {
                    Button(intent: CycleNextRecurringTaskIntent()) {
                        HStack(spacing: 4) {
                            Text(L10n.Widgets.recurringSwitch)
                                .font(.system(size: 10, weight: .semibold))
                            Image(systemName: "arrow.triangle.2.circlepath")
                                .font(.system(size: 9, weight: .bold))
                        }
                        .foregroundStyle(.white.opacity(0.8))
                        .padding(.horizontal, 7)
                        .padding(.vertical, 3)
                        .background(Color.white.opacity(0.08))
                        .clipShape(Capsule())
                    }
                    .buttonStyle(.plain)
                }

                if let task = activeRecurringTask, let count = task.subtasksCount, count > 0 {
                    let done = task.subtasksCompleted ?? 0
                    let p = SubtaskPaginationHelper.paginate(task: task, pageSize: 3)

                    HStack(spacing: 5) {
                        if p.totalPages > 1 {
                            CompactSubtaskPill(taskId: task.id, currentPage: p.page, totalPages: p.totalPages)
                        }

                        HStack(spacing: 4) {
                            Image(systemName: done == count ? "checkmark.seal.fill" : "circle.dashed")
                                .font(.system(size: 11))
                            Text(L10n.Widgets.recurringSubtasksDone(done: done, total: count))
                                .font(.system(size: 11, weight: .bold))
                        }
                        .foregroundStyle(done == count ? Color(red: 52/255, green: 211/255, blue: 153/255) : Color(red: 56/255, green: 189/255, blue: 248/255))
                        .padding(.horizontal, 7)
                        .padding(.vertical, 3)
                        .background(Color.white.opacity(0.08))
                        .clipShape(Capsule())
                    }
                }
            }

            if let task = activeRecurringTask {
                if task.status == "done" {
                    doneStateView(isCompact: false)
                } else {
                    HStack(alignment: .center) {
                        Text(task.title)
                            .font(.system(size: 13, weight: .bold))
                            .lineLimit(1)
                            .foregroundStyle(.white)

                        Spacer()

                        if let pattern = task.recurrencePattern {
                            Text(formatPattern(pattern))
                                .font(.system(size: 9, weight: .semibold))
                                .foregroundStyle(.white.opacity(0.6))
                                .padding(.horizontal, 5)
                                .padding(.vertical, 2)
                                .background(Color.white.opacity(0.06))
                                .clipShape(RoundedRectangle(cornerRadius: 4))
                        }
                    }

                    progressBar(for: task)

                    let p = SubtaskPaginationHelper.paginate(task: task, pageSize: 3)
                    let allChecked = !(task.subtasks ?? []).isEmpty && (task.subtasks ?? []).allSatisfy { $0.isCompleted }

                    if allChecked {
                        statusActionBar(taskId: task.id)
                    } else if p.subtasks.isEmpty {
                        Text(L10n.Widgets.recurringNoSubtasks)
                            .font(.system(size: 11))
                            .foregroundStyle(.white.opacity(0.6))
                            .padding(.vertical, 4)
                    } else {
                        VStack(alignment: .leading, spacing: 5) {
                            ForEach(p.subtasks) { st in
                                subtaskRow(taskId: task.id, subtask: st, isCompact: false)
                            }
                        }
                    }
                }
            } else {
                let isAuto = SharedWidgetDataManager.shared.isAutoSelectAllEnabled()
                let selectedIds = SharedWidgetDataManager.shared.getSelectedRecurringTaskIds()
                if !isAuto && selectedIds.isEmpty {
                    unselectedRecurringView(isCompact: false)
                } else if entry.bundle.completedToday > 0 {
                    doneStateView(isCompact: false)
                } else {
                    emptyRecurringView(isCompact: false)
                }
            }

            Spacer(minLength: 0)
        }
        .padding(14)
        .widgetBackground(accentColor: Color(red: 56/255, green: 189/255, blue: 248/255))
        .widgetURL(activeRecurringTask.flatMap { URL(string: "taskhub://tasks/\($0.id)") } ?? URL(string: "taskhub://settings/widget"))
    }

    // MARK: - Large Family

    private var largeView: some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack(spacing: 8) {
                HStack(spacing: 4) {
                    Image(systemName: "repeat.circle.fill")
                        .font(.system(size: 14, weight: .bold))
                    Text(L10n.Widgets.recurringTitle)
                        .font(.system(size: 11, weight: .black))
                        .tracking(1)
                }
                .foregroundStyle(Color(red: 56/255, green: 189/255, blue: 248/255))
                .padding(.horizontal, 8)
                .padding(.vertical, 4)
                .background(Color(red: 56/255, green: 189/255, blue: 248/255).opacity(0.15))
                .clipShape(Capsule())

                Spacer()

                if allRecurringTasks.count > 1 {
                    Button(intent: CycleNextRecurringTaskIntent()) {
                        HStack(spacing: 4) {
                            Text(L10n.Widgets.recurringSwitch)
                                .font(.system(size: 11, weight: .semibold))
                            Image(systemName: "arrow.triangle.2.circlepath")
                                .font(.system(size: 10, weight: .bold))
                        }
                        .foregroundStyle(.white.opacity(0.9))
                        .padding(.horizontal, 8)
                        .padding(.vertical, 4)
                        .background(Color.white.opacity(0.1))
                        .clipShape(Capsule())
                    }
                    .buttonStyle(.plain)
                }

                if let task = activeRecurringTask, let count = task.subtasksCount, count > 0 {
                    let done = task.subtasksCompleted ?? 0
                    Text(L10n.Widgets.recurringSubtasksDone(done: done, total: count))
                        .font(.system(size: 12, weight: .bold))
                        .foregroundStyle(Color(red: 52/255, green: 211/255, blue: 153/255))
                }
            }

            if let task = activeRecurringTask {
                if task.status == "done" {
                    doneStateView(isCompact: false)
                } else {
                    Text(task.title)
                        .font(.system(size: 15, weight: .bold))
                        .lineLimit(1)
                        .foregroundStyle(.white)

                    progressBar(for: task)

                    let p = SubtaskPaginationHelper.paginate(task: task, pageSize: 5)
                    let allChecked = !(task.subtasks ?? []).isEmpty && (task.subtasks ?? []).allSatisfy { $0.isCompleted }

                    VStack(alignment: .leading, spacing: 6) {
                        ForEach(p.subtasks) { st in
                            subtaskRow(taskId: task.id, subtask: st, isCompact: false)
                        }

                        if p.totalPages > 1 {
                            SubtaskPaginationBar(taskId: task.id, currentPage: p.page, totalPages: p.totalPages, totalCount: p.totalCount)
                        }
                    }

                    if allChecked {
                        statusActionBar(taskId: task.id)
                    }
                }
            } else {
                let isAuto = SharedWidgetDataManager.shared.isAutoSelectAllEnabled()
                let selectedIds = SharedWidgetDataManager.shared.getSelectedRecurringTaskIds()
                if !isAuto && selectedIds.isEmpty {
                    unselectedRecurringView(isCompact: false)
                } else if entry.bundle.completedToday > 0 {
                    doneStateView(isCompact: false)
                } else {
                    emptyRecurringView(isCompact: false)
                }
            }

            Spacer(minLength: 0)
        }
        .padding(16)
        .widgetBackground(accentColor: Color(red: 56/255, green: 189/255, blue: 248/255))
        .widgetURL(activeRecurringTask.flatMap { URL(string: "taskhub://tasks/\($0.id)") } ?? URL(string: "taskhub://settings/widget"))
    }

    // MARK: - Subcomponents

    @ViewBuilder
    private func progressBar(for task: TaskItem) -> some View {
        let total = max(task.subtasksCount ?? 0, 1)
        let done = min(task.subtasksCompleted ?? 0, total)
        let ratio = CGFloat(done) / CGFloat(total)

        GeometryReader { geo in
            ZStack(alignment: .leading) {
                Capsule()
                    .fill(Color.white.opacity(0.1))
                    .frame(height: 4)

                Capsule()
                    .fill(
                        LinearGradient(
                            colors: [
                                Color(red: 56/255, green: 189/255, blue: 248/255),
                                Color(red: 52/255, green: 211/255, blue: 153/255)
                            ],
                            startPoint: .leading,
                            endPoint: .trailing
                        )
                    )
                    .frame(width: max(geo.size.width * ratio, 6), height: 4)
            }
        }
        .frame(height: 4)
    }

    @ViewBuilder
    private func subtaskRow(taskId: Int64, subtask: Subtask, isCompact: Bool) -> some View {
        HStack(spacing: 8) {
            Button(intent: ToggleSubtaskIntent(taskId: taskId, subtaskId: subtask.id, isCompleted: !subtask.isCompleted)) {
                Image(systemName: subtask.isCompleted ? "checkmark.circle.fill" : "circle")
                    .font(.system(size: isCompact ? 13 : 15, weight: .semibold))
                    .foregroundStyle(subtask.isCompleted ? Color(red: 52/255, green: 211/255, blue: 153/255) : .white.opacity(0.45))
            }
            .buttonStyle(.plain)

            Text(subtask.title)
                .font(.system(size: isCompact ? 11 : 12, weight: subtask.isCompleted ? .regular : .medium))
                .lineLimit(2)
                .strikethrough(subtask.isCompleted, color: .white.opacity(0.35))
                .foregroundStyle(subtask.isCompleted ? .white.opacity(0.4) : .white)
                .fixedSize(horizontal: false, vertical: true)

            Spacer()
        }
        .padding(.horizontal, 8)
        .padding(.vertical, isCompact ? 4 : 5)
        .background(Color.white.opacity(0.04))
        .clipShape(RoundedRectangle(cornerRadius: 7, style: .continuous))
    }

    @ViewBuilder
    private func statusActionBar(taskId: Int64) -> some View {
        VStack(spacing: 6) {
            HStack(spacing: 4) {
                Image(systemName: "sparkles")
                    .font(.system(size: 10, weight: .bold))
                Text(L10n.Widgets.recurringAllSubtasksComplete)
                    .font(.system(size: 10, weight: .semibold))
            }
            .foregroundStyle(Color(red: 52/255, green: 211/255, blue: 153/255))

            HStack(spacing: 8) {
                Button(intent: UpdateTaskStatusIntent(taskId: taskId, newStatus: "done")) {
                    HStack(spacing: 4) {
                        Image(systemName: "checkmark.circle.fill")
                            .font(.system(size: 11, weight: .bold))
                        Text(L10n.Widgets.recurringActionDone)
                            .font(.system(size: 11, weight: .bold))
                    }
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 5)
                    .background(
                        LinearGradient(
                            colors: [Color(red: 16/255, green: 185/255, blue: 129/255), Color(red: 5/255, green: 150/255, blue: 105/255)],
                            startPoint: .topLeading,
                            endPoint: .bottomTrailing
                        )
                    )
                    .foregroundStyle(.white)
                    .clipShape(RoundedRectangle(cornerRadius: 8, style: .continuous))
                }
                .buttonStyle(.plain)

                Button(intent: UpdateTaskStatusIntent(taskId: taskId, newStatus: "in_review")) {
                    HStack(spacing: 4) {
                        Image(systemName: "eye.fill")
                            .font(.system(size: 11, weight: .bold))
                        Text(L10n.Widgets.recurringActionReview)
                            .font(.system(size: 11, weight: .bold))
                    }
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 5)
                    .background(Color.white.opacity(0.12))
                    .foregroundStyle(.white)
                    .clipShape(RoundedRectangle(cornerRadius: 8, style: .continuous))
                }
                .buttonStyle(.plain)
            }
        }
        .padding(8)
        .background(Color(red: 16/255, green: 185/255, blue: 129/255).opacity(0.12))
        .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))
    }

    @ViewBuilder
    private func smallActionsBar(taskId: Int64) -> some View {
        Button(intent: UpdateTaskStatusIntent(taskId: taskId, newStatus: "done")) {
            HStack(spacing: 4) {
                Image(systemName: "checkmark.circle.fill")
                    .font(.system(size: 10, weight: .bold))
                Text(L10n.Widgets.recurringActionDoneShort)
                    .font(.system(size: 10, weight: .bold))
            }
            .frame(maxWidth: .infinity)
            .padding(.vertical, 4)
            .background(Color(red: 16/255, green: 185/255, blue: 129/255))
            .foregroundStyle(.white)
            .clipShape(RoundedRectangle(cornerRadius: 6))
        }
        .buttonStyle(.plain)
    }

    @ViewBuilder
    private func doneStateView(isCompact: Bool) -> some View {
        VStack(spacing: isCompact ? 4 : 6) {
            Image(systemName: "checkmark.seal.fill")
                .font(.system(size: isCompact ? 22 : 28))
                .foregroundStyle(
                    LinearGradient(
                        colors: [Color(red: 52/255, green: 211/255, blue: 153/255), Color(red: 16/255, green: 185/255, blue: 129/255)],
                        startPoint: .topLeading,
                        endPoint: .bottomTrailing
                    )
                )

            Text(L10n.Widgets.recurringDoneTitle)
                .font(.system(size: isCompact ? 12 : 13, weight: .bold))
                .foregroundStyle(.white)

            Text(L10n.Widgets.recurringDoneDesc)
                .font(.system(size: isCompact ? 9 : 10))
                .foregroundStyle(.white.opacity(0.6))
                .multilineTextAlignment(.center)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }

    @ViewBuilder
    private func emptyRecurringView(isCompact: Bool) -> some View {
        VStack(spacing: isCompact ? 4 : 6) {
            Image(systemName: "repeat")
                .font(.system(size: isCompact ? 20 : 24))
                .foregroundStyle(Color(red: 56/255, green: 189/255, blue: 248/255).opacity(0.6))

            Text(L10n.Widgets.recurringEmptyTitle)
                .font(.system(size: isCompact ? 11 : 12, weight: .semibold))
                .foregroundStyle(.white.opacity(0.8))

            if !isCompact {
                Text(L10n.Widgets.recurringEmptyDesc)
                    .font(.system(size: 9))
                    .foregroundStyle(.white.opacity(0.5))
                    .multilineTextAlignment(.center)
            }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }

    @ViewBuilder
    private func unselectedRecurringView(isCompact: Bool) -> some View {
        VStack(spacing: isCompact ? 5 : 7) {
            Image(systemName: "repeat.circle")
                .font(.system(size: isCompact ? 22 : 28))
                .foregroundStyle(Color(red: 56/255, green: 189/255, blue: 248/255))

            Text(L10n.Widgets.recurringUnselectedTitle)
                .font(.system(size: isCompact ? 11 : 13, weight: .bold))
                .foregroundStyle(.white)
                .multilineTextAlignment(.center)

            Text(L10n.Widgets.recurringUnselectedDesc)
                .font(.system(size: isCompact ? 9 : 10))
                .foregroundStyle(.white.opacity(0.65))
                .multilineTextAlignment(.center)
                .padding(.horizontal, 4)

            if !isCompact {
                HStack(spacing: 4) {
                    Image(systemName: "slider.horizontal.3")
                        .font(.system(size: 10, weight: .semibold))
                    Text(L10n.Widgets.setupAction)
                        .font(.system(size: 11, weight: .bold))
                }
                .foregroundStyle(.white)
                .padding(.horizontal, 10)
                .padding(.vertical, 5)
                .background(Color(red: 56/255, green: 189/255, blue: 248/255).opacity(0.25))
                .clipShape(Capsule())
            }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }

    private func formatPattern(_ p: String) -> String {
        switch p.lowercased() {
        case "daily": return "pattern.daily".localized
        case "weekly": return "pattern.weekly".localized
        case "monthly": return "pattern.monthly".localized
        default: return p.capitalized
        }
    }
}

// MARK: - 2. Dedicated Custom Selected Task Checklist Widget View
public struct CustomTaskWidgetView: View {
    @Environment(\.widgetFamily) var family
    public let entry: TaskWidgetEntry

    public init(entry: TaskWidgetEntry) {
        self.entry = entry
    }

    private var activeTasks: [TaskItem] {
        let combined = entry.bundle.focusTasks + entry.bundle.upcomingDeadlines
        var seen = Set<Int64>()
        var result: [TaskItem] = []
        let isAuto = SharedWidgetDataManager.shared.isAutoSelectAllEnabled()
        let selectedIds = SharedWidgetDataManager.shared.getSelectedCustomTaskIds()
        let excluded = SharedWidgetDataManager.shared.getExcludedTaskIds()
        for t in combined {
            if !seen.contains(t.id) && t.status != "done" {
                let isSelected = isAuto ? !excluded.contains(t.id) : selectedIds.contains(t.id)
                if isSelected {
                    seen.insert(t.id)
                    result.append(t)
                }
            }
        }
        return result
    }

    private var currentTask: TaskItem? {
        guard !activeTasks.isEmpty else { return nil }
        let idx = SharedWidgetDataManager.shared.getCustomTaskIndex() % activeTasks.count
        return activeTasks[idx]
    }

    public var body: some View {
        switch family {
        case .systemSmall:
            smallView
        case .systemLarge:
            largeView
        default:
            mediumView
        }
    }

    // MARK: - Small Family

    private var smallView: some View {
        VStack(alignment: .leading, spacing: 6) {
            HStack(spacing: 4) {
                HStack(spacing: 3) {
                    Image(systemName: "checklist.checked")
                        .font(.system(size: 11, weight: .bold))
                    Text(L10n.Widgets.customBadge)
                        .font(.system(size: 9, weight: .black))
                        .tracking(0.5)
                }
                .foregroundStyle(Color(red: 168/255, green: 85/255, blue: 247/255))
                .padding(.horizontal, 6)
                .padding(.vertical, 2)
                .background(Color(red: 168/255, green: 85/255, blue: 247/255).opacity(0.15))
                .clipShape(Capsule())

                Spacer()

                if activeTasks.count > 1 {
                    Button(intent: CycleNextTaskIntent()) {
                        Image(systemName: "arrow.triangle.2.circlepath")
                            .font(.system(size: 10, weight: .bold))
                            .foregroundStyle(.white.opacity(0.7))
                            .padding(3)
                            .background(Color.white.opacity(0.1))
                            .clipShape(Circle())
                    }
                    .buttonStyle(.plain)
                }
            }

            if let task = currentTask {
                HStack {
                    Text(task.title)
                        .font(.system(size: 12, weight: .bold))
                        .lineLimit(1)
                        .foregroundStyle(.white)

                    Spacer(minLength: 2)

                    let p = SubtaskPaginationHelper.paginate(task: task, pageSize: 2)
                    if p.totalPages > 1 {
                        CompactSubtaskPill(taskId: task.id, currentPage: p.page, totalPages: p.totalPages)
                    }
                }

                let p = SubtaskPaginationHelper.paginate(task: task, pageSize: 2)
                let allChecked = !(task.subtasks ?? []).isEmpty && (task.subtasks ?? []).allSatisfy { $0.isCompleted }

                if allChecked {
                    Button(intent: UpdateTaskStatusIntent(taskId: task.id, newStatus: "done")) {
                        HStack(spacing: 4) {
                            Image(systemName: "checkmark.circle.fill")
                                .font(.system(size: 10, weight: .bold))
                            Text(L10n.Widgets.recurringActionDone)
                                .font(.system(size: 10, weight: .bold))
                        }
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 4)
                        .background(Color(red: 16/255, green: 185/255, blue: 129/255))
                        .foregroundStyle(.white)
                        .clipShape(RoundedRectangle(cornerRadius: 6))
                    }
                    .buttonStyle(.plain)
                } else if p.subtasks.isEmpty {
                    Text(L10n.Widgets.customNoSubtasks)
                        .font(.system(size: 10))
                        .foregroundStyle(.white.opacity(0.6))
                } else {
                    VStack(alignment: .leading, spacing: 4) {
                        ForEach(p.subtasks) { st in
                            subtaskRow(taskId: task.id, subtask: st, isCompact: true)
                        }
                    }
                }
            } else {
                let isAuto = SharedWidgetDataManager.shared.isAutoSelectAllEnabled()
                let selectedIds = SharedWidgetDataManager.shared.getSelectedCustomTaskIds()
                if !isAuto && selectedIds.isEmpty {
                    unselectedCustomView(isCompact: true)
                } else {
                    emptyTasksView(isCompact: true)
                }
            }

            Spacer(minLength: 0)
        }
        .padding(12)
        .widgetBackground(accentColor: Color(red: 168/255, green: 85/255, blue: 247/255))
        .widgetURL(currentTask.flatMap { URL(string: "taskhub://tasks/\($0.id)") } ?? URL(string: "taskhub://settings/widget"))
    }

    // MARK: - Medium Family

    private var mediumView: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack(spacing: 8) {
                HStack(spacing: 4) {
                    Image(systemName: "checklist.checked")
                        .font(.system(size: 12, weight: .bold))
                    Text(L10n.Widgets.customTitle)
                        .font(.system(size: 9, weight: .black))
                        .tracking(0.8)
                }
                .foregroundStyle(Color(red: 168/255, green: 85/255, blue: 247/255))
                .padding(.horizontal, 7)
                .padding(.vertical, 3)
                .background(Color(red: 168/255, green: 85/255, blue: 247/255).opacity(0.15))
                .clipShape(Capsule())

                Spacer()

                if activeTasks.count > 1 {
                    Button(intent: CycleNextTaskIntent()) {
                        HStack(spacing: 4) {
                            Text(L10n.Widgets.customSwitch)
                                .font(.system(size: 10, weight: .semibold))
                            Image(systemName: "arrow.triangle.2.circlepath")
                                .font(.system(size: 9, weight: .bold))
                        }
                        .foregroundStyle(.white.opacity(0.8))
                        .padding(.horizontal, 7)
                        .padding(.vertical, 3)
                        .background(Color.white.opacity(0.08))
                        .clipShape(Capsule())
                    }
                    .buttonStyle(.plain)
                }
            }

            if let task = currentTask {
                HStack {
                    Text(task.title)
                        .font(.system(size: 13, weight: .bold))
                        .lineLimit(1)
                        .foregroundStyle(.white)

                    Spacer()

                    let p = SubtaskPaginationHelper.paginate(task: task, pageSize: 3)
                    if p.totalPages > 1 {
                        CompactSubtaskPill(taskId: task.id, currentPage: p.page, totalPages: p.totalPages)
                    }

                    let pri = task.priority
                    Text(pri.uppercased())
                        .font(.system(size: 8, weight: .black))
                        .foregroundStyle(priorityColor(pri))
                        .padding(.horizontal, 5)
                        .padding(.vertical, 2)
                        .background(priorityColor(pri).opacity(0.15))
                        .clipShape(RoundedRectangle(cornerRadius: 4))
                }

                let p = SubtaskPaginationHelper.paginate(task: task, pageSize: 3)
                let allChecked = !(task.subtasks ?? []).isEmpty && (task.subtasks ?? []).allSatisfy { $0.isCompleted }

                if allChecked {
                    HStack(spacing: 8) {
                        Button(intent: UpdateTaskStatusIntent(taskId: task.id, newStatus: "done")) {
                            HStack(spacing: 4) {
                                Image(systemName: "checkmark.circle.fill")
                                    .font(.system(size: 11, weight: .bold))
                                Text(L10n.Widgets.customCompleteTask)
                                    .font(.system(size: 11, weight: .bold))
                            }
                            .frame(maxWidth: .infinity)
                            .padding(.vertical, 6)
                            .background(Color(red: 16/255, green: 185/255, blue: 129/255))
                            .foregroundStyle(.white)
                            .clipShape(RoundedRectangle(cornerRadius: 8))
                        }
                        .buttonStyle(.plain)

                        if activeTasks.count > 1 {
                            Button(intent: CycleNextTaskIntent()) {
                                Text(L10n.Widgets.customOtherTask)
                                    .font(.system(size: 11, weight: .semibold))
                                    .padding(.horizontal, 10)
                                    .padding(.vertical, 6)
                                    .background(Color.white.opacity(0.1))
                                    .foregroundStyle(.white)
                                    .clipShape(RoundedRectangle(cornerRadius: 8))
                            }
                            .buttonStyle(.plain)
                        }
                    }
                    .padding(8)
                    .background(Color(red: 16/255, green: 185/255, blue: 129/255).opacity(0.12))
                    .clipShape(RoundedRectangle(cornerRadius: 10))
                } else if p.subtasks.isEmpty {
                    Text(L10n.Widgets.customNoSubtasks)
                        .font(.system(size: 11))
                        .foregroundStyle(.white.opacity(0.6))
                        .padding(.vertical, 4)
                } else {
                    VStack(alignment: .leading, spacing: 5) {
                        ForEach(p.subtasks) { st in
                            subtaskRow(taskId: task.id, subtask: st, isCompact: false)
                        }
                    }
                }
            } else {
                let isAuto = SharedWidgetDataManager.shared.isAutoSelectAllEnabled()
                let selectedIds = SharedWidgetDataManager.shared.getSelectedCustomTaskIds()
                if !isAuto && selectedIds.isEmpty {
                    unselectedCustomView(isCompact: false)
                } else {
                    emptyTasksView(isCompact: false)
                }
            }

            Spacer(minLength: 0)
        }
        .padding(14)
        .widgetBackground(accentColor: Color(red: 168/255, green: 85/255, blue: 247/255))
        .widgetURL(currentTask.flatMap { URL(string: "taskhub://tasks/\($0.id)") } ?? URL(string: "taskhub://settings/widget"))
    }

    // MARK: - Large Family

    private var largeView: some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack(spacing: 8) {
                HStack(spacing: 4) {
                    Image(systemName: "checklist.checked")
                        .font(.system(size: 14, weight: .bold))
                    Text(L10n.Widgets.customTitleLarge)
                        .font(.system(size: 11, weight: .black))
                        .tracking(1)
                }
                .foregroundStyle(Color(red: 168/255, green: 85/255, blue: 247/255))
                .padding(.horizontal, 8)
                .padding(.vertical, 4)
                .background(Color(red: 168/255, green: 85/255, blue: 247/255).opacity(0.15))
                .clipShape(Capsule())

                Spacer()

                if activeTasks.count > 1 {
                    Button(intent: CycleNextTaskIntent()) {
                        HStack(spacing: 4) {
                            Text(L10n.Widgets.customOtherTasks)
                                .font(.system(size: 11, weight: .semibold))
                            Image(systemName: "arrow.triangle.2.circlepath")
                                .font(.system(size: 10, weight: .bold))
                        }
                        .foregroundStyle(.white.opacity(0.9))
                        .padding(.horizontal, 8)
                        .padding(.vertical, 4)
                        .background(Color.white.opacity(0.1))
                        .clipShape(Capsule())
                    }
                    .buttonStyle(.plain)
                }
            }

            if let task = currentTask {
                Text(task.title)
                    .font(.system(size: 15, weight: .bold))
                    .lineLimit(1)
                    .foregroundStyle(.white)

                let p = SubtaskPaginationHelper.paginate(task: task, pageSize: 5)
                let allChecked = !(task.subtasks ?? []).isEmpty && (task.subtasks ?? []).allSatisfy { $0.isCompleted }

                VStack(alignment: .leading, spacing: 6) {
                    ForEach(p.subtasks) { st in
                        subtaskRow(taskId: task.id, subtask: st, isCompact: false)
                    }

                    if p.totalPages > 1 {
                        SubtaskPaginationBar(taskId: task.id, currentPage: p.page, totalPages: p.totalPages, totalCount: p.totalCount)
                    }
                }

                if allChecked {
                    Button(intent: UpdateTaskStatusIntent(taskId: task.id, newStatus: "done")) {
                        HStack(spacing: 4) {
                            Image(systemName: "checkmark.circle.fill")
                                .font(.system(size: 12, weight: .bold))
                            Text(L10n.Widgets.customCompleteNow)
                                .font(.system(size: 12, weight: .bold))
                        }
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 7)
                        .background(Color(red: 16/255, green: 185/255, blue: 129/255))
                        .foregroundStyle(.white)
                        .clipShape(RoundedRectangle(cornerRadius: 8))
                    }
                    .buttonStyle(.plain)
                }
            } else {
                let isAuto = SharedWidgetDataManager.shared.isAutoSelectAllEnabled()
                let selectedIds = SharedWidgetDataManager.shared.getSelectedCustomTaskIds()
                if !isAuto && selectedIds.isEmpty {
                    unselectedCustomView(isCompact: false)
                } else {
                    emptyTasksView(isCompact: false)
                }
            }

            Spacer(minLength: 0)
        }
        .padding(16)
        .widgetBackground(accentColor: Color(red: 168/255, green: 85/255, blue: 247/255))
        .widgetURL(currentTask.flatMap { URL(string: "taskhub://tasks/\($0.id)") } ?? URL(string: "taskhub://settings/widget"))
    }

    // MARK: - Row & States

    @ViewBuilder
    private func subtaskRow(taskId: Int64, subtask: Subtask, isCompact: Bool) -> some View {
        HStack(spacing: 8) {
            Button(intent: ToggleSubtaskIntent(taskId: taskId, subtaskId: subtask.id, isCompleted: !subtask.isCompleted)) {
                Image(systemName: subtask.isCompleted ? "checkmark.circle.fill" : "circle")
                    .font(.system(size: isCompact ? 13 : 15, weight: .semibold))
                    .foregroundStyle(subtask.isCompleted ? Color(red: 52/255, green: 211/255, blue: 153/255) : .white.opacity(0.45))
            }
            .buttonStyle(.plain)

            Text(subtask.title)
                .font(.system(size: isCompact ? 11 : 12, weight: subtask.isCompleted ? .regular : .medium))
                .lineLimit(2)
                .strikethrough(subtask.isCompleted, color: .white.opacity(0.35))
                .foregroundStyle(subtask.isCompleted ? .white.opacity(0.4) : .white)
                .fixedSize(horizontal: false, vertical: true)

            Spacer()
        }
        .padding(.horizontal, 8)
        .padding(.vertical, isCompact ? 4 : 5)
        .background(Color.white.opacity(0.04))
        .clipShape(RoundedRectangle(cornerRadius: 7, style: .continuous))
    }

    @ViewBuilder
    private func emptyTasksView(isCompact: Bool) -> some View {
        VStack(spacing: isCompact ? 4 : 6) {
            Image(systemName: "checkmark.seal.fill")
                .font(.system(size: isCompact ? 20 : 26))
                .foregroundStyle(Color(red: 52/255, green: 211/255, blue: 153/255))

            Text(L10n.Widgets.customAllTasksDone)
                .font(.system(size: isCompact ? 11 : 12, weight: .semibold))
                .foregroundStyle(.white)

            if !isCompact {
                Text(L10n.Widgets.customAddNewTask)
                    .font(.system(size: 9))
                    .foregroundStyle(.white.opacity(0.5))
            }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }

    @ViewBuilder
    private func unselectedCustomView(isCompact: Bool) -> some View {
        VStack(spacing: isCompact ? 5 : 7) {
            Image(systemName: "checklist.checked")
                .font(.system(size: isCompact ? 22 : 28))
                .foregroundStyle(Color(red: 168/255, green: 85/255, blue: 247/255))

            Text(L10n.Widgets.customUnselectedTitle)
                .font(.system(size: isCompact ? 11 : 13, weight: .bold))
                .foregroundStyle(.white)
                .multilineTextAlignment(.center)

            Text(L10n.Widgets.customUnselectedDesc)
                .font(.system(size: isCompact ? 9 : 10))
                .foregroundStyle(.white.opacity(0.65))
                .multilineTextAlignment(.center)
                .padding(.horizontal, 4)

            if !isCompact {
                HStack(spacing: 4) {
                    Image(systemName: "slider.horizontal.3")
                        .font(.system(size: 10, weight: .semibold))
                    Text(L10n.Widgets.setupAction)
                        .font(.system(size: 11, weight: .bold))
                }
                .foregroundStyle(.white)
                .padding(.horizontal, 10)
                .padding(.vertical, 5)
                .background(Color(red: 168/255, green: 85/255, blue: 247/255).opacity(0.25))
                .clipShape(Capsule())
            }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }

    private func priorityColor(_ p: String) -> Color {
        switch p.lowercased() {
        case "urgent": return Color(red: 239/255, green: 68/255, blue: 68/255)
        case "high": return Color(red: 245/255, green: 158/255, blue: 11/255)
        case "medium": return Color(red: 59/255, green: 130/255, blue: 246/255)
        default: return Color(red: 148/255, green: 163/255, blue: 184/255)
        }
    }
}

// MARK: - 3. Legacy ChecklistWidgetView Alias (Compatibility)
public struct ChecklistWidgetView: View {
    public let entry: TaskWidgetEntry
    public init(entry: TaskWidgetEntry) {
        self.entry = entry
    }
    public var body: some View {
        RecurringTaskWidgetView(entry: entry)
    }
}

// MARK: - Premium Glassmorphic Widget Background
public extension View {
    func widgetBackground(accentColor: Color = Color(red: 56/255, green: 189/255, blue: 248/255)) -> some View {
        let content = ZStack {
            LinearGradient(
                colors: [
                    Color(red: 11/255, green: 17/255, blue: 32/255),
                    Color(red: 15/255, green: 23/255, blue: 42/255),
                    Color(red: 8/255, green: 12/255, blue: 22/255)
                ],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
            )

            GeometryReader { geo in
                Circle()
                    .fill(accentColor.opacity(0.12))
                    .blur(radius: 28)
                    .frame(width: geo.size.width * 0.7, height: geo.size.width * 0.7)
                    .position(x: geo.size.width * 0.85, y: geo.size.height * 0.15)
            }
        }

        #if canImport(WidgetKit)
        if #available(iOSApplicationExtension 17.0, iOS 17.0, *) {
            return self.containerBackground(for: .widget) {
                content
            }
        }
        #endif

        return self.background(content)
    }
}
