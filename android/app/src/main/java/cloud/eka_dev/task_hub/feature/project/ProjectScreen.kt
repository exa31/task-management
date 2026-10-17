package cloud.eka_dev.task_hub.feature.project

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
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
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import cloud.eka_dev.task_hub.core.designsystem.*
import cloud.eka_dev.task_hub.core.model.*
import cloud.eka_dev.task_hub.core.network.ApiClient
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

val PROJECT_PRESET_COLORS = listOf(
    "#6366F1", // Indigo
    "#8B5CF6", // Violet
    "#EC4899", // Pink
    "#F43F5E", // Rose
    "#F59E0B", // Amber
    "#10B981", // Emerald
    "#06B6D4", // Cyan
    "#3B82F6"  // Blue
)

class ProjectViewModel(private val apiClient: ApiClient) : ViewModel() {
    private val _projects = MutableStateFlow<List<Project>>(emptyList())
    val projects: StateFlow<List<Project>> = _projects.asStateFlow()

    private val _tasks = MutableStateFlow<List<Task>>(emptyList())
    val tasks: StateFlow<List<Task>> = _tasks.asStateFlow()

    private val _expandedProjectIds = MutableStateFlow<Set<Long>>(emptySet())
    val expandedProjectIds: StateFlow<Set<Long>> = _expandedProjectIds.asStateFlow()

    private val _searchQuery = MutableStateFlow("")
    val searchQuery: StateFlow<String> = _searchQuery.asStateFlow()

    private val _isLoading = MutableStateFlow(false)
    val isLoading: StateFlow<Boolean> = _isLoading.asStateFlow()

    init {
        loadData()
    }

    fun loadData() {
        viewModelScope.launch {
            _isLoading.value = true
            try {
                val projRes = apiClient.apiService.getProjects()
                if (projRes.success && projRes.data != null) {
                    _projects.value = projRes.data
                }
                val taskRes = apiClient.apiService.getTasks()
                if (taskRes.success && taskRes.data != null) {
                    _tasks.value = taskRes.data
                }
            } catch (e: Exception) {
                // handle error
            } finally {
                _isLoading.value = false
            }
        }
    }

    fun setSearchQuery(query: String) {
        _searchQuery.value = query
    }

    fun toggleExpand(projectId: Long) {
        val current = _expandedProjectIds.value
        if (current.contains(projectId)) {
            _expandedProjectIds.value = current - projectId
        } else {
            _expandedProjectIds.value = current + projectId
        }
    }

    fun createProject(name: String, color: String, description: String) {
        viewModelScope.launch {
            try {
                apiClient.apiService.createProject(CreateProjectRequest(name, color, description))
                loadData()
            } catch (e: Exception) {
                // handle error
            }
        }
    }

    fun updateProject(id: Long, name: String, color: String, description: String) {
        viewModelScope.launch {
            try {
                apiClient.apiService.updateProject(
                    id,
                    UpdateProjectRequest(name = name, color = color, description = description)
                )
                loadData()
            } catch (e: Exception) {
                // handle error
            }
        }
    }

