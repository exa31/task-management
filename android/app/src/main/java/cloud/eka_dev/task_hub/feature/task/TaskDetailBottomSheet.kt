package cloud.eka_dev.task_hub.feature.task

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextDecoration
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import cloud.eka_dev.task_hub.core.designsystem.*
import cloud.eka_dev.task_hub.core.model.Task
import cloud.eka_dev.task_hub.core.model.TaskActivity

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TaskDetailBottomSheet(
    task: Task,
    viewModel: TaskViewModel,
    onDismiss: () -> Unit
) {
    val uiState by viewModel.uiState.collectAsState()
    val currentTask = if (uiState.selectedTask?.id == task.id) uiState.selectedTask!! else task
    var selectedTab by remember { mutableIntStateOf(0) }
    var newNoteText by remember { mutableStateOf("") }
    var showEditDialog by remember { mutableStateOf(false) }

    ModalBottomSheet(
        onDismissRequest = onDismiss,
        sheetState = rememberModalBottomSheetState(skipPartiallyExpanded = true),
        containerColor = MaterialTheme.colorScheme.surface
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(bottom = 32.dp)
        ) {
            // Task Header
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 20.dp, vertical = 8.dp)
            ) {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.Top
                ) {
                    Column(modifier = Modifier.weight(1f)) {
                        Text(
                            text = currentTask.title,
                            fontSize = 18.sp,
                            fontWeight = FontWeight.Bold,
                            color = MaterialTheme.colorScheme.onSurface
                        )
                        Spacer(modifier = Modifier.height(6.dp))
                        Row(
                            horizontalArrangement = Arrangement.spacedBy(8.dp),
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            if (!currentTask.projectName.isNullOrEmpty()) {
                                ProjectPill(name = currentTask.projectName, hexColor = currentTask.projectColor)
                            }
                            PriorityBadge(priority = currentTask.priority)
                            StatusBadge(status = currentTask.status)
                        }
                    }

                    Row(verticalAlignment = Alignment.CenterVertically) {
                        IconButton(onClick = { viewModel.toggleTaskFocus(currentTask) }) {
                            Icon(
                                imageVector = if (currentTask.isFocus) Icons.Default.Star else Icons.Default.StarBorder,
                                contentDescription = "Fokus",
                                tint = if (currentTask.isFocus) Color(0xFFF59E0B) else MaterialTheme.colorScheme.onSurfaceVariant
                            )
                        }
                        IconButton(onClick = { showEditDialog = true }) {
                            Icon(
                                imageVector = Icons.Default.Edit,
                                contentDescription = "Edit Tugas",
                                tint = BrandPrimary
                            )
                        }
                    }
                }

                if (!currentTask.description.isNullOrEmpty()) {
                    Spacer(modifier = Modifier.height(8.dp))
                    Text(
                        text = currentTask.description,
                        fontSize = 13.sp,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                        maxLines = 3
                    )
                }
            }

            Spacer(modifier = Modifier.height(8.dp))

            // Tab Selector
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 20.dp),
                horizontalArrangement = Arrangement.spacedBy(10.dp)
            ) {
                FilterChip(
                    selected = selectedTab == 0,
                    onClick = { selectedTab = 0 },
                    leadingIcon = { Icon(Icons.Default.Checklist, contentDescription = null, modifier = Modifier.size(16.dp)) },
                    label = { Text("Detail & Subtask") },
                    modifier = Modifier.weight(1f)
                )

                FilterChip(
                    selected = selectedTab == 1,
                    onClick = { selectedTab = 1 },
                    leadingIcon = { Icon(Icons.Default.History, contentDescription = null, modifier = Modifier.size(16.dp)) },
                    label = {
                        Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(4.dp)) {
                            Text("Riwayat Log")
                            if (uiState.taskActivities.isNotEmpty()) {
                                Surface(
                                    shape = CircleShape,
                                    color = BrandPrimary.copy(alpha = 0.2f),
                                    modifier = Modifier.size(18.dp)
                                ) {
                                    Box(contentAlignment = Alignment.Center) {
                                        Text(
                                            "${uiState.taskActivities.size}",
                                            fontSize = 10.sp,
                                            fontWeight = FontWeight.Bold,
                                            color = BrandPrimary
                                        )
                                    }
                                }
                            }
                        }
                    },
                    modifier = Modifier.weight(1f)
                )
            }

            Spacer(modifier = Modifier.height(12.dp))

            // Tab Content
            if (selectedTab == 0) {
                DetailTabContent(task = currentTask, viewModel = viewModel, onTaskDeleted = onDismiss)
            } else {
                ActivityLogsTabContent(
                    task = currentTask,
                    activities = uiState.taskActivities,
                    isLoading = uiState.isLoadingActivities,
                    newNoteText = newNoteText,
                    onNoteChange = { newNoteText = it },
                    onPostNote = {
                        viewModel.addActivityNote(currentTask.id, newNoteText)
                        newNoteText = ""
                    }
                )
            }
        }
    }

    if (showEditDialog) {
        EditTaskDialog(
            task = currentTask,
            onDismiss = { showEditDialog = false },
            onSave = { updatedTitle, updatedDesc, updatedStatus, updatedPriority, updatedIsFocus ->
                viewModel.updateTask(
                    id = currentTask.id,
                    title = updatedTitle,
                    description = updatedDesc,
                    status = updatedStatus,
                    priority = updatedPriority,
                    isFocus = updatedIsFocus,
                    onSuccess = {
                        showEditDialog = false
                    }
                )
            }
        )
    }
}

