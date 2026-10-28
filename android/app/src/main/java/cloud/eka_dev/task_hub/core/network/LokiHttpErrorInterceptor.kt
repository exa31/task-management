package cloud.eka_dev.task_hub.core.network

import cloud.eka_dev.task_hub.core.logger.CrashReporterService
import okhttp3.Interceptor
import okhttp3.Response
import java.io.IOException

class LokiHttpErrorInterceptor : Interceptor {
    override fun intercept(chain: Interceptor.Chain): Response {
        val request = chain.request()
        val startTime = System.currentTimeMillis()
        val path = request.url.encodedPath
        val method = request.method

        val response: Response
        try {
            response = chain.proceed(request)
        } catch (e: IOException) {
            val durationMs = System.currentTimeMillis() - startTime
            CrashReporterService.instance.recordHttpError(
                method = method,
                path = path,
                statusCode = 0,
                errorMessage = e.message ?: "Network error",
                durationMs = durationMs
            )
            throw e
        }

        val durationMs = System.currentTimeMillis() - startTime
        if (!response.isSuccessful) {
            val responseBody = try {
                response.peekBody(1024 * 32).string()
            } catch (_: Throwable) {
                null
            }
            CrashReporterService.instance.recordHttpError(
                method = method,
                path = path,
                statusCode = response.code,
                errorMessage = "HTTP ${response.code}: ${response.message}",
                responseBody = responseBody,
                durationMs = durationMs
            )
        }

        return response
    }
}
