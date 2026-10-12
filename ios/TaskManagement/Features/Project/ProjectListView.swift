import SwiftUI

public let PROJECT_PRESET_COLORS: [String] = [
    "#6366f1", // Indigo
    "#8b5cf6", // Violet
    "#ec4899", // Pink
    "#f43f5e", // Rose
    "#f59e0b", // Amber
    "#10b981", // Emerald
    "#06b6d4", // Cyan
    "#3b82f6"  // Blue
]

@Observable
public final class ProjectViewModel {
    public var projects: [Project] = []
    public var tasks: [TaskItem] = []
    public var isLoading = false
    public var errorMessage: String?
    public var searchQuery: String = ""
    public var expandedProjectIds: Set<Int64> = []

    private let apiClient = APIClient.shared

    public init() {
        Task { await loadData() }
    }

    @MainActor
    public func loadData() async {
        isLoading = true
        errorMessage = nil
        do {
            async let fetchedProjects = apiClient.getProjects()
            async let fetchedTasks = apiClient.getTasks()
            self.projects = try await fetchedProjects
            self.tasks = try await fetchedTasks
        } catch {
            self.errorMessage = error.localizedDescription
        }
        self.isLoading = false
    }

    @MainActor
    public func loadProjects() async {
        await loadData()
    }

    public var filteredProjects: [Project] {
        let query = searchQuery.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
        guard !query.isEmpty else { return projects }
        return projects.filter {
            $0.name.lowercased().contains(query) ||
            $0.description.lowercased().contains(query)
        }
    }

    // Project Task Metrics matching Web
    public var projectTasks: [TaskItem] {
        tasks.filter { ($0.projectId ?? 0) > 0 }
    }

    public var activeProjectTasksCount: Int {
        projectTasks.filter { $0.status != "done" }.count
    }

    public var completedProjectTasksCount: Int {
        projectTasks.filter { $0.status == "done" }.count
    }

    public var completionRatePercent: Int {
        guard !projectTasks.isEmpty else { return 0 }
        return Int(round(Double(completedProjectTasksCount) / Double(projectTasks.count) * 100))
    }

    public func tasks(for projectId: Int64) -> [TaskItem] {
        tasks.filter { $0.projectId == projectId }
    }

    public func toggleExpand(projectId: Int64) {
        if expandedProjectIds.contains(projectId) {
            expandedProjectIds.remove(projectId)
        } else {
            expandedProjectIds.insert(projectId)
        }
    }

    public func expandAll() {
        expandedProjectIds = Set(projects.map { $0.id })
    }

    public func collapseAll() {
        expandedProjectIds.removeAll()
    }

    @MainActor
    public func createProject(name: String, color: String, description: String) async -> Bool {
        do {
            _ = try await apiClient.createProject(req: CreateProjectRequest(name: name, color: color, description: description))
            await loadData()
            return true
        } catch {
            self.errorMessage = error.localizedDescription
            return false
        }
    }

    @MainActor
    public func updateProject(id: Int64, name: String, color: String, description: String) async -> Bool {
        do {
            _ = try await apiClient.updateProject(id: id, req: UpdateProjectRequest(name: name, color: color, description: description))
            await loadData()
            return true
        } catch {
            self.errorMessage = error.localizedDescription
            return false
        }
    }

    @MainActor
    public func deleteProject(id: Int64) async -> Bool {
        do {
            try await apiClient.deleteProject(id: id)
            expandedProjectIds.remove(id)
            await loadData()
            return true
        } catch {
            self.errorMessage = error.localizedDescription
            return false
        }
    }

    @MainActor
    public func toggleTaskStatus(task: TaskItem) async {
        let newStatus = task.status == "done" ? "todo" : "done"
        do {
            _ = try await apiClient.updateTask(
                id: task.id,
                req: UpdateTaskRequest(status: newStatus)
            )
            await loadData()
        } catch {
            self.errorMessage = error.localizedDescription
        }
    }
}

public struct ProjectListView: View {
    @Bindable var viewModel: ProjectViewModel
    @Bindable var taskViewModel: TaskViewModel

    @State private var showProjectForm = false
    @State private var editingProject: Project? = nil
    @State private var projectToDelete: Project? = nil
    @State private var showDeleteConfirmation = false

    // Quick add task to project
    @State private var projectForNewTask: Project? = nil
    @State private var selectedTaskForDetail: TaskItem? = nil

