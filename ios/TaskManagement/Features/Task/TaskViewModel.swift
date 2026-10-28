import SwiftUI

@Observable
public final class TaskViewModel {
    public var tasks: [TaskItem] = []
    public var isLoading = false
    public var hasMore = false
    public var isLoadingMore = false
    public var errorMessage: String?
    public var selectedStatus: String? = nil
    public var selectedProjectId: Int64? = nil
    public var selectedPriority: String? = nil
    public var filterFocusOnly: Bool = false
    public var searchQuery: String = ""
    public var projects: [Project] = []
    public var taskActivities: [TaskActivity] = []
    public var isLoadingActivities = false
    public var selectedTaskForDetail: TaskItem? = nil

    // MARK: - Recurring Task Templates State
    public var recurringTasks: [RecurringTaskItem] = []
    public var isLoadingRecurring = false
    public var recurringSearchQuery: String = ""
    public var recurringStatusFilter: String = "all" // "all", "active", "paused"
    public var selectedRecurringTaskForEdit: RecurringTaskItem? = nil

    public var filteredRecurringTasks: [RecurringTaskItem] {
        recurringTasks.filter { item in
            let matchesSearch: Bool
            let q = recurringSearchQuery.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
            if q.isEmpty {
                matchesSearch = true
            } else {
                let titleMatch = item.title.lowercased().contains(q)
                let descMatch = item.description.lowercased().contains(q)
                let projectMatch = item.projectName?.lowercased().contains(q) ?? false
                matchesSearch = titleMatch || descMatch || projectMatch
            }

            let matchesStatus: Bool
            switch recurringStatusFilter {
            case "active":
                matchesStatus = item.isActive
            case "paused":
                matchesStatus = !item.isActive
            default:
                matchesStatus = true
            }

            return matchesSearch && matchesStatus
        }
    }

    private var nextCursor: Int64? = nil
    private let apiClient = APIClient.shared

    public init() {
        Task {
            await loadTasks()
            await loadProjects()
            await loadRecurringTasks()
        }
    }

    @MainActor
    public func loadTasks(reset: Bool = true) async {
        if reset {
            isLoading = true
            nextCursor = nil
            hasMore = false
        }
        errorMessage = nil
        do {
            let res = try await apiClient.getTasksPaginated(
                status: selectedStatus,
                search: searchQuery.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty ? nil : searchQuery,
                projectId: selectedProjectId,
                priority: selectedPriority,
                limit: 15,
                cursor: reset ? nil : nextCursor
            )
            var seen = Set<Int64>()
            let combined = reset ? res.items : (self.tasks + res.items)
            self.tasks = combined.filter { seen.insert($0.id).inserted }
            self.nextCursor = res.nextCursor
            self.hasMore = res.hasMore

            AppLogger.info(
                "Tasks loaded successfully", category: .task,
                metadata: ["count": tasks.count, "hasMore": hasMore, "nextCursor": nextCursor ?? -1]
            )
            SharedWidgetDataManager.shared.syncWithTaskList(self.tasks)
            Task {
                await SharedWidgetDataManager.shared.syncWithBackend()
            }

        } catch {
            if error.isTaskOrURLCancelled {
                AppLogger.debug("Tasks load cancelled", category: .task)
            } else {
                AppLogger.error("Failed to load tasks", error: error, category: .task)
                self.errorMessage = L10n.Errors.message(for: error)
            }
        }
        if reset {
            self.isLoading = false
        }
    }

    @MainActor
    public func loadMoreTasksIfNeeded(currentTask: TaskItem? = nil) async {
        guard !isLoading, !isLoadingMore, hasMore, nextCursor != nil else { return }
        if let current = currentTask {
            guard let index = tasks.firstIndex(where: { $0.id == current.id }),
                  index >= tasks.count - 3 else {
                return
            }
        }
        isLoadingMore = true
        do {
            let res = try await apiClient.getTasksPaginated(
                status: selectedStatus,
                search: searchQuery.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty ? nil : searchQuery,
                projectId: selectedProjectId,
                priority: selectedPriority,
                limit: 15,
                cursor: nextCursor
            )
            var seen = Set<Int64>()
            let combined = self.tasks + res.items
            self.tasks = combined.filter { seen.insert($0.id).inserted }
            self.nextCursor = res.nextCursor
            self.hasMore = res.hasMore

            AppLogger.info(
                "More tasks loaded successfully", category: .task,
                metadata: ["count": tasks.count, "hasMore": hasMore, "nextCursor": nextCursor ?? -1]
            )
            SharedWidgetDataManager.shared.syncWithTaskList(self.tasks)
        } catch {
            AppLogger.error("Failed to load more tasks", error: error, category: .task)
        }
        isLoadingMore = false
    }

