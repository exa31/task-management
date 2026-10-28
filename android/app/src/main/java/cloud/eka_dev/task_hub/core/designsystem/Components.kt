package cloud.eka_dev.task_hub.core.designsystem

import androidx.compose.animation.core.*
import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import cloud.eka_dev.task_hub.R
import cloud.eka_dev.task_hub.core.model.Task

@Composable
fun BrandLogo(size: Dp = 36.dp) {
    Box(
            modifier =
                    Modifier.size(size)
                            .clip(RoundedCornerShape(size * 0.28f))
                            .background(BrandGradientBrush),
            contentAlignment = Alignment.Center
    ) {
        Icon(
                imageVector = Icons.Default.Bolt,
                contentDescription = "Logo",
                tint = Color.White,
                modifier = Modifier.size(size * 0.55f)
        )
    }
}

@Composable
fun AppButton(
        text: String,
        onClick: () -> Unit,
        modifier: Modifier = Modifier,
        enabled: Boolean = true,
        isLoading: Boolean = false
) {
    Button(
            onClick = onClick,
            modifier = modifier.height(50.dp),
            enabled = enabled && !isLoading,
            shape = RoundedCornerShape(14.dp),
            colors =
                    ButtonDefaults.buttonColors(
                            containerColor = BrandPrimary,
                            contentColor = Color.White
                    )
    ) {
        if (isLoading) {
            CircularProgressIndicator(
                    modifier = Modifier.size(20.dp),
                    color = Color.White,
                    strokeWidth = 2.dp
            )
        } else {
            Text(text = text, fontSize = 15.sp, fontWeight = FontWeight.SemiBold)
        }
    }
}

@Composable
fun PriorityBadge(priority: String) {
    val (color, icon) =
            when (priority.lowercase()) {
                "urgent" -> Pair(PriorityUrgent, Icons.Default.Warning)
                "high" -> Pair(PriorityHigh, Icons.Default.KeyboardDoubleArrowUp)
                "medium" -> Pair(PriorityMedium, Icons.Default.Remove)
                "low" -> Pair(PriorityLow, Icons.Default.KeyboardArrowDown)
                else -> Pair(PriorityLow, Icons.Default.Remove)
            }

    Row(
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(3.dp),
            modifier =
                    Modifier.clip(RoundedCornerShape(6.dp))
                            .background(color.copy(alpha = 0.12f))
                            .border(1.dp, color.copy(alpha = 0.25f), RoundedCornerShape(6.dp))
                            .padding(horizontal = 6.dp, vertical = 3.dp)
    ) {
        Icon(icon, contentDescription = null, tint = color, modifier = Modifier.size(11.dp))
        Text(
                text = priority.uppercase(),
                color = color,
                fontSize = 10.sp,
                fontWeight = FontWeight.Bold
        )
    }
}

@Composable
fun StatusBadge(status: String) {
    val color =
            when (status.lowercase()) {
                "done" -> StatusDone
                "in_progress" -> StatusInProgress
                "review" -> StatusReview
                "backlog" -> StatusBacklog
                else -> StatusTodo
            }

    Box(
            modifier =
                    Modifier.clip(RoundedCornerShape(6.dp))
                            .background(color.copy(alpha = 0.12f))
                            .border(1.dp, color.copy(alpha = 0.25f), RoundedCornerShape(6.dp))
                            .padding(horizontal = 6.dp, vertical = 3.dp)
    ) {
        Text(
                text = status.replace("_", " ").uppercase(),
                color = color,
                fontSize = 10.sp,
                fontWeight = FontWeight.Bold
        )
    }
}

