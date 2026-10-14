import Foundation
import AppIntents
#if canImport(WidgetKit)
import WidgetKit
#endif

public struct ToggleSubtaskIntent: AppIntent {
    public static var title: LocalizedStringResource = "Toggle Subtask"
    public static var description = IntentDescription("Check or uncheck subtasks instantly from the widget.")

    @Parameter(title: "Task ID")
    public var taskId: Int

    @Parameter(title: "Subtask ID")
    public var subtaskId: Int

    @Parameter(title: "Is Completed")
    public var isCompleted: Bool

    public init() {
        self.taskId = 0
        self.subtaskId = 0
        self.isCompleted = false
    }

    public init(taskId: Int64, subtaskId: Int64, isCompleted: Bool) {
        self.taskId = Int(taskId)
        self.subtaskId = Int(subtaskId)
        self.isCompleted = isCompleted
    }

    public func perform() async throws -> some IntentResult {
        SharedWidgetDataManager.shared.toggleSubtaskLocally(
            taskId: Int64(taskId),
            subtaskId: Int64(subtaskId),
            isCompleted: isCompleted
        )

        #if canImport(WidgetKit)
        WidgetCenter.shared.reloadAllTimelines()
        #endif

        await SharedWidgetDataManager.shared.syncWithBackend()

        return .result()
    }
}

public struct UpdateTaskStatusIntent: AppIntent {
    public static var title: LocalizedStringResource = "Update Task Status"
    public static var description = IntentDescription("Update task status directly from the widget.")

    @Parameter(title: "Task ID")
    public var taskId: Int

    @Parameter(title: "New Status")
    public var newStatus: String

    public init() {
        self.taskId = 0
        self.newStatus = "done"
    }

    public init(taskId: Int64, newStatus: String) {
        self.taskId = Int(taskId)
        self.newStatus = newStatus
    }

    public func perform() async throws -> some IntentResult {
        SharedWidgetDataManager.shared.updateTaskStatusLocally(
            taskId: Int64(taskId),
            status: newStatus
        )

        #if canImport(WidgetKit)
        WidgetCenter.shared.reloadAllTimelines()
        #endif

        await SharedWidgetDataManager.shared.syncWithBackend()

        return .result()
    }
}

public struct CycleNextTaskIntent: AppIntent {
    public static var title: LocalizedStringResource = "Next Task"
    public static var description = IntentDescription("Switch to another active task on the widget.")

    public init() {}

    public func perform() async throws -> some IntentResult {
        SharedWidgetDataManager.shared.cycleCustomTaskIndex()
        return .result()
    }
}

public struct CycleSubtaskPageIntent: AppIntent {
    public static var title: LocalizedStringResource = "Cycle Subtask Page"
    public static var description = IntentDescription("Switch to the next or previous page of subtasks on the widget.")

    @Parameter(title: "Task ID")
    public var taskId: Int

    @Parameter(title: "Delta")
    public var delta: Int

    @Parameter(title: "Total Pages")
    public var totalPages: Int

    public init() {
        self.taskId = 0
        self.delta = 1
        self.totalPages = 1
    }

    public init(taskId: Int64, delta: Int = 1, totalPages: Int = 1) {
        self.taskId = Int(taskId)
        self.delta = delta
        self.totalPages = max(1, totalPages)
    }

    public func perform() async throws -> some IntentResult {
        SharedWidgetDataManager.shared.stepSubtaskPage(taskId: Int64(taskId), delta: delta, totalPages: totalPages)

        #if canImport(WidgetKit)
        WidgetCenter.shared.reloadAllTimelines()
        #endif

        return .result()
    }
}
