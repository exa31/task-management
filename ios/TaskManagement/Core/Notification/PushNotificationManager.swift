#if canImport(UIKit)
import UIKit
#elseif canImport(AppKit)
import AppKit
#endif
import UserNotifications

@MainActor
public final class PushNotificationManager: NSObject, UNUserNotificationCenterDelegate {
    public static let shared = PushNotificationManager()

    private let tokenKey = "apns_device_token"
    public private(set) var deviceToken: String?

    private override init() {
        super.init()
        self.deviceToken = UserDefaults.standard.string(forKey: tokenKey)
    }

    /// Sets up notification center delegate, alarm categories, and requests push permissions.
    public func configure() {
        UNUserNotificationCenter.current().delegate = self
        setupNotificationCategories()
        requestAuthorizationAndRegister()
    }

    private func setupNotificationCategories() {
        let snoozeAction = UNNotificationAction(
            identifier: "SNOOZE_10_MIN",
            title: "Tunda 10 Menit",
            options: []
        )
        let dismissAction = UNNotificationAction(
            identifier: "DISMISS_ALARM",
            title: "Selesai / Matikan",
            options: [.destructive]
        )
        let alarmCategory = UNNotificationCategory(
            identifier: "ALARM_REMINDER",
            actions: [snoozeAction, dismissAction],
            intentIdentifiers: [],
            options: [.customDismissAction]
        )
        UNUserNotificationCenter.current().setNotificationCategories([alarmCategory])
    }

    /// Requests user authorization for alerts, badges, and sounds, then registers for remote notifications.
    public func requestAuthorizationAndRegister() {
        UNUserNotificationCenter.current().requestAuthorization(options: [.alert, .badge, .sound, .criticalAlert]) { granted, error in
            Task { @MainActor in
                if let error = error {
                    AppLogger.error("Authorization request failed", error: error, category: .push)
                    return
                }

                if granted {
                    AppLogger.success("Notification permissions granted by user", category: .push)
                    #if canImport(UIKit)
                    UIApplication.shared.registerForRemoteNotifications()
                    #elseif canImport(AppKit)
                    NSApplication.shared.registerForRemoteNotifications()
                    #endif
                } else {
                    AppLogger.info("Notification permissions denied by user", category: .push)
                }
            }
        }
    }

    /// Handles device token received from APNs.
    public func handleDeviceTokenData(_ deviceTokenData: Data) {
        let token = deviceTokenData.map { String(format: "%02.2hhx", $0) }.joined()
        self.deviceToken = token
        UserDefaults.standard.set(token, forKey: tokenKey)

        AppLogger.info("APNs device token received", category: .push, metadata: ["token": token])

        Task {
            await registerDeviceTokenWithBackend(token: token)
        }
    }

    /// Auto-registers the device token to the backend if the user is authenticated.
    public func registerDeviceTokenWithBackend(token: String? = nil) async {
        let resolvedToken = token ?? self.deviceToken ?? UserDefaults.standard.string(forKey: tokenKey)
        guard let validToken = resolvedToken, !validToken.isEmpty else {
            AppLogger.debug("No valid device token available for registration", category: .push)
            return
        }

        guard KeychainManager.shared.get(key: "access_token") != nil else {
            AppLogger.info("Device token cached locally; awaiting user authentication", category: .push)
            return
        }

        do {
            try await APIClient.shared.registerDeviceToken(token: validToken)
            AppLogger.success("Successfully registered device token to backend", category: .push, metadata: ["platform": "ios"])
        } catch {
            AppLogger.warn("Failed to register device token to backend", category: .push, metadata: ["error": error.localizedDescription])
        }
    }

    // MARK: - UNUserNotificationCenterDelegate

    /// Allows notifications to be displayed as banners and play sounds even when the app is active in foreground.
    public nonisolated func userNotificationCenter(
        _ center: UNUserNotificationCenter,
        willPresent notification: UNNotification,
        withCompletionHandler completionHandler: @escaping (UNNotificationPresentationOptions) -> Void
    ) {
        completionHandler([.banner, .badge, .sound])
    }

    /// Handles user tapping on a received notification or interacting with action buttons.
    public nonisolated func userNotificationCenter(
        _ center: UNUserNotificationCenter,
        didReceive response: UNNotificationResponse,
        withCompletionHandler completionHandler: @escaping () -> Void
    ) {
        let userInfo = response.notification.request.content.userInfo
        let actionId = response.actionIdentifier
        AppLogger.info("Notification response received: \(actionId)", category: .push, metadata: ["action": actionId])

        var reminderId: Int64?
        if let idNum = userInfo["reminder_id"] as? NSNumber {
            reminderId = idNum.int64Value
        } else if let idStr = userInfo["reminder_id"] as? String, let parsed = Int64(idStr) {
            reminderId = parsed
        }

        if let rid = reminderId {
            if actionId == "SNOOZE_10_MIN" {
                Task {
                    _ = try? await APIClient.shared.snoozeReminder(id: rid, minutes: 10)
                    AppLogger.success("Reminder \(rid) snoozed for 10 minutes via notification action", category: .push)
                }
            } else if actionId == "DISMISS_ALARM" {
                Task {
                    _ = try? await APIClient.shared.dismissReminder(id: rid)
                    AppLogger.success("Reminder \(rid) dismissed via notification action", category: .push)
                }
            }
        }

        completionHandler()
    }
}
