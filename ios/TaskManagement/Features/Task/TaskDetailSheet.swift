import SwiftUI

public struct TaskDetailSheet: View {
    let task: TaskItem
    @Bindable var viewModel: TaskViewModel
    @Environment(\.dismiss) private var dismiss

    @State private var selectedTab = 0
    @State private var newNoteText = ""
    @State private var isPostingNote = false
    @State private var showEditSheet = false

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
                // Status Selector Row
                VStack(alignment: .leading, spacing: 8) {
                    Text(L10n.Tasks.statusHeader)
                        .font(.system(size: 11, weight: .bold))
                        .foregroundColor(.secondary)

                    ScrollView(.horizontal, showsIndicators: false) {
                        HStack(spacing: 8) {
                            ForEach([
                                ("todo", L10n.Tasks.statusTodo, "circle", AppColors.statusTodo),
                                ("in_progress", L10n.Tasks.statusInProgress, "arrow.triangle.2.circlepath", AppColors.statusInProgress),
                                ("review", L10n.Tasks.statusReview, "eye.fill", AppColors.statusReview),
                                ("done", L10n.Tasks.statusDone, "checkmark.circle.fill", AppColors.statusDone)
                            ], id: \.0) { item in
                                let isSelected = currentTask.status == item.0
                                Button(action: {
                                    viewModel.updateTaskStatus(task: currentTask, newStatus: item.0)
                                }) {
                                    HStack(spacing: 6) {
                                        Image(systemName: item.2)
                                            .font(.system(size: 12, weight: .semibold))
                                        Text(item.1)
                                            .font(.system(size: 13, weight: isSelected ? .bold : .medium))
                                    }
                                    .padding(.horizontal, 12)
                                    .padding(.vertical, 8)
                                    .background(isSelected ? item.3.opacity(0.18) : Color(.secondarySystemGroupedBackground))
                                    .foregroundColor(isSelected ? item.3 : .secondary)
                                    .clipShape(Capsule())
                                    .overlay(
                                        Capsule()
                                            .stroke(isSelected ? item.3.opacity(0.4) : Color.secondary.opacity(0.15), lineWidth: 1)
                                    )
                                }
                                .buttonStyle(.plain)
                            }
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
                                SubtaskItemRowView(
                                    subtask: subtask,
                                    parentDueDate: currentTask.dueDate,
                                    isRecurring: currentTask.isRecurring == true,
                                    onToggle: { isCompleted in
                                        viewModel.toggleSubtask(taskId: currentTask.id, subtaskId: subtask.id, currentCompleted: !isCompleted)
                                    },
                                    onUpdate: { title, description, dueDate, clearDueDate in
                                        await viewModel.updateSubtask(
                                            taskId: currentTask.id,
                                            subtaskId: subtask.id,
                                            title: title,
                                            description: description,
                                            dueDate: dueDate,
                                            clearDueDate: clearDueDate
                                        )
                                    },
                                    onDelete: {
                                        _ = await viewModel.deleteSubtask(taskId: currentTask.id, subtaskId: subtask.id)
                                    }
                                )
                            }
                        }
                    } else {
                        Text(L10n.Tasks.noSubtasks)
                            .font(.system(size: 13))
                            .foregroundColor(.secondary)
                            .padding(.vertical, 4)
                    }

