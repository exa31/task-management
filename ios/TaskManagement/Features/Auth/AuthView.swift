import SwiftUI
import AuthenticationServices

public struct AuthView: View {
    @Bindable var viewModel: AuthViewModel
    let onAuthSuccess: () -> Void

    public init(viewModel: AuthViewModel, onAuthSuccess: @escaping () -> Void) {
        self.viewModel = viewModel
        self.onAuthSuccess = onAuthSuccess
    }

    public var body: some View {
        ZStack {
            // Ambient Radial Gradient Background
            Color(.systemBackground)
                .ignoresSafeArea()

            RadialGradient(
                gradient: Gradient(colors: [
                    AppColors.primary.opacity(0.18),
                    AppColors.violet.opacity(0.12),
                    Color.clear
                ]),
                center: .top,
                startRadius: 40,
                endRadius: 450
            )
            .ignoresSafeArea()

            VStack(spacing: 0) {
                // Top Header Bar: Language Switcher Button (Powered by I18n Engine)
                HStack {
                    Spacer()
                    Button(action: {
                        I18n.shared.toggleLanguage()
                    }) {
                        HStack(spacing: 6) {
                            Image(systemName: "globe")
                                .font(.system(size: 11, weight: .semibold))
                            Text(I18n.shared.currentLanguage.displayName)
                                .font(.system(size: 12, weight: .semibold))
                            Text(I18n.shared.currentLanguage.flag)
                                .font(.system(size: 12))
                        }
                        .padding(.horizontal, 12)
                        .padding(.vertical, 7)
                        .background(Color(.secondarySystemGroupedBackground).opacity(0.85))
                        .foregroundColor(.primary)
                        .clipShape(Capsule())
                        .overlay(
                            Capsule().stroke(Color.primary.opacity(0.1), lineWidth: 1)
                        )
                        .shadow(color: Color.black.opacity(0.04), radius: 4, x: 0, y: 2)
                    }
                    .buttonStyle(.plain)
                }
                .padding(.horizontal, 20)
                .padding(.top, 12)

                ScrollView(showsIndicators: false) {
                    VStack(spacing: 28) {
                        Spacer().frame(height: 24)

                        // Brand Logo & Title
                        VStack(spacing: 16) {
                            BrandLogoView(size: 80)

                            VStack(spacing: 6) {
                                Text(L10n.Auth.title)
                                    .font(.system(size: 32, weight: .bold, design: .rounded))
                                    .foregroundStyle(
                                        LinearGradient(
                                            colors: [AppColors.indigo, AppColors.violet],
                                            startPoint: .leading,
                                            endPoint: .trailing
                                        )
                                    )

                                Text(L10n.Auth.subtitle)
                                    .font(.system(size: 15))
                                    .foregroundColor(.secondary)
                            }
                        }

                        // Feature highlights pill box
                        VStack(spacing: 12) {
                            FeatureHighlightRow(
                                icon: "bolt.fill",
                                title: L10n.Auth.syncTitle,
                                desc: L10n.Auth.syncDesc,
                                color: AppColors.primary
                            )
                            FeatureHighlightRow(
                                icon: "folder.fill",
                                title: L10n.Auth.projectsTitle,
                                desc: L10n.Auth.projectsDesc,
                                color: AppColors.violet
                            )
                            FeatureHighlightRow(
                                icon: "bell.badge.fill",
                                title: L10n.Auth.remindersTitle,
                                desc: L10n.Auth.remindersDesc,
                                color: AppColors.priorityHigh
                            )
                        }
                        .padding(18)
                        .glassCardStyle()
                        .padding(.horizontal, 24)

                        // Error Message Banner
                        if let error = viewModel.errorMessage {
                            HStack(spacing: 8) {
                                Image(systemName: "exclamationmark.triangle.fill")
                                    .foregroundColor(.red)
                                Text(error)
                                    .font(.system(size: 13))
                                    .foregroundColor(.red)
                            }
                            .padding()
                            .frame(maxWidth: .infinity)
                            .background(Color.red.opacity(0.1))
                            .clipShape(RoundedRectangle(cornerRadius: 12))
                            .padding(.horizontal, 24)
                        }

                        // Sign-In Actions
                        VStack(spacing: 14) {
                            // Sign in with Apple Button
                            ZStack {
                                SignInWithAppleButton(
                                    onRequest: { request in
                                        request.requestedScopes = [.fullName, .email]
                                    },
                                    onCompletion: { result in
                                        viewModel.handleAppleSignIn(result: result)
                                    }
                                )
                                .signInWithAppleButtonStyle(.black)
                                .frame(height: 52)
                                .clipShape(RoundedRectangle(cornerRadius: 14, style: .continuous))
                                .shadow(color: Color.black.opacity(0.12), radius: 6, x: 0, y: 3)
                                .environment(\.locale, Locale(identifier: I18n.shared.currentLanguage.rawValue))
                                .disabled(viewModel.isLoading)

                                if viewModel.isLoading {
                                    RoundedRectangle(cornerRadius: 14, style: .continuous)
                                        .fill(Color.black.opacity(0.65))
                                        .frame(height: 52)
                                        .overlay(
                                            ProgressView()
                                                .tint(.white)
                                        )
                                }
                            }

                            // Google Sign-In Button
                            Button(action: {
                                viewModel.handleGoogleSignIn()
                            }) {
                                HStack(spacing: 10) {
                                    if viewModel.isLoading {
                                        ProgressView()
                                            .tint(.primary)
                                    } else {
                                        Image(systemName: "g.circle.fill")
                                            .font(.system(size: 20, weight: .bold))
                                            .foregroundColor(AppColors.primary)
                                    }

                                    Text(L10n.Auth.signInGoogle)
                                        .font(.system(size: 16, weight: .semibold))
                                }
                                .frame(maxWidth: .infinity)
                                .frame(height: 52)
                                .background(Color(.secondarySystemGroupedBackground))
                                .foregroundColor(.primary)
                                .clipShape(RoundedRectangle(cornerRadius: 14, style: .continuous))
                                .overlay(
                                    RoundedRectangle(cornerRadius: 14, style: .continuous)
                                        .stroke(Color.secondary.opacity(0.2), lineWidth: 1)
                                 )
                                .shadow(color: Color.black.opacity(0.04), radius: 4, x: 0, y: 2)
                            }
                            .buttonStyle(.plain)
                            .disabled(viewModel.isLoading)
                        }
                        .padding(.horizontal, 24)

                        // Footer secure indicator (clean, non-technical)
                        HStack(spacing: 6) {
                            Circle()
                                .fill(Color.emeraldGreen)
                                .frame(width: 8, height: 8)
                            Text(L10n.Auth.secure)
                                .font(.system(size: 11, weight: .medium))
                                .foregroundColor(.secondary)
                        }
                        .padding(.top, 10)

                        Spacer().frame(height: 30)
                    }
                }
            }
        }
        .onChange(of: viewModel.isAuthenticated) { _, newValue in
            if newValue {
                onAuthSuccess()
            }
        }
    }
}

struct FeatureHighlightRow: View {
    let icon: String
    let title: String
    let desc: String
    let color: Color

    var body: some View {
        HStack(spacing: 12) {
            ZStack {
                Circle()
                    .fill(color.opacity(0.15))
                    .frame(width: 34, height: 34)
                Image(systemName: icon)
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundColor(color)
            }

            VStack(alignment: .leading, spacing: 2) {
                Text(title)
                    .font(.system(size: 14, weight: .semibold))
                Text(desc)
                    .font(.system(size: 12))
                    .foregroundColor(.secondary)
            }
            Spacer()
        }
    }
}

extension Color {
    static let emeraldGreen = Color(red: 16/255, green: 185/255, blue: 129/255)
}
