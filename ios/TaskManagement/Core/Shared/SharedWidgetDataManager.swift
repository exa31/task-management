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
    private let recurringTaskIndexKey = "task_hub_recurring_task_index"
    private let selectedRecurringTaskIdsKey = "task_hub_selected_recurring_task_ids"
    private let selectedCustomTaskIdsKey = "task_hub_selected_custom_task_ids"
    private let autoSelectAllKey = "task_hub_widget_auto_select_all"
    private let excludedTaskIdsKey = "task_hub_widget_excluded_task_ids"
    private let lastSyncKey = "task_hub_widget_last_sync"
    private let authTokenKey = "task_hub_shared_auth_token"
    private let refreshTokenKey = "task_hub_shared_refresh_token"

    private let defaults: UserDefaults
    private let lock = NSLock()

    // MARK: - Shared Auth Token Access

    public func saveAuthToken(_ token: String) {
        lock.lock()
        defer { lock.unlock() }
        defaults.set(token, forKey: authTokenKey)
    }

    public func getAuthToken() -> String? {
        lock.lock()
        defer { lock.unlock() }
        return defaults.string(forKey: authTokenKey)
    }

    public func clearAuthToken() {
        lock.lock()
        defer { lock.unlock() }
        defaults.removeObject(forKey: authTokenKey)
    }

    public func saveRefreshToken(_ token: String) {
        lock.lock()
        defer { lock.unlock() }
        defaults.set(token, forKey: refreshTokenKey)
    }

    public func getRefreshToken() -> String? {
        lock.lock()
        defer { lock.unlock() }
        return defaults.string(forKey: refreshTokenKey)
    }

    public func clearRefreshToken() {
        lock.lock()
        defer { lock.unlock() }
        defaults.removeObject(forKey: refreshTokenKey)
    }

    public func saveTokens(accessToken: String, refreshToken: String) {
        lock.lock()
        defer { lock.unlock() }
        defaults.set(accessToken, forKey: authTokenKey)
        defaults.set(refreshToken, forKey: refreshTokenKey)
    }

    public func clearAllTokens() {
        lock.lock()
        defer { lock.unlock() }
        defaults.removeObject(forKey: authTokenKey)
        defaults.removeObject(forKey: refreshTokenKey)
    }

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

        // Sanitize: ensure no done tasks are stored in focusTasks or upcomingDeadlines
        let sanitizedFocus = bundle.focusTasks.filter { $0.status != "done" }
        let sanitizedDeadlines = bundle.upcomingDeadlines.filter { $0.status != "done" }
        let sanitizedBundle = WidgetBundleResponse(
            focusTasks: sanitizedFocus,
            upcomingDeadlines: sanitizedDeadlines,
            completedToday: bundle.completedToday,
            totalToday: bundle.totalToday
        )

        if let encoded = try? JSONEncoder().encode(sanitizedBundle) {
            defaults.set(encoded, forKey: widgetBundleKey)
            defaults.set(Date().timeIntervalSince1970, forKey: lastSyncKey)
        }

        // Dynamic Matching: Reconcile active tasks into selection if auto-select is enabled
        let isAutoSelect = (defaults.object(forKey: autoSelectAllKey) == nil) ? true : defaults.bool(forKey: autoSelectAllKey)
        if isAutoSelect {
            let excluded = Set(defaults.array(forKey: excludedTaskIdsKey) as? [Int64] ?? [])
            var rec = Set(defaults.array(forKey: selectedRecurringTaskIdsKey) as? [Int64] ?? [])
            var cus = Set(defaults.array(forKey: selectedCustomTaskIdsKey) as? [Int64] ?? [])
            for t in sanitizedFocus + sanitizedDeadlines {
                if !excluded.contains(t.id) {
                    if t.isRecurring == true || t.recurringTaskId != nil {
                        rec.insert(t.id)
                        if let rid = t.recurringTaskId {
                            rec.insert(rid)
                        }
                    } else {
                        cus.insert(t.id)
                    }
                }
            }
            defaults.set(Array(rec), forKey: selectedRecurringTaskIdsKey)
            defaults.set(Array(cus), forKey: selectedCustomTaskIdsKey)
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
        // In case previous cache contained completed tasks, filter them out on load
        let cleanFocus = bundle.focusTasks.filter { $0.status != "done" }
        let cleanDeadlines = bundle.upcomingDeadlines.filter { $0.status != "done" }
        return WidgetBundleResponse(
            focusTasks: cleanFocus,
            upcomingDeadlines: cleanDeadlines,
            completedToday: bundle.completedToday,
            totalToday: bundle.totalToday
        )
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
                    let completedCount = subtasks.filter { $0.isCompleted }.count
                    task.subtasksCompleted = completedCount
                    task.progressPercent = subtasks.isEmpty ? 0 : Int(Double(completedCount) / Double(subtasks.count) * 100)
                }
                if task.status != "done" {
                    updatedFocus.append(task)
                }
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
                    let completedCount = subtasks.filter { $0.isCompleted }.count
                    task.subtasksCompleted = completedCount
                    task.progressPercent = subtasks.isEmpty ? 0 : Int(Double(completedCount) / Double(subtasks.count) * 100)
                }
                if task.status != "done" {
                    updatedDeadlines.append(task)
                }
            }

            let newBundle = WidgetBundleResponse(
                focusTasks: updatedFocus,
                upcomingDeadlines: updatedDeadlines,
                completedToday: bundle.completedToday,
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

            let wasDone = bundle.focusTasks.first(where: { $0.id == taskId })?.status == "done"
                || bundle.upcomingDeadlines.first(where: { $0.id == taskId })?.status == "done"
            let isNowDone = status == "done"
            var completedDelta = 0
            if !wasDone && isNowDone {
                completedDelta = 1
            } else if wasDone && !isNowDone {
                completedDelta = -1
            }

            var updatedFocus: [TaskItem] = []
            for var t in bundle.focusTasks {
                if t.id == taskId {
                    t.status = status
                }
                if t.status != "done" {
                    updatedFocus.append(t)
                }
            }

            var updatedDeadlines: [TaskItem] = []
            for var t in bundle.upcomingDeadlines {
                if t.id == taskId {
                    t.status = status
                }
                if t.status != "done" {
                    updatedDeadlines.append(t)
                }
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

            if status == "done" {
                var rec = Set(defaults.array(forKey: selectedRecurringTaskIdsKey) as? [Int64] ?? [])
                if rec.remove(taskId) != nil {
                    defaults.set(Array(rec), forKey: selectedRecurringTaskIdsKey)
                }
                var cus = Set(defaults.array(forKey: selectedCustomTaskIdsKey) as? [Int64] ?? [])
                if cus.remove(taskId) != nil {
                    defaults.set(Array(cus), forKey: selectedCustomTaskIdsKey)
                }
                var excl = Set(defaults.array(forKey: excludedTaskIdsKey) as? [Int64] ?? [])
                if excl.remove(taskId) != nil {
                    defaults.set(Array(excl), forKey: excludedTaskIdsKey)
                }
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

    // MARK: - Recurring Task Cycling

    public func cycleRecurringTaskIndex() {
        lock.lock()
        defer { lock.unlock() }
        let current = defaults.integer(forKey: recurringTaskIndexKey)
        defaults.set(current + 1, forKey: recurringTaskIndexKey)

        #if canImport(WidgetKit)
        WidgetCenter.shared.reloadAllTimelines()
        #endif
    }

    public func getRecurringTaskIndex() -> Int {
        lock.lock()
        defer { lock.unlock() }
        return defaults.integer(forKey: recurringTaskIndexKey)
    }

    // MARK: - BCA-Style Widget Selection Management

    public func getSelectedRecurringTaskIds() -> Set<Int64> {
        lock.lock()
        defer { lock.unlock() }
        let arr = defaults.array(forKey: selectedRecurringTaskIdsKey) as? [Int64] ?? []
        return Set(arr)
    }

    public func setSelectedRecurringTaskIds(_ ids: Set<Int64>) {
        lock.lock()
        defer { lock.unlock() }
        defaults.set(Array(ids), forKey: selectedRecurringTaskIdsKey)
        #if canImport(WidgetKit)
        WidgetCenter.shared.reloadAllTimelines()
        #endif
    }

    public func toggleSelectedRecurringTaskId(_ id: Int64) {
        lock.lock()
        defer { lock.unlock() }
        var current = Set(defaults.array(forKey: selectedRecurringTaskIdsKey) as? [Int64] ?? [])
        if current.contains(id) {
            current.remove(id)
        } else {
            current.insert(id)
        }
        defaults.set(Array(current), forKey: selectedRecurringTaskIdsKey)
        #if canImport(WidgetKit)
        WidgetCenter.shared.reloadAllTimelines()
        #endif
    }

    public func getSelectedCustomTaskIds() -> Set<Int64> {
        lock.lock()
        defer { lock.unlock() }
        let arr = defaults.array(forKey: selectedCustomTaskIdsKey) as? [Int64] ?? []
        return Set(arr)
    }

    public func setSelectedCustomTaskIds(_ ids: Set<Int64>) {
        lock.lock()
        defer { lock.unlock() }
        defaults.set(Array(ids), forKey: selectedCustomTaskIdsKey)
        #if canImport(WidgetKit)
        WidgetCenter.shared.reloadAllTimelines()
        #endif
    }

    public func toggleSelectedCustomTaskId(_ id: Int64) {
        lock.lock()
        defer { lock.unlock() }
        var current = Set(defaults.array(forKey: selectedCustomTaskIdsKey) as? [Int64] ?? [])
        if current.contains(id) {
            current.remove(id)
        } else {
            current.insert(id)
        }
        defaults.set(Array(current), forKey: selectedCustomTaskIdsKey)
        #if canImport(WidgetKit)
        WidgetCenter.shared.reloadAllTimelines()
        #endif
    }

    // MARK: - Subtask Pagination per Task

    public func getSubtaskPage(for taskId: Int64, totalPages: Int) -> Int {
        lock.lock()
        defer { lock.unlock() }
        guard totalPages > 1 else { return 0 }
        let page = defaults.integer(forKey: "task_hub_subtask_page_\(taskId)")
        return (page % totalPages + totalPages) % totalPages
    }

    public func stepSubtaskPage(taskId: Int64, delta: Int, totalPages: Int) {
        lock.lock()
        defer { lock.unlock() }
        guard totalPages > 1 else { return }
        let key = "task_hub_subtask_page_\(taskId)"
        let current = defaults.integer(forKey: key)
        let next = ((current + delta) % totalPages + totalPages) % totalPages
        defaults.set(next, forKey: key)
    }

    // MARK: - Auto-Select All & Dynamic Matching

    public func isAutoSelectAllEnabled() -> Bool {
        lock.lock()
        defer { lock.unlock() }
        if defaults.object(forKey: autoSelectAllKey) == nil {
            return true
        }
        return defaults.bool(forKey: autoSelectAllKey)
    }

    public func setAutoSelectAllEnabled(_ enabled: Bool) {
        lock.lock()
        defer { lock.unlock() }
        defaults.set(enabled, forKey: autoSelectAllKey)
        if enabled {
            defaults.removeObject(forKey: excludedTaskIdsKey)
        }
        #if canImport(WidgetKit)
        WidgetCenter.shared.reloadAllTimelines()
        #endif
    }

    public func getExcludedTaskIds() -> Set<Int64> {
        lock.lock()
        defer { lock.unlock() }
        let arr = defaults.array(forKey: excludedTaskIdsKey) as? [Int64] ?? []
        return Set(arr)
    }

    public func setExcludedTaskIds(_ ids: Set<Int64>) {
        lock.lock()
        defer { lock.unlock() }
        defaults.set(Array(ids), forKey: excludedTaskIdsKey)
        #if canImport(WidgetKit)
        WidgetCenter.shared.reloadAllTimelines()
        #endif
    }

    public func toggleExcludedTaskId(_ id: Int64) {
        lock.lock()
        defer { lock.unlock() }
        var current = Set(defaults.array(forKey: excludedTaskIdsKey) as? [Int64] ?? [])
        if current.contains(id) {
            current.remove(id)
        } else {
            current.insert(id)
        }
        defaults.set(Array(current), forKey: excludedTaskIdsKey)
        #if canImport(WidgetKit)
        WidgetCenter.shared.reloadAllTimelines()
        #endif
    }

    // MARK: - Synchronous Task List Sync (Instant Local Sync)

    public func syncWithTaskList(_ tasks: [TaskItem]) {
        lock.lock()
        defer { lock.unlock() }

        var selectedRec = Set(defaults.array(forKey: selectedRecurringTaskIdsKey) as? [Int64] ?? [])
        var selectedCus = Set(defaults.array(forKey: selectedCustomTaskIdsKey) as? [Int64] ?? [])
        var excluded = Set(defaults.array(forKey: excludedTaskIdsKey) as? [Int64] ?? [])

        // Clean up any tasks that are now done
        let doneIds = Set(tasks.filter { $0.status == "done" }.map { $0.id })
        selectedRec.subtract(doneIds)
        selectedCus.subtract(doneIds)
        excluded.subtract(doneIds)
        defaults.set(Array(excluded), forKey: excludedTaskIdsKey)

        let isAutoSelect = (defaults.object(forKey: autoSelectAllKey) == nil) ? true : defaults.bool(forKey: autoSelectAllKey)

        if isAutoSelect {
            // Dynamic Matching: Automatically select all active tasks that are not excluded
            for task in tasks where task.status != "done" && !excluded.contains(task.id) {
                if task.isRecurring == true || task.recurringTaskId != nil {
                    selectedRec.insert(task.id)
                    if let rid = task.recurringTaskId {
                        selectedRec.insert(rid)
                    }
                } else {
                    selectedCus.insert(task.id)
                }
            }
        }

        defaults.set(Array(selectedRec), forKey: selectedRecurringTaskIdsKey)
        defaults.set(Array(selectedCus), forKey: selectedCustomTaskIdsKey)

        let focus = tasks.filter { 
            $0.status != "done" && (
                isAutoSelect ? !excluded.contains($0.id) : (
                    $0.isFocus == true || 
                    $0.isRecurring == true || 
                    $0.recurringTaskId != nil ||
                    selectedRec.contains($0.id) || 
                    ($0.recurringTaskId != nil && selectedRec.contains($0.recurringTaskId!)) ||
                    selectedCus.contains($0.id)
                )
            ) 
        }
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

    public func clearPendingToggles(matching items: [PendingSubtaskToggle]) {
        lock.lock()
        defer { lock.unlock() }
        let set = Set(items.map { "\($0.taskId)_\($0.subtaskId)_\($0.isCompleted)_\($0.timestamp.timeIntervalSince1970)" })
        let remaining = getPendingTogglesLocked().filter {
            !set.contains("\($0.taskId)_\($0.subtaskId)_\($0.isCompleted)_\($0.timestamp.timeIntervalSince1970)")
        }
        if let encoded = try? JSONEncoder().encode(remaining) {
            defaults.set(encoded, forKey: pendingTogglesKey)
        }
    }

    public func clearPendingStatusUpdates(matching items: [PendingStatusUpdate]) {
        lock.lock()
        defer { lock.unlock() }
        let set = Set(items.map { "\($0.taskId)_\($0.status)_\($0.timestamp.timeIntervalSince1970)" })
        let remaining = getPendingStatusUpdatesLocked().filter {
            !set.contains("\($0.taskId)_\($0.status)_\($0.timestamp.timeIntervalSince1970)")
        }
        if let encoded = try? JSONEncoder().encode(remaining) {
            defaults.set(encoded, forKey: pendingStatusUpdatesKey)
        }
    }

    // MARK: - Online Sync

    public func syncWithBackend() async {
        let pending = getPendingToggles()
        if !pending.isEmpty {
            AppLogger.info("Flushing \(pending.count) offline widget subtask toggles to backend", category: .task)
            var synced: [PendingSubtaskToggle] = []
            for item in pending {
                do {
                    _ = try await APIClient.shared.updateSubtask(
                        taskId: item.taskId,
                        subtaskId: item.subtaskId,
                        isCompleted: item.isCompleted
                    )
                    synced.append(item)
                } catch {
                    AppLogger.warn("Failed to sync subtask \(item.subtaskId) toggle: \(error.localizedDescription)", category: .task)
                }
            }
            if !synced.isEmpty {
                clearPendingToggles(matching: synced)
            }
        }

        let pendingStatuses = getPendingStatusUpdates()
        if !pendingStatuses.isEmpty {
            AppLogger.info("Flushing \(pendingStatuses.count) widget task status updates to backend", category: .task)
            var syncedStatuses: [PendingStatusUpdate] = []
            for item in pendingStatuses {
                do {
                    _ = try await APIClient.shared.updateTask(
                        id: item.taskId,
                        req: UpdateTaskRequest(status: item.status)
                    )
                    syncedStatuses.append(item)
                } catch {
                    AppLogger.warn("Failed to sync task \(item.taskId) status update: \(error.localizedDescription)", category: .task)
                }
            }
            if !syncedStatuses.isEmpty {
                clearPendingStatusUpdates(matching: syncedStatuses)
            }
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
