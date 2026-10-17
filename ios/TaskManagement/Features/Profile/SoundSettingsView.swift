import SwiftUI
import UniformTypeIdentifiers

public struct SoundSettingsView: View {
    @State private var soundManager = SoundManager.shared
    @State private var isShowingFilePicker = false
    @State private var importErrorMessage: String?

    public init() {}

    public var body: some View {
        List {
            Section(
                header: Text(L10n.Settings.alarmRingtoneDefaultSection),
                footer: Text(L10n.Settings.alarmRingtoneDefaultDesc)
            ) {
                ForEach(SoundManager.presetSounds) { sound in
                    soundRow(sound: sound)
                }
            }

            let customSounds = soundManager.getCustomSounds()
            if !customSounds.isEmpty {
                Section(header: Text(L10n.Settings.alarmRingtoneCustomSection)) {
                    ForEach(customSounds) { sound in
                        soundRow(sound: sound)
                    }
                }
            }

            Section {
                Button(action: {
                    isShowingFilePicker = true
                }) {
                    HStack(spacing: 10) {
                        Image(systemName: "plus.circle.fill")
                            .font(.system(size: 16))
                            .foregroundColor(AppColors.primary)
                        Text(L10n.Settings.alarmRingtoneImport)
                            .font(.system(size: 15, weight: .medium))
                            .foregroundColor(.primary)
                    }
                }
            } footer: {
                Text(L10n.Settings.alarmRingtoneImportDesc)
            }

            if let error = importErrorMessage {
                Section {
                    Text(error)
                        .font(.system(size: 13))
                        .foregroundColor(.red)
                }
            }
        }
        .navigationTitle(L10n.Settings.alarmRingtone)
        #if canImport(UIKit)
        .navigationBarTitleDisplayMode(.inline)
        #endif
        .onDisappear {
            soundManager.stopPreview()
        }
        .fileImporter(
            isPresented: $isShowingFilePicker,
            allowedContentTypes: [UTType.audio],
            allowsMultipleSelection: false
        ) { result in
            do {
                guard let selectedFile = try result.get().first else { return }
                if selectedFile.startAccessingSecurityScopedResource() {
                    defer { selectedFile.stopAccessingSecurityScopedResource() }
                    let imported = try soundManager.importCustomSound(from: selectedFile)
                    soundManager.selectedSoundId = imported.id
                }
            } catch {
                importErrorMessage = L10n.Settings.alarmRingtoneImportError(message: error.localizedDescription)
            }
        }
    }

    @ViewBuilder
    private func soundRow(sound: SoundOption) -> some View {
        let isSelected = soundManager.selectedSoundId == sound.id
        let isPlaying = soundManager.currentPlayingSoundId == sound.id

        HStack(spacing: 12) {
            // Play/Stop Preview Button
            Button(action: {
                if isPlaying {
                    soundManager.stopPreview()
                } else {
                    soundManager.playPreview(sound: sound)
                }
            }) {
                ZStack {
                    Circle()
                        .fill(isPlaying ? Color.orange : AppColors.primary.opacity(0.12))
                        .frame(width: 36, height: 36)

                    Image(systemName: isPlaying ? "stop.fill" : "play.fill")
                        .font(.system(size: 14))
                        .foregroundColor(isPlaying ? .white : AppColors.primary)
                }
            }
            .buttonStyle(.plain)

            // Sound Name
            VStack(alignment: .leading, spacing: 2) {
                Text(sound.localizedName)
                    .font(.system(size: 15, weight: isSelected ? .bold : .regular))
                    .foregroundColor(.primary)

                if sound.isCustom {
                    Text(L10n.Settings.alarmRingtoneCustomBadge)
                        .font(.system(size: 11))
                        .foregroundColor(.secondary)
                }
            }

            Spacer()

            // Selection Checkmark
            if isSelected {
                Image(systemName: "checkmark")
                    .font(.system(size: 14, weight: .bold))
                    .foregroundColor(AppColors.primary)
            }
        }
        .contentShape(Rectangle())
        .onTapGesture {
            soundManager.selectedSoundId = sound.id
            soundManager.playPreview(sound: sound)
        }
    }
}
