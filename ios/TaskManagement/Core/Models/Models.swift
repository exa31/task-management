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
    public let reminderSound: String?

    enum CodingKeys: String, CodingKey {
        case id, name, email, provider, timezone
        case avatarUrl = "avatar_url"
        case googleBound = "google_bound"
        case appleBound = "apple_bound"
        case dueDateWarningThreshold = "due_date_warning_threshold"
        case emailNotificationsEnabled = "email_notifications_enabled"
        case reminderSound = "reminder_sound"
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
    public let reminderSound: String?

    public init(
        name: String? = nil,
        timezone: String? = nil,
        dueDateWarningThreshold: Int? = nil,
        emailNotificationsEnabled: Bool? = nil,
        reminderSound: String? = nil
    ) {
        self.name = name
        self.timezone = timezone
        self.dueDateWarningThreshold = dueDateWarningThreshold
        self.emailNotificationsEnabled = emailNotificationsEnabled
        self.reminderSound = reminderSound
    }

    enum CodingKeys: String, CodingKey {
        case name, timezone
        case dueDateWarningThreshold = "due_date_warning_threshold"
        case emailNotificationsEnabled = "email_notifications_enabled"
        case reminderSound = "reminder_sound"
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
public struct Subtask: Codable, Identifiable, Sendable {
    public let id: Int64
    public let taskId: Int64
    public var title: String
    public var description: String?
    public var isCompleted: Bool
    public let orderIndex: Int
    public let dueDate: String?

    public init(id: Int64, taskId: Int64, title: String, description: String? = nil, isCompleted: Bool, orderIndex: Int, dueDate: String? = nil) {
        self.id = id
        self.taskId = taskId
        self.title = title
        self.description = description
        self.isCompleted = isCompleted
        self.orderIndex = orderIndex
        self.dueDate = dueDate
    }

    enum CodingKeys: String, CodingKey {
        case id, title, description
        case taskId = "task_id"
        case isCompleted = "is_completed"
        case orderIndex = "order_index"
        case dueDate = "due_date"
    }
}

public struct TaskItem: Codable, Identifiable, Sendable {
    public let id: Int64
    public let userId: String
    public var title: String
    public var description: String
    public var status: String // backlog, todo, in_progress, review, done
    public var priority: String // low, medium, high, urgent
    public let dueDate: String?
    public var isFocus: Bool
    public let tags: [String]?
    public let projectId: Int64?
    public let projectName: String?
    public let projectColor: String?
    public let lastProgressNote: String?
    public var subtasks: [Subtask]?
    public var subtasksCount: Int?
    public var subtasksCompleted: Int?
    public var progressPercent: Int?
    public let isRecurring: Bool?
    public let recurrencePattern: String?
    public let recurrenceInterval: Int?
    public let recurrenceEndDate: String?
    public let recurrenceLastGeneratedAt: String?
    public let recurringTaskId: Int64?

    public init(
        id: Int64,
        userId: String,
        title: String,
        description: String,
        status: String,
        priority: String,
        dueDate: String?,
        isFocus: Bool,
        tags: [String]?,
        projectId: Int64?,
        projectName: String?,
        projectColor: String?,
        lastProgressNote: String?,
        subtasks: [Subtask]?,
        subtasksCount: Int?,
        subtasksCompleted: Int?,
        progressPercent: Int?,
        isRecurring: Bool?,
        recurrencePattern: String?,
        recurrenceInterval: Int?,
        recurrenceEndDate: String?,
        recurrenceLastGeneratedAt: String?,
        recurringTaskId: Int64? = nil
    ) {
        self.id = id
        self.userId = userId
        self.title = title
        self.description = description
        self.status = status
        self.priority = priority
        self.dueDate = dueDate
        self.isFocus = isFocus
        self.tags = tags
        self.projectId = projectId
        self.projectName = projectName
        self.projectColor = projectColor
        self.lastProgressNote = lastProgressNote
        self.subtasks = subtasks
        self.subtasksCount = subtasksCount
        self.subtasksCompleted = subtasksCompleted
        self.progressPercent = progressPercent
        self.isRecurring = isRecurring
        self.recurrencePattern = recurrencePattern
        self.recurrenceInterval = recurrenceInterval
        self.recurrenceEndDate = recurrenceEndDate
        self.recurrenceLastGeneratedAt = recurrenceLastGeneratedAt
        self.recurringTaskId = recurringTaskId
    }

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
        case isRecurring = "is_recurring"
        case recurrencePattern = "recurrence_pattern"
        case recurrenceInterval = "recurrence_interval"
        case recurrenceEndDate = "recurrence_end_date"
        case recurrenceLastGeneratedAt = "recurrence_last_generated_at"
        case recurringTaskId = "recurring_task_id"
    }
}

public struct RecurringSubtaskItem: Codable, Identifiable, Sendable {
    public let id: Int64
    public let recurringTaskId: Int64
    public var title: String
    public var orderIndex: Int

    enum CodingKeys: String, CodingKey {
        case id, title
        case recurringTaskId = "recurring_task_id"
        case orderIndex = "order_index"
    }
}

public struct RecurringTaskItem: Codable, Identifiable, Sendable {
    public let id: Int64
    public let userId: String
    public var title: String
    public var description: String
    public var priority: String
    public let projectId: Int64?
    public let projectName: String?
    public let projectColor: String?
    public let tags: [String]?
    public var recurrencePattern: String
    public var recurrenceInterval: Int
    public let recurrenceEndDate: String?
    public let dueTime: String?
    public var isActive: Bool
    public let lastGeneratedAt: String?
    public var subtasks: [RecurringSubtaskItem]?

    enum CodingKeys: String, CodingKey {
        case id, title, description, priority, tags
        case userId = "user_id"
        case projectId = "project_id"
        case projectName = "project_name"
        case projectColor = "project_color"
        case recurrencePattern = "recurrence_pattern"
        case recurrenceInterval = "recurrence_interval"
        case recurrenceEndDate = "recurrence_end_date"
        case dueTime = "due_time"
        case isActive = "is_active"
        case lastGeneratedAt = "last_generated_at"
        case subtasks
    }
}

public struct CreateRecurringSubtaskItem: Codable, Sendable {
    public let title: String
    public let orderIndex: Int

    public init(title: String, orderIndex: Int) {
        self.title = title
        self.orderIndex = orderIndex
    }

    enum CodingKeys: String, CodingKey {
        case title
        case orderIndex = "order_index"
    }
}

public struct CreateRecurringTaskRequest: Codable, Sendable {
    public let title: String
    public let description: String?
    public let priority: String?
    public let projectId: Int64?
    public let recurrencePattern: String?
    public let recurrenceInterval: Int?
    public let dueTime: String?
    public let isActive: Bool?
    public let subtasks: [CreateRecurringSubtaskItem]?

    public init(
        title: String,
        description: String? = nil,
        priority: String? = "medium",
        projectId: Int64? = nil,
        recurrencePattern: String? = "daily",
        recurrenceInterval: Int? = 1,
        dueTime: String? = nil,
        isActive: Bool? = true,
        subtasks: [CreateRecurringSubtaskItem]? = nil
    ) {
        self.title = title
        self.description = description
        self.priority = priority
        self.projectId = projectId
        self.recurrencePattern = recurrencePattern
        self.recurrenceInterval = recurrenceInterval
        self.dueTime = dueTime
        self.isActive = isActive
        self.subtasks = subtasks
    }

    enum CodingKeys: String, CodingKey {
        case title, description, priority
        case projectId = "project_id"
        case recurrencePattern = "recurrence_pattern"
        case recurrenceInterval = "recurrence_interval"
        case dueTime = "due_time"
        case isActive = "is_active"
        case subtasks
    }
}

public struct UpdateRecurringTaskRequest: Codable, Sendable {
    public let title: String?
    public let description: String?
    public let priority: String?
    public let projectId: Int64?
    public let recurrencePattern: String?
    public let recurrenceInterval: Int?
    public let dueTime: String?
    public let isActive: Bool?
    public let subtasks: [CreateRecurringSubtaskItem]?

    public init(
        title: String? = nil,
        description: String? = nil,
        priority: String? = nil,
        projectId: Int64? = nil,
        recurrencePattern: String? = nil,
        recurrenceInterval: Int? = nil,
        dueTime: String? = nil,
        isActive: Bool? = nil,
        subtasks: [CreateRecurringSubtaskItem]? = nil
    ) {
        self.title = title
        self.description = description
        self.priority = priority
        self.projectId = projectId
        self.recurrencePattern = recurrencePattern
        self.recurrenceInterval = recurrenceInterval
        self.dueTime = dueTime
        self.isActive = isActive
        self.subtasks = subtasks
    }

    enum CodingKeys: String, CodingKey {
        case title, description, priority
        case projectId = "project_id"
        case recurrencePattern = "recurrence_pattern"
        case recurrenceInterval = "recurrence_interval"
        case dueTime = "due_time"
        case isActive = "is_active"
        case subtasks
    }
}

public struct TaskCursorResponse: Codable {
    public let items: [TaskItem]
    public let nextCursor: Int64?
    public let hasMore: Bool

    enum CodingKeys: String, CodingKey {
        case items
        case nextCursor = "next_cursor"
        case hasMore = "has_more"
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
    public let isRecurring: Bool
    public let recurrencePattern: String?
    public let recurrenceInterval: Int?
    public let recurrenceEndDate: String?

    public init(
        title: String,
        description: String = "",
        status: String = "todo",
        priority: String = "medium",
        dueDate: String? = nil,
        isFocus: Bool = false,
        tags: [String] = [],
        projectId: Int64? = nil,
        subtasks: [String] = [],
        isRecurring: Bool = false,
        recurrencePattern: String? = nil,
        recurrenceInterval: Int? = nil,
        recurrenceEndDate: String? = nil
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
        self.isRecurring = isRecurring
        self.recurrencePattern = recurrencePattern
        self.recurrenceInterval = recurrenceInterval
        self.recurrenceEndDate = recurrenceEndDate
    }

    enum CodingKeys: String, CodingKey {
        case title, description, status, priority, tags, subtasks
        case dueDate = "due_date"
        case isFocus = "is_focus"
        case projectId = "project_id"
        case isRecurring = "is_recurring"
        case recurrencePattern = "recurrence_pattern"
        case recurrenceInterval = "recurrence_interval"
        case recurrenceEndDate = "recurrence_end_date"
    }
}

public struct UpdateTaskRequest: Codable {
    public let title: String?
    public let description: String?
    public let status: String?
    public let priority: String?
    public let dueDate: String?
    public let clearDueDate: Bool
    public let isFocus: Bool?
    public let tags: [String]?
    public let projectId: Int64?
    public let clearProjectId: Bool
    public let isRecurring: Bool?
    public let recurrencePattern: String?
    public let recurrenceInterval: Int?
    public let recurrenceEndDate: String?
    public let clearRecurrenceEndDate: Bool

    public init(
        title: String? = nil,
        description: String? = nil,
        status: String? = nil,
        priority: String? = nil,
        dueDate: String? = nil,
        clearDueDate: Bool = false,
        isFocus: Bool? = nil,
        tags: [String]? = nil,
        projectId: Int64? = nil,
        clearProjectId: Bool = false,
        isRecurring: Bool? = nil,
        recurrencePattern: String? = nil,
        recurrenceInterval: Int? = nil,
        recurrenceEndDate: String? = nil,
        clearRecurrenceEndDate: Bool = false
    ) {
        self.title = title
        self.description = description
        self.status = status
        self.priority = priority
        self.dueDate = dueDate
        self.clearDueDate = clearDueDate
        self.isFocus = isFocus
        self.tags = tags
        self.projectId = projectId
        self.clearProjectId = clearProjectId
        self.isRecurring = isRecurring
        self.recurrencePattern = recurrencePattern
        self.recurrenceInterval = recurrenceInterval
        self.recurrenceEndDate = recurrenceEndDate
        self.clearRecurrenceEndDate = clearRecurrenceEndDate
    }

    enum CodingKeys: String, CodingKey {
        case title, description, status, priority, tags
        case dueDate = "due_date"
        case isFocus = "is_focus"
        case projectId = "project_id"
        case isRecurring = "is_recurring"
        case recurrencePattern = "recurrence_pattern"
        case recurrenceInterval = "recurrence_interval"
        case recurrenceEndDate = "recurrence_end_date"
    }

    public func encode(to encoder: Encoder) throws {
        var container = encoder.container(keyedBy: CodingKeys.self)
        try container.encodeIfPresent(title, forKey: .title)
        try container.encodeIfPresent(description, forKey: .description)
        try container.encodeIfPresent(status, forKey: .status)
        try container.encodeIfPresent(priority, forKey: .priority)
        try container.encodeIfPresent(isFocus, forKey: .isFocus)
        try container.encodeIfPresent(tags, forKey: .tags)
        try container.encodeIfPresent(isRecurring, forKey: .isRecurring)
        try container.encodeIfPresent(recurrencePattern, forKey: .recurrencePattern)
        try container.encodeIfPresent(recurrenceInterval, forKey: .recurrenceInterval)

        if clearDueDate {
            try container.encodeNil(forKey: .dueDate)
        } else {
            try container.encodeIfPresent(dueDate, forKey: .dueDate)
        }

        if clearProjectId {
            try container.encodeNil(forKey: .projectId)
        } else {
            try container.encodeIfPresent(projectId, forKey: .projectId)
        }

        if clearRecurrenceEndDate {
            try container.encodeNil(forKey: .recurrenceEndDate)
        } else {
            try container.encodeIfPresent(recurrenceEndDate, forKey: .recurrenceEndDate)
        }
    }

    public init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        title = try container.decodeIfPresent(String.self, forKey: .title)
        description = try container.decodeIfPresent(String.self, forKey: .description)
        status = try container.decodeIfPresent(String.self, forKey: .status)
        priority = try container.decodeIfPresent(String.self, forKey: .priority)
        dueDate = try container.decodeIfPresent(String.self, forKey: .dueDate)
        clearDueDate = false
        isFocus = try container.decodeIfPresent(Bool.self, forKey: .isFocus)
        tags = try container.decodeIfPresent([String].self, forKey: .tags)
        projectId = try container.decodeIfPresent(Int64.self, forKey: .projectId)
        clearProjectId = false
        isRecurring = try container.decodeIfPresent(Bool.self, forKey: .isRecurring)
        recurrencePattern = try container.decodeIfPresent(String.self, forKey: .recurrencePattern)
        recurrenceInterval = try container.decodeIfPresent(Int.self, forKey: .recurrenceInterval)
        recurrenceEndDate = try container.decodeIfPresent(String.self, forKey: .recurrenceEndDate)
        clearRecurrenceEndDate = false
    }
}

public struct CreateSubtaskRequest: Codable {
    public let title: String
    public let description: String?
    public let dueDate: String?

    public init(title: String, description: String? = nil, dueDate: String? = nil) {
        self.title = title
        self.description = description
        self.dueDate = dueDate
    }

    enum CodingKeys: String, CodingKey {
        case title, description
        case dueDate = "due_date"
    }
}

public struct UpdateSubtaskRequest: Codable {
    public let title: String?
    public let description: String?
    public let isCompleted: Bool?
    public let dueDate: String?
    public let clearDueDate: Bool

    public init(
        title: String? = nil,
        description: String? = nil,
        isCompleted: Bool? = nil,
        dueDate: String? = nil,
        clearDueDate: Bool = false
    ) {
        self.title = title
        self.description = description
        self.isCompleted = isCompleted
        self.dueDate = dueDate
        self.clearDueDate = clearDueDate
    }

    enum CodingKeys: String, CodingKey {
        case title, description
        case isCompleted = "is_completed"
        case dueDate = "due_date"
    }

    public func encode(to encoder: Encoder) throws {
        var container = encoder.container(keyedBy: CodingKeys.self)
        try container.encodeIfPresent(title, forKey: .title)
        try container.encodeIfPresent(description, forKey: .description)
        try container.encodeIfPresent(isCompleted, forKey: .isCompleted)
        if clearDueDate {
            try container.encodeNil(forKey: .dueDate)
        } else {
            try container.encodeIfPresent(dueDate, forKey: .dueDate)
        }
    }

    public init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        title = try container.decodeIfPresent(String.self, forKey: .title)
        description = try container.decodeIfPresent(String.self, forKey: .description)
        isCompleted = try container.decodeIfPresent(Bool.self, forKey: .isCompleted)
        dueDate = try container.decodeIfPresent(String.self, forKey: .dueDate)
        clearDueDate = false
    }
}

