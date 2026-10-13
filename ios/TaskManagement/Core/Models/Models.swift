import Foundation

// MARK: - Auth & User
public struct User: Codable, Identifiable {
    public let id: String
    public let name: String
    public let email: String
    public let avatarUrl: String?
    public let provider: String?
    public let googleBound: Bool
    public let appleBound: Bool
    public let timezone: String?
    public let dueDateWarningThreshold: Int
    public let emailNotificationsEnabled: Bool

    enum CodingKeys: String, CodingKey {
        case id, name, email, provider, timezone
        case avatarUrl = "avatar_url"
        case googleBound = "google_bound"
        case appleBound = "apple_bound"
        case dueDateWarningThreshold = "due_date_warning_threshold"
        case emailNotificationsEnabled = "email_notifications_enabled"
    }
}

public struct TokenPair: Codable {
    public let accessToken: String
    public let refreshToken: String
    public let expiresAt: String?
    public let tokenType: String

    enum CodingKeys: String, CodingKey {
        case accessToken = "access_token"
        case refreshToken = "refresh_token"
        case expiresAt = "expires_at"
        case tokenType = "token_type"
    }
}

public struct AuthResponse: Codable {
    public let user: User
    public let tokens: TokenPair?
}

public struct OAuthLoginRequest: Codable {
    public let provider: String
    public let idToken: String?
    public let code: String?
    public let email: String?
    public let name: String?
    public let avatarUrl: String?

    enum CodingKeys: String, CodingKey {
        case provider, code, email, name
        case idToken = "id_token"
        case avatarUrl = "avatar_url"
    }
}

public struct RefreshTokenRequest: Codable {
    public let refreshToken: String

    enum CodingKeys: String, CodingKey {
        case refreshToken = "refresh_token"
    }
}

public struct UpdateProfileRequest: Codable {
    public let name: String?
    public let timezone: String?
    public let dueDateWarningThreshold: Int?
    public let emailNotificationsEnabled: Bool?

    public init(
        name: String? = nil,
        timezone: String? = nil,
        dueDateWarningThreshold: Int? = nil,
        emailNotificationsEnabled: Bool? = nil
    ) {
        self.name = name
        self.timezone = timezone
        self.dueDateWarningThreshold = dueDateWarningThreshold
        self.emailNotificationsEnabled = emailNotificationsEnabled
    }

    enum CodingKeys: String, CodingKey {
        case name, timezone
        case dueDateWarningThreshold = "due_date_warning_threshold"
        case emailNotificationsEnabled = "email_notifications_enabled"
    }
}

public struct BindOAuthRequest: Codable {
    public let provider: String
    public let idToken: String?
    public let code: String?
    public let email: String?
    public let name: String?

    public init(provider: String, idToken: String? = nil, code: String? = nil, email: String? = nil, name: String? = nil) {
        self.provider = provider
        self.idToken = idToken
        self.code = code
        self.email = email
        self.name = name
    }

    enum CodingKeys: String, CodingKey {
        case provider, code, email, name
        case idToken = "id_token"
    }
}

public struct UnbindOAuthRequest: Codable {
    public let provider: String

    public init(provider: String) {
        self.provider = provider
    }
}

public struct DeleteAccountRequest: Codable {
    public let confirmation: String

    public init(confirmation: String) {
        self.confirmation = confirmation
    }
}

// MARK: - Task & Subtask
public struct Subtask: Codable, Identifiable {
    public let id: Int64
    public let taskId: Int64
    public var title: String
    public var isCompleted: Bool
    public let orderIndex: Int

    enum CodingKeys: String, CodingKey {
        case id, title
        case taskId = "task_id"
        case isCompleted = "is_completed"
        case orderIndex = "order_index"
    }
}

public struct TaskItem: Codable, Identifiable {
    public let id: Int64
    public let userId: String
    public var title: String
    public var description: String
    public var status: String // backlog, todo, in_progress, review, done
    public var priority: String // low, medium, high, urgent
    public let dueDate: String?
    public let isFocus: Bool
    public let tags: [String]?
    public let projectId: Int64?
    public let projectName: String?
    public let projectColor: String?
    public let lastProgressNote: String?
    public var subtasks: [Subtask]?
    public let subtasksCount: Int?
    public let subtasksCompleted: Int?
    public let progressPercent: Int?

