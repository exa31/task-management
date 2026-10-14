import Foundation
import AppIntents
#if canImport(WidgetKit)
import WidgetKit
#endif

public struct ToggleSubtaskIntent: AppIntent {
    public static var title: LocalizedStringResource = "Centang Subtask"
    public static var description = IntentDescription("Tandai subtask selesai atau belum selesai secara instan langsung dari widget.")

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
    public static var title: LocalizedStringResource = "Ubah Status Tugas"
    public static var description = IntentDescription("Ubah status tugas secara instan langsung dari widget.")

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
    public static var title: LocalizedStringResource = "Tugas Berikutnya"
    public static var description = IntentDescription("Beralih ke tugas aktif lainnya pada widget.")

    public init() {}

    public func perform() async throws -> some IntentResult {
        SharedWidgetDataManager.shared.cycleCustomTaskIndex()
        return .result()
    }
}
