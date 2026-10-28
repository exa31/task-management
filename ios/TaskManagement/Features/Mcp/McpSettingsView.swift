import SwiftUI

#if canImport(UIKit)
import UIKit
#elseif canImport(AppKit)
import AppKit
#endif

// MARK: - Cross-Platform Clipboard Helper
public enum AppClipboard {
    public static func copy(_ text: String) {
        #if canImport(UIKit)
        UIPasteboard.general.string = text
        #elseif canImport(AppKit)
        NSPasteboard.general.clearContents()
        NSPasteboard.general.setString(text, forType: .string)
        #endif
    }
}

public struct McpSettingsView: View {
    @State private var viewModel = McpViewModel()
    @State private var newTokenName: String = ""
    @State private var tokenToRevoke: McpToken? = nil
    @State private var showRevokeConfirmation: Bool = false
    @State private var showGeneratedModal: Bool = false
    @State private var selectedSnippetIndex: Int = 0

    @Environment(\.dismiss) private var dismiss
    @Environment(\.colorScheme) private var colorScheme

    public init() {}

    private var mcpEndpointURL: String {
        "\(AppConfig.baseURL)/mcp"
    }

    public var body: some View {
        ScrollView {
            VStack(spacing: 20) {
                // Toast notification banner
                if let toast = viewModel.toastMessage {
                    toastBanner(message: toast, isError: false)
                }
                if let error = viewModel.errorMessage, !error.isEmpty {
                    toastBanner(message: error, isError: true)
                }

                // Header Banner
                headerBanner

                // Server Endpoint Card
                endpointCard

                // Generate New Token Card
                generateTokenCard

                // Active Tokens List
                activeTokensCard

                // Quick Setup Snippets Card
                configSnippetsCard
            }
            .padding(.horizontal, 16)
            .padding(.vertical, 16)
        }
        .background(backgroundSurface)
        .navigationTitle(L10n.Mcp.title)
        #if canImport(UIKit)
        .navigationBarTitleDisplayMode(.inline)
        #endif
        .sheet(item: $viewModel.newlyGeneratedToken) { tokenResponse in
            newlyGeneratedTokenSheet(tokenResponse)
        }
        .confirmationDialog(
            L10n.Mcp.revokeConfirmTitle,
            isPresented: $showRevokeConfirmation,
            presenting: tokenToRevoke
        ) { token in
            Button(L10n.Mcp.revoke, role: .destructive) {
                Task {
                    let success = await viewModel.revokeToken(id: token.id)
                    if success {
                        viewModel.toastMessage = "Token '\(token.name)' revoked."
                    }
                }
            }
            Button(L10n.General.cancel, role: .cancel) {}
        } message: { token in
            Text(L10n.Mcp.revokeConfirmMessage(name: token.name))
        }
        .refreshable {
            await viewModel.loadTokens()
        }
    }

    // MARK: - Surfaces
    private var backgroundSurface: Color {
        #if canImport(UIKit)
        return Color(uiColor: .systemGroupedBackground)
        #else
        return Color(nsColor: .windowBackgroundColor)
        #endif
    }

    private var cardSurface: Color {
        #if canImport(UIKit)
        return Color(uiColor: .secondarySystemGroupedBackground)
        #else
        return Color(nsColor: .controlBackgroundColor)
        #endif
    }

    // MARK: - Toast Banner
    private func toastBanner(message: String, isError: Bool) -> some View {
        HStack(spacing: 10) {
            Image(systemName: isError ? "exclamationmark.triangle.fill" : "checkmark.circle.fill")
                .foregroundColor(isError ? AppColors.priorityUrgent : AppColors.statusDone)
            Text(message)
                .font(.system(size: 13, weight: .medium))
                .foregroundColor(isError ? AppColors.priorityUrgent : AppColors.statusDone)
            Spacer()
            Button(action: {
                if isError {
                    viewModel.errorMessage = nil
                } else {
                    viewModel.toastMessage = nil
                }
            }) {
                Image(systemName: "xmark")
                    .font(.system(size: 11, weight: .bold))
                    .foregroundColor(isError ? AppColors.priorityUrgent : AppColors.statusDone)
            }
        }
        .padding(12)
        .background((isError ? AppColors.priorityUrgent : AppColors.statusDone).opacity(0.12))
        .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
    }

