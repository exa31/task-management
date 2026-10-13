import SwiftUI

@Observable
public final class NoteViewModel {
    public var notes: [Note] = []
    public var isLoading = false
    public var errorMessage: String?
    public var successMessage: String?
    public var searchQuery: String = ""
    public var selectedType: String? = nil // nil = All, "actionable", "knowledge"

    private let apiClient = APIClient.shared

    public init() {
        Task { await loadNotes() }
    }

    @MainActor
    public func loadNotes() async {
        isLoading = true
        do {
            self.notes = try await apiClient.getNotes()
            AppLogger.info("Notes loaded successfully", category: .app, metadata: ["count": notes.count])
        } catch {
            AppLogger.error("Failed to load notes", error: error, category: .app)
            self.errorMessage = error.localizedDescription
        }
        self.isLoading = false
    }

    public var filteredNotes: [Note] {
        var result = notes

        let q = searchQuery.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
        if !q.isEmpty {
            result = result.filter {
                $0.title.lowercased().contains(q) ||
                $0.content.lowercased().contains(q) ||
                ($0.tags?.contains(where: { $0.lowercased().contains(q) }) ?? false)
            }
        }

        if let type = selectedType, !type.isEmpty {
            result = result.filter { $0.type == type }
        }

        // Pinned notes appear first
        return result.sorted { ($0.isPinned ? 1 : 0) > ($1.isPinned ? 1 : 0) }
    }

    @MainActor
    public func createNote(title: String, content: String, type: String = "actionable", isPinned: Bool = false) async -> Bool {
        do {
            _ = try await apiClient.createNote(
                req: CreateNoteRequest(title: title, content: content, type: type, color: "#6366F1", isPinned: isPinned)
            )
            AppLogger.success("Note created successfully", category: .app, metadata: ["type": type])
            await loadNotes()
            return true
        } catch {
            AppLogger.error("Failed to create note", error: error, category: .app)
            self.errorMessage = error.localizedDescription
            return false
        }
    }

    @MainActor
    public func updateNote(id: Int64, title: String, content: String, type: String) async -> Bool {
        do {
            _ = try await apiClient.updateNote(
                id: id,
                req: UpdateNoteRequest(title: title, content: content, type: type)
            )
            AppLogger.success("Note updated successfully", category: .app, metadata: ["id": id])
            await loadNotes()
            return true
        } catch {
            AppLogger.error("Failed to update note", error: error, category: .app)
            self.errorMessage = error.localizedDescription
            return false
        }
    }

    @MainActor
    public func togglePin(note: Note) {
        Task {
            do {
                _ = try await apiClient.updateNote(
                    id: note.id,
                    req: UpdateNoteRequest(isPinned: !note.isPinned)
                )
                AppLogger.success("Note pin toggled", category: .app, metadata: ["id": note.id])
                await loadNotes()
            } catch {
                AppLogger.error("Failed to toggle note pin", error: error, category: .app)
                self.errorMessage = error.localizedDescription
            }
        }
    }

    @MainActor
    public func deleteNote(id: Int64) {
        Task {
            do {
                try await apiClient.deleteNote(id: id)
                AppLogger.success("Note deleted successfully", category: .app, metadata: ["id": id])
                await loadNotes()
            } catch {
                AppLogger.error("Failed to delete note", error: error, category: .app)
                self.errorMessage = error.localizedDescription
            }
        }
    }

    @MainActor
    public func convertToTask(note: Note) async -> Bool {
        do {
            let res = try await apiClient.convertNoteToTask(id: note.id)
            AppLogger.success("Note converted to task", category: .app, metadata: ["noteId": note.id, "taskId": res.taskId ?? 0])
            self.successMessage = L10n.Notes.convertedSuccess
            await loadNotes()
            return true
        } catch {
            AppLogger.error("Failed to convert note to task", error: error, category: .app)
            self.errorMessage = error.localizedDescription
            return false
        }
    }
}

public struct NoteListView: View {
    @Bindable var viewModel: NoteViewModel

    @State private var showCreateSheet = false
    @State private var noteToEdit: Note? = nil
    @State private var noteToDelete: Note? = nil
    @State private var showDeleteConfirm = false
    @State private var noteToConvert: Note? = nil
    @State private var isConverting = false

    public init(viewModel: NoteViewModel) {
        self.viewModel = viewModel
    }

