import Foundation
import AVFoundation
import Observation
#if canImport(UIKit)
import UIKit
#endif

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

@Observable
@MainActor
public final class SoundManager {
    public static let shared = SoundManager()

    public static let soundStorageKey = "selected_alarm_sound"

    public static let presetSounds: [SoundOption] = [
        SoundOption(id: "classic_alarm", name: "Alarm Tegas (Penting)", fileName: "classic_alarm.aiff"),
        SoundOption(id: "gentle_bell", name: "Lonceng Lembut (Tenang)", fileName: "gentle_bell.aiff"),
        SoundOption(id: "lofi_chime", name: "Lo-Fi Melody (Santai)", fileName: "lofi_chime.aiff"),
        SoundOption(id: "energetic_synth", name: "Energetic Beat (Fokus)", fileName: "energetic_synth.aiff"),
        SoundOption(id: "default", name: "Default (Bip Sistem)", fileName: "default")
    ]

    public var isAlarmPlaying: Bool = false
    public var currentPlayingSoundId: String? = nil

    @ObservationIgnored private var audioPlayer: AVAudioPlayer?
    @ObservationIgnored private var hapticTimer: Timer?
    @ObservationIgnored private let soundsDirectoryName = "Sounds"

    private init() {
        createSoundsDirectoryIfNeeded()
    }

    public var selectedSoundId: String {
        get {
            UserDefaults.standard.string(forKey: Self.soundStorageKey) ?? "classic_alarm"
        }
        set {
            UserDefaults.standard.set(newValue, forKey: Self.soundStorageKey)
        }
    }

    public var selectedSound: SoundOption {
        let currentId = selectedSoundId
        let all = getAllSounds()
        return all.first(where: { $0.id == currentId }) ?? Self.presetSounds[0]
    }

    public func getAllSounds() -> [SoundOption] {
        return Self.presetSounds + getCustomSounds()
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

    /// Resolves the URL of an audio resource whether from SPM module, App Bundle, or Library/Sounds
    private func resolveSoundURL(sound: SoundOption) -> URL? {
        if sound.isCustom, let dir = soundsDirectoryURL {
            let customURL = dir.appendingPathComponent(sound.fileName)
            if FileManager.default.fileExists(atPath: customURL.path) {
                return customURL
            }
        }

        let name = (sound.fileName as NSString).deletingPathExtension
        let ext = (sound.fileName as NSString).pathExtension

        #if SWIFT_PACKAGE
        if let moduleURL = Bundle.module.url(forResource: name, withExtension: ext) {
            return moduleURL
        }
        #endif

        if let mainURL = Bundle.main.url(forResource: name, withExtension: ext) {
            return mainURL
        }

        return nil
    }

    /// Preview/Play a sound option (plays once)
    public func playPreview(sound: SoundOption) {
        stopAlarm()
        currentPlayingSoundId = sound.id

        if sound.id == "default" {
            AudioServicesPlaySystemSound(1007) // Standard SMS/alert beep
            return
        }

        guard let targetURL = resolveSoundURL(sound: sound) else {
            AudioServicesPlaySystemSound(1005)
            return
        }

        do {
            #if canImport(UIKit)
            try AVAudioSession.sharedInstance().setCategory(.ambient, mode: .default)
            try AVAudioSession.sharedInstance().setActive(true)
            #endif

            audioPlayer = try AVAudioPlayer(contentsOf: targetURL)
            audioPlayer?.numberOfLoops = 0
            audioPlayer?.prepareToPlay()
            audioPlayer?.play()
        } catch {
            AppLogger.warn("Failed to play audio preview: \(error.localizedDescription)", category: .push)
        }
    }

    /// Starts playing the continuous alarm (infinite loop) with haptics
    public func startAlarm(soundId: String? = nil) {
        stopAlarm()

        let targetId = soundId ?? selectedSoundId
        let sound = getAllSounds().first(where: { $0.id == targetId }) ?? Self.presetSounds[0]
        currentPlayingSoundId = sound.id

        if sound.id == "default" {
            AudioServicesPlaySystemSound(1005)
            startHapticLoop()
            isAlarmPlaying = true
            return
        }

        guard let targetURL = resolveSoundURL(sound: sound) else {
            AudioServicesPlaySystemSound(1005)
            startHapticLoop()
            isAlarmPlaying = true
            return
        }

        do {
            #if canImport(UIKit)
            try AVAudioSession.sharedInstance().setCategory(.playback, mode: .default, options: [.duckOthers])
            try AVAudioSession.sharedInstance().setActive(true)
            #endif

            audioPlayer = try AVAudioPlayer(contentsOf: targetURL)
            audioPlayer?.numberOfLoops = -1 // Infinite loop
            audioPlayer?.volume = 1.0
            audioPlayer?.prepareToPlay()
            audioPlayer?.play()

            startHapticLoop()
            isAlarmPlaying = true
            AppLogger.info("Alarm playback started: \(sound.name)", category: .push)
        } catch {
            AppLogger.warn("Failed to start alarm sound: \(error.localizedDescription)", category: .push)
            AudioServicesPlaySystemSound(1005)
            startHapticLoop()
            isAlarmPlaying = true
        }
    }

    /// Stops any ringing alarm or preview sound, terminates haptics
    public func stopAlarm() {
        audioPlayer?.stop()
        audioPlayer = nil
        stopHapticLoop()
        isAlarmPlaying = false
        currentPlayingSoundId = nil

        #if canImport(UIKit)
        try? AVAudioSession.sharedInstance().setActive(false, options: .notifyOthersOnDeactivation)
        #endif
    }

    /// Stops audio preview playback
    public func stopPreview() {
        stopAlarm()
    }

    private func startHapticLoop() {
        stopHapticLoop()
        #if canImport(UIKit)
        triggerHaptic()
        hapticTimer = Timer.scheduledTimer(withTimeInterval: 1.2, repeats: true) { [weak self] _ in
            Task { @MainActor [weak self] in
                guard self?.isAlarmPlaying == true else {
                    self?.stopHapticLoop()
                    return
                }
                self?.triggerHaptic()
            }
        }
        #endif
    }

    private func stopHapticLoop() {
        hapticTimer?.invalidate()
        hapticTimer = nil
    }

    private func triggerHaptic() {
        #if canImport(UIKit)
        let feedback = UINotificationFeedbackGenerator()
        feedback.notificationOccurred(.warning)
        #endif
    }
}