                    // Add Subtask Section
                    AddSubtaskSectionView(
                        taskId: currentTask.id,
                        parentDueDate: currentTask.dueDate,
                        isRecurring: currentTask.isRecurring == true,
                        viewModel: viewModel
                    )
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
                TaskListSkeleton(count: 3)
                    .padding(.horizontal, 16)
                    .padding(.top, 12)
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
                        Text(L10n.Tasks.statusBacklog).tag("backlog")
                        Text(L10n.Tasks.statusTodo).tag("todo")
                        Text(L10n.Tasks.statusInProgress).tag("in_progress")
                        Text(L10n.Tasks.statusReview).tag("review")
                        Text(L10n.Tasks.statusDone).tag("done")
                    }

                    Picker(L10n.Tasks.priorityLabel, selection: $priority) {
                        Text(L10n.Tasks.priorityLow).tag("low")
                        Text(L10n.Tasks.priorityMedium).tag("medium")
                        Text(L10n.Tasks.priorityHigh).tag("high")
                        Text(L10n.Tasks.priorityUrgent).tag("urgent")
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

                if !isRecurring {
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
                            SubtaskItemRowView(
                                subtask: subtask,
                                parentDueDate: task.dueDate,
                                isRecurring: isRecurring,
                                onToggle: { isCompleted in
                                    viewModel.toggleSubtask(taskId: task.id, subtaskId: subtask.id, currentCompleted: !isCompleted)
                                },
                                onUpdate: { title, description, dueDate, clearDueDate in
                                    await viewModel.updateSubtask(
                                        taskId: task.id,
                                        subtaskId: subtask.id,
                                        title: title,
                                        description: description,
                                        dueDate: dueDate,
                                        clearDueDate: clearDueDate
                                    )
                                },
                                onDelete: {
                                    _ = await viewModel.deleteSubtask(taskId: task.id, subtaskId: subtask.id)
                                }
                            )
                        }
                    } else {
                        Text(L10n.Tasks.noSubtasks)
                            .font(.system(size: 13))
                            .foregroundColor(.secondary)
                    }

                    AddSubtaskSectionView(
                        taskId: task.id,
                        parentDueDate: task.dueDate,
                        isRecurring: isRecurring,
                        viewModel: viewModel
                    )
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
        let clearDueDate: Bool
        if isRecurring {
            isoDate = nil
            clearDueDate = true
        } else if hasDueDate {
            let formatter = ISO8601DateFormatter()
            formatter.formatOptions = [.withInternetDateTime]
            isoDate = formatter.string(from: dueDate)
            clearDueDate = false
        } else {
            isoDate = nil
            clearDueDate = true
        }

        Task {
            let success = await viewModel.updateTask(
                id: task.id,
                title: title.trimmingCharacters(in: .whitespaces),
                description: description.trimmingCharacters(in: .whitespaces),
                status: status,
                priority: priority,
                dueDate: isoDate,
                clearDueDate: clearDueDate,
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
                    errorMessage = viewModel.errorMessage ?? L10n.Tasks.saveFailed
                }
            }
        }
    }
}

// MARK: - Subtask Item Row View (Preview badges, Expandable inline editor, Delete confirm)
public struct SubtaskItemRowView: View {
    let subtask: Subtask
    let parentDueDate: String?
    var isRecurring: Bool = false
    let onToggle: (Bool) -> Void
    let onUpdate: (String, String?, String?, Bool) async -> Bool
    let onDelete: () async -> Void

    @State private var isExpanded: Bool = false
    @State private var editTitle: String
    @State private var editDescription: String
    @State private var hasDueDate: Bool
    @State private var editDueDate: Date
    @State private var isSaving: Bool = false
    @State private var showDeleteConfirm: Bool = false
    @State private var validationError: String? = nil

    public init(
        subtask: Subtask,
        parentDueDate: String? = nil,
        isRecurring: Bool = false,
        onToggle: @escaping (Bool) -> Void,
        onUpdate: @escaping (String, String?, String?, Bool) async -> Bool,
        onDelete: @escaping () async -> Void
    ) {
        self.subtask = subtask
        self.parentDueDate = parentDueDate
        self.isRecurring = isRecurring
        self.onToggle = onToggle
        self.onUpdate = onUpdate
        self.onDelete = onDelete

        _editTitle = State(initialValue: subtask.title)
        _editDescription = State(initialValue: subtask.description ?? "")

        if let dueStr = subtask.dueDate, let parsed = DateTimeHelper.parseISO8601(dueStr) {
            _hasDueDate = State(initialValue: true)
            _editDueDate = State(initialValue: parsed)
        } else {
            _hasDueDate = State(initialValue: false)
            _editDueDate = State(initialValue: Date().addingTimeInterval(3600 * 4))
        }
    }

    private var parentParsedDueDate: Date? {
        guard let parent = parentDueDate, !parent.isEmpty else { return nil }
        return DateTimeHelper.parseISO8601(parent)
    }

