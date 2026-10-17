package cloud.eka_dev.task_hub.feature.task

import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.rememberScrollState
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
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import cloud.eka_dev.task_hub.R
import cloud.eka_dev.task_hub.core.designsystem.*
import cloud.eka_dev.task_hub.core.model.Task

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TaskListScreen(
    viewModel: TaskViewModel,
    onTaskClick: (Long) -> Unit
) {
    val uiState by viewModel.uiState.collectAsState()
    var showCreateDialog by remember { mutableStateOf(false) }

    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    Row(
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.spacedBy(10.dp)
                    ) {
                        BrandLogo(size = 28.dp)
                        Text(
                            stringResource(R.string.task_list_title),
                            fontSize = 20.sp,
                            fontWeight = FontWeight.Bold
                        )
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = MaterialTheme.colorScheme.background
                )
            )
        },
        floatingActionButton = {
            FloatingActionButton(
                onClick = { showCreateDialog = true },
                containerColor = BrandPrimary,
                contentColor = Color.White,
                shape = CircleShape
            ) {
                Icon(Icons.Default.Add, contentDescription = stringResource(R.string.create_task))
            }
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(horizontal = 16.dp)
        ) {
            // Search Input with sleek rounded corners
            OutlinedTextField(
                value = uiState.searchQuery,
                onValueChange = { viewModel.setSearchQuery(it) },
                placeholder = { Text(stringResource(R.string.search_tasks_hint), fontSize = 14.sp) },
                leadingIcon = { Icon(Icons.Default.Search, contentDescription = null, tint = BrandPrimary) },
                shape = RoundedCornerShape(14.dp),
                modifier = Modifier.fillMaxWidth(),
                singleLine = true,
                colors = OutlinedTextFieldDefaults.colors(
                    unfocusedContainerColor = MaterialTheme.colorScheme.surface,
                    focusedContainerColor = MaterialTheme.colorScheme.surface
                )
            )

            Spacer(modifier = Modifier.height(10.dp))

            // Filter Chips with icons
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .horizontalScroll(rememberScrollState()),
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                FilterChip(
                    selected = uiState.selectedStatus == null,
                    onClick = { viewModel.setStatusFilter(null) },
                    leadingIcon = { Icon(Icons.Default.Inbox, contentDescription = null, modifier = Modifier.size(14.dp)) },
                    label = { Text(stringResource(R.string.filter_all)) }
                )
                FilterChip(
                    selected = uiState.selectedStatus == "todo",
                    onClick = { viewModel.setStatusFilter("todo") },
                    leadingIcon = { Icon(Icons.Default.RadioButtonUnchecked, contentDescription = null, modifier = Modifier.size(14.dp)) },
                    label = { Text(stringResource(R.string.filter_todo)) }
                )
                FilterChip(
                    selected = uiState.selectedStatus == "in_progress",
                    onClick = { viewModel.setStatusFilter("in_progress") },
                    leadingIcon = { Icon(Icons.Default.Sync, contentDescription = null, modifier = Modifier.size(14.dp)) },
                    label = { Text(stringResource(R.string.filter_in_progress)) }
                )
                FilterChip(
                    selected = uiState.selectedStatus == "review",
                    onClick = { viewModel.setStatusFilter("review") },
                    leadingIcon = { Icon(Icons.Default.Visibility, contentDescription = null, modifier = Modifier.size(14.dp)) },
                    label = { Text(stringResource(R.string.filter_review)) }
                )
                FilterChip(
                    selected = uiState.selectedStatus == "done",
                    onClick = { viewModel.setStatusFilter("done") },
                    leadingIcon = { Icon(Icons.Default.CheckCircle, contentDescription = null, modifier = Modifier.size(14.dp)) },
                    label = { Text(stringResource(R.string.filter_done)) }
                )
                FilterChip(
                    selected = uiState.filterFocusOnly,
                    onClick = { viewModel.toggleFocusFilter() },
                    leadingIcon = {
                        Icon(
                            if (uiState.filterFocusOnly) Icons.Default.Star else Icons.Default.StarBorder,
                            contentDescription = null,
                            tint = if (uiState.filterFocusOnly) Color(0xFFF59E0B) else MaterialTheme.colorScheme.onSurfaceVariant,
                            modifier = Modifier.size(14.dp)
                        )
                    },
                    label = {
                        val count = viewModel.focusTasks.size
                        Text(if (count > 0) stringResource(R.string.filter_focus_count, count) else stringResource(R.string.filter_focus))
                    }
                )
            }

            Spacer(modifier = Modifier.height(14.dp))

            // Task List Content
            if (uiState.isLoading) {
                TaskListSkeleton(count = 6)
            } else if (uiState.tasks.isEmpty()) {
                Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                    Column(
                        horizontalAlignment = Alignment.CenterHorizontally,
                        verticalArrangement = Arrangement.spacedBy(12.dp)
                    ) {
                        Box(
                            modifier = Modifier
                                .size(64.dp)
                                .clip(CircleShape)
                                .background(BrandPrimary.copy(alpha = 0.12f)),
                            contentAlignment = Alignment.Center
                        ) {
                            Icon(
                                Icons.Default.Checklist,
                                contentDescription = null,
                                tint = BrandPrimary,
                                modifier = Modifier.size(32.dp)
                            )
                        }
                        Text(
                            stringResource(R.string.empty_tasks),
                            fontSize = 15.sp,
                            fontWeight = FontWeight.SemiBold
                        )
                        Text(
                            stringResource(R.string.empty_tasks_sub),
                            fontSize = 13.sp,
                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                            textAlign = TextAlign.Center
                        )
                    }
                }
            } else {
                val listState = rememberLazyListState()

                val displayedTasks = remember(uiState.tasks, uiState.filterFocusOnly) {
                    if (uiState.filterFocusOnly) {
                        uiState.tasks.filter { it.isFocus }
                    } else {
                        uiState.tasks
                    }
                }

                val shouldLoadMore = remember {
                    derivedStateOf {
                        val total = listState.layoutInfo.totalItemsCount
                        val lastVisible = listState.layoutInfo.visibleItemsInfo.lastOrNull()?.index ?: 0
                        total > 0 && lastVisible >= total - 3
                    }
                }

                LaunchedEffect(shouldLoadMore.value) {
                    if (shouldLoadMore.value && uiState.hasMore && !uiState.isLoadingMore && !uiState.filterFocusOnly) {
                        viewModel.loadMoreTasks()
                    }
                }

                LazyColumn(
                    state = listState,
                    verticalArrangement = Arrangement.spacedBy(10.dp),
                    modifier = Modifier.fillMaxSize()
                ) {
                    // Today's Focus Horizon Banner
                    if (viewModel.focusTasks.isNotEmpty() && !uiState.filterFocusOnly) {
                        item {
                            TaskFocusHorizonBanner(
                                focusTasks = viewModel.focusTasks,
                                onTaskClick = { task ->
                                    viewModel.selectTask(task)
                                    onTaskClick(task.id)
                                },
                                onFocusToggle = { task -> viewModel.toggleTaskFocus(task) }
                            )
                        }
                    }

                    items(displayedTasks, key = { it.id }) { task ->
                        TaskItemCard(
                            task = task,
                            onTaskClick = {
                                viewModel.selectTask(task)
                                onTaskClick(task.id)
                            },
                            onStatusToggle = { viewModel.toggleTaskStatus(task) },
                            onFocusToggle = { viewModel.toggleTaskFocus(task) },
                            onStatusChange = { newStatus -> viewModel.updateTaskStatus(task, newStatus) }
                        )
                    }

                    if (uiState.isLoadingMore && !uiState.filterFocusOnly) {
                        item {
                            Box(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .padding(16.dp),
                                contentAlignment = Alignment.Center
                            ) {
                                CircularProgressIndicator(
                                    modifier = Modifier.size(24.dp),
                                    strokeWidth = 2.dp,
                                    color = BrandPrimary
                                )
                            }
                        }
                    }
                }
            }
        }
    }

    uiState.selectedTask?.let { task ->
        TaskDetailBottomSheet(
            task = task,
            viewModel = viewModel,
            onDismiss = { viewModel.selectTask(null) }
        )
    }

    if (showCreateDialog) {
        CreateTaskDialog(
            onDismiss = { showCreateDialog = false },
            onCreate = { title, desc, priority, isFocus ->
                viewModel.createTask(title, desc, priority, isFocus)
                showCreateDialog = false
            }
        )
    }
}

