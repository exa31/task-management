import SwiftUI

public struct MainTabView: View {
    let authViewModel: AuthViewModel
    @State private var taskViewModel = TaskViewModel()
    @State private var projectViewModel = ProjectViewModel()
    @State private var noteViewModel = NoteViewModel()
    @State private var reminderViewModel = ReminderViewModel()

    public init(authViewModel: AuthViewModel) {
        self.authViewModel = authViewModel
    }

    public var body: some View {
        TabView {
            TaskListView(viewModel: taskViewModel)
                .tabItem {
                    Label(L10n.Nav.tasks, systemImage: "checkmark.rectangle.stack.fill")
                }

            ProjectListView(viewModel: projectViewModel, taskViewModel: taskViewModel)
                .tabItem {
                    Label(L10n.Nav.projects, systemImage: "folder.fill")
                }

            CalendarView()
                .tabItem {
                    Label(L10n.Nav.calendar, systemImage: "calendar")
                }

            NoteListView(viewModel: noteViewModel)
                .tabItem {
                    Label(L10n.Nav.notes, systemImage: "note.text")
                }

            ReminderListView(viewModel: reminderViewModel)
                .tabItem {
                    Label(L10n.Nav.reminders, systemImage: "bell.badge.fill")
                }

            ProfileView(onSignOut: {
                authViewModel.signOut()
            })
            .tabItem {
                Label(L10n.Nav.settings, systemImage: "gearshape.fill")
            }
        }
        .tint(AppColors.primary)
    }
}
