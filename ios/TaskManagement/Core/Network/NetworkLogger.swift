import Foundation

public struct NetworkLogger: Sendable {
    public static func logRequest(_ request: URLRequest) {
        let url = request.url?.absoluteString ?? "Unknown URL"
        let method = request.httpMethod ?? "GET"

        var meta: [String: Any] = ["method": method, "url": url]
        if let headers = request.allHTTPHeaderFields, !headers.isEmpty {
            meta["headers"] = headers.keys.joined(separator: ", ")
        }
        if let body = request.httpBody, let bodyString = String(data: body, encoding: .utf8), !bodyString.isEmpty {
            meta["body"] = bodyString.count > 120 ? "\(bodyString.prefix(120))..." : bodyString
        }

        AppLogger.debug("→ \(method) \(url)", category: .network, metadata: meta)
    }

    public static func logResponse(url: String, method: String, statusCode: Int, data: Data?, duration: Double) {
        let durationMs = String(format: "%.0f", duration * 1000)
        let isSuccess = (200...299).contains(statusCode)

        var meta: [String: Any] = [
            "status": statusCode,
            "duration": "\(durationMs)ms"
        ]

        if let data = data, let bodyString = String(data: data, encoding: .utf8), !bodyString.isEmpty {
            meta["preview"] = bodyString.count > 100 ? "\(bodyString.prefix(100))..." : bodyString
        }

        if isSuccess {
            AppLogger.success("← [\(statusCode)] \(method) \(url) (\(durationMs)ms)", category: .network, metadata: meta)
        } else {
            AppLogger.warn("← [\(statusCode)] \(method) \(url) (\(durationMs)ms)", category: .network, metadata: meta)
        }
    }

    public static func logError(url: String, method: String, error: Error) {
        AppLogger.error("Network request failed: \(method) \(url)", error: error, category: .network)
    }

    public static func logSSE(event: String, data: String) {
        let preview = data.count > 80 ? "\(data.prefix(80))..." : data
        AppLogger.debug("SSE Event: [\(event)]", category: .sse, metadata: ["event": event, "preview": preview])
    }
}
