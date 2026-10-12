import SwiftUI

public struct TaskDetailSheet: View {
    let task: TaskItem
    @Bindable var viewModel: TaskViewModel
    @Environment(\.dismiss) private var dismiss

    @State private var selectedTab = 0
    @State private var newNoteText = ""
    @State private var isPostingNote = false

    public init(task: TaskItem, viewModel: TaskViewModel) {
        self.task = task
        self.viewModel = viewModel
    }

    public var body: some View {
        NavigationStack {
            VStack(spacing: 0) {
                // Task Header Banner
                VStack(alignment: .leading, spacing: 10) {
                    HStack(alignment: .top) {
                        VStack(alignment: .leading, spacing: 4) {
                            Text(task.title)
                                .font(.system(size: 18, weight: .bold))
                                .foregroundColor(.primary)

                            HStack(spacing: 8) {
                                if let projName = task.projectName {
                                    ProjectPill(name: projName, hexColor: task.projectColor)
                                }
                                PriorityBadge(priority: task.priority)
                                StatusBadge(status: task.status)
                            }
                        }

                        Spacer()

                        Button(action: { dismiss() }) {
                            Image(systemName: "xmark.circle.fill")
                                .font(.system(size: 24))
                                .foregroundColor(Color.secondary.opacity(0.6))
                        }
                    }

                    if !task.description.isEmpty {
                        Text(task.description)
                            .font(.system(size: 13))
                            .foregroundColor(.secondary)
                            .lineLimit(3)
                    }
                }
                .padding(16)
                .background(Color(.secondarySystemGroupedBackground))

                // Custom Segmented Bar
                HStack(spacing: 8) {
                    TabSegmentButton(
                        title: "Detail & Subtask",
                        icon: "checklist",
                        isSelected: selectedTab == 0
                    ) {
                        selectedTab = 0
                    }

                    TabSegmentButton(
                        title: "Riwayat Log",
                        icon: "clock.arrow.circlepath",
                        badgeCount: viewModel.taskActivities.count,
                        isSelected: selectedTab == 1
                    ) {
                        selectedTab = 1
                    }
                }
                .padding(.horizontal, 16)
                .padding(.vertical, 10)
                .background(Color(.systemGroupedBackground))

                // Content
                if selectedTab == 0 {
                    detailContentView
                } else {
                    activityTimelineView
                }
            }
            .navigationTitle("Detail Tugas")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button("Tutup") { dismiss() }
                        .font(.system(size: 14, weight: .semibold))
                }
            }
            .task {
                await viewModel.loadActivities(for: task.id)
            }
        }
    }

    // MARK: - Tab 1: Detail & Subtasks
    private var detailContentView: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                // Status Toggle Row
                VStack(alignment: .leading, spacing: 8) {
                    Text("STATUS TUGAS")
                        .font(.system(size: 11, weight: .bold))
                        .foregroundColor(.secondary)

                    HStack(spacing: 12) {
                        Button(action: { viewModel.toggleTaskStatus(task: task) }) {
                            HStack(spacing: 8) {
                                Image(systemName: task.status == "done" ? "checkmark.circle.fill" : "circle")
                                    .foregroundColor(task.status == "done" ? AppColors.statusDone : .secondary)
                                Text(task.status == "done" ? "Tandai Belum Selesai" : "Tandai Selesai")
                                    .font(.system(size: 14, weight: .medium))
                            }
                            .padding(.horizontal, 14)
                            .padding(.vertical, 10)
                            .background(Color(.secondarySystemGroupedBackground))
                            .clipShape(RoundedRectangle(cornerRadius: 10))
                        }
                    }
                }

                // Subtasks Section
                VStack(alignment: .leading, spacing: 10) {
                    HStack {
                        Text("SUBTASKS (\(task.subtasksCompleted ?? 0)/\(task.subtasksCount ?? 0))")
                            .font(.system(size: 11, weight: .bold))
                            .foregroundColor(.secondary)
                        Spacer()
                        Text("\(task.progressPercent ?? 0)%")
                            .font(.system(size: 12, weight: .bold))
                            .foregroundColor(AppColors.primary)
                    }

                    // Progress bar
                    GeometryReader { geo in
                        ZStack(alignment: .leading) {
                            Capsule().fill(Color.secondary.opacity(0.15))
                            Capsule()
                                .fill(AppColors.brandGradient)
                                .frame(width: geo.size.width * CGFloat(Double(task.progressPercent ?? 0) / 100.0))
                        }
                    }
                    .frame(height: 6)

                    if let subtasks = task.subtasks, !subtasks.isEmpty {
                        VStack(spacing: 8) {
                            ForEach(subtasks) { subtask in
                                HStack(spacing: 10) {
                                    Button(action: {
                                        viewModel.toggleSubtask(taskId: task.id, subtaskId: subtask.id, currentCompleted: subtask.isCompleted)
                                    }) {
                                        Image(systemName: subtask.isCompleted ? "checkmark.square.fill" : "square")
                                            .font(.system(size: 16))
                                            .foregroundColor(subtask.isCompleted ? AppColors.statusDone : .secondary)
                                    }
                                    .buttonStyle(.plain)

                                    Text(subtask.title)
                                        .font(.system(size: 13))
                                        .strikethrough(subtask.isCompleted)
                                        .foregroundColor(subtask.isCompleted ? .secondary : .primary)

                                    Spacer()
                                }
                                .padding(10)
                                .background(Color(.secondarySystemGroupedBackground))
                                .clipShape(RoundedRectangle(cornerRadius: 8))
                            }
                        }
                    } else {
                        Text(L10n.Tasks.noSubtasks)
                            .font(.system(size: 13))
                            .foregroundColor(.secondary)
                            .padding(.vertical, 8)
                    }
                }

                // Last Progress Note Preview
                if let lastNote = task.lastProgressNote, !lastNote.isEmpty {
                    VStack(alignment: .leading, spacing: 6) {
                        Text(L10n.Tasks.lastProgressNote)
                            .font(.system(size: 11, weight: .bold))
                            .foregroundColor(.secondary)

                        HStack(spacing: 8) {
                            Image(systemName: "quote.opening")
                                .font(.system(size: 12))
                                .foregroundColor(AppColors.primary)
                            Text(lastNote)
                                .font(.system(size: 13, weight: .medium))
                                .italic()
                        }
                        .padding(12)
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .background(AppColors.primary.opacity(0.08))
                        .clipShape(RoundedRectangle(cornerRadius: 10))
                    }
                }
            }
            .padding(16)
        }
        .background(Color(.systemGroupedBackground))
    }

    // MARK: - Tab 2: Activity Timeline (Task Logs)
    private var activityTimelineView: some View {
        VStack(spacing: 0) {
            // Post Note Input Card
            VStack(spacing: 8) {
                HStack(spacing: 8) {
                    TextField("Tulis catatan progres atau update log...", text: $newNoteText)
                        .font(.system(size: 13))
                        .padding(10)
                        .background(Color(.secondarySystemGroupedBackground))
                        .clipShape(RoundedRectangle(cornerRadius: 10))

                    Button(action: {
                        postNewNote()
                    }) {
                        HStack(spacing: 4) {
                            if isPostingNote {
                                ProgressView()
                                    .scaleEffect(0.8)
                            } else {
                                Image(systemName: "paperplane.fill")
                                    .font(.system(size: 12))
                            }
                            Text("Kirim")
                                .font(.system(size: 13, weight: .semibold))
                        }
                        .padding(.horizontal, 14)
                        .padding(.vertical, 10)
                        .background(newNoteText.trimmingCharacters(in: .whitespaces).isEmpty ? Color.secondary.opacity(0.3) : AppColors.primary)
                        .foregroundColor(.white)
                        .clipShape(RoundedRectangle(cornerRadius: 10))
                    }
                    .disabled(newNoteText.trimmingCharacters(in: .whitespaces).isEmpty || isPostingNote)
                }
            }
            .padding(12)
            .background(Color(.tertiarySystemGroupedBackground))

            // Timeline Scroll
            if viewModel.isLoadingActivities && viewModel.taskActivities.isEmpty {
                Spacer()
                ProgressView("Memuat riwayat log...")
                Spacer()
            } else if viewModel.taskActivities.isEmpty {
                Spacer()
                VStack(spacing: 12) {
                    Image(systemName: "clock.arrow.circlepath")
                        .font(.system(size: 40))
                        .foregroundColor(Color.secondary.opacity(0.5))
                    Text("Belum Ada Riwayat Log")
                        .font(.system(size: 16, weight: .semibold))
                    Text("Semua perubahan status, subtask, dan catatan progres akan tercatat di sini secara kronologis.")
                        .font(.system(size: 12))
                        .foregroundColor(.secondary)
                        .multilineTextAlignment(.center)
                        .padding(.horizontal, 32)
                }
                Spacer()
            } else {
                ScrollView {
                    LazyVStack(spacing: 0) {
                        ForEach(Array(viewModel.taskActivities.enumerated()), id: \.element.id) { index, activity in
                            ActivityTimelineItemView(
                                activity: activity,
                                isLast: index == viewModel.taskActivities.count - 1
                            )
                        }
                    }
                    .padding(.horizontal, 16)
                    .padding(.vertical, 12)
                }
            }
        }
        .background(Color(.systemGroupedBackground))
    }

    private func postNewNote() {
        let note = newNoteText
        isPostingNote = true
        Task {
            await viewModel.addActivityNote(taskId: task.id, note: note)
            newNoteText = ""
            isPostingNote = false
        }
    }
}

