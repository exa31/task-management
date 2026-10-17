package cloud.eka_dev.task_hub.feature.note

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import cloud.eka_dev.task_hub.R
import cloud.eka_dev.task_hub.core.designsystem.*
import cloud.eka_dev.task_hub.core.model.*
import cloud.eka_dev.task_hub.core.network.ApiClient
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

data class NoteUiState(
    val notes: List<Note> = emptyList(),
    val isLoading: Boolean = false,
    val searchQuery: String = "",
    val selectedType: String? = null,
    val feedbackMessage: String? = null,
    val error: String? = null
)

class NoteViewModel(private val apiClient: ApiClient) : ViewModel() {
    private val _uiState = MutableStateFlow(NoteUiState())
    val uiState: StateFlow<NoteUiState> = _uiState.asStateFlow()

    init {
        loadNotes()
    }

    fun loadNotes() {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true, error = null)
            try {
                val res = apiClient.apiService.getNotes(
                    type = _uiState.value.selectedType,
                    search = _uiState.value.searchQuery.ifEmpty { null }
                )
                if (res.success && res.data != null) {
                    _uiState.value = _uiState.value.copy(
                        isLoading = false,
                        notes = res.data
                    )
                } else {
                    _uiState.value = _uiState.value.copy(isLoading = false, error = res.message)
                }
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(isLoading = false, error = e.localizedMessage)
            }
        }
    }

    fun setSearchQuery(query: String) {
        _uiState.value = _uiState.value.copy(searchQuery = query)
        loadNotes()
    }

    fun setSelectedType(type: String?) {
        _uiState.value = _uiState.value.copy(selectedType = type)
        loadNotes()
    }

    fun clearFeedback() {
        _uiState.value = _uiState.value.copy(feedbackMessage = null, error = null)
    }

    fun createNote(title: String, content: String, type: String = "actionable", isPinned: Boolean = false) {
        viewModelScope.launch {
            try {
                val res = apiClient.apiService.createNote(
                    CreateNoteRequest(title = title, content = content, type = type, isPinned = isPinned)
                )
                if (res.success) {
                    _uiState.value = _uiState.value.copy(feedbackMessage = "Catatan berhasil dibuat!")
                    loadNotes()
                }
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(error = e.localizedMessage)
            }
        }
    }

    fun updateNote(id: Long, title: String, content: String, type: String) {
        viewModelScope.launch {
            try {
                val res = apiClient.apiService.updateNote(
                    id,
                    UpdateNoteRequest(title = title, content = content, type = type)
                )
                if (res.success) {
                    _uiState.value = _uiState.value.copy(feedbackMessage = "Catatan berhasil diperbarui!")
                    loadNotes()
                }
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(error = e.localizedMessage)
            }
        }
    }

    fun deleteNote(id: Long) {
        viewModelScope.launch {
            try {
                val res = apiClient.apiService.deleteNote(id)
                if (res.success) {
                    _uiState.value = _uiState.value.copy(feedbackMessage = "Catatan berhasil dihapus!")
                    loadNotes()
                }
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(error = e.localizedMessage)
            }
        }
    }

    fun togglePin(note: Note) {
        viewModelScope.launch {
            try {
                val res = apiClient.apiService.updateNote(
                    note.id,
                    UpdateNoteRequest(isPinned = !note.isPinned)
                )
                if (res.success) {
                    loadNotes()
                }
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(error = e.localizedMessage)
            }
        }
    }

    fun convertNoteToTask(note: Note) {
        viewModelScope.launch {
            try {
                val res = apiClient.apiService.convertNoteToTask(note.id, ConvertToTaskRequest())
                if (res.success) {
                    _uiState.value = _uiState.value.copy(feedbackMessage = "Catatan berhasil dikonversi menjadi Tugas!")
                    loadNotes()
                }
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(error = e.localizedMessage)
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun NoteScreen(viewModel: NoteViewModel) {
    val uiState by viewModel.uiState.collectAsState()
    var showAddDialog by remember { mutableStateOf(false) }
    var noteToEdit by remember { mutableStateOf<Note?>(null) }
    var noteToDelete by remember { mutableStateOf<Note?>(null) }
    var noteToConvert by remember { mutableStateOf<Note?>(null) }

    val sortedNotes = remember(uiState.notes) {
        uiState.notes.sortedByDescending { if (it.isPinned) 1 else 0 }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(stringResource(R.string.notes_title), fontWeight = FontWeight.Bold) }
            )
        },
        floatingActionButton = {
            FloatingActionButton(
                onClick = { showAddDialog = true },
                containerColor = BrandPrimary,
                contentColor = Color.White
            ) {
                Icon(Icons.Default.Add, contentDescription = stringResource(R.string.create_note))
            }
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
        ) {
            // Search Bar & Filter Row
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 16.dp, vertical = 8.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                OutlinedTextField(
                    value = uiState.searchQuery,
                    onValueChange = { viewModel.setSearchQuery(it) },
                    placeholder = { Text(stringResource(R.string.search_notes_hint)) },
                    leadingIcon = { Icon(Icons.Default.Search, contentDescription = null, tint = MaterialTheme.colorScheme.onSurfaceVariant) },
                    trailingIcon = {
                        if (uiState.searchQuery.isNotEmpty()) {
                            IconButton(onClick = { viewModel.setSearchQuery("") }) {
                                Icon(Icons.Default.Close, contentDescription = stringResource(R.string.clear_search))
                            }
                        }
                    },
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(12.dp),
                    singleLine = true
                )

                LazyRow(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    item {
                        FilterChip(
                            selected = uiState.selectedType == null,
                            onClick = { viewModel.setSelectedType(null) },
                            label = { Text(stringResource(R.string.filter_all_count, uiState.notes.size)) }
                        )
                    }
                    item {
                        FilterChip(
                            selected = uiState.selectedType == "actionable",
                            onClick = { viewModel.setSelectedType("actionable") },
                            label = { Text(stringResource(R.string.filter_actionable_count, uiState.notes.count { it.type == "actionable" })) }
                        )
                    }
                    item {
                        FilterChip(
                            selected = uiState.selectedType == "knowledge",
                            onClick = { viewModel.setSelectedType("knowledge") },
                            label = { Text(stringResource(R.string.filter_knowledge_count, uiState.notes.count { it.type == "knowledge" })) }
                        )
                    }
                }
            }

            // Feedback Message Banner
            uiState.feedbackMessage?.let { msg ->
                Surface(
                    color = Color(0xFF10B981).copy(alpha = 0.15f),
                    modifier = Modifier.fillMaxWidth()
                ) {
                    Row(
                        modifier = Modifier.padding(horizontal = 16.dp, vertical = 8.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Icon(Icons.Default.CheckCircle, contentDescription = null, tint = Color(0xFF10B981), modifier = Modifier.size(18.dp))
                        Spacer(modifier = Modifier.width(8.dp))
                        Text(msg, color = Color(0xFF10B981), fontSize = 13.sp, modifier = Modifier.weight(1f))
                        IconButton(onClick = { viewModel.clearFeedback() }, modifier = Modifier.size(24.dp)) {
                            Icon(Icons.Default.Close, contentDescription = null, tint = Color(0xFF10B981), modifier = Modifier.size(16.dp))
                        }
                    }
                }
            }

            uiState.error?.let { err ->
                Surface(
                    color = Color.Red.copy(alpha = 0.12f),
                    modifier = Modifier.fillMaxWidth()
                ) {
                    Row(
                        modifier = Modifier.padding(horizontal = 16.dp, vertical = 8.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Icon(Icons.Default.Warning, contentDescription = null, tint = Color.Red, modifier = Modifier.size(18.dp))
                        Spacer(modifier = Modifier.width(8.dp))
                        Text(err, color = Color.Red, fontSize = 13.sp, modifier = Modifier.weight(1f))
                        IconButton(onClick = { viewModel.clearFeedback() }, modifier = Modifier.size(24.dp)) {
                            Icon(Icons.Default.Close, contentDescription = null, tint = Color.Red, modifier = Modifier.size(16.dp))
                        }
                    }
                }
            }

            // List Content
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(horizontal = 16.dp, vertical = 8.dp)
            ) {
                if (uiState.isLoading && uiState.notes.isEmpty()) {
                    TaskListSkeleton(count = 4)
                } else if (sortedNotes.isEmpty()) {
                    Column(
                        modifier = Modifier.align(Alignment.Center),
                        horizontalAlignment = Alignment.CenterHorizontally,
                        verticalArrangement = Arrangement.spacedBy(8.dp)
                    ) {
                        Icon(
                            Icons.Default.Lightbulb,
                            contentDescription = null,
                            modifier = Modifier.size(48.dp),
                            tint = MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.5f)
                        )
                        Text(
                            text = stringResource(R.string.empty_notes),
                            fontWeight = FontWeight.SemiBold,
                            color = MaterialTheme.colorScheme.onSurfaceVariant
                        )
                        Text(
                            text = stringResource(R.string.empty_notes_sub),
                            fontSize = 13.sp,
                            color = MaterialTheme.colorScheme.onSurfaceVariant
                        )
                    }
                } else {
                    LazyColumn(
                        verticalArrangement = Arrangement.spacedBy(10.dp),
                        modifier = Modifier.fillMaxSize()
                    ) {
                        items(sortedNotes, key = { it.id }) { note ->
                            Card(
                                modifier = Modifier.fillMaxWidth(),
                                shape = RoundedCornerShape(14.dp),
                                colors = CardDefaults.cardColors(
                                    containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.5f)
                                )
                            ) {
                                Column(modifier = Modifier.padding(16.dp)) {
                                    Row(
                                        modifier = Modifier.fillMaxWidth(),
                                        horizontalArrangement = Arrangement.SpaceBetween,
                                        verticalAlignment = Alignment.Top
                                    ) {
                                        Column(modifier = Modifier.weight(1f)) {
                                            Row(
                                                verticalAlignment = Alignment.CenterVertically,
                                                horizontalArrangement = Arrangement.spacedBy(6.dp)
                                            ) {
                                                if (note.isPinned) {
                                                    Icon(
                                                        Icons.Default.PushPin,
                                                        contentDescription = "Pinned",
                                                        tint = BrandPrimary,
                                                        modifier = Modifier.size(16.dp)
                                                    )
                                                }
                                                Text(
                                                    text = note.title,
                                                    fontWeight = FontWeight.Bold,
                                                    fontSize = 16.sp,
                                                    color = MaterialTheme.colorScheme.onSurface
                                                )
                                            }

                                            Spacer(modifier = Modifier.height(4.dp))
                                            Surface(
                                                color = if (note.type == "actionable") Color(0xFF6366F1).copy(alpha = 0.15f) else Color(0xFF8B5CF6).copy(alpha = 0.15f),
                                                shape = CircleShape
                                            ) {
                                                Text(
                                                    text = if (note.type == "actionable") "ACTIONABLE" else "KNOWLEDGE",
                                                    fontSize = 10.sp,
                                                    fontWeight = FontWeight.Bold,
                                                    color = if (note.type == "actionable") Color(0xFF6366F1) else Color(0xFF8B5CF6),
                                                    modifier = Modifier.padding(horizontal = 8.dp, vertical = 2.dp)
                                                )
                                            }
                                        }

                                        IconButton(
                                            onClick = { viewModel.togglePin(note) },
                                            modifier = Modifier.size(28.dp)
                                        ) {
                                            Icon(
                                                if (note.isPinned) Icons.Default.PushPin else Icons.Default.OutlinedFlag,
                                                contentDescription = "Pin Note",
                                                tint = if (note.isPinned) BrandPrimary else MaterialTheme.colorScheme.onSurfaceVariant,
                                                modifier = Modifier.size(18.dp)
                                            )
                                        }
                                    }

                                    if (note.content.isNotBlank()) {
                                        Spacer(modifier = Modifier.height(8.dp))
                                        Text(
                                            text = note.content,
                                            fontSize = 14.sp,
                                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                                            maxLines = 4
                                        )
                                    }

                                    Spacer(modifier = Modifier.height(12.dp))

                                    // Action Buttons Row
                                    Row(
                                        modifier = Modifier.fillMaxWidth(),
                                        horizontalArrangement = Arrangement.SpaceBetween,
                                        verticalAlignment = Alignment.CenterVertically
                                    ) {
                                        if (note.type == "actionable") {
                                            FilledTonalButton(
                                                onClick = { noteToConvert = note },
                                                contentPadding = PaddingValues(horizontal = 10.dp, vertical = 4.dp),
                                                modifier = Modifier.height(32.dp)
                                            ) {
                                                Icon(Icons.Default.ArrowForward, contentDescription = null, modifier = Modifier.size(14.dp))
                                                Spacer(modifier = Modifier.width(4.dp))
                                                Text(stringResource(R.string.convert_to_task), fontSize = 11.sp, fontWeight = FontWeight.SemiBold)
                                            }
                                        } else {
                                            Spacer(modifier = Modifier.width(1.dp))
                                        }

                                        Row(horizontalArrangement = Arrangement.spacedBy(4.dp)) {
                                            IconButton(
                                                onClick = { noteToEdit = note },
                                                modifier = Modifier.size(32.dp)
                                            ) {
                                                Icon(Icons.Default.Edit, contentDescription = stringResource(R.string.edit), tint = MaterialTheme.colorScheme.onSurfaceVariant, modifier = Modifier.size(18.dp))
                                            }

                                            IconButton(
                                                onClick = { noteToDelete = note },
                                                modifier = Modifier.size(32.dp)
                                            ) {
                                                Icon(Icons.Default.Delete, contentDescription = stringResource(R.string.delete), tint = Color.Red.copy(alpha = 0.8f), modifier = Modifier.size(18.dp))
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    }

    // Dialog Create Note
    if (showAddDialog) {
        var title by remember { mutableStateOf("") }
        var content by remember { mutableStateOf("") }
        var type by remember { mutableStateOf("actionable") }
        var isPinned by remember { mutableStateOf(false) }

        AlertDialog(
            onDismissRequest = { showAddDialog = false },
            title = { Text(stringResource(R.string.create_note_title), fontWeight = FontWeight.Bold) },
            text = {
                Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                    OutlinedTextField(
                        value = title,
                        onValueChange = { title = it },
                        label = { Text(stringResource(R.string.note_title_label)) },
                        modifier = Modifier.fillMaxWidth(),
                        singleLine = true
                    )

                    Text(stringResource(R.string.note_type_label), fontWeight = FontWeight.SemiBold, fontSize = 12.sp)
                    Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                        FilterChip(
                            selected = type == "actionable",
                            onClick = { type = "actionable" },
                            label = { Text(stringResource(R.string.note_type_actionable)) }
                        )
                        FilterChip(
                            selected = type == "knowledge",
                            onClick = { type = "knowledge" },
                            label = { Text(stringResource(R.string.note_type_knowledge)) }
                        )
                    }

                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.SpaceBetween
                    ) {
                        Text(stringResource(R.string.pin_to_top), fontSize = 13.sp)
                        Switch(checked = isPinned, onCheckedChange = { isPinned = it })
                    }

                    OutlinedTextField(
                        value = content,
                        onValueChange = { content = it },
                        label = { Text(stringResource(R.string.note_content_label)) },
                        modifier = Modifier.fillMaxWidth(),
                        minLines = 4
                    )
                }
            },
            confirmButton = {
                AppButton(
                    text = stringResource(R.string.save),
                    enabled = title.isNotBlank(),
                    onClick = {
                        viewModel.createNote(title.trim(), content.trim(), type, isPinned)
                        showAddDialog = false
                    }
                )
            },
            dismissButton = {
                TextButton(onClick = { showAddDialog = false }) { Text(stringResource(R.string.cancel)) }
            }
        )
    }

    // Dialog Edit Note
    noteToEdit?.let { note ->
        var editTitle by remember { mutableStateOf(note.title) }
        var editContent by remember { mutableStateOf(note.content) }
        var editType by remember { mutableStateOf(note.type) }

        AlertDialog(
            onDismissRequest = { noteToEdit = null },
            title = { Text(stringResource(R.string.edit_note_title), fontWeight = FontWeight.Bold) },
            text = {
                Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                    OutlinedTextField(
                        value = editTitle,
                        onValueChange = { editTitle = it },
                        label = { Text(stringResource(R.string.note_title_label)) },
                        modifier = Modifier.fillMaxWidth(),
                        singleLine = true
                    )

                    Text(stringResource(R.string.note_type_label), fontWeight = FontWeight.SemiBold, fontSize = 12.sp)
                    Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                        FilterChip(
                            selected = editType == "actionable",
                            onClick = { editType = "actionable" },
                            label = { Text(stringResource(R.string.note_type_actionable)) }
                        )
                        FilterChip(
                            selected = editType == "knowledge",
                            onClick = { editType = "knowledge" },
                            label = { Text(stringResource(R.string.note_type_knowledge)) }
                        )
                    }

                    OutlinedTextField(
                        value = editContent,
                        onValueChange = { editContent = it },
                        label = { Text(stringResource(R.string.note_content_label)) },
                        modifier = Modifier.fillMaxWidth(),
                        minLines = 4
                    )
                }
            },
            confirmButton = {
                AppButton(
                    text = stringResource(R.string.save),
                    enabled = editTitle.isNotBlank(),
                    onClick = {
                        viewModel.updateNote(note.id, editTitle.trim(), editContent.trim(), editType)
                        noteToEdit = null
                    }
                )
            },
            dismissButton = {
                TextButton(onClick = { noteToEdit = null }) { Text(stringResource(R.string.cancel)) }
            }
        )
    }

    // Confirm Delete Dialog
    noteToDelete?.let { note ->
        AlertDialog(
            onDismissRequest = { noteToDelete = null },
            title = { Text(stringResource(R.string.delete_note_confirm_title), fontWeight = FontWeight.Bold) },
            text = { Text(stringResource(R.string.delete_note_confirm_desc, note.title)) },
            confirmButton = {
                Button(
                    onClick = {
                        viewModel.deleteNote(note.id)
                        noteToDelete = null
                    },
                    colors = ButtonDefaults.buttonColors(containerColor = Color.Red)
                ) {
                    Text(stringResource(R.string.delete), color = Color.White)
                }
            },
            dismissButton = {
                TextButton(onClick = { noteToDelete = null }) { Text(stringResource(R.string.cancel)) }
            }
        )
    }

    // Confirm Convert Note to Task Dialog
    noteToConvert?.let { note ->
        AlertDialog(
            onDismissRequest = { noteToConvert = null },
            title = { Text(stringResource(R.string.convert_note_dialog_title), fontWeight = FontWeight.Bold) },
            text = { Text(stringResource(R.string.convert_note_dialog_desc, note.title)) },
            confirmButton = {
                AppButton(
                    text = stringResource(R.string.convert_btn),
                    onClick = {
                        viewModel.convertNoteToTask(note)
                        noteToConvert = null
                    }
                )
            },
            dismissButton = {
                TextButton(onClick = { noteToConvert = null }) { Text(stringResource(R.string.cancel)) }
            }
        )
    }
}
