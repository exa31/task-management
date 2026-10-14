import SwiftUI
import AppIntents
#if canImport(WidgetKit)
import WidgetKit
#endif

// MARK: - 1. Dedicated Recurring Task Checklist Widget View
public struct RecurringTaskWidgetView: View {
    @Environment(\.widgetFamily) var family
    public let entry: TaskWidgetEntry

    public init(entry: TaskWidgetEntry) {
        self.entry = entry
    }

    private var allRecurringTasks: [TaskItem] {
        let combined = entry.bundle.focusTasks + entry.bundle.upcomingDeadlines
        var seen = Set<Int64>()
        var result: [TaskItem] = []
        for t in combined {
            if t.isRecurring == true && !seen.contains(t.id) {
                seen.insert(t.id)
                result.append(t)
            }
        }
        return result
    }

    private var activeRecurringTask: TaskItem? {
        // Prioritize recurring task that is not done yet
        allRecurringTasks.first { $0.status != "done" } ?? allRecurringTasks.first
    }

    public var body: some View {
        switch family {
        case .systemSmall:
            smallView
        case .systemLarge:
            largeView
        default:
            mediumView
        }
    }

    // MARK: - Small Family

    private var smallView: some View {
        VStack(alignment: .leading, spacing: 6) {
            // Header Pill & Ratio
            HStack(spacing: 4) {
                HStack(spacing: 3) {
                    Image(systemName: "repeat.circle.fill")
                        .font(.system(size: 11, weight: .bold))
                    Text("RUTIN")
                        .font(.system(size: 9, weight: .black))
                        .tracking(0.5)
                }
                .foregroundStyle(Color(red: 56/255, green: 189/255, blue: 248/255))
                .padding(.horizontal, 6)
                .padding(.vertical, 2)
                .background(Color(red: 56/255, green: 189/255, blue: 248/255).opacity(0.15))
                .clipShape(Capsule())

                Spacer()

                if let task = activeRecurringTask, let count = task.subtasksCount, count > 0 {
                    let done = task.subtasksCompleted ?? 0
                    Text("\(done)/\(count)")
                        .font(.system(size: 10, weight: .bold, design: .monospaced))
                        .foregroundStyle(done == count ? Color(red: 52/255, green: 211/255, blue: 153/255) : Color(red: 56/255, green: 189/255, blue: 248/255))
                        .padding(.horizontal, 5)
                        .padding(.vertical, 1.5)
                        .background(Color.white.opacity(0.08))
                        .clipShape(Capsule())
                }
            }

            if let task = activeRecurringTask {
                if task.status == "done" {
                    doneStateView(isCompact: true)
                } else {
                    Text(task.title)
                        .font(.system(size: 12, weight: .bold))
                        .lineLimit(1)
                        .foregroundStyle(.white)

                    progressBar(for: task)

                    let subtasks = (task.subtasks ?? []).prefix(2)
                    let allChecked = !(task.subtasks ?? []).isEmpty && (task.subtasks ?? []).allSatisfy { $0.isCompleted }

                    if allChecked {
                        smallActionsBar(taskId: task.id)
                    } else if subtasks.isEmpty {
                        Text("Tidak ada subtask")
                            .font(.system(size: 10))
                            .foregroundStyle(.white.opacity(0.6))
                    } else {
                        VStack(alignment: .leading, spacing: 4) {
                            ForEach(subtasks) { st in
                                subtaskRow(taskId: task.id, subtask: st, isCompact: true)
                            }
                        }
                    }
                }
            } else {
                emptyRecurringView(isCompact: true)
            }

            Spacer(minLength: 0)
        }
        .padding(12)
        .widgetBackground(accentColor: Color(red: 56/255, green: 189/255, blue: 248/255))
    }

    // MARK: - Medium Family