public struct WidgetBundleResponse: Codable, Sendable {
    public let focusTasks: [TaskItem]
    public let upcomingDeadlines: [TaskItem]
    public let completedToday: Int
    public let totalToday: Int

    enum CodingKeys: String, CodingKey {
        case focusTasks = "focus_tasks"
        case upcomingDeadlines = "upcoming_deadlines"
        case completedToday = "completed_today"
        case totalToday = "total_today"
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

    public init(
        title: String,
        content: String,
        type: String,
        color: String = "#6366F1",
        isPinned: Bool = false
    ) {
        self.title = title
        self.content = content
        self.type = type
        self.color = color
        self.isPinned = isPinned
    }

    enum CodingKeys: String, CodingKey {
        case title, content, type, color
        case isPinned = "is_pinned"
    }
}

public struct UpdateNoteRequest: Codable {
    public let title: String?
    public let content: String?
    public let type: String?
    public let tags: [String]?
    public let color: String?
    public let isPinned: Bool?

    public init(
        title: String? = nil,
        content: String? = nil,
        type: String? = nil,
        tags: [String]? = nil,
        color: String? = nil,
        isPinned: Bool? = nil
    ) {
        self.title = title
        self.content = content
        self.type = type
        self.tags = tags
        self.color = color
        self.isPinned = isPinned
    }

