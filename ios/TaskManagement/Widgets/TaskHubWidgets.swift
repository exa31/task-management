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
        .configurationDisplayName(LocalizedStringKey("widget.recurring.configName"))
        .description(LocalizedStringKey("widget.recurring.configDesc"))
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
        .configurationDisplayName(LocalizedStringKey("widget.custom.configName"))
        .description(LocalizedStringKey("widget.custom.configDesc"))
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
        .configurationDisplayName(LocalizedStringKey("widget.deadlines.configName"))
        .description(LocalizedStringKey("widget.deadlines.configDesc"))
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
        .configurationDisplayName(LocalizedStringKey("widget.overview.configName"))
        .description(LocalizedStringKey("widget.overview.configDesc"))
        .supportedFamilies([.systemSmall, .systemMedium])
    }
}

#if os(iOS)
// MARK: - 5. Lock Screen & StandBy Widget
public struct LockScreenWidget: Widget {
    public static let kind: String = "LockScreenWidget"

    public init() {}

    public var body: some WidgetConfiguration {
        StaticConfiguration(kind: Self.kind, provider: TaskWidgetTimelineProvider()) { entry in
            LockScreenWidgetView(entry: entry)
        }
        .configurationDisplayName(LocalizedStringKey("widget.lockscreen.configName"))
        .description(LocalizedStringKey("widget.lockscreen.configDesc"))
        .supportedFamilies([.accessoryInline, .accessoryCircular, .accessoryRectangular])
    }
}
#endif

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
        #if os(iOS)
        LockScreenWidget()
        #endif
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
        #if os(iOS)
        LockScreenWidget()
        #endif
    }
}
#endif
#endif
