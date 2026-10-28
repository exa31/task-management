import SwiftUI

@MainActor
@Observable
public final class McpViewModel {
    public var tokens: [McpToken] = []
    public var isLoading: Bool = false
    public var isGenerating: Bool = false
    public var revokingTokenId: Int64? = nil
    public var errorMessage: String? = nil
    public var toastMessage: String? = nil
    public var newlyGeneratedToken: GenerateMcpTokenResponse? = nil

    private let apiClient = APIClient.shared

    public init() {
        Task {
            await loadTokens()
        }
    }

    public func loadTokens() async {
        isLoading = true
        errorMessage = nil
        do {
            tokens = try await apiClient.getMcpTokens()
            AppLogger.info("Loaded \(tokens.count) MCP tokens successfully", category: .app)
        } catch {
            if error.isTaskOrURLCancelled {
                AppLogger.debug("MCP tokens load cancelled", category: .app)
            } else {
                AppLogger.error("Failed to load MCP tokens", error: error, category: .app)
                errorMessage = L10n.Errors.message(for: error)
            }
        }
        isLoading = false
    }

    public func generateToken(name: String) async -> Bool {
        let trimmed = name.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return false }

        isGenerating = true
        errorMessage = nil
        do {
            let res = try await apiClient.generateMcpToken(name: trimmed)
            self.newlyGeneratedToken = res
            AppLogger.success("Generated new MCP token: \(trimmed)", category: .app, metadata: ["id": res.id])

            // Refresh tokens list to reflect the new active token
            await loadTokens()
            isGenerating = false
            return true
        } catch {
            if error.isTaskOrURLCancelled {
                AppLogger.debug("MCP token generation cancelled", category: .app)
            } else {
                AppLogger.error("Failed to generate MCP token", error: error, category: .app)
                errorMessage = L10n.Errors.message(for: error)
            }
            isGenerating = false
            return false
        }
    }

    public func revokeToken(id: Int64) async -> Bool {
        revokingTokenId = id
        errorMessage = nil
        do {
            try await apiClient.deleteMcpToken(id: id)
            tokens.removeAll { $0.id == id }
            AppLogger.success("Revoked MCP token", category: .app, metadata: ["id": id])
            revokingTokenId = nil
            return true
        } catch {
            if error.isTaskOrURLCancelled {
                AppLogger.debug("MCP token revocation cancelled", category: .app)
            } else {
                AppLogger.error("Failed to revoke MCP token", error: error, category: .app, metadata: ["id": id])
                errorMessage = L10n.Errors.message(for: error)
            }
            revokingTokenId = nil
            return false
        }
    }
}
