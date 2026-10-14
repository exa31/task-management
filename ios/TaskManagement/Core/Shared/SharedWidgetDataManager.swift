import Foundation
#if canImport(WidgetKit)
import WidgetKit
#endif

public struct PendingSubtaskToggle: Codable, Sendable {
    public let taskId: Int64
    public let subtaskId: Int64
    public let isCompleted: Bool
    public let timestamp: Date

    public init(taskId: Int64, subtaskId: Int64, isCompleted: Bool, timestamp: Date = Date()) {
        self.taskId = taskId
        self.subtaskId = subtaskId
        self.isCompleted = isCompleted
        self.timestamp = timestamp
    }
}

public final class SharedWidgetDataManager: @unchecked Sendable {
    public static let shared = SharedWidgetDataManager()

    public static let appGroupId = "group.cloud.eka-dev.task-hub"
    private let widgetBundleKey = "task_hub_widget_bundle"
    private let pendingTogglesKey = "task_hub_pending_subtask_toggles"
    private let pendingStatusUpdatesKey = "task_hub_pending_status_updates"
    private let customTaskIndexKey = "task_hub_custom_task_index"
    private let lastSyncKey = "task_hub_widget_last_sync"

    private let defaults: UserDefaults
    private let lock = NSLock()

    public struct PendingStatusUpdate: Codable, Sendable {
        public let taskId: Int64
        public let status: String
        public let timestamp: Date

        public init(taskId: Int64, status: String, timestamp: Date = Date()) {
            self.taskId = taskId
            self.status = status
            self.timestamp = timestamp
        }
    }

    private init() {
        if let groupDefaults = UserDefaults(suiteName: SharedWidgetDataManager.appGroupId) {
            self.defaults = groupDefaults
        } else {
            self.defaults = .standard
        }
    }

    // MARK: - Widget Bundle Persistence

    public func saveWidgetBundle(_ bundle: WidgetBundleResponse) {
        lock.lock()
        defer { lock.unlock() }

        if let encoded = try? JSONEncoder().encode(bundle) {
            defaults.set(encoded, forKey: widgetBundleKey)
            defaults.set(Date().timeIntervalSince1970, forKey: lastSyncKey)
        }

        #if canImport(WidgetKit)
        WidgetCenter.shared.reloadAllTimelines()
        #endif
    }

    public func loadWidgetBundle() -> WidgetBundleResponse? {
        lock.lock()
        defer { lock.unlock() }

        guard let data = defaults.data(forKey: widgetBundleKey),
              let bundle = try? JSONDecoder().decode(WidgetBundleResponse.self, from: data) else {
            return nil
        }
        return bundle
    }

    // MARK: - Offline Subtask Checkoff

    public func toggleSubtaskLocally(taskId: Int64, subtaskId: Int64, isCompleted: Bool) {
        lock.lock()
        defer { lock.unlock() }

        // 1. Immediately mutate cached bundle so widget/UI displays state instantly
        if let data = defaults.data(forKey: widgetBundleKey),
           let bundle = try? JSONDecoder().decode(WidgetBundleResponse.self, from: data) {
            
            var updatedFocus: [TaskItem] = []
            for var task in bundle.focusTasks {
                if task.id == taskId, var subtasks = task.subtasks {
                    for i in 0..<subtasks.count {
                        if subtasks[i].id == subtaskId {
                            subtasks[i].isCompleted = isCompleted
                        }
                    }
                    task.subtasks = subtasks
                    let allCompleted = !subtasks.isEmpty && subtasks.allSatisfy { $0.isCompleted }
                    if allCompleted {
                        task.status = "done"
                        task.isFocus = false
                    } else if !isCompleted && task.status == "done" {
                        task.status = "in_progress"
                    }
                }
                updatedFocus.append(task)
            }

            var updatedDeadlines: [TaskItem] = []
            for var task in bundle.upcomingDeadlines {
                if task.id == taskId, var subtasks = task.subtasks {
                    for i in 0..<subtasks.count {
                        if subtasks[i].id == subtaskId {
                            subtasks[i].isCompleted = isCompleted
                        }
                    }
                    task.subtasks = subtasks
                    let allCompleted = !subtasks.isEmpty && subtasks.allSatisfy { $0.isCompleted }
                    if allCompleted {
                        task.status = "done"
                        task.isFocus = false
                    } else if !isCompleted && task.status == "done" {
                        task.status = "in_progress"
                    }
                }
                updatedDeadlines.append(task)
            }

            let newBundle = WidgetBundleResponse(
                focusTasks: updatedFocus,
                upcomingDeadlines: updatedDeadlines,
                completedToday: bundle.completedToday + (isCompleted ? 1 : -1),
                totalToday: bundle.totalToday
            )

            if let reencoded = try? JSONEncoder().encode(newBundle) {
                defaults.set(reencoded, forKey: widgetBundleKey)
            }
        }

        // 2. Enqueue pending sync for backend
        var pending = getPendingTogglesLocked()
        pending.append(PendingSubtaskToggle(taskId: taskId, subtaskId: subtaskId, isCompleted: isCompleted))
        if let encoded = try? JSONEncoder().encode(pending) {
            defaults.set(encoded, forKey: pendingTogglesKey)
        }

        #if canImport(WidgetKit)
        WidgetCenter.shared.reloadAllTimelines()
        #endif
    }