@Composable
private fun DetailTabContent(
    task: Task,
    viewModel: TaskViewModel,
    onTaskDeleted: () -> Unit
) {
    var newSubtaskText by remember { mutableStateOf("") }
    var showDeleteDialog by remember { mutableStateOf(false) }

    Column(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 20.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        // Status Stepper / Selector
        Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
            Text(
                "STATUS TUGAS",
                fontSize = 11.sp,
                fontWeight = FontWeight.Bold,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .horizontalScroll(rememberScrollState()),
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                listOf(
                    Triple("todo", "To Do", StatusTodo),
                    Triple("in_progress", "In Progress", StatusInProgress),
                    Triple("review", "Review", StatusReview),
                    Triple("done", "Selesai", StatusDone)
                ).forEach { (st, label, color) ->
                    val isSelected = task.status.lowercase() == st
                    FilterChip(
                        selected = isSelected,
                        onClick = { viewModel.updateTaskStatus(task, st) },
                        leadingIcon = {
                            Icon(
                                when (st) {
                                    "done" -> Icons.Default.CheckCircle
                                    "in_progress" -> Icons.Default.Sync
                                    "review" -> Icons.Default.Visibility
                                    else -> Icons.Default.RadioButtonUnchecked
                                },
                                contentDescription = null,
                                tint = if (isSelected) color else MaterialTheme.colorScheme.onSurfaceVariant,
                                modifier = Modifier.size(15.dp)
                            )
                        },
                        label = {
                            Text(
                                label,
                                fontWeight = if (isSelected) FontWeight.Bold else FontWeight.Medium,
                                color = if (isSelected) color else MaterialTheme.colorScheme.onSurface
                            )
                        }
                    )
                }
            }
        }

        // Subtask Progress Header
        Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Text(
                    "SUBTASKS (${task.subtasksCompleted}/${task.subtasksCount})",
                    fontSize = 12.sp,
                    fontWeight = FontWeight.Bold,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
                Text(
                    "${task.progressPercent}%",
                    fontSize = 12.sp,
                    fontWeight = FontWeight.Bold,
                    color = BrandPrimary
                )
            }

            LinearProgressIndicator(
                progress = { task.progressPercent / 100f },
                modifier = Modifier
                    .fillMaxWidth()
                    .height(6.dp)
                    .clip(RoundedCornerShape(3.dp)),
                color = BrandPrimary,
                trackColor = MaterialTheme.colorScheme.surfaceVariant
            )
        }

        // Add Subtask Form
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .clip(RoundedCornerShape(10.dp))
                .background(MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.35f))
                .padding(6.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            OutlinedTextField(
                value = newSubtaskText,
                onValueChange = { newSubtaskText = it },
                placeholder = { Text("Tambah checklist subtask...", fontSize = 12.sp) },
                modifier = Modifier.weight(1f),
                singleLine = true,
                colors = OutlinedTextFieldDefaults.colors(
                    unfocusedContainerColor = MaterialTheme.colorScheme.surface,
                    focusedContainerColor = MaterialTheme.colorScheme.surface
                ),
                shape = RoundedCornerShape(8.dp)
            )
            Spacer(modifier = Modifier.width(6.dp))
            IconButton(
                onClick = {
                    if (newSubtaskText.isNotBlank()) {
                        viewModel.createSubtask(task.id, newSubtaskText.trim())
                        newSubtaskText = ""
                    }
                },
                enabled = newSubtaskText.isNotBlank()
            ) {
                Icon(
                    Icons.Default.AddCircle,
                    contentDescription = "Tambah Subtask",
                    tint = if (newSubtaskText.isNotBlank()) BrandPrimary else MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.4f)
                )
            }
        }

        // Subtask Checklist Items
        if (task.subtasks.isNotEmpty()) {
            Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                task.subtasks.forEach { subtask ->
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .clip(RoundedCornerShape(10.dp))
                            .background(MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.5f))
                            .padding(10.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Checkbox(
                            checked = subtask.isCompleted,
                            onCheckedChange = {
                                viewModel.toggleSubtask(task.id, subtask.id, subtask.isCompleted)
                            },
                            colors = CheckboxDefaults.colors(checkedColor = StatusDone)
                        )
                        Spacer(modifier = Modifier.width(8.dp))
                        Column(modifier = Modifier.weight(1f)) {
                            Text(
                                text = subtask.title,
                                fontSize = 13.sp,
                                textDecoration = if (subtask.isCompleted) TextDecoration.LineThrough else TextDecoration.None,
                                color = if (subtask.isCompleted) MaterialTheme.colorScheme.onSurfaceVariant else MaterialTheme.colorScheme.onSurface
                            )
                            if (subtask.description.isNotBlank()) {
                                Spacer(modifier = Modifier.height(2.dp))
                                Text(
                                    text = subtask.description,
                                    fontSize = 11.sp,
                                    color = MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.8f),
                                    maxLines = 2
                                )
                            }
                            if (!subtask.dueDate.isNullOrBlank()) {
                                Spacer(modifier = Modifier.height(2.dp))
                                Text(
                                    text = subtask.dueDate,
                                    fontSize = 10.sp,
                                    color = BrandPrimary
                                )
                            }
                        }
                        IconButton(
                            onClick = { viewModel.deleteSubtask(task.id, subtask.id) },
                            modifier = Modifier.size(24.dp)
                        ) {
                            Icon(
                                Icons.Default.DeleteOutline,
                                contentDescription = "Hapus Subtask",
                                tint = MaterialTheme.colorScheme.error.copy(alpha = 0.7f),
                                modifier = Modifier.size(16.dp)
                            )
                        }
                    }
                }
            }
        } else {
            Text(
                "Tidak ada subtask checklist.",
                fontSize = 13.sp,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }

        Spacer(modifier = Modifier.height(8.dp))

        // Delete Task Button
        OutlinedButton(
            onClick = { showDeleteDialog = true },
            modifier = Modifier.fillMaxWidth(),
            shape = RoundedCornerShape(12.dp),
            colors = ButtonDefaults.outlinedButtonColors(contentColor = MaterialTheme.colorScheme.error),
            border = BorderStroke(1.dp, MaterialTheme.colorScheme.error.copy(alpha = 0.5f))
        ) {
            Icon(Icons.Default.Delete, contentDescription = null, modifier = Modifier.size(16.dp))
            Spacer(modifier = Modifier.width(6.dp))
            Text("Hapus Tugas Ini", fontWeight = FontWeight.SemiBold)
        }
    }

    if (showDeleteDialog) {
        AlertDialog(
            onDismissRequest = { showDeleteDialog = false },
            title = { Text("Hapus Tugas?", fontWeight = FontWeight.Bold) },
            text = { Text("Apakah Anda yakin ingin menghapus tugas '${task.title}'? Tindakan ini tidak dapat dibatalkan.") },
            confirmButton = {
                Button(
                    onClick = {
                        viewModel.deleteTask(task.id) {
                            showDeleteDialog = false
                            onTaskDeleted()
                        }
                    },
                    colors = ButtonDefaults.buttonColors(containerColor = MaterialTheme.colorScheme.error)
                ) {
                    Text("Hapus", color = Color.White)
                }
            },
            dismissButton = {
                TextButton(onClick = { showDeleteDialog = false }) {
                    Text("Batal")
                }
            }
        )
    }
}