@Composable
fun ProjectPill(name: String, colorHex: String?) {
    val color =
            try {
                if (!colorHex.isNullOrEmpty()) Color(android.graphics.Color.parseColor(colorHex))
                else BrandPrimary
            } catch (e: Exception) {
                BrandPrimary
            }

    Row(
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(4.dp),
            modifier =
                    Modifier.clip(RoundedCornerShape(6.dp))
                            .background(color.copy(alpha = 0.12f))
                            .border(1.dp, color.copy(alpha = 0.3f), RoundedCornerShape(6.dp))
                            .padding(horizontal = 6.dp, vertical = 2.5.dp)
    ) {
        Icon(
                Icons.Default.Folder,
                contentDescription = null,
                tint = color,
                modifier = Modifier.size(10.dp)
        )
        Text(name, color = color, fontSize = 10.sp, fontWeight = FontWeight.SemiBold)
    }
}

@Composable
fun TaskItemCard(
        task: Task,
        onTaskClick: () -> Unit,
        onStatusToggle: () -> Unit,
        onFocusToggle: () -> Unit = {},
        onStatusChange: (String) -> Unit = {},
        modifier: Modifier = Modifier
) {
    val isDone = task.status == "done"

    Card(
            modifier = modifier.fillMaxWidth().clickable { onTaskClick() },
            shape = RoundedCornerShape(16.dp),
            colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
            elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(verticalAlignment = Alignment.Top, modifier = Modifier.fillMaxWidth()) {
                // Done Checkbox button
                IconButton(onClick = onStatusToggle, modifier = Modifier.size(24.dp)) {
                    Box(
                            modifier =
                                    Modifier.size(20.dp)
                                            .clip(RoundedCornerShape(6.dp))
                                            .background(
                                                    if (isDone) StatusDone else Color.Transparent
                                            )
                                            .border(
                                                    1.5.dp,
                                                    if (isDone) StatusDone
                                                    else MaterialTheme.colorScheme.outlineVariant,
                                                    RoundedCornerShape(6.dp)
                                            ),
                            contentAlignment = Alignment.Center
                    ) {
                        if (isDone) {
                            Icon(
                                    Icons.Default.Check,
                                    contentDescription = null,
                                    tint = Color.White,
                                    modifier = Modifier.size(13.dp)
                            )
                        }
                    }
                }

                Spacer(modifier = Modifier.width(10.dp))

                Column(modifier = Modifier.weight(1f)) {
                    Row(
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.spacedBy(6.dp)
                    ) {
                        Text(
                                text = task.title,
                                fontSize = 15.sp,
                                fontWeight = FontWeight.SemiBold,
                                color =
                                        if (isDone) MaterialTheme.colorScheme.onSurfaceVariant
                                        else MaterialTheme.colorScheme.onSurface
                        )
                        if (task.isFocus) {
                            Icon(
                                    Icons.Default.Star,
                                    contentDescription = "Focus",
                                    tint = Color(0xFFF59E0B),
                                    modifier = Modifier.size(14.dp)
                            )
                        }
                    }

                    // Project & Tags
                    Row(
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.spacedBy(6.dp),
                            modifier = Modifier.padding(top = 4.dp)
                    ) {
                        if (!task.projectName.isNullOrEmpty()) {
                            ProjectPill(name = task.projectName, colorHex = task.projectColor)
                        }
                        task.tags.take(2).forEach { tag ->
                            Text(
                                    text = "#$tag",
                                    fontSize = 10.sp,
                                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                                    modifier =
                                            Modifier.clip(RoundedCornerShape(4.dp))
                                                    .background(
                                                            MaterialTheme.colorScheme.surfaceVariant
                                                    )
                                                    .padding(horizontal = 4.dp, vertical = 2.dp)
                            )
                        }
                    }

                    if (task.description.isNotBlank()) {
                        Text(
                                text = task.description,
                                fontSize = 12.sp,
                                color = MaterialTheme.colorScheme.onSurfaceVariant,
                                maxLines = 2,
                                modifier = Modifier.padding(top = 4.dp)
                        )
                    }

                    // Last Progress Note ("Inget-inget udah sampai mana")
                    if (!task.lastProgressNote.isNullOrEmpty()) {
                        Row(
                                verticalAlignment = Alignment.CenterVertically,
                                horizontalArrangement = Arrangement.spacedBy(4.dp),
                                modifier = Modifier.padding(top = 4.dp)
                        ) {
                            Icon(
                                    Icons.Default.History,
                                    contentDescription = null,
                                    tint = BrandPrimary,
                                    modifier = Modifier.size(12.dp)
                            )
                            Text(
                                    text = task.lastProgressNote,
                                    fontSize = 11.sp,
                                    fontStyle = FontStyle.Italic,
                                    color = BrandPrimary
                            )
                        }
                    }
                }

                Spacer(modifier = Modifier.width(6.dp))

                Row(
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.spacedBy(4.dp)
                ) {
                    IconButton(onClick = onFocusToggle, modifier = Modifier.size(24.dp)) {
                        Icon(
                                imageVector =
                                        if (task.isFocus) Icons.Default.Star
                                        else Icons.Default.StarBorder,
                                contentDescription = stringResource(R.string.toggle_focus),
                                tint =
                                        if (task.isFocus) Color(0xFFF59E0B)
                                        else MaterialTheme.colorScheme.outlineVariant,
                                modifier = Modifier.size(18.dp)
                        )
                    }

                    PriorityBadge(priority = task.priority)
                }
            }

            // Progress Bar (if subtasks exist)
            if (task.subtasksCount > 0) {
                Spacer(modifier = Modifier.height(10.dp))
                LinearProgressIndicator(
                        progress = { task.progressPercent / 100f },
                        modifier = Modifier.fillMaxWidth().height(5.dp).clip(CircleShape),
                        color = BrandPrimary,
                        trackColor = MaterialTheme.colorScheme.surfaceVariant
                )
                Row(
                        modifier = Modifier.fillMaxWidth().padding(top = 3.dp),
                        horizontalArrangement = Arrangement.SpaceBetween
                ) {
                    Text(
                            text =
                                    stringResource(
                                            R.string.subtasks_count,
                                            task.subtasksCompleted,
                                            task.subtasksCount
                                    ),
                            fontSize = 10.sp,
                            color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                    Text(
                            text = "${task.progressPercent}%",
                            fontSize = 10.sp,
                            fontWeight = FontWeight.Bold,
                            color = BrandPrimary
                    )
                }
            }

            Spacer(modifier = Modifier.height(10.dp))
            Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
            ) {
                // Status with quick picker dropdown & quick advance/revert buttons
                Row(
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.spacedBy(6.dp)
                ) {
                    var showStatusMenu by remember { mutableStateOf(false) }

                    Box {
                        Surface(
                                shape = RoundedCornerShape(6.dp),
                                color =
                                        when (task.status.lowercase()) {
                                            "done" -> StatusDone.copy(alpha = 0.12f)
                                            "in_progress" -> StatusInProgress.copy(alpha = 0.12f)
                                            "review" -> StatusReview.copy(alpha = 0.12f)
                                            else -> StatusTodo.copy(alpha = 0.12f)
                                        },
                                border =
                                        BorderStroke(
                                                1.dp,
                                                when (task.status.lowercase()) {
                                                    "done" -> StatusDone.copy(alpha = 0.25f)
                                                    "in_progress" ->
                                                            StatusInProgress.copy(alpha = 0.25f)
                                                    "review" -> StatusReview.copy(alpha = 0.25f)
                                                    else -> StatusTodo.copy(alpha = 0.25f)
                                                }
                                        ),
                                modifier = Modifier.clickable { showStatusMenu = true }
                        ) {
                            Row(
                                    modifier =
                                            Modifier.padding(horizontal = 7.dp, vertical = 3.5.dp),
                                    verticalAlignment = Alignment.CenterVertically,
                                    horizontalArrangement = Arrangement.spacedBy(3.dp)
                            ) {
                                Text(
                                        task.status.replace("_", " ").uppercase(),
                                        fontSize = 10.sp,
                                        fontWeight = FontWeight.Bold,
                                        color =
                                                when (task.status.lowercase()) {
                                                    "done" -> StatusDone
                                                    "in_progress" -> StatusInProgress
                                                    "review" -> StatusReview
                                                    else -> StatusTodo
                                                }
                                )
                                Icon(
                                        Icons.Default.ArrowDropDown,
                                        contentDescription = null,
                                        tint = MaterialTheme.colorScheme.onSurfaceVariant,
                                        modifier = Modifier.size(12.dp)
                                )
                            }
                        }

                        DropdownMenu(
                                expanded = showStatusMenu,
                                onDismissRequest = { showStatusMenu = false }
                        ) {
                            DropdownMenuItem(
                                    text = {
                                        Text(stringResource(R.string.filter_todo), fontSize = 13.sp)
                                    },
                                    leadingIcon = {
                                        Icon(
                                                Icons.Default.RadioButtonUnchecked,
                                                contentDescription = null,
                                                tint = StatusTodo,
                                                modifier = Modifier.size(16.dp)
                                        )
                                    },
                                    onClick = {
                                        showStatusMenu = false
                                        onStatusChange("todo")
                                    }
                            )
                            DropdownMenuItem(
                                    text = {
                                        Text(
                                                stringResource(R.string.filter_in_progress),
                                                fontSize = 13.sp
                                        )
                                    },
                                    leadingIcon = {
                                        Icon(
                                                Icons.Default.Sync,
                                                contentDescription = null,
                                                tint = StatusInProgress,
                                                modifier = Modifier.size(16.dp)
                                        )
                                    },
                                    onClick = {
                                        showStatusMenu = false
                                        onStatusChange("in_progress")
                                    }
                            )
                            DropdownMenuItem(
                                    text = {
                                        Text(
                                                stringResource(R.string.filter_review),
                                                fontSize = 13.sp
                                        )
                                    },
                                    leadingIcon = {
                                        Icon(
                                                Icons.Default.Visibility,
                                                contentDescription = null,
                                                tint = StatusReview,
                                                modifier = Modifier.size(16.dp)
                                        )
                                    },
                                    onClick = {
                                        showStatusMenu = false
                                        onStatusChange("review")
                                    }
                            )
                            DropdownMenuItem(
                                    text = {
                                        Text(stringResource(R.string.filter_done), fontSize = 13.sp)
                                    },
                                    leadingIcon = {
                                        Icon(
                                                Icons.Default.CheckCircle,
                                                contentDescription = null,
                                                tint = StatusDone,
                                                modifier = Modifier.size(16.dp)
                                        )
                                    },
                                    onClick = {
                                        showStatusMenu = false
                                        onStatusChange("done")
                                    }
                            )
                        }
                    }

                    // Quick Action Shortcut Buttons
                    when (task.status.lowercase()) {
                        "todo" -> {
                            Surface(
                                    shape = CircleShape,
                                    color = StatusInProgress.copy(alpha = 0.15f),
                                    modifier = Modifier.clickable { onStatusChange("in_progress") }
                            ) {
                                Row(
                                        modifier =
                                                Modifier.padding(
                                                        horizontal = 7.dp,
                                                        vertical = 3.dp
                                                ),
                                        verticalAlignment = Alignment.CenterVertically,
                                        horizontalArrangement = Arrangement.spacedBy(2.dp)
                                ) {
                                    Icon(
                                            Icons.Default.PlayArrow,
                                            contentDescription = null,
                                            tint = StatusInProgress,
                                            modifier = Modifier.size(10.dp)
                                    )
                                    Text(
                                            stringResource(R.string.quick_action_start),
                                            fontSize = 10.sp,
                                            fontWeight = FontWeight.Bold,
                                            color = StatusInProgress
                                    )
                                }
                            }
                        }
                        "in_progress" -> {
                            Surface(
                                    shape = CircleShape,
                                    color = StatusReview.copy(alpha = 0.15f),
                                    modifier = Modifier.clickable { onStatusChange("review") }
                            ) {
                                Row(
                                        modifier =
                                                Modifier.padding(
                                                        horizontal = 7.dp,
                                                        vertical = 3.dp
                                                ),
                                        verticalAlignment = Alignment.CenterVertically,
                                        horizontalArrangement = Arrangement.spacedBy(2.dp)
                                ) {
                                    Icon(
                                            Icons.Default.Visibility,
                                            contentDescription = null,
                                            tint = StatusReview,
                                            modifier = Modifier.size(10.dp)
                                    )
                                    Text(
                                            stringResource(R.string.quick_action_review),
                                            fontSize = 10.sp,
                                            fontWeight = FontWeight.Bold,
                                            color = StatusReview
                                    )
                                }
                            }
                            Surface(
                                    shape = CircleShape,
                                    color = StatusDone.copy(alpha = 0.15f),
                                    modifier = Modifier.clickable { onStatusChange("done") }
                            ) {
                                Row(
                                        modifier =
                                                Modifier.padding(
                                                        horizontal = 7.dp,
                                                        vertical = 3.dp
                                                ),
                                        verticalAlignment = Alignment.CenterVertically,
                                        horizontalArrangement = Arrangement.spacedBy(2.dp)
                                ) {
                                    Icon(
                                            Icons.Default.Check,
                                            contentDescription = null,
                                            tint = StatusDone,
                                            modifier = Modifier.size(10.dp)
                                    )
                                    Text(
                                            stringResource(R.string.quick_action_done),
                                            fontSize = 10.sp,
                                            fontWeight = FontWeight.Bold,
                                            color = StatusDone
                                    )
                                }
                            }
                        }
                        "review" -> {
                            Surface(
                                    shape = CircleShape,
                                    color = StatusInProgress.copy(alpha = 0.15f),
                                    modifier = Modifier.clickable { onStatusChange("in_progress") }
                            ) {
                                Row(
                                        modifier =
                                                Modifier.padding(
                                                        horizontal = 7.dp,
                                                        vertical = 3.dp
                                                ),
                                        verticalAlignment = Alignment.CenterVertically,
                                        horizontalArrangement = Arrangement.spacedBy(2.dp)
                                ) {
                                    Icon(
                                            Icons.Default.Undo,
                                            contentDescription = null,
                                            tint = StatusInProgress,
                                            modifier = Modifier.size(10.dp)
                                    )
                                    Text(
                                            stringResource(R.string.quick_action_revert),
                                            fontSize = 10.sp,
                                            fontWeight = FontWeight.Bold,
                                            color = StatusInProgress
                                    )
                                }
                            }
                            Surface(
                                    shape = CircleShape,
                                    color = StatusDone.copy(alpha = 0.15f),
                                    modifier = Modifier.clickable { onStatusChange("done") }
                            ) {
                                Row(
                                        modifier =
                                                Modifier.padding(
                                                        horizontal = 7.dp,
                                                        vertical = 3.dp
                                                ),
                                        verticalAlignment = Alignment.CenterVertically,
                                        horizontalArrangement = Arrangement.spacedBy(2.dp)
                                ) {
                                    Icon(
                                            Icons.Default.Check,
                                            contentDescription = null,
                                            tint = StatusDone,
                                            modifier = Modifier.size(10.dp)
                                    )
                                    Text(
                                            stringResource(R.string.quick_action_done),
                                            fontSize = 10.sp,
                                            fontWeight = FontWeight.Bold,
                                            color = StatusDone
                                    )
                                }
                            }
                        }
                        "done" -> {
                            Surface(
                                    shape = CircleShape,
                                    color = StatusInProgress.copy(alpha = 0.15f),
                                    modifier = Modifier.clickable { onStatusChange("in_progress") }
                            ) {
                                Row(
                                        modifier =
                                                Modifier.padding(
                                                        horizontal = 7.dp,
                                                        vertical = 3.dp
                                                ),
                                        verticalAlignment = Alignment.CenterVertically,
                                        horizontalArrangement = Arrangement.spacedBy(2.dp)
                                ) {
                                    Icon(
                                            Icons.Default.Refresh,
                                            contentDescription = null,
                                            tint = StatusInProgress,
                                            modifier = Modifier.size(10.dp)
                                    )
                                    Text(
                                            stringResource(R.string.quick_action_redo),
                                            fontSize = 10.sp,
                                            fontWeight = FontWeight.Bold,
                                            color = StatusInProgress
                                    )
                                }
                            }
                        }
                    }
                }

                Row(
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    if (!task.dueDate.isNullOrEmpty()) {
                        Row(
                                verticalAlignment = Alignment.CenterVertically,
                                horizontalArrangement = Arrangement.spacedBy(4.dp)
                        ) {
                            Icon(
                                    Icons.Default.CalendarToday,
                                    contentDescription = null,
                                    tint = PriorityUrgent,
                                    modifier = Modifier.size(11.dp)
                            )
                            Text(
                                    text = task.dueDate.take(10),
                                    fontSize = 11.sp,
                                    color = PriorityUrgent
                            )
                        }
                    }

                    Surface(
                            shape = RoundedCornerShape(8.dp),
                            color = BrandPrimary.copy(alpha = 0.12f),
                            modifier = Modifier.clickable { onTaskClick() }
                    ) {
                        Row(
                                modifier = Modifier.padding(horizontal = 8.dp, vertical = 3.dp),
                                verticalAlignment = Alignment.CenterVertically,
                                horizontalArrangement = Arrangement.spacedBy(4.dp)
                        ) {
                            Icon(
                                    Icons.Default.History,
                                    contentDescription = null,
                                    tint = BrandPrimary,
                                    modifier = Modifier.size(11.dp)
                            )
                            Text(
                                    stringResource(R.string.tab_activity_logs),
                                    fontSize = 10.sp,
                                    fontWeight = FontWeight.SemiBold,
                                    color = BrandPrimary
                            )
                        }
                    }
                }
            }
        }
    }
}

