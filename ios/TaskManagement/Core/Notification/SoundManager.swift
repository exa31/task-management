import Foundation
import AVFoundation

public struct SoundOption: Identifiable, Hashable, Sendable {
    public let id: String
    public let name: String
    public let fileName: String
    public let isCustom: Bool

    public init(id: String, name: String, fileName: String, isCustom: Bool = false) {
        self.id = id
        self.name = name
        self.fileName = fileName
        self.isCustom = isCustom
    }
}

@MainActor
public final class SoundManager: ObservableObject {
    public static let shared = SoundManager()

    public static let presetSounds: [SoundOption] = [
        SoundOption(id: "default", name: "Default (Sistem)", fileName: "default"),
        SoundOption(id: "gentle_bell", name: "Lonceng Lembut (Tenang)", fileName: "gentle_bell.aiff"),
        SoundOption(id: "lofi_chime", name: "Lo-Fi Melody (Santai)", fileName: "lofi_chime.aiff"),
        SoundOption(id: "energetic_synth", name: "Energetic Beat (Fokus)", fileName: "energetic_synth.aiff"),
        SoundOption(id: "classic_alarm", name: "Alarm Tegas (Penting)", fileName: "classic_alarm.aiff")
    ]

    private var audioPlayer: AVAudioPlayer?
    private let soundsDirectoryName = "Sounds"

    private init() {
        createSoundsDirectoryIfNeeded()
    }

    private var soundsDirectoryURL: URL? {
        guard let libraryURL = FileManager.default.urls(for: .libraryDirectory, in: .userDomainMask).first else {
            return nil
        }
        return libraryURL.appendingPathComponent(soundsDirectoryName, isDirectory: true)
    }

    private func createSoundsDirectoryIfNeeded() {
        guard let url = soundsDirectoryURL else { return }
        if !FileManager.default.fileExists(atPath: url.path) {
            try? FileManager.default.createDirectory(at: url, withIntermediateDirectories: true)
        }
    }

    /// List all custom audio files imported by user in Library/Sounds/
    public func getCustomSounds() -> [SoundOption] {
        guard let url = soundsDirectoryURL,
              let files = try? FileManager.default.contentsOfDirectory(atPath: url.path) else {
            return []
        }
        return files.compactMap { file in
            guard file.hasSuffix(".caf") || file.hasSuffix(".aiff") || file.hasSuffix(".wav") || file.hasSuffix(".m4a") else {
                return nil
            }
            let name = (file as NSString).deletingPathExtension
            return SoundOption(id: "custom_\(file)", name: "Musik: \(name)", fileName: file, isCustom: true)
        }
    }

    /// Import user chosen audio file into Library/Sounds for APNs custom sounds playback
    public func importCustomSound(from sourceURL: URL) throws -> SoundOption {
        createSoundsDirectoryIfNeeded()
        guard let targetDir = soundsDirectoryURL else {
            throw NSError(domain: "SoundManager", code: 1, userInfo: [NSLocalizedDescriptionKey: "Folder suara tidak ditemukan"])
        }

        let originalName = sourceURL.lastPathComponent
        let sanitizedName = originalName.replacingOccurrences(of: " ", with: "_")
        let destinationURL = targetDir.appendingPathComponent(sanitizedName)

        if FileManager.default.fileExists(atPath: destinationURL.path) {
            try? FileManager.default.removeItem(at: destinationURL)
        }

        try FileManager.default.copyItem(at: sourceURL, to: destinationURL)
        AppLogger.success("Custom audio file imported: \(sanitizedName)", category: .push)

        let displayName = (sanitizedName as NSString).deletingPathExtension
        return SoundOption(id: "custom_\(sanitizedName)", name: "Musik: \(displayName)", fileName: sanitizedName, isCustom: true)
    }

    /// Preview/Play a sound option
    public func playPreview(sound: SoundOption) {
        audioPlayer?.stop()
        if sound.id == "default" {
            AudioServicesPlaySystemSound(1007) // Standard SMS/alert beep
            return
        }

        let fileURL: URL?
        if sound.isCustom, let dir = soundsDirectoryURL {
            fileURL = dir.appendingPathComponent(sound.fileName)
        } else {
            fileURL = Bundle.main.url(forResource: (sound.fileName as NSString).deletingPathExtension, withExtension: (sound.fileName as NSString).pathExtension)
        }

        guard let targetURL = fileURL, FileManager.default.fileExists(atPath: targetURL.path) else {
            AudioServicesPlaySystemSound(1005)
            return
        }

        do {
            audioPlayer = try AVAudioPlayer(contentsOf: targetURL)
            audioPlayer?.prepareToPlay()
            audioPlayer?.play()
        } catch {
            AppLogger.warn("Failed to play audio preview: \(error.localizedDescription)", category: .push)
        }
    }

    public func stopPreview() {
        audioPlayer?.stop()
    }
}
