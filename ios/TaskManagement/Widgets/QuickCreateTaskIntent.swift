import Foundation
import AppIntents

public struct QuickCreateTaskIntent: AppIntent {
    public static var title: LocalizedStringResource = "Buat Tugas Cepat"
    public static var description = IntentDescription("Buat tugas baru secara instan dari Pintasan (Shortcuts), Pusat Kontrol, atau Otomasi.")

    public static var parameterSummary: some ParameterSummary {
        Summary("Catat \(\.$title)")
    }

    @Parameter(title: "Judul Tugas", requestValueDialog: IntentDialog("Apa judul tugas yang ingin Anda catat?"))
    public var title: String

    @Parameter(title: "Jadikan Fokus Hari Ini", default: true)
    public var isFocus: Bool

    public init() {
        self.title = ""
        self.isFocus = true
    }

    public init(title: String, isFocus: Bool = true) {
        self.title = title
        self.isFocus = isFocus
    }

    public func perform() async throws -> some IntentResult & ReturnsValue<String> {
        let req = CreateTaskRequest(title: title, isFocus: isFocus)
        _ = try? await APIClient.shared.createTask(req: req)
        await SharedWidgetDataManager.shared.syncWithBackend()
        return .result(value: "Tugas '\(title)' berhasil dicatat!")
    }
}

public struct TaskHubShortcutsProvider: AppShortcutsProvider {
    public static var appShortcuts: [AppShortcut] {
        AppShortcut(
            intent: QuickCreateTaskIntent(),
            phrases: [
                "Buat tugas baru di \(.applicationName)",
                "Catat tugas di \(.applicationName)",
                "Tambah fokus hari ini di \(.applicationName)"
            ],
            shortTitle: "Buat Tugas",
            systemImageName: "plus.circle.fill"
        )
    }
}