@Composable
fun TaskItemSkeleton() {
    val shimmerColors =
            listOf(
                    MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.5f),
                    MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.2f),
                    MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.5f)
            )
    val transition = rememberInfiniteTransition(label = "shimmer")
    val translateAnim by
            transition.animateFloat(
                    initialValue = 0f,
                    targetValue = 1000f,
                    animationSpec =
                            infiniteRepeatable(
                                    animation = tween(durationMillis = 1000, easing = LinearEasing),
                                    repeatMode = RepeatMode.Restart
                            ),
                    label = "shimmer_anim"
            )
    val brush =
            Brush.linearGradient(
                    colors = shimmerColors,
                    start = Offset.Zero,
                    end = Offset(translateAnim, translateAnim)
            )

    Card(
            modifier = Modifier.fillMaxWidth(),
            shape = RoundedCornerShape(16.dp),
            colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
            border = BorderStroke(1.dp, MaterialTheme.colorScheme.outlineVariant.copy(alpha = 0.4f))
    ) {
        Column(modifier = Modifier.padding(14.dp)) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Box(
                        modifier =
                                Modifier.size(20.dp)
                                        .clip(RoundedCornerShape(6.dp))
                                        .background(brush)
                )
                Spacer(modifier = Modifier.width(10.dp))
                Column(
                        modifier = Modifier.weight(1f),
                        verticalArrangement = Arrangement.spacedBy(6.dp)
                ) {
                    Box(
                            modifier =
                                    Modifier.fillMaxWidth(0.65f)
                                            .height(14.dp)
                                            .clip(RoundedCornerShape(4.dp))
                                            .background(brush)
                    )
                    Box(
                            modifier =
                                    Modifier.fillMaxWidth(0.35f)
                                            .height(10.dp)
                                            .clip(RoundedCornerShape(4.dp))
                                            .background(brush)
                    )
                }
            }
        }
    }
}

