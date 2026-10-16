import SwiftUI
import AuthenticationServices
import GoogleSignIn

@MainActor
@Observable
public final class AuthViewModel {
    public var isLoading = false
    public var isAppleLoading = false
    public var isGoogleLoading = false
    public var errorMessage: String?
    public var isAuthenticated = false

    private let apiClient = APIClient.shared
    private let keychain = KeychainManager.shared

    public init() {
        checkSession()
        setupNotificationObservers()
    }

    private func setupNotificationObservers() {
        NotificationCenter.default.addObserver(
            forName: .sessionExpired,
            object: nil,
            queue: .main
        ) { [weak self] _ in
            Task { @MainActor [weak self] in
                AppLogger.warn("Session expired notification received; signing out user", category: .auth)
                self?.signOut(withMessage: L10n.Auth.sessionExpired)
            }
        }
    }

    public func checkSession() {
        if let token = keychain.get(key: "access_token") {
            isAuthenticated = true
            SharedWidgetDataManager.shared.saveAuthToken(token)
            AppLogger.info("Active session token found in Keychain", category: .auth)
            Task {
                await PushNotificationManager.shared.registerDeviceTokenWithBackend()
            }
        } else {
            isAuthenticated = false
            SharedWidgetDataManager.shared.clearAuthToken()
        }
    }

    public func handleAppleSignIn(result: Result<ASAuthorization, Error>) {
        isAppleLoading = true
        isLoading = true
        errorMessage = nil

        switch result {
        case .success(let auth):
            if let appleIDCredential = auth.credential as? ASAuthorizationAppleIDCredential,
               let tokenData = appleIDCredential.identityToken,
               let idToken = String(data: tokenData, encoding: .utf8) {

                let fullName = [
                    appleIDCredential.fullName?.givenName,
                    appleIDCredential.fullName?.familyName
                ].compactMap { $0 }.joined(separator: " ")

                AppLogger.info("Initiating Apple Sign-In exchange with backend", category: .auth)
                Task {
                    do {
                        _ = try await apiClient.oauthLogin(
                            req: OAuthLoginRequest(
                                provider: "apple",
                                idToken: idToken,
                                code: nil,
                                email: appleIDCredential.email,
                                name: fullName.isEmpty ? "Apple User" : fullName,
                                avatarUrl: nil
                            )
                        )
                        self.isAuthenticated = true
                        AppLogger.success("Apple Sign-In completed successfully", category: .auth)
                        await PushNotificationManager.shared.registerDeviceTokenWithBackend()
                    } catch {
                        AppLogger.error("Apple Sign-In backend verification failed", error: error, category: .auth)
                        self.errorMessage = error.localizedDescription
                    }
                    self.isAppleLoading = false
                    self.isLoading = false
                }
            } else {
                AppLogger.error("Failed to extract Apple ID Token from authorization credential", category: .auth)
                isAppleLoading = false
                isLoading = false
                errorMessage = L10n.Auth.appleTokenFailed
            }
        case .failure(let error):
            let authError = error as? ASAuthorizationError
            if authError?.code != .canceled {
                AppLogger.warn("Apple Sign-In authorization flow cancelled or failed", category: .auth, metadata: ["error": error.localizedDescription])
                errorMessage = error.localizedDescription
            } else {
                AppLogger.info("User cancelled Apple Sign-In sheet", category: .auth)
            }
            isAppleLoading = false
            isLoading = false
        }
    }

    public func handleGoogleSignIn() {
        guard let windowScene = UIApplication.shared.connectedScenes.first as? UIWindowScene,
              let rootVC = windowScene.windows.first(where: { $0.isKeyWindow })?.rootViewController ?? windowScene.windows.first?.rootViewController else {
            AppLogger.warn("Could not find key window view controller for Google Sign-In presentation", category: .auth)
            errorMessage = L10n.Auth.windowNotFound
            return
        }

        isGoogleLoading = true
        isLoading = true
        errorMessage = nil
        AppLogger.info("Presenting Google Sign-In dialog", category: .auth)

        GIDSignIn.sharedInstance.signIn(withPresenting: rootVC) { [weak self] signInResult, error in
            Task { @MainActor [weak self] in
                guard let self = self else { return }

                if let error = error {
                    let nsError = error as NSError
                    // GIDSignInErrorCodeCanceled = -5
                    if nsError.code != -5 {
                        AppLogger.error("Google Sign-In presented error", error: error, category: .auth)
                        self.errorMessage = error.localizedDescription
                    } else {
                        AppLogger.info("User cancelled Google Sign-In dialog", category: .auth)
                    }
                    self.isGoogleLoading = false
                    self.isLoading = false
                    return
                }

                guard let user = signInResult?.user,
                      let idToken = user.idToken?.tokenString else {
                    AppLogger.error("Google Sign-In returned empty user or idToken", category: .auth)
                    self.errorMessage = L10n.Auth.googleTokenFailed
                    self.isGoogleLoading = false
                    self.isLoading = false
                    return
                }

                let email = user.profile?.email
                let name = user.profile?.name ?? "Google User"
                let avatarUrl = user.profile?.imageURL(withDimension: 200)?.absoluteString

                AppLogger.info("Exchanging Google ID token with backend", category: .auth)
                do {
                    _ = try await self.apiClient.oauthLogin(
                        req: OAuthLoginRequest(
                            provider: "google",
                            idToken: idToken,
                            code: nil,
                            email: email,
                            name: name,
                            avatarUrl: avatarUrl
                        )
                    )
                    self.isAuthenticated = true
                    AppLogger.success("Google Sign-In completed successfully", category: .auth)
                    await PushNotificationManager.shared.registerDeviceTokenWithBackend()
                } catch {
                    AppLogger.error("Google Sign-In backend verification failed", error: error, category: .auth)
                    self.errorMessage = error.localizedDescription
                }
                self.isGoogleLoading = false
                self.isLoading = false
            }
        }
    }

    public func signOut(withMessage message: String? = nil) {
        GIDSignIn.sharedInstance.signOut()
        keychain.clearAll()
        SharedWidgetDataManager.shared.clearAuthToken()
        isAuthenticated = false
        if let message {
            errorMessage = message
        }
    }
}
