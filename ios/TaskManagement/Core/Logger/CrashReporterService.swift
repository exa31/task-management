import Foundation
import Network

/// High-reliability crash and telemetry reporter dispatching structured events
/// to Grafana Loki Ingress with offline buffering and automatic network reconnection sync.
public final class CrashReporterService: @unchecked Sendable {
    public static let shared = CrashReporterService()

    private static let storageKeyQueue = "task_management_loki_offline_queue_v1"
    private static let maxQueueSize = 100

    private let session: URLSession
    private let pathMonitor: NWPathMonitor
    private let monitorQueue = DispatchQueue(label: "cloud.eka-dev.task-hub.crashreporter.network")
    private let syncQueue = DispatchQueue(label: "cloud.eka-dev.task-hub.crashreporter.sync")

    private var isFlushing = false
    private var isInitialized = false
    private var cachedUserId: String? = nil

    private init() {
        let configuration = URLSessionConfiguration.default
        configuration.timeoutIntervalForRequest = 10
        configuration.timeoutIntervalForResource = 15
        self.session = URLSession(configuration: configuration)
        self.pathMonitor = NWPathMonitor()
    }

    public var lokiURL: URL {
        let urlStr = AppConfig.lokiURL
        return URL(string: urlStr) ?? URL(string: "https://loki.eka-dev.cloud/loki/api/v1/push")!
    }

    public var environmentName: String {
        #if DEBUG
        return "development"
        #else
        return "production"
        #endif
    }

    public var platformName: String {
        #if os(iOS)
        return "ios"
        #elseif os(macOS)
        return "macos"
        #else
        return "unknown"
        #endif
    }

    public var appName: String {
        return "task-management-ios"
    }

    /// Set or update the active user identifier for telemetry attribution
    public func setUserId(_ userId: String?) {
        syncQueue.async { [weak self] in
            self?.cachedUserId = userId
            if let userId = userId, !userId.isEmpty {
                UserDefaults.standard.set(userId, forKey: "current_user_id")
            } else {
                UserDefaults.standard.removeObject(forKey: "current_user_id")
            }
        }
    }

    /// Initialize CrashReporter, register uncaught exception handler, and start offline synchronization.
    public func initialize() {
        syncQueue.async { [weak self] in
            guard let self = self, !self.isInitialized else { return }
            self.isInitialized = true

            // Restore cached user ID if available
            if self.cachedUserId == nil {
                self.cachedUserId = UserDefaults.standard.string(forKey: "current_user_id")
            }

            // Setup uncaught exception handler for fatal crashes
            NSSetUncaughtExceptionHandler { exception in
                CrashReporterService.shared.recordFatalCrash(
                    name: exception.name.rawValue,
                    reason: exception.reason ?? "Unknown fatal error",
                    callStackSymbols: exception.callStackSymbols
                )
            }

            // Monitor connectivity restoration to flush offline logs
            self.pathMonitor.pathUpdateHandler = { [weak self] path in
                if path.status == .satisfied {
                    self?.flushOfflineLogs()
                }
            }
            self.pathMonitor.start(queue: self.monitorQueue)

            // Initial flush on startup
            self.flushOfflineLogs()
        }
    }

    /// Record fatal crash / uncaught exception
    public func recordFatalCrash(
        name: String,
        reason: String,
        callStackSymbols: [String]
    ) {
        let payload: [String: Any] = [
            "message": "\(name): \(reason)",
            "reason": reason,
            "stackTrace": callStackSymbols.joined(separator: "\n"),
            "fatal": true,
            "tag": "FATAL_CRASH",
            "timestamp": ISO8601DateFormatter().string(from: Date())
        ]

        dispatch(level: "CRITICAL", tag: "FATAL_CRASH", body: payload)
    }

    /// Record runtime Error / Exception
    public func recordError(
        _ error: Error,
        stackTrace: String? = nil,
        reason: String? = nil,
        fatal: Bool = false,
        tag: String = "ERROR",
        customData: [String: Any]? = nil
    ) {
        var payload: [String: Any] = [
            "message": error.localizedDescription,
            "reason": reason ?? "\(error)",
            "stackTrace": stackTrace ?? Thread.callStackSymbols.joined(separator: "\n"),
            "fatal": fatal,
            "tag": tag,
            "timestamp": ISO8601DateFormatter().string(from: Date())
        ]
        if let customData = customData {
            payload["data"] = customData
        }

        dispatch(level: fatal ? "CRITICAL" : "ERROR", tag: tag, body: payload)
    }

    /// Record general informational / warning logs
    public func recordLog(
        message: String,
        level: String = "INFO",
        tag: String = "APP",
        extra: [String: Any]? = nil
    ) {
        var payload: [String: Any] = [
            "message": message,
            "tag": tag,
            "timestamp": ISO8601DateFormatter().string(from: Date())
        ]
        if let extra = extra {
            payload["extra"] = extra
        }

        dispatch(level: level.uppercased(), tag: tag, body: payload)
    }

