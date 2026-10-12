import Foundation

public struct RealtimeEvent: Sendable {
    public let type: String
    public let data: String
}

public final class SSEClient: Sendable {
    public static let shared = SSEClient()

    private let baseURL = AppConfig.sseURL

    private init() {}

    public func events() -> AsyncStream<RealtimeEvent> {
        AsyncStream { continuation in
            let task = Task {
                guard let token = KeychainManager.shared.get(key: "access_token"),
                      let url = URL(string: "\(baseURL)?token=\(token)") else {
                    continuation.finish()
                    return
                }

                var request = URLRequest(url: url)
                request.setValue("text/event-stream", forHTTPHeaderField: "Accept")
                request.timeoutInterval = .infinity

                do {
                    #if DEBUG
                    print("⚡ [SSE CONNECTING] Connecting to \(url.absoluteString)...")
                    #endif
                    let (asyncBytes, _) = try await URLSession.shared.bytes(for: request)
                    #if DEBUG
                    print("⚡ [SSE CONNECTED] Stream established successfully.")
                    #endif
                    var currentEvent = "message"

                    for try await line in asyncBytes.lines {
                        if line.hasPrefix("event:") {
                            currentEvent = line.replacingOccurrences(of: "event:", with: "").trimmingCharacters(in: .whitespaces)
                        } else if line.hasPrefix("data:") {
                            let dataContent = line.replacingOccurrences(of: "data:", with: "").trimmingCharacters(in: .whitespaces)
                            NetworkLogger.logSSE(event: currentEvent, data: dataContent)
                            continuation.yield(RealtimeEvent(type: currentEvent, data: dataContent))
                        }
                    }
                } catch {
                    #if DEBUG
                    print("⚡ [SSE DISCONNECTED / ERROR] \(error.localizedDescription)")
                    #endif
                }
                continuation.finish()
            }

            continuation.onTermination = { _ in
                task.cancel()
            }
        }
    }
}
