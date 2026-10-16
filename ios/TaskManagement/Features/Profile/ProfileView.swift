import SwiftUI
import PhotosUI
import AuthenticationServices

@Observable
public final class ProfileViewModel {
    public var user: User? = nil
    public var isLoading = false
    public var isSavingProfile = false
    public var isSavingNotifications = false
    public var isUploadingAvatar = false
    public var isDeletingAccount = false
    public var errorMessage: String?
    public var successToast: String?

    // Profile form
    public var name: String = ""
    public var email: String = ""

    // Notification preferences
    public var timezone: String = "Asia/Jakarta"
    public var threshold: Int = 1440
    public var emailNotificationsEnabled: Bool = true

    // OAuth binding status
    public var googleBound: Bool = false
    public var appleBound: Bool = false
    public var isAppleBinding: Bool = false

    private let apiClient = APIClient.shared
    private let keychain = KeychainManager.shared

    public init() {
        Task { await loadProfile() }
    }

    @MainActor
    public func loadProfile() async {
        isLoading = true
        errorMessage = nil
        do {
            let profile = try await apiClient.getProfile()
            self.user = profile
            self.name = profile.name
            self.email = profile.email
            self.timezone = profile.timezone ?? "Asia/Jakarta"
            self.threshold = profile.dueDateWarningThreshold
            self.emailNotificationsEnabled = profile.emailNotificationsEnabled
            self.googleBound = profile.googleBound
            self.appleBound = profile.appleBound
        } catch {
            self.errorMessage = error.localizedDescription
            self.name = keychain.get(key: "user_name") ?? ""
            self.email = keychain.get(key: "user_email") ?? ""
        }
        self.isLoading = false
    }

    @MainActor
    public func saveProfile() async -> Bool {
        guard !name.trimmingCharacters(in: .whitespaces).isEmpty else { return false }
        isSavingProfile = true
        errorMessage = nil
        successToast = nil
        do {
            let updated = try await apiClient.updateProfile(
                req: UpdateProfileRequest(name: name.trimmingCharacters(in: .whitespaces))
            )
            self.user = updated
            self.name = updated.name
            self.successToast = L10n.Settings.profileUpdated
            AppLogger.success("Profile details updated successfully", category: .app)
            self.isSavingProfile = false
            return true
        } catch {
            AppLogger.error("Failed to update profile", error: error, category: .app)
            self.errorMessage = error.localizedDescription
            self.isSavingProfile = false
            return false
        }
    }

    @MainActor
    public func uploadAvatar(imageData: Data) async {
        isUploadingAvatar = true
        errorMessage = nil
        successToast = nil
        do {
            let updated = try await apiClient.uploadAvatar(data: imageData)
            self.user = updated
            self.successToast = L10n.Settings.avatarUpdated
            AppLogger.success("Avatar image uploaded successfully", category: .app)
        } catch {
            AppLogger.error("Failed to upload avatar", error: error, category: .app)
            self.errorMessage = error.localizedDescription
        }
        self.isUploadingAvatar = false
    }

    @MainActor
    public func saveNotifications() async -> Bool {
        isSavingNotifications = true
        errorMessage = nil
        successToast = nil
        do {
            let updated = try await apiClient.updateProfile(
                req: UpdateProfileRequest(
                    timezone: timezone,
                    dueDateWarningThreshold: threshold,
                    emailNotificationsEnabled: emailNotificationsEnabled
                )
            )
            self.user = updated
            self.successToast = L10n.Settings.preferencesSaved
            AppLogger.success("Notification preferences updated successfully", category: .app)
            self.isSavingNotifications = false
            return true
        } catch {
            AppLogger.error("Failed to update notification preferences", error: error, category: .app)
            self.errorMessage = error.localizedDescription
            self.isSavingNotifications = false
            return false
        }
    }

    @MainActor
    public func unbindOAuth(provider: String) async -> Bool {
        errorMessage = nil
        successToast = nil
        do {
            let updated = try await apiClient.unbindOAuth(provider: provider)
            self.user = updated
            self.googleBound = updated.googleBound
            self.appleBound = updated.appleBound
            self.successToast = L10n.Settings.accountDisconnected(provider: provider.capitalized)
            AppLogger.success("OAuth provider unlinked successfully", category: .auth, metadata: ["provider": provider])
            return true
        } catch {
            AppLogger.error("Failed to unbind OAuth provider", error: error, category: .auth, metadata: ["provider": provider])
            self.errorMessage = error.localizedDescription
            return false
        }
    }

