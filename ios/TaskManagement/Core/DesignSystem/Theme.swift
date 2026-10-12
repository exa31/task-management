import SwiftUI

public struct AppColors {
    // Brand Gradient (matches frontend BrandLogo)
    public static let indigo = Color(red: 79/255, green: 70/255, blue: 229/255) // #4F46E5
    public static let primary = Color(red: 99/255, green: 102/255, blue: 241/255) // #6366F1
    public static let violet = Color(red: 124/255, green: 58/255, blue: 237/255) // #7C3AED
    public static let sky = Color(red: 56/255, green: 189/255, blue: 248/255) // #38BDF8

    // Background surfaces (Dark / Light)
    public static let darkSurface = Color(red: 9/255, green: 13/255, blue: 22/255) // #090D16
    public static let darkCard = Color(red: 15/255, green: 23/255, blue: 42/255) // #0F172A
    public static let darkCardHover = Color(red: 30/255, green: 41/255, blue: 59/255) // #1E293B

    // Status Colors (Matching web frontend STATUS_CLASSES)
    public static let statusBacklog = Color(red: 100/255, green: 116/255, blue: 139/255) // Slate
    public static let statusTodo = Color(red: 59/255, green: 130/255, blue: 246/255) // Blue
    public static let statusInProgress = Color(red: 217/255, green: 119/255, blue: 6/255) // Amber
    public static let statusReview = Color(red: 147/255, green: 51/255, blue: 234/255) // Purple
    public static let statusDone = Color(red: 16/255, green: 185/255, blue: 129/255) // Emerald

    // Priority Colors (Matching web frontend PRIORITY_META)
    public static let priorityUrgent = Color(red: 244/255, green: 63/255, blue: 94/255) // Rose-500
    public static let priorityHigh = Color(red: 245/255, green: 158/255, blue: 11/255) // Amber-500
    public static let priorityMedium = Color(red: 59/255, green: 130/255, blue: 246/255) // Blue-500
    public static let priorityLow = Color(red: 100/255, green: 116/255, blue: 139/255) // Slate-500

    public static var brandGradient: LinearGradient {
        LinearGradient(
            colors: [indigo, primary, violet],
            startPoint: .topLeading,
            endPoint: .bottomTrailing
        )
    }
}

// Brand Logo View replicating frontend/src/components/brand/BrandLogo.tsx
public struct BrandLogoView: View {
    public let size: CGFloat

    public init(size: CGFloat = 36) {
        self.size = size
    }

    public var body: some View {
        ZStack {
            // Squircle gradient background
            RoundedRectangle(cornerRadius: size * 0.25, style: .continuous)
                .fill(AppColors.brandGradient)
                .overlay(
                    RoundedRectangle(cornerRadius: size * 0.25, style: .continuous)
                        .stroke(Color.white.opacity(0.3), lineWidth: 1)
                )
                .shadow(color: AppColors.indigo.opacity(0.4), radius: size * 0.15, x: 0, y: size * 0.08)

            // Lightning Bolt emblem
            Image(systemName: "bolt.fill")
                .resizable()
                .scaledToFit()
                .foregroundColor(.white)
                .frame(width: size * 0.45, height: size * 0.45)
                .offset(x: -size * 0.04, y: -size * 0.02)

            // Small glowing check accent
            Circle()
                .fill(AppColors.sky)
                .frame(width: size * 0.16, height: size * 0.16)
                .offset(x: size * 0.22, y: size * 0.2)
        }
        .frame(width: size, height: size)
    }
}

public struct GlassCardModifier: ViewModifier {
    @Environment(\.colorScheme) var colorScheme

    public func body(content: Content) -> some View {
        content
            .padding(14)
            .background(
                RoundedRectangle(cornerRadius: 16, style: .continuous)
                    .fill(colorScheme == .dark ? AppColors.darkCard.opacity(0.8) : Color.white)
            )
            .overlay(
                RoundedRectangle(cornerRadius: 16, style: .continuous)
                    .stroke(
                        colorScheme == .dark ? Color.white.opacity(0.08) : Color.black.opacity(0.06),
                        lineWidth: 1
                    )
            )
            .shadow(
                color: colorScheme == .dark ? Color.black.opacity(0.3) : Color.black.opacity(0.04),
                radius: 8,
                x: 0,
                y: 3
            )
    }
}

public extension View {
    func glassCardStyle() -> some View {
        self.modifier(GlassCardModifier())
    }
    
    // Backward compatibility
    func appCardStyle() -> some View {
        self.modifier(GlassCardModifier())
    }
}

public extension Color {
    init?(hex: String) {
        var hexSanitized = hex.trimmingCharacters(in: .whitespacesAndNewlines)
        hexSanitized = hexSanitized.replacingOccurrences(of: "#", with: "")

        var rgb: UInt64 = 0
        guard Scanner(string: hexSanitized).scanHexInt64(&rgb) else { return nil }

        let r, g, b, a: Double
        let length = hexSanitized.count

        if length == 6 {
            r = Double((rgb & 0xFF0000) >> 16) / 255.0
            g = Double((rgb & 0x00FF00) >> 8) / 255.0
            b = Double(rgb & 0x0000FF) / 255.0
            a = 1.0
        } else if length == 8 {
            r = Double((rgb & 0xFF000000) >> 24) / 255.0
            g = Double((rgb & 0x00FF0000) >> 16) / 255.0
            b = Double((rgb & 0x0000FF00) >> 8) / 255.0
            a = Double(rgb & 0x000000FF) / 255.0
        } else {
            return nil
        }

        self.init(.sRGB, red: r, green: g, blue: b, opacity: a)
    }
}