@Composable
private fun ActivityLogsTabContent(
    task: Task,
    activities: List<TaskActivity>,
    isLoading: Boolean,
    newNoteText: String,
    onNoteChange: (String) -> Unit,
    onPostNote: () -> Unit
) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 20.dp)
    ) {
        // Post Note Form
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .clip(RoundedCornerShape(14.dp))
                .background(BrandPrimary.copy(alpha = 0.08f))
                .padding(10.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            OutlinedTextField(
                value = newNoteText,
                onValueChange = onNoteChange,
                placeholder = { Text("Tulis catatan progres...", fontSize = 12.sp) },
                shape = RoundedCornerShape(10.dp),
                modifier = Modifier.weight(1f),
                singleLine = true,
                colors = OutlinedTextFieldDefaults.colors(
                    unfocusedContainerColor = MaterialTheme.colorScheme.surface,
                    focusedContainerColor = MaterialTheme.colorScheme.surface
                )
            )
            Spacer(modifier = Modifier.width(8.dp))
            Button(
                onClick = onPostNote,
                enabled = newNoteText.isNotBlank(),
                shape = RoundedCornerShape(10.dp),
                colors = ButtonDefaults.buttonColors(containerColor = BrandPrimary)
            ) {
                Icon(Icons.Default.Send, contentDescription = "Kirim", modifier = Modifier.size(14.dp))
            }
        }

        Spacer(modifier = Modifier.height(14.dp))

        // Activity Timeline
        if (isLoading && activities.isEmpty()) {
            TaskListSkeleton(count = 2)
        } else if (activities.isEmpty()) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(32.dp),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                Icon(Icons.Default.History, contentDescription = null, tint = MaterialTheme.colorScheme.onSurfaceVariant, modifier = Modifier.size(36.dp))
                Text("Belum Ada Riwayat Log", fontWeight = FontWeight.SemiBold, fontSize = 15.sp)
                Text(
                    "Catatan progres dan perubahan tugas akan tercatat secara kronologis di sini.",
                    fontSize = 12.sp,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                    textAlign = androidx.compose.ui.text.style.TextAlign.Center
                )
            }
        } else {
            LazyColumn(
                verticalArrangement = Arrangement.spacedBy(10.dp),
                modifier = Modifier.fillMaxWidth().heightIn(max = 350.dp)
            ) {
                itemsIndexed(activities, key = { _, act -> act.id }) { index, act ->
                    val badge = getActivityBadge(act.activityType)
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .clip(RoundedCornerShape(12.dp))
                            .background(MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.45f))
                            .padding(12.dp),
                        verticalAlignment = Alignment.Top
                    ) {
                        Surface(
                            shape = CircleShape,
                            color = badge.color.copy(alpha = 0.15f),
                            modifier = Modifier.size(32.dp)
                        ) {
                            Box(contentAlignment = Alignment.Center) {
                                Icon(badge.icon, contentDescription = null, tint = badge.color, modifier = Modifier.size(16.dp))
                            }
                        }
                        Spacer(modifier = Modifier.width(12.dp))
                        Column(modifier = Modifier.weight(1f)) {
                            Row(
                                modifier = Modifier.fillMaxWidth(),
                                horizontalArrangement = Arrangement.SpaceBetween
                            ) {
                                Text(
                                    text = badge.label,
                                    fontSize = 12.sp,
                                    fontWeight = FontWeight.Bold,
                                    color = badge.color
                                )
                                Text(
                                    text = act.createdAt.take(16).replace("T", " "),
                                    fontSize = 11.sp,
                                    color = MaterialTheme.colorScheme.onSurfaceVariant
                                )
                            }
                            Spacer(modifier = Modifier.height(2.dp))
                            Text(
                                text = act.description,
                                fontSize = 13.sp,
                                color = MaterialTheme.colorScheme.onSurface
                            )
                        }
                    }
                }
            }
        }
    }
}