    public init(viewModel: ProjectViewModel, taskViewModel: TaskViewModel? = nil) {
        self.viewModel = viewModel
        self.taskViewModel = taskViewModel ?? TaskViewModel()
    }

    public var body: some View {
        NavigationStack {
            ScrollView {
                VStack(spacing: 16) {
                    // Metrics Summary Header (matching Web)
                    ProjectMetricsHeader(viewModel: viewModel)

                    // Search & Actions Bar
                    ProjectSearchBar(
                        searchQuery: $viewModel.searchQuery,
                        allExpanded: !viewModel.projects.isEmpty && viewModel.expandedProjectIds.count == viewModel.projects.count,
                        onToggleExpandAll: {
                            if viewModel.expandedProjectIds.count == viewModel.projects.count {
                                viewModel.collapseAll()
                            } else {
                                viewModel.expandAll()
                            }
                        },
                        onAddProject: {
                            editingProject = nil
                            showProjectForm = true
                        }
                    )

                    // Content State
                    if viewModel.isLoading && viewModel.projects.isEmpty {
                        VStack(spacing: 12) {
                            ProgressView()
                                .scaleEffect(1.2)
                            Text("Memuat daftar proyek…")
                                .font(.system(size: 13))
                                .foregroundColor(.secondary)
                        }
                        .frame(maxWidth: .infinity, minHeight: 200)
                    } else if viewModel.filteredProjects.isEmpty {
                        VStack(spacing: 14) {
                            ZStack {
                                Circle()
                                    .fill(AppColors.primary.opacity(0.12))
                                    .frame(width: 72, height: 72)
                                Image(systemName: "folder.badge.plus")
                                    .font(.system(size: 32))
                                    .foregroundColor(AppColors.primary)
                            }
                            Text(viewModel.searchQuery.isEmpty ? "Belum Ada Proyek" : "Tidak Ada Hasil")
                                .font(.system(size: 17, weight: .bold))
                            Text(viewModel.searchQuery.isEmpty
                                 ? "Kelompokkan tugas ke dalam proyek untuk memudahkan pelacakan progres."
                                 : "Tidak ada proyek yang sesuai dengan kata kunci \"\(viewModel.searchQuery)\".")
                                .font(.system(size: 13))
                                .foregroundColor(.secondary)
                                .multilineTextAlignment(.center)
                                .padding(.horizontal, 32)

                            if viewModel.searchQuery.isEmpty {
                                Button(action: {
                                    editingProject = nil
                                    showProjectForm = true
                                }) {
                                    HStack(spacing: 6) {
                                        Image(systemName: "plus")
                                        Text("Buat Proyek Pertama")
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
                        .frame(maxWidth: .infinity, minHeight: 240)
                    } else {
                        // Project Accordion Items
                        LazyVStack(spacing: 12) {
                            ForEach(viewModel.filteredProjects) { project in
                                ProjectAccordionCard(
                                    project: project,
                                    tasks: viewModel.tasks(for: project.id),
                                    isExpanded: viewModel.expandedProjectIds.contains(project.id),
                                    onToggleExpand: {
                                        withAnimation(.easeInOut(duration: 0.25)) {
                                            viewModel.toggleExpand(projectId: project.id)
                                        }
                                    },
                                    onEdit: {
                                        editingProject = project
                                        showProjectForm = true
                                    },
                                    onDelete: {
                                        projectToDelete = project
                                        showDeleteConfirmation = true
                                    },
                                    onToggleTaskStatus: { task in
                                        Task {
                                            await viewModel.toggleTaskStatus(task: task)
                                            await taskViewModel.loadTasks()
                                        }
                                    },
                                    onSelectTask: { task in
                                        selectedTaskForDetail = task
                                    },
                                    onAddTask: {
                                        projectForNewTask = project
                                    }
                                )
                            }
                        }
                    }
                }
                .padding(.horizontal, 16)
                .padding(.vertical, 12)
            }
            .background(Color(.systemGroupedBackground))
            .navigationTitle("Proyek & Modul")
            .toolbar {
                ToolbarItem(placement: .primaryAction) {
                    Button(action: {
                        editingProject = nil
                        showProjectForm = true
                    }) {
                        HStack(spacing: 4) {
                            Image(systemName: "plus.circle.fill")
                                .font(.system(size: 18))
                            Text("Proyek")
                                .font(.system(size: 14, weight: .semibold))
                        }
                        .foregroundStyle(AppColors.brandGradient)
                    }
                }
            }
            .refreshable {
                await viewModel.loadData()
                await taskViewModel.loadTasks()
            }
            .sheet(isPresented: $showProjectForm) {
                ProjectFormSheet(
                    editingProject: editingProject,
                    onSave: { name, color, desc in
                        Task {
                            if let proj = editingProject {
                                _ = await viewModel.updateProject(id: proj.id, name: name, color: color, description: desc)
                            } else {
                                _ = await viewModel.createProject(name: name, color: color, description: desc)
                            }
                            await taskViewModel.loadProjects()
                        }
                    }
                )
            }
            .sheet(item: $projectForNewTask) { project in
                CreateTaskSheet(
                    viewModel: taskViewModel,
                    initialProjectId: project.id,
                    onCreated: {
                        Task {
                            await viewModel.loadData()
                        }
                    }
                )
            }
            .sheet(item: $selectedTaskForDetail) { task in
                TaskDetailSheet(task: task, viewModel: taskViewModel)
            }
            .confirmationDialog(
                "Hapus Proyek?",
                isPresented: $showDeleteConfirmation,
                titleVisibility: .visible,
                presenting: projectToDelete
            ) { proj in
                Button("Hapus Proyek \"\(proj.name)\"", role: .destructive) {
                    Task {
                        _ = await viewModel.deleteProject(id: proj.id)
                        await taskViewModel.loadProjects()
                    }
                }
                Button("Batal", role: .cancel) {}
            } message: { proj in
                Text("Tugas yang terhubung dengan proyek ini tidak akan dihapus, tetapi relasi proyek akan dilepas.")
            }
        }
    }
}

// MARK: - Metrics Summary Header
private struct ProjectMetricsHeader: View {
    let viewModel: ProjectViewModel

    var body: some View {
        ScrollView(.horizontal, showsIndicators: false) {
            HStack(spacing: 10) {
                MetricCard(
                    title: "Total Proyek",
                    value: "\(viewModel.projects.count)",
                    subtitle: "Proyek Aktif",
                    icon: "folder.fill",
                    tint: AppColors.primary
                )

                MetricCard(
                    title: "Tugas Proyek",
                    value: "\(viewModel.projectTasks.count)",
                    subtitle: "\(viewModel.activeProjectTasksCount) berjalan",
                    icon: "checklist",
                    tint: AppColors.statusInProgress
                )

                MetricCard(
                    title: "Selesai",
                    value: "\(viewModel.completedProjectTasksCount)",
                    subtitle: "Tugas tuntas",
                    icon: "checkmark.circle.fill",
                    tint: AppColors.statusDone
                )

                MetricCard(
                    title: "Tingkat Selesai",
                    value: "\(viewModel.completionRatePercent)%",
                    subtitle: "Rasio progres",
                    icon: "chart.pie.fill",
                    tint: AppColors.violet
                )
            }
            .padding(.vertical, 2)
        }
    }
}

private struct MetricCard: View {
    let title: String
    let value: String
    let subtitle: String
    let icon: String
    let tint: Color

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            HStack {
                Text(title)
                    .font(.system(size: 11, weight: .medium))
                    .foregroundColor(.secondary)
                Spacer()
                Image(systemName: icon)
                    .font(.system(size: 13))
                    .foregroundColor(tint)
            }

            Text(value)
                .font(.system(size: 20, weight: .bold, design: .rounded))
                .foregroundColor(.primary)

            Text(subtitle)
                .font(.system(size: 10))
                .foregroundColor(.secondary)
        }
        .padding(12)
        .frame(width: 130)
        .background(Color(.secondarySystemGroupedBackground))
        .clipShape(RoundedRectangle(cornerRadius: 14, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: 14, style: .continuous)
                .stroke(Color.primary.opacity(0.06), lineWidth: 1)
        )
    }
}

// MARK: - Search & Action Bar
private struct ProjectSearchBar: View {
    @Binding var searchQuery: String
    let allExpanded: Bool
    let onToggleExpandAll: () -> Void
    let onAddProject: () -> Void

    var body: some View {
        HStack(spacing: 8) {
            // Search Input Field
            HStack(spacing: 6) {
                Image(systemName: "magnifyingglass")
                    .font(.system(size: 13))
                    .foregroundColor(.secondary)
                TextField("Cari proyek…", text: $searchQuery)
                    .font(.system(size: 13))
                if !searchQuery.isEmpty {
                    Button(action: { searchQuery = "" }) {
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

            // Expand / Collapse All
            Button(action: onToggleExpandAll) {
                Image(systemName: allExpanded ? "arrow.up.and.down.circle.fill" : "arrow.up.and.down.circle")
                    .font(.system(size: 16))
                    .foregroundColor(.secondary)
                    .padding(8)
                    .background(Color(.secondarySystemGroupedBackground))
                    .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))
            }
            .buttonStyle(.plain)
        }
    }
}

// MARK: - Project Accordion Card
private struct ProjectAccordionCard: View {
    let project: Project
    let tasks: [TaskItem]
    let isExpanded: Bool
    let onToggleExpand: () -> Void
    let onEdit: () -> Void
    let onDelete: () -> Void
    let onToggleTaskStatus: (TaskItem) -> Void
    let onSelectTask: (TaskItem) -> Void
    let onAddTask: () -> Void

    @State private var selectedStatusTab: String? = nil

    private var projectColor: Color {
        Color(hex: project.color) ?? AppColors.primary
    }

    private var totalTasks: Int {
        tasks.count
    }

    private var completedTasks: Int {
        tasks.filter { $0.status == "done" }.count
    }

    private var percentDone: Int {
        guard totalTasks > 0 else { return 0 }
        return Int(round(Double(completedTasks) / Double(totalTasks) * 100))
    }

    private var filteredTasks: [TaskItem] {
        guard let tab = selectedStatusTab else { return tasks }
        return tasks.filter { $0.status == tab }
    }

    var body: some View {
        VStack(spacing: 0) {
            // Header Bar (Tappable)
            Button(action: onToggleExpand) {
                VStack(alignment: .leading, spacing: 8) {
                    HStack(alignment: .center, spacing: 12) {
                        // Project Color Indicator
                        Circle()
                            .fill(projectColor)
                            .frame(width: 14, height: 14)
                            .overlay(
                                Circle()
                                    .stroke(projectColor.opacity(0.3), lineWidth: 3)
                            )

                        // Title & Description
                        VStack(alignment: .leading, spacing: 2) {
                            Text(project.name)
                                .font(.system(size: 16, weight: .bold))
                                .foregroundColor(.primary)
                            if !project.description.isEmpty {
                                Text(project.description)
                                    .font(.system(size: 12))
                                    .foregroundColor(.secondary)
                                    .lineLimit(1)
                            }
                        }

                        Spacer()

                        // Action Menu (Edit / Delete)
                        Menu {
                            Button(action: onEdit) {
                                Label("Edit Proyek", systemImage: "pencil")
                            }
                            Button(role: .destructive, action: onDelete) {
                                Label("Hapus Proyek", systemImage: "trash")
                            }
                        } label: {
                            Image(systemName: "ellipsis")
                                .font(.system(size: 14, weight: .semibold))
                                .foregroundColor(.secondary)
                                .padding(6)
                                .contentShape(Rectangle())
                        }
                        .buttonStyle(.plain)

                        // Chevron indicator
                        Image(systemName: "chevron.right")
                            .font(.system(size: 12, weight: .semibold))
                            .foregroundColor(.secondary)
                            .rotationEffect(.degrees(isExpanded ? 90 : 0))
                    }

                    // Progress Bar & Stats
                    VStack(spacing: 4) {
                        GeometryReader { geo in
                            ZStack(alignment: .leading) {
                                Capsule()
                                    .fill(Color.secondary.opacity(0.15))
                                    .frame(height: 6)

                                Capsule()
                                    .fill(projectColor)
                                    .frame(width: geo.size.width * CGFloat(Double(percentDone) / 100.0), height: 6)
                            }
                        }
                        .frame(height: 6)

                        HStack {
                            Text("\(totalTasks) tugas (\(completedTasks) selesai)")
                                .font(.system(size: 11, weight: .medium))
                                .foregroundColor(.secondary)
                            Spacer()
                            Text("\(percentDone)%")
                                .font(.system(size: 11, weight: .bold))
                                .foregroundColor(projectColor)
                        }
                    }
                }
                .padding(14)
            }
            .buttonStyle(.plain)

            // Expanded Tasks Section
            if isExpanded {
                Divider()
                    .padding(.horizontal, 14)

                VStack(spacing: 12) {
                    // Status Filter Tabs
                    ScrollView(.horizontal, showsIndicators: false) {
                        HStack(spacing: 6) {
                            StatusFilterPill(
                                title: "Semua (\(totalTasks))",
                                isSelected: selectedStatusTab == nil
                            ) {
                                selectedStatusTab = nil
                            }

                            StatusFilterPill(
                                title: "To Do (\(tasks.filter { $0.status == "todo" }.count))",
                                isSelected: selectedStatusTab == "todo"
                            ) {
                                selectedStatusTab = "todo"
                            }

                            StatusFilterPill(
                                title: "In Progress (\(tasks.filter { $0.status == "in_progress" }.count))",
                                isSelected: selectedStatusTab == "in_progress"
                            ) {
                                selectedStatusTab = "in_progress"
                            }

                            StatusFilterPill(
                                title: "Review (\(tasks.filter { $0.status == "review" }.count))",
                                isSelected: selectedStatusTab == "review"
                            ) {
                                selectedStatusTab = "review"
                            }

                            StatusFilterPill(
                                title: "Selesai (\(completedTasks))",
                                isSelected: selectedStatusTab == "done"
                            ) {
                                selectedStatusTab = "done"
                            }
                        }
                        .padding(.horizontal, 14)
                        .padding(.top, 10)
                    }

                    // Tasks List
                    if filteredTasks.isEmpty {
                        VStack(spacing: 8) {
                            Text("Belum ada tugas di kategori ini")
                                .font(.system(size: 12))
                                .foregroundColor(.secondary)
                        }
                        .padding(.vertical, 14)
                    } else {
                        VStack(spacing: 8) {
                            ForEach(filteredTasks) { task in
                                ProjectTaskRow(
                                    task: task,
                                    onToggle: { onToggleTaskStatus(task) },
                                    onSelect: { onSelectTask(task) }
                                )
                            }
                        }
                        .padding(.horizontal, 14)
                    }

                    // Quick "+ Tambah Tugas di Proyek Ini" Button
                    Button(action: onAddTask) {
                        HStack(spacing: 6) {
                            Image(systemName: "plus.circle.fill")
                                .font(.system(size: 14))
                            Text("Tambah Tugas di Proyek Ini")
                                .font(.system(size: 13, weight: .semibold))
                        }
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 10)
                        .background(projectColor.opacity(0.12))
                        .foregroundColor(projectColor)
                        .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))
                    }
                    .buttonStyle(.plain)
                    .padding(.horizontal, 14)
                    .padding(.bottom, 14)
                }
                .background(Color(.tertiarySystemGroupedBackground).opacity(0.5))
            }
        }
        .background(Color(.secondarySystemGroupedBackground))
        .clipShape(RoundedRectangle(cornerRadius: 16, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: 16, style: .continuous)
                .stroke(Color.primary.opacity(0.06), lineWidth: 1)
        )
    }
}