    enum CodingKeys: String, CodingKey {
        case title, content, type, tags, color
        case isPinned = "is_pinned"
    }
}

public struct ConvertNoteToTaskRequest: Codable {
    public let priority: String?
    public let projectId: Int64?
    public let dueDate: String?

    public init(priority: String? = nil, projectId: Int64? = nil, dueDate: String? = nil) {
        self.priority = priority
        self.projectId = projectId
        self.dueDate = dueDate
    }

    enum CodingKeys: String, CodingKey {
        case priority
        case projectId = "project_id"
        case dueDate = "due_date"
    }
}

public struct ConvertNoteResponse: Codable {
    public let note: Note
    public let taskId: Int64?

    enum CodingKeys: String, CodingKey {
        case note
        case taskId = "task_id"
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

public struct UpdateReminderRequest: Codable {
    public let title: String?
    public let description: String?
    public let remindAt: String?
    public let status: String?
    public let taskId: Int64?
    public let clearTaskId: Bool

    public init(
        title: String? = nil,
        description: String? = nil,
        remindAt: String? = nil,
        status: String? = nil,
        taskId: Int64? = nil,
        clearTaskId: Bool = false
    ) {
        self.title = title
        self.description = description
        self.remindAt = remindAt
        self.status = status
        self.taskId = taskId
        self.clearTaskId = clearTaskId
    }

    enum CodingKeys: String, CodingKey {
        case title, description, status
        case remindAt = "remind_at"
        case taskId = "task_id"
        case clearTaskId = "clear_task_id"
    }

    public func encode(to encoder: Encoder) throws {
        var container = encoder.container(keyedBy: CodingKeys.self)
        try container.encodeIfPresent(title, forKey: .title)
        try container.encodeIfPresent(description, forKey: .description)
        try container.encodeIfPresent(remindAt, forKey: .remindAt)
        try container.encodeIfPresent(status, forKey: .status)
        try container.encode(clearTaskId, forKey: .clearTaskId)
        if clearTaskId {
            try container.encodeNil(forKey: .taskId)
        } else {
            try container.encodeIfPresent(taskId, forKey: .taskId)
        }
    }

    public init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        title = try container.decodeIfPresent(String.self, forKey: .title)
        description = try container.decodeIfPresent(String.self, forKey: .description)
        remindAt = try container.decodeIfPresent(String.self, forKey: .remindAt)
        status = try container.decodeIfPresent(String.self, forKey: .status)
        taskId = try container.decodeIfPresent(Int64.self, forKey: .taskId)
        clearTaskId = try container.decodeIfPresent(Bool.self, forKey: .clearTaskId) ?? false
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