    fun deleteProject(id: Long) {
        viewModelScope.launch {
            try {
                apiClient.apiService.deleteProject(id)
                loadData()
            } catch (e: Exception) {
                // handle error
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ProjectScreen(viewModel: ProjectViewModel) {
    val projects by viewModel.projects.collectAsState()
    val tasks by viewModel.tasks.collectAsState()
    val expandedIds by viewModel.expandedProjectIds.collectAsState()
    val searchQuery by viewModel.searchQuery.collectAsState()
    val isLoading by viewModel.isLoading.collectAsState()

    var showAddDialog by remember { mutableStateOf(false) }
    var editingProject by remember { mutableStateOf<Project?>(null) }
    var deletingProject by remember { mutableStateOf<Project?>(null) }

    val filteredProjects = remember(projects, searchQuery) {
        if (searchQuery.isBlank()) projects else {
            projects.filter {
                it.name.contains(searchQuery, ignoreCase = true) ||
                it.description.contains(searchQuery, ignoreCase = true)
            }
        }
    }

    // Metrics matching iOS
    val projectTasks = remember(tasks) { tasks.filter { (it.projectId ?: 0) > 0 } }
    val activeCount = remember(projectTasks) { projectTasks.count { it.status != "done" } }
    val completedCount = remember(projectTasks) { projectTasks.count { it.status == "done" } }
    val completionRate = remember(projectTasks, completedCount) {
        if (projectTasks.isNotEmpty()) (completedCount * 100) / projectTasks.size else 0
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    Row(
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.spacedBy(8.dp)
                    ) {
                        BrandLogo(size = 28.dp)
                        Text("Proyek & Modul", fontWeight = FontWeight.Bold)
                    }
                }
            )
        },
        floatingActionButton = {
            FloatingActionButton(
                onClick = { showAddDialog = true },
                containerColor = BrandPrimary,
                contentColor = Color.White,
                shape = CircleShape
            ) {
                Icon(Icons.Default.Add, contentDescription = "Tambah Proyek")
            }
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(horizontal = 16.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            // Search Bar
            OutlinedTextField(
                value = searchQuery,
                onValueChange = { viewModel.setSearchQuery(it) },
                placeholder = { Text("Cari proyek...", fontSize = 14.sp) },
                leadingIcon = { Icon(Icons.Default.Search, contentDescription = null, tint = BrandPrimary) },
                shape = RoundedCornerShape(14.dp),
                modifier = Modifier.fillMaxWidth(),
                singleLine = true,
                colors = OutlinedTextFieldDefaults.colors(
                    unfocusedContainerColor = MaterialTheme.colorScheme.surface,
                    focusedContainerColor = MaterialTheme.colorScheme.surface
                )
            )

            // Metrics Summary Card
            Card(
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(16.dp),
                colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.5f))
            ) {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(14.dp),
                    horizontalArrangement = Arrangement.SpaceAround
                ) {
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Text("Aktif", fontSize = 11.sp, color = MaterialTheme.colorScheme.onSurfaceVariant)
                        Text("$activeCount", fontSize = 18.sp, fontWeight = FontWeight.Bold, color = BrandPrimary)
                    }
                    Divider(modifier = Modifier.height(32.dp).width(1.dp))
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Text("Selesai", fontSize = 11.sp, color = MaterialTheme.colorScheme.onSurfaceVariant)
                        Text("$completedCount", fontSize = 18.sp, fontWeight = FontWeight.Bold, color = StatusDone)
                    }
                    Divider(modifier = Modifier.height(32.dp).width(1.dp))
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Text("Progres", fontSize = 11.sp, color = MaterialTheme.colorScheme.onSurfaceVariant)
                        Text("$completionRate%", fontSize = 18.sp, fontWeight = FontWeight.Bold, color = Color(0xFFF59E0B))
                    }
                }
            }

            // Projects List
            if (isLoading && projects.isEmpty()) {
                TaskListSkeleton(count = 4)
            } else if (filteredProjects.isEmpty()) {
                Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                    Column(
                        horizontalAlignment = Alignment.CenterHorizontally,
                        verticalArrangement = Arrangement.spacedBy(8.dp)
                    ) {
                        Icon(Icons.Default.FolderOpen, contentDescription = null, modifier = Modifier.size(48.dp), tint = MaterialTheme.colorScheme.onSurfaceVariant)
                        Text("Belum ada proyek.", fontWeight = FontWeight.SemiBold)
                        Text("Tambahkan proyek pertama Anda untuk mengelompokkan tugas.", fontSize = 12.sp, color = MaterialTheme.colorScheme.onSurfaceVariant)
                    }
                }
            } else {
                LazyColumn(
                    verticalArrangement = Arrangement.spacedBy(10.dp),
                    modifier = Modifier.fillMaxSize()
                ) {
                    items(filteredProjects, key = { it.id }) { proj ->
                        val isExpanded = expandedIds.contains(proj.id)
                        val projectTasksList = remember(tasks, proj.id) {
                            tasks.filter { it.projectId == proj.id }
                        }
                        val projColor = remember(proj.color) {
                            try {
                                Color(android.graphics.Color.parseColor(proj.color))
                            } catch (e: Exception) {
                                BrandPrimary
                            }
                        }

                        Card(
                            modifier = Modifier
                                .fillMaxWidth()
                                .clip(RoundedCornerShape(16.dp))
                                .clickable { viewModel.toggleExpand(proj.id) },
                            shape = RoundedCornerShape(16.dp),
                            colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
                            elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
                        ) {
                            Column(modifier = Modifier.padding(14.dp)) {
                                Row(
                                    modifier = Modifier.fillMaxWidth(),
                                    verticalAlignment = Alignment.CenterVertically
                                ) {
                                    Box(
                                        modifier = Modifier
                                            .size(16.dp)
                                            .clip(CircleShape)
                                            .background(projColor)
                                    )
                                    Spacer(modifier = Modifier.width(10.dp))
                                    Column(modifier = Modifier.weight(1f)) {
                                        Text(proj.name, fontWeight = FontWeight.Bold, fontSize = 15.sp)
                                        if (proj.description.isNotBlank()) {
                                            Text(proj.description, fontSize = 12.sp, color = MaterialTheme.colorScheme.onSurfaceVariant, maxLines = 1)
                                        }
                                    }

                                    Surface(
                                        shape = CircleShape,
                                        color = projColor.copy(alpha = 0.12f)
                                    ) {
                                        Text(
                                            "${projectTasksList.size} Tugas",
                                            fontSize = 11.sp,
                                            fontWeight = FontWeight.SemiBold,
                                            color = projColor,
                                            modifier = Modifier.padding(horizontal = 8.dp, vertical = 2.dp)
                                        )
                                    }

                                    Spacer(modifier = Modifier.width(4.dp))

                                    IconButton(
                                        onClick = { editingProject = proj },
                                        modifier = Modifier.size(28.dp)
                                    ) {
                                        Icon(Icons.Default.Edit, contentDescription = "Edit Proyek", tint = MaterialTheme.colorScheme.onSurfaceVariant, modifier = Modifier.size(16.dp))
                                    }

                                    IconButton(
                                        onClick = { deletingProject = proj },
                                        modifier = Modifier.size(28.dp)
                                    ) {
                                        Icon(Icons.Default.DeleteOutline, contentDescription = "Hapus Proyek", tint = MaterialTheme.colorScheme.error.copy(alpha = 0.7f), modifier = Modifier.size(16.dp))
                                    }

                                    Icon(
                                        if (isExpanded) Icons.Default.ExpandLess else Icons.Default.ExpandMore,
                                        contentDescription = null,
                                        tint = MaterialTheme.colorScheme.onSurfaceVariant,
                                        modifier = Modifier.size(20.dp)
                                    )
                                }

                                // Expanded Tasks List inside project
                                if (isExpanded) {
                                    Spacer(modifier = Modifier.height(12.dp))
                                    Divider(color = MaterialTheme.colorScheme.surfaceVariant)
                                    Spacer(modifier = Modifier.height(8.dp))

                                    if (projectTasksList.isEmpty()) {
                                        Text(
                                            "Belum ada tugas di dalam proyek ini.",
                                            fontSize = 12.sp,
                                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                                            modifier = Modifier.padding(vertical = 6.dp)
                                        )
                                    } else {
                                        Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
                                            projectTasksList.forEach { pTask ->
                                                val isDone = pTask.status == "done"
                                                Row(
                                                    modifier = Modifier
                                                        .fillMaxWidth()
                                                        .clip(RoundedCornerShape(8.dp))
                                                        .background(MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.4f))
                                                        .padding(horizontal = 10.dp, vertical = 6.dp),
                                                    verticalAlignment = Alignment.CenterVertically
                                                ) {
                                                    Icon(
                                                        if (isDone) Icons.Default.CheckCircle else Icons.Default.RadioButtonUnchecked,
                                                        contentDescription = null,
                                                        tint = if (isDone) StatusDone else MaterialTheme.colorScheme.onSurfaceVariant,
                                                        modifier = Modifier.size(14.dp)
                                                    )
                                                    Spacer(modifier = Modifier.width(8.dp))
                                                    Text(
                                                        text = pTask.title,
                                                        fontSize = 13.sp,
                                                        fontWeight = FontWeight.Medium,
                                                        color = if (isDone) MaterialTheme.colorScheme.onSurfaceVariant else MaterialTheme.colorScheme.onSurface,
                                                        modifier = Modifier.weight(1f)
                                                    )
                                                    Spacer(modifier = Modifier.width(6.dp))
                                                    StatusBadge(status = pTask.status)
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
    }

    if (showAddDialog) {
        ProjectFormDialog(
            title = "Tambah Proyek Baru",
            initialName = "",
            initialColor = PROJECT_PRESET_COLORS.first(),
            initialDesc = "",
            onDismiss = { showAddDialog = false },
            onConfirm = { name, color, desc ->
                viewModel.createProject(name, color, desc)
                showAddDialog = false
            }
        )
    }

    editingProject?.let { proj ->
        ProjectFormDialog(
            title = "Edit Proyek",
            initialName = proj.name,
            initialColor = proj.color,
            initialDesc = proj.description,
            onDismiss = { editingProject = null },
            onConfirm = { name, color, desc ->
                viewModel.updateProject(proj.id, name, color, desc)
                editingProject = null
            }
        )
    }

    deletingProject?.let { proj ->
        AlertDialog(
            onDismissRequest = { deletingProject = null },
            title = { Text("Hapus Proyek?", fontWeight = FontWeight.Bold) },
            text = { Text("Apakah Anda yakin ingin menghapus proyek '${proj.name}'? Tugas yang ada di dalamnya tidak akan terhapus.") },
            confirmButton = {
                Button(
                    onClick = {
                        viewModel.deleteProject(proj.id)
                        deletingProject = null
                    },
                    colors = ButtonDefaults.buttonColors(containerColor = MaterialTheme.colorScheme.error)
                ) {
                    Text("Hapus", color = Color.White)
                }
            },
            dismissButton = {
                TextButton(onClick = { deletingProject = null }) {
                    Text("Batal")
                }
            }
        )
    }
}

@Composable
fun ProjectFormDialog(
    title: String,
    initialName: String,
    initialColor: String,
    initialDesc: String,
    onDismiss: () -> Unit,
    onConfirm: (String, String, String) -> Unit
) {
    var name by remember { mutableStateOf(initialName) }
    var color by remember { mutableStateOf(initialColor) }
    var desc by remember { mutableStateOf(initialDesc) }

    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text(title, fontWeight = FontWeight.Bold) },
        text = {
            Column(
                modifier = Modifier.verticalScroll(rememberScrollState()),
                verticalArrangement = Arrangement.spacedBy(12.dp)
            ) {
                OutlinedTextField(
                    value = name,
                    onValueChange = { name = it },
                    label = { Text("Nama Proyek *") },
                    shape = RoundedCornerShape(12.dp),
                    modifier = Modifier.fillMaxWidth(),
                    singleLine = true
                )

                OutlinedTextField(
                    value = desc,
                    onValueChange = { desc = it },
                    label = { Text("Deskripsi") },
                    shape = RoundedCornerShape(12.dp),
                    modifier = Modifier.fillMaxWidth(),
                    maxLines = 3
                )

                Text("Pilih Warna", fontSize = 12.sp, fontWeight = FontWeight.SemiBold)
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .horizontalScroll(rememberScrollState()),
                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    PROJECT_PRESET_COLORS.forEach { hex ->
                        val parsedColor = try {
                            Color(android.graphics.Color.parseColor(hex))
                        } catch (e: Exception) {
                            BrandPrimary
                        }
                        val isSelected = color.equals(hex, ignoreCase = true)
                        Box(
                            modifier = Modifier
                                .size(32.dp)
                                .clip(CircleShape)
                                .background(parsedColor)
                                .border(
                                    2.dp,
                                    if (isSelected) MaterialTheme.colorScheme.onSurface else Color.Transparent,
                                    CircleShape
                                )
                                .clickable { color = hex },
                            contentAlignment = Alignment.Center
                        ) {
                            if (isSelected) {
                                Icon(Icons.Default.Check, contentDescription = null, tint = Color.White, modifier = Modifier.size(16.dp))
                            }
                        }
                    }
                }
            }
        },
        confirmButton = {
            AppButton(
                text = "Simpan",
                enabled = name.isNotBlank(),
                onClick = { onConfirm(name.trim(), color, desc.trim()) }
            )
        },
        dismissButton = {
            TextButton(onClick = onDismiss) { Text("Batal") }
        }
    )
}
