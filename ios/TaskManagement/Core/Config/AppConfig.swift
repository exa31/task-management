import Foundation

public enum AppEnvironment: String, CaseIterable, Sendable {
    case development = "Development"
    case staging = "Staging"
    case production = "Production"

    public static var current: AppEnvironment {
        #if DEBUG
        return .development
        #else
        return .production
        #endif
    }

    public static let bundleIdentifier = "cloud.eka-dev.task-hub"

    public var baseURL: String {
        switch self {
        case .development:
            return "https://be-task-hub.eka-dev.cloud/api/v1"
        case .staging:
            return "https://be-task-hub.eka-dev.cloud/api/v1"
        case .production:
            return "https://be-task-hub.eka-dev.cloud/api/v1"
        }
    }

    public var sseURL: String {
        return "\(baseURL)/realtime/sse/events"
    }

    public var googleClientID: String {
        switch self {
        case .development:
            return "YOUR_DEV_IOS_GOOGLE_CLIENT_ID.apps.googleusercontent.com"
        case .staging, .production:
            return "YOUR_PROD_IOS_GOOGLE_CLIENT_ID.apps.googleusercontent.com"
        }
    }
}

public struct AppConfig: Sendable {
    public static var environment: AppEnvironment {
        return AppEnvironment.current
    }

    public static var baseURL: String {
        return environment.baseURL
    }

    public static var sseURL: String {
        return environment.sseURL
    }

    public static var googleClientID: String {
        return environment.googleClientID
    }
}
