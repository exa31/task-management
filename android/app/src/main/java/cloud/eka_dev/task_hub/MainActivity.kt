package cloud.eka_dev.task_hub

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.annotation.StringRes
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.navigation.NavGraph.Companion.findStartDestination
import androidx.navigation.compose.*
import cloud.eka_dev.task_hub.core.designsystem.TaskManagementTheme
import cloud.eka_dev.task_hub.feature.auth.AuthScreen
import cloud.eka_dev.task_hub.feature.auth.AuthViewModel
import cloud.eka_dev.task_hub.feature.calendar.CalendarScreen
import cloud.eka_dev.task_hub.feature.calendar.CalendarViewModel
import cloud.eka_dev.task_hub.feature.mcp.McpScreen
import cloud.eka_dev.task_hub.feature.mcp.McpViewModel
import cloud.eka_dev.task_hub.feature.note.NoteScreen
import cloud.eka_dev.task_hub.feature.note.NoteViewModel
import cloud.eka_dev.task_hub.feature.profile.ProfileScreen
import cloud.eka_dev.task_hub.feature.profile.ProfileViewModel
import cloud.eka_dev.task_hub.feature.project.ProjectScreen
import cloud.eka_dev.task_hub.feature.project.ProjectViewModel
import cloud.eka_dev.task_hub.feature.reminder.ReminderScreen
import cloud.eka_dev.task_hub.feature.reminder.ReminderViewModel
import cloud.eka_dev.task_hub.feature.task.TaskListScreen
import cloud.eka_dev.task_hub.feature.task.TaskViewModel

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        val app = application as TaskApp

        val authViewModel = AuthViewModel(app.apiClient, app.tokenManager)
        val taskViewModel = TaskViewModel(app.apiClient)
        val projectViewModel = ProjectViewModel(app.apiClient)
        val calendarViewModel = CalendarViewModel(app.apiClient)
        val noteViewModel = NoteViewModel(app.apiClient)
        val reminderViewModel = ReminderViewModel(app.apiClient, app.sseClient)
        val profileViewModel = ProfileViewModel(app.apiClient, app.tokenManager)
        val mcpViewModel = McpViewModel(app.apiClient)

        setContent {
            TaskManagementTheme {
                val tokenState by app.tokenManager.accessTokenFlow.collectAsState(initial = "loading")

                if (tokenState == "loading") {
                    // Splash loading
                } else if (tokenState.isNullOrEmpty()) {
                    AuthScreen(
                        viewModel = authViewModel,
                        onAuthSuccess = {
                            taskViewModel.loadTasks()
                            projectViewModel.loadData()
                            calendarViewModel.loadMonthData(calendarViewModel.currentMonth.value)
                            noteViewModel.loadNotes()
                            reminderViewModel.loadReminders()
                            profileViewModel.loadProfile()
                        }
                    )
                } else {
                    MainNavigationScaffold(
                        taskViewModel = taskViewModel,
                        projectViewModel = projectViewModel,
                        calendarViewModel = calendarViewModel,
                        noteViewModel = noteViewModel,
                        reminderViewModel = reminderViewModel,
                        profileViewModel = profileViewModel,
                        mcpViewModel = mcpViewModel,
                        onSignOut = { authViewModel.signOut() }
                    )
                }
            }
        }
    }
}

sealed class Screen(val route: String, @StringRes val titleRes: Int) {
    object Tasks : Screen("tasks", R.string.nav_tasks)
    object Projects : Screen("projects", R.string.nav_projects)
    object Calendar : Screen("calendar", R.string.nav_calendar)
    object Notes : Screen("notes", R.string.nav_notes)
    object Reminders : Screen("reminders", R.string.nav_reminders)
    object Profile : Screen("profile", R.string.nav_profile)
    object Mcp : Screen("mcp", R.string.nav_mcp)
}

