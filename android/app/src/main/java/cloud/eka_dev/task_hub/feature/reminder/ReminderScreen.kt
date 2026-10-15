package cloud.eka_dev.task_hub.feature.reminder

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import cloud.eka_dev.task_hub.core.designsystem.AppButton
import cloud.eka_dev.task_hub.core.model.RegisterDeviceTokenRequest
import cloud.eka_dev.task_hub.core.model.Reminder
import cloud.eka_dev.task_hub.core.network.ApiClient
import cloud.eka_dev.task_hub.core.network.SSEClient
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

class ReminderViewModel(
    private val apiClient: ApiClient,
    private val sseClient: SSEClient
) : ViewModel() {

    private val _reminders = MutableStateFlow<List<Reminder>>(emptyList())
    val reminders: StateFlow<List<Reminder>> = _reminders.asStateFlow()

    private val _activeAlert = MutableStateFlow<String?>(null)
    val activeAlert: StateFlow<String?> = _activeAlert.asStateFlow()

    init {
        loadReminders()
        startSSEListener()
    }

    fun loadReminders() {
        viewModelScope.launch {
            try {
                val res = apiClient.apiService.getReminders()
                if (res.success && res.data != null) {
                    _reminders.value = res.data
                }
            } catch (e: Exception) {
                // handle error
            }
        }
    }

    private fun startSSEListener() {
        viewModelScope.launch {
            try {
                sseClient.observeEvents().collect { event ->
                    if (event.type == "reminder:triggered") {
                        _activeAlert.value = "Pengingat Aktif: ${event.data}"
                        loadReminders()
                    }
                }
            } catch (e: Exception) {
                // SSE error / auto reconnect
            }
        }
    }

    fun snooze(id: Long) {
        viewModelScope.launch {
            try {
                apiClient.apiService.snoozeReminder(id, 15)
                loadReminders()
            } catch (e: Exception) {
                // handle error
            }
        }
    }

    fun updateReminder(
        id: Long,
        title: String,
        description: String?,
        remindAt: String? = null,
        status: String? = null
    ) {
        viewModelScope.launch {
            try {
                apiClient.apiService.updateReminder(
                    id,
                    cloud.eka_dev.task_hub.core.model.UpdateReminderRequest(
                        title = title,
                        description = description,
                        remindAt = remindAt,
                        status = status
                    )
                )
                loadReminders()
            } catch (e: Exception) {
                // handle error
            }
        }
    }

    fun deleteReminder(id: Long) {
        viewModelScope.launch {
            try {
                apiClient.apiService.deleteReminder(id)
                loadReminders()
            } catch (e: Exception) {
                // handle error
            }
        }
    }

    fun dismiss(id: Long) {
        viewModelScope.launch {
            try {
                apiClient.apiService.dismissReminder(id)
                loadReminders()
            } catch (e: Exception) {
                // handle error
            }
        }
    }

    fun clearActiveAlert() {
        _activeAlert.value = null
    }

    fun registerFcmToken(token: String) {
        viewModelScope.launch {
            try {
                apiClient.apiService.registerDeviceToken(
                    RegisterDeviceTokenRequest(token = token, platform = "android")
                )
            } catch (e: Exception) {
                // handle error
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ReminderScreen(viewModel: ReminderViewModel) {
    val reminders by viewModel.reminders.collectAsState()
    val activeAlert by viewModel.activeAlert.collectAsState()
    var editingReminder by remember { mutableStateOf<Reminder?>(null) }

    Scaffold(
        topBar = { TopAppBar(title = { Text("Pengingat & Notifikasi") }) }
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(16.dp)
        ) {
            // Live SSE Banner
            if (activeAlert != null) {
                Card(
                    colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.primaryContainer),
                    shape = RoundedCornerShape(12.dp),
                    modifier = Modifier.fillMaxWidth().padding(bottom = 16.dp)
                ) {
                    Row(
                        modifier = Modifier.padding(16.dp),
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.SpaceBetween
                    ) {
                        Text(activeAlert ?: "", modifier = Modifier.weight(1f), fontSize = 14.sp)
                        TextButton(onClick = { viewModel.clearActiveAlert() }) {
                            Text("Tutup")
                        }
                    }
                }
            }

            if (reminders.isEmpty()) {
                Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                    Text("Tidak ada pengingat tertunda.", color = MaterialTheme.colorScheme.onSurfaceVariant)
                }
            } else {
                LazyColumn(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                    items(reminders, key = { it.id }) { reminder ->
                        Card(
                            modifier = Modifier.fillMaxWidth(),
                            shape = RoundedCornerShape(12.dp)
                        ) {
                            Column(modifier = Modifier.padding(16.dp)) {
                                Text(reminder.title, fontWeight = FontWeight.Bold, fontSize = 16.sp)
                                if (reminder.description.isNotBlank()) {
                                    Text(reminder.description, fontSize = 13.sp, color = MaterialTheme.colorScheme.onSurfaceVariant)
                                }
                                Spacer(modifier = Modifier.height(6.dp))
                                Text("Jadwal: ${reminder.remindAt}", fontSize = 12.sp, color = MaterialTheme.colorScheme.primary)

                                Spacer(modifier = Modifier.height(12.dp))
                                Row(
                                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                                    verticalAlignment = Alignment.CenterVertically
                                ) {
                                    OutlinedButton(onClick = { editingReminder = reminder }) {
                                        Text("Edit")
                                    }
                                    OutlinedButton(onClick = { viewModel.snooze(reminder.id) }) {
                                        Text("Tunda 15 mnt")
                                    }
                                    Button(onClick = { viewModel.dismiss(reminder.id) }) {
                                        Text("Selesai")
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    }

    if (editingReminder != null) {
        val current = editingReminder!!
        var editTitle by remember { mutableStateOf(current.title) }
        var editDescription by remember { mutableStateOf(current.description) }

        AlertDialog(
            onDismissRequest = { editingReminder = null },
            title = { Text("Edit Pengingat") },
            text = {
                Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                    OutlinedTextField(
                        value = editTitle,
                        onValueChange = { editTitle = it },
                        label = { Text("Judul Pengingat") },
                        modifier = Modifier.fillMaxWidth()
                    )
                    OutlinedTextField(
                        value = editDescription,
                        onValueChange = { editDescription = it },
                        label = { Text("Deskripsi / Catatan") },
                        modifier = Modifier.fillMaxWidth(),
                        maxLines = 3
                    )
                }
            },
            confirmButton = {
                Button(
                    onClick = {
                        if (editTitle.isNotBlank()) {
                            viewModel.updateReminder(
                                id = current.id,
                                title = editTitle.trim(),
                                description = editDescription.trim()
                            )
                            editingReminder = null
                        }
                    },
                    enabled = editTitle.isNotBlank()
                ) {
                    Text("Simpan")
                }
            },
            dismissButton = {
                TextButton(onClick = { editingReminder = null }) {
                    Text("Batal")
                }
            }
        )
    }
}
