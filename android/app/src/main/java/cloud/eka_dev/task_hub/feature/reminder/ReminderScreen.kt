package cloud.eka_dev.task_hub.feature.reminder

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import cloud.eka_dev.task_hub.R
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
        listenToAlerts()
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

    private fun listenToAlerts() {
        viewModelScope.launch {
            sseClient.events.collect { event ->
                if (event.data.contains("reminder") || event.type?.contains("reminder") == true) {
                    _activeAlert.value = "Pengingat Baru Diterima: ${event.data}"
                    loadReminders()
                }
            }
        }
    }

    fun snooze(id: Long) {
        viewModelScope.launch {
            try {
                apiClient.apiService.snoozeReminder(id)
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

    fun registerPushToken(fcmToken: String) {
        viewModelScope.launch {
            try {
                apiClient.apiService.registerDeviceToken(
                    RegisterDeviceTokenRequest(token = fcmToken, platform = "android")
                )
            } catch (e: Exception) {
                // handle error
            }
        }
    }

    fun clearActiveAlert() {
        _activeAlert.value = null
    }

    fun updateReminder(id: Long, title: String, description: String) {
        viewModelScope.launch {
            try {
                apiClient.apiService.updateReminder(
                    id = id,
                    req = cloud.eka_dev.task_hub.core.model.UpdateReminderRequest(
                        title = title,
                        description = description
                    )
                )
                loadReminders()
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
        topBar = { TopAppBar(title = { Text(stringResource(R.string.reminders_title)) }) }
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
                            Text(stringResource(R.string.close))
                        }
                    }
                }
            }

            if (reminders.isEmpty()) {
                Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                    Text(stringResource(R.string.empty_reminders), color = MaterialTheme.colorScheme.onSurfaceVariant)
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
                                Text(stringResource(R.string.reminder_schedule_prefix, reminder.remindAt), fontSize = 12.sp, color = MaterialTheme.colorScheme.primary)

                                Spacer(modifier = Modifier.height(12.dp))
                                Row(
                                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                                    verticalAlignment = Alignment.CenterVertically
                                ) {
                                    OutlinedButton(onClick = { editingReminder = reminder }) {
                                        Text(stringResource(R.string.edit))
                                    }
                                    OutlinedButton(onClick = { viewModel.snooze(reminder.id) }) {
                                        Text(stringResource(R.string.snooze_15m))
                                    }
                                    Button(onClick = { viewModel.dismiss(reminder.id) }) {
                                        Text(stringResource(R.string.dismiss))
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
            title = { Text(stringResource(R.string.edit_reminder_title)) },
            text = {
                Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                    OutlinedTextField(
                        value = editTitle,
                        onValueChange = { editTitle = it },
                        label = { Text(stringResource(R.string.reminder_title_label)) },
                        modifier = Modifier.fillMaxWidth()
                    )
                    OutlinedTextField(
                        value = editDescription,
                        onValueChange = { editDescription = it },
                        label = { Text(stringResource(R.string.reminder_desc_label)) },
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
                    Text(stringResource(R.string.save))
                }
            },
            dismissButton = {
                TextButton(onClick = { editingReminder = null }) {
                    Text(stringResource(R.string.cancel))
                }
            }
        )
    }
}