    enum CodingKeys: String, CodingKey {
        case id, title, description, status, priority, tags
        case userId = "user_id"
        case dueDate = "due_date"
        case isFocus = "is_focus"
        case projectId = "project_id"
        case projectName = "project_name"
        case projectColor = "project_color"
        case lastProgressNote = "last_progress_note"
        case subtasks
        case subtasksCount = "subtasks_count"
        case subtasksCompleted = "subtasks_completed"
        case progressPercent = "progress_percent"
    }
}

public struct CreateTaskRequest: Codable {
    public let title: String
    public let description: String
    public let status: String
    public let priority: String
    public let dueDate: String?
    public let isFocus: Bool
    public let tags: [String]
    public let projectId: Int64?
    public let subtasks: [String]

    public init(
        title: String,
        description: String = "",
        status: String = "todo",
        priority: String = "medium",
        dueDate: String? = nil,
        isFocus: Bool = false,
        tags: [String] = [],
        projectId: Int64? = nil,
        subtasks: [String] = []
    ) {
        self.title = title
        self.description = description
        self.status = status
        self.priority = priority
        self.dueDate = dueDate
        self.isFocus = isFocus
        self.tags = tags
        self.projectId = projectId
        self.subtasks = subtasks
    }

    enum CodingKeys: String, CodingKey {
        case title, description, status, priority, tags, subtasks
        case dueDate = "due_date"
        case isFocus = "is_focus"
        case projectId = "project_id"
    }
}

public struct UpdateTaskRequest: Codable {
    public let title: String?
    public let description: String?
    public let status: String?
    public let priority: String?
    public let dueDate: String?
    public let isFocus: Bool?
    public let tags: [String]?
    public let projectId: Int64?

    public init(
        title: String? = nil,
        description: String? = nil,
        status: String? = nil,
        priority: String? = nil,
        dueDate: String? = nil,
        isFocus: Bool? = nil,
        tags: [String]? = nil,
        projectId: Int64? = nil
    ) {
        self.title = title
        self.description = description
        self.status = status
        self.priority = priority
        self.dueDate = dueDate
        self.isFocus = isFocus
        self.tags = tags
        self.projectId = projectId
    }

    enum CodingKeys: String, CodingKey {
        case title, description, status, priority, tags
        case dueDate = "due_date"
        case isFocus = "is_focus"
        case projectId = "project_id"
    }
}

public struct UpdateSubtaskRequest: Codable {
    public let title: String?
    public let isCompleted: Bool?

    enum CodingKeys: String, CodingKey {
        case title
        case isCompleted = "is_completed"
    }
}

// MARK: - Task Activity (Audit Trail & Activity Log)
public struct TaskActivity: Codable, Identifiable {
    public let id: Int64
    public let taskId: Int64
    public let userId: String
    public let activityType: String
    public let description: String
    public let createdAt: String

    enum CodingKeys: String, CodingKey {
        case id, description
        case taskId = "task_id"
        case userId = "user_id"
        case activityType = "activity_type"
        case createdAt = "created_at"
    }
}

public struct CreateTaskActivityRequest: Codable {
    public let description: String
    public let activityType: String?

    public init(description: String, activityType: String? = nil) {
        self.description = description
        self.activityType = activityType
    }

    enum CodingKeys: String, CodingKey {
        case description
        case activityType = "activity_type"
    }
}

// MARK: - Project
public struct Project: Codable, Identifiable {
    public let id: Int64
    public let name: String
    public let color: String
    public let description: String
    public let taskCount: Int?

    enum CodingKeys: String, CodingKey {
        case id, name, color, description
        case taskCount = "task_count"
    }
}

public struct CreateProjectRequest: Codable {
    public let name: String
    public let color: String
    public let description: String

    public init(name: String, color: String = "#6366f1", description: String = "") {
        self.name = name
        self.color = color
        self.description = description
    }
}

public struct UpdateProjectRequest: Codable {
    public let name: String?
    public let color: String?
    public let description: String?

    public init(name: String? = nil, color: String? = nil, description: String? = nil) {
        self.name = name
        self.color = color
        self.description = description
    }
}