private data class ActivityBadge(val icon: androidx.compose.ui.graphics.vector.ImageVector, val color: Color, val label: String)

private fun getActivityBadge(type: String): ActivityBadge {
    return when (type) {
        "task_created" -> ActivityBadge(Icons.Default.AutoAwesome, Color(0xFF10B981), "Tugas Dibuat")
        "task_updated" -> ActivityBadge(Icons.Default.Edit, Color(0xFF3B82F6), "Tugas Diperbarui")
        "progress_logged" -> ActivityBadge(Icons.Default.History, Color(0xFF6366F1), "Catatan Progres")
        "subtask_created" -> ActivityBadge(Icons.Default.Add, Color(0xFF06B6D4), "Subtask Ditambahkan")
        "subtask_toggled" -> ActivityBadge(Icons.Default.CheckCircle, Color(0xFF14B8A6), "Subtask Diperbarui")
        "subtask_deleted" -> ActivityBadge(Icons.Default.Delete, Color(0xFFF43F5E), "Subtask Dihapus")
        "focus_toggled" -> ActivityBadge(Icons.Default.Star, Color(0xFFF59E0B), "Fokus Diubah")
        else -> ActivityBadge(Icons.Default.Comment, Color(0xFF64748B), "Catatan Log")
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun EditTaskDialog(
    task: Task,
    onDismiss: () -> Unit,
    onSave: (String, String, String, String, Boolean) -> Unit
) {
    var title by remember { mutableStateOf(task.title) }
    var description by remember { mutableStateOf(task.description ?: "") }
    var status by remember { mutableStateOf(task.status) }
    var priority by remember { mutableStateOf(task.priority) }
    var isFocus by remember { mutableStateOf(task.isFocus) }

    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("Edit Tugas", fontWeight = FontWeight.Bold) },
        text = {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .verticalScroll(rememberScrollState()),
                verticalArrangement = Arrangement.spacedBy(12.dp)
            ) {
                OutlinedTextField(
                    value = title,
                    onValueChange = { title = it },
                    label = { Text("Judul Tugas *") },
                    modifier = Modifier.fillMaxWidth(),
                    singleLine = true
                )

                OutlinedTextField(
                    value = description,
                    onValueChange = { description = it },
                    label = { Text("Deskripsi") },
                    modifier = Modifier.fillMaxWidth(),
                    minLines = 3,
                    maxLines = 5
                )

                Text("Status", fontWeight = FontWeight.SemiBold, fontSize = 13.sp)
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(6.dp)
                ) {
                    listOf("todo" to "To Do", "in_progress" to "In Progress", "review" to "Review", "done" to "Done").forEach { (st, label) ->
                        FilterChip(
                            selected = status == st,
                            onClick = { status = st },
                            label = { Text(label, fontSize = 11.sp) }
                        )
                    }
                }

                Text("Prioritas", fontWeight = FontWeight.SemiBold, fontSize = 13.sp)
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(6.dp)
                ) {
                    listOf("low" to "Low", "medium" to "Med", "high" to "High", "urgent" to "Urg").forEach { (pr, label) ->
                        FilterChip(
                            selected = priority == pr,
                            onClick = { priority = pr },
                            label = { Text(label, fontSize = 11.sp) }
                        )
                    }
                }

                Row(
                    modifier = Modifier.fillMaxWidth(),
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.SpaceBetween
                ) {
                    Text("Fokus Hari Ini", fontSize = 13.sp)
                    Switch(
                        checked = isFocus,
                        onCheckedChange = { isFocus = it }
                    )
                }
            }
        },
        confirmButton = {
            AppButton(
                text = "Simpan",
                onClick = {
                    if (title.isNotBlank()) {
                        onSave(title.trim(), description.trim(), status, priority, isFocus)
                    }
                },
                enabled = title.isNotBlank()
            )
        },
        dismissButton = {
            TextButton(onClick = onDismiss) {
                Text("Batal")
            }
        }
    )
}

