import SwiftUI

public struct AlarmRingingOverlayView: View {
    @Bindable var alarmManager: ActiveAlarmManager
    @State private var pulseScale: CGFloat = 1.0
    @State private var bellAngle: Double = 0.0
    @State private var selectedSnoozeMinutes: Int = 10

    public init(alarmManager: ActiveAlarmManager) {
        self.alarmManager = alarmManager
    }

    public var body: some View {
        if let reminder = alarmManager.activeReminder {
            ZStack {
                // Dimmed glassmorphism backdrop
                Color.black.opacity(0.7)
                    .ignoresSafeArea()
                    .background(.ultraThinMaterial)

                VStack(spacing: 24) {
                    Spacer()

                    // Animated ringing bell with pulsing waves
                    ZStack {
                        // Pulsing outer ripple
                        Circle()
                            .stroke(Color.orange.opacity(0.3), lineWidth: 4)
                            .frame(width: 140, height: 140)
                            .scaleEffect(pulseScale)
                            .opacity(2.0 - Double(pulseScale))

                        // Middle ripple
                        Circle()
                            .fill(
                                RadialGradient(
                                    colors: [Color.orange.opacity(0.35), Color.clear],
                                    center: .center,
                                    startRadius: 20,
                                    endRadius: 60
                                )
                            )
                            .frame(width: 120, height: 120)

                        // Bell icon circle
                        Circle()
                            .fill(
                                LinearGradient(
                                    colors: [Color.orange, Color.red],
                                    startPoint: .topLeading,
                                    endPoint: .bottomTrailing
                                )
                            )
                            .frame(width: 84, height: 84)
                            .shadow(color: Color.orange.opacity(0.6), radius: 16, x: 0, y: 6)

                        Image(systemName: "bell.badge.fill")
                            .font(.system(size: 38, weight: .bold))
                            .foregroundColor(.white)
                            .rotationEffect(.degrees(bellAngle))
                    }
                    .onAppear {
                        withAnimation(.easeInOut(duration: 0.9).repeatForever(autoreverses: true)) {
                            pulseScale = 1.35
                        }
                        withAnimation(.easeInOut(duration: 0.15).repeatForever(autoreverses: true)) {
                            bellAngle = 14
                        }
                    }

                    // Alarm Tag & Current Time
                    VStack(spacing: 8) {
                        Text(L10n.Reminders.liveAlert.uppercased())
                            .font(.system(size: 13, weight: .black, design: .rounded))
                            .tracking(2)
                            .foregroundColor(.orange)
                            .padding(.horizontal, 12)
                            .padding(.vertical, 4)
                            .background(Color.orange.opacity(0.18))
                            .clipShape(Capsule())

                        Text(currentTimeString)
                            .font(.system(size: 48, weight: .heavy, design: .rounded))
                            .foregroundColor(.white)

                        HStack(spacing: 6) {
                            Image(systemName: "alarm.fill")
                                .font(.system(size: 12))
                            Text(L10n.Alarm.scheduled(time: DateTimeHelper.formatTime(reminder.remindAt)))
                                .font(.system(size: 13, weight: .semibold))
                        }
                        .foregroundColor(.white.opacity(0.85))
                        .padding(.horizontal, 10)
                        .padding(.vertical, 4)
                        .background(Color.white.opacity(0.12))
                        .clipShape(Capsule())
                    }

                    // Reminder Info Card
                    VStack(spacing: 10) {
                        Text(reminder.title)
                            .font(.system(size: 22, weight: .bold, design: .rounded))
                            .foregroundColor(.white)
                            .multilineTextAlignment(.center)
                            .padding(.horizontal, 20)

                        if !reminder.description.isEmpty {
                            Text(reminder.description)
                                .font(.system(size: 14, weight: .medium))
                                .foregroundColor(.white.opacity(0.8))
                                .multilineTextAlignment(.center)
                                .padding(.horizontal, 24)
                                .lineLimit(3)
                        }

                        if let taskTitle = reminder.taskTitle, !taskTitle.isEmpty {
                            HStack(spacing: 6) {
                                Image(systemName: "checklist")
                                    .font(.system(size: 12))
                                Text(taskTitle)
                                    .font(.system(size: 13, weight: .semibold))
                                    .lineLimit(1)
                            }
                            .foregroundColor(.blue.opacity(0.9))
                            .padding(.horizontal, 12)
                            .padding(.vertical, 6)
                            .background(Color.blue.opacity(0.2))
                            .clipShape(Capsule())
                        }
                    }
                    .padding(.vertical, 12)

                    if let error = alarmManager.errorMessage {
                        Text(error)
                            .font(.system(size: 12, weight: .medium))
                            .foregroundColor(.red)
                            .padding(.horizontal, 16)
                    }

                    Spacer()

                    // Action Controls: Snooze & Dismiss
                    VStack(spacing: 14) {
                        // Dismiss / Matikan Button (Primary Action)
                        Button(action: {
                            Task {
                                await alarmManager.dismiss()
                            }
                        }) {
                            HStack(spacing: 10) {
                                if alarmManager.isProcessingAction {
                                    ProgressView()
                                        .tint(.white)
                                } else {
                                    Image(systemName: "checkmark.circle.fill")
                                        .font(.system(size: 20, weight: .bold))
                                    Text(L10n.Alarm.dismiss)
                                        .font(.system(size: 17, weight: .bold, design: .rounded))
                                }
                            }
                            .frame(maxWidth: .infinity)
                            .padding(.vertical, 16)
                            .background(
                                LinearGradient(
                                    colors: [Color.green, Color(red: 16/255, green: 185/255, blue: 129/255)],
                                    startPoint: .leading,
                                    endPoint: .trailing
                                )
                            )
                            .foregroundColor(.white)
                            .clipShape(RoundedRectangle(cornerRadius: 18, style: .continuous))
                            .shadow(color: Color.green.opacity(0.4), radius: 10, x: 0, y: 4)
                        }
                        .disabled(alarmManager.isProcessingAction)

                        // Snooze Button with options menu
                        HStack(spacing: 8) {
                            Button(action: {
                                Task {
                                    await alarmManager.snooze(minutes: selectedSnoozeMinutes)
                                }
                            }) {
                                HStack(spacing: 8) {
                                    Image(systemName: "clock.arrow.circlepath")
                                        .font(.system(size: 16, weight: .bold))
                                    Text(L10n.Alarm.snooze(minutes: selectedSnoozeMinutes))
                                        .font(.system(size: 15, weight: .semibold))
                                }
                                .frame(maxWidth: .infinity)
                                .padding(.vertical, 13)
                                .background(Color.white.opacity(0.15))
                                .foregroundColor(.white)
                                .clipShape(RoundedRectangle(cornerRadius: 14, style: .continuous))
                            }
                            .disabled(alarmManager.isProcessingAction)

                            // Menu for custom snooze duration
                            Menu {
                                Button(L10n.Alarm.snooze(minutes: 5)) {
                                    selectedSnoozeMinutes = 5
                                    Task { await alarmManager.snooze(minutes: 5) }
                                }
                                Button(L10n.Alarm.snooze(minutes: 10)) {
                                    selectedSnoozeMinutes = 10
                                    Task { await alarmManager.snooze(minutes: 10) }
                                }
                                Button(L10n.Alarm.snooze(minutes: 15)) {
                                    selectedSnoozeMinutes = 15
                                    Task { await alarmManager.snooze(minutes: 15) }
                                }
                                Button(L10n.Alarm.snooze(minutes: 30)) {
                                    selectedSnoozeMinutes = 30
                                    Task { await alarmManager.snooze(minutes: 30) }
                                }
                            } label: {
                                Image(systemName: "chevron.up.circle.fill")
                                    .font(.system(size: 22))
                                    .foregroundColor(.white.opacity(0.7))
                                    .padding(.horizontal, 12)
                                    .padding(.vertical, 12)
                                    .background(Color.white.opacity(0.12))
                                    .clipShape(RoundedRectangle(cornerRadius: 14, style: .continuous))
                            }
                            .disabled(alarmManager.isProcessingAction)
                        }
                    }
                    .padding(.horizontal, 24)
                    .padding(.bottom, 36)
                }
            }
            .transition(.opacity.combined(with: .scale(scale: 0.95)))
            .zIndex(9999)
        }
    }

    private var currentTimeString: String {
        let formatter = DateFormatter()
        formatter.dateFormat = "HH:mm"
        formatter.timeZone = DateTimeHelper.userTimeZone
        formatter.locale = Locale(identifier: "en_US_POSIX")
        return formatter.string(from: Date())
    }
}
