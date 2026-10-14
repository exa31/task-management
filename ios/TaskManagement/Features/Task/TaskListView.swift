import SwiftUI

public struct TaskListView: View {
    @Bindable var viewModel: TaskViewModel
    @State private var showCreateSheet = false

    public init(viewModel: TaskViewModel) {
        self.viewModel = viewModel
    }

    public var body: some View {
        NavigationStack {
            VStack(spacing: 0) {
                // Search Input Field
                HStack(spacing: 8) {
                    HStack(spacing: 6) {
                        Image(systemName: "magnifyingglass")
                            .font(.system(size: 13))
                            .foregroundColor(.secondary)
                        TextField("Cari tugas, tag, atau deskripsi…", text: $viewModel.searchQuery)
                            .font(.system(size: 13))
                        if !viewModel.searchQuery.isEmpty {
                            Button(action: { viewModel.searchQuery = "" }) {
                                Image(systemName: "xmark.circle.fill")
                                    .font(.system(size: 14))
                                    .foregroundColor(.secondary)
                            }
                            .buttonStyle(.plain)
                        }
                    }
                    .padding(.horizontal, 10)
                    .padding(.vertical, 8)
                    .background(Color(.secondarySystemGroupedBackground))
                    .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))
                    .overlay(
                        RoundedRectangle(cornerRadius: 10, style: .continuous)
                            .stroke(Color.primary.opacity(0.06), lineWidth: 1)
                    )
                }
                .padding(.horizontal, 16)
                .padding(.top, 8)
                .padding(.bottom, 6)

                // Secondary Filter Controls: Project, Priority, Focus, and Reset
                ScrollView(.horizontal, showsIndicators: false) {
                    HStack(spacing: 8) {
                        // Project Filter Menu
                        Menu {
                            Button("Semua Proyek") {
                                viewModel.selectedProjectId = nil
                            }
                            Divider()
                            ForEach(viewModel.projects) { proj in
                                Button(action: { viewModel.selectedProjectId = proj.id }) {
                                    HStack {
                                        Text(proj.name)
                                        if viewModel.selectedProjectId == proj.id {
                                            Image(systemName: "checkmark")
                                        }
                                    }
                                }
                            }
                        } label: {
                            HStack(spacing: 4) {
                                Image(systemName: "folder.fill")
                                    .font(.system(size: 11))
                                if let projId = viewModel.selectedProjectId,
                                   let proj = viewModel.projects.first(where: { $0.id == projId }) {
                                    Text(proj.name)
                                        .font(.system(size: 12, weight: .bold))
                                } else {
                                    Text("Proyek")
                                        .font(.system(size: 12, weight: .medium))
                                }
                                Image(systemName: "chevron.down")
                                    .font(.system(size: 9))
                            }
                            .padding(.horizontal, 10)
                            .padding(.vertical, 6)
                            .background(
                                viewModel.selectedProjectId != nil
                                    ? AppColors.primary.opacity(0.15)
                                    : Color(.secondarySystemGroupedBackground)
                            )
                            .foregroundColor(
                                viewModel.selectedProjectId != nil
                                    ? AppColors.primary
                                    : .secondary
                            )
                            .clipShape(Capsule())
                            .overlay(
                                Capsule()
                                    .stroke(
                                        viewModel.selectedProjectId != nil
                                            ? AppColors.primary.opacity(0.3)
                                            : Color.secondary.opacity(0.15),
                                        lineWidth: 1
                                    )
                            )
                        }
                        .buttonStyle(.plain)

                        // Priority Filter Menu
                        Menu {
                            Button("Semua Prioritas") {
                                viewModel.selectedPriority = nil
                            }
                            Divider()
                            Button("Urgent") { viewModel.selectedPriority = "urgent" }
                            Button("High") { viewModel.selectedPriority = "high" }
                            Button("Medium") { viewModel.selectedPriority = "medium" }
                            Button("Low") { viewModel.selectedPriority = "low" }
                        } label: {
                            HStack(spacing: 4) {
                                Image(systemName: "flag.fill")
                                    .font(.system(size: 11))
                                if let priority = viewModel.selectedPriority {
                                    Text(priority.uppercased())
                                        .font(.system(size: 12, weight: .bold))
                                } else {
                                    Text("Prioritas")
                                        .font(.system(size: 12, weight: .medium))
                                }
                                Image(systemName: "chevron.down")
                                    .font(.system(size: 9))
                            }
                            .padding(.horizontal, 10)
                            .padding(.vertical, 6)
                            .background(
                                viewModel.selectedPriority != nil
                                    ? AppColors.statusInProgress.opacity(0.15)
                                    : Color(.secondarySystemGroupedBackground)
                            )
                            .foregroundColor(
                                viewModel.selectedPriority != nil
                                    ? AppColors.statusInProgress
                                    : .secondary
                            )
                            .clipShape(Capsule())
                            .overlay(
                                Capsule()
                                    .stroke(
                                        viewModel.selectedPriority != nil
                                            ? AppColors.statusInProgress.opacity(0.3)
                                            : Color.secondary.opacity(0.15),
                                        lineWidth: 1
                                    )
                            )
                        }
                        .buttonStyle(.plain)

                        // Focus Filter Toggle
                        Button(action: {
                            viewModel.filterFocusOnly.toggle()
                        }) {
                            HStack(spacing: 4) {
                                Image(systemName: "star.fill")
                                    .font(.system(size: 11))
                                Text(viewModel.focusTasks.count > 0 ? "Fokus (\(viewModel.focusTasks.count))" : "Fokus")
                                    .font(.system(size: 12, weight: viewModel.filterFocusOnly ? .bold : .medium))
                            }
                            .padding(.horizontal, 10)
                            .padding(.vertical, 6)
                            .background(
                                viewModel.filterFocusOnly
                                    ? Color.yellow.opacity(0.2)
                                    : Color(.secondarySystemGroupedBackground)
                            )
                            .foregroundColor(
                                viewModel.filterFocusOnly
                                    ? .orange
                                    : .secondary
                            )
                            .clipShape(Capsule())
                            .overlay(
                                Capsule()
                                    .stroke(
                                        viewModel.filterFocusOnly
                                            ? Color.orange.opacity(0.4)
                                            : Color.secondary.opacity(0.15),
                                        lineWidth: 1
                                    )
                            )
                        }
                        .buttonStyle(.plain)

                        // Clear / Reset All Filters Button
                        if viewModel.hasActiveFilters {
                            Button(action: {
                                withAnimation {
                                    viewModel.resetFilters()
                                }
                            }) {
                                HStack(spacing: 4) {
                                    Image(systemName: "xmark.circle.fill")
                                        .font(.system(size: 11))
                                    Text("Reset (\(viewModel.activeFiltersCount))")
                                        .font(.system(size: 12, weight: .bold))
                                }
                                .padding(.horizontal, 10)
                                .padding(.vertical, 6)
                                .background(AppColors.priorityUrgent.opacity(0.12))
                                .foregroundColor(AppColors.priorityUrgent)
                                .clipShape(Capsule())
                            }
                            .buttonStyle(.plain)
                        }
                    }
                    .padding(.horizontal, 16)
                    .padding(.vertical, 4)
                }

                // Filter Segment Pills for Status (Semua, To Do, In Progress, Review, Selesai)
                ScrollView(.horizontal, showsIndicators: false) {
                    HStack(spacing: 8) {
                        FilterButton(
                            title: "Semua (\(viewModel.tasks.count))",
                            icon: "tray.fill",
                            isSelected: viewModel.selectedStatus == nil
                        ) {
                            viewModel.setStatusFilter(nil)
                        }
                        FilterButton(
                            title: "To Do (\(viewModel.tasks.filter { $0.status == "todo" }.count))",
                            icon: "circle",
                            isSelected: viewModel.selectedStatus == "todo"
                        ) {
                            viewModel.setStatusFilter("todo")
                        }
                        FilterButton(
                            title: "In Progress (\(viewModel.tasks.filter { $0.status == "in_progress" }.count))",
                            icon: "arrow.triangle.2.circlepath",
                            isSelected: viewModel.selectedStatus == "in_progress"
                        ) {
                            viewModel.setStatusFilter("in_progress")
                        }
                        FilterButton(
                            title: "Review (\(viewModel.tasks.filter { $0.status == "review" }.count))",
                            icon: "eye.fill",
                            isSelected: viewModel.selectedStatus == "review"
                        ) {
                            viewModel.setStatusFilter("review")
                        }
                        FilterButton(
                            title: "Selesai (\(viewModel.tasks.filter { $0.status == "done" }.count))",
                            icon: "checkmark.circle.fill",
                            isSelected: viewModel.selectedStatus == "done"
                        ) {
                            viewModel.setStatusFilter("done")
                        }
                    }
                    .padding(.horizontal, 16)
                    .padding(.vertical, 8)
                }

                // Today's Focus Horizon Banner
                if !viewModel.focusTasks.isEmpty && !viewModel.filterFocusOnly {
                    TaskFocusHorizonBanner(
                        focusTasks: viewModel.focusTasks,
                        onComplete: { task in
                            viewModel.toggleTaskStatus(task: task)
                        },
                        onSelect: { task in
                            viewModel.selectedTaskForDetail = task
                        }
                    )
                    .padding(.horizontal, 16)
                    .padding(.bottom, 6)
                }

                // Task List Content
                if viewModel.isLoading && viewModel.tasks.isEmpty {
                    Spacer()
                    VStack(spacing: 12) {
                        ProgressView()
                            .scaleEffect(1.2)
                        Text("Memuat daftar tugas…")
                            .font(.system(size: 13))
                            .foregroundColor(.secondary)
                    }
                    Spacer()
                } else if viewModel.filteredTasks.isEmpty {
                    Spacer()
                    VStack(spacing: 14) {
                        ZStack {
                            Circle()
                                .fill(AppColors.primary.opacity(0.12))
                                .frame(width: 72, height: 72)
                            Image(systemName: viewModel.hasActiveFilters ? "line.3.horizontal.decrease.circle" : "checklist")
                                .font(.system(size: 32))
                                .foregroundColor(AppColors.primary)
                        }

                        VStack(spacing: 4) {
                            Text(viewModel.hasActiveFilters ? L10n.Tasks.noTasksFilterTitle : L10n.Tasks.emptyTitle)
                                .font(.system(size: 17, weight: .bold))
                            Text(viewModel.hasActiveFilters
                                 ? L10n.Tasks.noTasksFilterDesc
                                 : L10n.Tasks.emptyDesc)
                                .font(.system(size: 13))
                                .foregroundColor(.secondary)
                                .multilineTextAlignment(.center)
                                .padding(.horizontal, 32)
                        }

                        if viewModel.hasActiveFilters {
                            Button(action: {
                                withAnimation {
                                    viewModel.resetFilters()
                                }
                            }) {
                                HStack(spacing: 6) {
                                    Image(systemName: "arrow.counterclockwise")
                                    Text(L10n.Tasks.resetAllFilters)
                                }
                                .font(.system(size: 14, weight: .semibold))
                                .padding(.horizontal, 16)
                                .padding(.vertical, 10)
                                .background(AppColors.brandGradient)
                                .foregroundColor(.white)
                                .clipShape(Capsule())
                            }
                            .padding(.top, 4)
                        } else {
                            Button(action: { showCreateSheet = true }) {
                                HStack(spacing: 6) {
                                    Image(systemName: "plus")
                                    Text(L10n.Tasks.createTask)
                                }
                                .font(.system(size: 14, weight: .semibold))
                                .padding(.horizontal, 16)
                                .padding(.vertical, 10)
                                .background(AppColors.brandGradient)
                                .foregroundColor(.white)
                                .clipShape(Capsule())
                            }
                            .padding(.top, 4)
                        }
                    }
                    Spacer()
                } else {
                    List {
                        ForEach(viewModel.filteredTasks) { task in
                            TaskRowView(
                                task: task,
                                onToggle: { viewModel.toggleTaskStatus(task: task) },
                                onSubtaskToggle: { subtaskId, isDone in
                                    viewModel.toggleSubtask(taskId: task.id, subtaskId: subtaskId, currentCompleted: isDone)
                                },
                                onSelect: {
                                    viewModel.selectedTaskForDetail = task
                                }
                            )
                            .swipeActions(edge: .leading) {
                                Button {
                                    viewModel.toggleFocus(task: task)
                                } label: {
                                    Label(task.isFocus ? L10n.Tasks.unfocusToggle : L10n.Tasks.focusToggle, systemImage: task.isFocus ? "star.slash.fill" : "star.fill")
                                }
                                .tint(.orange)
                            }
                            .swipeActions(edge: .trailing) {
                                Button(role: .destructive) {
                                    Task { await viewModel.deleteTask(id: task.id) }
                                } label: {
                                    Label(L10n.General.delete, systemImage: "trash")
                                }
                            }
                            .listRowSeparator(.hidden)
                            .listRowBackground(Color.clear)
                            .listRowInsets(EdgeInsets(top: 6, leading: 16, bottom: 6, trailing: 16))
                        }
                    }
                    .listStyle(.plain)
                    .refreshable {
                        await viewModel.loadTasks()
                        await viewModel.loadProjects()
                    }
                }
            }
            .navigationTitle(L10n.Tasks.title)
            .toolbar {
                ToolbarItem(placement: .navigationBarLeading) {
                    BrandLogoView(size: 26)
                }

                ToolbarItem(placement: .primaryAction) {
                    Button(action: { showCreateSheet = true }) {
                        Image(systemName: "plus.circle.fill")
                            .font(.system(size: 20))
                            .foregroundStyle(AppColors.brandGradient)
                    }
                }
            }
            .sheet(isPresented: $showCreateSheet) {
                CreateTaskSheet(viewModel: viewModel)
            }
            .sheet(item: $viewModel.selectedTaskForDetail) { task in
                TaskDetailSheet(task: task, viewModel: viewModel)
            }
        }
    }
}

