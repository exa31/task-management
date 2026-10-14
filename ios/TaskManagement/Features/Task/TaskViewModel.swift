import SwiftUI

@Observable
public final class TaskViewModel {
    public var tasks: [TaskItem] = []
    public var isLoading = false
    public var errorMessage: String?
    public var selectedStatus: String? = nil
    public var selectedProjectId: Int64? = nil
    public var selectedPriority: String? = nil
    public var filterFocusOnly: Bool = false
    public var searchQuery: String = ""
    public var projects: [Project] = []
    public var taskActivities: [TaskActivity] = []
    public var isLoadingActivities = false

    private let apiClient = APIClient.shared

    public init() {
        Task {
            await loadTasks()
            await loadProjects()
        }
    }

    @MainActor
    public func loadTasks() async {
        isLoading = true
        errorMessage = nil
        do {
            self.tasks = try await apiClient.getTasks()
            AppLogger.info("Tasks loaded successfully", category: .task, metadata: ["count": tasks.count])
            SharedWidgetDataManager.shared.syncWithTaskList(self.tasks)
            Task {
                await SharedWidgetDataManager.shared.syncWithBackend()
            }
        } catch {
            AppLogger.error("Failed to load tasks", error: error, category: .task)
            self.errorMessage = error.localizedDescription
        }
        self.isLoading = false
    }

    @MainActor
    public func loadProjects() async {
        do {
            self.projects = try await apiClient.getProjects()
            AppLogger.debug("Projects loaded for tasks filter", category: .task, metadata: ["count": projects.count])
        } catch {
            AppLogger.warn("Failed to load projects for task association", category: .task, metadata: ["error": error.localizedDescription])
        }
    }

    public var filteredTasks: [TaskItem] {
        var result = tasks

        // Search Query filter (checks title, description, tags, project name)
        let q = searchQuery.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
        if !q.isEmpty {
            result = result.filter { task in
                task.title.lowercased().contains(q) ||
                task.description.lowercased().contains(q) ||
                (task.tags?.contains(where: { $0.lowercased().contains(q) }) ?? false) ||
                (task.projectName?.lowercased().contains(q) ?? false)
            }
        }

        // Status filter
        if let status = selectedStatus, !status.isEmpty {
            result = result.filter { $0.status == status }
        }

        // Project filter
        if let projectId = selectedProjectId {
            result = result.filter { $0.projectId == projectId }
        }

        // Priority filter
        if let priority = selectedPriority, !priority.isEmpty {
            result = result.filter { $0.priority.lowercased() == priority.lowercased() }
        }

        // Focus filter
        if filterFocusOnly {
            result = result.filter { $0.isFocus }
        }

        let statusOrder: [String: Int] = [
            "in_progress": 1,
            "todo": 2,
            "review": 3,
            "backlog": 4,
            "done": 5
        ]
        let priorityOrder: [String: Int] = [
            "urgent": 1,
            "high": 2,
            "medium": 3,
            "low": 4
        ]

        return result.sorted { a, b in
            let aFocus = a.isFocus && a.status != "done" ? 1 : 0
            let bFocus = b.isFocus && b.status != "done" ? 1 : 0
            if aFocus != bFocus { return aFocus > bFocus }

            let aStatus = statusOrder[a.status] ?? 99
            let bStatus = statusOrder[b.status] ?? 99
            if aStatus != bStatus { return aStatus < bStatus }

            let aPriority = priorityOrder[a.priority.lowercased()] ?? 99
            let bPriority = priorityOrder[b.priority.lowercased()] ?? 99
            if aPriority != bPriority { return aPriority < bPriority }

            return a.id > b.id
        }
    }

    public var focusTasks: [TaskItem] {
        tasks.filter { $0.isFocus && $0.status != "done" }
    }

    public var hasActiveFilters: Bool {
        selectedStatus != nil ||
        selectedProjectId != nil ||
        selectedPriority != nil ||
        filterFocusOnly ||
        !searchQuery.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
    }

    public var activeFiltersCount: Int {
        var count = 0
        if selectedStatus != nil { count += 1 }
        if selectedProjectId != nil { count += 1 }
        if selectedPriority != nil { count += 1 }
        if filterFocusOnly { count += 1 }
        if !searchQuery.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty { count += 1 }
        return count
    }

    public func resetFilters() {
        selectedStatus = nil
        selectedProjectId = nil
        selectedPriority = nil
        filterFocusOnly = false
        searchQuery = ""
    }

    @MainActor
    public func setStatusFilter(_ status: String?) {
        self.selectedStatus = status
    }

    @MainActor
    public func toggleFocus(task: TaskItem) {
        Task {
            do {
                let updated = try await apiClient.toggleTaskFocus(id: task.id)
                if let idx = tasks.firstIndex(where: { $0.id == task.id }) {
                    tasks[idx] = updated
                }
            } catch {
                self.errorMessage = error.localizedDescription
            }
        }
    }

    @MainActor
    public func toggleTaskStatus(task: TaskItem) {
        let newStatus = task.status == "done" ? "todo" : "done"
        if let idx = tasks.firstIndex(where: { $0.id == task.id }) {
            var copy = tasks[idx]
            copy.status = newStatus
            tasks[idx] = copy
            SharedWidgetDataManager.shared.syncWithTaskList(self.tasks)
        }
        Task {
            do {
                _ = try await apiClient.updateTask(
                    id: task.id,
                    req: UpdateTaskRequest(status: newStatus)
                )
                await loadTasks()
            } catch {
                self.errorMessage = error.localizedDescription
            }
        }
    }