    public func getPendingToggles() -> [PendingSubtaskToggle] {
        lock.lock()
        defer { lock.unlock() }
        return getPendingTogglesLocked()
    }

    private func getPendingTogglesLocked() -> [PendingSubtaskToggle] {
        guard let data = defaults.data(forKey: pendingTogglesKey),
              let list = try? JSONDecoder().decode([PendingSubtaskToggle].self, from: data) else {
            return []
        }
        return list
    }

    public func clearPendingToggles(olderThan: Date? = nil) {
        lock.lock()
        defer { lock.unlock() }

        if let cutoff = olderThan {
            let remaining = getPendingTogglesLocked().filter { $0.timestamp > cutoff }
            if let encoded = try? JSONEncoder().encode(remaining) {
                defaults.set(encoded, forKey: pendingTogglesKey)
            }
        } else {
            defaults.removeObject(forKey: pendingTogglesKey)
        }
    }

    // MARK: - Task Status Updates Directly from Widget

    public func updateTaskStatusLocally(taskId: Int64, status: String) {
        lock.lock()
        defer { lock.unlock() }

        if let data = defaults.data(forKey: widgetBundleKey),
           let bundle = try? JSONDecoder().decode(WidgetBundleResponse.self, from: data) {

            var updatedFocus: [TaskItem] = []
            for var t in bundle.focusTasks {
                if t.id == taskId {
                    t.status = status
                }
                updatedFocus.append(t)
            }

            var updatedDeadlines: [TaskItem] = []
            for var t in bundle.upcomingDeadlines {
                if t.id == taskId {
                    t.status = status
                }
                if status != "done" || t.id != taskId {
                    updatedDeadlines.append(t)
                }
            }

            let wasDone = bundle.focusTasks.first(where: { $0.id == taskId })?.status == "done"
            let isNowDone = status == "done"
            var completedDelta = 0
            if !wasDone && isNowDone {
                completedDelta = 1
            } else if wasDone && !isNowDone {
                completedDelta = -1
            }

            let newBundle = WidgetBundleResponse(
                focusTasks: updatedFocus,
                upcomingDeadlines: updatedDeadlines,
                completedToday: max(0, bundle.completedToday + completedDelta),
                totalToday: bundle.totalToday
            )

            if let reencoded = try? JSONEncoder().encode(newBundle) {
                defaults.set(reencoded, forKey: widgetBundleKey)
            }
        }

        var pending = getPendingStatusUpdatesLocked()
        pending.append(PendingStatusUpdate(taskId: taskId, status: status))
        if let encoded = try? JSONEncoder().encode(pending) {
            defaults.set(encoded, forKey: pendingStatusUpdatesKey)
        }

        #if canImport(WidgetKit)
        WidgetCenter.shared.reloadAllTimelines()
        #endif
    }

    public func getPendingStatusUpdates() -> [PendingStatusUpdate] {
        lock.lock()
        defer { lock.unlock() }
        return getPendingStatusUpdatesLocked()
    }