// MARK: - Today's Focus Horizon Banner
struct TaskFocusHorizonBanner: View {
    let focusTasks: [TaskItem]
    let onComplete: (TaskItem) -> Void
    let onSelect: (TaskItem) -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack(spacing: 6) {
                Image(systemName: "star.fill")
                    .font(.system(size: 12))
                    .foregroundColor(.orange)
                Text(L10n.Tasks.focusToday)
                    .font(.system(size: 13, weight: .bold))
                    .foregroundColor(.primary)
                Text(L10n.Tasks.activeCount(count: focusTasks.count))
                    .font(.system(size: 10, weight: .bold))
                    .padding(.horizontal, 6)
                    .padding(.vertical, 2)
                    .background(Color.orange.opacity(0.15))
                    .foregroundColor(.orange)
                    .clipShape(Capsule())
                Spacer()
            }

            ScrollView(.horizontal, showsIndicators: false) {
                HStack(spacing: 10) {
                    ForEach(focusTasks) { task in
                        HStack(spacing: 8) {
                            Button(action: { onComplete(task) }) {
                                Image(systemName: task.status == "done" ? "checkmark.circle.fill" : "circle")
                                    .font(.system(size: 16))
                                    .foregroundColor(task.status == "done" ? AppColors.statusDone : .secondary)
                            }
                            .buttonStyle(.plain)

                            VStack(alignment: .leading, spacing: 2) {
                                Text(task.title)
                                    .font(.system(size: 13, weight: .semibold))
                                    .lineLimit(1)
                                HStack(spacing: 4) {
                                    if let pName = task.projectName {
                                        Text(pName)
                                            .font(.system(size: 10))
                                            .foregroundColor(.secondary)
                                    }
                                    PriorityBadge(priority: task.priority)
                                }
                            }
                        }
                        .padding(.horizontal, 10)
                        .padding(.vertical, 8)
                        .background(Color(.secondarySystemGroupedBackground))
                        .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
                        .overlay(
                            RoundedRectangle(cornerRadius: 12, style: .continuous)
                                .stroke(Color.orange.opacity(0.2), lineWidth: 1)
                        )
                        .contentShape(Rectangle())
                        .onTapGesture {
                            onSelect(task)
                        }
                    }
                }
            }
        }
        .padding(12)
        .background(Color(.secondarySystemGroupedBackground).opacity(0.6))
        .clipShape(RoundedRectangle(cornerRadius: 14, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: 14, style: .continuous)
                .stroke(Color.primary.opacity(0.06), lineWidth: 1)
        )
    }
}

