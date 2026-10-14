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