    @MainActor
    public func selectTask(byId id: Int64) async {
        if selectedTaskForDetail != nil {
            selectedTaskForDetail = nil
            try? await Task.sleep(nanoseconds: 120_000_000)
        }

        if let existing = tasks.first(where: { $0.id == id }) {
            self.selectedTaskForDetail = existing
            return
        }

        if tasks.isEmpty {
            await loadTasks()
            if let existing = tasks.first(where: { $0.id == id }) {
                self.selectedTaskForDetail = existing
                return
            }
        }

        do {
            let task = try await apiClient.getTask(id: id)
            self.selectedTaskForDetail = task
        } catch {
            AppLogger.error(
                "Failed to fetch task by id for deep link",
                error: error,
                category: .task,
                metadata: ["taskId": id]
            )
        }
    }

    @MainActor
    public func loadProjects() async {
        do {
            self.projects = try await apiClient.getProjects()
            AppLogger.debug(
                "Projects loaded for tasks filter", category: .task,
                metadata: ["count": projects.count])
        } catch {
            AppLogger.warn(
                "Failed to load projects for task association", category: .task,
                metadata: ["error": error.localizedDescription])
        }
    }

    public var filteredTasks: [TaskItem] {
        var result = tasks

        // Search Query filter (checks title, description, tags, project name)
        let q = searchQuery.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
        if !q.isEmpty {
            result = result.filter { task in
                task.title.lowercased().contains(q) || task.description.lowercased().contains(q)
                    || (task.tags?.contains(where: { $0.lowercased().contains(q) }) ?? false)
                    || (task.projectName?.lowercased().contains(q) ?? false)
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
            "done": 5,
        ]
        let priorityOrder: [String: Int] = [
            "urgent": 1,
            "high": 2,
            "medium": 3,
            "low": 4,
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
        selectedStatus != nil || selectedProjectId != nil || selectedPriority != nil
            || filterFocusOnly
            || !searchQuery.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
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

    @MainActor
    public func resetFilters() {
        selectedStatus = nil
        selectedProjectId = nil
        selectedPriority = nil
        filterFocusOnly = false
        searchQuery = ""
        Task {
            await loadTasks(reset: true)
        }
    }

    @MainActor
    public func setStatusFilter(_ status: String?) {
        self.selectedStatus = status
        Task {
            await loadTasks(reset: true)
        }
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
    public func updateTaskStatus(task: TaskItem, newStatus: String) {
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
            AppLogger.success(
                "Task created successfully", category: .task, metadata: ["title": title])
            await loadTasks()
            return true
        } catch {
            AppLogger.error("Failed to create task", error: error, category: .task)
            self.errorMessage = error.localizedDescription
            return false
        }
    }

    @MainActor
    public func createTask(
        title: String, description: String, priority: String, projectId: Int64? = nil
    ) {
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

    @discardableResult
    @MainActor
    public func deleteTask(id: Int64) async -> Bool {
        do {
            try await apiClient.deleteTask(id: id)
            tasks.removeAll { $0.id == id }
            SharedWidgetDataManager.shared.syncWithTaskList(self.tasks)
            AppLogger.success("Task deleted successfully", category: .task, metadata: ["id": id])
            await loadTasks()
            return true
        } catch {
            if error.isTaskOrURLCancelled {
                AppLogger.debug("Delete task cancelled", category: .task)
            } else {
                AppLogger.error(
                    "Failed to delete task", error: error, category: .task, metadata: ["id": id])
                self.errorMessage = L10n.Errors.message(for: error)
            }
            return false
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
                AppLogger.success(
                    "Subtask completion toggled", category: .task,
                    metadata: ["taskId": taskId, "subtaskId": subtaskId])
                await loadTasks()
            } catch {
                AppLogger.error("Failed to toggle subtask", error: error, category: .task)
                self.errorMessage = error.localizedDescription
            }
        }
    }

    @MainActor
    public func addSubtask(taskId: Int64, title: String, description: String? = nil, dueDate: String? = nil) async -> Bool {
        let trimmed = title.trimmingCharacters(in: .whitespaces)
        guard !trimmed.isEmpty else { return false }
        do {
            _ = try await apiClient.createSubtask(taskId: taskId, req: CreateSubtaskRequest(title: trimmed, description: description, dueDate: dueDate))
            AppLogger.success("Subtask created", category: .task, metadata: ["taskId": taskId, "title": trimmed])
            await loadTasks()
            return true
        } catch {
            AppLogger.error("Failed to create subtask", error: error, category: .task)
            self.errorMessage = error.localizedDescription
            return false
        }
    }

    @MainActor
    public func updateSubtask(
        taskId: Int64,
        subtaskId: Int64,
        title: String? = nil,
        description: String? = nil,
        isCompleted: Bool? = nil,
        dueDate: String? = nil,
        clearDueDate: Bool = false
    ) async -> Bool {
        do {
            _ = try await apiClient.updateSubtask(
                taskId: taskId,
                subtaskId: subtaskId,
                req: UpdateSubtaskRequest(
                    title: title,
                    description: description,
                    isCompleted: isCompleted,
                    dueDate: dueDate,
                    clearDueDate: clearDueDate
                )
            )
            AppLogger.success("Subtask updated", category: .task, metadata: ["taskId": taskId, "subtaskId": subtaskId])
            await loadTasks()
            return true
        } catch {
            AppLogger.error("Failed to update subtask", error: error, category: .task)
            self.errorMessage = error.localizedDescription
            return false
        }
    }

    @MainActor
    public func deleteSubtask(taskId: Int64, subtaskId: Int64) async -> Bool {
        do {
            try await apiClient.deleteSubtask(taskId: taskId, subtaskId: subtaskId)
            AppLogger.success("Subtask deleted", category: .task, metadata: ["taskId": taskId, "subtaskId": subtaskId])
            await loadTasks()
            return true
        } catch {
            AppLogger.error("Failed to delete subtask", error: error, category: .task)
            self.errorMessage = error.localizedDescription
            return false
        }
    }

    @MainActor
    public func loadActivities(for taskId: Int64) async {
        isLoadingActivities = true
        do {
            self.taskActivities = try await apiClient.getTaskActivities(taskId: taskId)
            AppLogger.debug(
                "Task activities loaded", category: .task,
                metadata: ["taskId": taskId, "count": taskActivities.count])
        } catch {
            AppLogger.error(
                "Failed to load task activities", error: error, category: .task,
                metadata: ["taskId": taskId])
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
                req: CreateTaskActivityRequest(
                    description: note.trimmingCharacters(in: .whitespacesAndNewlines))
            )
            AppLogger.success(
                "Task activity note appended", category: .task, metadata: ["taskId": taskId])
            self.taskActivities.insert(newActivity, at: 0)
        } catch {
            AppLogger.error(
                "Failed to add task activity note", error: error, category: .task,
                metadata: ["taskId": taskId])
            self.errorMessage = error.localizedDescription
        }
    }

    // MARK: - Recurring Task Templates CRUD
    @MainActor
    public func loadRecurringTasks() async {
        isLoadingRecurring = true
        do {
            self.recurringTasks = try await apiClient.getRecurringTasks()
            AppLogger.info(
                "Recurring tasks loaded", category: .task,
                metadata: ["count": recurringTasks.count]
            )
        } catch {
            AppLogger.error("Failed to load recurring tasks", error: error, category: .task)
            self.errorMessage = error.localizedDescription
        }
        isLoadingRecurring = false
    }

    @MainActor
    public func toggleActiveRecurringTask(id: Int64) async {
        do {
            let updated = try await apiClient.toggleActiveRecurringTask(id: id)
            if let index = recurringTasks.firstIndex(where: { $0.id == id }) {
                recurringTasks[index] = updated
            }
            AppLogger.success("Toggled recurring task active state", category: .task, metadata: ["id": id, "isActive": updated.isActive])
        } catch {
            AppLogger.error("Failed to toggle recurring task", error: error, category: .task)
            self.errorMessage = error.localizedDescription
        }
    }

    @MainActor
    public func deleteRecurringTask(id: Int64) async -> Bool {
        do {
            try await apiClient.deleteRecurringTask(id: id)
            recurringTasks.removeAll { $0.id == id }
            AppLogger.success("Deleted recurring task template", category: .task, metadata: ["id": id])
            return true
        } catch {
            AppLogger.error("Failed to delete recurring task", error: error, category: .task)
            self.errorMessage = error.localizedDescription
            return false
        }
    }

    @MainActor
    public func createRecurringTask(req: CreateRecurringTaskRequest) async -> Bool {
        do {
            let item = try await apiClient.createRecurringTask(req: req)
            recurringTasks.insert(item, at: 0)
            AppLogger.success("Created recurring task template", category: .task, metadata: ["id": item.id])
            await loadTasks()
            return true
        } catch {
            AppLogger.error("Failed to create recurring task", error: error, category: .task)
            self.errorMessage = error.localizedDescription
            return false
        }
    }

    @MainActor
    public func updateRecurringTask(id: Int64, req: UpdateRecurringTaskRequest) async -> Bool {
        do {
            let item = try await apiClient.updateRecurringTask(id: id, req: req)
            if let index = recurringTasks.firstIndex(where: { $0.id == id }) {
                recurringTasks[index] = item
            }
            AppLogger.success("Updated recurring task template", category: .task, metadata: ["id": id])
            return true
        } catch {
            AppLogger.error("Failed to update recurring task", error: error, category: .task)
            self.errorMessage = error.localizedDescription
            return false
        }
    }
}
