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

    private var activeDeadlines: [TaskItem] {
        entry.bundle.upcomingDeadlines.filter { $0.status != "done" }
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
                HStack(spacing: 3) {
                    Image(systemName: "flame.fill")
                        .font(.system(size: 10, weight: .bold))
                    Text(L10n.Widgets.deadlineBadge)
                        .font(.system(size: 9, weight: .black))
                        .tracking(0.5)
                }
                .foregroundStyle(Color(red: 244/255, green: 63/255, blue: 94/255))
                .padding(.horizontal, 6)
                .padding(.vertical, 2)
                .background(Color(red: 244/255, green: 63/255, blue: 94/255).opacity(0.18))
                .clipShape(Capsule())

                Spacer()

                if !activeDeadlines.isEmpty {
                    Text("\(activeDeadlines.count)")
                        .font(.system(size: 10, weight: .bold))
                        .foregroundStyle(Color(red: 244/255, green: 63/255, blue: 94/255))
                        .padding(.horizontal, 5)
                        .padding(.vertical, 1.5)
                        .background(Color.white.opacity(0.08))
                        .clipShape(Capsule())
                }
            }

            if let top = activeDeadlines.first {
                VStack(alignment: .leading, spacing: 4) {
                    Text(top.title)
                        .font(.system(size: 12, weight: .bold))
                        .lineLimit(2)
                        .foregroundStyle(.white)

                    deadlineBadge(for: top.dueDate)

                    if let proj = top.projectName {
                        HStack(spacing: 4) {
                            Circle()
                                .fill(Color(hex: top.projectColor ?? "#3b82f6") ?? AppColors.primary)
                                .frame(width: 5, height: 5)
                            Text(proj)
                                .font(.system(size: 9, weight: .medium))
                                .foregroundStyle(.white.opacity(0.6))
                                .lineLimit(1)
                        }
                    }
                }
            } else {
                noDeadlinesView(isCompact: true)
            }

            Spacer(minLength: 0)
        }
        .padding(12)
        .widgetBackground(accentColor: Color(red: 244/255, green: 63/255, blue: 94/255))
        .widgetURL(activeDeadlines.first.flatMap { URL(string: "taskhub://tasks/\($0.id)") })
    }

    // MARK: - Medium Family

    private var mediumView: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack(spacing: 8) {
                HStack(spacing: 4) {
                    Image(systemName: "flame.fill")
                        .font(.system(size: 11, weight: .bold))
                    Text(L10n.Widgets.upcomingHeader)
                        .font(.system(size: 9, weight: .black))
                        .tracking(0.8)
                }
                .foregroundStyle(Color(red: 244/255, green: 63/255, blue: 94/255))
                .padding(.horizontal, 7)
                .padding(.vertical, 3)
                .background(Color(red: 244/255, green: 63/255, blue: 94/255).opacity(0.18))
                .clipShape(Capsule())

                Spacer()

                Text(L10n.Widgets.activeTasksCount(count: activeDeadlines.count))
                    .font(.system(size: 10, weight: .semibold))
                    .foregroundStyle(.white.opacity(0.6))
            }

            let deadlines = activeDeadlines.prefix(3)
            if deadlines.isEmpty {
                noDeadlinesView(isCompact: false)
            } else {
                VStack(spacing: 5) {
                    ForEach(deadlines) { task in
                        Link(destination: URL(string: "taskhub://tasks/\(task.id)")!) {
                            deadlineRow(task: task)
                        }
                        .buttonStyle(.plain)
                    }
                }
            }

            Spacer(minLength: 0)
        }
        .padding(14)
        .widgetBackground(accentColor: Color(red: 244/255, green: 63/255, blue: 94/255))
        .widgetURL(activeDeadlines.first.flatMap { URL(string: "taskhub://tasks/\($0.id)") })
    }

    // MARK: - Large Family

    private var largeView: some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack(spacing: 8) {
                HStack(spacing: 4) {
                    Image(systemName: "flame.fill")
                        .font(.system(size: 13, weight: .bold))
                    Text(L10n.Widgets.deadlinesTitleLarge)
                        .font(.system(size: 11, weight: .black))
                        .tracking(1)
                }
                .foregroundStyle(Color(red: 244/255, green: 63/255, blue: 94/255))
                .padding(.horizontal, 8)
                .padding(.vertical, 4)
                .background(Color(red: 244/255, green: 63/255, blue: 94/255).opacity(0.18))
                .clipShape(Capsule())

                Spacer()

                Text(L10n.Widgets.upcomingTasksCount(count: activeDeadlines.count))
                    .font(.system(size: 11, weight: .semibold))
                    .foregroundStyle(.white.opacity(0.7))
            }

            let deadlines = activeDeadlines.prefix(5)
            if deadlines.isEmpty {
                noDeadlinesView(isCompact: false)
            } else {
                VStack(spacing: 6) {
                    ForEach(deadlines) { task in
                        Link(destination: URL(string: "taskhub://tasks/\(task.id)")!) {
                            deadlineRow(task: task, showDetails: true)
                        }
                        .buttonStyle(.plain)
                    }
                }
            }

            Spacer(minLength: 0)
        }
        .padding(16)
        .widgetBackground(accentColor: Color(red: 244/255, green: 63/255, blue: 94/255))
        .widgetURL(activeDeadlines.first.flatMap { URL(string: "taskhub://tasks/\($0.id)") })
    }

    // MARK: - Deadline Row

    @ViewBuilder
    private func deadlineRow(task: TaskItem, showDetails: Bool = false) -> some View {
        HStack(spacing: 8) {
            RoundedRectangle(cornerRadius: 2)
                .fill(priorityColor(task.priority))
                .frame(width: 3, height: showDetails ? 28 : 20)

            VStack(alignment: .leading, spacing: 2) {
                Text(task.title)
                    .font(.system(size: 12, weight: .semibold))
                    .lineLimit(1)
                    .foregroundStyle(.white)

                if let proj = task.projectName {
                    HStack(spacing: 3) {
                        Circle()
                            .fill(Color(hex: task.projectColor ?? "#3b82f6") ?? AppColors.primary)
                            .frame(width: 4, height: 4)
                        Text(proj)
                            .font(.system(size: 9))
                            .foregroundStyle(.white.opacity(0.55))
                    }
                }
            }

            Spacer()

            deadlineBadge(for: task.dueDate)
        }
        .padding(.horizontal, 8)
        .padding(.vertical, showDetails ? 6 : 5)
        .background(Color.white.opacity(0.04))
        .clipShape(RoundedRectangle(cornerRadius: 8, style: .continuous))
    }

    @ViewBuilder
    private func deadlineBadge(for dateStr: String?) -> some View {
        if let str = dateStr, let date = ISO8601DateFormatter().date(from: str) {
            let info = formatDeadlineUrgency(date)
            HStack(spacing: 3) {
                Image(systemName: info.icon)
                    .font(.system(size: 8, weight: .bold))
                Text(info.label)
                    .font(.system(size: 9, weight: .bold))
            }
            .padding(.horizontal, 6)
            .padding(.vertical, 2.5)
            .background(info.color.opacity(0.18))
            .foregroundStyle(info.color)
            .clipShape(Capsule())
        }
    }

    private func formatDeadlineUrgency(_ date: Date) -> (label: String, color: Color, icon: String) {
        let calendar = Calendar.current
        if date < Date() {
            return (L10n.Widgets.overdue, Color(red: 239/255, green: 68/255, blue: 68/255), "exclamationmark.circle.fill")
        } else if calendar.isDateInToday(date) {
            let f = DateFormatter()
            f.dateFormat = "HH:mm"
            return (L10n.Widgets.todayAt(time: f.string(from: date)), Color(red: 244/255, green: 63/255, blue: 94/255), "flame.fill")
        } else if calendar.isDateInTomorrow(date) {
            return (L10n.Widgets.tomorrow, Color(red: 245/255, green: 158/255, blue: 11/255), "clock.fill")
        } else {
            let f = DateFormatter()
            f.dateFormat = "d MMM"
            return (f.string(from: date), Color(red: 56/255, green: 189/255, blue: 248/255), "calendar")
        }
    }

    private func priorityColor(_ priority: String) -> Color {
        switch priority.lowercased() {
        case "urgent": return Color(red: 239/255, green: 68/255, blue: 68/255)
        case "high": return Color(red: 245/255, green: 158/255, blue: 11/255)
        case "medium": return Color(red: 59/255, green: 130/255, blue: 246/255)
        default: return Color(red: 148/255, green: 163/255, blue: 184/255)
        }
    }

    @ViewBuilder
    private func noDeadlinesView(isCompact: Bool) -> some View {
        VStack(spacing: isCompact ? 4 : 6) {
            Image(systemName: "checkmark.shield.fill")
                .font(.system(size: isCompact ? 20 : 24))
                .foregroundStyle(Color(red: 52/255, green: 211/255, blue: 153/255))

            Text(L10n.Widgets.noUrgentTitle)
                .font(.system(size: isCompact ? 11 : 12, weight: .semibold))
                .foregroundStyle(.white)

            if !isCompact {
                Text(L10n.Widgets.noUrgentDesc)
                    .font(.system(size: 9))
                    .foregroundStyle(.white.opacity(0.5))
            }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}
