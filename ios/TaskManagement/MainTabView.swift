import SwiftUI

public struct MainTabView: View {
    let authViewModel: AuthViewModel
    @State private var taskViewModel = TaskViewModel()
    @State private var projectViewModel = ProjectViewModel()
    @State private var noteViewModel = NoteViewModel()
    @State private var reminderViewModel = ReminderViewModel()
    @State private var selectedTab: Int = 0

    public init(authViewModel: AuthViewModel) {
        self.authViewModel = authViewModel
    }

    public var body: some View {
        TabView(selection: $selectedTab) {
            TaskListView(viewModel: taskViewModel)
                .tabItem {
                    Label(L10n.Nav.tasks, systemImage: "checkmark.rectangle.stack.fill")
                }
                .tag(0)

            ProjectListView(viewModel: projectViewModel, taskViewModel: taskViewModel)
                .tabItem {
                    Label(L10n.Nav.projects, systemImage: "folder.fill")
                }
                .tag(1)

            CalendarView()
                .tabItem {
                    Label(L10n.Nav.calendar, systemImage: "calendar")
                }
                .tag(2)

            NoteListView(viewModel: noteViewModel)
                .tabItem {
                    Label(L10n.Nav.notes, systemImage: "note.text")
                }
                .tag(3)

            ReminderListView(viewModel: reminderViewModel)
                .tabItem {
                    Label(L10n.Nav.reminders, systemImage: "bell.badge.fill")
                }
                .tag(4)

            ProfileView(onSignOut: {
                authViewModel.signOut()
            })
            .tabItem {
                Label(L10n.Nav.settings, systemImage: "gearshape.fill")
            }
            .tag(5)
        }
        .tint(AppColors.primary)
        .onOpenURL { url in
            handleDeepLink(url)
        }
    }

    private func handleDeepLink(_ url: URL) {
        guard url.scheme == "taskhub" else { return }
        AppLogger.info("Received deep link in MainTabView: \(url.absoluteString)", category: .task)

        var targetTaskId: Int64?
        let host = url.host?.lowercased()
        if host == "settings" || host == "profile" || host == "widget" || host == "widget-settings" {
            selectedTab = 5
            return
        }

        if host == "tasks" || host == "task" {
            let path = url.path.trimmingCharacters(in: CharacterSet(charactersIn: "/"))
            if let id = Int64(path) {
                targetTaskId = id
            } else if let components = URLComponents(url: url, resolvingAgainstBaseURL: false),
                      let queryItems = components.queryItems,
                      let idStr = queryItems.first(where: { $0.name == "id" })?.value,
                      let id = Int64(idStr) {
                targetTaskId = id
            }
        }

        if let taskId = targetTaskId {
            selectedTab = 0
            Task { @MainActor in
                await taskViewModel.selectTask(byId: taskId)
            }
        }
    }
}