// MARK: - Timeline Item Row
struct ActivityTimelineItemView: View {
    let activity: TaskActivity
    let isLast: Bool

    private var badgeInfo: (icon: String, color: Color, label: String) {
        switch activity.activityType {
        case "task_created":
            return ("sparkles", Color.green, "Tugas Dibuat")
        case "task_updated":
            return ("pencil", Color.blue, "Tugas Diperbarui")
        case "progress_logged":
            return ("clock.arrow.circlepath", AppColors.indigo, "Catatan Progres")
        case "subtask_created":
            return ("plus.circle", Color.cyan, "Subtask Ditambahkan")
        case "subtask_toggled":
            return ("checkmark.circle", Color.teal, "Subtask Diperbarui")
        case "subtask_deleted":
            return ("trash", Color.red, "Subtask Dihapus")
        case "focus_toggled":
            return ("star.fill", Color.yellow, "Fokus Diubah")
        default:
            return ("bubble.left.fill", Color.secondary, "Catatan Log")
        }
    }

    var body: some View {
        HStack(alignment: .top, spacing: 14) {
            // Node circle & line
            VStack(spacing: 0) {
                ZStack {
                    Circle()
                        .fill(badgeInfo.color.opacity(0.18))
                        .frame(width: 28, height: 28)
                    Image(systemName: badgeInfo.icon)
                        .font(.system(size: 12, weight: .bold))
                        .foregroundColor(badgeInfo.color)
                }

                if !isLast {
                    Rectangle()
                        .fill(Color.secondary.opacity(0.2))
                        .frame(width: 2)
                        .frame(minHeight: 36)
                }
            }

            // Content card
            VStack(alignment: .leading, spacing: 4) {
                HStack {
                    Text(badgeInfo.label)
                        .font(.system(size: 11, weight: .bold))
                        .foregroundColor(badgeInfo.color)
                        .padding(.horizontal, 6)
                        .padding(.vertical, 2)
                        .background(badgeInfo.color.opacity(0.1))
                        .clipShape(RoundedRectangle(cornerRadius: 4))

                    Spacer()

                    Text(formatRelativeDate(activity.createdAt))
                        .font(.system(size: 11))
                        .foregroundColor(.secondary)
                }

                Text(activity.description)
                    .font(.system(size: 13))
                    .foregroundColor(.primary)
                    .padding(.top, 1)
            }
            .padding(10)
            .background(Color(.secondarySystemGroupedBackground))
            .clipShape(RoundedRectangle(cornerRadius: 10))
            .padding(.bottom, isLast ? 8 : 12)
        }
    }

