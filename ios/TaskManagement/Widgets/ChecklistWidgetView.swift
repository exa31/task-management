import SwiftUI
import AppIntents
#if canImport(WidgetKit)
import WidgetKit
#endif

public struct ChecklistWidgetView: View {
    @Environment(\.widgetFamily) var family
    public let entry: TaskWidgetEntry

    public init(entry: TaskWidgetEntry) {
        self.entry = entry
    }

    private var activeTask: TaskItem? {
        entry.bundle.focusTasks.first { $0.status != "done" } ?? entry.bundle.focusTasks.first
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
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                BrandLogoView(size: 20)
                Spacer()
                if let task = activeTask {
                    Text("\(task.subtasksCompleted ?? 0)/\(task.subtasksCount ?? 0)")
                        .font(.system(size: 11, weight: .bold))
                        .foregroundStyle(AppColors.primary)
                        .padding(.horizontal, 6)
                        .padding(.vertical, 2)
                        .background(AppColors.primary.opacity(0.15))
                        .clipShape(Capsule())
                }
            }

            if let task = activeTask {
                Text(task.title)
                    .font(.system(size: 13, weight: .bold))
                    .lineLimit(2)
                    .foregroundStyle(.primary)

                Divider().opacity(0.3)

                let subtasks = (task.subtasks ?? []).prefix(2)
                if subtasks.isEmpty {
                    Text(L10n.Widgets.allSubtasksDone)
                        .font(.system(size: 11))
                        .foregroundStyle(.secondary)
                        .italic()
                } else {
                    VStack(alignment: .leading, spacing: 6) {
                        ForEach(subtasks) { st in
                            subtaskRow(taskId: task.id, subtask: st, isCompact: true)
                        }
                    }
                }
            } else {
                emptyView
            }
            Spacer(minLength: 0)
        }
        .padding(12)
        .widgetBackground()
    }

    // MARK: - Medium Family

    private var mediumView: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack(spacing: 8) {
                BrandLogoView(size: 22)
                VStack(alignment: .leading, spacing: 1) {
                    Text(L10n.Widgets.focusToday)
                        .font(.system(size: 9, weight: .black))
                        .tracking(1)
                        .foregroundStyle(AppColors.primary)
                    if let task = activeTask {
                        Text(task.title)
                            .font(.system(size: 14, weight: .bold))
                            .lineLimit(1)
                            .foregroundStyle(.primary)
                    }
                }
                Spacer()
                if let task = activeTask, let count = task.subtasksCount, count > 0 {
                    let done = task.subtasksCompleted ?? 0
                    HStack(spacing: 4) {
                        Image(systemName: done == count ? "checkmark.circle.fill" : "circle.dashed")
                            .font(.system(size: 12))
                        Text("\(done)/\(count)")
                            .font(.system(size: 12, weight: .bold))
                    }
                    .foregroundStyle(done == count ? AppColors.statusDone : AppColors.primary)
                    .padding(.horizontal, 8)
                    .padding(.vertical, 3)
                    .background((done == count ? AppColors.statusDone : AppColors.primary).opacity(0.15))
                    .clipShape(Capsule())
                }
            }

            Divider().opacity(0.3)

            if let task = activeTask {
                let subtasks = (task.subtasks ?? []).prefix(3)
                if subtasks.isEmpty {
                    Text(L10n.Widgets.noActiveSubtasks)
                        .font(.system(size: 12))
                        .foregroundStyle(.secondary)
                } else {
                    VStack(alignment: .leading, spacing: 6) {
                        ForEach(subtasks) { st in
                            subtaskRow(taskId: task.id, subtask: st, isCompact: false)
                        }
                    }
                }
            } else {
                emptyView
            }

            Spacer(minLength: 0)
        }
        .padding(14)
        .widgetBackground()
    }

    // MARK: - Large Family

    private var largeView: some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack {
                BrandLogoView(size: 24)
                Text(L10n.Widgets.interactiveTitle)
                    .font(.system(size: 15, weight: .bold))
                    .foregroundStyle(.primary)
                Spacer()
                Text(L10n.Widgets.completedCount(done: entry.bundle.completedToday, total: entry.bundle.totalToday))
                    .font(.system(size: 11, weight: .semibold))
                    .foregroundStyle(AppColors.statusDone)
            }

            Divider().opacity(0.3)

            if entry.bundle.focusTasks.isEmpty {
                emptyView
            } else {
                ForEach(entry.bundle.focusTasks.prefix(2)) { task in
                    VStack(alignment: .leading, spacing: 6) {
                        HStack {
                            Text(task.title)
                                .font(.system(size: 13, weight: .bold))
                                .lineLimit(1)
                            Spacer()
                            if task.isRecurring == true {
                                Image(systemName: "repeat")
                                    .font(.system(size: 10, weight: .bold))
                                    .foregroundStyle(AppColors.sky)
                            }
                        }

                        let subtasks = (task.subtasks ?? []).prefix(3)
                        ForEach(subtasks) { st in
                            subtaskRow(taskId: task.id, subtask: st, isCompact: false)
                        }
                    }
                    .padding(8)
                    .background(Color.primary.opacity(0.04))
                    .clipShape(RoundedRectangle(cornerRadius: 8, style: .continuous))
                }
            }

            Spacer(minLength: 0)
        }
        .padding(14)
        .widgetBackground()
    }

    // MARK: - Interactive Subtask Row

    @ViewBuilder
    private func subtaskRow(taskId: Int64, subtask: Subtask, isCompact: Bool) -> some View {
        HStack(spacing: 8) {
            Button(intent: ToggleSubtaskIntent(taskId: taskId, subtaskId: subtask.id, isCompleted: !subtask.isCompleted)) {
                Image(systemName: subtask.isCompleted ? "checkmark.square.fill" : "square")
                    .font(.system(size: isCompact ? 13 : 15, weight: .semibold))
                    .foregroundStyle(subtask.isCompleted ? AppColors.statusDone : .secondary)
            }
            .buttonStyle(.plain)

            Text(subtask.title)
                .font(.system(size: isCompact ? 11 : 12))
                .lineLimit(1)
                .strikethrough(subtask.isCompleted)
                .foregroundStyle(subtask.isCompleted ? .secondary : .primary)

            Spacer()

            if let due = subtask.dueDate, !due.isEmpty {
                Text(formatSubtaskDue(due))
                    .font(.system(size: 9, weight: .medium))
                    .foregroundStyle(AppColors.priorityHigh)
            }
        }
    }

    private var emptyView: some View {
        VStack(spacing: 4) {
            Image(systemName: "checkmark.seal.fill")
                .font(.system(size: 20))
                .foregroundStyle(AppColors.statusDone)
            Text(L10n.Widgets.allDone)
                .font(.system(size: 12, weight: .semibold))
                .foregroundStyle(.primary)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }

    private func formatSubtaskDue(_ dateStr: String) -> String {
        guard let date = ISO8601DateFormatter().date(from: dateStr) else { return "" }
        let f = DateFormatter()
        f.dateFormat = "HH:mm"
        return f.string(from: date)
    }
}

// Helper to provide smooth gradient background to widgets
public extension View {
    func widgetBackground() -> some View {
        #if canImport(WidgetKit)
        if #available(iOSApplicationExtension 17.0, iOS 17.0, *) {
            return self.containerBackground(for: .widget) {
                LinearGradient(
                    colors: [
                        Color(red: 15/255, green: 23/255, blue: 42/255),
                        Color(red: 9/255, green: 13/255, blue: 22/255)
                    ],
                    startPoint: .topLeading,
                    endPoint: .bottomTrailing
                )
            }
        }
        #endif
        return self.background(
            LinearGradient(
                colors: [
                    Color(red: 15/255, green: 23/255, blue: 42/255),
                    Color(red: 9/255, green: 13/255, blue: 22/255)
                ],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
            )
        )
    }
}
