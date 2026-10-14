import Foundation
import SwiftUI
#if canImport(WidgetKit)
import WidgetKit
#endif

public struct TaskWidgetEntry: TimelineEntry, Sendable {
    public let date: Date
    public let bundle: WidgetBundleResponse
    public let isOffline: Bool

    public init(date: Date = Date(), bundle: WidgetBundleResponse, isOffline: Bool = false) {
        self.date = date
        self.bundle = bundle
        self.isOffline = isOffline
    }

    public static var sample: TaskWidgetEntry {
        let sampleTask1 = TaskItem(
            id: 1,
            userId: "sample",
            title: "Desain Sistem Widget iOS",
            description: "Implementasi widget interaktif",
            status: "in_progress",
            priority: "urgent",
            dueDate: "2026-10-14T17:00:00Z",
            isFocus: true,
            tags: ["ios", "ui"],
            projectId: 1,
            projectName: "Mobile App",
            projectColor: "#3b82f6",
            lastProgressNote: "Model dan AppIntents selesai",
            subtasks: [
                Subtask(id: 101, taskId: 1, title: "Buat checklist interaktif", isCompleted: true, orderIndex: 0),
                Subtask(id: 102, taskId: 1, title: "Hubungkan App Group cache", isCompleted: true, orderIndex: 1),
                Subtask(id: 103, taskId: 1, title: "Pasang tombol centang offline", isCompleted: false, orderIndex: 2)
            ],
            subtasksCount: 3,
            subtasksCompleted: 2,
            progressPercent: 67,
            isRecurring: true,
            recurrencePattern: "daily",
            recurrenceInterval: 1,
            recurrenceEndDate: nil,
            recurrenceLastGeneratedAt: nil
        )

        let sampleTask2 = TaskItem(
            id: 2,
            userId: "sample",
            title: "Review PR Backend Recurrence",
            description: "Pemeriksaan rollover asynq",
            status: "todo",
            priority: "high",
            dueDate: "2026-10-14T21:00:00Z",
            isFocus: true,
            tags: ["backend"],
            projectId: 2,
            projectName: "Backend",
            projectColor: "#10b981",
            lastProgressNote: nil,
            subtasks: [
                Subtask(id: 104, taskId: 2, title: "Review migrasi database", isCompleted: true, orderIndex: 0),
                Subtask(id: 105, taskId: 2, title: "Cek endpoint trigger", isCompleted: false, orderIndex: 1)
            ],
            subtasksCount: 2,
            subtasksCompleted: 1,
            progressPercent: 50,
            isRecurring: false,
            recurrencePattern: nil,
            recurrenceInterval: nil,
            recurrenceEndDate: nil,
            recurrenceLastGeneratedAt: nil
        )

        let bundle = WidgetBundleResponse(
            focusTasks: [sampleTask1, sampleTask2],
            upcomingDeadlines: [sampleTask1, sampleTask2],
            completedToday: 4,
            totalToday: 6
        )

        return TaskWidgetEntry(date: Date(), bundle: bundle, isOffline: false)
    }

    public static var empty: TaskWidgetEntry {
        TaskWidgetEntry(
            date: Date(),
            bundle: WidgetBundleResponse(
                focusTasks: [],
                upcomingDeadlines: [],
                completedToday: 0,
                totalToday: 0
            ),
            isOffline: false
        )
    }
}

#if canImport(WidgetKit)
public struct TaskWidgetTimelineProvider: TimelineProvider {
    public typealias Entry = TaskWidgetEntry

    public init() {}

    public func placeholder(in context: Context) -> TaskWidgetEntry {
        TaskWidgetEntry.sample
    }

    public func getSnapshot(in context: Context, completion: @escaping (TaskWidgetEntry) -> Void) {
        if let cached = SharedWidgetDataManager.shared.loadWidgetBundle() {
            completion(TaskWidgetEntry(date: Date(), bundle: cached, isOffline: false))
        } else {
            completion(TaskWidgetEntry.sample)
        }
    }

    public func getTimeline(in context: Context, completion: @escaping (Timeline<TaskWidgetEntry>) -> Void) {
        let bundle = SharedWidgetDataManager.shared.loadWidgetBundle() ?? TaskWidgetEntry.empty.bundle
        let entry = TaskWidgetEntry(date: Date(), bundle: bundle, isOffline: false)

        // Reload timeline every 15 minutes
        let nextUpdate = Calendar.current.date(byAdding: .minute, value: 15, to: Date()) ?? Date().addingTimeInterval(900)
        let timeline = Timeline(entries: [entry], policy: .after(nextUpdate))
        completion(timeline)
    }
}
#endif