    private var isExceedingParentDueDate: Bool {
        guard !isRecurring, hasDueDate, let parent = parentParsedDueDate else { return false }
        return editDueDate > parent
    }

    private var canSave: Bool {
        let trimmed = editTitle.trimmingCharacters(in: .whitespacesAndNewlines)
        return !trimmed.isEmpty && !isExceedingParentDueDate && !isSaving
    }

    private func handleSave() {
        let trimmedTitle = editTitle.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmedTitle.isEmpty else { return }
        if isExceedingParentDueDate {
            validationError = L10n.Tasks.subtaskDueDateExceeds
            return
        }

        let trimmedDesc = editDescription.trimmingCharacters(in: .whitespacesAndNewlines)
        let descParam: String? = trimmedDesc.isEmpty ? "" : trimmedDesc

        let dueParam: String?
        let clearDue: Bool
        if !isRecurring && hasDueDate {
            dueParam = DateTimeHelper.toISOString(editDueDate)
            clearDue = false
        } else {
            dueParam = nil
            clearDue = (subtask.dueDate != nil && !subtask.dueDate!.isEmpty)
        }

        isSaving = true
        validationError = nil

        Task {
            let success = await onUpdate(trimmedTitle, descParam, dueParam, clearDue)
            await MainActor.run {
                isSaving = false
                if success {
                    isExpanded = false
                } else {
                    validationError = L10n.Tasks.updateSubtaskFailed
                }
            }
        }
    }

    private func handleCancel() {
        editTitle = subtask.title
        editDescription = subtask.description ?? ""
        if let dueStr = subtask.dueDate, let parsed = DateTimeHelper.parseISO8601(dueStr) {
            hasDueDate = true
            editDueDate = parsed
        } else {
            hasDueDate = false
            editDueDate = Date().addingTimeInterval(3600 * 4)
        }
        validationError = nil
        withAnimation(.easeInOut(duration: 0.2)) {
            isExpanded = false
        }
    }