    private func formatRelativeDate(_ isoString: String) -> String {
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        var date = formatter.date(from: isoString)
        if date == nil {
            formatter.formatOptions = [.withInternetDateTime]
            date = formatter.date(from: isoString)
        }

        guard let d = date else { return isoString }
        let diff = Int(Date().timeIntervalSince(d))
        if diff < 60 { return "Baru saja" }
        if diff < 3600 { return "\(diff / 60)m lalu" }
        if diff < 86400 { return "\(diff / 3600)j lalu" }
        let dayFormat = DateFormatter()
        dayFormat.dateFormat = "dd MMM HH:mm"
        return dayFormat.string(from: d)
    }
}

// MARK: - Tab Segment Button
struct TabSegmentButton: View {
    let title: String
    let icon: String
    var badgeCount: Int? = nil
    let isSelected: Bool
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            HStack(spacing: 6) {
                Image(systemName: icon)
                    .font(.system(size: 12, weight: .semibold))
                Text(title)
                    .font(.system(size: 13, weight: .semibold))

                if let count = badgeCount, count > 0 {
                    Text("\(count)")
                        .font(.system(size: 10, weight: .bold, design: .monospaced))
                        .padding(.horizontal, 5)
                        .padding(.vertical, 1)
                        .background(isSelected ? Color.white.opacity(0.25) : Color.secondary.opacity(0.15))
                        .clipShape(Capsule())
                }
            }
            .frame(maxWidth: .infinity)
            .padding(.vertical, 8)
            .background(
                isSelected ?
                    AnyShapeStyle(AppColors.brandGradient) :
                    AnyShapeStyle(Color(.secondarySystemGroupedBackground))
            )
            .foregroundColor(isSelected ? .white : .primary)
            .clipShape(RoundedRectangle(cornerRadius: 10))
            .shadow(color: isSelected ? AppColors.indigo.opacity(0.2) : Color.clear, radius: 4, y: 2)
        }
        .buttonStyle(.plain)
    }
}
