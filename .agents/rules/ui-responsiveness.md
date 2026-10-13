---
description: Mandatory rules for responsive UI design, calendar multi-dot indicators, network debounce, and non-technical user communication
globs: ["ios/TaskManagement/**/*.swift", "frontend/src/**/*.tsx", "android/**/*.kt"]
---

# UI Responsiveness, Calendar Standards & User Communication Rules

## 1. Non-Technical User Communication Standard
- **Never display technical jargon, raw code snippets, function names, latency numbers, or inner architecture details** in chat responses to the user unless explicitly asked.
- Always communicate in clear, friendly, everyday language focused on feature benefits and usability.

## 2. Universal UI Responsiveness
- All views in Mobile (iOS / Android) and Web must be fully responsive:
  - Must look and feel great from small devices (iPhone SE / 375pt width) to large Pro Max, iPad, and desktop screens.
  - Avoid fixed widths that cause horizontal clipping or overflow.
  - Buttons and interactive items must maintain comfortable tap targets (minimum 44x44pt on touch devices).
  - Text must wrap cleanly and support font scaling without breaking layout cards.

## 3. Calendar Visual Indicators (Multi-Dot System)
- Calendar date cells must visually indicate scheduled items using horizontal color dots beneath the day number:
  - 🔵 **Blue Dot**: Tasks due on this date.
  - 🟠 **Orange Dot**: Scheduled reminders.
  - 🟢 **Green Dot**: Work log / activity history records.
- If a date has 2 or all 3 categories, display the dots neatly side-by-side (`🔵 🟠 🟢`) with proper micro-spacing so they fit harmoniously inside the date box across all screen sizes.
- Tapping any date opens the organized agenda breakdown below the calendar.

## 4. Rapid Navigation & Overcall Protection
- Any interactive date picker, calendar switcher, or rapid navigation must include:
  - **Debouncing (200ms–250ms)** before making network calls.
  - **In-flight request cancellation** (using `AbortController` on web and `Task.cancel()` on mobile) so older slow responses never overwrite newer user selections.
