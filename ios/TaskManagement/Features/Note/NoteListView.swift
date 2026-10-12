import SwiftUI

@Observable
public final class NoteViewModel {
    public var notes: [Note] = []
    public var isLoading = false
    public var errorMessage: String?

    private let apiClient = APIClient.shared

    public init() {
        Task { await loadNotes() }
    }

    @MainActor
    public func loadNotes() async {
        isLoading = true
        do {
            self.notes = try await apiClient.getNotes()
        } catch {
            self.errorMessage = error.localizedDescription
        }
        self.isLoading = false
    }

    @MainActor
    public func createNote(title: String, content: String, type: String = "actionable") {
        Task {
            do {
                _ = try await apiClient.createNote(
                    req: CreateNoteRequest(title: title, content: content, type: type, color: "#6366F1", isPinned: false)
                )
                await loadNotes()
            } catch {
                self.errorMessage = error.localizedDescription
            }
        }
    }
}

public struct NoteListView: View {
    @Bindable var viewModel: NoteViewModel
    @State private var showCreateSheet = false
    @State private var newTitle = ""
    @State private var newContent = ""
    @State private var newType = "actionable"

    public init(viewModel: NoteViewModel) {
        self.viewModel = viewModel
    }

    public var body: some View {
        NavigationStack {
            Group {
                if viewModel.isLoading && viewModel.notes.isEmpty {
                    VStack(spacing: 12) {
                        ProgressView()
                        Text("Memuat catatan…").font(.system(size: 13)).foregroundColor(.secondary)
                    }
                } else if viewModel.notes.isEmpty {
                    VStack(spacing: 14) {
                        ZStack {
                            Circle()
                                .fill(AppColors.primary.opacity(0.12))
                                .frame(width: 72, height: 72)
                            Image(systemName: "note.text")
                                .font(.system(size: 32))
                                .foregroundColor(AppColors.primary)
                        }
                        Text("Belum Ada Catatan")
                            .font(.system(size: 17, weight: .bold))
                        Text("Tulis ide, referensi pengetahuan, atau catatan actionable.")
                            .font(.system(size: 13))
                            .foregroundColor(.secondary)
                            .multilineTextAlignment(.center)
                            .padding(.horizontal, 32)

                        Button(action: { showCreateSheet = true }) {
                            HStack(spacing: 6) {
                                Image(systemName: "plus")
                                Text("Tulis Catatan")
                            }
                            .font(.system(size: 14, weight: .semibold))
                            .padding(.horizontal, 16)
                            .padding(.vertical, 10)
                            .background(AppColors.brandGradient)
                            .foregroundColor(.white)
                            .clipShape(Capsule())
                        }
                    }
                } else {
                    List {
                        ForEach(viewModel.notes) { note in
                            VStack(alignment: .leading, spacing: 8) {
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
                                        }

                                        // Type pill (Actionable vs Knowledge)
                                        Text(note.type == "actionable" ? "ACTIONABLE" : "KNOWLEDGE")
                                            .font(.system(size: 9, weight: .bold))
                                            .foregroundColor(note.type == "actionable" ? AppColors.indigo : AppColors.violet)
                                            .padding(.horizontal, 6)
                                            .padding(.vertical, 2)
                                            .background((note.type == "actionable" ? AppColors.indigo : AppColors.violet).opacity(0.1))
                                            .clipShape(Capsule())
                                    }

                                    Spacer()
                                }

                                if !note.content.isEmpty {
                                    Text(note.content)
                                        .font(.system(size: 13))
                                        .foregroundColor(.secondary)
                                        .lineLimit(4)
                                        .padding(.top, 2)
                                }
                            }
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
            .navigationTitle("Catatan & Ide")
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
                NavigationStack {
                    Form {
                        Section(header: Text("Informasi Catatan")) {
                            TextField("Judul Catatan *", text: $newTitle)
                            Picker("Tipe", selection: $newType) {
                                Text("Actionable (Bisa jadi tugas)").tag("actionable")
                                Text("Knowledge (Wawasan)").tag("knowledge")
                            }
                        }

                        Section(header: Text("Isi Catatan Markdown")) {
                            TextEditor(text: $newContent)
                                .frame(minHeight: 140)
                        }
                    }
                    .navigationTitle("Catatan Baru")
                    .navigationBarTitleDisplayMode(.inline)
                    .toolbar {
                        ToolbarItem(placement: .cancellationAction) {
                            Button("Batal") { showCreateSheet = false }
                        }
                        ToolbarItem(placement: .confirmationAction) {
                            Button("Simpan") {
                                viewModel.createNote(title: newTitle, content: newContent, type: newType)
                                newTitle = ""
                                newContent = ""
                                showCreateSheet = false
                            }
                            .disabled(newTitle.trimmingCharacters(in: .whitespaces).isEmpty)
                            .fontWeight(.semibold)
                        }
                    }
                }
            }
        }
    }
}