struct FilterButton: View {
    let title: String
    let icon: String
    let isSelected: Bool
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            HStack(spacing: 5) {
                Image(systemName: icon)
                    .font(.system(size: 11, weight: .semibold))
                Text(title)
                    .font(.system(size: 13, weight: .semibold))
            }
            .padding(.horizontal, 12)
            .padding(.vertical, 7)
            .background(
                isSelected ?
                    AnyShapeStyle(AppColors.brandGradient) :
                    AnyShapeStyle(Color(.secondarySystemGroupedBackground))
            )
            .foregroundColor(isSelected ? .white : .primary)
            .clipShape(Capsule())
            .overlay(
                Capsule()
                    .stroke(isSelected ? Color.clear : Color.secondary.opacity(0.15), lineWidth: 1)
            )
            .shadow(color: isSelected ? AppColors.indigo.opacity(0.3) : Color.clear, radius: 4, x: 0, y: 2)
        }
        .buttonStyle(.plain)
    }
}

struct TaskRowView: View {
    let task: TaskItem
    let onToggle: () -> Void
    let onSubtaskToggle: (Int64, Bool) -> Void
    let onSelect: () -> Void

    private var isDone: Bool {
        task.status == "done"
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            // Header: Checkbox, Title, Project Pill, Priority
            HStack(alignment: .top, spacing: 10) {
                // Done Checkbox button
                Button(action: onToggle) {
                    ZStack {
                        RoundedRectangle(cornerRadius: 6, style: .continuous)
                            .stroke(isDone ? AppColors.statusDone : Color.secondary.opacity(0.35), lineWidth: 1.5)
                            .frame(width: 20, height: 20)
                            .background(
                                RoundedRectangle(cornerRadius: 6, style: .continuous)
                                    .fill(isDone ? AppColors.statusDone : Color.clear)
                            )

                        if isDone {
                            Image(systemName: "checkmark")
                                .font(.system(size: 11, weight: .bold))
                                .foregroundColor(.white)
                        }
                    }
                }
                .buttonStyle(.plain)
                .padding(.top, 2)

                // Title & Subtitles
                VStack(alignment: .leading, spacing: 4) {
                    HStack(spacing: 6) {
                        Text(task.title)
                            .font(.system(size: 15, weight: .semibold))
                            .strikethrough(isDone, color: .secondary)
                            .foregroundColor(isDone ? .secondary : .primary)

                        if task.isFocus {
                            Image(systemName: "star.fill")
                                .font(.system(size: 11))
                                .foregroundColor(.yellow)
                        }
                    }

                    // Project Pill & Tags
                    HStack(spacing: 6) {
                        if let projName = task.projectName {
                            ProjectPill(name: projName, hexColor: task.projectColor)
                        }

                        if task.isRecurring == true {
                            HStack(spacing: 3) {
                                Image(systemName: "repeat")
                                    .font(.system(size: 9, weight: .bold))
                                Text(task.recurrencePattern?.capitalized ?? L10n.Tasks.repeatBadge)
                                    .font(.system(size: 10, weight: .semibold))
                            }
                            .padding(.horizontal, 6)
                            .padding(.vertical, 2)
                            .background(Color.purple.opacity(0.12))
                            .foregroundColor(.purple)
                            .clipShape(Capsule())
                        }

                        if let tags = task.tags, !tags.isEmpty {
                            ForEach(tags.prefix(2), id: \.self) { tag in
                                Text("#\(tag)")
                                    .font(.system(size: 10, design: .monospaced))
                                    .foregroundColor(.secondary)
                                    .padding(.horizontal, 4)
                                    .padding(.vertical, 2)
                                    .background(Color.secondary.opacity(0.1))
                                    .clipShape(RoundedRectangle(cornerRadius: 4))
                            }
                        }
                    }

                    // Description
                    if !task.description.isEmpty {
                        Text(task.description)
                            .font(.system(size: 12))
                            .foregroundColor(.secondary)
                            .lineLimit(2)
                            .padding(.top, 2)
                    }

                    // Last Progress Note ("Inget-inget udah sampai mana")
                    if let progressNote = task.lastProgressNote, !progressNote.isEmpty {
                        HStack(spacing: 4) {
                            Image(systemName: "clock.arrow.circlepath")
                                .font(.system(size: 10))
                            Text(progressNote)
                                .font(.system(size: 11, weight: .medium))
                                .italic()
                        }
                        .foregroundColor(AppColors.primary)
                        .padding(.top, 2)
                    }
                }

                Spacer()

                PriorityBadge(priority: task.priority)
            }

            // Subtasks checklist preview
            if let subtasks = task.subtasks, !subtasks.isEmpty {
                VStack(alignment: .leading, spacing: 6) {
                    Divider()
                        .padding(.vertical, 2)

                    ForEach(subtasks.prefix(3)) { subtask in
                        HStack(spacing: 8) {
                            Button(action: { onSubtaskToggle(subtask.id, subtask.isCompleted) }) {
                                Image(systemName: subtask.isCompleted ? "checkmark.square.fill" : "square")
                                    .font(.system(size: 13))
                                    .foregroundColor(subtask.isCompleted ? AppColors.statusDone : .secondary)
                            }
                            .buttonStyle(.plain)

                            Text(subtask.title)
                                .font(.system(size: 12))
                                .strikethrough(subtask.isCompleted)
                                .foregroundColor(subtask.isCompleted ? .secondary : .primary)
                            Spacer()
                        }
                    }
                }
            }

            // Progress Bar (if subtasks exist)
            if let count = task.subtasksCount, count > 0 {
                VStack(spacing: 4) {
                    ProgressBar(percent: task.progressPercent ?? 0)
                    HStack {
                        Text("\(task.subtasksCompleted ?? 0)/\(count) checklist")
                            .font(.system(size: 10, weight: .medium))
                            .foregroundColor(.secondary)
                        Spacer()
                        Text("\(task.progressPercent ?? 0)%")
                            .font(.system(size: 10, weight: .bold))
                            .foregroundColor(AppColors.primary)
                    }
                }
                .padding(.top, 2)
            }

            // Footer: Status Badge, Due Date, and Log History Button
            HStack(spacing: 8) {
                StatusBadge(status: task.status)
                Spacer()
                if let due = task.dueDate {
                    HStack(spacing: 4) {
                        Image(systemName: "calendar")
                            .font(.system(size: 10))
                        Text(due.prefix(10))
                            .font(.system(size: 11))
                    }
                    .foregroundColor(AppColors.priorityHigh)
                }

                Button(action: onSelect) {
                    HStack(spacing: 4) {
                        Image(systemName: "clock.arrow.circlepath")
                            .font(.system(size: 10, weight: .semibold))
                        Text("Riwayat Log")
                            .font(.system(size: 11, weight: .semibold))
                    }
                    .padding(.horizontal, 9)
                    .padding(.vertical, 4)
                    .background(AppColors.primary.opacity(0.12))
                    .foregroundColor(AppColors.primary)
                    .clipShape(Capsule())
                }
                .buttonStyle(.plain)
            }
            .padding(.top, 4)
        }
        .contentShape(Rectangle())
        .onTapGesture {
            onSelect()
        }
        .glassCardStyle()
    }
}

