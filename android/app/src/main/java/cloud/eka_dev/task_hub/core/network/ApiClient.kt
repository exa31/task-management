package cloud.eka_dev.task_hub.core.network

import com.jakewharton.retrofit2.converter.kotlinx.serialization.asConverterFactory
import cloud.eka_dev.task_hub.core.datastore.TokenManager
import cloud.eka_dev.task_hub.core.model.*
import kotlinx.coroutines.runBlocking
import kotlinx.serialization.json.Json
import okhttp3.*
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.logging.HttpLoggingInterceptor
import retrofit2.Retrofit
import retrofit2.http.*
import java.util.concurrent.TimeUnit

interface ApiService {
    // Auth
    @POST("auth/oauth")
    suspend fun oauthLogin(@Body req: OAuthLoginRequest): BaseResponse<AuthResponse>

    @POST("auth/refresh")
    suspend fun refreshToken(@Body req: RefreshTokenRequest): BaseResponse<TokenPair>

    @GET("auth/profile")
    suspend fun getCurrentUser(): BaseResponse<User>

    // Tasks
    @GET("tasks")
    suspend fun getTasks(
        @Query("status") status: String? = null,
        @Query("priority") priority: String? = null,
        @Query("project_id") projectId: Long? = null,
        @Query("search") search: String? = null
    ): BaseResponse<List<Task>>

    @POST("tasks")
    suspend fun createTask(@Body req: CreateTaskRequest): BaseResponse<Task>

    @GET("tasks/{id}")
    suspend fun getTaskById(@Path("id") id: Long): BaseResponse<Task>

    @PUT("tasks/{id}")
    suspend fun updateTask(@Path("id") id: Long, @Body req: UpdateTaskRequest): BaseResponse<Task>

    @DELETE("tasks/{id}")
    suspend fun deleteTask(@Path("id") id: Long): BaseResponse<Unit>

    @POST("tasks/{id}/progress")
    suspend fun logProgress(@Path("id") id: Long, @Body req: LogProgressRequest): BaseResponse<Task>

    @POST("tasks/{id}/subtasks")
    suspend fun createSubtask(@Path("id") id: Long, @Body req: CreateSubtaskRequest): BaseResponse<Subtask>

    @PUT("tasks/{taskId}/subtasks/{subtaskId}")
    suspend fun updateSubtask(
        @Path("taskId") taskId: Long,
        @Path("subtaskId") subtaskId: Long,
        @Body req: UpdateSubtaskRequest
    ): BaseResponse<Subtask>

    @GET("tasks/{id}/activities")
    suspend fun getTaskActivities(@Path("id") taskId: Long): BaseResponse<List<TaskActivity>>

    @POST("tasks/{id}/activities")
    suspend fun createTaskActivity(
        @Path("id") taskId: Long,
        @Body req: CreateTaskActivityRequest
    ): BaseResponse<TaskActivity>

    // Projects
    @GET("projects")
    suspend fun getProjects(): BaseResponse<List<Project>>

    @POST("projects")
    suspend fun createProject(@Body req: CreateProjectRequest): BaseResponse<Project>

    @DELETE("projects/{id}")
    suspend fun deleteProject(@Path("id") id: Long): BaseResponse<Unit>

    // Notes
    @GET("notes")
    suspend fun getNotes(
        @Query("type") type: String? = null,
        @Query("search") search: String? = null
    ): BaseResponse<List<Note>>

    @POST("notes")
    suspend fun createNote(@Body req: CreateNoteRequest): BaseResponse<Note>

    @DELETE("notes/{id}")
    suspend fun deleteNote(@Path("id") id: Long): BaseResponse<Unit>

    @PUT("notes/{id}")
    suspend fun updateNote(@Path("id") id: Long, @Body req: UpdateNoteRequest): BaseResponse<Note>

    @POST("notes/{id}/convert-task")
    suspend fun convertNoteToTask(@Path("id") id: Long, @Body req: ConvertToTaskRequest = ConvertToTaskRequest()): BaseResponse<ConvertNoteResponse>

    // Reminders
    @GET("reminders")
    suspend fun getReminders(@Query("status") status: String? = null): BaseResponse<List<Reminder>>

    @POST("reminders")
    suspend fun createReminder(@Body req: CreateReminderRequest): BaseResponse<Reminder>

    @POST("reminders/device-token")
    suspend fun registerDeviceToken(@Body req: RegisterDeviceTokenRequest): BaseResponse<Unit>

    @POST("reminders/{id}/snooze")
    suspend fun snoozeReminder(@Path("id") id: Long, @Query("minutes") minutes: Int = 15): BaseResponse<Reminder>

    @POST("reminders/{id}/dismiss")
    suspend fun dismissReminder(@Path("id") id: Long): BaseResponse<Reminder>
}

class AuthInterceptor(private val tokenManager: TokenManager) : Interceptor {
    override fun intercept(chain: Interceptor.Chain): Response {
        val original = chain.request()
        val token = runBlocking { tokenManager.getAccessToken() }
        val builder = original.newBuilder()
        if (!token.isNullOrEmpty()) {
            builder.header("Authorization", "Bearer $token")
        }
        return chain.proceed(builder.build())
    }
}

class ApiClient(
    val tokenManager: TokenManager,
    baseUrl: String = BuildConfig.BASE_URL
) {
    private val json = Json {
        ignoreUnknownKeys = true
        coerceInputValues = true
        isLenient = true
    }

    val okHttpClient: OkHttpClient = OkHttpClient.Builder()
        .addInterceptor(AuthInterceptor(tokenManager))
        .addInterceptor(HttpLoggingInterceptor().apply {
            level = HttpLoggingInterceptor.Level.BODY
        })
        .connectTimeout(30, TimeUnit.SECONDS)
        .readTimeout(30, TimeUnit.SECONDS)
        .build()

    val apiService: ApiService = Retrofit.Builder()
        .baseUrl(baseUrl)
        .client(okHttpClient)
        .addConverterFactory(json.asConverterFactory("application/json".toMediaType()))
        .build()
        .create(ApiService::class.java)
}