    // MARK: - Header Banner
    private var headerBanner: some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack(spacing: 12) {
                ZStack {
                    RoundedRectangle(cornerRadius: 14, style: .continuous)
                        .fill(AppColors.brandGradient)
                        .frame(width: 44, height: 44)
                        .shadow(color: AppColors.primary.opacity(0.3), radius: 6, x: 0, y: 3)

                    Image(systemName: "cpu.fill")
                        .font(.system(size: 20, weight: .semibold))
                        .foregroundColor(.white)
                }

                VStack(alignment: .leading, spacing: 3) {
                    Text(L10n.Mcp.title)
                        .font(.system(size: 17, weight: .bold))
                        .foregroundColor(.primary)

                    HStack(spacing: 6) {
                        Circle()
                            .fill(AppColors.statusDone)
                            .frame(width: 7, height: 7)
                        Text("JSON-RPC 2.0 • Active")
                            .font(.system(size: 11, weight: .semibold))
                            .foregroundColor(AppColors.statusDone)
                    }
                }
                Spacer()
            }

            Text(L10n.Mcp.subtitle)
                .font(.system(size: 13))
                .foregroundColor(.secondary)
                .fixedSize(horizontal: false, vertical: true)
        }
        .padding(16)
        .background(cardSurface)
        .clipShape(RoundedRectangle(cornerRadius: 16, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: 16, style: .continuous)
                .stroke(AppColors.primary.opacity(0.15), lineWidth: 1)
        )
    }

    // MARK: - MCP Server Endpoint
    private var endpointCard: some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack {
                Label(L10n.Mcp.endpoint, systemImage: "link")
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundColor(.primary)
                Spacer()
            }

            HStack(spacing: 10) {
                Text(mcpEndpointURL)
                    .font(.system(size: 12, weight: .medium, design: .monospaced))
                    .foregroundColor(.primary)
                    .lineLimit(1)
                    .truncationMode(.middle)
                    .padding(.horizontal, 10)
                    .padding(.vertical, 8)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .background(Color.primary.opacity(0.04))
                    .clipShape(RoundedRectangle(cornerRadius: 8, style: .continuous))

                Button(action: {
                    AppClipboard.copy(mcpEndpointURL)
                    viewModel.toastMessage = L10n.Mcp.endpointCopied
                }) {
                    HStack(spacing: 5) {
                        Image(systemName: "doc.on.doc.fill")
                        Text(L10n.Mcp.copyToken)
                    }
                    .font(.system(size: 12, weight: .semibold))
                    .foregroundColor(.white)
                    .padding(.horizontal, 12)
                    .padding(.vertical, 8)
                    .background(AppColors.primary)
                    .clipShape(RoundedRectangle(cornerRadius: 8, style: .continuous))
                }
                .buttonStyle(.plain)
            }
        }
        .padding(16)
        .background(cardSurface)
        .clipShape(RoundedRectangle(cornerRadius: 16, style: .continuous))
    }

    // MARK: - Generate Token Card
    private var generateTokenCard: some View {
        VStack(alignment: .leading, spacing: 12) {
            Label(L10n.Mcp.generateToken, systemImage: "key.fill")
                .font(.system(size: 14, weight: .semibold))
                .foregroundColor(.primary)

            HStack(spacing: 10) {
                TextField(L10n.Mcp.tokenNamePlaceholder, text: $newTokenName)
                    .font(.system(size: 14))
                    .padding(.horizontal, 12)
                    .padding(.vertical, 9)
                    .background(Color.primary.opacity(0.04))
                    .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))

                Button(action: {
                    Task {
                        let name = newTokenName
                        let success = await viewModel.generateToken(name: name)
                        if success {
                            newTokenName = ""
                        }
                    }
                }) {
                    HStack(spacing: 6) {
                        if viewModel.isGenerating {
                            ProgressView()
                                .progressViewStyle(CircularProgressViewStyle(tint: .white))
                                .scaleEffect(0.8)
                        } else {
                            Image(systemName: "plus")
                            Text(L10n.General.create)
                        }
                    }
                    .font(.system(size: 13, weight: .semibold))
                    .foregroundColor(.white)
                    .padding(.horizontal, 14)
                    .padding(.vertical, 9)
                    .background(
                        newTokenName.trimmingCharacters(in: .whitespaces).isEmpty || viewModel.isGenerating
                            ? AppColors.primary.opacity(0.5)
                            : AppColors.primary
                    )
                    .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))
                }
                .buttonStyle(.plain)
                .disabled(newTokenName.trimmingCharacters(in: .whitespaces).isEmpty || viewModel.isGenerating)
            }
        }
        .padding(16)
        .background(cardSurface)
        .clipShape(RoundedRectangle(cornerRadius: 16, style: .continuous))
    }

    // MARK: - Active Tokens Card
    private var activeTokensCard: some View {
        VStack(alignment: .leading, spacing: 14) {
            HStack {
                Label(L10n.Mcp.tokensTitle, systemImage: "lock.shield.fill")
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundColor(.primary)

                Spacer()

                if viewModel.isLoading {
                    ProgressView()
                        .scaleEffect(0.8)
                } else {
                    Text("\(viewModel.tokens.count)")
                        .font(.system(size: 12, weight: .bold))
                        .foregroundColor(AppColors.primary)
                        .padding(.horizontal, 8)
                        .padding(.vertical, 2)
                        .background(AppColors.primary.opacity(0.12))
                        .clipShape(Capsule())
                }
            }

            Text(L10n.Mcp.tokensSubtitle)
                .font(.system(size: 12))
                .foregroundColor(.secondary)

            if viewModel.tokens.isEmpty && !viewModel.isLoading {
                HStack {
                    Spacer()
                    VStack(spacing: 8) {
                        Image(systemName: "key.slash.fill")
                            .font(.system(size: 28))
                            .foregroundColor(.secondary.opacity(0.6))
                        Text(L10n.Mcp.noTokens)
                            .font(.system(size: 13))
                            .foregroundColor(.secondary)
                    }
                    .padding(.vertical, 24)
                    Spacer()
                }
            } else {
                VStack(spacing: 10) {
                    ForEach(viewModel.tokens) { token in
                        tokenRow(token)
                    }
                }
            }
        }
        .padding(16)
        .background(cardSurface)
        .clipShape(RoundedRectangle(cornerRadius: 16, style: .continuous))
    }

    // MARK: - Token Row Item
    private func tokenRow(_ token: McpToken) -> some View {
        HStack(spacing: 12) {
            ZStack {
                Circle()
                    .fill(AppColors.indigo.opacity(0.15))
                    .frame(width: 36, height: 36)

                Image(systemName: "key.fill")
                    .font(.system(size: 14))
                    .foregroundColor(AppColors.indigo)
            }

            VStack(alignment: .leading, spacing: 4) {
                Text(token.name)
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundColor(.primary)

                HStack(spacing: 8) {
                    Text("\(token.tokenPrefix)...")
                        .font(.system(size: 11, weight: .medium, design: .monospaced))
                        .foregroundColor(.secondary)
                        .padding(.horizontal, 6)
                        .padding(.vertical, 2)
                        .background(Color.primary.opacity(0.04))
                        .clipShape(RoundedRectangle(cornerRadius: 4, style: .continuous))

                    if let lastUsed = token.lastUsedAt, !lastUsed.isEmpty {
                        Text(L10n.Mcp.lastUsed(date: formatShortDate(lastUsed)))
                            .font(.system(size: 11))
                            .foregroundColor(.secondary)
                    } else {
                        Text(L10n.Mcp.neverUsed)
                            .font(.system(size: 11))
                            .foregroundColor(.secondary)
                    }
                }
            }

            Spacer()

            // Revoke Button
            Button(action: {
                tokenToRevoke = token
                showRevokeConfirmation = true
            }) {
                if viewModel.revokingTokenId == token.id {
                    ProgressView()
                        .scaleEffect(0.7)
                } else {
                    HStack(spacing: 4) {
                        Image(systemName: "trash")
                        Text(L10n.Mcp.revoke)
                    }
                    .font(.system(size: 12, weight: .semibold))
                    .foregroundColor(AppColors.priorityUrgent)
                    .padding(.horizontal, 10)
                    .padding(.vertical, 6)
                    .background(AppColors.priorityUrgent.opacity(0.1))
                    .clipShape(RoundedRectangle(cornerRadius: 8, style: .continuous))
                }
            }
            .buttonStyle(.plain)
            .disabled(viewModel.revokingTokenId == token.id)
        }
        .padding(10)
        .background(Color.primary.opacity(0.02))
        .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
    }

    // MARK: - Configuration Snippets Card
    private var configSnippetsCard: some View {
        VStack(alignment: .leading, spacing: 12) {
            Label(L10n.Mcp.configSnippets, systemImage: "applescript.fill")
                .font(.system(size: 14, weight: .semibold))
                .foregroundColor(.primary)

            // Platform Picker
            Picker("", selection: $selectedSnippetIndex) {
                Text("Antigravity / VS Code").tag(0)
                Text("Cursor").tag(1)
                Text("Claude Desktop").tag(2)
            }
            .pickerStyle(.segmented)

            // Code block view
            let snippet = currentConfigSnippet
            VStack(alignment: .trailing, spacing: 8) {
                ScrollView(.horizontal, showsIndicators: false) {
                    Text(snippet)
                        .font(.system(size: 11, weight: .regular, design: .monospaced))
                        .foregroundColor(.primary)
                        .padding(12)
                }
                .frame(maxWidth: .infinity, alignment: .leading)
                .background(Color.primary.opacity(0.05))
                .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))

                Button(action: {
                    AppClipboard.copy(snippet)
                    viewModel.toastMessage = L10n.Mcp.configCopied
                }) {
                    HStack(spacing: 5) {
                        Image(systemName: "doc.on.doc")
                        Text(L10n.Mcp.copyConfig)
                    }
                    .font(.system(size: 12, weight: .semibold))
                    .foregroundColor(AppColors.primary)
                    .padding(.horizontal, 10)
                    .padding(.vertical, 6)
                    .background(AppColors.primary.opacity(0.12))
                    .clipShape(RoundedRectangle(cornerRadius: 8, style: .continuous))
                }
                .buttonStyle(.plain)
            }
        }
        .padding(16)
        .background(cardSurface)
        .clipShape(RoundedRectangle(cornerRadius: 16, style: .continuous))
    }

    private var currentConfigSnippet: String {
        switch selectedSnippetIndex {
        case 0:
            // Antigravity / VS Code
            return """
            {
              "mcpServers": {
                "task-hub": {
                  "url": "\(mcpEndpointURL)",
                  "headers": {
                    "Authorization": "Bearer YOUR_MCP_PAT"
                  }
                }
              }
            }
            """
        case 1:
            // Cursor
            return """
            {
              "mcpServers": {
                "task-hub": {
                  "url": "\(mcpEndpointURL)",
                  "headers": {
                    "Authorization": "Bearer YOUR_MCP_PAT"
                  }
                }
              }
            }
            """
        default:
            // Claude Desktop
            return """
            {
              "mcpServers": {
                "task-hub": {
                  "command": "npx",
                  "args": ["-y", "@modelcontextprotocol/server-everything"],
                  "env": {
                    "TASK_HUB_URL": "\(mcpEndpointURL)"
                  }
                }
              }
            }
            """
        }
    }

    // MARK: - Newly Generated Token Sheet
    private func newlyGeneratedTokenSheet(_ res: GenerateMcpTokenResponse) -> some View {
        NavigationStack {
            VStack(spacing: 20) {
                ZStack {
                    Circle()
                        .fill(AppColors.statusDone.opacity(0.15))
                        .frame(width: 64, height: 64)

                    Image(systemName: "key.fill")
                        .font(.system(size: 28))
                        .foregroundColor(AppColors.statusDone)
                }

                Text(L10n.Mcp.tokenCreatedTitle)
                    .font(.system(size: 20, weight: .bold))
                    .foregroundColor(.primary)

                // Warning Notice
                HStack(alignment: .top, spacing: 10) {
                    Image(systemName: "exclamationmark.triangle.fill")
                        .foregroundColor(AppColors.priorityHigh)
                        .font(.system(size: 16))

                    Text(L10n.Mcp.tokenCreatedWarning)
                        .font(.system(size: 13, weight: .medium))
                        .foregroundColor(AppColors.priorityHigh)
                        .fixedSize(horizontal: false, vertical: true)
                }
                .padding(12)
                .background(AppColors.priorityHigh.opacity(0.12))
                .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))

                // Raw Token display box
                VStack(alignment: .leading, spacing: 8) {
                    Text(res.name)
                        .font(.system(size: 13, weight: .bold))
                        .foregroundColor(.secondary)

                    HStack {
                        Text(res.rawToken)
                            .font(.system(size: 13, weight: .medium, design: .monospaced))
                            .foregroundColor(.primary)
                            .lineLimit(2)
                            .textSelection(.enabled)

                        Spacer()
                    }
                    .padding(12)
                    .background(Color.primary.opacity(0.05))
                    .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))
                }

                // Copy Action Button
                Button(action: {
                    AppClipboard.copy(res.rawToken)
                    viewModel.toastMessage = L10n.Mcp.tokenCopied
                }) {
                    HStack(spacing: 8) {
                        Image(systemName: "doc.on.doc.fill")
                        Text(L10n.Mcp.copyToken)
                    }
                    .font(.system(size: 15, weight: .bold))
                    .foregroundColor(.white)
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 14)
                    .background(AppColors.brandGradient)
                    .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
                    .shadow(color: AppColors.primary.opacity(0.4), radius: 6, x: 0, y: 3)
                }
                .buttonStyle(.plain)

                Spacer()
            }
            .padding(20)
            .navigationTitle(L10n.Mcp.tokensTitle)
            #if canImport(UIKit)
            .navigationBarTitleDisplayMode(.inline)
            #endif
            .toolbar {
                ToolbarItem(placement: .confirmationAction) {
                    Button(L10n.General.done) {
                        viewModel.newlyGeneratedToken = nil
                    }
                    .font(.system(size: 15, weight: .bold))
                }
            }
        }
        .presentationDetents([.medium, .large])
    }

    private func formatShortDate(_ iso: String) -> String {
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        guard let date = formatter.date(from: iso) ?? ISO8601DateFormatter().date(from: iso) else {
            return String(iso.prefix(10))
        }
        let display = DateFormatter()
        display.dateStyle = .medium
        display.timeStyle = .short
        return display.string(from: date)
    }
}