    public var body: some View {
        NavigationStack {
            VStack(spacing: 0) {
                // Search & Filter Header
                VStack(spacing: 10) {
                    // Search Bar
                    HStack(spacing: 8) {
                        Image(systemName: "magnifyingglass")
                            .foregroundColor(.secondary)
                            .font(.system(size: 14))

                        TextField(L10n.Notes.searchPlaceholder, text: $viewModel.searchQuery)
                            .font(.system(size: 14))

                        if !viewModel.searchQuery.isEmpty {
                            Button(action: { viewModel.searchQuery = "" }) {
                                Image(systemName: "xmark.circle.fill")
                                    .foregroundColor(.secondary)
                                    .font(.system(size: 14))
                            }
                        }
                    }
                    .padding(.horizontal, 12)
                    .padding(.vertical, 8)
                    .background(Color(.secondarySystemGroupedBackground))
                    .clipShape(RoundedRectangle(cornerRadius: 10))

                    // Type Filter Pills
                    ScrollView(.horizontal, showsIndicators: false) {
                        HStack(spacing: 8) {
                            FilterPill(
                                title: L10n.Notes.filterAll,
                                isSelected: viewModel.selectedType == nil,
                                count: viewModel.notes.count
                            ) {
                                viewModel.selectedType = nil
                            }

                            FilterPill(
                                title: "Actionable",
                                isSelected: viewModel.selectedType == "actionable",
                                count: viewModel.notes.filter { $0.type == "actionable" }.count
                            ) {
                                viewModel.selectedType = "actionable"
                            }

                            FilterPill(
                                title: "Knowledge",
                                isSelected: viewModel.selectedType == "knowledge",
                                count: viewModel.notes.filter { $0.type == "knowledge" }.count
                            ) {
                                viewModel.selectedType = "knowledge"
                            }
                        }
                    }
                }
                .padding(.horizontal, 16)
                .padding(.vertical, 10)
                .background(Color(.systemGroupedBackground))

                // Feedback Banner
                if let success = viewModel.successMessage {
                    HStack(spacing: 8) {
                        Image(systemName: "checkmark.circle.fill")
                            .foregroundColor(.emeraldGreen)
                        Text(success)
                            .font(.system(size: 13))
                            .foregroundColor(.emeraldGreen)
                        Spacer()
                        Button(action: { viewModel.successMessage = nil }) {
                            Image(systemName: "xmark")
                                .font(.system(size: 11))
                                .foregroundColor(.secondary)
                        }
                    }
                    .padding(.horizontal, 16)
                    .padding(.vertical, 8)
                    .background(Color.emeraldGreen.opacity(0.12))
                }

                if let err = viewModel.errorMessage {
                    HStack(spacing: 8) {
                        Image(systemName: "exclamationmark.triangle.fill")
                            .foregroundColor(.red)
                        Text(err)
                            .font(.system(size: 13))
                            .foregroundColor(.red)
                        Spacer()
                        Button(action: { viewModel.errorMessage = nil }) {
                            Image(systemName: "xmark")
                                .font(.system(size: 11))
                                .foregroundColor(.secondary)
                        }
                    }
                    .padding(.horizontal, 16)
                    .padding(.vertical, 8)
                    .background(Color.red.opacity(0.12))
                }

                // Main Content List
                Group {
                    if viewModel.isLoading && viewModel.notes.isEmpty {
                        VStack(spacing: 12) {
                            ProgressView()
                            Text(L10n.Notes.loading).font(.system(size: 13)).foregroundColor(.secondary)
                        }
                        .frame(maxHeight: .infinity)
                    } else if viewModel.filteredNotes.isEmpty {
                        VStack(spacing: 14) {
                            ZStack {
                                Circle()
                                    .fill(AppColors.primary.opacity(0.12))
                                    .frame(width: 72, height: 72)
                                Image(systemName: "note.text")
                                    .font(.system(size: 32))
                                    .foregroundColor(AppColors.primary)
                            }
                            Text(L10n.Notes.emptyTitle)
                                .font(.system(size: 17, weight: .bold))
                            Text(L10n.Notes.emptyDesc)
                                .font(.system(size: 13))
                                .foregroundColor(.secondary)
                                .multilineTextAlignment(.center)
                                .padding(.horizontal, 32)

                            Button(action: { showCreateSheet = true }) {
                                HStack(spacing: 6) {
                                    Image(systemName: "plus")
                                    Text(L10n.Notes.newNote)
                                }
                                .font(.system(size: 14, weight: .semibold))
                                .padding(.horizontal, 16)
                                .padding(.vertical, 10)
                                .background(AppColors.brandGradient)
                                .foregroundColor(.white)
                                .clipShape(Capsule())
                            }
                        }
                        .frame(maxHeight: .infinity)
                    } else {
                        List {
                            ForEach(viewModel.filteredNotes) { note in
                                NoteCardRow(
                                    note: note,
                                    onTogglePin: {
                                        viewModel.togglePin(note: note)
                                    },
                                    onEdit: {
                                        noteToEdit = note
                                    },
                                    onConvert: {
                                        noteToConvert = note
                                    },
                                    onDelete: {
                                        noteToDelete = note
                                        showDeleteConfirm = true
                                    }
                                )
                                .glassCardStyle()
                                .listRowSeparator(.hidden)
                                .listRowBackground(Color.clear)
                                .listRowInsets(EdgeInsets(top: 6, leading: 16, bottom: 6, trailing: 16))
                            }
                        }
                        .listStyle(.plain)
                        .refreshable {
                            await viewModel.loadNotes()
                        }
                    }
                }
            }
            .navigationTitle(L10n.Notes.title)
            .toolbar {
                ToolbarItem(placement: .primaryAction) {
                    Button(action: { showCreateSheet = true }) {
                        Image(systemName: "plus.circle.fill")
                            .font(.system(size: 20))
                            .foregroundStyle(AppColors.brandGradient)
                    }
                }
            }
            .sheet(isPresented: $showCreateSheet) {
                CreateNoteSheet(viewModel: viewModel)
            }
            .sheet(item: $noteToEdit) { note in
                EditNoteSheet(note: note, viewModel: viewModel)
            }
            .confirmationDialog(
                L10n.Notes.deleteConfirmTitle,
                isPresented: $showDeleteConfirm,
                titleVisibility: .visible
            ) {
                Button(L10n.Notes.deleteNote, role: .destructive) {
                    if let note = noteToDelete {
                        viewModel.deleteNote(id: note.id)
                    }
                }
                Button(L10n.General.cancel, role: .cancel) {}
            } message: {
                Text(L10n.Notes.deleteConfirmMessage)
            }
            .confirmationDialog(
                L10n.Notes.convertConfirmTitle,
                isPresented: Binding(
                    get: { noteToConvert != nil },
                    set: { if !$0 { noteToConvert = nil } }
                ),
                titleVisibility: .visible
            ) {
                Button(L10n.Notes.convertToTask) {
                    if let note = noteToConvert {
                        Task {
                            _ = await viewModel.convertToTask(note: note)
                            noteToConvert = nil
                        }
                    }
                }
                Button(L10n.General.cancel, role: .cancel) {
                    noteToConvert = nil
                }
            } message: {
                Text(L10n.Notes.convertConfirmMessage)
            }
        }
    }
}

