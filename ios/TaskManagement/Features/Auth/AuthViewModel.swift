import SwiftUI
import AuthenticationServices
import GoogleSignIn

@MainActor
@Observable
public final class AuthViewModel {
    public var isLoading = false
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
                self?.signOut(withMessage: "Sesi telah berakhir atau tidak valid. Silakan login kembali.")
            }
        }
    }

    public func checkSession() {
        if keychain.get(key: "access_token") != nil {
            isAuthenticated = true
        } else {
            isAuthenticated = false
        }
    }

    public func handleAppleSignIn(result: Result<ASAuthorization, Error>) {
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
                    } catch {
                        self.errorMessage = error.localizedDescription
                    }
                    self.isLoading = false
                }
            } else {
                isLoading = false
                errorMessage = "Failed to extract Apple ID Token"
            }
        case .failure(let error):
            isLoading = false
            errorMessage = error.localizedDescription
        }
    }

    public func handleGoogleSignIn() {
        guard let windowScene = UIApplication.shared.connectedScenes.first as? UIWindowScene,
              let rootVC = windowScene.windows.first(where: { $0.isKeyWindow })?.rootViewController ?? windowScene.windows.first?.rootViewController else {
            errorMessage = "Tidak dapat menemukan jendela aplikasi untuk Google Sign-In"
            return
        }

        isLoading = true
        errorMessage = nil

        GIDSignIn.sharedInstance.signIn(withPresenting: rootVC) { [weak self] signInResult, error in
            Task { @MainActor [weak self] in
                guard let self = self else { return }

                if let error = error {
                    let nsError = error as NSError
                    // GIDSignInErrorCodeCanceled = -5
                    if nsError.code != -5 {
                        self.errorMessage = error.localizedDescription
                    }
                    self.isLoading = false
                    return
                }

                guard let user = signInResult?.user,
                      let idToken = user.idToken?.tokenString else {
                    self.errorMessage = "Gagal memverifikasi token Google."
                    self.isLoading = false
                    return
                }

                let email = user.profile?.email
                let name = user.profile?.name ?? "Google User"
                let avatarUrl = user.profile?.imageURL(withDimension: 200)?.absoluteString

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
                } catch {
                    self.errorMessage = error.localizedDescription
                }
                self.isLoading = false
            }
        }
    }

    public func signOut(withMessage message: String? = nil) {
        GIDSignIn.sharedInstance.signOut()
        keychain.clearAll()
        isAuthenticated = false
        if let message {
            errorMessage = message
        }
    }
}
