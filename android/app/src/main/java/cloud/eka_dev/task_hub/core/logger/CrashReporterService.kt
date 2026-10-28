package cloud.eka_dev.task_hub.core.logger

import android.content.Context
import android.content.SharedPreferences
import android.net.ConnectivityManager
import android.net.Network
import android.net.NetworkCapabilities
import android.net.NetworkRequest
import android.os.Build
import android.util.Log
import cloud.eka_dev.task_hub.BuildConfig
import kotlinx.coroutines.*
import kotlinx.serialization.json.*
import okhttp3.*
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.RequestBody.Companion.toRequestBody
import java.io.PrintWriter
import java.io.StringWriter
import java.text.SimpleDateFormat
import java.util.*
import java.util.concurrent.TimeUnit
import java.util.concurrent.atomic.AtomicBoolean

class CrashReporterService private constructor() {

    companion object {
        val instance: CrashReporterService by lazy { CrashReporterService() }
        private const val TAG = "CrashReporter"
        private const val STORAGE_KEY_QUEUE = "task_hub_loki_offline_queue_v1"
        private const val PREFS_NAME = "crash_reporter_prefs"
        private const val KEY_CACHED_USER_ID = "cached_user_id"
        private const val MAX_QUEUE_SIZE = 100
    }

    private val coroutineScope = CoroutineScope(SupervisorJob() + Dispatchers.IO)
    private val isFlushing = AtomicBoolean(false)
    private var isInitialized = false

    private var preferences: SharedPreferences? = null
    private var cachedUserId: String? = null

    private val json = Json {
        ignoreUnknownKeys = true
        isLenient = true
        encodeDefaults = true
    }

    private val client: OkHttpClient = OkHttpClient.Builder()
        .connectTimeout(10, TimeUnit.SECONDS)
        .readTimeout(10, TimeUnit.SECONDS)
        .writeTimeout(10, TimeUnit.SECONDS)
        .build()

    private val lokiUrl: String
        get() = try {
            BuildConfig.LOKI_URL
        } catch (_: Throwable) {
            "https://loki.eka-dev.cloud/loki/api/v1/push"
        }

    private val environment: String
        get() = if (BuildConfig.DEBUG) "development" else "production"

    private val platformName: String
        get() = "android"

    private val appName: String
        get() = "task-management-android"

    fun setUserId(userId: String?) {
        cachedUserId = userId
        preferences?.edit()?.apply {
            if (userId.isNullOrEmpty()) {
                remove(KEY_CACHED_USER_ID)
            } else {
                putString(KEY_CACHED_USER_ID, userId)
            }
            apply()
        }
    }

    fun initialize(context: Context) {
        if (isInitialized) return
        isInitialized = true

        val appContext = context.applicationContext
        preferences = appContext.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)
        cachedUserId = preferences?.getString(KEY_CACHED_USER_ID, null)

        // Setup global uncaught exception handler
        val previousHandler = Thread.getDefaultUncaughtExceptionHandler()
        Thread.setDefaultUncaughtExceptionHandler { thread, throwable ->
            try {
                recordFatalError(throwable, reason = "Uncaught Exception in thread ${thread.name}")
                // Give a short grace period for sending fatal report before termination
                runBlocking {
                    delay(300)
                }
            } catch (e: Throwable) {
                Log.e(TAG, "Failed to record fatal crash", e)
            } finally {
                previousHandler?.uncaughtException(thread, throwable)
            }
        }

        // Monitor network connectivity changes
        registerNetworkCallback(appContext)

