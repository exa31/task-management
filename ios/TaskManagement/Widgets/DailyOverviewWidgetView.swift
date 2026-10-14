import SwiftUI
#if canImport(WidgetKit)
import WidgetKit
#endif

public struct DailyOverviewWidgetView: View {
    @Environment(\.widgetFamily) var family
    public let entry: TaskWidgetEntry

    public init(entry: TaskWidgetEntry) {
        self.entry = entry
    }

    private var percent: Double {
        guard entry.bundle.totalToday > 0 else { return 0 }
        return Double(entry.bundle.completedToday) / Double(entry.bundle.totalToday)
    }

    public var body: some View {
        switch family {
        case .systemSmall:
            smallView
        default:
            mediumView
        }
    }

    // MARK: - Small Family

    private var smallView: some View {
        VStack(spacing: 8) {
            HStack {
                Text(L10n.Widgets.todaySummary)
                    .font(.system(size: 9, weight: .black))
                    .tracking(1)
                    .foregroundStyle(AppColors.primary)
                Spacer()
                BrandLogoView(size: 16)
            }

            Spacer()

            ZStack {
                Circle()
                    .stroke(Color.primary.opacity(0.1), lineWidth: 8)
                Circle()
                    .trim(from: 0, to: CGFloat(min(percent, 1.0)))
                    .stroke(
                        AppColors.brandGradient,
                        style: StrokeStyle(lineWidth: 8, lineCap: .round)
                    )
                    .rotationEffect(.degrees(-90))

                VStack(spacing: 1) {
                    Text("\(Int(percent * 100))%")
                        .font(.system(size: 18, weight: .black))
                        .foregroundStyle(.primary)
                    Text("\(entry.bundle.completedToday)/\(entry.bundle.totalToday)")
                        .font(.system(size: 10, weight: .semibold))
                        .foregroundStyle(.secondary)
                }
            }
            .frame(width: 75, height: 75)

            Spacer()

            Text(statusQuote)
                .font(.system(size: 10, weight: .medium))
                .foregroundStyle(.secondary)
                .lineLimit(1)
        }
        .padding(12)
        .widgetBackground()
    }

    // MARK: - Medium Family

    private var mediumView: some View {
        HStack(spacing: 16) {
            ZStack {
                Circle()
                    .stroke(Color.primary.opacity(0.1), lineWidth: 10)
                Circle()
                    .trim(from: 0, to: CGFloat(min(percent, 1.0)))
                    .stroke(
                        AppColors.brandGradient,
                        style: StrokeStyle(lineWidth: 10, lineCap: .round)
                    )
                    .rotationEffect(.degrees(-90))

                VStack(spacing: 2) {
                    Text("\(Int(percent * 100))%")
                        .font(.system(size: 22, weight: .black))
                        .foregroundStyle(.primary)
                    Text(L10n.Widgets.doneBadge)
                        .font(.system(size: 9, weight: .bold))
                        .tracking(1)
                        .foregroundStyle(AppColors.statusDone)
                }
            }
            .frame(width: 95, height: 95)

            VStack(alignment: .leading, spacing: 8) {
                HStack {
                    BrandLogoView(size: 20)
                    Text(L10n.Widgets.productivitySummary)
                        .font(.system(size: 13, weight: .bold))
                    Spacer()
                }

                Divider().opacity(0.3)

                VStack(alignment: .leading, spacing: 4) {
                    HStack {
                        Image(systemName: "checkmark.circle.fill")
                            .foregroundStyle(AppColors.statusDone)
                            .font(.system(size: 11))
                        Text(L10n.Widgets.tasksDone(completed: entry.bundle.completedToday, total: entry.bundle.totalToday))
                            .font(.system(size: 11, weight: .semibold))
                    }

                    HStack {
                        Image(systemName: "target")
                            .foregroundStyle(AppColors.sky)
                            .font(.system(size: 11))
                        Text(L10n.Widgets.tasksInFocus(count: entry.bundle.focusTasks.count))
                            .font(.system(size: 11, weight: .medium))
                            .foregroundStyle(.secondary)
                    }

                    HStack {
                        Image(systemName: "flame.fill")
                            .foregroundStyle(AppColors.priorityHigh)
                            .font(.system(size: 11))
                        Text(statusQuote)
                            .font(.system(size: 11, weight: .medium))
                            .foregroundStyle(AppColors.primary)
                    }
                }
            }
        }
        .padding(14)
        .widgetBackground()
    }

    private var statusQuote: String {
        if percent >= 1.0 {
            return L10n.Widgets.quoteAllDone
        } else if percent >= 0.7 {
            return L10n.Widgets.quoteAlmost
        } else if percent >= 0.3 {
            return L10n.Widgets.quoteSteady
        } else {
            return L10n.Widgets.quoteStart
        }
    }
}
