import SwiftUI

// MARK: - GraphQL Models
public struct GraphQLTaskItem: Codable, Identifiable, Sendable {
    public let id: Int64
    public let title: String
    public let description: String?
    public let status: String
    public let priority: String
    public let dueDate: String?
    public let isFocus: Bool?
    public let progress: Int?
    public let tags: [String]?
    public let project: GraphQLProjectItem?

    enum CodingKeys: String, CodingKey {
        case id
        case title
        case description
        case status
        case priority
        case dueDate = "due_date"
        case isFocus = "is_focus"
        case progress
        case tags
        case project
    }
}

public struct GraphQLProjectItem: Codable, Sendable {
    public let id: Int64
    public let name: String
    public let color: String?
}

public struct GraphQLReminderItem: Codable, Identifiable, Sendable {
    public let id: Int64
    public let title: String
    public let description: String?
    public let remindAt: String
    public let status: String
    public let taskId: Int64?
    public let taskTitle: String?
    public let snoozeCount: Int?
    public let createdAt: String?

    enum CodingKeys: String, CodingKey {
        case id
        case title
        case description
        case remindAt = "remind_at"
        case status
        case taskId = "task_id"
        case taskTitle = "task_title"
        case snoozeCount = "snooze_count"
        case createdAt = "created_at"
    }
}

public struct GraphQLActivityItem: Codable, Identifiable, Sendable {
    public let id: Int64
    public let taskId: Int64
    public let taskTitle: String?
    public let activityType: String
    public let description: String
    public let createdAt: String

    enum CodingKeys: String, CodingKey {
        case id
        case taskId = "task_id"
        case taskTitle = "task_title"
        case activityType = "activity_type"
        case description
        case createdAt = "created_at"
    }
}

public struct CalendarGraphQLData: Codable, Sendable {
    public let tasks: [GraphQLTaskItem]
    public let reminders: [GraphQLReminderItem]
    public let activities: [GraphQLActivityItem]
}

public enum CalendarFilter: String, CaseIterable, Identifiable {
    case all
    case tasks
    case reminders
    case activities

    public var id: String { rawValue }

    public var title: String {
        switch self {
        case .all: return L10n.Calendar.all
        case .tasks: return L10n.Calendar.tasks
        case .reminders: return L10n.Calendar.reminders
        case .activities: return L10n.Calendar.activities
        }
    }
}

// MARK: - View Model
@Observable
public final class CalendarViewModel {
    public var selectedDate: Date = Date()
    public var filter: CalendarFilter = .all
    public var tasks: [GraphQLTaskItem] = []
    public var reminders: [GraphQLReminderItem] = []
    public var activities: [GraphQLActivityItem] = []
    public var isLoading: Bool = false
    public var errorMessage: String? = nil

    private let apiClient = APIClient.shared

    private static let dayFormatter: DateFormatter = {
        let fmt = DateFormatter()
        fmt.dateFormat = "yyyy-MM-dd"
        return fmt
    }()

    private static let fallbackIsoFormatter: ISO8601DateFormatter = {
        let fmt = ISO8601DateFormatter()
        fmt.formatOptions = [.withInternetDateTime]
        return fmt
    }()

    public init() {
        Task {
            await loadData()
        }
    }

    private func dateString(from date: Date) -> String {
        CalendarViewModel.dayFormatter.string(from: date)
    }

    public var selectedDateString: String {
        dateString(from: selectedDate)
    }

    public var tasksForSelectedDate: [GraphQLTaskItem] {
        let target = selectedDateString
        return tasks.filter { task in
            guard let due = task.dueDate, due.count >= 10 else { return false }
            return String(due.prefix(10)) == target
        }
    }

    public var remindersForSelectedDate: [GraphQLReminderItem] {
        let target = selectedDateString
        return reminders.filter { rem in
            guard rem.remindAt.count >= 10 else { return false }
            return String(rem.remindAt.prefix(10)) == target
        }
    }

