import Foundation
import os.log

/// Log severity levels for structured diagnostic output.
public enum LogLevel: Int, Comparable, Sendable {
    case debug = 0
    case info = 1
    case success = 2
    case warning = 3
    case error = 4

    public static func < (lhs: LogLevel, rhs: LogLevel) -> Bool {
        lhs.rawValue < rhs.rawValue
    }

    public var title: String {
        switch self {
        case .debug:   return "DEBUG"
        case .info:    return "INFO"
        case .success: return "SUCCESS"
        case .warning: return "WARN"
        case .error:   return "ERROR"
        }
    }

    public var icon: String {
        switch self {
        case .debug:   return "🔍"
        case .info:    return "ℹ️"
        case .success: return "✅"
        case .warning: return "⚠️"
        case .error:   return "❌"
        }
    }

    public var badge: String {
        "\(icon) [\(title)]"
    }

    public var osLogType: OSLogType {
        switch self {
        case .debug:   return .debug
        case .info:    return .info
        case .success: return .default
        case .warning: return .default
        case .error:   return .error
        }
    }
}

/// Standardized log categories to group related subsystem activities.
public struct LogCategory: Sendable, RawRepresentable, ExpressibleByStringLiteral, Hashable {
    public let rawValue: String

    public init(rawValue: String) {
        self.rawValue = rawValue
    }

    public init(stringLiteral value: String) {
        self.rawValue = value
    }

    public static let app: LogCategory = "APP"
    public static let auth: LogCategory = "AUTH"
    public static let network: LogCategory = "NETWORK"
    public static let push: LogCategory = "PUSH"
    public static let sse: LogCategory = "SSE"
    public static let task: LogCategory = "TASK"
    public static let reminder: LogCategory = "REMINDER"
    public static let storage: LogCategory = "STORAGE"
}

/// Unified, high-fidelity logging system for TaskManagement iOS app.
/// Combines native Apple unified `os.Logger` for Console.app integration
/// with aesthetic formatted console output for developer debugging.
public final class AppLogger: Sendable {
    public static let shared = AppLogger()

    #if DEBUG
    public static var minimumLevel: LogLevel = .debug
    #else
    public static var minimumLevel: LogLevel = .info
    #endif

    private static let dateFormatter: DateFormatter = {
        let formatter = DateFormatter()
        formatter.dateFormat = "HH:mm:ss.SSS"
        return formatter
    }()

    private static let subsystem = Bundle.main.bundleIdentifier ?? "cloud.eka-dev.task-hub"

    private init() {}

    // MARK: - Core Logging Methods

    public static func debug(
        _ message: String,
        category: LogCategory = .app,
        metadata: [String: Any]? = nil,
        file: String = #file,
        line: Int = #line
    ) {
        log(level: .debug, message: message, category: category, error: nil, metadata: metadata, file: file, line: line)
    }

    public static func info(
        _ message: String,
        category: LogCategory = .app,
        metadata: [String: Any]? = nil,
        file: String = #file,
        line: Int = #line
    ) {
        log(level: .info, message: message, category: category, error: nil, metadata: metadata, file: file, line: line)
    }

    public static func success(
        _ message: String,
        category: LogCategory = .app,
        metadata: [String: Any]? = nil,
        file: String = #file,
        line: Int = #line
    ) {
        log(level: .success, message: message, category: category, error: nil, metadata: metadata, file: file, line: line)
    }

    public static func warn(
        _ message: String,
        category: LogCategory = .app,
        metadata: [String: Any]? = nil,
        file: String = #file,
        line: Int = #line
    ) {
        log(level: .warning, message: message, category: category, error: nil, metadata: metadata, file: file, line: line)
    }

    public static func error(
        _ message: String,
        error: Error? = nil,
        category: LogCategory = .app,
        metadata: [String: Any]? = nil,
        file: String = #file,
        line: Int = #line
    ) {
        log(level: .error, message: message, category: category, error: error, metadata: metadata, file: file, line: line)
    }

    // MARK: - Internal Formatter & Dispatcher

    private static func log(
        level: LogLevel,
        message: String,
        category: LogCategory,
        error: Error?,
        metadata: [String: Any]?,
        file: String,
        line: Int
    ) {
        guard level >= minimumLevel else { return }

        let fileName = (file as NSString).lastPathComponent
        let timestamp = dateFormatter.string(from: Date())

        var fullMessage = message
        if let error = error {
            fullMessage += " | Error: \(error.localizedDescription)"
        }

        // Format metadata if present
        var metaString = ""
        if let metadata = metadata, !metadata.isEmpty {
            let pairs = metadata.map { "\($0.key)=\($0.value)" }.joined(separator: ", ")
            metaString = " {\(pairs)}"
        }

        // Format: [12:34:56.789] ℹ️ [INFO] [AUTH] (AuthViewModel.swift:45) > User logged in {user_id=1}
        let formattedString = "[\(timestamp)] \(level.badge) [\(category.rawValue)] (\(fileName):\(line)) › \(fullMessage)\(metaString)"

        #if DEBUG
        print(formattedString)
        #endif

        // Forward to native os.Logger for Console.app & system log recording
        let osLog = OSLog(subsystem: subsystem, category: category.rawValue)
        os_log("%{public}@", log: osLog, type: level.osLogType, formattedString)

        // Forward warning and error events to Grafana Loki via CrashReporterService
        if level >= .warning {
            if let error = error {
                CrashReporterService.shared.recordError(
                    error,
                    reason: message,
                    fatal: level == .error,
                    tag: category.rawValue,
                    customData: metadata
                )
            } else {
                CrashReporterService.shared.recordLog(
                    message: "\(message)\(metaString)",
                    level: level.title,
                    tag: category.rawValue,
                    extra: metadata
                )
            }
        }
    }
}