    /// Record network & API failures
    public func recordHttpError(
        method: String,
        path: String,
        statusCode: Int?,
        errorMessage: String?,
        responseBody: String? = nil,
        durationMs: Int64? = nil
    ) {
        var payload: [String: Any] = [
            "http_method": method,
            "path": path,
            "status_code": statusCode ?? 0,
            "error": errorMessage ?? "",
            "timestamp": ISO8601DateFormatter().string(from: Date())
        ]
        if let durationMs = durationMs {
            payload["duration_ms"] = durationMs
        }
        if let responseBody = responseBody {
            payload["response"] = responseBody
        }

        let level = (statusCode != nil && statusCode! >= 500) ? "ERROR" : "WARN"
        dispatch(level: level, tag: "HTTP_API", body: payload)
    }

    // MARK: - Internal Dispatcher

    private func dispatch(
        level: String,
        tag: String,
        body: [String: Any]
    ) {
        let nowNs = "\(Int64(Date().timeIntervalSince1970 * 1_000_000_000))"

        let bodyData = (try? JSONSerialization.data(withJSONObject: body, options: [])) ?? Data()
        let bodyString = String(data: bodyData, encoding: .utf8) ?? "{}"

        syncQueue.async { [weak self] in
            guard let self = self else { return }
            let userId = (self.cachedUserId?.isEmpty == false) ? self.cachedUserId! : "anonymous"

            let logEntry: [String: String] = [
                "timestamp_ns": nowNs,
                "level": level,
                "tag": tag,
                "user_id": userId,
                "platform": self.platformName,
                "environment": self.environmentName,
                "app": self.appName,
                "body": bodyString
            ]

            self.sendEntries([logEntry]) { success in
                if !success {
                    self.enqueueOffline(logEntry)
                }
            }
        }
    }

    // MARK: - Offline Storage & Sync

    private func enqueueOffline(_ entry: [String: String]) {
        guard let data = try? JSONSerialization.data(withJSONObject: entry, options: []),
              let jsonString = String(data: data, encoding: .utf8) else {
            return
        }

        var queue = UserDefaults.standard.stringArray(forKey: Self.storageKeyQueue) ?? []
        if queue.count >= Self.maxQueueSize {
            queue.removeFirst()
        }
        queue.append(jsonString)
        UserDefaults.standard.set(queue, forKey: Self.storageKeyQueue)
    }

    public func flushOfflineLogs() {
        syncQueue.async { [weak self] in
            guard let self = self, !self.isFlushing else { return }
            self.isFlushing = true

            let queue = UserDefaults.standard.stringArray(forKey: Self.storageKeyQueue) ?? []
            guard !queue.isEmpty else {
                self.isFlushing = false
                return
            }

            var entries: [[String: String]] = []
            for item in queue {
                if let data = item.data(using: .utf8),
                   let dict = try? JSONSerialization.jsonObject(with: data) as? [String: String] {
                    entries.append(dict)
                }
            }

            guard !entries.isEmpty else {
                UserDefaults.standard.removeObject(forKey: Self.storageKeyQueue)
                self.isFlushing = false
                return
            }

            self.sendEntries(entries) { success in
                if success {
                    UserDefaults.standard.removeObject(forKey: Self.storageKeyQueue)
                }
                self.isFlushing = false
            }
        }
    }

    // MARK: - Loki POST Request

    private func sendEntries(
        _ entries: [[String: String]],
        completion: @escaping (Bool) -> Void
    ) {
        guard !entries.isEmpty else {
            completion(true)
            return
        }

        var streamsMap: [String: [[String]]] = [:]
        var labelMap: [String: [String: String]] = [:]

        for item in entries {
            let app = item["app"] ?? appName
            let level = item["level"] ?? "INFO"
            let platform = item["platform"] ?? platformName
            let env = item["environment"] ?? environmentName
            let tag = item["tag"] ?? "APP"
            let userId = item["user_id"] ?? "anonymous"

            let streamKey = "\(app)-\(level)-\(platform)-\(env)-\(tag)"

            if streamsMap[streamKey] == nil {
                streamsMap[streamKey] = []
                labelMap[streamKey] = [
                    "app": app,
                    "level": level,
                    "platform": platform,
                    "environment": env,
                    "tag": tag,
                    "user_id": userId
                ]
            }

            let ts = item["timestamp_ns"] ?? "\(Int64(Date().timeIntervalSince1970 * 1_000_000_000))"
            let body = item["body"] ?? "{}"
            streamsMap[streamKey]?.append([ts, body])
        }

        var streamsList: [[String: Any]] = []
        for (key, values) in streamsMap {
            if let labels = labelMap[key] {
                streamsList.append([
                    "stream": labels,
                    "values": values
                ])
            }
        }

        let payload: [String: Any] = ["streams": streamsList]
        guard let jsonData = try? JSONSerialization.data(withJSONObject: payload, options: []) else {
            completion(false)
            return
        }

        var request = URLRequest(url: lokiURL)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = jsonData

        let task = session.dataTask(with: request) { _, response, error in
            if error != nil {
                completion(false)
                return
            }
            if let httpRes = response as? HTTPURLResponse, (200...299).contains(httpRes.statusCode) {
                completion(true)
            } else {
                completion(false)
            }
        }
        task.resume()
    }
}
