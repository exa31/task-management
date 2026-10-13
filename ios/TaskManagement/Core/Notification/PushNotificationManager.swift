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

    /// Sets up notification center delegate and requests push permissions.
    public func configure() {
        UNUserNotificationCenter.current().delegate = self
        requestAuthorizationAndRegister()
    }

    /// Requests user authorization for alerts, badges, and sounds, then registers for remote notifications.
    public func requestAuthorizationAndRegister() {
        UNUserNotificationCenter.current().requestAuthorization(options: [.alert, .badge, .sound]) { granted, error in
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

    /// Handles user tapping on a received notification.
    public nonisolated func userNotificationCenter(
        _ center: UNUserNotificationCenter,
        didReceive response: UNNotificationResponse,
        withCompletionHandler completionHandler: @escaping () -> Void
    ) {
        let userInfo = response.notification.request.content.userInfo
        AppLogger.info("User tapped notification", category: .push, metadata: ["keys": userInfo.keys.map(String.init).joined(separator: ",")])
        completionHandler()
    }
}