    private var mediumView: some View {
        VStack(alignment: .leading, spacing: 8) {
            // Header Bar
            HStack(spacing: 8) {
                HStack(spacing: 4) {
                    Image(systemName: "repeat.circle.fill")
                        .font(.system(size: 12, weight: .bold))
                    Text("TUGAS RUTIN HARIAN")
                        .font(.system(size: 9, weight: .black))
                        .tracking(0.8)
                }
                .foregroundStyle(Color(red: 56/255, green: 189/255, blue: 248/255))
                .padding(.horizontal, 7)
                .padding(.vertical, 3)
                .background(Color(red: 56/255, green: 189/255, blue: 248/255).opacity(0.15))
                .clipShape(Capsule())

                Spacer()

                if let task = activeRecurringTask, let count = task.subtasksCount, count > 0 {
                    let done = task.subtasksCompleted ?? 0
                    HStack(spacing: 4) {
                        Image(systemName: done == count ? "checkmark.seal.fill" : "circle.dashed")
                            .font(.system(size: 11))
                        Text("\(done)/\(count) selesai")
                            .font(.system(size: 11, weight: .bold))
                    }
                    .foregroundStyle(done == count ? Color(red: 52/255, green: 211/255, blue: 153/255) : Color(red: 56/255, green: 189/255, blue: 248/255))
                    .padding(.horizontal, 7)
                    .padding(.vertical, 3)
                    .background(Color.white.opacity(0.08))
                    .clipShape(Capsule())
                }
            }

            if let task = activeRecurringTask {
                if task.status == "done" {
                    doneStateView(isCompact: false)
                } else {
                    HStack(alignment: .center) {
                        Text(task.title)
                            .font(.system(size: 13, weight: .bold))
                            .lineLimit(1)
                            .foregroundStyle(.white)

                        Spacer()

                        if let pattern = task.recurrencePattern {
                            Text(formatPattern(pattern))
                                .font(.system(size: 9, weight: .semibold))
                                .foregroundStyle(.white.opacity(0.6))
                                .padding(.horizontal, 5)
                                .padding(.vertical, 2)
                                .background(Color.white.opacity(0.06))
                                .clipShape(RoundedRectangle(cornerRadius: 4))
                        }
                    }

                    progressBar(for: task)

                    let subtasks = (task.subtasks ?? []).prefix(3)
                    let allChecked = !(task.subtasks ?? []).isEmpty && (task.subtasks ?? []).allSatisfy { $0.isCompleted }

                    if allChecked {
                        statusActionBar(taskId: task.id)
                    } else if subtasks.isEmpty {
                        Text("Belum ada subtask untuk tugas ini")
                            .font(.system(size: 11))
                            .foregroundStyle(.white.opacity(0.6))
                            .padding(.vertical, 4)
                    } else {
                        VStack(alignment: .leading, spacing: 5) {
                            ForEach(subtasks) { st in
                                subtaskRow(taskId: task.id, subtask: st, isCompact: false)
                            }
                        }
                    }
                }
            } else {
                emptyRecurringView(isCompact: false)
            }

            Spacer(minLength: 0)
        }
        .padding(14)
        .widgetBackground(accentColor: Color(red: 56/255, green: 189/255, blue: 248/255))
    }

    // MARK: - Large Family

    private var largeView: some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack(spacing: 8) {
                HStack(spacing: 4) {
                    Image(systemName: "repeat.circle.fill")
                        .font(.system(size: 14, weight: .bold))
                    Text("TUGAS RUTIN HARIAN")
                        .font(.system(size: 11, weight: .black))
                        .tracking(1)
                }
                .foregroundStyle(Color(red: 56/255, green: 189/255, blue: 248/255))
                .padding(.horizontal, 8)
                .padding(.vertical, 4)
                .background(Color(red: 56/255, green: 189/255, blue: 248/255).opacity(0.15))
                .clipShape(Capsule())

                Spacer()

                if let task = activeRecurringTask, let count = task.subtasksCount, count > 0 {
                    let done = task.subtasksCompleted ?? 0
                    Text("\(done) dari \(count) Subtask")
                        .font(.system(size: 12, weight: .bold))
                        .foregroundStyle(Color(red: 52/255, green: 211/255, blue: 153/255))
                }
            }

            if let task = activeRecurringTask {
                if task.status == "done" {
                    doneStateView(isCompact: false)
                } else {
                    Text(task.title)
                        .font(.system(size: 15, weight: .bold))
                        .lineLimit(1)
                        .foregroundStyle(.white)

                    progressBar(for: task)

                    let subtasks = (task.subtasks ?? []).prefix(5)
                    let allChecked = !(task.subtasks ?? []).isEmpty && (task.subtasks ?? []).allSatisfy { $0.isCompleted }

                    VStack(alignment: .leading, spacing: 6) {
                        ForEach(subtasks) { st in
                            subtaskRow(taskId: task.id, subtask: st, isCompact: false)
                        }
                    }

                    if allChecked {
                        statusActionBar(taskId: task.id)
                    }
                }
            } else {
                emptyRecurringView(isCompact: false)
            }