    @MainActor
    public func handleAppleBind(result: Result<ASAuthorization, Error>) {
        switch result {
        case .success(let auth):
            if let credential = auth.credential as? ASAuthorizationAppleIDCredential,
               let tokenData = credential.identityToken,
               let idToken = String(data: tokenData, encoding: .utf8) {
                let fullName = [credential.fullName?.givenName, credential.fullName?.familyName]
                    .compactMap { $0 }.joined(separator: " ")
                isAppleBinding = true
                Task {
                    do {
                        let updated = try await apiClient.bindOAuth(
                            req: BindOAuthRequest(
                                provider: "apple",
                                idToken: idToken,
                                code: nil,
                                email: credential.email,
                                name: fullName.isEmpty ? nil : fullName
                            )
                        )
                        self.user = updated
                        self.appleBound = updated.appleBound
                        self.successToast = L10n.Settings.appleConnected
                        AppLogger.success("Apple ID linked successfully to profile", category: .auth)
                    } catch {
                        AppLogger.error("Failed to bind Apple ID", error: error, category: .auth)
                        self.errorMessage = error.localizedDescription
                    }
                    self.isAppleBinding = false
                }
            }
        case .failure(let error):
            let authError = error as? ASAuthorizationError
            if authError?.code != .canceled {
                self.errorMessage = error.localizedDescription
            } else {
                AppLogger.info("User cancelled Apple ID bind sheet", category: .auth)
            }
        }
    }

    @MainActor
    public func deleteAccount(confirmation: String) async -> Bool {
        guard confirmation.trimmingCharacters(in: .whitespaces).uppercased() == "DELETE" else {
            self.errorMessage = L10n.Settings.deleteConfirmPrompt
            return false
        }
        isDeletingAccount = true
        errorMessage = nil
        do {
            try await apiClient.deleteAccount(confirmation: "DELETE")
            self.isDeletingAccount = false
            return true
        } catch {
            self.errorMessage = error.localizedDescription
            self.isDeletingAccount = false
            return false
        }
    }
}

public struct ProfileView: View {
    let onSignOut: () -> Void
    @State private var viewModel = ProfileViewModel()
    @State private var selectedPhotoItem: PhotosPickerItem? = nil

    // Unbind confirmation state
    @State private var providerToUnbind: String? = nil
    @State private var showUnbindConfirmation = false

    // Sign out confirmation state
    @State private var showSignOutConfirmation = false

    // Danger Zone Delete Modal state
    @State private var showDeleteModal = false
    @State private var deleteConfirmText = ""

    @AppStorage("app_appearance") private var appAppearance: String = "system"
    @AppStorage("app_language") private var appLanguage: String = "en"

    public init(onSignOut: @escaping () -> Void) {
        self.onSignOut = onSignOut
    }

