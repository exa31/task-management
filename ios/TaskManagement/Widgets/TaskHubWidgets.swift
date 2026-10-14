import SwiftUI
#if canImport(WidgetKit)
import WidgetKit

// MARK: - 1. Interactive Checklist Widget
public struct TaskChecklistWidget: Widget {
    public static let kind: String = "TaskChecklistWidget"

    public init() {}

    public var body: some WidgetConfiguration {
        StaticConfiguration(kind: Self.kind, provider: TaskWidgetTimelineProvider()) { entry in
            ChecklistWidgetView(entry: entry)
        }
        .configurationDisplayName("Checklist Subtask")
        .description("Centang subtask fokus hari ini secara instan langsung dari layar utama tanpa membuka aplikasi.")
        .supportedFamilies([.systemSmall, .systemMedium, .systemLarge])
    }
}

// MARK: - 2. Upcoming Deadlines Widget
public struct TaskDeadlinesWidget: Widget {
    public static let kind: String = "TaskDeadlinesWidget"

    public init() {}

    public var body: some WidgetConfiguration {
        StaticConfiguration(kind: Self.kind, provider: TaskWidgetTimelineProvider()) { entry in
            DeadlinesWidgetView(entry: entry)
        }
        .configurationDisplayName("Tenggat Waktu")
        .description("Pantau tugas-tugas penting yang mendekati batas waktu dengan penanda urgensi warna.")
        .supportedFamilies([.systemSmall, .systemMedium, .systemLarge])
    }
}

// MARK: - 3. Daily Overview & Productivity Ring
public struct DailyOverviewWidget: Widget {
    public static let kind: String = "DailyOverviewWidget"

    public init() {}

    public var body: some WidgetConfiguration {
        StaticConfiguration(kind: Self.kind, provider: TaskWidgetTimelineProvider()) { entry in
            DailyOverviewWidgetView(entry: entry)
        }
        .configurationDisplayName("Ringkasan Harian")
        .description("Lihat rasio penyelesaian tugas harianmu dengan progress ring yang memotivasi.")
        .supportedFamilies([.systemSmall, .systemMedium])
    }
}

// MARK: - 4. Lock Screen & StandBy Widget
public struct LockScreenWidget: Widget {
    public static let kind: String = "LockScreenWidget"

    public init() {}

    public var body: some WidgetConfiguration {
        StaticConfiguration(kind: Self.kind, provider: TaskWidgetTimelineProvider()) { entry in
            LockScreenWidgetView(entry: entry)
        }
        .configurationDisplayName("Layar Kunci Task Hub")
        .description("Pantau progres dan tugas terdekatmu langsung di Lock Screen atau StandBy Mode.")
        .supportedFamilies([.accessoryInline, .accessoryCircular, .accessoryRectangular])
    }
}

// MARK: - Widget Bundle
#if WIDGET_EXTENSION
@main
public struct TaskHubWidgetBundle: WidgetBundle {
    public init() {}

    public var body: some Widget {
        TaskChecklistWidget()
        TaskDeadlinesWidget()
        DailyOverviewWidget()
        LockScreenWidget()
    }
}
#else
public struct TaskHubWidgetBundle: WidgetBundle {
    public init() {}

    public var body: some Widget {
        TaskChecklistWidget()
        TaskDeadlinesWidget()
        DailyOverviewWidget()
        LockScreenWidget()
    }
}
#endif
#endif
