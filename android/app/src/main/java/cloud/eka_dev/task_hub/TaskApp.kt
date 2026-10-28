package cloud.eka_dev.task_hub

import android.app.Application
import cloud.eka_dev.task_hub.core.datastore.TokenManager
import cloud.eka_dev.task_hub.core.logger.CrashReporterService
import cloud.eka_dev.task_hub.core.network.ApiClient
import cloud.eka_dev.task_hub.core.network.SSEClient
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch

/*
Application class: Titik mula seluruh lifecycle aplikasi Android.
Di sini kita menginisialisasi singleton dependencies (AppContainer)
agar bersih dan bisa diakses oleh Activity dan ViewModel.
*/
class TaskApp : Application() {

    lateinit var tokenManager: TokenManager
        private set

    lateinit var apiClient: ApiClient
        private set

    lateinit var sseClient: SSEClient
        private set

    override fun onCreate() {
        super.onCreate()
        CrashReporterService.instance.initialize(this)
        tokenManager = TokenManager(this)
        apiClient = ApiClient(tokenManager)
        sseClient = SSEClient(tokenManager)

        CoroutineScope(Dispatchers.IO).launch {
            val userId = tokenManager.getUserId()
            if (!userId.isNullOrEmpty()) {
                CrashReporterService.instance.setUserId(userId)
            }
        }
    }
}
