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
    private let lastSyncKey = "task_hub_widget_last_sync"

    private let defaults: UserDefaults
    private let lock = NSLock()

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

        do {
            let bundle = try await APIClient.shared.getWidgetBundle()
            saveWidgetBundle(bundle)
            AppLogger.success("Widget bundle updated from network", category: .task)
        } catch {
            AppLogger.warn("Failed to fetch fresh widget bundle: \(error.localizedDescription)", category: .task)
        }
    }
}
