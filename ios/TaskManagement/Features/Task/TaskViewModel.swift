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
        } catch {
            self.errorMessage = error.localizedDescription
        }
        self.isLoading = false
    }

    @MainActor
    public func loadProjects() async {
        do {
            self.projects = try await apiClient.getProjects()
        } catch {
            // Silently handled
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

        return result
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
    public func createTask(
        title: String,
        description: String = "",
        status: String = "todo",
        priority: String = "medium",
        dueDate: String? = nil,
        isFocus: Bool = false,
        tags: [String] = [],
        projectId: Int64? = nil,
        subtasks: [String] = []
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
                    subtasks: subtasks
                )
            )
            await loadTasks()
            return true
        } catch {
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
            await loadTasks()
        } catch {
            self.errorMessage = error.localizedDescription
        }
    }

    @MainActor
    public func toggleSubtask(taskId: Int64, subtaskId: Int64, currentCompleted: Bool) {
        Task {
            do {
                _ = try await apiClient.updateSubtask(
                    taskId: taskId,
                    subtaskId: subtaskId,
                    isCompleted: !currentCompleted
                )
                await loadTasks()
            } catch {
                self.errorMessage = error.localizedDescription
            }
        }
    }

    @MainActor
    public func loadActivities(for taskId: Int64) async {
        isLoadingActivities = true
        do {
            self.taskActivities = try await apiClient.getTaskActivities(taskId: taskId)
        } catch {
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
            self.taskActivities.insert(newActivity, at: 0)
        } catch {
            self.errorMessage = error.localizedDescription
        }
    }
}