@Composable
fun TaskListSkeleton(count: Int = 6) {
    Column(verticalArrangement = Arrangement.spacedBy(10.dp), modifier = Modifier.fillMaxSize()) {
        repeat(count) { TaskItemSkeleton() }
    }
}

@Composable
fun TaskFocusHorizonBanner(
        focusTasks: List<Task>,
        onTaskClick: (Task) -> Unit,
        onFocusToggle: (Task) -> Unit,
        modifier: Modifier = Modifier
) {
    if (focusTasks.isEmpty()) return

    Card(
            modifier = modifier.fillMaxWidth(),
            shape = RoundedCornerShape(16.dp),
            colors = CardDefaults.cardColors(containerColor = Color(0xFFFFFBEB)),
            border = BorderStroke(1.dp, Color(0xFFFDE68A))
    ) {
        Column(modifier = Modifier.padding(14.dp)) {
            Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
            ) {
                Row(
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.spacedBy(6.dp)
                ) {
                    Icon(
                            Icons.Default.Star,
                            contentDescription = null,
                            tint = Color(0xFFF59E0B),
                            modifier = Modifier.size(16.dp)
                    )
                    Text(
                            stringResource(R.string.focus_horizon_title),
                            fontSize = 11.sp,
                            fontWeight = FontWeight.Bold,
                            color = Color(0xFFB45309),
                            letterSpacing = 0.5.sp
                    )
                }
                Surface(shape = CircleShape, color = Color(0xFFF59E0B).copy(alpha = 0.15f)) {
                    Text(
                            stringResource(R.string.focus_horizon_count, focusTasks.size),
                            fontSize = 11.sp,
                            fontWeight = FontWeight.Bold,
                            color = Color(0xFFB45309),
                            modifier = Modifier.padding(horizontal = 8.dp, vertical = 2.dp)
                    )
                }
            }

            Spacer(modifier = Modifier.height(10.dp))

            Row(
                    modifier = Modifier.fillMaxWidth().horizontalScroll(rememberScrollState()),
                    horizontalArrangement = Arrangement.spacedBy(10.dp)
            ) {
                focusTasks.forEach { task ->
                    Surface(
                            modifier =
                                    Modifier.width(220.dp)
                                            .clip(RoundedCornerShape(12.dp))
                                            .clickable { onTaskClick(task) },
                            color = MaterialTheme.colorScheme.surface,
                            shape = RoundedCornerShape(12.dp),
                            border = BorderStroke(1.dp, Color(0xFFFDE68A)),
                            shadowElevation = 1.dp
                    ) {
                        Column(modifier = Modifier.padding(10.dp)) {
                            Row(
                                    modifier = Modifier.fillMaxWidth(),
                                    horizontalArrangement = Arrangement.SpaceBetween,
                                    verticalAlignment = Alignment.Top
                            ) {
                                Text(
                                        text = task.title,
                                        fontSize = 13.sp,
                                        fontWeight = FontWeight.SemiBold,
                                        maxLines = 1,
                                        modifier = Modifier.weight(1f)
                                )
                                IconButton(
                                        onClick = { onFocusToggle(task) },
                                        modifier = Modifier.size(20.dp)
                                ) {
                                    Icon(
                                            Icons.Default.Star,
                                            contentDescription =
                                                    stringResource(R.string.remove_focus),
                                            tint = Color(0xFFF59E0B),
                                            modifier = Modifier.size(14.dp)
                                    )
                                }
                            }

                            Spacer(modifier = Modifier.height(4.dp))

                            Row(
                                    modifier = Modifier.fillMaxWidth(),
                                    horizontalArrangement = Arrangement.SpaceBetween,
                                    verticalAlignment = Alignment.CenterVertically
                            ) {
                                if (!task.projectName.isNullOrEmpty()) {
                                    ProjectPill(
                                            name = task.projectName,
                                            colorHex = task.projectColor
                                    )
                                } else {
                                    Spacer(modifier = Modifier.width(1.dp))
                                }
                                PriorityBadge(priority = task.priority)
                            }
                        }
                    }
                }
            }
        }
    }
}