private struct StatusFilterPill: View {
    let title: String
    let isSelected: Bool
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            Text(title)
                .font(.system(size: 11, weight: isSelected ? .bold : .medium))
                .padding(.horizontal, 10)
                .padding(.vertical, 5)
                .background(isSelected ? AppColors.primary : Color.secondary.opacity(0.12))
                .foregroundColor(isSelected ? .white : .secondary)
                .clipShape(Capsule())
        }
        .buttonStyle(.plain)
    }
}

// MARK: - Task Row Inside Project Accordion
private struct ProjectTaskRow: View {
    let task: TaskItem
    let onToggle: () -> Void
    let onSelect: () -> Void

    var body: some View {
        HStack(alignment: .center, spacing: 10) {
            // Checkbox Button
            Button(action: onToggle) {
                Image(systemName: task.status == "done" ? "checkmark.circle.fill" : "circle")
                    .font(.system(size: 18))
                    .foregroundColor(task.status == "done" ? AppColors.statusDone : .secondary.opacity(0.6))
            }
            .buttonStyle(.plain)

            // Content
            VStack(alignment: .leading, spacing: 3) {
                HStack(spacing: 6) {
                    Text(task.title)
                        .font(.system(size: 14, weight: .medium))
                        .foregroundColor(task.status == "done" ? .secondary : .primary)
                        .strikethrough(task.status == "done")
                        .lineLimit(1)

                    if task.isFocus {
                        Image(systemName: "star.fill")
                            .font(.system(size: 10))
                            .foregroundColor(.yellow)
                    }
                }

                HStack(spacing: 6) {
                    PriorityBadge(priority: task.priority)
                    StatusBadge(status: task.status)

                    if let due = task.dueDate {
                        HStack(spacing: 3) {
                            Image(systemName: "calendar")
                                .font(.system(size: 9))
                            Text(due.prefix(10))
                                .font(.system(size: 10))
                        }
                        .foregroundColor(AppColors.priorityHigh)
                    }

                    if let subCount = task.subtasksCount, subCount > 0 {
                        Text("\(task.subtasksCompleted ?? 0)/\(subCount)")
                            .font(.system(size: 10, weight: .semibold))
                            .foregroundColor(.secondary)
                    }
                }
            }

            Spacer()

            // Chevron to view details
            Button(action: onSelect) {
                Image(systemName: "chevron.right")
                    .font(.system(size: 11, weight: .semibold))
                    .foregroundColor(.secondary.opacity(0.6))
                    .padding(4)
            }
            .buttonStyle(.plain)
        }
        .padding(.horizontal, 10)
        .padding(.vertical, 8)
        .background(Color(.secondarySystemGroupedBackground))
        .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))
        .contentShape(Rectangle())
        .onTapGesture {
            onSelect()
        }
    }
}

