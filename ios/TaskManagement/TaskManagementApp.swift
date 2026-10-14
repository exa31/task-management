import SwiftUI
import GoogleSignIn

#if canImport(UIKit)
import UIKit
final class AppDelegate: NSObject, UIApplicationDelegate {
    func application(
        _ application: UIApplication,
        didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]? = nil
    ) -> Bool {
        Task { @MainActor in
            PushNotificationManager.shared.configure()
            TaskHubShortcutsProvider.updateAppShortcutParameters()
        }
        return true
    }

    func application(
        _ application: UIApplication,
        didRegisterForRemoteNotificationsWithDeviceToken deviceToken: Data
    ) {
        Task { @MainActor in
            PushNotificationManager.shared.handleDeviceTokenData(deviceToken)
        }
    }

    func application(
        _ application: UIApplication,
        didFailToRegisterForRemoteNotificationsWithError error: Error
    ) {
        AppLogger.warn("Failed to register for remote notifications", category: .push, metadata: ["error": error.localizedDescription])
    }

    func application(
        _ application: UIApplication,
        didReceiveRemoteNotification userInfo: [AnyHashable: Any],
        fetchCompletionHandler completionHandler: @escaping (UIBackgroundFetchResult) -> Void
    ) {
        Task {
            AppLogger.info("Remote notification received in background; syncing widget data", category: .push)
            await SharedWidgetDataManager.shared.syncWithBackend()
            completionHandler(.newData)
        }
    }
}
#elseif canImport(AppKit)
import AppKit
final class AppDelegate: NSObject, NSApplicationDelegate {
    func applicationDidFinishLaunching(_ notification: Notification) {
        Task { @MainActor in
            PushNotificationManager.shared.configure()
        }
    }

    func application(_ application: NSApplication, didRegisterForRemoteNotificationsWithDeviceToken deviceToken: Data) {
        Task { @MainActor in
            PushNotificationManager.shared.handleDeviceTokenData(deviceToken)
        }
    }

    func application(_ application: NSApplication, didFailToRegisterForRemoteNotificationsWithError error: Error) {
        AppLogger.warn("Failed to register for remote notifications", category: .push, metadata: ["error": error.localizedDescription])
    }
}
#endif

/*
Entry Point iOS / macOS (SwiftUI Modern iOS 17+ / macOS 14+):
@main menandai struct ini sebagai gerbang awal eksekusi aplikasi.
*/
@main
struct TaskManagementApp: App {
    #if canImport(UIKit)
    @UIApplicationDelegateAdaptor(AppDelegate.self) private var appDelegate
    #elseif canImport(AppKit)
    @NSApplicationDelegateAdaptor(AppDelegate.self) private var appDelegate
    #endif
    @Environment(\.scenePhase) private var scenePhase
    @State private var authViewModel = AuthViewModel()
    @AppStorage("app_appearance") private var appAppearance: String = "system"

    private var colorScheme: ColorScheme? {
        switch appAppearance {
        case "light":
            return .light
        case "dark":
            return .dark
        default:
            return nil
        }
    }

    var body: some Scene {
        WindowGroup {
            Group {
                if authViewModel.isAuthenticated {
                    MainTabView(authViewModel: authViewModel)
                } else {
                    AuthView(viewModel: authViewModel, onAuthSuccess: {
                        authViewModel.isAuthenticated = true
                    })
                }
            }
            .preferredColorScheme(colorScheme)
            .environment(\.locale, Locale(identifier: I18n.shared.currentLanguage.rawValue))
            .onOpenURL { url in
                _ = GIDSignIn.sharedInstance.handle(url)
            }
            .onChange(of: authViewModel.isAuthenticated) { _, isAuthenticated in
                if isAuthenticated {
                    Task {
                        await PushNotificationManager.shared.registerDeviceTokenWithBackend()
                    }
                }
            }
            .onChange(of: scenePhase) { _, newPhase in
                if newPhase == .active && authViewModel.isAuthenticated {
                    Task {
                        await SharedWidgetDataManager.shared.syncWithBackend()
                    }
                }
            }
        }
    }
}