    public var activitiesForSelectedDate: [GraphQLActivityItem] {
        let target = selectedDateString
        return activities.filter { act in
            guard act.createdAt.count >= 10 else { return false }
            return String(act.createdAt.prefix(10)) == target
        }
    }

    public func hasTasks(on date: Date) -> Bool {
        let key = dateString(from: date)
        return tasks.contains { task in
            guard let due = task.dueDate, due.count >= 10 else { return false }
            return String(due.prefix(10)) == key
        }
    }

    public func hasReminders(on date: Date) -> Bool {
        let key = dateString(from: date)
        return reminders.contains { rem in
            guard rem.remindAt.count >= 10 else { return false }
            return String(rem.remindAt.prefix(10)) == key
        }
    }

    public func hasActivities(on date: Date) -> Bool {
        let key = dateString(from: date)
        return activities.contains { act in
            guard act.createdAt.count >= 10 else { return false }
            return String(act.createdAt.prefix(10)) == key
        }
    }

    public var monthTitle: String {
        let fmt = DateFormatter()
        fmt.dateFormat = "LLLL yyyy"
        return fmt.string(from: selectedDate).capitalized
    }

    public var weekdaySymbols: [String] {
        DateFormatter().shortStandaloneWeekdaySymbols ?? [
            "Min", "Sen", "Sel", "Rab", "Kam", "Jum", "Sab",
        ]
    }

    public var currentMonthDates: [Date] {
        let cal = Calendar.current
        guard let monthInterval = cal.dateInterval(of: .month, for: selectedDate),
            let firstWeek = cal.dateInterval(of: .weekOfMonth, for: monthInterval.start)
        else {
            return []
        }

        var dates: [Date] = []
        var current = firstWeek.start
        while dates.count < 42 {
            dates.append(current)
            guard let next = cal.date(byAdding: .day, value: 1, to: current) else { break }
            current = next
            if dates.count >= 35 && current >= monthInterval.end && (dates.count % 7 == 0) {
                break
            }
        }
        return dates
    }

    @MainActor
    public func nextMonth() {
        if let next = Calendar.current.date(byAdding: .month, value: 1, to: selectedDate) {
            onDateChanged(to: next)
        }
    }

    @MainActor
    public func prevMonth() {
        if let prev = Calendar.current.date(byAdding: .month, value: -1, to: selectedDate) {
            onDateChanged(to: prev)
        }
    }

    public func isSameMonth(_ date1: Date, _ date2: Date) -> Bool {
        Calendar.current.isDate(date1, equalTo: date2, toGranularity: .month)
    }

    public func isSameDay(_ date1: Date, _ date2: Date) -> Bool {
        Calendar.current.isDate(date1, equalTo: date2, toGranularity: .day)
    }

    private var loadDataTask: Task<Void, Never>?

    @MainActor
    public func onDateChanged(to newDate: Date) {
        selectedDate = newDate

        // Cancel previous in-flight task to avoid over-calling and race conditions
        loadDataTask?.cancel()
        loadDataTask = Task {
            // 250ms debounce for rapid date clicking
            try? await Task.sleep(nanoseconds: 250_000_000)
            guard !Task.isCancelled else { return }
            await loadData()
        }
    }