@Composable
fun MainNavigationScaffold(
    taskViewModel: TaskViewModel,
    projectViewModel: ProjectViewModel,
    calendarViewModel: CalendarViewModel,
    noteViewModel: NoteViewModel,
    reminderViewModel: ReminderViewModel,
    profileViewModel: ProfileViewModel,
    mcpViewModel: McpViewModel,
    onSignOut: () -> Unit
) {
    val navController = rememberNavController()
    val navBackStackEntry by navController.currentBackStackEntryAsState()
    val currentRoute = navBackStackEntry?.destination?.route

    Scaffold(
        bottomBar = {
            if (currentRoute != Screen.Mcp.route) {
                NavigationBar {
                    NavigationBarItem(
                        selected = currentRoute == Screen.Tasks.route,
                        onClick = {
                            navController.navigate(Screen.Tasks.route) {
                                popUpTo(navController.graph.findStartDestination().id) { saveState = true }
                                launchSingleTop = true
                                restoreState = true
                            }
                        },
                        icon = { Icon(Icons.Default.Checklist, contentDescription = null) },
                        label = { Text(stringResource(Screen.Tasks.titleRes)) }
                    )
                    NavigationBarItem(
                        selected = currentRoute == Screen.Projects.route,
                        onClick = {
                            navController.navigate(Screen.Projects.route) {
                                popUpTo(navController.graph.findStartDestination().id) { saveState = true }
                                launchSingleTop = true
                                restoreState = true
                            }
                        },
                        icon = { Icon(Icons.Default.Folder, contentDescription = null) },
                        label = { Text(stringResource(Screen.Projects.titleRes)) }
                    )
                    NavigationBarItem(
                        selected = currentRoute == Screen.Calendar.route,
                        onClick = {
                            navController.navigate(Screen.Calendar.route) {
                                popUpTo(navController.graph.findStartDestination().id) { saveState = true }
                                launchSingleTop = true
                                restoreState = true
                            }
                        },
                        icon = { Icon(Icons.Default.CalendarMonth, contentDescription = null) },
                        label = { Text(stringResource(Screen.Calendar.titleRes)) }
                    )
                    NavigationBarItem(
                        selected = currentRoute == Screen.Notes.route,
                        onClick = {
                            navController.navigate(Screen.Notes.route) {
                                popUpTo(navController.graph.findStartDestination().id) { saveState = true }
                                launchSingleTop = true
                                restoreState = true
                            }
                        },
                        icon = { Icon(Icons.Default.EditNote, contentDescription = null) },
                        label = { Text(stringResource(Screen.Notes.titleRes)) }
                    )
                    NavigationBarItem(
                        selected = currentRoute == Screen.Reminders.route,
                        onClick = {
                            navController.navigate(Screen.Reminders.route) {
                                popUpTo(navController.graph.findStartDestination().id) { saveState = true }
                                launchSingleTop = true
                                restoreState = true
                            }
                        },
                        icon = { Icon(Icons.Default.Notifications, contentDescription = null) },
                        label = { Text(stringResource(Screen.Reminders.titleRes)) }
                    )
                    NavigationBarItem(
                        selected = currentRoute == Screen.Profile.route,
                        onClick = {
                            navController.navigate(Screen.Profile.route) {
                                popUpTo(navController.graph.findStartDestination().id) { saveState = true }
                                launchSingleTop = true
                                restoreState = true
                            }
                        },
                        icon = { Icon(Icons.Default.Person, contentDescription = null) },
                        label = { Text(stringResource(Screen.Profile.titleRes)) }
                    )
                }
            }
        }
    ) { padding ->
        NavHost(
            navController = navController,
            startDestination = Screen.Tasks.route,
            modifier = Modifier.padding(padding)
        ) {
            composable(Screen.Tasks.route) {
                TaskListScreen(
                    viewModel = taskViewModel,
                    onTaskClick = { /* detail */ }
                )
            }
            composable(Screen.Projects.route) {
                ProjectScreen(viewModel = projectViewModel)
            }
            composable(Screen.Calendar.route) {
                CalendarScreen(viewModel = calendarViewModel)
            }
            composable(Screen.Notes.route) {
                NoteScreen(viewModel = noteViewModel)
            }
            composable(Screen.Reminders.route) {
                ReminderScreen(viewModel = reminderViewModel)
            }
            composable(Screen.Profile.route) {
                ProfileScreen(
                    viewModel = profileViewModel,
                    onNavigateToMcp = { navController.navigate(Screen.Mcp.route) },
                    onSignOut = onSignOut
                )
            }
            composable(Screen.Mcp.route) {
                McpScreen(
                    viewModel = mcpViewModel,
                    onBack = { navController.popBackStack() }
                )
            }
        }
    }
}
