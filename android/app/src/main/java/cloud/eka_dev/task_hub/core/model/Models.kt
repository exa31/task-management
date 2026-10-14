package cloud.eka_dev.task_hub.core.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class User(
    val id: String,
    val name: String,
    val email: String,
    @SerialName("avatar_url") val avatarUrl: String? = null,
    val provider: String? = null,
    @SerialName("google_bound") val googleBound: Boolean = false,
    @SerialName("apple_bound") val appleBound: Boolean = false,
    val timezone: String? = null,
    @SerialName("due_date_warning_threshold") val dueDateWarningThreshold: Int = 24,
    @SerialName("email_notifications_enabled") val emailNotificationsEnabled: Boolean = true
)

@Serializable
data class TokenPair(
    @SerialName("access_token") val accessToken: String,
    @SerialName("refresh_token") val refreshToken: String,
    @SerialName("expires_at") val expiresAt: String? = null,
    @SerialName("token_type") val tokenType: String = "Bearer"
)

@Serializable
data class AuthResponse(
    val user: User,
    val tokens: TokenPair? = null
)

@Serializable
data class OAuthLoginRequest(
    val provider: String,
    @SerialName("id_token") val idToken: String? = null,
    val code: String? = null,
    val email: String? = null,
    val name: String? = null,
    @SerialName("avatar_url") val avatarUrl: String? = null
)

@Serializable
data class RefreshTokenRequest(
    @SerialName("refresh_token") val refreshToken: String
)

@Serializable
data class Subtask(
    val id: Long,
    @SerialName("task_id") val taskId: Long,
    val title: String,
    @SerialName("is_completed") val isCompleted: Boolean,
    @SerialName("order_index") val orderIndex: Int = 0
)

@Serializable
data class Task(
    val id: Long,
    @SerialName("user_id") val userId: String,
    val title: String,
    val description: String = "",
    val status: String = "todo", // backlog, todo, in_progress, review, done
    val priority: String = "medium", // low, medium, high, urgent
    @SerialName("due_date") val dueDate: String? = null,
    @SerialName("is_focus") val isFocus: Boolean = false,
    val tags: List<String> = emptyList(),
    @SerialName("project_id") val projectId: Long? = null,
    @SerialName("project_name") val projectName: String? = null,
    @SerialName("project_color") val projectColor: String? = null,
    @SerialName("last_progress_note") val lastProgressNote: String? = null,
    val subtasks: List<Subtask> = emptyList(),
    @SerialName("subtasks_count") val subtasksCount: Int = 0,
    @SerialName("subtasks_completed") val subtasksCompleted: Int = 0,
    @SerialName("progress_percent") val progressPercent: Int = 0
)

@Serializable
data class CreateTaskRequest(
    val title: String,
    val description: String = "",
    val status: String = "todo",
    val priority: String = "medium",
    @SerialName("due_date") val dueDate: String? = null,
    @SerialName("is_focus") val isFocus: Boolean = false,
    val tags: List<String> = emptyList(),
    @SerialName("project_id") val projectId: Long? = null,
    val subtasks: List<String> = emptyList()
)

@Serializable
data class UpdateTaskRequest(
    val title: String? = null,
    val description: String? = null,
    val status: String? = null,
    val priority: String? = null,
    @SerialName("due_date") val dueDate: String? = null,
    @SerialName("is_focus") val isFocus: Boolean? = null,
    val tags: List<String>? = null,
    @SerialName("project_id") val projectId: Long? = null,
    @SerialName("last_progress_note") val lastProgressNote: String? = null
)

@Serializable
data class LogProgressRequest(
    @SerialName("progress_note") val progressNote: String,
    val status: String? = null,
    @SerialName("completed_subtask_ids") val completedSubtaskIds: List<Long> = emptyList()
)

@Serializable
data class CreateSubtaskRequest(
    val title: String
)

@Serializable
data class UpdateSubtaskRequest(
    val title: String? = null,
    @SerialName("is_completed") val isCompleted: Boolean? = null
)

@Serializable
data class TaskActivity(
    val id: Long,
    @SerialName("task_id") val taskId: Long,
    @SerialName("user_id") val userId: String,
    @SerialName("activity_type") val activityType: String,
    val description: String,
    @SerialName("created_at") val createdAt: String
)

@Serializable
data class CreateTaskActivityRequest(
    val description: String,
    @SerialName("activity_type") val activityType: String? = null
)

@Serializable
data class Project(
    val id: Long,
    val name: String,
    val color: String = "#3B82F6",
    val description: String = "",
    @SerialName("task_count") val taskCount: Int = 0
)

@Serializable
data class CreateProjectRequest(
    val name: String,
    val color: String = "#3B82F6",
    val description: String = ""
)

@Serializable
data class Note(
    val id: Long,
    val title: String,
    val content: String = "",
    val type: String = "actionable", // actionable, knowledge
    val tags: List<String> = emptyList(),
    val color: String = "#F3F4F6",
    @SerialName("is_pinned") val isPinned: Boolean = false
)

@Serializable
data class CreateNoteRequest(
    val title: String,
    val content: String = "",
    val type: String = "actionable",
    val tags: List<String> = emptyList(),
    val color: String = "#F3F4F6",
    @SerialName("is_pinned") val isPinned: Boolean = false
)

@Serializable
data class UpdateNoteRequest(
    val title: String? = null,
    val content: String? = null,
    val type: String? = null,
    val tags: List<String>? = null,
    val color: String? = null,
    @SerialName("is_pinned") val isPinned: Boolean? = null
)

@Serializable
data class ConvertToTaskRequest(
    val priority: String? = null,
    @SerialName("project_id") val projectId: Long? = null,
    @SerialName("due_date") val dueDate: String? = null
)

@Serializable
data class ConvertNoteResponse(
    val note: Note? = null,
    @SerialName("task_id") val taskId: Long? = null
)

@Serializable
data class Reminder(
    val id: Long,
    @SerialName("task_id") val taskId: Long? = null,
    @SerialName("task_title") val taskTitle: String? = null,
    val title: String,
    val description: String = "",
    @SerialName("remind_at") val remindAt: String,
    val status: String = "pending", // pending, sent, dismissed
    @SerialName("snooze_count") val snoozeCount: Int = 0
)

@Serializable
data class CreateReminderRequest(
    @SerialName("task_id") val taskId: Long? = null,
    val title: String,
    val description: String = "",
    @SerialName("remind_at") val remindAt: String
)

@Serializable
data class RegisterDeviceTokenRequest(
    val token: String,
    val platform: String = "android"
)

@Serializable
data class BaseResponse<T>(
    val success: Boolean = true,
    val data: T? = null,
    val message: String = ""
)