            Spacer(minLength: 0)
        }
        .padding(16)
        .widgetBackground(accentColor: Color(red: 56/255, green: 189/255, blue: 248/255))
    }

    // MARK: - Subcomponents

    @ViewBuilder
    private func progressBar(for task: TaskItem) -> some View {
        let total = max(task.subtasksCount ?? 0, 1)
        let done = min(task.subtasksCompleted ?? 0, total)
        let ratio = CGFloat(done) / CGFloat(total)

        GeometryReader { geo in
            ZStack(alignment: .leading) {
                Capsule()
                    .fill(Color.white.opacity(0.1))
                    .frame(height: 4)

                Capsule()
                    .fill(
                        LinearGradient(
                            colors: [
                                Color(red: 56/255, green: 189/255, blue: 248/255),
                                Color(red: 52/255, green: 211/255, blue: 153/255)
                            ],
                            startPoint: .leading,
                            endPoint: .trailing
                        )
                    )
                    .frame(width: max(geo.size.width * ratio, 6), height: 4)
            }
        }
        .frame(height: 4)
    }

    @ViewBuilder
    private func subtaskRow(taskId: Int64, subtask: Subtask, isCompact: Bool) -> some View {
        HStack(spacing: 8) {
            Button(intent: ToggleSubtaskIntent(taskId: taskId, subtaskId: subtask.id, isCompleted: !subtask.isCompleted)) {
                Image(systemName: subtask.isCompleted ? "checkmark.circle.fill" : "circle")
                    .font(.system(size: isCompact ? 13 : 15, weight: .semibold))
                    .foregroundStyle(subtask.isCompleted ? Color(red: 52/255, green: 211/255, blue: 153/255) : .white.opacity(0.45))
            }
            .buttonStyle(.plain)

            Text(subtask.title)
                .font(.system(size: isCompact ? 11 : 12, weight: subtask.isCompleted ? .regular : .medium))
                .lineLimit(1)
                .strikethrough(subtask.isCompleted, color: .white.opacity(0.35))
                .foregroundStyle(subtask.isCompleted ? .white.opacity(0.4) : .white)

            Spacer()
        }
        .padding(.horizontal, 8)
        .padding(.vertical, isCompact ? 4 : 5)
        .background(Color.white.opacity(0.04))
        .clipShape(RoundedRectangle(cornerRadius: 7, style: .continuous))
    }

    // Action buttons bar when all subtasks completed
    @ViewBuilder
    private func statusActionBar(taskId: Int64) -> some View {
        VStack(spacing: 6) {
            HStack(spacing: 4) {
                Image(systemName: "sparkles")
                    .font(.system(size: 10, weight: .bold))
                Text("Semua subtask selesai! Pilih tindakan:")
                    .font(.system(size: 10, weight: .semibold))
            }
            .foregroundStyle(Color(red: 52/255, green: 211/255, blue: 153/255))

            HStack(spacing: 8) {
                Button(intent: UpdateTaskStatusIntent(taskId: taskId, newStatus: "done")) {
                    HStack(spacing: 4) {
                        Image(systemName: "checkmark.circle.fill")
                            .font(.system(size: 11, weight: .bold))
                        Text("Selesaikan")
                            .font(.system(size: 11, weight: .bold))
                    }
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 5)
                    .background(
                        LinearGradient(
                            colors: [Color(red: 16/255, green: 185/255, blue: 129/255), Color(red: 5/255, green: 150/255, blue: 105/255)],
                            startPoint: .topLeading,
                            endPoint: .bottomTrailing
                        )
                    )
                    .foregroundStyle(.white)
                    .clipShape(RoundedRectangle(cornerRadius: 8, style: .continuous))
                }
                .buttonStyle(.plain)

                Button(intent: UpdateTaskStatusIntent(taskId: taskId, newStatus: "in_review")) {
                    HStack(spacing: 4) {
                        Image(systemName: "eye.fill")
                            .font(.system(size: 11, weight: .bold))
                        Text("Review")
                            .font(.system(size: 11, weight: .bold))
                    }
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 5)
                    .background(Color.white.opacity(0.12))
                    .foregroundStyle(.white)
                    .clipShape(RoundedRectangle(cornerRadius: 8, style: .continuous))
                }
                .buttonStyle(.plain)
            }
        }
        .padding(8)
        .background(Color(red: 16/255, green: 185/255, blue: 129/255).opacity(0.12))
        .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))
    }

    @ViewBuilder
    private func smallActionsBar(taskId: Int64) -> some View {
        Button(intent: UpdateTaskStatusIntent(taskId: taskId, newStatus: "done")) {
            HStack(spacing: 4) {
                Image(systemName: "checkmark.circle.fill")
                    .font(.system(size: 10, weight: .bold))
                Text("Tandai Selesai")
                    .font(.system(size: 10, weight: .bold))
            }
            .frame(maxWidth: .infinity)
            .padding(.vertical, 4)
            .background(Color(red: 16/255, green: 185/255, blue: 129/255))
            .foregroundStyle(.white)
            .clipShape(RoundedRectangle(cornerRadius: 6))
        }
        .buttonStyle(.plain)
    }

    // Done state view
    @ViewBuilder
    private func doneStateView(isCompact: Bool) -> some View {
        VStack(spacing: isCompact ? 4 : 6) {
            Image(systemName: "checkmark.seal.fill")
                .font(.system(size: isCompact ? 22 : 28))
                .foregroundStyle(
                    LinearGradient(
                        colors: [Color(red: 52/255, green: 211/255, blue: 153/255), Color(red: 16/255, green: 185/255, blue: 129/255)],
                        startPoint: .topLeading,
                        endPoint: .bottomTrailing
                    )
                )

            Text("Rutin Hari Ini Selesai!")
                .font(.system(size: isCompact ? 12 : 13, weight: .bold))
                .foregroundStyle(.white)

            Text("Istirahat dulu, tugas akan muncul kembali besok.")
                .font(.system(size: isCompact ? 9 : 10))
                .foregroundStyle(.white.opacity(0.6))
                .multilineTextAlignment(.center)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }

    // Empty state view
    @ViewBuilder
    private func emptyRecurringView(isCompact: Bool) -> some View {
        VStack(spacing: isCompact ? 4 : 6) {
            Image(systemName: "repeat")
                .font(.system(size: isCompact ? 20 : 24))
                .foregroundStyle(Color(red: 56/255, green: 189/255, blue: 248/255).opacity(0.6))

            Text("Tidak Ada Tugas Rutin")
                .font(.system(size: isCompact ? 11 : 12, weight: .semibold))
                .foregroundStyle(.white.opacity(0.8))

            if !isCompact {
                Text("Aktifkan 'Ulangi Otomatis' pada tugas di aplikasi.")
                    .font(.system(size: 9))
                    .foregroundStyle(.white.opacity(0.5))
                    .multilineTextAlignment(.center)
            }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }

    private func formatPattern(_ p: String) -> String {
        switch p.lowercased() {
        case "daily": return "Harian"
        case "weekly": return "Mingguan"
        case "monthly": return "Bulanan"
        default: return p.capitalized
        }
    }
}

// MARK: - 2. Dedicated Custom Selected Task Checklist Widget View
public struct CustomTaskWidgetView: View {
    @Environment(\.widgetFamily) var family
    public let entry: TaskWidgetEntry

    public init(entry: TaskWidgetEntry) {
        self.entry = entry
    }

    private var activeTasks: [TaskItem] {
        let combined = entry.bundle.focusTasks + entry.bundle.upcomingDeadlines
        var seen = Set<Int64>()
        var result: [TaskItem] = []
        for t in combined {
            if !seen.contains(t.id) && t.status != "done" {
                seen.insert(t.id)
                result.append(t)
            }
        }
        return result
    }

    private var currentTask: TaskItem? {
        guard !activeTasks.isEmpty else { return nil }
        let idx = SharedWidgetDataManager.shared.getCustomTaskIndex() % activeTasks.count
        return activeTasks[idx]
    }

    public var body: some View {
        switch family {
        case .systemSmall:
            smallView
        case .systemLarge:
            largeView
        default:
            mediumView
        }
    }

    // MARK: - Small Family

    private var smallView: some View {
        VStack(alignment: .leading, spacing: 6) {
            HStack(spacing: 4) {
                HStack(spacing: 3) {
                    Image(systemName: "checklist.checked")
                        .font(.system(size: 11, weight: .bold))
                    Text("PILIHAN")
                        .font(.system(size: 9, weight: .black))
                        .tracking(0.5)
                }
                .foregroundStyle(Color(red: 168/255, green: 85/255, blue: 247/255))
                .padding(.horizontal, 6)
                .padding(.vertical, 2)
                .background(Color(red: 168/255, green: 85/255, blue: 247/255).opacity(0.15))
                .clipShape(Capsule())

                Spacer()

                if activeTasks.count > 1 {
                    Button(intent: CycleNextTaskIntent()) {
                        Image(systemName: "arrow.triangle.2.circlepath")
                            .font(.system(size: 10, weight: .bold))
                            .foregroundStyle(.white.opacity(0.7))
                            .padding(3)
                            .background(Color.white.opacity(0.1))
                            .clipShape(Circle())
                    }
                    .buttonStyle(.plain)
                }
            }

            if let task = currentTask {
                Text(task.title)
                    .font(.system(size: 12, weight: .bold))
                    .lineLimit(1)
                    .foregroundStyle(.white)

                let subtasks = (task.subtasks ?? []).prefix(2)
                let allChecked = !(task.subtasks ?? []).isEmpty && (task.subtasks ?? []).allSatisfy { $0.isCompleted }

                if allChecked {
                    Button(intent: UpdateTaskStatusIntent(taskId: task.id, newStatus: "done")) {
                        HStack(spacing: 4) {
                            Image(systemName: "checkmark.circle.fill")
                                .font(.system(size: 10, weight: .bold))
                            Text("Selesaikan")
                                .font(.system(size: 10, weight: .bold))
                        }
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 4)
                        .background(Color(red: 16/255, green: 185/255, blue: 129/255))
                        .foregroundStyle(.white)
                        .clipShape(RoundedRectangle(cornerRadius: 6))
                    }
                    .buttonStyle(.plain)
                } else if subtasks.isEmpty {
                    Text("Tidak ada subtask")
                        .font(.system(size: 10))
                        .foregroundStyle(.white.opacity(0.6))
                } else {
                    VStack(alignment: .leading, spacing: 4) {
                        ForEach(subtasks) { st in
                            subtaskRow(taskId: task.id, subtask: st, isCompact: true)
                        }
                    }
                }
            } else {
                emptyTasksView(isCompact: true)
            }

            Spacer(minLength: 0)
        }
        .padding(12)
        .widgetBackground(accentColor: Color(red: 168/255, green: 85/255, blue: 247/255))
    }

    // MARK: - Medium Family

    private var mediumView: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack(spacing: 8) {
                HStack(spacing: 4) {
                    Image(systemName: "checklist.checked")
                        .font(.system(size: 12, weight: .bold))
                    Text("CHECKLIST TUGAS")
                        .font(.system(size: 9, weight: .black))
                        .tracking(0.8)
                }
                .foregroundStyle(Color(red: 168/255, green: 85/255, blue: 247/255))
                .padding(.horizontal, 7)
                .padding(.vertical, 3)
                .background(Color(red: 168/255, green: 85/255, blue: 247/255).opacity(0.15))
                .clipShape(Capsule())

                Spacer()

                if activeTasks.count > 1 {
                    Button(intent: CycleNextTaskIntent()) {
                        HStack(spacing: 4) {
                            Text("Ganti")
                                .font(.system(size: 10, weight: .semibold))
                            Image(systemName: "arrow.triangle.2.circlepath")
                                .font(.system(size: 9, weight: .bold))
                        }
                        .foregroundStyle(.white.opacity(0.8))
                        .padding(.horizontal, 7)
                        .padding(.vertical, 3)
                        .background(Color.white.opacity(0.08))
                        .clipShape(Capsule())
                    }
                    .buttonStyle(.plain)
                }
            }

            if let task = currentTask {
                HStack {
                    Text(task.title)
                        .font(.system(size: 13, weight: .bold))
                        .lineLimit(1)
                        .foregroundStyle(.white)

                    Spacer()

                    let pri = task.priority
                    Text(pri.uppercased())
                        .font(.system(size: 8, weight: .black))
                        .foregroundStyle(priorityColor(pri))
                        .padding(.horizontal, 5)
                        .padding(.vertical, 2)
                        .background(priorityColor(pri).opacity(0.15))
                        .clipShape(RoundedRectangle(cornerRadius: 4))
                }

                let subtasks = (task.subtasks ?? []).prefix(3)
                let allChecked = !(task.subtasks ?? []).isEmpty && (task.subtasks ?? []).allSatisfy { $0.isCompleted }

                if allChecked {
                    HStack(spacing: 8) {
                        Button(intent: UpdateTaskStatusIntent(taskId: task.id, newStatus: "done")) {
                            HStack(spacing: 4) {
                                Image(systemName: "checkmark.circle.fill")
                                    .font(.system(size: 11, weight: .bold))
                                Text("Selesaikan Tugas")
                                    .font(.system(size: 11, weight: .bold))
                            }
                            .frame(maxWidth: .infinity)
                            .padding(.vertical, 6)
                            .background(Color(red: 16/255, green: 185/255, blue: 129/255))
                            .foregroundStyle(.white)
                            .clipShape(RoundedRectangle(cornerRadius: 8))
                        }
                        .buttonStyle(.plain)

                        if activeTasks.count > 1 {
                            Button(intent: CycleNextTaskIntent()) {
                                Text("Tugas Lain")
                                    .font(.system(size: 11, weight: .semibold))
                                    .padding(.horizontal, 10)
                                    .padding(.vertical, 6)
                                    .background(Color.white.opacity(0.1))
                                    .foregroundStyle(.white)
                                    .clipShape(RoundedRectangle(cornerRadius: 8))
                            }
                            .buttonStyle(.plain)
                        }
                    }
                    .padding(8)
                    .background(Color(red: 16/255, green: 185/255, blue: 129/255).opacity(0.12))
                    .clipShape(RoundedRectangle(cornerRadius: 10))
                } else if subtasks.isEmpty {
                    Text("Tugas ini tidak memiliki subtask.")
                        .font(.system(size: 11))
                        .foregroundStyle(.white.opacity(0.6))
                        .padding(.vertical, 4)
                } else {
                    VStack(alignment: .leading, spacing: 5) {
                        ForEach(subtasks) { st in
                            subtaskRow(taskId: task.id, subtask: st, isCompact: false)
                        }
                    }
                }
            } else {
                emptyTasksView(isCompact: false)
            }

            Spacer(minLength: 0)
        }
        .padding(14)
        .widgetBackground(accentColor: Color(red: 168/255, green: 85/255, blue: 247/255))
    }

    // MARK: - Large Family

    private var largeView: some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack(spacing: 8) {
                HStack(spacing: 4) {
                    Image(systemName: "checklist.checked")
                        .font(.system(size: 14, weight: .bold))
                    Text("CHECKLIST TUGAS PILIHAN")
                        .font(.system(size: 11, weight: .black))
                        .tracking(1)
                }
                .foregroundStyle(Color(red: 168/255, green: 85/255, blue: 247/255))
                .padding(.horizontal, 8)
                .padding(.vertical, 4)
                .background(Color(red: 168/255, green: 85/255, blue: 247/255).opacity(0.15))
                .clipShape(Capsule())

                Spacer()

                if activeTasks.count > 1 {
                    Button(intent: CycleNextTaskIntent()) {
                        HStack(spacing: 4) {
                            Text("Tugas Lainnya")
                                .font(.system(size: 11, weight: .semibold))
                            Image(systemName: "arrow.triangle.2.circlepath")
                                .font(.system(size: 10, weight: .bold))
                        }
                        .foregroundStyle(.white.opacity(0.9))
                        .padding(.horizontal, 8)
                        .padding(.vertical, 4)
                        .background(Color.white.opacity(0.1))
                        .clipShape(Capsule())
                    }
                    .buttonStyle(.plain)
                }
            }

            if let task = currentTask {
                Text(task.title)
                    .font(.system(size: 15, weight: .bold))
                    .lineLimit(1)
                    .foregroundStyle(.white)

                let subtasks = (task.subtasks ?? []).prefix(5)
                let allChecked = !(task.subtasks ?? []).isEmpty && (task.subtasks ?? []).allSatisfy { $0.isCompleted }

                VStack(alignment: .leading, spacing: 6) {
                    ForEach(subtasks) { st in
                        subtaskRow(taskId: task.id, subtask: st, isCompact: false)
                    }
                }

                if allChecked {
                    Button(intent: UpdateTaskStatusIntent(taskId: task.id, newStatus: "done")) {
                        HStack(spacing: 4) {
                            Image(systemName: "checkmark.circle.fill")
                                .font(.system(size: 12, weight: .bold))
                            Text("Selesaikan Tugas Sekarang")
                                .font(.system(size: 12, weight: .bold))
                        }
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 7)
                        .background(Color(red: 16/255, green: 185/255, blue: 129/255))
                        .foregroundStyle(.white)
                        .clipShape(RoundedRectangle(cornerRadius: 8))
                    }
                    .buttonStyle(.plain)
                }
            } else {
                emptyTasksView(isCompact: false)
            }

            Spacer(minLength: 0)
        }
        .padding(16)
        .widgetBackground(accentColor: Color(red: 168/255, green: 85/255, blue: 247/255))
    }

    // MARK: - Row & States

    @ViewBuilder
    private func subtaskRow(taskId: Int64, subtask: Subtask, isCompact: Bool) -> some View {
        HStack(spacing: 8) {
            Button(intent: ToggleSubtaskIntent(taskId: taskId, subtaskId: subtask.id, isCompleted: !subtask.isCompleted)) {
                Image(systemName: subtask.isCompleted ? "checkmark.circle.fill" : "circle")
                    .font(.system(size: isCompact ? 13 : 15, weight: .semibold))
                    .foregroundStyle(subtask.isCompleted ? Color(red: 52/255, green: 211/255, blue: 153/255) : .white.opacity(0.45))
            }
            .buttonStyle(.plain)

            Text(subtask.title)
                .font(.system(size: isCompact ? 11 : 12, weight: subtask.isCompleted ? .regular : .medium))
                .lineLimit(1)
                .strikethrough(subtask.isCompleted, color: .white.opacity(0.35))
                .foregroundStyle(subtask.isCompleted ? .white.opacity(0.4) : .white)

            Spacer()
        }
        .padding(.horizontal, 8)
        .padding(.vertical, isCompact ? 4 : 5)
        .background(Color.white.opacity(0.04))
        .clipShape(RoundedRectangle(cornerRadius: 7, style: .continuous))
    }

    @ViewBuilder
    private func emptyTasksView(isCompact: Bool) -> some View {
        VStack(spacing: isCompact ? 4 : 6) {
            Image(systemName: "checkmark.seal.fill")
                .font(.system(size: isCompact ? 20 : 26))
                .foregroundStyle(Color(red: 52/255, green: 211/255, blue: 153/255))

            Text("Semua Tugas Selesai!")
                .font(.system(size: isCompact ? 11 : 12, weight: .semibold))
                .foregroundStyle(.white)

            if !isCompact {
                Text("Tambahkan tugas baru melalui aplikasi.")
                    .font(.system(size: 9))
                    .foregroundStyle(.white.opacity(0.5))
            }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }

    private func priorityColor(_ p: String) -> Color {
        switch p.lowercased() {
        case "urgent": return Color(red: 239/255, green: 68/255, blue: 68/255)
        case "high": return Color(red: 245/255, green: 158/255, blue: 11/255)
        case "medium": return Color(red: 59/255, green: 130/255, blue: 246/255)
        default: return Color(red: 148/255, green: 163/255, blue: 184/255)
        }
    }
}

