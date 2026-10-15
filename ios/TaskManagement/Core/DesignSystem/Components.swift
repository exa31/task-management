import SwiftUI

public struct AppButton: View {
    let title: String
    let icon: String?
    let isLoading: Bool
    let action: () -> Void

    public init(
        title: String,
        icon: String? = nil,
        isLoading: Bool = false,
        action: @escaping () -> Void
    ) {
        self.title = title
        self.icon = icon
        self.isLoading = isLoading
        self.action = action
    }

    public var body: some View {
        Button(action: action) {
            HStack(spacing: 8) {
                if isLoading {
                    ProgressView()
                        .progressViewStyle(CircularProgressViewStyle(tint: .white))
                } else {
                    if let icon = icon {
                        Image(systemName: icon)
                            .font(.system(size: 15, weight: .semibold))
                    }
                    Text(title)
                        .font(.system(size: 15, weight: .semibold))
                }
            }
            .frame(maxWidth: .infinity)
            .frame(height: 48)
            .background(AppColors.brandGradient)
            .foregroundColor(.white)
            .clipShape(RoundedRectangle(cornerRadius: 14, style: .continuous))
            .shadow(color: AppColors.primary.opacity(0.35), radius: 8, x: 0, y: 4)
        }
        .buttonStyle(.plain)
        .disabled(isLoading)
    }
}

public struct PriorityBadge: View {
    let priority: String

    public init(priority: String) {
        self.priority = priority
    }

    private var meta: (icon: String, color: Color) {
        switch priority.lowercased() {
        case "urgent":
            return ("exclamationmark.triangle.fill", AppColors.priorityUrgent)
        case "high":
            return ("chevron.up.2", AppColors.priorityHigh)
        case "medium":
            return ("minus", AppColors.priorityMedium)
        case "low":
            return ("chevron.down", AppColors.priorityLow)
        default:
            return ("minus", AppColors.priorityLow)
        }
    }

    public var body: some View {
        HStack(spacing: 3) {
            Image(systemName: meta.icon)
                .font(.system(size: 9, weight: .bold))
            Text(priority.uppercased())
                .font(.system(size: 10, weight: .bold))
        }
        .foregroundColor(meta.color)
        .padding(.horizontal, 7)
        .padding(.vertical, 3.5)
        .background(meta.color.opacity(0.12))
        .overlay(
            RoundedRectangle(cornerRadius: 6, style: .continuous)
                .stroke(meta.color.opacity(0.25), lineWidth: 1)
        )
        .clipShape(RoundedRectangle(cornerRadius: 6, style: .continuous))
    }
}

public struct StatusBadge: View {
    let status: String

    public init(status: String) {
        self.status = status
    }

    private var color: Color {
        switch status.lowercased() {
        case "done": return AppColors.statusDone
        case "in_progress": return AppColors.statusInProgress
        case "review": return AppColors.statusReview
        case "backlog": return AppColors.statusBacklog
        default: return AppColors.statusTodo
        }
    }

    public var body: some View {
        Text(status.replacingOccurrences(of: "_", with: " ").uppercased())
            .font(.system(size: 10, weight: .bold))
            .foregroundColor(color)
            .padding(.horizontal, 7)
            .padding(.vertical, 3.5)
            .background(color.opacity(0.12))
            .overlay(
                RoundedRectangle(cornerRadius: 6, style: .continuous)
                    .stroke(color.opacity(0.25), lineWidth: 1)
            )
            .clipShape(RoundedRectangle(cornerRadius: 6, style: .continuous))
    }
}

public struct ProjectPill: View {
    let name: String
    let hexColor: String?

    public init(name: String, hexColor: String? = nil) {
        self.name = name
        self.hexColor = hexColor
    }

    private var parsedColor: Color {
        if let hex = hexColor, let c = Color(hex: hex) {
            return c
        }
        return AppColors.primary
    }

    public var body: some View {
        HStack(spacing: 4) {
            Image(systemName: "folder.fill")
                .font(.system(size: 9))
            Text(name)
                .font(.system(size: 10, weight: .semibold))
        }
        .foregroundColor(parsedColor)
        .padding(.horizontal, 6)
        .padding(.vertical, 3)
        .background(parsedColor.opacity(0.12))
        .overlay(
            RoundedRectangle(cornerRadius: 6, style: .continuous)
                .stroke(parsedColor.opacity(0.3), lineWidth: 1)
        )
        .clipShape(RoundedRectangle(cornerRadius: 6, style: .continuous))
    }
}

public struct ProgressBar: View {
    let percent: Int

    public init(percent: Int) {
        self.percent = max(0, min(100, percent))
    }

    public var body: some View {
        GeometryReader { proxy in
            ZStack(alignment: .leading) {
                Capsule()
                    .fill(Color.secondary.opacity(0.15))
                    .frame(height: 6)

                Capsule()
                    .fill(
                        LinearGradient(
                            colors: [AppColors.indigo, AppColors.sky],
                            startPoint: .leading,
                            endPoint: .trailing
                        )
                    )
                    .frame(width: proxy.size.width * CGFloat(percent) / 100.0, height: 6)
                    .animation(.easeInOut(duration: 0.3), value: percent)
            }
        }
        .frame(height: 6)
    }
}

public struct TaskItemSkeleton: View {
    public init() {}

    public var body: some View {
        HStack(spacing: 12) {
            RoundedRectangle(cornerRadius: 6, style: .continuous)
                .fill(Color.secondary.opacity(0.15))
                .frame(width: 22, height: 22)

            VStack(alignment: .leading, spacing: 6) {
                RoundedRectangle(cornerRadius: 4, style: .continuous)
                    .fill(Color.secondary.opacity(0.2))
                    .frame(width: 180, height: 14)

                RoundedRectangle(cornerRadius: 4, style: .continuous)
                    .fill(Color.secondary.opacity(0.15))
                    .frame(width: 100, height: 10)
            }

            Spacer()

            RoundedRectangle(cornerRadius: 8, style: .continuous)
                .fill(Color.secondary.opacity(0.15))
                .frame(width: 50, height: 22)
        }
        .padding(.horizontal, 14)
        .padding(.vertical, 12)
        .background(
            RoundedRectangle(cornerRadius: 16, style: .continuous)
                .fill(Color(uiColor: .secondarySystemGroupedBackground))
        )
        .redacted(reason: .placeholder)
    }
}

public struct TaskListSkeleton: View {
    let count: Int
    public init(count: Int = 6) {
        self.count = count
    }

    public var body: some View {
        VStack(spacing: 10) {
            ForEach(0..<count, id: \.self) { _ in
                TaskItemSkeleton()
            }
        }
    }
}

