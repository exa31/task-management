package cloud.eka_dev.task_hub.feature.calendar

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
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
import androidx.compose.ui.text.style.TextAlign
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
import java.time.DayOfWeek
import java.time.LocalDate
import java.time.YearMonth
import java.time.format.DateTimeFormatter
import java.time.format.TextStyle
import java.util.Locale

class CalendarViewModel(private val apiClient: ApiClient) : ViewModel() {
    private val _selectedDate = MutableStateFlow(LocalDate.now())
    val selectedDate: StateFlow<LocalDate> = _selectedDate.asStateFlow()

    private val _currentMonth = MutableStateFlow(YearMonth.now())
    val currentMonth: StateFlow<YearMonth> = _currentMonth.asStateFlow()

    private val _calendarData = MutableStateFlow(CalendarGraphQLData())
    val calendarData: StateFlow<CalendarGraphQLData> = _calendarData.asStateFlow()

    private val _isLoading = MutableStateFlow(false)
    val isLoading: StateFlow<Boolean> = _isLoading.asStateFlow()

    private val _filterCategory = MutableStateFlow("all") // all, tasks, reminders, activities
    val filterCategory: StateFlow<String> = _filterCategory.asStateFlow()

    init {
        loadMonthData(_currentMonth.value)
    }

    fun selectDate(date: LocalDate) {
        _selectedDate.value = date
        if (YearMonth.from(date) != _currentMonth.value) {
            _currentMonth.value = YearMonth.from(date)
            loadMonthData(_currentMonth.value)
        }
    }

    fun nextMonth() {
        val next = _currentMonth.value.plusMonths(1)
        _currentMonth.value = next
        loadMonthData(next)
    }

    fun prevMonth() {
        val prev = _currentMonth.value.minusMonths(1)
        _currentMonth.value = prev
        loadMonthData(prev)
    }

    fun jumpToToday() {
        val today = LocalDate.now()
        _selectedDate.value = today
        _currentMonth.value = YearMonth.now()
        loadMonthData(_currentMonth.value)
    }

    fun setFilterCategory(category: String) {
        _filterCategory.value = category
    }

