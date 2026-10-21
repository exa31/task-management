#if canImport(UIKit)
import UIKit
#elseif canImport(AppKit)
import AppKit
#endif
import UserNotifications
import FirebaseCore
import FirebaseMessaging

@MainActor
public final class PushNotificationManager: NSObject, UNUserNotificationCenterDelegate, MessagingDelegate {
    public static let shared = PushNotificationManager()

    private let tokenKey = "fcm_device_token"
    public private(set) var deviceToken: String?

    private override init() {
        super.init()
        if let cached = UserDefaults.standard.string(forKey: tokenKey), cached.count > 80 {
            self.deviceToken = cached
        }
    }

    /// Sets up notification center delegate, alarm categories, Firebase Messaging, and requests push permissions.
    public func configure() {
        if FirebaseApp.app() == nil {
            FirebaseApp.configure()
        }
        Messaging.messaging().delegate = self
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

    /// Handles device token received from APNs, forwards to Firebase Messaging, and retrieves the FCM token.
    public func handleDeviceTokenData(_ deviceTokenData: Data) {
        Messaging.messaging().apnsToken = deviceTokenData
        let apnsHex = deviceTokenData.map { String(format: "%02.2hhx", $0) }.joined()
        AppLogger.info("APNs device token forwarded to Firebase Messaging", category: .push, metadata: ["apns_token": apnsHex])

        Messaging.messaging().token { [weak self] fcmToken, error in
            Task { @MainActor in
                if let error = error {
                    AppLogger.warn("Failed to retrieve FCM token from Firebase", category: .push, metadata: ["error": error.localizedDescription])
                    return
                }
                if let token = fcmToken, !token.isEmpty {
                    self?.handleFCMRegistrationToken(token)
                }
            }
        }
    }

    /// Handles FCM registration token received from Firebase Messaging.
    public func handleFCMRegistrationToken(_ token: String) {
        guard !token.isEmpty else { return }
        self.deviceToken = token
        UserDefaults.standard.set(token, forKey: tokenKey)

        let preview = token.count > 15 ? String(token.prefix(15)) + "..." : token
        AppLogger.info("FCM registration token received", category: .push, metadata: ["token_preview": preview])

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

        let hasToken = KeychainManager.shared.get(key: "access_token") != nil ||
                       SharedWidgetDataManager.shared.getAuthToken() != nil ||
                       KeychainManager.shared.get(key: "refresh_token") != nil ||
                       SharedWidgetDataManager.shared.getRefreshToken() != nil

        guard hasToken else {
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

    /// Allows notifications to be displayed as banners and triggers the in-app alarm overlay when app is active in foreground.
    public nonisolated func userNotificationCenter(
        _ center: UNUserNotificationCenter,
        willPresent notification: UNNotification,
        withCompletionHandler completionHandler: @escaping (UNNotificationPresentationOptions) -> Void
    ) {
        let userInfo = notification.request.content.userInfo
        Task { @MainActor in
            await SharedWidgetDataManager.shared.syncWithBackend()
            if userInfo["reminder_id"] != nil || (userInfo["category"] as? String) == "ALARM_REMINDER" {
                ActiveAlarmManager.shared.triggerAlarm(fromUserInfo: userInfo)
            }
        }
        completionHandler([.banner, .badge, .sound])
    }

    /// Handles user tapping on a received notification or interacting with action buttons.
    public nonisolated func userNotificationCenter(
        _ center: UNUserNotificationCenter,
        didReceive response: UNNotificationResponse,
        withCompletionHandler completionHandler: @escaping () -> Void
    ) {
        Task {
            await SharedWidgetDataManager.shared.syncWithBackend()
        }
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
                Task { @MainActor in
                    SoundManager.shared.stopAlarm()
                    _ = try? await APIClient.shared.snoozeReminder(id: rid, minutes: 10)
                    AppLogger.success("Reminder \(rid) snoozed for 10 minutes via notification action", category: .push)
                }
            } else if actionId == "DISMISS_ALARM" {
                Task { @MainActor in
                    SoundManager.shared.stopAlarm()
                    _ = try? await APIClient.shared.dismissReminder(id: rid)
                    AppLogger.success("Reminder \(rid) dismissed via notification action", category: .push)
                }
            } else if actionId == UNNotificationDefaultActionIdentifier {
                Task { @MainActor in
                    ActiveAlarmManager.shared.triggerAlarm(fromUserInfo: userInfo)
                }
            }
        }

        completionHandler()
    }

    // MARK: - MessagingDelegate

    /// Called by Firebase Messaging when a new FCM registration token is generated or refreshed.
    public nonisolated func messaging(_ messaging: Messaging, didReceiveRegistrationToken fcmToken: String?) {
        guard let token = fcmToken, !token.isEmpty else { return }
        Task { @MainActor in
            PushNotificationManager.shared.handleFCMRegistrationToken(token)
        }
    }
}