    public var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            summaryRow
            if isExpanded {
                Divider()
                    .padding(.horizontal, 10)
                inlineEditor
            }
        }
        .background(Color(.secondarySystemGroupedBackground))
        .clipShape(RoundedRectangle(cornerRadius: 8))
        .overlay(
            RoundedRectangle(cornerRadius: 8)
                .stroke(isExpanded ? AppColors.primary.opacity(0.4) : Color.secondary.opacity(0.12), lineWidth: 1)
        )
        .confirmationDialog(L10n.Tasks.deleteSubtaskConfirm, isPresented: $showDeleteConfirm, titleVisibility: .visible) {
            Button(L10n.General.delete, role: .destructive) {
                Task {
                    await onDelete()
                }
            }
            Button(L10n.General.cancel, role: .cancel) {}
        }
    }

    @ViewBuilder
    private var summaryRow: some View {
        HStack(spacing: 10) {
            Button(action: {
                UIImpactFeedbackGenerator(style: .light).impactOccurred()
                onToggle(!subtask.isCompleted)
            }) {
                Image(systemName: subtask.isCompleted ? "checkmark.square.fill" : "square")
                    .font(.system(size: 16))
                    .foregroundColor(subtask.isCompleted ? AppColors.statusDone : .secondary)
            }
            .buttonStyle(.plain)

            VStack(alignment: .leading, spacing: 3) {
                Text(subtask.title)
                    .font(.system(size: 13, weight: .medium))
                    .strikethrough(subtask.isCompleted)
                    .foregroundColor(subtask.isCompleted ? .secondary : .primary)
                    .lineLimit(2)

                if !isExpanded {
                    badgesRow
                }
            }

            Spacer()

            Button(action: {
                withAnimation(.easeInOut(duration: 0.2)) {
                    if !isExpanded {
                        editTitle = subtask.title
                        editDescription = subtask.description ?? ""
                        if let dueStr = subtask.dueDate, let parsed = DateTimeHelper.parseISO8601(dueStr) {
                            hasDueDate = true
                            editDueDate = parsed
                        } else {
                            hasDueDate = false
                        }
                    }
                    isExpanded.toggle()
                }
            }) {
                Image(systemName: isExpanded ? "chevron.up" : "chevron.down")
                    .font(.system(size: 11, weight: .bold))
                    .foregroundColor(isExpanded ? AppColors.primary : Color.secondary.opacity(0.7))
                    .frame(width: 26, height: 26)
                    .background(isExpanded ? AppColors.primary.opacity(0.1) : Color.clear)
                    .clipShape(RoundedRectangle(cornerRadius: 6))
            }
            .buttonStyle(.plain)

            Button(action: {
                showDeleteConfirm = true
            }) {
                Image(systemName: "trash")
                    .font(.system(size: 12))
                    .foregroundColor(Color.secondary.opacity(0.6))
                    .frame(width: 26, height: 26)
            }
            .buttonStyle(.plain)
        }
        .padding(10)
    }

    @ViewBuilder
    private var badgesRow: some View {
        let hasDesc = (subtask.description != nil && !subtask.description!.isEmpty)
        let hasDue = !isRecurring && (subtask.dueDate != nil && !subtask.dueDate!.isEmpty)
        if hasDesc || hasDue {
            HStack(spacing: 6) {
                if let desc = subtask.description, !desc.isEmpty {
                    HStack(spacing: 3) {
                        Image(systemName: "text.alignleft")
                            .font(.system(size: 8))
                        Text(desc)
                            .font(.system(size: 10))
                            .lineLimit(1)
                    }
                    .padding(.horizontal, 5)
                    .padding(.vertical, 2)
                    .background(Color.secondary.opacity(0.12))
                    .foregroundColor(.secondary)
                    .clipShape(RoundedRectangle(cornerRadius: 4))
                }

                if !isRecurring, let due = subtask.dueDate, !due.isEmpty {
                    HStack(spacing: 3) {
                        Image(systemName: "calendar")
                            .font(.system(size: 8))
                        Text(DateTimeHelper.formatFullDateTime(due))
                            .font(.system(size: 10, weight: .medium))
                    }
                    .padding(.horizontal, 6)
                    .padding(.vertical, 2)
                    .background(AppColors.indigo.opacity(0.12))
                    .foregroundColor(AppColors.indigo)
                    .clipShape(Capsule())
                }
            }
        }
    }

    @ViewBuilder
    private var inlineEditor: some View {
        VStack(alignment: .leading, spacing: 9) {
            VStack(alignment: .leading, spacing: 3) {
                Text(L10n.Tasks.subtaskTitlePlaceholder)
                    .font(.system(size: 10, weight: .semibold))
                    .foregroundColor(.secondary)
                TextField(L10n.Tasks.subtaskTitlePlaceholder, text: $editTitle)
                    .font(.system(size: 13))
                    .textFieldStyle(.plain)
                    .padding(8)
                    .background(Color(.tertiarySystemGroupedBackground))
                    .clipShape(RoundedRectangle(cornerRadius: 6))
            }

            VStack(alignment: .leading, spacing: 3) {
                Text(L10n.Tasks.subtaskDetails)
                    .font(.system(size: 10, weight: .semibold))
                    .foregroundColor(.secondary)
                TextField(L10n.Tasks.subtaskDescriptionPlaceholder, text: $editDescription, axis: .vertical)
                    .font(.system(size: 12))
                    .textFieldStyle(.plain)
                    .lineLimit(2...4)
                    .padding(8)
                    .background(Color(.tertiarySystemGroupedBackground))
                    .clipShape(RoundedRectangle(cornerRadius: 6))
            }

            if !isRecurring {
                dueDatePicker
            }

            if let err = validationError {
                Text(err)
                    .font(.system(size: 11))
                    .foregroundColor(AppColors.priorityUrgent)
            }

            editorButtons
        }
        .padding(10)
    }

    @ViewBuilder
    private var dueDatePicker: some View {
        VStack(alignment: .leading, spacing: 6) {
            Toggle(isOn: $hasDueDate.animation()) {
                Label(L10n.Tasks.subtaskDueDate, systemImage: "calendar")
                    .font(.system(size: 12, weight: .medium))
            }
            .tint(AppColors.primary)

            if hasDueDate {
                DatePicker(
                    "",
                    selection: $editDueDate,
                    displayedComponents: [.date, .hourAndMinute]
                )
                .datePickerStyle(.compact)
                .labelsHidden()

                if isExceedingParentDueDate {
                    HStack(spacing: 4) {
                        Image(systemName: "exclamationmark.triangle.fill")
                            .font(.system(size: 10))
                        Text(L10n.Tasks.subtaskDueDateExceeds)
                            .font(.system(size: 10))
                    }
                    .foregroundColor(AppColors.priorityUrgent)
                }
            }
        }
    }

    @ViewBuilder
    private var editorButtons: some View {
        HStack(spacing: 8) {
            Spacer()
            Button(action: handleCancel) {
                Text(L10n.General.cancel)
                    .font(.system(size: 11, weight: .medium))
                    .foregroundColor(.secondary)
                    .padding(.horizontal, 10)
                    .padding(.vertical, 6)
                    .background(Color.secondary.opacity(0.1))
                    .clipShape(RoundedRectangle(cornerRadius: 6))
            }
            .buttonStyle(.plain)

            Button(action: handleSave) {
                HStack(spacing: 4) {
                    if isSaving {
                        ProgressView()
                            .scaleEffect(0.7)
                    }
                    Text(L10n.General.save)
                        .font(.system(size: 11, weight: .semibold))
                }
                .foregroundColor(.white)
                .padding(.horizontal, 14)
                .padding(.vertical, 6)
                .background(canSave ? AppColors.primary : Color.secondary.opacity(0.3))
                .clipShape(RoundedRectangle(cornerRadius: 6))
            }
            .buttonStyle(.plain)
            .disabled(!canSave || isSaving)
        }
    }
}

