import Foundation

public struct NetworkLogger: Sendable {
    public static func logRequest(_ request: URLRequest) {
        #if DEBUG
        let url = request.url?.absoluteString ?? "Unknown URL"
        let method = request.httpMethod ?? "GET"
        var log = "\n======================================================\n"
        log += "🚀 [HTTP REQUEST] \(method) \(url)\n"
        
        if let headers = request.allHTTPHeaderFields, !headers.isEmpty {
            log += "📋 Headers:\n"
            for (key, val) in headers {
                let displayVal = key.lowercased() == "authorization" ? "\(val.prefix(18))...[TRUNCATED]" : val
                log += "   • \(key): \(displayVal)\n"
            }
        }

        if let body = request.httpBody, let bodyString = String(data: body, encoding: .utf8) {
            log += "📦 Body:\n\(prettyPrintJSON(bodyString))\n"
        }
        log += "======================================================\n"
        print(log)
        #endif
    }

    public static func logResponse(url: String, method: String, statusCode: Int, data: Data?, duration: Double) {
        #if DEBUG
        let isSuccess = (200...299).contains(statusCode)
        let icon = isSuccess ? "✅" : "⚠️"
        var log = "\n------------------------------------------------------\n"
        log += "\(icon) [HTTP RESPONSE \(statusCode)] \(method) \(url) (\(String(format: "%.0f", duration * 1000))ms)\n"

        if let data = data, let bodyString = String(data: data, encoding: .utf8) {
            log += "📄 Response Body:\n\(prettyPrintJSON(bodyString))\n"
        }
        log += "------------------------------------------------------\n"
        print(log)
        #endif
    }

    public static func logError(url: String, method: String, error: Error) {
        #if DEBUG
        var log = "\n❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌\n"
        log += "❌ [HTTP NETWORK FAILURE] \(method) \(url)\n"
        log += "🚨 Error: \(error.localizedDescription)\n"
        log += "❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌\n"
        print(log)
        #endif
    }

    public static func logSSE(event: String, data: String) {
        #if DEBUG
        print("⚡ [SSE EVENT RECEIVED] Type: \(event) | Data: \(data)")
        #endif
    }

    private static func prettyPrintJSON(_ jsonString: String) -> String {
        guard let data = jsonString.data(using: .utf8),
              let json = try? JSONSerialization.jsonObject(with: data, options: []),
              let prettyData = try? JSONSerialization.data(withJSONObject: json, options: [.prettyPrinted]),
              let prettyString = String(data: prettyData, encoding: .utf8) else {
            return jsonString
        }
        return prettyString
    }
}
