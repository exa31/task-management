import SwiftUI
#if canImport(WidgetKit)
import WidgetKit
#endif

public struct LockScreenWidgetView: View {
    @Environment(\.widgetFamily) var family
    public let entry: TaskWidgetEntry

    public init(entry: TaskWidgetEntry) {
        self.entry = entry
    }

    private var percent: Double {
        guard entry.bundle.totalToday > 0 else { return 0 }
        return Double(entry.bundle.completedToday) / Double(entry.bundle.totalToday)
    }

    private var topTask: TaskItem? {
        entry.bundle.upcomingDeadlines.first ?? entry.bundle.focusTasks.first
    }

    public var body: some View {
        switch family {
        #if canImport(WidgetKit)
        case .accessoryInline:
            inlineView
        case .accessoryCircular:
            circularView
        case .accessoryRectangular:
            rectangularView
        #endif
        default:
            inlineView
        }
    }

    // MARK: - Inline View

    private var inlineView: some View {
        HStack(spacing: 4) {
            Image(systemName: "bolt.fill")
            if let task = topTask {
                Text("\(entry.bundle.completedToday)/\(entry.bundle.totalToday) • \(task.title)")
            } else {
                Text(L10n.Widgets.lockscreenTasksDone(completed: entry.bundle.completedToday, total: entry.bundle.totalToday))
            }
        }
    }

    // MARK: - Circular Accessory

    private var circularView: some View {
        Gauge(value: min(percent, 1.0)) {
            Image(systemName: "bolt.fill")
        } currentValueLabel: {
            Text("\(Int(percent * 100))%")
                .font(.system(size: 11, weight: .black))
        }
        .gaugeStyle(.accessoryCircular)
    }

    // MARK: - Rectangular Accessory

    private var rectangularView: some View {
        VStack(alignment: .leading, spacing: 2) {
            HStack(spacing: 4) {
                Image(systemName: "bolt.fill")
                    .font(.system(size: 10, weight: .bold))
                Text("TASK HUB")
                    .font(.system(size: 9, weight: .heavy))
                Spacer()
                Text("\(entry.bundle.completedToday)/\(entry.bundle.totalToday)")
                    .font(.system(size: 10, weight: .bold))
            }

            if let task = topTask {
                Text(task.title)
                    .font(.system(size: 12, weight: .bold))
                    .lineLimit(1)

                if let due = task.dueDate, !formatTime(due).isEmpty {
                    HStack(spacing: 2) {
                        Image(systemName: "clock")
                            .font(.system(size: 8))
                        Text(L10n.Widgets.duePrefix(time: formatTime(due)))
                            .font(.system(size: 9))
                    }
                    .foregroundStyle(.secondary)
                } else if let note = task.lastProgressNote, !note.isEmpty {
                    Text(note)
                        .font(.system(size: 9))
                        .foregroundStyle(.secondary)
                        .lineLimit(1)
                }
            } else {
                Text(L10n.Widgets.lockscreenAllDone)
                    .font(.system(size: 11))
                    .foregroundStyle(.secondary)
            }
        }
    }

    private func formatTime(_ dateStr: String) -> String {
        guard let date = ISO8601DateFormatter().date(from: dateStr) else { return "" }
        let f = DateFormatter()
        f.dateFormat = "HH:mm"
        return f.string(from: date)
    }
}