    @MainActor
    public func loadData() async {
        guard !Task.isCancelled else { return }
        isLoading = true
        errorMessage = nil

        let cal = Calendar.current
        let startOfMonth = cal.date(byAdding: .day, value: -14, to: selectedDate) ?? selectedDate
        let endOfMonth = cal.date(byAdding: .day, value: 35, to: selectedDate) ?? selectedDate

        let fromStr = CalendarViewModel.fallbackIsoFormatter.string(from: startOfMonth)
        let toStr = CalendarViewModel.fallbackIsoFormatter.string(from: endOfMonth)

        let query = """
            query GetCalendarData($from: String, $to: String) {
              tasks {
                id
                title
                description
                status
                priority
                due_date
                is_focus
                progress
                tags
                project {
                  id
                  name
                  color
                }
              }
              reminders {
                id
                title
                description
                remind_at
                status
                task_id
                task_title
                snooze_count
                created_at
              }
              activities(from: $from, to: $to) {
                id
                task_id
                task_title
                activity_type
                description
                created_at
              }
            }
            """

        let variables: [String: Any] = [
            "from": fromStr,
            "to": toStr,
        ]

        do {
            let data: CalendarGraphQLData = try await apiClient.graphql(
                query: query, variables: variables)
            guard !Task.isCancelled else { return }
            self.tasks = data.tasks
            self.reminders = data.reminders
            self.activities = data.activities
            AppLogger.info(
                "Calendar GraphQL data loaded", category: .network,
                metadata: [
                    "tasks": data.tasks.count,
                    "reminders": data.reminders.count,
                    "activities": data.activities.count,
                ])
        } catch {
            guard !Task.isCancelled else { return }
            AppLogger.error(
                "Failed to load Calendar GraphQL data", error: error, category: .network)
            self.errorMessage = error.localizedDescription
        }

        if !Task.isCancelled {
            self.isLoading = false
        }
    }

    @MainActor
    public func snoozeReminder(id: Int64, minutes: Int = 60) async {
        do {
            _ = try await apiClient.snoozeReminder(id: id, minutes: minutes)
            await loadData()
        } catch {
            self.errorMessage = error.localizedDescription
        }
    }

    @MainActor
    public func dismissReminder(id: Int64) async {
        do {
            _ = try await apiClient.dismissReminder(id: id)
            await loadData()
        } catch {
            self.errorMessage = error.localizedDescription
        }
    }
}

// MARK: - Calendar View
public struct CalendarView: View {
    @State private var viewModel = CalendarViewModel()

    public init() {}

