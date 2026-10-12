import UIKit
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
                    #if DEBUG
                    print("⚠️ [PushNotification] Authorization request error: \(error.localizedDescription)")
                    #endif
                    return
                }

                if granted {
                    #if DEBUG
                    print("✅ [PushNotification] User granted notification permissions.")
                    #endif
                    UIApplication.shared.registerForRemoteNotifications()
                } else {
                    #if DEBUG
                    print("ℹ️ [PushNotification] User denied notification permissions.")
                    #endif
                }
            }
        }
    }

    /// Handles device token received from APNs.
    public func handleDeviceTokenData(_ deviceTokenData: Data) {
        let token = deviceTokenData.map { String(format: "%02.2hhx", $0) }.joined()
        self.deviceToken = token
        UserDefaults.standard.set(token, forKey: tokenKey)

        #if DEBUG
        print("📲 [PushNotification] Received APNs device token: \(token)")
        #endif

        Task {
            await registerDeviceTokenWithBackend(token: token)
        }
    }

    /// Auto-registers the device token to the backend if the user is authenticated.
    public func registerDeviceTokenWithBackend(token: String? = nil) async {
        let resolvedToken = token ?? self.deviceToken ?? UserDefaults.standard.string(forKey: tokenKey)
        guard let validToken = resolvedToken, !validToken.isEmpty else {
            #if DEBUG
            print("ℹ️ [PushNotification] No valid device token available for registration.")
            #endif
            return
        }

        guard KeychainManager.shared.get(key: "access_token") != nil else {
            #if DEBUG
            print("ℹ️ [PushNotification] Device token cached locally. Waiting for user login before registering to backend.")
            #endif
            return
        }

        do {
            try await APIClient.shared.registerDeviceToken(token: validToken)
            #if DEBUG
            print("🚀 [PushNotification] Successfully registered device token to backend.")
            #endif
        } catch {
            #if DEBUG
            print("⚠️ [PushNotification] Failed to register device token to backend: \(error.localizedDescription)")
            #endif
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
        #if DEBUG
        print("🔔 [PushNotification] User tapped on notification with payload: \(userInfo)")
        #endif
        completionHandler()
    }
}