// MARK: - Add Subtask Section View (With expandable Details & Due Date)
public struct AddSubtaskSectionView: View {
    let taskId: Int64
    let parentDueDate: String?
    var isRecurring: Bool = false
    let viewModel: TaskViewModel

    public init(
        taskId: Int64,
        parentDueDate: String? = nil,
        isRecurring: Bool = false,
        viewModel: TaskViewModel
    ) {
        self.taskId = taskId
        self.parentDueDate = parentDueDate
        self.isRecurring = isRecurring
        self.viewModel = viewModel
    }

    @State private var newSubtaskTitle: String = ""
    @State private var newSubtaskDescription: String = ""
    @State private var hasDueDate: Bool = false
    @State private var newSubtaskDueDate: Date = Date().addingTimeInterval(3600 * 4)
    @State private var isShowingDetails: Bool = false
    @State private var isAdding: Bool = false
    @State private var errorMessage: String? = nil

    private var parentParsedDueDate: Date? {
        guard let parent = parentDueDate, !parent.isEmpty else { return nil }
        return DateTimeHelper.parseISO8601(parent)
    }

    private var isExceedingParentDueDate: Bool {
        guard !isRecurring, hasDueDate, let parent = parentParsedDueDate else { return false }
        return newSubtaskDueDate > parent
    }

    private var canSubmit: Bool {
        let trimmed = newSubtaskTitle.trimmingCharacters(in: .whitespacesAndNewlines)
        return !trimmed.isEmpty && !isExceedingParentDueDate && !isAdding
    }

    private func handleSubmit() {
        let trimmedTitle = newSubtaskTitle.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmedTitle.isEmpty else { return }
        if isExceedingParentDueDate {
            errorMessage = L10n.Tasks.subtaskDueDateExceeds
            return
        }

        let trimmedDesc = newSubtaskDescription.trimmingCharacters(in: .whitespacesAndNewlines)
        let descParam = trimmedDesc.isEmpty ? nil : trimmedDesc
        let dueParam = (!isRecurring && hasDueDate) ? DateTimeHelper.toISOString(newSubtaskDueDate) : nil

        isAdding = true
        errorMessage = nil

        Task {
            let success = await viewModel.addSubtask(
                taskId: taskId,
                title: trimmedTitle,
                description: descParam,
                dueDate: dueParam
            )
            await MainActor.run {
                isAdding = false
                if success {
                    newSubtaskTitle = ""
                    newSubtaskDescription = ""
                    hasDueDate = false
                    isShowingDetails = false
                } else {
                    errorMessage = viewModel.errorMessage ?? L10n.Tasks.addSubtaskFailed
                }
            }
        }
    }