    public var body: some View {
        AdaptiveTabNavigationStack {
            ScrollView {
                VStack(spacing: 20) {
                    // Success / Error Banners
                    if let success = viewModel.successToast {
                        HStack(spacing: 8) {
                            Image(systemName: "checkmark.circle.fill")
                                .foregroundColor(AppColors.statusDone)
                            Text(success)
                                .font(.system(size: 13, weight: .medium))
                                .foregroundColor(AppColors.statusDone)
                            Spacer()
                            Button(action: { viewModel.successToast = nil }) {
                                Image(systemName: "xmark")
                                    .font(.system(size: 11))
                                    .foregroundColor(AppColors.statusDone)
                            }
                        }
                        .padding(12)
                        .background(AppColors.statusDone.opacity(0.12))
                        .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
                    }

                    if let error = viewModel.errorMessage {
                        HStack(spacing: 8) {
                            Image(systemName: "exclamationmark.triangle.fill")
                                .foregroundColor(AppColors.priorityUrgent)
                            Text(error)
                                .font(.system(size: 13, weight: .medium))
                                .foregroundColor(AppColors.priorityUrgent)
                            Spacer()
                            Button(action: { viewModel.errorMessage = nil }) {
                                Image(systemName: "xmark")
                                    .font(.system(size: 11))
                                    .foregroundColor(AppColors.priorityUrgent)
                            }
                        }
                        .padding(12)
                        .background(AppColors.priorityUrgent.opacity(0.12))
                        .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
                    }

                    // 1. Profile Information Section
                    profileInformationSection

                    // 2. Linked Accounts Section (Google & Apple)
                    linkedAccountsSection

                    // 3. Notification & Due Date Warning Preferences
                    notificationPreferencesSection

                    // 4. Appearance & Language (Dark Mode & Locale)
                    appearanceAndLanguageSection

                    // 5. Widget Configuration (BCA-Style Task Hub Widgets)
                    widgetConfigurationSection

                    // 6. AI & MCP Integration
                    mcpIntegrationSection

                    // 7. Danger Zone (Permanent Account Deletion)
                    dangerZoneSection

                    // 7. Sign Out Button
                    Button(action: { showSignOutConfirmation = true }) {
                        HStack(spacing: 8) {
                            Image(systemName: "rectangle.portrait.and.arrow.right")
                            Text(L10n.Settings.signOut)
                        }
                        .font(.system(size: 15, weight: .semibold))
                        .foregroundColor(.secondary)
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 14)
                        .background(Color(.secondarySystemGroupedBackground))
                        .clipShape(RoundedRectangle(cornerRadius: 14, style: .continuous))
                    }
                    .buttonStyle(.plain)
                    .padding(.top, 4)
                }
                .padding(.horizontal, 16)
                .padding(.vertical, 14)
            }
            .background(Color(.systemGroupedBackground))
            .navigationTitle(L10n.Settings.title)
            .navigationBarTitleDisplayMode(.inline)
            .refreshable {
                await viewModel.loadProfile()
            }
            .confirmationDialog(
                L10n.Settings.disconnectDialogTitle,
                isPresented: $showUnbindConfirmation,
                titleVisibility: .visible,
                presenting: providerToUnbind
            ) { provider in
                Button(L10n.Settings.disconnectConfirmButton(provider: provider.capitalized), role: .destructive) {
                    Task {
                        _ = await viewModel.unbindOAuth(provider: provider)
                    }
                }
                Button(L10n.General.cancel, role: .cancel) {}
            } message: { provider in
                Text(L10n.Settings.disconnectDialogMessage(provider: provider.capitalized))
            }
            .confirmationDialog(
                L10n.Settings.signOutDialogTitle,
                isPresented: $showSignOutConfirmation,
                titleVisibility: .visible
            ) {
                Button(L10n.Settings.signOutConfirmButton, role: .destructive) {
                    onSignOut()
                }
                Button(L10n.General.cancel, role: .cancel) {}
            } message: {
                Text(L10n.Settings.signOutDialogMessage)
            }
            .sheet(isPresented: $showDeleteModal) {
                deleteAccountModalView
            }
            .onChange(of: selectedPhotoItem) { _, newItem in
                Task {
                    if let data = try? await newItem?.loadTransferable(type: Data.self) {
                        await viewModel.uploadAvatar(imageData: data)
                    }
                }
            }
        }
    }

    // MARK: - Section 1: Profile Information
    private var profileInformationSection: some View {
        VStack(alignment: .leading, spacing: 16) {
            Label(L10n.Settings.profileInfo, systemImage: "person.crop.circle")
                .font(.system(size: 16, weight: .bold))
                .foregroundColor(.primary)

            // Avatar & Header
            HStack(spacing: 16) {
                ZStack(alignment: .bottomTrailing) {
                    Group {
                        if let avatarUrlString = viewModel.user?.avatarUrl,
                           let url = URL(string: avatarUrlString) {
                            AsyncImage(url: url) { phase in
                                switch phase {
                                case .success(let image):
                                    image
                                        .resizable()
                                        .scaledToFill()
                                case .failure, .empty:
                                    avatarFallbackView
                                @unknown default:
                                    avatarFallbackView
                                }
                            }
                        } else {
                            avatarFallbackView
                        }
                    }
                    .frame(width: 72, height: 72)
                    .clipShape(Circle())
                    .overlay(
                        Circle()
                            .stroke(AppColors.primary.opacity(0.3), lineWidth: 2)
                    )

                    // Photo Picker Overlay Button
                    PhotosPicker(selection: $selectedPhotoItem, matching: .images) {
                        ZStack {
                            Circle()
                                .fill(AppColors.primary)
                                .frame(width: 26, height: 26)
                            if viewModel.isUploadingAvatar {
                                ProgressView()
                                    .scaleEffect(0.6)
                                    .tint(.white)
                            } else {
                                Image(systemName: "camera.fill")
                                    .font(.system(size: 11))
                                    .foregroundColor(.white)
                            }
                        }
                    }
                    .buttonStyle(.plain)
                }

                VStack(alignment: .leading, spacing: 4) {
                    Text(viewModel.user?.name ?? viewModel.name)
                        .font(.system(size: 17, weight: .bold))
                        .foregroundColor(.primary)
                    Text(viewModel.user?.email ?? viewModel.email)
                        .font(.system(size: 13))
                        .foregroundColor(.secondary)

                    PhotosPicker(selection: $selectedPhotoItem, matching: .images) {
                        HStack(spacing: 4) {
                            Image(systemName: "arrow.triangle.2.circlepath")
                                .font(.system(size: 10))
                            Text(L10n.Settings.changeAvatar)
                                .font(.system(size: 11, weight: .semibold))
                        }
                        .padding(.horizontal, 8)
                        .padding(.vertical, 4)
                        .background(AppColors.primary.opacity(0.12))
                        .foregroundColor(AppColors.primary)
                        .clipShape(Capsule())
                    }
                    .padding(.top, 2)
                }

                Spacer()
            }

            Divider()

            // Name Edit Field
            VStack(alignment: .leading, spacing: 6) {
                Text(L10n.Settings.fullName)
                    .font(.system(size: 12, weight: .medium))
                    .foregroundColor(.secondary)

                TextField(L10n.Settings.fullName, text: $viewModel.name)
                    .font(.system(size: 14))
                    .padding(10)
                    .background(Color(.tertiarySystemGroupedBackground))
                    .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))
            }

            // Email Field (Read Only)
            VStack(alignment: .leading, spacing: 6) {
                HStack {
                    Text(L10n.Settings.emailAddress)
                        .font(.system(size: 12, weight: .medium))
                        .foregroundColor(.secondary)
                    Spacer()
                    Image(systemName: "lock.fill")
                        .font(.system(size: 10))
                        .foregroundColor(.secondary)
                }

                Text(viewModel.email.isEmpty ? "-" : viewModel.email)
                    .font(.system(size: 14))
                    .foregroundColor(.secondary)
                    .padding(10)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .background(Color(.tertiarySystemGroupedBackground).opacity(0.6))
                    .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))
            }

            // Save Profile Button
            Button(action: {
                Task { _ = await viewModel.saveProfile() }
            }) {
                HStack(spacing: 6) {
                    if viewModel.isSavingProfile {
                        ProgressView()
                            .scaleEffect(0.8)
                            .tint(.white)
                    } else {
                        Image(systemName: "checkmark")
                            .font(.system(size: 13, weight: .bold))
                        Text(L10n.Settings.saveProfile)
                            .font(.system(size: 14, weight: .semibold))
                    }
                }
                .frame(maxWidth: .infinity)
                .padding(.vertical, 11)
                .background(AppColors.brandGradient)
                .foregroundColor(.white)
                .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
            }
            .buttonStyle(.plain)
            .disabled(viewModel.isSavingProfile || viewModel.name.trimmingCharacters(in: .whitespaces).isEmpty || viewModel.name == (viewModel.user?.name ?? ""))
            .opacity(viewModel.name == (viewModel.user?.name ?? "") ? 0.6 : 1.0)
        }
        .padding(16)
        .background(Color(.secondarySystemGroupedBackground))
        .clipShape(RoundedRectangle(cornerRadius: 16, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: 16, style: .continuous)
                .stroke(Color.primary.opacity(0.06), lineWidth: 1)
        )
    }

    private var avatarFallbackView: some View {
        ZStack {
            AppColors.brandGradient
            if let firstChar = (viewModel.user?.name.first ?? viewModel.name.first) {
                Text(String(firstChar).uppercased())
                    .font(.system(size: 28, weight: .bold))
                    .foregroundColor(.white)
            } else {
                Image(systemName: "person.fill")
                    .font(.system(size: 28))
                    .foregroundColor(.white)
            }
        }
    }

    // MARK: - Section 2: Linked Accounts (Google & Apple)
    private var linkedAccountsSection: some View {
        VStack(alignment: .leading, spacing: 14) {
            VStack(alignment: .leading, spacing: 2) {
                Label(L10n.Settings.linkedAccounts, systemImage: "link")
                    .font(.system(size: 16, weight: .bold))
                    .foregroundColor(.primary)
                Text(L10n.Settings.linkedAccountsDesc)
                    .font(.system(size: 12))
                    .foregroundColor(.secondary)
            }

            // Google Card
            HStack(spacing: 12) {
                ZStack {
                    RoundedRectangle(cornerRadius: 10, style: .continuous)
                        .fill(Color(.tertiarySystemGroupedBackground))
                        .frame(width: 40, height: 40)
                    Text("G")
                        .font(.system(size: 20, weight: .heavy, design: .rounded))
                        .foregroundColor(.blue)
                }

                VStack(alignment: .leading, spacing: 2) {
                    Text("Google Account")
                        .font(.system(size: 15, weight: .semibold))
                    HStack(spacing: 5) {
                        Circle()
                            .fill(viewModel.googleBound ? AppColors.statusDone : Color.secondary)
                            .frame(width: 6, height: 6)
                        Text(viewModel.googleBound ? L10n.Settings.connected : L10n.Settings.notConnected)
                            .font(.system(size: 12))
                            .foregroundColor(.secondary)
                    }
                }

                Spacer()

                if viewModel.googleBound {
                    Button(L10n.Settings.disconnect) {
                        providerToUnbind = "google"
                        showUnbindConfirmation = true
                    }
                    .font(.system(size: 12, weight: .semibold))
                    .foregroundColor(AppColors.priorityUrgent)
                    .padding(.horizontal, 10)
                    .padding(.vertical, 5)
                    .background(AppColors.priorityUrgent.opacity(0.12))
                    .clipShape(Capsule())
                }
            }
            .padding(12)
            .background(Color(.tertiarySystemGroupedBackground).opacity(0.4))
            .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))

            // Apple Card
            HStack(spacing: 12) {
                ZStack {
                    RoundedRectangle(cornerRadius: 10, style: .continuous)
                        .fill(Color(.tertiarySystemGroupedBackground))
                        .frame(width: 40, height: 40)
                    Image(systemName: "apple.logo")
                        .font(.system(size: 19))
                        .foregroundColor(.primary)
                }

                VStack(alignment: .leading, spacing: 2) {
                    Text("Apple ID")
                        .font(.system(size: 15, weight: .semibold))
                    HStack(spacing: 5) {
                        Circle()
                            .fill(viewModel.appleBound ? AppColors.statusDone : Color.secondary)
                            .frame(width: 6, height: 6)
                        Text(viewModel.appleBound ? L10n.Settings.connected : L10n.Settings.notConnected)
                            .font(.system(size: 12))
                            .foregroundColor(.secondary)
                    }
                }

                Spacer()

                if viewModel.appleBound {
                    Button(L10n.Settings.disconnect) {
                        providerToUnbind = "apple"
                        showUnbindConfirmation = true
                    }
                    .font(.system(size: 12, weight: .semibold))
                    .foregroundColor(AppColors.priorityUrgent)
                    .padding(.horizontal, 10)
                    .padding(.vertical, 5)
                    .background(AppColors.priorityUrgent.opacity(0.12))
                    .clipShape(Capsule())
                } else if viewModel.isAppleBinding {
                    HStack(spacing: 6) {
                        ProgressView()
                            .scaleEffect(0.75)
                        Text(L10n.General.loading)
                            .font(.system(size: 12, weight: .medium))
                            .foregroundColor(.secondary)
                    }
                    .padding(.horizontal, 12)
                    .padding(.vertical, 6)
                    .background(Color(.tertiarySystemGroupedBackground))
                    .clipShape(Capsule())
                } else {
                    SignInWithAppleButton(.continue) { request in
                        request.requestedScopes = [.fullName, .email]
                    } onCompletion: { result in
                        viewModel.handleAppleBind(result: result)
                    }
                    .signInWithAppleButtonStyle(.black)
                    .frame(width: 130, height: 34)
                    .clipShape(Capsule())
                }
            }
            .padding(12)
            .background(Color(.tertiarySystemGroupedBackground).opacity(0.4))
            .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
        }
        .padding(16)
        .background(Color(.secondarySystemGroupedBackground))
        .clipShape(RoundedRectangle(cornerRadius: 16, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: 16, style: .continuous)
                .stroke(Color.primary.opacity(0.06), lineWidth: 1)
        )
    }

    // MARK: - Section 3: Notification & Due Date Warning Preferences
    private var notificationPreferencesSection: some View {
        VStack(alignment: .leading, spacing: 16) {
            VStack(alignment: .leading, spacing: 2) {
                Label(L10n.Settings.notifications, systemImage: "bell.badge.fill")
                    .font(.system(size: 16, weight: .bold))
                    .foregroundColor(.primary)
                Text(L10n.Settings.notificationsDesc)
                    .font(.system(size: 12))
                    .foregroundColor(.secondary)
            }

            // Timezone Picker
            VStack(alignment: .leading, spacing: 6) {
                Text(L10n.Settings.timezone)
                    .font(.system(size: 12, weight: .medium))
                    .foregroundColor(.secondary)

                Picker(L10n.Settings.timezone, selection: $viewModel.timezone) {
                    Text("Asia/Jakarta (WIB, UTC+7)").tag("Asia/Jakarta")
                    Text("Asia/Makassar (WITA, UTC+8)").tag("Asia/Makassar")
                    Text("Asia/Jayapura (WIT, UTC+9)").tag("Asia/Jayapura")
                    Text("Asia/Singapore (SGT, UTC+8)").tag("Asia/Singapore")
                    Text("Asia/Tokyo (JST, UTC+9)").tag("Asia/Tokyo")
                    Text("UTC (UTC+0)").tag("UTC")
                    Text("Europe/London (BST/GMT)").tag("Europe/London")
                    Text("America/New_York (EDT/EST)").tag("America/New_York")
                    Text("America/Los_Angeles (PDT/PST)").tag("America/Los_Angeles")
                }
                .pickerStyle(.menu)
                .padding(6)
                .frame(maxWidth: .infinity, alignment: .leading)
                .background(Color(.tertiarySystemGroupedBackground))
                .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))
            }

            // Due Date Threshold Picker
            VStack(alignment: .leading, spacing: 6) {
                Text(L10n.Settings.threshold)
                    .font(.system(size: 12, weight: .medium))
                    .foregroundColor(.secondary)

                Picker(L10n.Settings.threshold, selection: $viewModel.threshold) {
                    Text(L10n.Settings.threshold15m).tag(15)
                    Text(L10n.Settings.threshold1h).tag(60)
                    Text(L10n.Settings.threshold3h).tag(180)
                    Text(L10n.Settings.threshold12h).tag(720)
                    Text(L10n.Settings.threshold24h).tag(1440)
                    Text(L10n.Settings.threshold48h).tag(2880)
                }
                .pickerStyle(.menu)
                .padding(6)
                .frame(maxWidth: .infinity, alignment: .leading)
                .background(Color(.tertiarySystemGroupedBackground))
                .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))
            }

            // Email Notification Toggle
            Toggle(isOn: $viewModel.emailNotificationsEnabled) {
                VStack(alignment: .leading, spacing: 2) {
                    Text(L10n.Settings.emailNotifications)
                        .font(.system(size: 14, weight: .medium))
                    Text(L10n.Settings.emailNotificationsDesc)
                        .font(.system(size: 11))
                        .foregroundColor(.secondary)
                }
            }
            .tint(AppColors.primary)
            .padding(.vertical, 2)

            // Save Notification Preferences Button
            Button(action: {
                Task { _ = await viewModel.saveNotifications() }
            }) {
                HStack(spacing: 6) {
                    if viewModel.isSavingNotifications {
                        ProgressView()
                            .scaleEffect(0.8)
                            .tint(.white)
                    } else {
                        Image(systemName: "checkmark")
                            .font(.system(size: 13, weight: .bold))
                        Text(L10n.Settings.savePreferences)
                            .font(.system(size: 14, weight: .semibold))
                    }
                }
                .frame(maxWidth: .infinity)
                .padding(.vertical, 11)
                .background(AppColors.brandGradient)
                .foregroundColor(.white)
                .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
            }
            .buttonStyle(.plain)
            .disabled(viewModel.isSavingNotifications)
        }
        .padding(16)
        .background(Color(.secondarySystemGroupedBackground))
        .clipShape(RoundedRectangle(cornerRadius: 16, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: 16, style: .continuous)
                .stroke(Color.primary.opacity(0.06), lineWidth: 1)
        )
    }

    // MARK: - Section 4: Appearance & Language
    private var appearanceAndLanguageSection: some View {
        VStack(alignment: .leading, spacing: 16) {
            Label(L10n.Settings.appearance, systemImage: "paintpalette.fill")
                .font(.system(size: 16, weight: .bold))
                .foregroundColor(.primary)

            // Theme Mode (Dark Mode)
            VStack(alignment: .leading, spacing: 8) {
                HStack {
                    Image(systemName: appAppearance == "dark" ? "moon.stars.fill" : (appAppearance == "light" ? "sun.max.fill" : "circle.lefthalf.filled"))
                        .foregroundColor(AppColors.primary)
                    Text(L10n.Settings.themeMode)
                        .font(.system(size: 13, weight: .semibold))
                    Spacer()
                }

                Picker("Theme", selection: $appAppearance) {
                    Text(L10n.Settings.themeSystem).tag("system")
                    Text(L10n.Settings.themeLight).tag("light")
                    Text(L10n.Settings.themeDark).tag("dark")
                }
                .pickerStyle(.segmented)
            }

            Divider()

            // Language (Default English)
            VStack(alignment: .leading, spacing: 8) {
                HStack {
                    Image(systemName: "globe")
                        .foregroundColor(AppColors.violet)
                    Text(L10n.Settings.language)
                        .font(.system(size: 13, weight: .semibold))
                    Spacer()
                }

                Picker(L10n.Settings.language, selection: Binding(
                    get: { I18n.shared.currentLanguage },
                    set: { I18n.shared.setLanguage($0) }
                )) {
                    ForEach(AppLanguage.allCases) { lang in
                        Text("\(lang.displayName) (\(lang.flag))").tag(lang)
                    }
                }
                .pickerStyle(.segmented)
            }

            Divider()

            // Version info (clean)
            HStack {
                Image(systemName: "info.circle.fill")
                    .font(.system(size: 12))
                    .foregroundColor(.secondary)
                Text(L10n.Settings.appVersion)
                    .font(.system(size: 13))
                    .foregroundColor(.secondary)
                Spacer()
                Text("1.0.0")
                    .font(.system(size: 13, weight: .medium))
                    .foregroundColor(.secondary)
            }
        }
        .padding(16)
        .background(Color(.secondarySystemGroupedBackground))
        .clipShape(RoundedRectangle(cornerRadius: 16, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: 16, style: .continuous)
                .stroke(Color.primary.opacity(0.06), lineWidth: 1)
        )
    }

    // MARK: - Section 5: Widget Configuration (BCA-Style Task Hub Widgets)
    private var widgetConfigurationSection: some View {
        NavigationLink(destination: WidgetSettingsView()) {
            HStack(spacing: 14) {
                ZStack {
                    RoundedRectangle(cornerRadius: 12, style: .continuous)
                        .fill(
                            LinearGradient(
                                colors: [
                                    Color(red: 56/255, green: 189/255, blue: 248/255),
                                    Color(red: 168/255, green: 85/255, blue: 247/255)
                                ],
                                startPoint: .topLeading,
                                endPoint: .bottomTrailing
                            )
                        )
                        .frame(width: 42, height: 42)
                        .shadow(color: Color(red: 56/255, green: 189/255, blue: 248/255).opacity(0.3), radius: 4, x: 0, y: 2)

                    Image(systemName: "slider.horizontal.2.square.stack")
                        .font(.system(size: 18, weight: .semibold))
                        .foregroundColor(.white)
                }

                VStack(alignment: .leading, spacing: 3) {
                    Text(L10n.Settings.widgetSettings)
                        .font(.system(size: 15, weight: .bold))
                        .foregroundColor(.primary)

                    Text(L10n.Settings.widgetSettingsDesc)
                        .font(.system(size: 12))
                        .foregroundColor(.secondary)
                        .lineLimit(2)
                }

                Spacer()

                let totalSelected = SharedWidgetDataManager.shared.getSelectedRecurringTaskIds().count
                    + SharedWidgetDataManager.shared.getSelectedCustomTaskIds().count
                if totalSelected > 0 {
                    Text("\(totalSelected)")
                        .font(.system(size: 12, weight: .bold, design: .rounded))
                        .foregroundColor(.white)
                        .padding(.horizontal, 8)
                        .padding(.vertical, 3)
                        .background(Color(red: 56/255, green: 189/255, blue: 248/255))
                        .clipShape(Capsule())
                }

                Image(systemName: "chevron.right")
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundColor(.secondary)
            }
            .padding(16)
            .background(Color(.secondarySystemGroupedBackground))
            .clipShape(RoundedRectangle(cornerRadius: 16, style: .continuous))
            .overlay(
                RoundedRectangle(cornerRadius: 16, style: .continuous)
                    .stroke(Color.primary.opacity(0.06), lineWidth: 1)
            )
        }
        .buttonStyle(.plain)
    }

    // MARK: - Section 6: AI & MCP Integration
    private var mcpIntegrationSection: some View {
        NavigationLink(destination: McpSettingsView()) {
            HStack(spacing: 14) {
                ZStack {
                    RoundedRectangle(cornerRadius: 12, style: .continuous)
                        .fill(AppColors.brandGradient)
                        .frame(width: 42, height: 42)
                        .shadow(color: AppColors.primary.opacity(0.3), radius: 4, x: 0, y: 2)

                    Image(systemName: "cpu.fill")
                        .font(.system(size: 18, weight: .semibold))
                        .foregroundColor(.white)
                }

                VStack(alignment: .leading, spacing: 3) {
                    Text(L10n.Settings.mcpIntegration)
                        .font(.system(size: 15, weight: .bold))
                        .foregroundColor(.primary)

                    Text(L10n.Settings.mcpIntegrationDesc)
                        .font(.system(size: 12))
                        .foregroundColor(.secondary)
                        .lineLimit(2)
                }

                Spacer()

                Image(systemName: "chevron.right")
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundColor(.secondary)
            }
            .padding(16)
            .background(Color(.secondarySystemGroupedBackground))
            .clipShape(RoundedRectangle(cornerRadius: 16, style: .continuous))
            .overlay(
                RoundedRectangle(cornerRadius: 16, style: .continuous)
                    .stroke(AppColors.primary.opacity(0.15), lineWidth: 1)
            )
        }
        .buttonStyle(.plain)
    }

    // MARK: - Section 6: Danger Zone
    private var dangerZoneSection: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack(spacing: 8) {
                Image(systemName: "exclamationmark.triangle.fill")
                    .foregroundColor(AppColors.priorityUrgent)
                Text(L10n.Settings.dangerZone)
                    .font(.system(size: 16, weight: .bold))
                    .foregroundColor(AppColors.priorityUrgent)
            }

            Text(L10n.Settings.deleteWarning)
                .font(.system(size: 12))
                .foregroundColor(.secondary)
                .lineSpacing(2)

            Button(action: {
                deleteConfirmText = ""
                showDeleteModal = true
            }) {
                HStack(spacing: 6) {
                    Image(systemName: "trash.fill")
                    Text(L10n.Settings.deleteAccount)
                }
                .font(.system(size: 13, weight: .bold))
                .foregroundColor(.white)
                .padding(.horizontal, 14)
                .padding(.vertical, 9)
                .background(AppColors.priorityUrgent)
                .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))
            }
            .buttonStyle(.plain)
            .padding(.top, 4)
        }
        .padding(16)
        .background(AppColors.priorityUrgent.opacity(0.06))
        .clipShape(RoundedRectangle(cornerRadius: 16, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: 16, style: .continuous)
                .stroke(AppColors.priorityUrgent.opacity(0.2), lineWidth: 1)
        )
    }

    // MARK: - Delete Account Modal
    private var deleteAccountModalView: some View {
        NavigationStack {
            VStack(spacing: 16) {
                ZStack {
                    Circle()
                        .fill(AppColors.priorityUrgent.opacity(0.12))
                        .frame(width: 56, height: 56)
                    Image(systemName: "exclamationmark.triangle.fill")
                        .font(.system(size: 24))
                        .foregroundColor(AppColors.priorityUrgent)
                }

                Text(L10n.Settings.deleteAccount)
                    .font(.system(size: 18, weight: .bold))

                Text(L10n.Settings.deletePermanentNotice)
                    .font(.system(size: 13))
                    .foregroundColor(.secondary)
                    .multilineTextAlignment(.center)
                    .padding(.horizontal, 16)

                VStack(alignment: .leading, spacing: 6) {
                    Text(L10n.Settings.deleteConfirmPrompt)
                        .font(.system(size: 12, weight: .semibold))
                        .foregroundColor(.primary)

                    TextField("DELETE", text: $deleteConfirmText)
                        .font(.system(size: 14, design: .monospaced))
                        .autocapitalization(.allCharacters)
                        .padding(10)
                        .background(Color(.tertiarySystemGroupedBackground))
                        .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))
                }
                .padding(.horizontal, 16)
                .padding(.top, 6)

                Spacer()

                HStack(spacing: 12) {
                    Button(L10n.General.cancel) {
                        showDeleteModal = false
                    }
                    .font(.system(size: 14, weight: .medium))
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 12)
                    .background(Color(.tertiarySystemGroupedBackground))
                    .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))

                    Button(action: {
                        Task {
                            let success = await viewModel.deleteAccount(confirmation: deleteConfirmText)
                            if success {
                                showDeleteModal = false
                                onSignOut()
                            }
                        }
                    }) {
                        if viewModel.isDeletingAccount {
                            ProgressView()
                                .tint(.white)
                        } else {
                            Text(L10n.Settings.deleteAccount)
                                .font(.system(size: 14, weight: .bold))
                        }
                    }
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 12)
                    .background(deleteConfirmText.trimmingCharacters(in: .whitespaces).uppercased() == "DELETE" ? AppColors.priorityUrgent : Color.secondary.opacity(0.3))
                    .foregroundColor(.white)
                    .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
                    .disabled(deleteConfirmText.trimmingCharacters(in: .whitespaces).uppercased() != "DELETE" || viewModel.isDeletingAccount)
                }
                .padding(.horizontal, 16)
                .padding(.bottom, 16)
            }
            .padding(.top, 20)
            .navigationTitle(L10n.Settings.deleteConfirmTitle)
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button(L10n.General.close) { showDeleteModal = false }
                }
            }
        }
    }
}
