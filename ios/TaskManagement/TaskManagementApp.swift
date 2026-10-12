import SwiftUI
import GoogleSignIn

final class AppDelegate: NSObject, UIApplicationDelegate {
    func application(
        _ application: UIApplication,
        didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]? = nil
    ) -> Bool {
        Task { @MainActor in
            PushNotificationManager.shared.configure()
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
        #if DEBUG
        print("⚠️ [PushNotification] Failed to register for remote notifications: \(error.localizedDescription)")
        #endif
    }
}

/*
Entry Point iOS (SwiftUI Modern iOS 17+):
@main menandai struct ini sebagai gerbang awal eksekusi aplikasi oleh sistem operasi iOS.
*/
@main
struct TaskManagementApp: App {
    @UIApplicationDelegateAdaptor(AppDelegate.self) private var appDelegate
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
        }
    }
}