@Composable
fun CreateTaskDialog(
    onDismiss: () -> Unit,
    onCreate: (String, String, String, Boolean) -> Unit
) {
    var title by remember { mutableStateOf("") }
    var description by remember { mutableStateOf("") }
    var priority by remember { mutableStateOf("medium") }
    var isFocus by remember { mutableStateOf(false) }

    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text(stringResource(R.string.create_task), fontWeight = FontWeight.Bold) },
        text = {
            Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
                OutlinedTextField(
                    value = title,
                    onValueChange = { title = it },
                    label = { Text(stringResource(R.string.task_title_label)) },
                    shape = RoundedCornerShape(12.dp),
                    modifier = Modifier.fillMaxWidth()
                )
                OutlinedTextField(
                    value = description,
                    onValueChange = { description = it },
                    label = { Text(stringResource(R.string.task_desc_label)) },
                    shape = RoundedCornerShape(12.dp),
                    modifier = Modifier.fillMaxWidth(),
                    maxLines = 3
                )
                Text(stringResource(R.string.priority_label), fontSize = 12.sp, fontWeight = FontWeight.SemiBold)
                Row(horizontalArrangement = Arrangement.spacedBy(6.dp)) {
                    listOf("low", "medium", "high", "urgent").forEach { p ->
                        val pLabel = when (p) {
                            "low" -> stringResource(R.string.priority_low)
                            "medium" -> stringResource(R.string.priority_medium)
                            "high" -> stringResource(R.string.priority_high)
                            "urgent" -> stringResource(R.string.priority_urgent)
                            else -> p
                        }
                        FilterChip(
                            selected = priority == p,
                            onClick = { priority = p },
                            label = { Text(pLabel, fontSize = 11.sp) }
                        )
                    }
                }
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.SpaceBetween
                ) {
                    Text(stringResource(R.string.focus_today), fontSize = 13.sp)
                    Switch(
                        checked = isFocus,
                        onCheckedChange = { isFocus = it }
                    )
                }
            }
        },
        confirmButton = {
            AppButton(
                text = stringResource(R.string.save),
                enabled = title.isNotBlank(),
                onClick = { onCreate(title, description, priority, isFocus) }
            )
        },
        dismissButton = {
            TextButton(onClick = onDismiss) {
                Text(stringResource(R.string.cancel))
            }
        }
    )
}
