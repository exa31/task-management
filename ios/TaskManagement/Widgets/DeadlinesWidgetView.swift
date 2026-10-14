import SwiftUI
#if canImport(WidgetKit)
import WidgetKit
#endif

public struct DeadlinesWidgetView: View {
    @Environment(\.widgetFamily) var family
    public let entry: TaskWidgetEntry

    public init(entry: TaskWidgetEntry) {
        self.entry = entry
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
            HStack {
                Image(systemName: "clock.badge.exclamationmark")
                    .font(.system(size: 13, weight: .bold))
                    .foregroundStyle(AppColors.priorityUrgent)
                Text(L10n.Widgets.deadlineBadge)
                    .font(.system(size: 10, weight: .black))
                    .tracking(1)
                    .foregroundStyle(AppColors.priorityUrgent)
                Spacer()
            }

            if let top = entry.bundle.upcomingDeadlines.first {
                VStack(alignment: .leading, spacing: 4) {
                    Text(top.title)
                        .font(.system(size: 13, weight: .bold))
                        .lineLimit(2)
                        .foregroundStyle(.primary)

                    deadlineBadge(for: top.dueDate)

                    if let proj = top.projectName {
                        Text(proj)
                            .font(.system(size: 10, weight: .medium))
                            .foregroundStyle(.secondary)
                    }
                }
            } else {
                noDeadlinesView
            }
            Spacer(minLength: 0)
        }
        .padding(12)
        .widgetBackground()
    }

    // MARK: - Medium Family

    private var mediumView: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                HStack(spacing: 6) {
                    Image(systemName: "calendar.badge.clock")
                        .font(.system(size: 14, weight: .bold))
                        .foregroundStyle(AppColors.priorityUrgent)
                    Text(L10n.Widgets.upcomingHeader)
                        .font(.system(size: 10, weight: .black))
                        .tracking(1)
                        .foregroundStyle(AppColors.priorityUrgent)
                }
                Spacer()
                Text(L10n.Widgets.tasksCount(count: entry.bundle.upcomingDeadlines.count))
                    .font(.system(size: 11, weight: .semibold))
                    .foregroundStyle(.secondary)
            }

            Divider().opacity(0.3)

            let deadlines = entry.bundle.upcomingDeadlines.prefix(3)
            if deadlines.isEmpty {
                noDeadlinesView
            } else {
                VStack(spacing: 6) {
                    ForEach(deadlines) { task in
                        deadlineRow(task: task)
                    }
                }
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
                BrandLogoView(size: 22)
                Text(L10n.Widgets.listTitle)
                    .font(.system(size: 15, weight: .bold))
                Spacer()
                Text(L10n.Widgets.activeTasksCount(count: entry.bundle.upcomingDeadlines.count))
                    .font(.system(size: 11, weight: .semibold))
                    .foregroundStyle(.secondary)
            }

            Divider().opacity(0.3)

            let deadlines = entry.bundle.upcomingDeadlines.prefix(5)
            if deadlines.isEmpty {
                noDeadlinesView
            } else {
                VStack(spacing: 8) {
                    ForEach(deadlines) { task in
                        deadlineRow(task: task, showDetails: true)
                    }
                }
            }

            Spacer(minLength: 0)
        }
        .padding(14)
        .widgetBackground()
    }

    // MARK: - Deadline Row

    @ViewBuilder
    private func deadlineRow(task: TaskItem, showDetails: Bool = false) -> some View {
        HStack(spacing: 8) {
            Circle()
                .fill(priorityColor(task.priority))
                .frame(width: 8, height: 8)

            VStack(alignment: .leading, spacing: 2) {
                Text(task.title)
                    .font(.system(size: 12, weight: .semibold))
                    .lineLimit(1)
                    .foregroundStyle(.primary)

                if showDetails, let desc = task.projectName {
                    Text(desc)
                        .font(.system(size: 10))
                        .foregroundStyle(.secondary)
                }
            }

            Spacer()

            deadlineBadge(for: task.dueDate)
        }
        .padding(showDetails ? 8 : 4)
        .background(Color.primary.opacity(0.04))
        .clipShape(RoundedRectangle(cornerRadius: 6, style: .continuous))
    }

    @ViewBuilder
    private func deadlineBadge(for dateStr: String?) -> some View {
        if let str = dateStr, let date = ISO8601DateFormatter().date(from: str) {
            let info = formatDeadlineUrgency(date)
            HStack(spacing: 3) {
                Image(systemName: info.icon)
                    .font(.system(size: 9))
                Text(info.label)
                    .font(.system(size: 10, weight: .bold))
            }
            .padding(.horizontal, 6)
            .padding(.vertical, 2)
            .background(info.color.opacity(0.18))
            .foregroundStyle(info.color)
            .clipShape(Capsule())
        } else {
            EmptyView()
        }
    }

    private func formatDeadlineUrgency(_ date: Date) -> (label: String, color: Color, icon: String) {
        let calendar = Calendar.current
        if calendar.isDateInToday(date) {
            let f = DateFormatter()
            f.dateFormat = "HH:mm"
            return (L10n.Widgets.todayAt(time: f.string(from: date)), AppColors.priorityUrgent, "flame.fill")
        } else if calendar.isDateInTomorrow(date) {
            return (L10n.Widgets.tomorrow, AppColors.priorityHigh, "clock.fill")
        } else if date < Date() {
            return (L10n.Widgets.overdue, AppColors.priorityUrgent, "exclamationmark.triangle.fill")
        } else {
            let f = DateFormatter()
            f.dateFormat = "dd MMM"
            return (f.string(from: date), AppColors.primary, "calendar")
        }
    }

    private func priorityColor(_ priority: String) -> Color {
        switch priority.lowercased() {
        case "urgent": return AppColors.priorityUrgent
        case "high": return AppColors.priorityHigh
        case "medium": return AppColors.priorityMedium
        default: return AppColors.priorityLow
        }
    }

    private var noDeadlinesView: some View {
        VStack(spacing: 4) {
            Image(systemName: "checkmark.circle.fill")
                .font(.system(size: 20))
                .foregroundStyle(AppColors.statusDone)
            Text(L10n.Widgets.noUrgent)
                .font(.system(size: 12, weight: .semibold))
                .foregroundStyle(.primary)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}
