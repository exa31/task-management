# Agent Operational Guidelines & Proactive Task Management

## 1. Autonomous Task Tracking with `task-management` MCP
Whenever working on any user request (feature, bug fix, UI enhancement, refactoring):
1. **Always search existing tasks first**: Run `list_tasks` with project ID and keywords to prevent duplicate tasks.
2. **Adopt or Auto-Create**:
   - If an active task matches the request, adopt it and set status to `in_progress`.
   - If no task exists and the request is substantive, auto-create a task with descriptive title, subtasks, tags, and project ID (`2` for Task-management, `9` for Cyber, `8` for F-Tracker, etc.).
3. **Auto-Complete upon Resolution**:
   - When implementation is completed, verified, and committed/pushed, automatically invoke `log_task_progress` to mark the task and its subtasks as `done`.
   - The user does NOT need to explicitly request task creation or updates.

## 2. Quality Gates
- **iOS**: Always verify with `xcodebuild -project ios/TaskManagement.xcodeproj -scheme TaskManagement -destination 'generic/platform=iOS Simulator' CODE_SIGNING_ALLOWED=NO build`.
- **Backend**: Always verify Go code compilation and tests.
- **Frontend**: Verify with `npm run build` or typecheck where appropriate.
