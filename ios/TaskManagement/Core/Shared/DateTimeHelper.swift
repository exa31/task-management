import Foundation

public enum DateTimeHelper {
    public static let timezoneStorageKey = "user_configured_timezone"
    public static let localeStorageKey = "user_configured_locale"

    /// User's configured TimeZone from Settings/Profile (default: Asia/Jakarta)
    public static var userTimeZone: TimeZone {
        if let saved = UserDefaults.standard.string(forKey: timezoneStorageKey),
           let tz = TimeZone(identifier: saved) {
            return tz
        }
        return TimeZone(identifier: "Asia/Jakarta") ?? TimeZone.current
    }

    public static func setUserTimeZone(_ identifier: String?) {
        guard let id = identifier, !id.isEmpty, TimeZone(identifier: id) != nil else { return }
        UserDefaults.standard.set(id, forKey: timezoneStorageKey)
    }

    /// User's active Locale based on application language setting or explicit configuration
    public static var userLocale: Locale {
        if let saved = UserDefaults.standard.string(forKey: localeStorageKey), !saved.isEmpty {
            return Locale(identifier: saved)
        }
        let lang = UserDefaults.standard.string(forKey: "app_language") ?? I18n.shared.currentLanguage.rawValue
        return Locale(identifier: lang == "id" ? "id_ID" : "en_US")
    }

    public static func setUserLocale(_ identifier: String?) {
        guard let id = identifier, !id.isEmpty else { return }
        UserDefaults.standard.set(id, forKey: localeStorageKey)
    }

    private static let isoFormatterFractional: ISO8601DateFormatter = {
        let fmt = ISO8601DateFormatter()
        fmt.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        return fmt
    }()

    private static let isoFormatterBasic: ISO8601DateFormatter = {
        let fmt = ISO8601DateFormatter()
        fmt.formatOptions = [.withInternetDateTime]
        return fmt
    }()

    public static func parseISO8601(_ string: String) -> Date? {
        let trimmed = string.trimmingCharacters(in: .whitespacesAndNewlines)
        if let d = isoFormatterFractional.date(from: trimmed) { return d }
        return isoFormatterBasic.date(from: trimmed)
    }

    /// Formats time in HH:mm (e.g. "14:00") strictly according to the user's configured timezone and 24h standard
    public static func formatTime(_ isoString: String) -> String {
        guard let date = parseISO8601(isoString) else {
            if isoString.count >= 16, let tRange = isoString.range(of: "T") {
                let start = tRange.upperBound
                let end = isoString.index(start, offsetBy: 5, limitedBy: isoString.endIndex) ?? isoString.endIndex
                return String(isoString[start..<end])
            }
            return isoString
        }
        let formatter = DateFormatter()
        formatter.dateFormat = "HH:mm"
        formatter.timeZone = userTimeZone
        formatter.locale = Locale(identifier: "en_US_POSIX")
        return formatter.string(from: date)
    }

    /// Formats relative alarm date: "Hari ini" / "Today", "Besok" / "Tomorrow", or localized "Sab, 17 Okt" / "Sat, 17 Oct"
    public static func formatRelativeAlarmDate(_ isoString: String) -> String {
        guard let date = parseISO8601(isoString) else { return "" }
        var cal = Calendar.current
        cal.timeZone = userTimeZone
        cal.locale = userLocale

        let nowInTz = Date()
        if cal.isDate(date, inSameDayAs: nowInTz) {
            return I18n.shared.t("calendar.today")
        } else if let tomorrow = cal.date(byAdding: .day, value: 1, to: nowInTz),
                  cal.isDate(date, inSameDayAs: tomorrow) {
            return I18n.shared.t("widget.deadlines.tomorrow")
        } else {
            let fmt = DateFormatter()
            fmt.dateFormat = "EEE, d MMM"
            fmt.timeZone = userTimeZone
            fmt.locale = userLocale
            return fmt.string(from: date)
        }
    }

    /// Formats full localized date & time: "Sab, 17 Okt 2026, 14:00" / "Sat, 17 Oct 2026, 14:00"
    public static func formatFullDateTime(_ isoString: String) -> String {
        guard let date = parseISO8601(isoString) else { return isoString }
        let fmt = DateFormatter()
        fmt.dateFormat = "EEE, d MMM yyyy, HH:mm"
        fmt.timeZone = userTimeZone
        fmt.locale = userLocale
        return fmt.string(from: date)
    }

    /// Returns "YYYY-MM-DD" in user's timezone for calendar date grouping
    public static func toLocalDateKey(_ isoString: String) -> String {
        guard let date = parseISO8601(isoString) else {
            return String(isoString.prefix(10))
        }
        let fmt = DateFormatter()
        fmt.dateFormat = "yyyy-MM-dd"
        fmt.timeZone = userTimeZone
        fmt.locale = Locale(identifier: "en_US_POSIX")
        return fmt.string(from: date)
    }

    /// Creates ISO-8601 UTC string to send to backend
    public static func toISOString(_ date: Date) -> String {
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime]
        formatter.timeZone = TimeZone(secondsFromGMT: 0)
        return formatter.string(from: date)
    }
}