// MARK: - Filter Pill
struct FilterPill: View {
    let title: String
    let isSelected: Bool
    let count: Int
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            HStack(spacing: 6) {
                Text(title)
                    .font(.system(size: 13, weight: isSelected ? .bold : .medium))
                Text("\(count)")
                    .font(.system(size: 11, weight: .bold))
                    .padding(.horizontal, 6)
                    .padding(.vertical, 2)
                    .background(isSelected ? Color.white.opacity(0.25) : Color.secondary.opacity(0.15))
                    .clipShape(Capsule())
            }
            .padding(.horizontal, 12)
            .padding(.vertical, 6)
            .background(isSelected ? AppColors.primary : Color(.secondarySystemGroupedBackground))
            .foregroundColor(isSelected ? .white : .primary)
            .clipShape(Capsule())
        }
        .buttonStyle(.plain)
    }
}

// MARK: - Note Card Row
struct NoteCardRow: View {
    let note: Note
    let onTogglePin: () -> Void
    let onEdit: () -> Void
    let onConvert: () -> Void
    let onDelete: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack(alignment: .top) {
                VStack(alignment: .leading, spacing: 4) {
                    HStack(spacing: 6) {
                        if note.isPinned {
                            Image(systemName: "pin.fill")
                                .font(.system(size: 11))
                                .foregroundColor(AppColors.primary)
                        }
                        Text(note.title)
                            .font(.system(size: 16, weight: .semibold))
                            .foregroundColor(.primary)
                    }

                    // Type pill
                    Text(note.type == "actionable" ? "ACTIONABLE" : "KNOWLEDGE")
                        .font(.system(size: 9, weight: .bold))
                        .foregroundColor(note.type == "actionable" ? AppColors.indigo : AppColors.violet)
                        .padding(.horizontal, 6)
                        .padding(.vertical, 2)
                        .background((note.type == "actionable" ? AppColors.indigo : AppColors.violet).opacity(0.12))
                        .clipShape(Capsule())
                }

                Spacer()

                // Pin Button
                Button(action: onTogglePin) {
                    Image(systemName: note.isPinned ? "pin.slash.fill" : "pin")
                        .font(.system(size: 14))
                        .foregroundColor(note.isPinned ? AppColors.primary : .secondary.opacity(0.7))
                }
                .buttonStyle(.plain)
            }

            if !note.content.isEmpty {
                Text(note.content)
                    .font(.system(size: 13))
                    .foregroundColor(.secondary)
                    .lineLimit(4)
            }

