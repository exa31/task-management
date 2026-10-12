package cloud.eka_dev.task_hub

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.navigation.NavGraph.Companion.findStartDestination
import androidx.navigation.compose.*
import cloud.eka_dev.task_hub.core.designsystem.TaskManagementTheme
import cloud.eka_dev.task_hub.feature.auth.AuthScreen
import cloud.eka_dev.task_hub.feature.auth.AuthViewModel
import cloud.eka_dev.task_hub.feature.note.NoteScreen
import cloud.eka_dev.task_hub.feature.note.NoteViewModel
import cloud.eka_dev.task_hub.feature.profile.ProfileScreen
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
        val noteViewModel = NoteViewModel(app.apiClient)
        val reminderViewModel = ReminderViewModel(app.apiClient, app.sseClient)

        setContent {
            TaskManagementTheme {
                val navController = rememberNavController()
                val tokenState by app.tokenManager.accessTokenFlow.collectAsState(initial = "loading")

                if (tokenState == "loading") {
                    // Splash loading
                } else if (tokenState.isNullOrEmpty()) {
                    AuthScreen(
                        viewModel = authViewModel,
                        onAuthSuccess = {
                            taskViewModel.loadTasks()
                            projectViewModel.loadProjects()
                            noteViewModel.loadNotes()
                            reminderViewModel.loadReminders()
                        }
                    )
                } else {
                    MainNavigationScaffold(
                        taskViewModel = taskViewModel,
                        projectViewModel = projectViewModel,
                        noteViewModel = noteViewModel,
                        reminderViewModel = reminderViewModel,
                        tokenManager = app.tokenManager,
                        onSignOut = { authViewModel.signOut() }
                    )
                }
            }
        }
    }
}

sealed class Screen(val route: String, val title: String) {
    object Tasks : Screen("tasks", "Tugas")
    object Projects : Screen("projects", "Proyek")
    object Notes : Screen("notes", "Catatan")
    object Reminders : Screen("reminders", "Pengingat")
    object Profile : Screen("profile", "Profil")
}

@Composable
fun MainNavigationScaffold(
    taskViewModel: TaskViewModel,
    projectViewModel: ProjectViewModel,
    noteViewModel: NoteViewModel,
    reminderViewModel: ReminderViewModel,
    tokenManager: cloud.eka_dev.task_hub.core.datastore.TokenManager,
    onSignOut: () -> Unit
) {
    val navController = rememberNavController()
    val navBackStackEntry by navController.currentBackStackEntryAsState()
    val currentRoute = navBackStackEntry?.destination?.route

    Scaffold(
        bottomBar = {
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
                    label = { Text(Screen.Tasks.title) }
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
                    label = { Text(Screen.Projects.title) }
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
                    label = { Text(Screen.Notes.title) }
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
                    label = { Text(Screen.Reminders.title) }
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
                    label = { Text(Screen.Profile.title) }
                )
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
            composable(Screen.Notes.route) {
                NoteScreen(viewModel = noteViewModel)
            }
            composable(Screen.Reminders.route) {
                ReminderScreen(viewModel = reminderViewModel)
            }
            composable(Screen.Profile.route) {
                ProfileScreen(tokenManager = tokenManager, onSignOut = onSignOut)
            }
        }
    }
}