    public var body: some View {
        NavigationStack {
            ScrollView {
                VStack(spacing: 16) {
                    // Header Banner with Single-Query GraphQL indicator
                    headerCard

                    // Graphical Calendar Picker
                    calendarSection

                    // Filter Segmented Control
                    filterSegment

                    // Agenda Content for Selected Date
                    agendaSection
                }
                .padding(.horizontal, 16)
                .padding(.bottom, 32)
            }
            .background(Color(uiColor: .systemGroupedBackground))
            .navigationTitle(L10n.Calendar.title)
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Button {
                        Task { await viewModel.loadData() }
                    } label: {
                        if viewModel.isLoading {
                            ProgressView()
                                .tint(AppColors.primary)
                        } else {
                            Image(systemName: "arrow.clockwise")
                                .font(.subheadline)
                        }
                    }
                }
            }
            .refreshable {
                await viewModel.loadData()
            }
        }
    }

    // MARK: - Header Card
    private var headerCard: some View {
        HStack(spacing: 12) {
            Image(systemName: "calendar.badge.clock")
                .font(.system(size: 26))
                .foregroundStyle(AppColors.primary)
                .frame(width: 44, height: 44)
                .background(AppColors.primary.opacity(0.12))
                .clipShape(RoundedRectangle(cornerRadius: 12))

            VStack(alignment: .leading, spacing: 2) {
                HStack(spacing: 6) {
                    Text(L10n.Calendar.title)
                        .font(.headline)
                        .foregroundStyle(Color.primary)

                    HStack(spacing: 3) {
                        Circle()
                            .fill(Color.green)
                            .frame(width: 6, height: 6)
                        Text(L10n.Calendar.graphqlPowered)
                            .font(.system(size: 9, weight: .semibold, design: .monospaced))
                            .foregroundStyle(Color.green)
                    }
                    .padding(.horizontal, 6)
                    .padding(.vertical, 2)
                    .background(Color.green.opacity(0.1))
                    .clipShape(Capsule())
                }

                Text(L10n.Calendar.subtitle)
                    .font(.caption)
                    .foregroundStyle(Color.secondary)
            }

            Spacer()
        }
        .padding(12)
        .background(Color(uiColor: .secondarySystemGroupedBackground))
        .clipShape(RoundedRectangle(cornerRadius: 14))
        .overlay(
            RoundedRectangle(cornerRadius: 14)
                .stroke(Color(uiColor: .separator).opacity(0.4), lineWidth: 1)
        )
    }

    // MARK: - Calendar Section (Responsive Month Grid with Multi-Dot Indicators)
    private var calendarSection: some View {
        VStack(spacing: 12) {
            // Month Header with Prev/Next and Today button
            HStack {
                VStack(alignment: .leading, spacing: 2) {
                    Text(viewModel.monthTitle)
                        .font(.headline.weight(.bold))
                        .foregroundStyle(Color.primary)
                }

                Spacer()

                HStack(spacing: 8) {
                    Button {
                        withAnimation(.easeInOut(duration: 0.2)) {
                            viewModel.prevMonth()
                        }
                    } label: {
                        Image(systemName: "chevron.left")
                            .font(.system(size: 13, weight: .semibold))
                            .foregroundStyle(Color.primary)
                            .frame(width: 32, height: 32)
                            .background(Color(uiColor: .tertiarySystemFill))
                            .clipShape(Circle())
                    }

                    Button {
                        withAnimation(.easeInOut(duration: 0.2)) {
                            viewModel.onDateChanged(to: Date())
                        }
                    } label: {
                        Text(L10n.Calendar.today)
                            .font(.caption2.weight(.semibold))
                            .foregroundStyle(AppColors.primary)
                            .padding(.horizontal, 8)
                            .padding(.vertical, 6)
                            .background(AppColors.primary.opacity(0.12))
                            .clipShape(Capsule())
                    }

                    Button {
                        withAnimation(.easeInOut(duration: 0.2)) {
                            viewModel.nextMonth()
                        }
                    } label: {
                        Image(systemName: "chevron.right")
                            .font(.system(size: 13, weight: .semibold))
                            .foregroundStyle(Color.primary)
                            .frame(width: 32, height: 32)
                            .background(Color(uiColor: .tertiarySystemFill))
                            .clipShape(Circle())
                    }
                }
            }

            // Weekday Headers
            HStack(spacing: 0) {
                ForEach(viewModel.weekdaySymbols, id: \.self) { day in
                    Text(day)
                        .font(.caption2.weight(.semibold))
                        .foregroundStyle(Color.secondary)
                        .frame(maxWidth: .infinity)
                }
            }
            .padding(.top, 4)

            // Month Dates Grid
            LazyVGrid(
                columns: Array(repeating: GridItem(.flexible(), spacing: 4), count: 7),
                spacing: 8
            ) {
                ForEach(viewModel.currentMonthDates, id: \.self) { date in
                    dayCell(for: date)
                }
            }

            // Category Dots Legend
            HStack(spacing: 12) {
                legendItem(color: AppColors.primary, label: L10n.Calendar.tasks)
                legendItem(color: Color.orange, label: L10n.Calendar.reminders)
                legendItem(color: Color.green, label: L10n.Calendar.activities)
                Spacer()
            }
            .padding(.top, 6)
            .padding(.horizontal, 4)
        }
        .padding(14)
        .background(Color(uiColor: .secondarySystemGroupedBackground))
        .clipShape(RoundedRectangle(cornerRadius: 16))
        .overlay(
            RoundedRectangle(cornerRadius: 16)
                .stroke(Color(uiColor: .separator).opacity(0.4), lineWidth: 1)
        )
    }

    private func legendItem(color: Color, label: String) -> some View {
        HStack(spacing: 4) {
            Circle()
                .fill(color)
                .frame(width: 6, height: 6)
            Text(label)
                .font(.system(size: 10, weight: .medium))
                .foregroundStyle(Color.secondary)
        }
    }

    private func dayCell(for date: Date) -> some View {
        let isSelected = viewModel.isSameDay(date, viewModel.selectedDate)
        let isToday = viewModel.isSameDay(date, Date())
        let isCurrentMonth = viewModel.isSameMonth(date, viewModel.selectedDate)
        let dayNum = Calendar.current.component(.day, from: date)

        let hasTasks = viewModel.hasTasks(on: date)
        let hasReminders = viewModel.hasReminders(on: date)
        let hasActivities = viewModel.hasActivities(on: date)

        return Button {
            withAnimation(.spring(response: 0.25, dampingFraction: 0.7)) {
                viewModel.onDateChanged(to: date)
            }
        } label: {
            VStack(spacing: 4) {
                Text("\(dayNum)")
                    .font(.system(size: 14, weight: isSelected || isToday ? .bold : .regular))
                    .foregroundStyle(
                        isSelected
                            ? Color.white
                            : (isToday
                                ? AppColors.primary
                                : (isCurrentMonth ? Color.primary : Color.secondary.opacity(0.35)))
                    )
                    .frame(width: 32, height: 32)
                    .background(
                        Group {
                            if isSelected {
                                Circle().fill(AppColors.primary)
                            } else if isToday {
                                Circle()
                                    .stroke(AppColors.primary.opacity(0.5), lineWidth: 1.5)
                            } else {
                                Color.clear
                            }
                        }
                    )

                // Multi-dot indicators: side-by-side (Tasks 🔵, Reminders 🟠, Activities 🟢)
                HStack(spacing: 3) {
                    if hasTasks {
                        Circle()
                            .fill(isSelected ? Color.white : AppColors.primary)
                            .frame(width: 4, height: 4)
                    }
                    if hasReminders {
                        Circle()
                            .fill(isSelected ? Color.orange.opacity(0.9) : Color.orange)
                            .frame(width: 4, height: 4)
                    }
                    if hasActivities {
                        Circle()
                            .fill(isSelected ? Color.green.opacity(0.9) : Color.green)
                            .frame(width: 4, height: 4)
                    }
                }
                .frame(height: 5)
            }
            .frame(maxWidth: .infinity)
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
    }

    // MARK: - Filter Segment
    private var filterSegment: some View {
        Picker("Filter", selection: $viewModel.filter) {
            ForEach(CalendarFilter.allCases) { filter in
                Text(filter.title).tag(filter)
            }
        }
        .pickerStyle(.segmented)
    }

    // MARK: - Agenda Section
    private var agendaSection: some View {
        VStack(alignment: .leading, spacing: 14) {
            // Section Header
            HStack {
                Text(formattedSelectedDate)
                    .font(.subheadline.weight(.bold))
                    .foregroundStyle(Color.primary)

                Spacer()

                HStack(spacing: 8) {
                    chipCount(
                        count: viewModel.tasksForSelectedDate.count, color: .blue,
                        icon: "checkmark.square")
                    chipCount(
                        count: viewModel.remindersForSelectedDate.count, color: .orange,
                        icon: "bell")
                    chipCount(
                        count: viewModel.activitiesForSelectedDate.count, color: .green,
                        icon: "clock.arrow.circlepath")
                }
            }

            if viewModel.filter == .all || viewModel.filter == .tasks {
                tasksGroup
            }

            if viewModel.filter == .all || viewModel.filter == .reminders {
                remindersGroup
            }

            if viewModel.filter == .all || viewModel.filter == .activities {
                activitiesGroup
            }

            if isAllEmpty {
                emptyDayView
            }
        }
    }

    private var formattedSelectedDate: String {
        let fmt = DateFormatter()
        fmt.dateStyle = .full
        return fmt.string(from: viewModel.selectedDate)
    }

    private var isAllEmpty: Bool {
        switch viewModel.filter {
        case .all:
            return viewModel.tasksForSelectedDate.isEmpty
                && viewModel.remindersForSelectedDate.isEmpty
                && viewModel.activitiesForSelectedDate.isEmpty
        case .tasks:
            return viewModel.tasksForSelectedDate.isEmpty
        case .reminders:
            return viewModel.remindersForSelectedDate.isEmpty
        case .activities:
            return viewModel.activitiesForSelectedDate.isEmpty
        }
    }

    private func chipCount(count: Int, color: Color, icon: String) -> some View {
        HStack(spacing: 3) {
            Image(systemName: icon)
                .font(.system(size: 9))
            Text("\(count)")
                .font(.system(size: 10, weight: .bold, design: .monospaced))
        }
        .foregroundStyle(color)
        .padding(.horizontal, 6)
        .padding(.vertical, 2)
        .background(color.opacity(0.12))
        .clipShape(Capsule())
    }

    // MARK: - Tasks Group
    private var tasksGroup: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack(spacing: 6) {
                Image(systemName: "checkmark.square.fill")
                    .font(.caption)
                    .foregroundStyle(Color.blue)
                Text(L10n.Calendar.tasksDue)
                    .font(.caption.weight(.bold))
                    .foregroundStyle(Color.secondary)
                Text("(\(viewModel.tasksForSelectedDate.count))")
                    .font(.caption2.monospaced())
                    .foregroundStyle(Color.secondary)
            }

            if viewModel.tasksForSelectedDate.isEmpty {
                Text(L10n.Calendar.noTasks)
                    .font(.caption)
                    .foregroundStyle(Color.secondary.opacity(0.7))
                    .padding(.leading, 8)
            } else {
                ForEach(viewModel.tasksForSelectedDate) { task in
                    HStack(alignment: .center, spacing: 10) {
                        Circle()
                            .fill(Color.blue)
                            .frame(width: 8, height: 8)

                        VStack(alignment: .leading, spacing: 2) {
                            Text(task.title)
                                .font(.subheadline.weight(.medium))
                                .foregroundStyle(Color.primary)

                            if let proj = task.project {
                                ProjectPill(name: proj.name, hexColor: proj.color)
                            }
                        }

                        Spacer()

                        PriorityBadge(priority: task.priority)
                    }
                    .padding(10)
                    .background(Color(uiColor: .secondarySystemGroupedBackground))
                    .clipShape(RoundedRectangle(cornerRadius: 10))
                    .overlay(
                        RoundedRectangle(cornerRadius: 10)
                            .stroke(Color(uiColor: .separator).opacity(0.4), lineWidth: 1)
                    )
                }
            }
        }
    }

    // MARK: - Reminders Group
    private var remindersGroup: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack(spacing: 6) {
                Image(systemName: "bell.fill")
                    .font(.caption)
                    .foregroundStyle(Color.orange)
                Text(L10n.Calendar.scheduledReminders)
                    .font(.caption.weight(.bold))
                    .foregroundStyle(Color.secondary)
                Text("(\(viewModel.remindersForSelectedDate.count))")
                    .font(.caption2.monospaced())
                    .foregroundStyle(Color.secondary)
            }

            if viewModel.remindersForSelectedDate.isEmpty {
                Text(L10n.Calendar.noReminders)
                    .font(.caption)
                    .foregroundStyle(Color.secondary.opacity(0.7))
                    .padding(.leading, 8)
            } else {
                ForEach(viewModel.remindersForSelectedDate) { rem in
                    HStack(spacing: 10) {
                        Circle()
                            .fill(Color.orange)
                            .frame(width: 8, height: 8)

                        VStack(alignment: .leading, spacing: 2) {
                            Text(rem.title)
                                .font(.subheadline.weight(.medium))
                                .foregroundStyle(Color.primary)

                            if let taskTitle = rem.taskTitle {
                                Text("🔗 \(taskTitle)")
                                    .font(.caption2)
                                    .foregroundStyle(Color.orange)
                            }

                            if rem.remindAt.count >= 16 {
                                Text(
                                    "⏰ \(rem.remindAt.suffix(from: rem.remindAt.index(rem.remindAt.startIndex, offsetBy: 11)).prefix(5))"
                                )
                                .font(.caption2)
                                .foregroundStyle(Color.secondary)
                            }
                        }

                        Spacer()

                        HStack(spacing: 6) {
                            Button {
                                Task { await viewModel.snoozeReminder(id: rem.id, minutes: 60) }
                            } label: {
                                Text(L10n.Calendar.snooze1h)
                                    .font(.system(size: 10, weight: .semibold))
                                    .foregroundStyle(AppColors.primary)
                                    .padding(.horizontal, 6)
                                    .padding(.vertical, 3)
                                    .background(AppColors.primary.opacity(0.1))
                                    .clipShape(RoundedRectangle(cornerRadius: 6))
                            }

                            Button {
                                Task { await viewModel.dismissReminder(id: rem.id) }
                            } label: {
                                Image(systemName: "checkmark")
                                    .font(.system(size: 10, weight: .semibold))
                                    .foregroundStyle(Color.green)
                                    .padding(6)
                                    .background(Color.green.opacity(0.1))
                                    .clipShape(Circle())
                            }
                        }
                    }
                    .padding(10)
                    .background(Color(uiColor: .secondarySystemGroupedBackground))
                    .clipShape(RoundedRectangle(cornerRadius: 10))
                    .overlay(
                        RoundedRectangle(cornerRadius: 10)
                            .stroke(Color(uiColor: .separator).opacity(0.4), lineWidth: 1)
                    )
                }
            }
        }
    }

    // MARK: - Activities Group (Work Log / History)
    private var activitiesGroup: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack(spacing: 6) {
                Image(systemName: "clock.arrow.circlepath")
                    .font(.caption)
                    .foregroundStyle(Color.green)
                Text(L10n.Calendar.workLog)
                    .font(.caption.weight(.bold))
                    .foregroundStyle(Color.secondary)
                Text("(\(viewModel.activitiesForSelectedDate.count))")
                    .font(.caption2.monospaced())
                    .foregroundStyle(Color.secondary)
            }

            if viewModel.activitiesForSelectedDate.isEmpty {
                Text(L10n.Calendar.noActivities)
                    .font(.caption)
                    .foregroundStyle(Color.secondary.opacity(0.7))
                    .padding(.leading, 8)
            } else {
                VStack(spacing: 10) {
                    ForEach(viewModel.activitiesForSelectedDate) { act in
                        HStack(alignment: .top, spacing: 10) {
                            Circle()
                                .fill(Color.green)
                                .frame(width: 8, height: 8)
                                .padding(.top, 4)

                            VStack(alignment: .leading, spacing: 2) {
                                HStack {
                                    if let tTitle = act.taskTitle {
                                        Text(tTitle)
                                            .font(.caption.weight(.semibold))
                                            .foregroundStyle(Color.primary)
                                    }
                                    Spacer()
                                    if act.createdAt.count >= 16 {
                                        Text(String(act.createdAt.dropFirst(11).prefix(5)))
                                            .font(.caption2.monospaced())
                                            .foregroundStyle(Color.secondary)
                                    }
                                }

                                Text(act.description)
                                    .font(.caption)
                                    .foregroundStyle(Color.secondary)
                            }
                        }
                        .padding(10)
                        .background(Color(uiColor: .secondarySystemGroupedBackground))
                        .clipShape(RoundedRectangle(cornerRadius: 10))
                        .overlay(
                            RoundedRectangle(cornerRadius: 10)
                                .stroke(Color(uiColor: .separator).opacity(0.4), lineWidth: 1)
                        )
                    }
                }
            }
        }
    }

    // MARK: - Empty State
    private var emptyDayView: some View {
        VStack(spacing: 8) {
            Image(systemName: "calendar.badge.exclamationmark")
                .font(.system(size: 32))
                .foregroundStyle(Color.secondary.opacity(0.5))
            Text(L10n.Calendar.noItems)
                .font(.subheadline)
                .foregroundStyle(Color.secondary)
        }
        .frame(maxWidth: .infinity)
        .padding(.vertical, 32)
    }
}
