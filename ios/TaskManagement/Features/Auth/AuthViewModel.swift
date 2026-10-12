import SwiftUI
import AuthenticationServices

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
            self?.signOut(withMessage: "Sesi telah berakhir atau tidak valid. Silakan login kembali.")
        }
    }

    public func checkSession() {
        if keychain.get(key: "access_token") != nil {
            isAuthenticated = true
        } else {
            isAuthenticated = false
        }
    }

    @MainActor
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

    @MainActor
    public func handleGoogleSignInSimulated() {
        isLoading = true
        errorMessage = nil

        Task {
            do {
                _ = try await apiClient.oauthLogin(
                    req: OAuthLoginRequest(
                        provider: "google",
                        idToken: "mock_google_id_token",
                        code: nil,
                        email: "user@gmail.com",
                        name: "Google User",
                        avatarUrl: nil
                    )
                )
                self.isAuthenticated = true
            } catch {
                self.errorMessage = error.localizedDescription
            }
            self.isLoading = false
        }
    }

    public func signOut(withMessage message: String? = nil) {
        keychain.clearAll()
        isAuthenticated = false
        if let message {
            errorMessage = message
        }
    }
}