    fun loadMonthData(month: YearMonth) {
        viewModelScope.launch {
            _isLoading.value = true
            val fromStr = month.atDay(1).atStartOfDay().format(DateTimeFormatter.ISO_LOCAL_DATE_TIME) + "Z"
            val toStr = month.atEndOfMonth().atTime(23, 59, 59).format(DateTimeFormatter.ISO_LOCAL_DATE_TIME) + "Z"

            val query = """
                query GetCalendarData(${'$'}from: String, ${'$'}to: String) {
                  tasks {
                    id
                    title
                    description
                    status
                    priority
                    due_date
                    is_focus
                    progress
                    tags
                    project {
                      id
                      name
                      color
                    }
                  }
                  reminders {
                    id
                    title
                    description
                    remind_at
                    status
                    task_id
                    task_title
                    snooze_count
                    created_at
                  }
                  activities(from: ${'$'}from, to: ${'$'}to) {
                    id
                    task_id
                    task_title
                    activity_type
                    description
                    created_at
                  }
                }
            """.trimIndent()

            try {
                val res = apiClient.apiService.executeGraphQL(
                    GraphQLRequest(query = query, variables = mapOf("from" to fromStr, "to" to toStr))
                )
                if (res.data != null) {
                    _calendarData.value = res.data
                }
            } catch (e: Exception) {
                // Fallback to REST APIs if GraphQL encounter issues
                try {
                    val tasksRes = apiClient.apiService.getTasks()
                    val remRes = apiClient.apiService.getReminders()
                    val tasks = tasksRes.data?.map { t ->
                        GraphQLTaskItem(
                            id = t.id,
                            title = t.title,
                            description = t.description,
                            status = t.status,
                            priority = t.priority,
                            dueDate = t.dueDate,
                            isFocus = t.isFocus,
                            progress = t.progressPercent,
                            tags = t.tags,
                            project = if (!t.projectName.isNullOrEmpty()) GraphQLProjectItem(t.projectId ?: 0, t.projectName, t.projectColor) else null
                        )
                    } ?: emptyList()
                    val reminders = remRes.data?.map { r ->
                        GraphQLReminderItem(
                            id = r.id,
                            title = r.title,
                            description = r.description,
                            remindAt = r.remindAt,
                            status = r.status,
                            taskId = r.taskId,
                            taskTitle = r.taskTitle,
                            snoozeCount = r.snoozeCount
                        )
                    } ?: emptyList()
                    _calendarData.value = CalendarGraphQLData(tasks = tasks, reminders = reminders, activities = emptyList())
                } catch (ex: Exception) {
                    // Ignore
                }
            } finally {
                _isLoading.value = false
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun CalendarScreen(
    viewModel: CalendarViewModel
) {
    val selectedDate by viewModel.selectedDate.collectAsState()
    val currentMonth by viewModel.currentMonth.collectAsState()
    val calendarData by viewModel.calendarData.collectAsState()
    val isLoading by viewModel.isLoading.collectAsState()
    val filterCategory by viewModel.filterCategory.collectAsState()

    val selectedDateStr = remember(selectedDate) {
        selectedDate.format(DateTimeFormatter.ISO_LOCAL_DATE)
    }

    // Filter items for selected date
    val dayTasks = remember(calendarData.tasks, selectedDateStr) {
        calendarData.tasks.filter { it.dueDate?.take(10) == selectedDateStr }
    }
    val dayReminders = remember(calendarData.reminders, selectedDateStr) {
        calendarData.reminders.filter { it.remindAt.take(10) == selectedDateStr }
    }
    val dayActivities = remember(calendarData.activities, selectedDateStr) {
        calendarData.activities.filter { it.createdAt.take(10) == selectedDateStr }
    }

    // Dates with items in current month
    val taskDates = remember(calendarData.tasks) {
        calendarData.tasks.mapNotNull { it.dueDate?.take(10) }.toSet()
    }
    val reminderDates = remember(calendarData.reminders) {
        calendarData.reminders.map { it.remindAt.take(10) }.toSet()
    }
    val activityDates = remember(calendarData.activities) {
        calendarData.activities.map { it.createdAt.take(10) }.toSet()
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
                        Text(stringResource(R.string.calendar_title), fontWeight = FontWeight.Bold)
                    }
                },
                actions = {
                    TextButton(onClick = { viewModel.jumpToToday() }) {
                        Text(stringResource(R.string.calendar_today), fontWeight = FontWeight.SemiBold, color = BrandPrimary)
                    }
                }
            )
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(horizontal = 16.dp),
            verticalArrangement = Arrangement.spacedBy(10.dp)
        ) {
            // Month Header Selector
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                IconButton(onClick = { viewModel.prevMonth() }) {
                    Icon(Icons.Default.ChevronLeft, contentDescription = null)
                }
                Text(
                    text = "${currentMonth.month.getDisplayName(TextStyle.FULL, Locale.getDefault())} ${currentMonth.year}",
                    fontSize = 17.sp,
                    fontWeight = FontWeight.Bold
                )
                IconButton(onClick = { viewModel.nextMonth() }) {
                    Icon(Icons.Default.ChevronRight, contentDescription = null)
                }
            }

            // Calendar Grid Card
            Card(
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(16.dp),
                colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
                elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
            ) {
                Column(modifier = Modifier.padding(10.dp)) {
                    // Weekday Labels
                    val weekdays = remember {
                        listOf(
                            DayOfWeek.SUNDAY, DayOfWeek.MONDAY, DayOfWeek.TUESDAY,
                            DayOfWeek.WEDNESDAY, DayOfWeek.THURSDAY, DayOfWeek.FRIDAY, DayOfWeek.SATURDAY
                        ).map { it.getDisplayName(TextStyle.SHORT, Locale.getDefault()) }
                    }
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.SpaceAround
                    ) {
                        weekdays.forEach { day ->
                            Text(
                                text = day,
                                fontSize = 11.sp,
                                fontWeight = FontWeight.SemiBold,
                                color = MaterialTheme.colorScheme.onSurfaceVariant,
                                modifier = Modifier.width(36.dp),
                                textAlign = TextAlign.Center
                            )
                        }
                    }

                    Spacer(modifier = Modifier.height(6.dp))

                    // Days Grid
                    val firstDayOfMonth = currentMonth.atDay(1)
                    val daysInMonth = currentMonth.lengthOfMonth()
                    // java.time DayOfWeek: 1 (Mon) to 7 (Sun). We want 0 (Sun) to 6 (Sat)
                    val startOffset = firstDayOfMonth.dayOfWeek.value % 7

                    val totalCells = ((startOffset + daysInMonth + 6) / 7) * 7

                    Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
                        for (row in 0 until (totalCells / 7)) {
                            Row(
                                modifier = Modifier.fillMaxWidth(),
                                horizontalArrangement = Arrangement.SpaceAround
                            ) {
                                for (col in 0 until 7) {
                                    val cellIndex = row * 7 + col
                                    val dayNumber = cellIndex - startOffset + 1

                                    if (dayNumber in 1..daysInMonth) {
                                        val date = currentMonth.atDay(dayNumber)
                                        val dateKey = date.format(DateTimeFormatter.ISO_LOCAL_DATE)
                                        val isSelected = date == selectedDate
                                        val isToday = date == LocalDate.now()
                                        val hasTasks = taskDates.contains(dateKey)
                                        val hasReminders = reminderDates.contains(dateKey)
                                        val hasActivities = activityDates.contains(dateKey)

                                        Box(
                                            modifier = Modifier
                                                .size(36.dp)
                                                .clip(CircleShape)
                                                .background(
                                                    when {
                                                        isSelected -> BrandPrimary
                                                        isToday -> BrandPrimary.copy(alpha = 0.12f)
                                                        else -> Color.Transparent
                                                    }
                                                )
                                                .clickable { viewModel.selectDate(date) },
                                            contentAlignment = Alignment.Center
                                        ) {
                                            Column(
                                                horizontalAlignment = Alignment.CenterHorizontally,
                                                verticalArrangement = Arrangement.Center
                                            ) {
                                                Text(
                                                    text = "$dayNumber",
                                                    fontSize = 12.sp,
                                                    fontWeight = if (isSelected || isToday) FontWeight.Bold else FontWeight.Normal,
                                                    color = when {
                                                        isSelected -> Color.White
                                                        isToday -> BrandPrimary
                                                        else -> MaterialTheme.colorScheme.onSurface
                                                    }
                                                )
                                                Row(
                                                    horizontalArrangement = Arrangement.spacedBy(1.5.dp),
                                                    modifier = Modifier.height(4.dp)
                                                ) {
                                                    if (hasTasks) {
                                                        Box(modifier = Modifier.size(3.dp).clip(CircleShape).background(if (isSelected) Color.White else Color(0xFF3B82F6)))
                                                    }
                                                    if (hasReminders) {
                                                        Box(modifier = Modifier.size(3.dp).clip(CircleShape).background(if (isSelected) Color.White else Color(0xFFF59E0B)))
                                                    }
                                                    if (hasActivities) {
                                                        Box(modifier = Modifier.size(3.dp).clip(CircleShape).background(if (isSelected) Color.White else Color(0xFF10B981)))
                                                    }
                                                }
                                            }
                                        }
                                    } else {
                                        Spacer(modifier = Modifier.size(36.dp))
                                    }
                                }
                            }
                        }
                    }
                }
            }

            // Category Filter Chips
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                FilterChip(
                    selected = filterCategory == "all",
                    onClick = { viewModel.setFilterCategory("all") },
                    label = { Text(stringResource(R.string.calendar_all_filter, dayTasks.size + dayReminders.size + dayActivities.size), fontSize = 11.sp) }
                )
                FilterChip(
                    selected = filterCategory == "tasks",
                    onClick = { viewModel.setFilterCategory("tasks") },
                    leadingIcon = { Box(modifier = Modifier.size(6.dp).clip(CircleShape).background(Color(0xFF3B82F6))) },
                    label = { Text(stringResource(R.string.calendar_tasks_filter, dayTasks.size), fontSize = 11.sp) }
                )
                FilterChip(
                    selected = filterCategory == "reminders",
                    onClick = { viewModel.setFilterCategory("reminders") },
                    leadingIcon = { Box(modifier = Modifier.size(6.dp).clip(CircleShape).background(Color(0xFFF59E0B))) },
                    label = { Text(stringResource(R.string.calendar_reminders_filter, dayReminders.size), fontSize = 11.sp) }
                )
                FilterChip(
                    selected = filterCategory == "activities",
                    onClick = { viewModel.setFilterCategory("activities") },
                    leadingIcon = { Box(modifier = Modifier.size(6.dp).clip(CircleShape).background(Color(0xFF10B981))) },
                    label = { Text(stringResource(R.string.calendar_activities_filter, dayActivities.size), fontSize = 11.sp) }
                )
            }

            // Day Agenda Timeline
            val formattedDate = "${selectedDate.dayOfMonth} ${selectedDate.month.getDisplayName(TextStyle.FULL, Locale.getDefault())} ${selectedDate.year}"
            Text(
                text = stringResource(R.string.calendar_agenda_title, formattedDate),
                fontSize = 14.sp,
                fontWeight = FontWeight.Bold,
                color = MaterialTheme.colorScheme.onSurface
            )

            val showTasks = filterCategory == "all" || filterCategory == "tasks"
            val showReminders = filterCategory == "all" || filterCategory == "reminders"
            val showActivities = filterCategory == "all" || filterCategory == "activities"

            val totalItems = (if (showTasks) dayTasks.size else 0) +
                             (if (showReminders) dayReminders.size else 0) +
                             (if (showActivities) dayActivities.size else 0)

            if (isLoading) {
                TaskListSkeleton(count = 2)
            } else if (totalItems == 0) {
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .weight(1f),
                    contentAlignment = Alignment.Center
                ) {
                    Column(
                        horizontalAlignment = Alignment.CenterHorizontally,
                        verticalArrangement = Arrangement.spacedBy(6.dp)
                    ) {
                        Icon(Icons.Default.EventAvailable, contentDescription = null, tint = MaterialTheme.colorScheme.onSurfaceVariant, modifier = Modifier.size(36.dp))
                        Text(stringResource(R.string.calendar_no_schedule), fontSize = 13.sp, color = MaterialTheme.colorScheme.onSurfaceVariant)
                    }
                }
            } else {
                LazyColumn(
                    verticalArrangement = Arrangement.spacedBy(8.dp),
                    modifier = Modifier
                        .fillMaxWidth()
                        .weight(1f)
                ) {
                    if (showTasks && dayTasks.isNotEmpty()) {
                        item {
                            Text(stringResource(R.string.calendar_tasks_due), fontSize = 11.sp, fontWeight = FontWeight.Bold, color = Color(0xFF3B82F6))
                        }
                        items(dayTasks, key = { "task-${it.id}" }) { t ->
                            Card(
                                modifier = Modifier.fillMaxWidth(),
                                shape = RoundedCornerShape(12.dp),
                                colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface)
                            ) {
                                Row(
                                    modifier = Modifier.padding(12.dp),
                                    verticalAlignment = Alignment.CenterVertically
                                ) {
                                    Icon(
                                        if (t.status == "done") Icons.Default.CheckCircle else Icons.Default.RadioButtonUnchecked,
                                        contentDescription = null,
                                        tint = if (t.status == "done") StatusDone else Color(0xFF3B82F6),
                                        modifier = Modifier.size(16.dp)
                                    )
                                    Spacer(modifier = Modifier.width(10.dp))
                                    Column(modifier = Modifier.weight(1f)) {
                                        Text(t.title, fontWeight = FontWeight.SemiBold, fontSize = 13.sp)
                                        t.project?.let { p ->
                                            Text(p.name, fontSize = 11.sp, color = MaterialTheme.colorScheme.onSurfaceVariant)
                                        }
                                    }
                                    StatusBadge(status = t.status)
                                }
                            }
                        }
                    }

                    if (showReminders && dayReminders.isNotEmpty()) {
                        item {
                            Spacer(modifier = Modifier.height(4.dp))
                            Text(stringResource(R.string.calendar_active_reminders), fontSize = 11.sp, fontWeight = FontWeight.Bold, color = Color(0xFFF59E0B))
                        }
                        items(dayReminders, key = { "rem-${it.id}" }) { r ->
                            Card(
                                modifier = Modifier.fillMaxWidth(),
                                shape = RoundedCornerShape(12.dp),
                                colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface)
                            ) {
                                Row(
                                    modifier = Modifier.padding(12.dp),
                                    verticalAlignment = Alignment.CenterVertically
                                ) {
                                    Icon(Icons.Default.Notifications, contentDescription = null, tint = Color(0xFFF59E0B), modifier = Modifier.size(16.dp))
                                    Spacer(modifier = Modifier.width(10.dp))
                                    Column(modifier = Modifier.weight(1f)) {
                                        Text(r.title, fontWeight = FontWeight.SemiBold, fontSize = 13.sp)
                                        Text(stringResource(R.string.reminder_schedule_prefix, r.remindAt.take(16).replace("T", " ")), fontSize = 11.sp, color = MaterialTheme.colorScheme.onSurfaceVariant)
                                    }
                                    Surface(
                                        shape = CircleShape,
                                        color = if (r.status == "sent") StatusDone.copy(alpha = 0.12f) else Color(0xFFF59E0B).copy(alpha = 0.12f)
                                    ) {
                                        Text(
                                            r.status.uppercase(),
                                            fontSize = 9.sp,
                                            fontWeight = FontWeight.Bold,
                                            color = if (r.status == "sent") StatusDone else Color(0xFFF59E0B),
                                            modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp)
                                        )
                                    }
                                }
                            }
                        }
                    }

                    if (showActivities && dayActivities.isNotEmpty()) {
                        item {
                            Spacer(modifier = Modifier.height(4.dp))
                            Text(stringResource(R.string.calendar_activity_history), fontSize = 11.sp, fontWeight = FontWeight.Bold, color = Color(0xFF10B981))
                        }
                        items(dayActivities, key = { "act-${it.id}" }) { a ->
                            Card(
                                modifier = Modifier.fillMaxWidth(),
                                shape = RoundedCornerShape(12.dp),
                                colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface)
                            ) {
                                Row(
                                    modifier = Modifier.padding(12.dp),
                                    verticalAlignment = Alignment.CenterVertically
                                ) {
                                    Icon(Icons.Default.History, contentDescription = null, tint = Color(0xFF10B981), modifier = Modifier.size(16.dp))
                                    Spacer(modifier = Modifier.width(10.dp))
                                    Column(modifier = Modifier.weight(1f)) {
                                        Text(a.description, fontSize = 12.sp)
                                        Text(a.createdAt.take(16).replace("T", " "), fontSize = 10.sp, color = MaterialTheme.colorScheme.onSurfaceVariant)
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