        // Flush queued offline logs on app start
        flushOfflineLogs()
    }

    private fun registerNetworkCallback(context: Context) {
        try {
            val cm = context.getSystemService(Context.CONNECTIVITY_SERVICE) as? ConnectivityManager ?: return
            val request = NetworkRequest.Builder()
                .addCapability(NetworkCapabilities.NET_CAPABILITY_INTERNET)
                .build()

            cm.registerNetworkCallback(request, object : ConnectivityManager.NetworkCallback() {
                override fun onAvailable(network: Network) {
                    flushOfflineLogs()
                }
            })
        } catch (e: Throwable) {
            Log.w(TAG, "Failed to register network callback: ${e.message}")
        }
    }

    fun recordFatalError(throwable: Throwable, reason: String = "Fatal Uncaught Error") {
        recordError(
            throwable = throwable,
            reason = reason,
            fatal = true,
            tag = "FATAL_CRASH"
        )
    }

    fun recordError(
        throwable: Throwable,
        reason: String? = null,
        fatal: Boolean = false,
        tag: String = "ERROR",
        customData: Map<String, Any?>? = null
    ) {
        val sw = StringWriter()
        throwable.printStackTrace(PrintWriter(sw))
        val stackTrace = sw.toString()

        val payload = buildJsonObject {
            put("message", throwable.message ?: throwable.toString())
            put("reason", reason ?: (throwable.localizedMessage ?: ""))
            put("stackTrace", stackTrace)
            put("fatal", fatal)
            put("tag", tag)
            put("timestamp", getIso8601Timestamp())
            if (customData != null) {
                put("data", customData.toJsonObject())
            }
        }

        dispatch(
            level = if (fatal) "CRITICAL" else "ERROR",
            tag = tag,
            body = payload
        )
    }

    fun recordLog(
        message: String,
        level: String = "INFO",
        tag: String = "APP",
        extra: Map<String, Any?>? = null
    ) {
        val payload = buildJsonObject {
            put("message", message)
            put("tag", tag)
            put("timestamp", getIso8601Timestamp())
            if (extra != null) {
                put("extra", extra.toJsonObject())
            }
        }

        dispatch(
            level = level.uppercase(Locale.ROOT),
            tag = tag,
            body = payload
        )
    }

    fun recordHttpError(
        method: String,
        path: String,
        statusCode: Int?,
        errorMessage: String?,
        responseBody: String? = null,
        durationMs: Long? = null
    ) {
        val payload = buildJsonObject {
            put("http_method", method)
            put("path", path)
            put("status_code", statusCode ?: 0)
            put("error", errorMessage ?: "")
            if (durationMs != null) {
                put("duration_ms", durationMs)
            }
            if (!responseBody.isNullOrBlank()) {
                put("response", responseBody.take(4096))
            }
            put("timestamp", getIso8601Timestamp())
        }

        val level = if (statusCode != null && statusCode >= 500) "ERROR" else "WARN"
        dispatch(
            level = level,
            tag = "HTTP_API",
            body = payload
        )
    }

    // MARK: - Internal Dispatcher

    private fun dispatch(
        level: String,
        tag: String,
        body: JsonObject
    ) {
        val userId = if (!cachedUserId.isNullOrEmpty()) cachedUserId!! else "anonymous"
        val nowNs = (System.currentTimeMillis() * 1_000_000).toString()

        val logEntry = buildJsonObject {
            put("timestamp_ns", nowNs)
            put("level", level)
            put("tag", tag)
            put("user_id", userId)
            put("platform", platformName)
            put("environment", environment)
            put("app", appName)
            put("body", body.toString())
        }

        coroutineScope.launch {
            val success = sendEntries(listOf(logEntry))
            if (!success) {
                enqueueOffline(logEntry)
            }
        }
    }

    // MARK: - Offline Storage

    private fun enqueueOffline(entry: JsonObject) {
        val prefs = preferences ?: return
        synchronized(this) {
            try {
                val currentQueueJson = prefs.getString(STORAGE_KEY_QUEUE, "[]") ?: "[]"
                val jsonArray = try {
                    json.parseToJsonElement(currentQueueJson).jsonArray.toMutableList()
                } catch (_: Throwable) {
                    mutableListOf()
                }

                if (jsonArray.size >= MAX_QUEUE_SIZE) {
                    jsonArray.removeAt(0)
                }
                jsonArray.add(entry)

                prefs.edit().putString(STORAGE_KEY_QUEUE, JsonArray(jsonArray).toString()).apply()
            } catch (e: Throwable) {
                Log.w(TAG, "Failed to enqueue offline log", e)
            }
        }
    }

    fun flushOfflineLogs() {
        if (!isFlushing.compareAndSet(false, true)) return

        coroutineScope.launch {
            try {
                val prefs = preferences ?: return@launch
                val queueJson = prefs.getString(STORAGE_KEY_QUEUE, "[]") ?: "[]"
                val jsonArray = try {
                    json.parseToJsonElement(queueJson).jsonArray
                } catch (_: Throwable) {
                    JsonArray(emptyList())
                }

                if (jsonArray.isEmpty()) return@launch

                val entries = jsonArray.mapNotNull { it as? JsonObject }
                if (entries.isNotEmpty()) {
                    val success = sendEntries(entries)
                    if (success) {
                        prefs.edit().remove(STORAGE_KEY_QUEUE).apply()
                        Log.d(TAG, "Flushed ${entries.size} offline log(s) to Loki")
                    }
                }
            } catch (e: Throwable) {
                Log.w(TAG, "Failed to flush offline logs", e)
            } finally {
                isFlushing.set(false)
            }
        }
    }

    // MARK: - Loki POST Request

    private suspend fun sendEntries(entries: List<JsonObject>): Boolean = withContext(Dispatchers.IO) {
        if (entries.isEmpty()) return@withContext true

        val streamsMap = mutableMapOf<String, MutableList<JsonArray>>()
        val labelMap = mutableMapOf<String, JsonObject>()

        for (item in entries) {
            val app = item["app"]?.jsonPrimitive?.contentOrNull ?: appName
            val level = item["level"]?.jsonPrimitive?.contentOrNull ?: "INFO"
            val platform = item["platform"]?.jsonPrimitive?.contentOrNull ?: platformName
            val env = item["environment"]?.jsonPrimitive?.contentOrNull ?: environment
            val tag = item["tag"]?.jsonPrimitive?.contentOrNull ?: "APP"
            val userId = item["user_id"]?.jsonPrimitive?.contentOrNull ?: "anonymous"

            val streamKey = "$app-$level-$platform-$env-$tag"

            if (!streamsMap.containsKey(streamKey)) {
                streamsMap[streamKey] = mutableListOf()
                labelMap[streamKey] = buildJsonObject {
                    put("app", app)
                    put("level", level)
                    put("platform", platform)
                    put("environment", env)
                    put("tag", tag)
                    put("user_id", userId)
                }
            }

            val ts = item["timestamp_ns"]?.jsonPrimitive?.contentOrNull
                ?: (System.currentTimeMillis() * 1_000_000).toString()
            val body = item["body"]?.jsonPrimitive?.contentOrNull ?: "{}"

            val valueArray = buildJsonArray {
                add(ts)
                add(body)
            }
            streamsMap[streamKey]?.add(valueArray)
        }

        val streamsList = buildJsonArray {
            for ((key, values) in streamsMap) {
                val labels = labelMap[key] ?: continue
                add(buildJsonObject {
                    put("stream", labels)
                    put("values", JsonArray(values))
                })
            }
        }

        val payload = buildJsonObject {
            put("streams", streamsList)
        }

        val requestBody = payload.toString().toRequestBody("application/json".toMediaType())
        val request = Request.Builder()
            .url(lokiUrl)
            .post(requestBody)
            .build()

        return@withContext try {
            client.newCall(request).execute().use { response ->
                response.isSuccessful
            }
        } catch (e: Throwable) {
            false
        }
    }

    private fun getIso8601Timestamp(): String {
        val sdf = SimpleDateFormat("yyyy-MM-dd'T'HH:mm:ss.SSS'Z'", Locale.US)
        sdf.timeZone = TimeZone.getTimeZone("UTC")
        return sdf.format(Date())
    }

    private fun Map<String, Any?>.toJsonObject(): JsonObject {
        return buildJsonObject {
            for ((k, v) in this@toJsonObject) {
                when (v) {
                    null -> put(k, JsonNull)
                    is Boolean -> put(k, v)
                    is Number -> put(k, v)
                    is String -> put(k, v)
                    else -> put(k, v.toString())
                }
            }
        }
    }
}
