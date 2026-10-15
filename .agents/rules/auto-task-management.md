---
description: Proactive task management lifecycle rule - automatically search, create, update, and complete tasks in Task Hub MCP without requiring explicit user prompt reminders
globs: ["**/*"]
---

# Proactive Task Management Lifecycle Rule

Whenever the user requests work on features, bug fixes, UI improvements, refactoring, or maintenance, the AI agent must **proactively and autonomously** manage the task lifecycle in the `task-management` MCP tool without waiting for the user to say "bikin task ya" or "jangan lupa tasknya".

---

## 1. Project Identification

Determine the target `project_id` from the current workspace, active document path, or discussion context:

| Project Name | Project ID | Context / Keywords |
| :--- | :--- | :--- |
| **Task-management** | `2` | Task Hub repo (`ios/`, `android/`, `backend/`, `frontend/`, `notes/`) |
| **Cyber** | `9` | Cyber Mobile Flutter, Express Backend, Apple Store Web |
| **F-Tracker** | `8` | Personal finance tracker, Nuxt 3, Elysia Bun, Flutter |
| **Horus** | `4` | Horus office tasks |
| **Daily** | `6` | Daily routine tasks |
| **Career** | `1` | Interview, career schedule |
| **Coffe** | `3` | E-commerce POS coffee shop |
| **Template** | `5` | CI/CD templates |
| **Mikrotik** | `7` | Network / Mikrotik tasks |

*If uncertain, call `list_projects` to look up available project IDs.*

---

## 2. Autonomous Task Lifecycle

### Step 1: Pre-Work Task Discovery (Check Existing Tasks First)
- **Do not create duplicates.** Before starting or writing code, search if a relevant task or subtask already exists:
  ```json
  list_tasks({ "project_id": <id>, "search": "<keywords>" })
  ```
- **If an existing task or subtask matches**:
  - Adopt that task!
  - If its status is `todo` or `backlog`, update it to `in_progress`.
  - Briefly inform the user (e.g. `📌 Menggunakan task aktif #X: [Judul Task]`).
- **If NO matching task exists** and the work represents a substantive task, bug fix, or feature:
  - Create a new task via `create_task`:
    - `title`: Clear, concise Indonesian or English title reflecting the goal.
    - `description`: Summary of context, scope, and technical plan.
    - `project_id`: Appropriate project ID from table above.
    - `status`: `"in_progress"`
    - `priority`: `"medium"` or `"high"` depending on severity/impact.
    - `tags`: Relevant tags (e.g. `["ios", "ui-fix"]`, `["backend", "bug"]`).
    - `subtasks`: Concrete, actionable checklist items.
  - Briefly inform the user (e.g. `📝 Task baru otomatis dibuat #X: [Judul Task]`).

### Step 2: In-Progress Updates (Checkpointing)
- For multi-step tasks, record checkpoints using `log_task_progress` when major milestones or subtasks are verified.

### Step 3: Post-Work Completion & Push (Auto-Closure)
- Once the implementation is tested, verified (builds succeed), and pushed/committed:
  - Automatically complete the task and check off subtasks using `log_task_progress`:
    ```json
    log_task_progress({
      "task_id": <id>,
      "status": "done",
      "progress_note": "Ringkasan pekerjaan yang telah diselesaikan dan diverifikasi.",
      "completed_subtask_ids": [...]
    })
    ```
  - In the final response to the user, include a concise completion summary showing the task ID, title, and `done` status.

---

## 3. Communication Guidelines
- **Autonomous & Non-Intrusive**: Do not ask "Apakah Anda ingin saya buatkan task?". Just perform the lookup/creation/update seamlessly as part of your standard development workflow.
- **Transparent Summary**: State clearly in the final reply which task ID was updated or completed.
