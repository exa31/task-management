import Foundation
import SwiftUI
import Observation

@Observable
@MainActor
public final class ActiveAlarmManager {
    public static let shared = ActiveAlarmManager()

    public var activeReminder: Reminder? = nil
    public var isProcessingAction: Bool = false
    public var errorMessage: String? = nil

    @ObservationIgnored private let soundManager = SoundManager.shared
    @ObservationIgnored private let apiClient = APIClient.shared
    @ObservationIgnored private let sseClient = SSEClient.shared
    @ObservationIgnored private var sseTask: Task<Void, Never>? = nil

    private init() {}

    /// Starts global SSE listener to intercept reminder events regardless of active tab
    public func startGlobalListener() {
        guard sseTask == nil else { return }
        sseTask = Task { [weak self] in
            for await event in SSEClient.shared.events() {
                guard let self = self else { break }
                if event.type == "reminder:triggered" {
                    AppLogger.info("Global SSE reminder:triggered received: \(event.data)", category: .reminder)
                    self.handleRealtimeReminderEvent(dataString: event.data)
                }
            }
        }
    }

    /// Stops global SSE listener
    public func stopGlobalListener() {
        sseTask?.cancel()
        sseTask = nil
    }

    /// Parses incoming SSE data payload and triggers the alarm
    public func handleRealtimeReminderEvent(dataString: String) {
        guard let data = dataString.data(using: .utf8) else { return }
        let decoder = JSONDecoder()
        if let reminder = try? decoder.decode(Reminder.self, from: data) {
            triggerAlarm(reminder: reminder)
            return
        }

        // Fallback JSON dictionary parser
        if let json = try? JSONSerialization.jsonObject(with: data) as? [String: Any],
           let idVal = json["id"] {
            let id: Int64
            if let num = idVal as? NSNumber {
                id = num.int64Value
            } else if let str = idVal as? String, let parsed = Int64(str) {
                id = parsed
            } else {
                return
            }

            let title = (json["title"] as? String) ?? "Pengingat"
            let desc = (json["description"] as? String) ?? ""
            let remindAt = (json["remind_at"] as? String) ?? ISO8601DateFormatter().string(from: Date())
            let status = (json["status"] as? String) ?? "sent"
            let taskId = (json["task_id"] as? NSNumber)?.int64Value
            let taskTitle = json["task_title"] as? String

            let reminder = Reminder(
                id: id,
                taskId: taskId,
                taskTitle: taskTitle,
                title: title,
                description: desc,
                remindAt: remindAt,
                status: status,
                snoozeCount: 0
            )
            triggerAlarm(reminder: reminder)
        }
    }

    /// Triggers an in-app alarm with ringing audio and haptics
    public func triggerAlarm(reminder: Reminder) {
        self.activeReminder = reminder
        self.errorMessage = nil
        soundManager.startAlarm()
        AppLogger.info("Active alarm triggered for reminder #\(reminder.id): \(reminder.title)", category: .reminder)
    }

    /// Triggers alarm from APNs notification userInfo dictionary
    public func triggerAlarm(fromUserInfo userInfo: [AnyHashable: Any]) {
        var reminderId: Int64?
        if let idNum = userInfo["reminder_id"] as? NSNumber {
            reminderId = idNum.int64Value
        } else if let idStr = userInfo["reminder_id"] as? String, let parsed = Int64(idStr) {
            reminderId = parsed
        }

        guard let rid = reminderId else { return }
        let title = (userInfo["title"] as? String) ?? (userInfo["aps"] as? [String: Any])?["alert"] as? String ?? "Pengingat"
        let taskTitle = userInfo["task_title"] as? String

        var taskId: Int64?
        if let tidNum = userInfo["task_id"] as? NSNumber {
            taskId = tidNum.int64Value
        } else if let tidStr = userInfo["task_id"] as? String, let parsed = Int64(tidStr) {
            taskId = parsed
        }

        let reminder = Reminder(
            id: rid,
            taskId: taskId,
            taskTitle: taskTitle,
            title: title,
            description: (userInfo["body"] as? String) ?? "",
            remindAt: ISO8601DateFormatter().string(from: Date()),
            status: "sent",
            snoozeCount: 0
        )
        triggerAlarm(reminder: reminder)
    }

    /// Snoozes current active alarm
    public func snooze(minutes: Int = 10) async {
        guard let reminder = activeReminder else { return }
        isProcessingAction = true
        soundManager.stopAlarm()

        do {
            _ = try await apiClient.snoozeReminder(id: reminder.id, minutes: minutes)
            AppLogger.success("Reminder #\(reminder.id) snoozed for \(minutes) minutes", category: .reminder)
            await SharedWidgetDataManager.shared.syncWithBackend()
            self.activeReminder = nil
        } catch {
            AppLogger.error("Failed to snooze reminder", error: error, category: .reminder)
            self.errorMessage = error.localizedDescription
            // Re-trigger alarm sound if failed to snooze
            soundManager.startAlarm()
        }
        self.isProcessingAction = false
    }

    /// Dismisses current active alarm
    public func dismiss() async {
        guard let reminder = activeReminder else { return }
        isProcessingAction = true
        soundManager.stopAlarm()

        do {
            _ = try await apiClient.dismissReminder(id: reminder.id)
            AppLogger.success("Reminder #\(reminder.id) dismissed successfully", category: .reminder)
            await SharedWidgetDataManager.shared.syncWithBackend()
            self.activeReminder = nil
        } catch {
            AppLogger.error("Failed to dismiss reminder", error: error, category: .reminder)
            self.errorMessage = error.localizedDescription
            // Re-trigger alarm sound if failed to dismiss
            soundManager.startAlarm()
        }
        self.isProcessingAction = false
    }

    /// Closes the overlay without calling backend
    public func silence() {
        soundManager.stopAlarm()
        self.activeReminder = nil
    }
}