// MARK: - 3. Legacy ChecklistWidgetView Alias (Compatibility)
public struct ChecklistWidgetView: View {
    public let entry: TaskWidgetEntry
    public init(entry: TaskWidgetEntry) {
        self.entry = entry
    }
    public var body: some View {
        RecurringTaskWidgetView(entry: entry)
    }
}

// MARK: - Premium Glassmorphic Widget Background
public extension View {
    func widgetBackground(accentColor: Color = Color(red: 56/255, green: 189/255, blue: 248/255)) -> some View {
        let content = ZStack {
            // Deep obsidian gradient base
            LinearGradient(
                colors: [
                    Color(red: 11/255, green: 17/255, blue: 32/255),
                    Color(red: 15/255, green: 23/255, blue: 42/255),
                    Color(red: 8/255, green: 12/255, blue: 22/255)
                ],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
            )

            // Subtle ambient neon glow at top right
            GeometryReader { geo in
                Circle()
                    .fill(accentColor.opacity(0.12))
                    .blur(radius: 28)
                    .frame(width: geo.size.width * 0.7, height: geo.size.width * 0.7)
                    .position(x: geo.size.width * 0.85, y: geo.size.height * 0.15)
            }
        }

        #if canImport(WidgetKit)
        if #available(iOSApplicationExtension 17.0, iOS 17.0, *) {
            return self.containerBackground(for: .widget) {
                content
            }
        }
        #endif

        return self.background(content)
    }
}
