import SwiftUI
#if canImport(WidgetKit)
import WidgetKit

// MARK: - 1. Recurring Task Checklist Widget
public struct TaskRecurringWidget: Widget {
    public static let kind: String = "TaskRecurringWidget"

    public init() {}

    public var body: some WidgetConfiguration {
        StaticConfiguration(kind: Self.kind, provider: TaskWidgetTimelineProvider()) { entry in
            RecurringTaskWidgetView(entry: entry)
        }
        .configurationDisplayName("Tugas Rutin Harian")
        .description("Pantau checklist tugas berulang harianmu dan selesaikan langsung dari layar utama.")
        .supportedFamilies([.systemSmall, .systemMedium, .systemLarge])
    }
}

// MARK: - 2. Custom Selected Task Checklist Widget
public struct TaskCustomChecklistWidget: Widget {
    public static let kind: String = "TaskCustomChecklistWidget"

    public init() {}

    public var body: some WidgetConfiguration {
        StaticConfiguration(kind: Self.kind, provider: TaskWidgetTimelineProvider()) { entry in
            CustomTaskWidgetView(entry: entry)
        }
        .configurationDisplayName("Checklist Tugas Pilihan")
        .description("Pantau subtask dan progres tugas pilihan aktifmu langsung dari layar utama.")
        .supportedFamilies([.systemSmall, .systemMedium, .systemLarge])
    }
}

// MARK: - 3. Upcoming Deadlines Widget
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

// MARK: - 4. Daily Overview & Productivity Ring
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

// MARK: - 5. Lock Screen & StandBy Widget
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
        TaskRecurringWidget()
        TaskCustomChecklistWidget()
        TaskDeadlinesWidget()
        DailyOverviewWidget()
        LockScreenWidget()
    }
}
#else
public struct TaskHubWidgetBundle: WidgetBundle {
    public init() {}

    public var body: some Widget {
        TaskRecurringWidget()
        TaskCustomChecklistWidget()
        TaskDeadlinesWidget()
        DailyOverviewWidget()
        LockScreenWidget()
    }
}
#endif
#endif