            // Card Action Buttons
            HStack(spacing: 12) {
                // Convert to Task (Actionable notes)
                if note.type == "actionable" {
                    Button(action: onConvert) {
                        HStack(spacing: 4) {
                            Image(systemName: "arrow.right.circle.fill")
                            Text(L10n.Notes.convertToTask)
                        }
                        .font(.system(size: 11, weight: .semibold))
                        .foregroundColor(AppColors.primary)
                        .padding(.horizontal, 8)
                        .padding(.vertical, 4)
                        .background(AppColors.primary.opacity(0.1))
                        .clipShape(Capsule())
                    }
                    .buttonStyle(.plain)
                }

                Spacer()

                // Edit
                Button(action: onEdit) {
                    Image(systemName: "pencil")
                        .font(.system(size: 13))
                        .foregroundColor(.secondary)
                }
                .buttonStyle(.plain)

                // Delete
                Button(action: onDelete) {
                    Image(systemName: "trash")
                        .font(.system(size: 13))
                        .foregroundColor(.red.opacity(0.8))
                }
                .buttonStyle(.plain)
            }
            .padding(.top, 4)
        }
        .contextMenu {
            Button(action: onTogglePin) {
                Label(note.isPinned ? L10n.Notes.unpin : L10n.Notes.pin, systemImage: note.isPinned ? "pin.slash" : "pin")
            }
            Button(action: onEdit) {
                Label(L10n.Notes.editNote, systemImage: "pencil")
            }
            if note.type == "actionable" {
                Button(action: onConvert) {
                    Label(L10n.Notes.convertToTask, systemImage: "arrow.right.circle")
                }
            }
            Divider()
            Button(role: .destructive, action: onDelete) {
                Label(L10n.General.delete, systemImage: "trash")
            }
        }
    }
}

// MARK: - Create Note Sheet
struct CreateNoteSheet: View {
    @Environment(\.dismiss) private var dismiss
    @Bindable var viewModel: NoteViewModel

    @State private var title = ""
    @State private var content = ""
    @State private var type = "actionable"
    @State private var isPinned = false
    @State private var isSubmitting = false

    var body: some View {
        NavigationStack {
            Form {
                Section(header: Text(L10n.Notes.sectionInfo)) {
                    TextField(L10n.Notes.fieldTitle, text: $title)
                    Picker(L10n.Notes.fieldType, selection: $type) {
                        Text(L10n.Notes.typeActionable).tag("actionable")
                        Text(L10n.Notes.typeKnowledge).tag("knowledge")
                    }
                    Toggle(L10n.Notes.pinToTop, isOn: $isPinned)
                }

                Section(header: Text(L10n.Notes.fieldContent)) {
                    TextEditor(text: $content)
                        .frame(minHeight: 140)
                }
            }
            .navigationTitle(L10n.Notes.newNote)
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button(L10n.General.cancel) { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button(L10n.General.save) {
                        isSubmitting = true
                        Task {
                            _ = await viewModel.createNote(title: title, content: content, type: type, isPinned: isPinned)
                            isSubmitting = false
                            dismiss()
                        }
                    }
                    .disabled(title.trimmingCharacters(in: .whitespaces).isEmpty || isSubmitting)
                    .fontWeight(.semibold)
                }
            }
        }
    }
}

// MARK: - Edit Note Sheet
struct EditNoteSheet: View {
    @Environment(\.dismiss) private var dismiss
    let note: Note
    @Bindable var viewModel: NoteViewModel

    @State private var title: String
    @State private var content: String
    @State private var type: String
    @State private var isSubmitting = false

    init(note: Note, viewModel: NoteViewModel) {
        self.note = note
        self.viewModel = viewModel
        _title = State(initialValue: note.title)
        _content = State(initialValue: note.content)
        _type = State(initialValue: note.type)
    }

    var body: some View {
        NavigationStack {
            Form {
                Section(header: Text(L10n.Notes.sectionInfo)) {
                    TextField(L10n.Notes.fieldTitle, text: $title)
                    Picker(L10n.Notes.fieldType, selection: $type) {
                        Text(L10n.Notes.typeActionable).tag("actionable")
                        Text(L10n.Notes.typeKnowledge).tag("knowledge")
                    }
                }

                Section(header: Text(L10n.Notes.fieldContent)) {
                    TextEditor(text: $content)
                        .frame(minHeight: 140)
                }
            }
            .navigationTitle(L10n.Notes.editNote)
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button(L10n.General.cancel) { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button(L10n.General.save) {
                        isSubmitting = true
                        Task {
                            _ = await viewModel.updateNote(id: note.id, title: title, content: content, type: type)
                            isSubmitting = false
                            dismiss()
                        }
                    }
                    .disabled(title.trimmingCharacters(in: .whitespaces).isEmpty || isSubmitting)
                    .fontWeight(.semibold)
                }
            }
        }
    }
}