// MARK: - Note
public struct Note: Codable, Identifiable {
    public let id: Int64
    public let title: String
    public let content: String
    public let type: String // actionable, knowledge
    public let tags: [String]?
    public let color: String
    public let isPinned: Bool

    enum CodingKeys: String, CodingKey {
        case id, title, content, type, tags, color
        case isPinned = "is_pinned"
    }
}

public struct CreateNoteRequest: Codable {
    public let title: String
    public let content: String
    public let type: String
    public let color: String
    public let isPinned: Bool

    enum CodingKeys: String, CodingKey {
        case title, content, type, color
        case isPinned = "is_pinned"
    }
}

// MARK: - Reminder
public struct Reminder: Codable, Identifiable {
    public let id: Int64
    public let taskId: Int64?
    public let taskTitle: String?
    public let title: String
    public let description: String
    public let remindAt: String
    public var status: String
    public let snoozeCount: Int?

    enum CodingKeys: String, CodingKey {
        case id, title, description, status
        case taskId = "task_id"
        case taskTitle = "task_title"
        case remindAt = "remind_at"
        case snoozeCount = "snooze_count"
    }
}

public struct CreateReminderRequest: Codable {
    public let title: String
    public let description: String?
    public let remindAt: String
    public let taskId: Int64?

    public init(title: String, description: String? = nil, remindAt: String, taskId: Int64? = nil) {
        self.title = title
        self.description = description
        self.remindAt = remindAt
        self.taskId = taskId
    }

    enum CodingKeys: String, CodingKey {
        case title, description
        case remindAt = "remind_at"
        case taskId = "task_id"
    }
}

public struct RegisterDeviceTokenRequest: Codable {
    public let token: String
    public let platform: String
}

// MARK: - Base API Response Envelope
public struct BaseResponse<T: Codable>: Codable {
    public let success: Bool
    public let data: T?
    public let message: String
}

// MARK: - Paginated API Response Envelope
public struct PaginatedResponse<T: Codable>: Codable {
    public let items: [T]
    public let total: Int?
    public let page: Int?
    public let perPage: Int?
    public let totalPages: Int?

    enum CodingKeys: String, CodingKey {
        case items, total, page
        case perPage = "perPage"
        case totalPages = "totalPages"
    }
}

// MARK: - MCP (Model Context Protocol) Tokens
public struct McpToken: Codable, Identifiable, Sendable {
    public let id: Int64
    public let userId: String
    public let name: String
    public let tokenPrefix: String
    public let lastUsedAt: String?
    public let createdAt: String
    public let expiresAt: String?

    public init(
        id: Int64,
        userId: String,
        name: String,
        tokenPrefix: String,
        lastUsedAt: String? = nil,
        createdAt: String,
        expiresAt: String? = nil
    ) {
        self.id = id
        self.userId = userId
        self.name = name
        self.tokenPrefix = tokenPrefix
        self.lastUsedAt = lastUsedAt
        self.createdAt = createdAt
        self.expiresAt = expiresAt
    }

    enum CodingKeys: String, CodingKey {
        case id, name
        case userId = "user_id"
        case tokenPrefix = "token_prefix"
        case lastUsedAt = "last_used_at"
        case createdAt = "created_at"
        case expiresAt = "expires_at"
    }
}

public struct GenerateMcpTokenRequest: Codable, Sendable {
    public let name: String

    public init(name: String) {
        self.name = name
    }
}

public struct GenerateMcpTokenResponse: Codable, Identifiable, Sendable {
    public let id: Int64
    public let name: String
    public let tokenPrefix: String
    public let rawToken: String
    public let createdAt: String
    public let expiresAt: String?

    public init(
        id: Int64,
        name: String,
        tokenPrefix: String,
        rawToken: String,
        createdAt: String,
        expiresAt: String? = nil
    ) {
        self.id = id
        self.name = name
        self.tokenPrefix = tokenPrefix
        self.rawToken = rawToken
        self.createdAt = createdAt
        self.expiresAt = expiresAt
    }

    enum CodingKeys: String, CodingKey {
        case id, name
        case tokenPrefix = "token_prefix"
        case rawToken = "raw_token"
        case createdAt = "created_at"
        case expiresAt = "expires_at"
    }
}

public struct EmptyResponse: Codable {
    public init() {}
}
