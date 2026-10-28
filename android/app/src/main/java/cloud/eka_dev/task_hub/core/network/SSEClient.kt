package cloud.eka_dev.task_hub.core.network

import cloud.eka_dev.task_hub.BuildConfig
import cloud.eka_dev.task_hub.core.datastore.TokenManager
import kotlinx.coroutines.channels.awaitClose
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.callbackFlow
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.Response
import okhttp3.sse.EventSource
import okhttp3.sse.EventSourceListener
import okhttp3.sse.EventSources
import java.util.concurrent.TimeUnit

data class SSEEvent(
    val id: String?,
    val type: String?,
    val data: String
)

class SSEClient(
    private val tokenManager: TokenManager,
    private val baseUrl: String = BuildConfig.SSE_URL
) {
    private val sseHttpClient = OkHttpClient.Builder()
        .readTimeout(0, TimeUnit.MILLISECONDS) // infinite for SSE stream
        .build()

    val events: Flow<SSEEvent>
        get() = observeEvents()

    fun observeEvents(): Flow<SSEEvent> = callbackFlow {
        val token = tokenManager.getAccessToken()
        val url = if (!token.isNullOrEmpty()) "$baseUrl?token=$token" else baseUrl

        val request = Request.Builder()
            .url(url)
            .addHeader("Accept", "text/event-stream")
            .build()

        val listener = object : EventSourceListener() {
            override fun onEvent(eventSource: EventSource, id: String?, type: String?, data: String) {
                trySend(SSEEvent(id = id, type = type, data = data))
            }

            override fun onFailure(eventSource: EventSource, t: Throwable?, response: Response?) {
                // Connection dropped or reconnection needed
            }

            override fun onClosed(eventSource: EventSource) {
                close()
            }
        }

        val factory = EventSources.createFactory(sseHttpClient)
        val eventSource = factory.newEventSource(request, listener)

        awaitClose {
            eventSource.cancel()
        }
    }
}