    public var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            mainInputRow
            if isShowingDetails {
                detailsSection
            }
            if let err = errorMessage {
                Text(err)
                    .font(.system(size: 11))
                    .foregroundColor(AppColors.priorityUrgent)
            }
        }
    }

    @ViewBuilder
    private var mainInputRow: some View {
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
                    if canSubmit {
                        handleSubmit()
                    }
                }

            Button(action: {
                withAnimation(.easeInOut(duration: 0.2)) {
                    isShowingDetails.toggle()
                }
            }) {
                Image(systemName: "slider.horizontal.3")
                    .font(.system(size: 14))
                    .foregroundColor(isShowingDetails ? AppColors.primary : .secondary)
                    .frame(width: 34, height: 34)
                    .background(isShowingDetails ? AppColors.primary.opacity(0.12) : Color(.secondarySystemGroupedBackground))
                    .clipShape(RoundedRectangle(cornerRadius: 8))
                    .overlay(
                        RoundedRectangle(cornerRadius: 8)
                            .stroke(isShowingDetails ? AppColors.primary.opacity(0.3) : Color.secondary.opacity(0.15), lineWidth: 1)
                    )
            }
            .buttonStyle(.plain)

            Button(action: handleSubmit) {
                if isAdding {
                    ProgressView()
                        .scaleEffect(0.8)
                        .frame(width: 34, height: 34)
                } else {
                    Image(systemName: "plus.circle.fill")
                        .font(.system(size: 26))
                        .foregroundColor(canSubmit ? AppColors.primary : Color.secondary.opacity(0.3))
                        .frame(width: 34, height: 34)
                }
            }
            .buttonStyle(.plain)
            .disabled(!canSubmit)
        }
    }

    @ViewBuilder
    private var detailsSection: some View {
        VStack(alignment: .leading, spacing: 8) {
            TextField(L10n.Tasks.subtaskDescriptionPlaceholder, text: $newSubtaskDescription, axis: .vertical)
                .font(.system(size: 12))
                .lineLimit(2...3)
                .textFieldStyle(.plain)
                .padding(8)
                .background(Color(.secondarySystemGroupedBackground))
                .clipShape(RoundedRectangle(cornerRadius: 6))
                .overlay(
                    RoundedRectangle(cornerRadius: 6)
                        .stroke(Color.secondary.opacity(0.15), lineWidth: 1)
                )

            if !isRecurring {
                Toggle(isOn: $hasDueDate.animation()) {
                    Label(L10n.Tasks.subtaskDueDate, systemImage: "calendar")
                        .font(.system(size: 12, weight: .medium))
                }
                .tint(AppColors.primary)

                if hasDueDate {
                    DatePicker(
                        "",
                        selection: $newSubtaskDueDate,
                        displayedComponents: [.date, .hourAndMinute]
                    )
                    .datePickerStyle(.compact)
                    .labelsHidden()

                    if isExceedingParentDueDate {
                        HStack(spacing: 4) {
                            Image(systemName: "exclamationmark.triangle.fill")
                                .font(.system(size: 10))
                            Text(L10n.Tasks.subtaskDueDateExceeds)
                                .font(.system(size: 10))
                        }
                        .foregroundColor(AppColors.priorityUrgent)
                    }
                }
            }
        }
        .padding(10)
        .background(Color(.secondarySystemGroupedBackground).opacity(0.6))
        .clipShape(RoundedRectangle(cornerRadius: 8))
        .overlay(
            RoundedRectangle(cornerRadius: 8)
                .stroke(Color.secondary.opacity(0.1), lineWidth: 1)
        )
    }
}


