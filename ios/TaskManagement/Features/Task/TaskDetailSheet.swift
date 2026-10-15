import SwiftUI

public struct TaskDetailSheet: View {
    let task: TaskItem
    @Bindable var viewModel: TaskViewModel
    @Environment(\.dismiss) private var dismiss

    @State private var selectedTab = 0
    @State private var newNoteText = ""
    @State private var isPostingNote = false
    @State private var showEditSheet = false
    @State private var newSubtaskTitle = ""
    @State private var isAddingSubtask = false

    private var currentTask: TaskItem {
        viewModel.tasks.first(where: { $0.id == task.id }) ?? task
    }

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
                            Text(currentTask.title)
                                .font(.system(size: 18, weight: .bold))
                                .foregroundColor(.primary)

                            HStack(spacing: 8) {
                                if let projName = currentTask.projectName {
                                    ProjectPill(name: projName, hexColor: currentTask.projectColor)
                                }
                                PriorityBadge(priority: currentTask.priority)
                                StatusBadge(status: currentTask.status)
                                if currentTask.isRecurring == true {
                                    HStack(spacing: 4) {
                                        Image(systemName: "repeat")
                                            .font(.system(size: 10, weight: .semibold))
                                        Text(currentTask.recurrencePattern?.capitalized ?? L10n.Tasks.repeatBadge)
                                            .font(.system(size: 11, weight: .semibold))
                                    }
                                    .padding(.horizontal, 8)
                                    .padding(.vertical, 3)
                                    .background(Color.purple.opacity(0.12))
                                    .foregroundColor(.purple)
                                    .clipShape(Capsule())
                                }
                            }
                        }

                        Spacer()

                        Button(action: { dismiss() }) {
                            Image(systemName: "xmark.circle.fill")
                                .font(.system(size: 24))
                                .foregroundColor(Color.secondary.opacity(0.6))
                        }
                    }

                    if !currentTask.description.isEmpty {
                        Text(currentTask.description)
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
                        title: L10n.Tasks.tabDetails,
                        icon: "checklist",
                        isSelected: selectedTab == 0
                    ) {
                        selectedTab = 0
                    }

                    TabSegmentButton(
                        title: L10n.Tasks.tabActivities,
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
            .navigationTitle(L10n.Tasks.detailTitle)
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .navigationBarLeading) {
                    Button(action: { showEditSheet = true }) {
                        HStack(spacing: 4) {
                            Image(systemName: "pencil")
                            Text(L10n.Tasks.edit)
                        }
                        .font(.system(size: 14, weight: .semibold))
                        .foregroundColor(AppColors.primary)
                    }
                }
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button(L10n.General.close) { dismiss() }
                        .font(.system(size: 14, weight: .semibold))
                }
            }
            .sheet(isPresented: $showEditSheet) {
                EditTaskSheet(task: currentTask, viewModel: viewModel)
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
                    Text(L10n.Tasks.statusHeader)
                        .font(.system(size: 11, weight: .bold))
                        .foregroundColor(.secondary)

                    HStack(spacing: 12) {
                        Button(action: { viewModel.toggleTaskStatus(task: currentTask) }) {
                            HStack(spacing: 8) {
                                Image(systemName: currentTask.status == "done" ? "checkmark.circle.fill" : "circle")
                                    .foregroundColor(currentTask.status == "done" ? AppColors.statusDone : .secondary)
                                Text(currentTask.status == "done" ? L10n.Tasks.markUndone : L10n.Tasks.markDone)
                                    .font(.system(size: 14, weight: .medium))
                            }
                            .padding(.horizontal, 14)
                            .padding(.vertical, 10)
                            .background(Color(.secondarySystemGroupedBackground))
                            .clipShape(RoundedRectangle(cornerRadius: 10))
                        }
                    }
                }

                // Recurrence Details (if recurring)
                if currentTask.isRecurring == true {
                    VStack(alignment: .leading, spacing: 6) {
                        Text(L10n.Tasks.sectionRecurrence.uppercased())
                            .font(.system(size: 11, weight: .bold))
                            .foregroundColor(.secondary)

                        HStack(spacing: 8) {
                            Image(systemName: "repeat")
                                .foregroundColor(.purple)
                                .font(.system(size: 14, weight: .semibold))
                            Text(currentTask.recurrencePattern?.capitalized ?? L10n.Tasks.repeatBadge)
                                .font(.system(size: 13, weight: .medium))

                            if let interval = currentTask.recurrenceInterval, interval > 1 {
                                let unit = currentTask.recurrencePattern == "monthly" ? L10n.Tasks.unitMonths : currentTask.recurrencePattern == "weekly" ? L10n.Tasks.unitWeeks : L10n.Tasks.unitDays
                                Text("(\(L10n.Tasks.repeatInterval(interval: interval, unit: unit)))")
                                    .font(.system(size: 12))
                                    .foregroundColor(.secondary)
                            }
                        }
                        .padding(12)
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .background(Color.purple.opacity(0.08))
                        .clipShape(RoundedRectangle(cornerRadius: 10))
                    }
                }

                // Subtasks Section
                VStack(alignment: .leading, spacing: 10) {
                    HStack {
                        Text("SUBTASKS (\(currentTask.subtasksCompleted ?? 0)/\(currentTask.subtasksCount ?? 0))")
                            .font(.system(size: 11, weight: .bold))
                            .foregroundColor(.secondary)
                        Spacer()
                        Text("\(currentTask.progressPercent ?? 0)%")
                            .font(.system(size: 12, weight: .bold))
                            .foregroundColor(AppColors.primary)
                    }

                    // Progress bar
                    GeometryReader { geo in
                        ZStack(alignment: .leading) {
                            Capsule().fill(Color.secondary.opacity(0.15))
                            Capsule()
                                .fill(AppColors.brandGradient)
                                .frame(width: geo.size.width * CGFloat(Double(currentTask.progressPercent ?? 0) / 100.0))
                        }
                    }
                    .frame(height: 6)

                    if let subtasks = currentTask.subtasks, !subtasks.isEmpty {
                        VStack(spacing: 8) {
                            ForEach(subtasks) { subtask in
                                HStack(spacing: 10) {
                                    Button(action: {
                                        viewModel.toggleSubtask(taskId: currentTask.id, subtaskId: subtask.id, currentCompleted: subtask.isCompleted)
                                    }) {
                                        Image(systemName: subtask.isCompleted ? "checkmark.square.fill" : "square")
                                            .font(.system(size: 16))
                                            .foregroundColor(subtask.isCompleted ? AppColors.statusDone : .secondary)
                                    }
                                    .buttonStyle(.plain)

                                    VStack(alignment: .leading, spacing: 3) {
                                        Text(subtask.title)
                                            .font(.system(size: 13))
                                            .strikethrough(subtask.isCompleted)
                                            .foregroundColor(subtask.isCompleted ? .secondary : .primary)

                                        if let desc = subtask.description, !desc.isEmpty {
                                            Text(desc)
                                                .font(.system(size: 11))
                                                .foregroundColor(.secondary)
                                                .lineLimit(2)
                                        }

                                        if let due = subtask.dueDate, !due.isEmpty {
                                            HStack(spacing: 4) {
                                                Image(systemName: "calendar")
                                                    .font(.system(size: 9))
                                                Text(due)
                                                    .font(.system(size: 10))
                                            }
                                            .foregroundColor(AppColors.primary)
                                        }
                                    }

                                    Spacer()

                                    Button(action: {
                                        Task {
                                            _ = await viewModel.deleteSubtask(taskId: currentTask.id, subtaskId: subtask.id)
                                        }
                                    }) {
                                        Image(systemName: "trash")
                                            .font(.system(size: 12))
                                            .foregroundColor(Color.secondary.opacity(0.6))
                                    }
                                    .buttonStyle(.plain)
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
                            .padding(.vertical, 4)
                    }

                    // Add Subtask Row
                    HStack(spacing: 8) {
                        TextField(L10n.Tasks.addSubtaskPlaceholder, text: $newSubtaskTitle)
                            .font(.system(size: 13))
                            .textFieldStyle(.plain)
                            .padding(9)
                            .background(Color(.secondarySystemGroupedBackground))
                            .clipShape(RoundedRectangle(cornerRadius: 8))
                            .overlay(
                                RoundedRectangle(cornerRadius: 8)
                                    .stroke(Color.secondary.opacity(0.15), lineWidth: 1)
                            )
                            .onSubmit {
                                handleAddSubtask()
                            }

                        Button(action: handleAddSubtask) {
                            if isAddingSubtask {
                                ProgressView()
                                    .scaleEffect(0.8)
                                    .frame(width: 32, height: 32)
                            } else {
                                Image(systemName: "plus.circle.fill")
                                    .font(.system(size: 24))
                                    .foregroundColor(newSubtaskTitle.trimmingCharacters(in: .whitespaces).isEmpty ? Color.secondary.opacity(0.3) : AppColors.primary)
                                    .frame(width: 32, height: 32)
                            }
                        }
                        .disabled(newSubtaskTitle.trimmingCharacters(in: .whitespaces).isEmpty || isAddingSubtask)
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
                    TextField(L10n.Tasks.postNotePlaceholder, text: $newNoteText)
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
                            Text(L10n.Tasks.send)
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
                ProgressView(L10n.Tasks.loadingLogs)
                Spacer()
            } else if viewModel.taskActivities.isEmpty {
                Spacer()
                VStack(spacing: 12) {
                    Image(systemName: "clock.arrow.circlepath")
                        .font(.system(size: 40))
                        .foregroundColor(Color.secondary.opacity(0.5))
                    Text(L10n.Tasks.noLogsTitle)
                        .font(.system(size: 16, weight: .semibold))
                    Text(L10n.Tasks.noLogsDesc)
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

    private func handleAddSubtask() {
        let trimmed = newSubtaskTitle.trimmingCharacters(in: .whitespaces)
        guard !trimmed.isEmpty else { return }
        isAddingSubtask = true
        Task {
            let success = await viewModel.addSubtask(taskId: currentTask.id, title: trimmed)
            await MainActor.run {
                isAddingSubtask = false
                if success {
                    newSubtaskTitle = ""
                }
            }
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
            return ("sparkles", Color.green, L10n.Tasks.activityCreated)
        case "task_updated":
            return ("pencil", Color.blue, L10n.Tasks.activityUpdated)
        case "progress_logged":
            return ("clock.arrow.circlepath", AppColors.indigo, L10n.Tasks.activityProgress)
        case "subtask_created":
            return ("plus.circle", Color.cyan, L10n.Tasks.activitySubtaskCreated)
        case "subtask_toggled":
            return ("checkmark.circle", Color.teal, L10n.Tasks.activitySubtaskUpdated)
        case "subtask_deleted":
            return ("trash", Color.red, L10n.Tasks.activitySubtaskDeleted)
        case "focus_toggled":
            return ("star.fill", Color.yellow, L10n.Tasks.activityFocusChanged)
        default:
            return ("bubble.left.fill", Color.secondary, L10n.Tasks.activityDefault)
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
        if diff < 60 { return L10n.Tasks.relativeJustNow }
        if diff < 3600 { return L10n.Tasks.relativeMinutesAgo(count: diff / 60) }
        if diff < 86400 { return L10n.Tasks.relativeHoursAgo(count: diff / 3600) }
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

// MARK: - Edit Task Sheet
public struct EditTaskSheet: View {
    @Environment(\.dismiss) private var dismiss
    @Bindable var viewModel: TaskViewModel
    let task: TaskItem
    public var onUpdated: (() -> Void)? = nil

    @State private var title: String
    @State private var description: String
    @State private var status: String
    @State private var priority: String
    @State private var selectedProjectId: Int64?
    @State private var hasDueDate: Bool
    @State private var dueDate: Date
    @State private var isFocus: Bool
    @State private var tagsText: String
    @State private var isRecurring: Bool
    @State private var recurrencePattern: String
    @State private var recurrenceInterval: Int
    @State private var isSubmitting: Bool = false
    @State private var errorMessage: String? = nil
    @State private var newSubtaskTitle: String = ""
    @State private var isAddingSubtask: Bool = false

    private var currentTask: TaskItem {
        viewModel.tasks.first(where: { $0.id == task.id }) ?? task
    }

    public init(
        task: TaskItem,
        viewModel: TaskViewModel,
        onUpdated: (() -> Void)? = nil
    ) {
        self.task = task
        self.viewModel = viewModel
        self.onUpdated = onUpdated

        _title = State(initialValue: task.title)
        _description = State(initialValue: task.description)
        _status = State(initialValue: task.status)
        _priority = State(initialValue: task.priority)
        _selectedProjectId = State(initialValue: task.projectId)
        _isFocus = State(initialValue: task.isFocus)
        _tagsText = State(initialValue: (task.tags ?? []).joined(separator: ", "))
        _isRecurring = State(initialValue: task.isRecurring ?? false)
        _recurrencePattern = State(initialValue: task.recurrencePattern ?? "daily")
        _recurrenceInterval = State(initialValue: task.recurrenceInterval ?? 1)

        if let dueStr = task.dueDate, !dueStr.isEmpty {
            let formatter = ISO8601DateFormatter()
            formatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
            if let parsed = formatter.date(from: dueStr) {
                _hasDueDate = State(initialValue: true)
                _dueDate = State(initialValue: parsed)
            } else {
                let simpleFormatter = ISO8601DateFormatter()
                simpleFormatter.formatOptions = [.withInternetDateTime]
                if let simpleParsed = simpleFormatter.date(from: dueStr) {
                    _hasDueDate = State(initialValue: true)
                    _dueDate = State(initialValue: simpleParsed)
                } else {
                    _hasDueDate = State(initialValue: false)
                    _dueDate = State(initialValue: Date().addingTimeInterval(3600 * 24))
                }
            }
        } else {
            _hasDueDate = State(initialValue: false)
            _dueDate = State(initialValue: Date().addingTimeInterval(3600 * 24))
        }
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
                        Text("Done").tag("done")
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

                Section(header: Text(L10n.Tasks.sectionSubtasks)) {
                    if let subtasks = currentTask.subtasks, !subtasks.isEmpty {
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

                                VStack(alignment: .leading, spacing: 3) {
                                    Text(subtask.title)
                                        .font(.system(size: 14))
                                        .strikethrough(subtask.isCompleted)
                                        .foregroundColor(subtask.isCompleted ? .secondary : .primary)

                                    if let desc = subtask.description, !desc.isEmpty {
                                        Text(desc)
                                            .font(.system(size: 11))
                                            .foregroundColor(.secondary)
                                            .lineLimit(2)
                                    }

                                    if let due = subtask.dueDate, !due.isEmpty {
                                        HStack(spacing: 4) {
                                            Image(systemName: "calendar")
                                                .font(.system(size: 9))
                                            Text(due)
                                                .font(.system(size: 10))
                                        }
                                        .foregroundColor(AppColors.primary)
                                    }
                                }

                                Spacer()

                                Button(action: {
                                    Task {
                                        _ = await viewModel.deleteSubtask(taskId: task.id, subtaskId: subtask.id)
                                    }
                                }) {
                                    Image(systemName: "trash")
                                        .font(.system(size: 13))
                                        .foregroundColor(AppColors.priorityUrgent.opacity(0.8))
                                }
                                .buttonStyle(.plain)
                            }
                        }
                    }

                    HStack(spacing: 8) {
                        TextField(L10n.Tasks.addSubtaskPlaceholder, text: $newSubtaskTitle)
                            .font(.system(size: 14))
                            .onSubmit {
                                handleAddEditSubtask()
                            }

                        Button(action: handleAddEditSubtask) {
                            if isAddingSubtask {
                                ProgressView()
                                    .scaleEffect(0.8)
                            } else {
                                Image(systemName: "plus.circle.fill")
                                    .font(.system(size: 20))
                                    .foregroundColor(newSubtaskTitle.trimmingCharacters(in: .whitespaces).isEmpty ? Color.secondary.opacity(0.3) : AppColors.primary)
                            }
                        }
                        .disabled(newSubtaskTitle.trimmingCharacters(in: .whitespaces).isEmpty || isAddingSubtask)
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
            .navigationTitle(L10n.Tasks.editTitle)
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button(L10n.General.cancel) { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button(action: handleSave) {
                        HStack(spacing: 6) {
                            if isSubmitting {
                                ProgressView()
                                    .scaleEffect(0.8)
                            }
                            Text(isSubmitting ? L10n.General.saving : L10n.General.save)
                                .fontWeight(.semibold)
                        }
                    }
                    .disabled(title.trimmingCharacters(in: .whitespaces).isEmpty || isSubmitting)
                }
            }
        }
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

        Task {
            let success = await viewModel.updateTask(
                id: task.id,
                title: title.trimmingCharacters(in: .whitespaces),
                description: description.trimmingCharacters(in: .whitespaces),
                status: status,
                priority: priority,
                dueDate: isoDate,
                clearDueDate: !hasDueDate,
                isFocus: isFocus,
                tags: parsedTags,
                projectId: selectedProjectId,
                clearProjectId: selectedProjectId == nil,
                isRecurring: isRecurring,
                recurrencePattern: isRecurring ? recurrencePattern : nil,
                recurrenceInterval: isRecurring ? recurrenceInterval : nil,
                clearRecurrenceEndDate: !isRecurring
            )
            await MainActor.run {
                isSubmitting = false
                if success {
                    onUpdated?()
                    dismiss()
                } else {
                    errorMessage = viewModel.errorMessage ?? "Gagal menyimpan perubahan tugas"
                }
            }
        }
    }

    private func handleAddEditSubtask() {
        let trimmed = newSubtaskTitle.trimmingCharacters(in: .whitespaces)
        guard !trimmed.isEmpty else { return }
        isAddingSubtask = true
        Task {
            let success = await viewModel.addSubtask(taskId: task.id, title: trimmed)
            await MainActor.run {
                isAddingSubtask = false
                if success {
                    newSubtaskTitle = ""
                }
            }
        }
    }
}