    @MainActor
    public func updateTask(
        id: Int64,
        title: String? = nil,
        description: String? = nil,
        status: String? = nil,
        priority: String? = nil,
        dueDate: String? = nil,
        clearDueDate: Bool = false,
        isFocus: Bool? = nil,
        tags: [String]? = nil,
        projectId: Int64? = nil,
        clearProjectId: Bool = false,
        isRecurring: Bool? = nil,
        recurrencePattern: String? = nil,
        recurrenceInterval: Int? = nil,
        recurrenceEndDate: String? = nil,
        clearRecurrenceEndDate: Bool = false
    ) async -> Bool {
        do {
            let updated = try await apiClient.updateTask(
                id: id,
                req: UpdateTaskRequest(
                    title: title,
                    description: description,
                    status: status,
                    priority: priority,
                    dueDate: dueDate,
                    clearDueDate: clearDueDate,
                    isFocus: isFocus,
                    tags: tags,
                    projectId: projectId,
                    clearProjectId: clearProjectId,
                    isRecurring: isRecurring,
                    recurrencePattern: recurrencePattern,
                    recurrenceInterval: recurrenceInterval,
                    recurrenceEndDate: recurrenceEndDate,
                    clearRecurrenceEndDate: clearRecurrenceEndDate
                )
            )
            if let idx = tasks.firstIndex(where: { $0.id == id }) {
                tasks[idx] = updated
                SharedWidgetDataManager.shared.syncWithTaskList(self.tasks)
            }
            AppLogger.success("Task updated successfully", category: .task, metadata: ["id": id])
            await loadTasks()
            return true
        } catch {
            AppLogger.error("Failed to update task", error: error, category: .task)
            self.errorMessage = error.localizedDescription
            return false
        }
    }

    @MainActor
    public func createTask(
        title: String,
        description: String = "",
        status: String = "todo",
        priority: String = "medium",
        dueDate: String? = nil,
        isFocus: Bool = false,
        tags: [String] = [],
        projectId: Int64? = nil,
        subtasks: [String] = [],
        isRecurring: Bool = false,
        recurrencePattern: String? = nil,
        recurrenceInterval: Int? = nil,
        recurrenceEndDate: String? = nil
    ) async -> Bool {
        do {
            _ = try await apiClient.createTask(
                req: CreateTaskRequest(
                    title: title,
                    description: description,
                    status: status,
                    priority: priority,
                    dueDate: dueDate,
                    isFocus: isFocus,
                    tags: tags,
                    projectId: projectId,
                    subtasks: subtasks,
                    isRecurring: isRecurring,
                    recurrencePattern: recurrencePattern,
                    recurrenceInterval: recurrenceInterval,
                    recurrenceEndDate: recurrenceEndDate
                )
            )
            AppLogger.success("Task created successfully", category: .task, metadata: ["title": title])
            await loadTasks()
            return true
        } catch {
            AppLogger.error("Failed to create task", error: error, category: .task)
            self.errorMessage = error.localizedDescription
            return false
        }
    }

    @MainActor
    public func createTask(title: String, description: String, priority: String, projectId: Int64? = nil) {
        Task {
            _ = await createTask(
                title: title,
                description: description,
                status: "todo",
                priority: priority,
                projectId: projectId
            )
        }
    }

    @MainActor
    public func deleteTask(id: Int64) async {
        do {
            try await apiClient.deleteTask(id: id)
            tasks.removeAll { $0.id == id }
            SharedWidgetDataManager.shared.syncWithTaskList(self.tasks)
            AppLogger.success("Task deleted successfully", category: .task, metadata: ["id": id])
            await loadTasks()
        } catch {
            AppLogger.error("Failed to delete task", error: error, category: .task, metadata: ["id": id])
            self.errorMessage = error.localizedDescription
        }
    }

    @MainActor
    public func toggleSubtask(taskId: Int64, subtaskId: Int64, currentCompleted: Bool) {
        // Immediate local mutation for offline widget and instant response
        SharedWidgetDataManager.shared.toggleSubtaskLocally(
            taskId: taskId,
            subtaskId: subtaskId,
            isCompleted: !currentCompleted
        )

        Task {
            do {
                _ = try await apiClient.updateSubtask(
                    taskId: taskId,
                    subtaskId: subtaskId,
                    isCompleted: !currentCompleted
                )
                AppLogger.success("Subtask completion toggled", category: .task, metadata: ["taskId": taskId, "subtaskId": subtaskId])
                await loadTasks()
            } catch {
                AppLogger.error("Failed to toggle subtask", error: error, category: .task)
                self.errorMessage = error.localizedDescription
            }
        }
    }

    @MainActor
    public func loadActivities(for taskId: Int64) async {
        isLoadingActivities = true
        do {
            self.taskActivities = try await apiClient.getTaskActivities(taskId: taskId)
            AppLogger.debug("Task activities loaded", category: .task, metadata: ["taskId": taskId, "count": taskActivities.count])
        } catch {
            AppLogger.error("Failed to load task activities", error: error, category: .task, metadata: ["taskId": taskId])
            self.errorMessage = error.localizedDescription
        }
        isLoadingActivities = false
    }

    @MainActor
    public func addActivityNote(taskId: Int64, note: String) async {
        guard !note.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty else { return }
        do {
            let newActivity = try await apiClient.createTaskActivity(
                taskId: taskId,
                req: CreateTaskActivityRequest(description: note.trimmingCharacters(in: .whitespacesAndNewlines))
            )
            AppLogger.success("Task activity note appended", category: .task, metadata: ["taskId": taskId])
            self.taskActivities.insert(newActivity, at: 0)
        } catch {
            AppLogger.error("Failed to add task activity note", error: error, category: .task, metadata: ["taskId": taskId])
            self.errorMessage = error.localizedDescription
        }
    }
}