    private func getPendingStatusUpdatesLocked() -> [PendingStatusUpdate] {
        guard let data = defaults.data(forKey: pendingStatusUpdatesKey),
              let list = try? JSONDecoder().decode([PendingStatusUpdate].self, from: data) else {
            return []
        }
        return list
    }

    public func clearPendingStatusUpdates(olderThan: Date? = nil) {
        lock.lock()
        defer { lock.unlock() }

        if let cutoff = olderThan {
            let remaining = getPendingStatusUpdatesLocked().filter { $0.timestamp > cutoff }
            if let encoded = try? JSONEncoder().encode(remaining) {
                defaults.set(encoded, forKey: pendingStatusUpdatesKey)
            }
        } else {
            defaults.removeObject(forKey: pendingStatusUpdatesKey)
        }
    }

    // MARK: - Custom Task Cycling

    public func cycleCustomTaskIndex() {
        lock.lock()
        defer { lock.unlock() }
        let current = defaults.integer(forKey: customTaskIndexKey)
        defaults.set(current + 1, forKey: customTaskIndexKey)

        #if canImport(WidgetKit)
        WidgetCenter.shared.reloadAllTimelines()
        #endif
    }

    public func getCustomTaskIndex() -> Int {
        lock.lock()
        defer { lock.unlock() }
        return defaults.integer(forKey: customTaskIndexKey)
    }

    // MARK: - Synchronous Task List Sync (Instant Local Sync)

    public func syncWithTaskList(_ tasks: [TaskItem]) {
        lock.lock()
        defer { lock.unlock() }

        let focus = tasks.filter { ($0.isFocus == true || $0.isRecurring == true) && $0.status != "done" }
        let deadlines = tasks
            .filter { $0.status != "done" && $0.dueDate != nil && !($0.dueDate?.isEmpty ?? true) }
            .sorted { ($0.dueDate ?? "") < ($1.dueDate ?? "") }

        let todayDone = tasks.filter { $0.status == "done" }.count
        let total = tasks.count

        let bundle = WidgetBundleResponse(
            focusTasks: focus,
            upcomingDeadlines: Array(deadlines.prefix(10)),
            completedToday: todayDone,
            totalToday: total
        )

        if let encoded = try? JSONEncoder().encode(bundle) {
            defaults.set(encoded, forKey: widgetBundleKey)
            defaults.set(Date().timeIntervalSince1970, forKey: lastSyncKey)
        }

        #if canImport(WidgetKit)
        WidgetCenter.shared.reloadAllTimelines()
        #endif
    }

    // MARK: - Online Sync

    public func syncWithBackend() async {
        let pending = getPendingToggles()
        if !pending.isEmpty {
            AppLogger.info("Flushing \(pending.count) offline widget subtask toggles to backend", category: .task)
            let syncTime = Date()
            for item in pending {
                do {
                    _ = try await APIClient.shared.updateSubtask(
                        taskId: item.taskId,
                        subtaskId: item.subtaskId,
                        isCompleted: item.isCompleted
                    )
                } catch {
                    AppLogger.warn("Failed to sync subtask \(item.subtaskId) toggle: \(error.localizedDescription)", category: .task)
                }
            }
            clearPendingToggles(olderThan: syncTime)
        }

        let pendingStatuses = getPendingStatusUpdates()
        if !pendingStatuses.isEmpty {
            AppLogger.info("Flushing \(pendingStatuses.count) widget task status updates to backend", category: .task)
            let syncTime = Date()
            for item in pendingStatuses {
                do {
                    _ = try await APIClient.shared.updateTask(
                        id: item.taskId,
                        req: UpdateTaskRequest(status: item.status)
                    )
                } catch {
                    AppLogger.warn("Failed to sync task \(item.taskId) status update: \(error.localizedDescription)", category: .task)
                }
            }
            clearPendingStatusUpdates(olderThan: syncTime)
        }

        do {
            let bundle = try await APIClient.shared.getWidgetBundle()
            saveWidgetBundle(bundle)
            AppLogger.success("Widget bundle updated from network", category: .task)
        } catch {
            AppLogger.warn("Failed to fetch fresh widget bundle: \(error.localizedDescription)", category: .task)
        }
    }
}
