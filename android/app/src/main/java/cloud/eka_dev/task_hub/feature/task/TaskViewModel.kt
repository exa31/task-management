package cloud.eka_dev.task_hub.feature.task

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import cloud.eka_dev.task_hub.core.model.*
import cloud.eka_dev.task_hub.core.network.ApiClient
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

data class TaskUiState(
    val isLoading: Boolean = false,
    val tasks: List<Task> = emptyList(),
    val selectedStatus: String? = null,
    val searchQuery: String = "",
    val error: String? = null,
    val selectedTask: Task? = null,
    val taskActivities: List<TaskActivity> = emptyList(),
    val isLoadingActivities: Boolean = false
)

class TaskViewModel(private val apiClient: ApiClient) : ViewModel() {

    private val _uiState = MutableStateFlow(TaskUiState())
    val uiState: StateFlow<TaskUiState> = _uiState.asStateFlow()

    init {
        loadTasks()
    }

    fun loadTasks() {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true, error = null)
            try {
                val res = apiClient.apiService.getTasks(
                    status = _uiState.value.selectedStatus,
                    search = _uiState.value.searchQuery.ifEmpty { null }
                )
                if (res.success && res.data != null) {
                    _uiState.value = _uiState.value.copy(
                        isLoading = false,
                        tasks = res.data
                    )
                } else {
                    _uiState.value = _uiState.value.copy(
                        isLoading = false,
                        error = res.message
                    )
                }
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    error = e.localizedMessage ?: "Failed to load tasks"
                )
            }
        }
    }

    fun setStatusFilter(status: String?) {
        _uiState.value = _uiState.value.copy(selectedStatus = status)
        loadTasks()
    }

    fun setSearchQuery(query: String) {
        _uiState.value = _uiState.value.copy(searchQuery = query)
        loadTasks()
    }

    fun toggleTaskStatus(task: Task) {
        viewModelScope.launch {
            val newStatus = if (task.status == "done") "todo" else "done"
            try {
                val res = apiClient.apiService.updateTask(task.id, UpdateTaskRequest(status = newStatus))
                if (res.success && res.data != null) {
                    if (_uiState.value.selectedTask?.id == task.id) {
                        _uiState.value = _uiState.value.copy(selectedTask = res.data)
                    }
                }
                loadTasks()
            } catch (e: Exception) {
                // handle error
            }
        }
    }

    fun updateTaskStatus(task: Task, newStatus: String) {
        viewModelScope.launch {
            try {
                val updatedTasks = _uiState.value.tasks.map {
                    if (it.id == task.id) it.copy(status = newStatus) else it
                }
                _uiState.value = _uiState.value.copy(tasks = updatedTasks)
                if (_uiState.value.selectedTask?.id == task.id) {
                    _uiState.value = _uiState.value.copy(selectedTask = _uiState.value.selectedTask?.copy(status = newStatus))
                }

                val res = apiClient.apiService.updateTask(task.id, UpdateTaskRequest(status = newStatus))
                if (res.success && res.data != null) {
                    if (_uiState.value.selectedTask?.id == task.id) {
                        _uiState.value = _uiState.value.copy(selectedTask = res.data)
                    }
                }
                loadTasks()
            } catch (e: Exception) {
                loadTasks()
            }
        }
    }

    fun updateTask(
        id: Long,
        title: String? = null,
        description: String? = null,
        status: String? = null,
        priority: String? = null,
        dueDate: String? = null,
        isFocus: Boolean? = null,
        tags: List<String>? = null,
        projectId: Long? = null,
        onSuccess: (() -> Unit)? = null
    ) {
        viewModelScope.launch {
            try {
                val res = apiClient.apiService.updateTask(
                    id,
                    UpdateTaskRequest(
                        title = title,
                        description = description,
                        status = status,
                        priority = priority,
                        dueDate = dueDate,
                        isFocus = isFocus,
                        tags = tags,
                        projectId = projectId
                    )
                )
                if (res.success && res.data != null) {
                    if (_uiState.value.selectedTask?.id == id) {
                        _uiState.value = _uiState.value.copy(selectedTask = res.data)
                    }
                    loadTasks()
                    onSuccess?.invoke()
                }
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    error = e.localizedMessage ?: "Failed to update task"
                )
            }
        }
    }

    fun createTask(title: String, description: String, priority: String, projectId: Long? = null) {
        viewModelScope.launch {
            try {
                apiClient.apiService.createTask(
                    CreateTaskRequest(
                        title = title,
                        description = description,
                        priority = priority,
                        projectId = projectId
                    )
                )
                loadTasks()
            } catch (e: Exception) {
                // handle error
            }
        }
    }

    fun toggleSubtask(taskId: Long, subtaskId: Long, currentCompleted: Boolean) {
        viewModelScope.launch {
            try {
                apiClient.apiService.updateSubtask(
                    taskId,
                    subtaskId,
                    UpdateSubtaskRequest(isCompleted = !currentCompleted)
                )
                loadTasks()
            } catch (e: Exception) {
                // handle error
            }
        }
    }

    fun selectTask(task: Task?) {
        _uiState.value = _uiState.value.copy(selectedTask = task, taskActivities = emptyList())
        if (task != null) {
            loadActivities(task.id)
        }
    }

    fun loadActivities(taskId: Long) {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoadingActivities = true)
            try {
                val res = apiClient.apiService.getTaskActivities(taskId)
                if (res.success && res.data != null) {
                    _uiState.value = _uiState.value.copy(
                        isLoadingActivities = false,
                        taskActivities = res.data
                    )
                } else {
                    _uiState.value = _uiState.value.copy(isLoadingActivities = false)
                }
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(isLoadingActivities = false)
            }
        }
    }

    fun addActivityNote(taskId: Long, note: String) {
        if (note.isBlank()) return
        viewModelScope.launch {
            try {
                val res = apiClient.apiService.createTaskActivity(
                    taskId,
                    CreateTaskActivityRequest(description = note.trim())
                )
                if (res.success && res.data != null) {
                    val currentList = _uiState.value.taskActivities
                    _uiState.value = _uiState.value.copy(
                        taskActivities = listOf(res.data) + currentList
                    )
                }
            } catch (e: Exception) {
                // handle error
            }
        }
    }
}