// MARK: - Project Form Sheet (Create & Edit)
public struct ProjectFormSheet: View {
    @Environment(\.dismiss) private var dismiss
    let editingProject: Project?
    let onSave: (String, String, String) -> Void

    @State private var name: String = ""
    @State private var description: String = ""
    @State private var color: String = PROJECT_PRESET_COLORS[0]
    @State private var isSubmitting = false

    public init(
        editingProject: Project? = nil,
        onSave: @escaping (String, String, String) -> Void
    ) {
        self.editingProject = editingProject
        self.onSave = onSave
    }

    public var body: some View {
        NavigationStack {
            Form {
                Section(header: Text("Informasi Proyek")) {
                    TextField("Nama Proyek *", text: $name)
                        .font(.system(size: 15))
                    TextField("Deskripsi (Opsional)", text: $description, axis: .vertical)
                        .font(.system(size: 14))
                        .lineLimit(2...4)
                }

                Section(header: Text("Pilihan Warna")) {
                    VStack(alignment: .leading, spacing: 10) {
                        Text("Pilih warna aksen untuk modul/proyek ini:")
                            .font(.system(size: 12))
                            .foregroundColor(.secondary)

                        LazyVGrid(columns: Array(repeating: GridItem(.flexible(), spacing: 12), count: 4), spacing: 14) {
                            ForEach(PROJECT_PRESET_COLORS, id: \.self) { hex in
                                let swatchColor = Color(hex: hex) ?? AppColors.primary
                                let isSelected = color.lowercased() == hex.lowercased()

                                Button(action: {
                                    color = hex
                                }) {
                                    ZStack {
                                        Circle()
                                            .fill(swatchColor)
                                            .frame(width: 44, height: 44)
                                            .shadow(color: swatchColor.opacity(isSelected ? 0.4 : 0.1), radius: 4, x: 0, y: 2)

                                        if isSelected {
                                            Circle()
                                                .stroke(Color.white, lineWidth: 3)
                                                .frame(width: 38, height: 38)
                                            Image(systemName: "checkmark")
                                                .font(.system(size: 15, weight: .bold))
                                                .foregroundColor(.white)
                                        }
                                    }
                                }
                                .buttonStyle(.plain)
                            }
                        }
                        .padding(.vertical, 6)
                    }
                    .padding(.vertical, 4)
                }
            }
            .navigationTitle(editingProject == nil ? "Buat Proyek Baru" : "Edit Proyek")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Batal") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button(action: handleSave) {
                        if isSubmitting {
                            ProgressView()
                                .scaleEffect(0.8)
                        } else {
                            Text("Simpan")
                                .fontWeight(.semibold)
                        }
                    }
                    .disabled(name.trimmingCharacters(in: .whitespaces).isEmpty || isSubmitting)
                }
            }
            .onAppear {
                if let proj = editingProject {
                    name = proj.name
                    description = proj.description
                    color = proj.color
                }
            }
        }
    }

    private func handleSave() {
        let trimmed = name.trimmingCharacters(in: .whitespaces)
        guard !trimmed.isEmpty else { return }
        isSubmitting = true
        onSave(trimmed, color, description.trimmingCharacters(in: .whitespaces))
        dismiss()
    }
}
