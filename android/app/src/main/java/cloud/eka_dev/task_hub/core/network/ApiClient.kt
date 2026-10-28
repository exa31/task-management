package cloud.eka_dev.task_hub.core.network

import com.jakewharton.retrofit2.converter.kotlinx.serialization.asConverterFactory
import cloud.eka_dev.task_hub.BuildConfig
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

    @PUT("auth/profile")
    suspend fun updateProfile(@Body req: UpdateProfileRequest): BaseResponse<User>

    @Multipart
    @POST("auth/avatar")
    suspend fun uploadAvatar(@Part avatar: MultipartBody.Part): BaseResponse<User>

    @POST("auth/bind-oauth")
    suspend fun bindOAuth(@Body req: BindOAuthRequest): BaseResponse<User>

    @POST("auth/unbind-oauth")
    suspend fun unbindOAuth(@Body req: UnbindOAuthRequest): BaseResponse<User>

    @HTTP(method = "DELETE", path = "auth/account", hasBody = true)
    suspend fun deleteAccount(@Body req: DeleteAccountRequest): BaseResponse<Unit>

    // Tasks
    @GET("tasks")
    suspend fun getTasks(
        @Query("status") status: String? = null,
        @Query("priority") priority: String? = null,
        @Query("project_id") projectId: Long? = null,
        @Query("search") search: String? = null
    ): BaseResponse<List<Task>>

    @GET("tasks")
    suspend fun getTasksPaginated(
        @Query("paginate") paginate: String = "cursor",
        @Query("limit") limit: Int = 15,
        @Query("cursor") cursor: Long? = null,
        @Query("status") status: String? = null,
        @Query("priority") priority: String? = null,
        @Query("project_id") projectId: Long? = null,
        @Query("search") search: String? = null
    ): BaseResponse<TaskCursorResponse>

    @POST("tasks")
    suspend fun createTask(@Body req: CreateTaskRequest): BaseResponse<Task>

    @GET("tasks/{id}")
    suspend fun getTaskById(@Path("id") id: Long): BaseResponse<Task>

    @PUT("tasks/{id}")
    suspend fun updateTask(@Path("id") id: Long, @Body req: UpdateTaskRequest): BaseResponse<Task>

    @POST("tasks/{id}/focus")
    suspend fun toggleTaskFocus(@Path("id") id: Long): BaseResponse<Task>

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

    @DELETE("tasks/{taskId}/subtasks/{subtaskId}")
    suspend fun deleteSubtask(
        @Path("taskId") taskId: Long,
        @Path("subtaskId") subtaskId: Long
    ): BaseResponse<Unit>

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

    @PUT("projects/{id}")
    suspend fun updateProject(@Path("id") id: Long, @Body req: UpdateProjectRequest): BaseResponse<Project>

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

    @PUT("reminders/{id}")
    suspend fun updateReminder(@Path("id") id: Long, @Body req: UpdateReminderRequest): BaseResponse<Reminder>

    @DELETE("reminders/{id}")
    suspend fun deleteReminder(@Path("id") id: Long): BaseResponse<Unit>

    @POST("reminders/device-token")
    suspend fun registerDeviceToken(@Body req: RegisterDeviceTokenRequest): BaseResponse<Unit>

    @POST("reminders/{id}/snooze")
    suspend fun snoozeReminder(@Path("id") id: Long, @Query("minutes") minutes: Int = 15): BaseResponse<Reminder>

    @POST("reminders/{id}/dismiss")
    suspend fun dismissReminder(@Path("id") id: Long): BaseResponse<Reminder>

    // MCP Tokens
    @GET("tokens")
    suspend fun getMcpTokens(): BaseResponse<List<McpToken>>

    @POST("tokens")
    suspend fun generateMcpToken(@Body req: GenerateMcpTokenRequest): BaseResponse<GenerateMcpTokenResponse>

    @DELETE("tokens/{id}")
    suspend fun deleteMcpToken(@Path("id") id: Long): BaseResponse<Unit>

    // GraphQL for Calendar
    @POST("graphql")
    suspend fun executeGraphQL(@Body req: GraphQLRequest): GraphQLResponse<CalendarGraphQLData>
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

class TokenAuthenticator(
    private val tokenManager: TokenManager,
    private val baseUrl: String,
    private val json: Json
) : Authenticator {
    override fun authenticate(route: Route?, response: Response): Request? {
        if (response.responseCount > 2) {
            return null
        }

        val refreshToken = runBlocking { tokenManager.getRefreshToken() } ?: return null

        val refreshClient = OkHttpClient.Builder()
            .connectTimeout(15, TimeUnit.SECONDS)
            .readTimeout(15, TimeUnit.SECONDS)
            .build()

        val requestBody = json.encodeToString(
            RefreshTokenRequest.serializer(),
            RefreshTokenRequest(refreshToken = refreshToken)
        ).let { RequestBody.create("application/json".toMediaType(), it) }

        val refreshRequest = Request.Builder()
            .url("${baseUrl.trimEnd('/')}/auth/refresh")
            .post(requestBody)
            .build()

        try {
            val refreshResponse = refreshClient.newCall(refreshRequest).execute()
            if (refreshResponse.isSuccessful) {
                val bodyString = refreshResponse.body?.string() ?: return null
                val baseRes = json.decodeFromString(
                    BaseResponse.serializer(TokenPair.serializer()),
                    bodyString
                )
                val newTokens = baseRes.data
                if (newTokens != null) {
                    runBlocking {
                        tokenManager.saveTokens(newTokens.accessToken, newTokens.refreshToken)
                    }
                    return response.request.newBuilder()
                        .header("Authorization", "Bearer ${newTokens.accessToken}")
                        .build()
                }
            } else {
                runBlocking {
                    tokenManager.clearSession()
                }
            }
        } catch (e: Exception) {
            // Error during refresh
        }
        return null
    }

    private val Response.responseCount: Int
        get() {
            var result = 1
            var prior = priorResponse
            while (prior != null) {
                result++
                prior = prior.priorResponse
            }
            return result
        }
}

class ApiClient(
    val tokenManager: TokenManager,
    baseUrl: String = BuildConfig.BASE_URL
) {
    val json = Json {
        ignoreUnknownKeys = true
        coerceInputValues = true
        isLenient = true
    }

    val okHttpClient: OkHttpClient = OkHttpClient.Builder()
        .addInterceptor(AuthInterceptor(tokenManager))
        .addInterceptor(LokiHttpErrorInterceptor())
        .authenticator(TokenAuthenticator(tokenManager, baseUrl, json))
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