public struct CreateTaskSheet: View {
    @Environment(\.dismiss) private var dismiss
    @Bindable var viewModel: TaskViewModel
    public var initialProjectId: Int64? = nil
    public var onCreated: (() -> Void)? = nil

    @State private var title: String = ""
    @State private var description: String = ""
    @State private var status: String = "todo"
    @State private var priority: String = "medium"
    @State private var selectedProjectId: Int64? = nil
    @State private var hasDueDate: Bool = false
    @State private var dueDate: Date = Date().addingTimeInterval(3600 * 24)
    @State private var isFocus: Bool = false
    @State private var isRecurring: Bool = false
    @State private var recurrencePattern: String = "daily"
    @State private var recurrenceInterval: Int = 1
    @State private var tagsText: String = ""
    @State private var subtasks: [String] = []
    @State private var newSubtaskText: String = ""
    @State private var isSubmitting: Bool = false
    @State private var errorMessage: String? = nil

    public init(
        viewModel: TaskViewModel,
        initialProjectId: Int64? = nil,
        onCreated: (() -> Void)? = nil
    ) {
        self.viewModel = viewModel
        self.initialProjectId = initialProjectId
        self.onCreated = onCreated
        _selectedProjectId = State(initialValue: initialProjectId)
    }

    public var body: some View {
        NavigationStack {
            Form {
                if let error = errorMessage {
                    Section {
                        HStack(spacing: 8) {
                            Image(systemName: "exclamationmark.triangle.fill")
                                .foregroundColor(AppColors.priorityUrgent)
                            Text(error)
                                .font(.system(size: 13))
                                .foregroundColor(AppColors.priorityUrgent)
                        }
                    }
                }

                Section(header: Text(L10n.Tasks.sectionInfo)) {
                    TextField(L10n.Tasks.fieldTitle, text: $title)
                        .font(.system(size: 15))
                    TextField(L10n.Tasks.fieldDescription, text: $description, axis: .vertical)
                        .font(.system(size: 14))
                        .lineLimit(3...5)
                }

                Section(header: Text(L10n.Tasks.sectionStatusPriority)) {
                    Picker(L10n.Tasks.statusLabel, selection: $status) {
                        Text("Backlog").tag("backlog")
                        Text("To Do").tag("todo")
                        Text("In Progress").tag("in_progress")
                        Text("Review").tag("review")
                    }

                    Picker(L10n.Tasks.priorityLabel, selection: $priority) {
                        Text("Low").tag("low")
                        Text("Medium").tag("medium")
                        Text("High").tag("high")
                        Text("Urgent").tag("urgent")
                    }
                    .pickerStyle(.segmented)
                }

                Section(header: Text(L10n.Tasks.sectionProject)) {
                    Picker(L10n.Tasks.selectProject, selection: $selectedProjectId) {
                        Text(L10n.Tasks.noProject).tag(nil as Int64?)
                        ForEach(viewModel.projects) { proj in
                            HStack(spacing: 6) {
                                Circle()
                                    .fill(Color(hex: proj.color) ?? AppColors.primary)
                                    .frame(width: 8, height: 8)
                                Text(proj.name)
                            }
                            .tag(proj.id as Int64?)
                        }
                    }
                }

                Section(header: Text(L10n.Tasks.sectionDueDate)) {
                    Toggle(L10n.Tasks.setDueDate, isOn: $hasDueDate)
                    if hasDueDate {
                        DatePicker(
                            L10n.Tasks.dateTime,
                            selection: $dueDate,
                            displayedComponents: [.date, .hourAndMinute]
                        )
                        .datePickerStyle(.compact)
                    }
                }

                Section(header: Text(L10n.Tasks.sectionFocus)) {
                    Toggle(isOn: $isFocus) {
                        HStack(spacing: 8) {
                            Image(systemName: "star.fill")
                                .foregroundColor(isFocus ? .yellow : .secondary)
                            Text(L10n.Tasks.focusToday)
                                .font(.system(size: 15))
                        }
                    }
                }

                Section(header: Text(L10n.Tasks.sectionRecurrence)) {
                    Toggle(L10n.Tasks.repeatAuto, isOn: $isRecurring)
                    if isRecurring {
                        Picker(L10n.Tasks.repeatPattern, selection: $recurrencePattern) {
                            Text(L10n.Tasks.patternDaily).tag("daily")
                            Text(L10n.Tasks.patternWeekdays).tag("weekdays")
                            Text(L10n.Tasks.patternWeekly).tag("weekly")
                            Text(L10n.Tasks.patternMonthly).tag("monthly")
                        }
                        let unit = recurrencePattern == "monthly" ? L10n.Tasks.unitMonths : recurrencePattern == "weekly" ? L10n.Tasks.unitWeeks : L10n.Tasks.unitDays
                        Stepper(L10n.Tasks.repeatInterval(interval: recurrenceInterval, unit: unit), value: $recurrenceInterval, in: 1...365)
                    }
                }

                Section(header: HStack {
                    Text(L10n.Tasks.sectionSubtasks)
                    if !subtasks.isEmpty {
                        Text("(\(subtasks.count))")
                            .foregroundColor(AppColors.primary)
                            .fontWeight(.bold)
                    }
                }) {
                    ForEach(Array(subtasks.enumerated()), id: \.offset) { index, item in
                        HStack {
                            Image(systemName: "checkmark.circle")
                                .foregroundColor(.secondary)
                                .font(.system(size: 14))
                            Text(item)
                                .font(.system(size: 14))
                            Spacer()
                            Button(role: .destructive) {
                                subtasks.remove(at: index)
                            } label: {
                                Image(systemName: "trash")
                                    .foregroundColor(.red.opacity(0.8))
                                    .font(.system(size: 12))
                            }
                            .buttonStyle(.plain)
                        }
                    }

                    HStack {
                        TextField(L10n.Tasks.addSubtaskPlaceholder, text: $newSubtaskText)
                            .font(.system(size: 14))
                            .onSubmit {
                                addSubtask()
                            }
                        if !newSubtaskText.trimmingCharacters(in: .whitespaces).isEmpty {
                            Button(action: addSubtask) {
                                Image(systemName: "plus.circle.fill")
                                    .font(.system(size: 20))
                                    .foregroundColor(AppColors.primary)
                            }
                            .buttonStyle(.plain)
                        }
                    }
                }

                Section(header: Text(L10n.Tasks.sectionTags)) {
                    TextField(L10n.Tasks.tagsPlaceholder, text: $tagsText)
                        .font(.system(size: 14))

                    let parsedTags = parseTags(tagsText)
                    if !parsedTags.isEmpty {
                        ScrollView(.horizontal, showsIndicators: false) {
                            HStack(spacing: 6) {
                                ForEach(parsedTags, id: \.self) { tag in
                                    Text("#\(tag)")
                                        .font(.system(size: 11, weight: .semibold))
                                        .padding(.horizontal, 8)
                                        .padding(.vertical, 4)
                                        .background(AppColors.primary.opacity(0.12))
                                        .foregroundColor(AppColors.primary)
                                        .clipShape(Capsule())
                                }
                            }
                            .padding(.vertical, 2)
                        }
                    }
                }
            }
            .navigationTitle(L10n.Tasks.createTitle)
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button(L10n.General.cancel) { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button(action: handleSave) {
                        if isSubmitting {
                            ProgressView()
                                .scaleEffect(0.8)
                        } else {
                            Text(L10n.General.save)
                                .fontWeight(.semibold)
                        }
                    }
                    .disabled(title.trimmingCharacters(in: .whitespaces).isEmpty || isSubmitting)
                }
            }
            .onAppear {
                if let initial = initialProjectId {
                    selectedProjectId = initial
                }
                if viewModel.projects.isEmpty {
                    Task {
                        await viewModel.loadProjects()
                    }
                }
            }
        }
    }

    private func addSubtask() {
        let text = newSubtaskText.trimmingCharacters(in: .whitespaces)
        guard !text.isEmpty else { return }
        subtasks.append(text)
        newSubtaskText = ""
    }

    private func parseTags(_ input: String) -> [String] {
        input.components(separatedBy: ",")
            .map { $0.trimmingCharacters(in: .whitespaces).lowercased() }
            .filter { !$0.isEmpty }
    }

    private func handleSave() {
        guard !title.trimmingCharacters(in: .whitespaces).isEmpty else { return }
        isSubmitting = true
        errorMessage = nil

        let parsedTags = parseTags(tagsText)
        let isoDate: String?
        if hasDueDate {
            let formatter = ISO8601DateFormatter()
            formatter.formatOptions = [.withInternetDateTime]
            isoDate = formatter.string(from: dueDate)
        } else {
            isoDate = nil
        }

        var finalSubtasks = subtasks
        let pendingSubtask = newSubtaskText.trimmingCharacters(in: .whitespaces)
        if !pendingSubtask.isEmpty {
            finalSubtasks.append(pendingSubtask)
        }

        Task {
            let success = await viewModel.createTask(
                title: title.trimmingCharacters(in: .whitespaces),
                description: description.trimmingCharacters(in: .whitespaces),
                status: status,
                priority: priority,
                dueDate: isoDate,
                isFocus: isFocus,
                tags: parsedTags,
                projectId: selectedProjectId,
                subtasks: finalSubtasks,
                isRecurring: isRecurring,
                recurrencePattern: isRecurring ? recurrencePattern : nil,
                recurrenceInterval: isRecurring ? recurrenceInterval : nil
            )
            await MainActor.run {
                isSubmitting = false
                if success {
                    onCreated?()
                    dismiss()
                } else {
                    errorMessage = viewModel.errorMessage ?? L10n.Tasks.saveFailed
                }
            }
        }
    }
}
