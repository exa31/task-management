import SwiftUI
import GoogleSignIn

/*
Entry Point iOS (SwiftUI Modern iOS 17+):
@main menandai struct ini sebagai gerbang awal eksekusi aplikasi oleh sistem operasi iOS.
*/
@main
struct TaskManagementApp: App {
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
        }
    }
}
