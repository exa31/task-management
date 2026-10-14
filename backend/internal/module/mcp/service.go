package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang/database"
	"golang/internal/module/note"
	"golang/internal/module/project"
	"golang/internal/module/reminder"
	"golang/internal/module/task"

	"github.com/jmoiron/sqlx"
)

type Service interface {
	CreateToken(ctx context.Context, userID string, name string) (*GenerateTokenResponse, error)
	ListTokens(ctx context.Context, userID string) ([]APITokenResponse, error)
	DeleteToken(ctx context.Context, userID string, tokenID int64) error
	ValidateToken(ctx context.Context, rawToken string) (string, error)
	HandleJSONRPC(ctx context.Context, userID string, req JSONRPCRequest) JSONRPCResponse
	SetReminderService(reminderService reminder.Service)
	SetProjectService(projectService project.Service)
}

type service struct {
	tokenRepo       TokenRepository
	noteService     note.Service
	taskService     task.Service
	reminderService reminder.Service
	projectService  project.Service
	db              database.Transactor
}

func NewService(tokenRepo TokenRepository, noteService note.Service, taskService task.Service, db database.Transactor) Service {
	return &service{
		tokenRepo:   tokenRepo,
		noteService: noteService,
		taskService: taskService,
		db:          db,
	}
}

func (s *service) SetReminderService(reminderService reminder.Service) {
	s.reminderService = reminderService
}

func (s *service) SetProjectService(projectService project.Service) {
	s.projectService = projectService
}

func (s *service) CreateToken(ctx context.Context, userID string, name string) (*GenerateTokenResponse, error) {
	var res *GenerateTokenResponse
	err := s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		txCtx := database.ContextWithTx(ctx, tx)
		var err error
		res, err = s.tokenRepo.CreateToken(txCtx, userID, name)
		return err
	})
	return res, err
}

func (s *service) ListTokens(ctx context.Context, userID string) ([]APITokenResponse, error) {
	var res []APITokenResponse
	err := s.db.WithReadOnlyTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		txCtx := database.ContextWithTx(ctx, tx)
		var err error
		res, err = s.tokenRepo.ListTokens(txCtx, userID)
		return err
	})
	return res, err
}

func (s *service) DeleteToken(ctx context.Context, userID string, tokenID int64) error {
	return s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		txCtx := database.ContextWithTx(ctx, tx)
		return s.tokenRepo.DeleteToken(txCtx, userID, tokenID)
	})
}

func (s *service) ValidateToken(ctx context.Context, rawToken string) (string, error) {
	var userID string
	err := s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		txCtx := database.ContextWithTx(ctx, tx)
		var err error
		userID, err = s.tokenRepo.ValidateToken(txCtx, rawToken)
		return err
	})
	return userID, err
}

func (s *service) HandleJSONRPC(ctx context.Context, userID string, req JSONRPCRequest) JSONRPCResponse {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
	}

	switch req.Method {
	case "initialize":
		resp.Result = map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"serverInfo": map[string]interface{}{
				"name":    "task-management-mcp",
				"version": "1.0.0",
			},
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
		}
		return resp

	case "ping":
		resp.Result = map[string]interface{}{}
		return resp

	case "tools/list":
		resp.Result = map[string]interface{}{
			"tools": s.getToolDefinitions(),
		}
		return resp

	case "tools/call":
		toolName, _ := req.Params["name"].(string)
		arguments, _ := req.Params["arguments"].(map[string]interface{})
		if arguments == nil {
			arguments = make(map[string]interface{})
		}

		result, err := s.callTool(ctx, userID, toolName, arguments)
		if err != nil {
			resp.Result = MCPToolResult{
				IsError: true,
				Content: []MCPToolContent{
					{Type: "text", Text: fmt.Sprintf("Error executing tool %s: %v", toolName, err)},
				},
			}
			return resp
		}

		resp.Result = result
		return resp

	default:
		resp.Error = &JSONRPCError{
			Code:    -32601,
			Message: fmt.Sprintf("Method not found: %s", req.Method),
		}
		return resp
	}
}

func (s *service) getToolDefinitions() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":        "create_idea_note",
			"description": "Quickly record a new idea or thought into the personal scratchpad.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"title":   map[string]interface{}{"type": "string", "description": "Idea title"},
					"content": map[string]interface{}{"type": "string", "description": "Markdown content or details"},
					"tags":    map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Tags (e.g. bug, feature, explore)"},
					"color":   map[string]interface{}{"type": "string", "description": "Card color label (default, blue, purple, green, amber, rose)"},
				},
				"required": []string{"title"},
			},
		},
		{
			"name":        "list_idea_notes",
			"description": "List existing idea notes from the user's scratchpad.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"search": map[string]interface{}{"type": "string", "description": "Search keyword in title or content"},
					"tag":    map[string]interface{}{"type": "string", "description": "Filter by tag"},
					"limit":  map[string]interface{}{"type": "number", "description": "Limit results (max 50)"},
				},
			},
		},
		{
			"name":        "list_projects",
			"description": "List existing projects/modules with their IDs, names, colors, descriptions, and task counts.",
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			"name":        "convert_note_to_task",
			"description": "Promote an existing idea note into an actionable task with optional subtasks.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"note_id":    map[string]interface{}{"type": "number", "description": "ID of the note to convert"},
					"priority":   map[string]interface{}{"type": "string", "enum": []string{"low", "medium", "high", "urgent"}, "description": "Task priority"},
					"due_date":   map[string]interface{}{"type": "string", "description": "Optional deadline (ISO-8601 string, e.g. '2026-10-15T18:00:00Z' or '2026-10-15')"},
					"subtasks":   map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Initial checklist steps"},
					"project_id": map[string]interface{}{"type": "number", "description": "Optional Project ID to link the converted task to"},
				},
				"required": []string{"note_id"},
			},
		},
		{
			"name":        "create_task",
			"description": "Create a new task directly in the user's task pipeline.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"title":               map[string]interface{}{"type": "string", "description": "Task title"},
					"description":         map[string]interface{}{"type": "string", "description": "Detailed description or markdown"},
					"status":              map[string]interface{}{"type": "string", "enum": []string{"backlog", "todo", "in_progress", "review", "done"}, "description": "Initial status"},
					"priority":            map[string]interface{}{"type": "string", "enum": []string{"low", "medium", "high", "urgent"}, "description": "Priority level"},
					"due_date":            map[string]interface{}{"type": "string", "description": "Due date/deadline ISO-8601 string (e.g. '2026-10-15T18:00:00Z' or '2026-10-15')"},
					"is_focus":            map[string]interface{}{"type": "boolean", "description": "Pin to Today's Focus"},
					"subtasks":            map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Checklist subtasks"},
					"tags":                map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Category tags"},
					"project_id":          map[string]interface{}{"type": "number", "description": "Optional Project ID to link this task to"},
					"is_recurring":        map[string]interface{}{"type": "boolean", "description": "Set whether this task automatically recurs after completion"},
					"recurrence_pattern":  map[string]interface{}{"type": "string", "enum": []string{"daily", "weekdays", "weekly", "monthly"}, "description": "Frequency of recurrence"},
					"recurrence_interval": map[string]interface{}{"type": "number", "description": "Interval between recurrences (default: 1)"},
					"recurrence_end_date": map[string]interface{}{"type": "string", "description": "Optional end date for recurring cycle (ISO-8601)"},
				},
				"required": []string{"title"},
			},
		},
		{
			"name":        "list_tasks",
			"description": "Retrieve tasks filtered by status, priority, or focus.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"status":     map[string]interface{}{"type": "string", "enum": []string{"backlog", "todo", "in_progress", "review", "done"}},
					"priority":   map[string]interface{}{"type": "string", "enum": []string{"low", "medium", "high", "urgent"}},
					"is_focus":   map[string]interface{}{"type": "boolean", "description": "Filter by pinned Today's Focus tasks"},
					"search":     map[string]interface{}{"type": "string", "description": "Search keyword"},
					"project_id": map[string]interface{}{"type": "number", "description": "Filter by Project ID"},
					"limit":      map[string]interface{}{"type": "number", "description": "Limit count"},
				},
			},
		},
		{
			"name":        "update_task_status",
			"description": "Update the status of a task in the pipeline.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id": map[string]interface{}{"type": "number", "description": "Task ID"},
					"status":  map[string]interface{}{"type": "string", "enum": []string{"backlog", "todo", "in_progress", "review", "done"}},
				},
				"required": []string{"task_id", "status"},
			},
		},
		{
			"name":        "update_task",
			"description": "Update task details such as title, description, priority, due date, focus, tags, and recurrence.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id":             map[string]interface{}{"type": "number", "description": "Task ID to update"},
					"title":               map[string]interface{}{"type": "string", "description": "New title"},
					"description":         map[string]interface{}{"type": "string", "description": "New description"},
					"status":              map[string]interface{}{"type": "string", "enum": []string{"backlog", "todo", "in_progress", "review", "done"}},
					"priority":            map[string]interface{}{"type": "string", "enum": []string{"low", "medium", "high", "urgent"}},
					"is_focus":            map[string]interface{}{"type": "boolean", "description": "Pin to Today's Focus"},
					"due_date":            map[string]interface{}{"type": "string", "description": "Due date ISO-8601 string"},
					"tags":                map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
					"project_id":          map[string]interface{}{"type": "number", "description": "Project ID to link this task to (pass 0 to unlink)"},
					"is_recurring":        map[string]interface{}{"type": "boolean", "description": "Update recurring status"},
					"recurrence_pattern":  map[string]interface{}{"type": "string", "enum": []string{"daily", "weekdays", "weekly", "monthly"}, "description": "Frequency of recurrence"},
					"recurrence_interval": map[string]interface{}{"type": "number", "description": "Interval between recurrences"},
					"recurrence_end_date": map[string]interface{}{"type": "string", "description": "End date for recurring cycle (ISO-8601)"},
				},
				"required": []string{"task_id"},
			},
		},
		{
			"name":        "delete_task",
			"description": "Permanently delete a task by ID.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id": map[string]interface{}{"type": "number", "description": "Task ID to delete"},
				},
				"required": []string{"task_id"},
			},
		},
		{
			"name":        "log_task_progress",
			"description": "Record recent progress update on a task ('inget-inget udah sampai mana') and optionally mark subtasks done.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id":               map[string]interface{}{"type": "number", "description": "Task ID"},
					"progress_note":         map[string]interface{}{"type": "string", "description": "Summary of work done or current checkpoint"},
					"status":                map[string]interface{}{"type": "string", "enum": []string{"backlog", "todo", "in_progress", "review", "done"}},
					"completed_subtask_ids": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "number"}},
				},
				"required": []string{"task_id", "progress_note"},
			},
		},
		{
			"name":        "get_task_details",
			"description": "Fetch detailed information of a single task including subtask checklists and last progress note.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id": map[string]interface{}{"type": "number", "description": "Task ID"},
				},
				"required": []string{"task_id"},
			},
		},
		{
			"name":        "create_reminder",
			"description": "Create a reminder/notification alert for a deadline or event.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"title":       map[string]interface{}{"type": "string", "description": "Reminder title/headline"},
					"description": map[string]interface{}{"type": "string", "description": "Optional notes or details"},
					"remind_at":    map[string]interface{}{"type": "string", "description": "Scheduled trigger timestamp (ISO-8601/RFC3339, e.g. '2026-10-15T09:00:00Z')"},
					"task_id":      map[string]interface{}{"type": "number", "description": "Optional linked Task ID"},
				},
				"required": []string{"title", "remind_at"},
			},
		},
		{
			"name":        "list_reminders",
			"description": "List reminders filtered by status (pending, sent, dismissed, all).",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"status":  map[string]interface{}{"type": "string", "enum": []string{"pending", "sent", "dismissed", "all"}, "description": "Filter by status (default: all)"},
					"task_id": map[string]interface{}{"type": "number", "description": "Filter by associated task ID"},
					"limit":   map[string]interface{}{"type": "number", "description": "Max results to return (default 50)"},
					"offset":  map[string]interface{}{"type": "number", "description": "Pagination offset"},
				},
			},
		},
		{
			"name":        "snooze_reminder",
			"description": "Snooze/postpone an existing reminder by a specified number of minutes.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"reminder_id": map[string]interface{}{"type": "number", "description": "Reminder ID"},
					"minutes":     map[string]interface{}{"type": "number", "description": "Duration in minutes to postpone (default 15)"},
				},
				"required": []string{"reminder_id"},
			},
		},
		{
			"name":        "dismiss_reminder",
			"description": "Dismiss/acknowledge a reminder so it no longer triggers notifications.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"reminder_id": map[string]interface{}{"type": "number", "description": "Reminder ID"},
				},
				"required": []string{"reminder_id"},
			},
		},
		{
			"name":        "delete_reminder",
			"description": "Permanently delete a reminder by ID.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"reminder_id": map[string]interface{}{"type": "number", "description": "Reminder ID"},
				},
				"required": []string{"reminder_id"},
			},
		},
	}
}

func (s *service) callTool(ctx context.Context, userID string, name string, args map[string]interface{}) (*MCPToolResult, error) {
	switch name {
	case "create_idea_note":
		title, _ := args["title"].(string)
		content, _ := args["content"].(string)
		color, _ := args["color"].(string)
		tags := parseStringSlice(args["tags"])

		res, err := s.noteService.Create(ctx, userID, note.CreateNoteRequest{
			Title:   title,
			Content: content,
			Tags:    tags,
			Color:   color,
		})
		if err != nil {
			return nil, err
		}
		jsonStr, _ := json.MarshalIndent(res, "", "  ")
		return &MCPToolResult{
			Content: []MCPToolContent{{Type: "text", Text: string(jsonStr)}},
		}, nil

	case "list_idea_notes":
		search, _ := args["search"].(string)
		tag, _ := args["tag"].(string)
		limit := parseInt(args["limit"], 30)

		res, err := s.noteService.FindAll(ctx, userID, note.NoteFilter{
			Search: search,
			Tag:    tag,
			Limit:  limit,
		})
		if err != nil {
			return nil, err
		}
		jsonStr, _ := json.MarshalIndent(res, "", "  ")
		return &MCPToolResult{
			Content: []MCPToolContent{{Type: "text", Text: string(jsonStr)}},
		}, nil

	case "list_projects":
		if s.projectService == nil {
			return nil, errors.New("project service is not available")
		}
		projects, err := s.projectService.List(ctx, userID)
		if err != nil {
			return nil, err
		}
		jsonStr, _ := json.MarshalIndent(projects, "", "  ")
		return &MCPToolResult{
			Content: []MCPToolContent{{Type: "text", Text: string(jsonStr)}},
		}, nil

	case "convert_note_to_task":
		noteID := int64(parseInt(args["note_id"], 0))
		priority, _ := args["priority"].(string)
		subtasks := parseStringSlice(args["subtasks"])
		var projectIDPtr *int64
		if pid := parseInt(args["project_id"], 0); pid > 0 {
			p := int64(pid)
			projectIDPtr = &p
		}
		var dueTimePtr *time.Time
		if dd, ok := args["due_date"].(string); ok && strings.TrimSpace(dd) != "" {
			if t, err := parseDateTime(dd); err == nil {
				dueTimePtr = &t
			}
		}

		n, taskID, err := s.noteService.ConvertToTask(ctx, userID, noteID, note.ConvertToTaskRequest{
			Priority:  priority,
			DueDate:   dueTimePtr,
			Subtasks:  subtasks,
			ProjectID: projectIDPtr,
		})
		if err != nil {
			return nil, err
		}
		out := map[string]interface{}{
			"message": "Note successfully converted to task",
			"task_id": taskID,
			"note":    n,
		}
		jsonStr, _ := json.MarshalIndent(out, "", "  ")
		return &MCPToolResult{
			Content: []MCPToolContent{{Type: "text", Text: string(jsonStr)}},
		}, nil

	case "create_task":
		title, _ := args["title"].(string)
		desc, _ := args["description"].(string)
		status, _ := args["status"].(string)
		priority, _ := args["priority"].(string)
		isFocus, _ := args["is_focus"].(bool)
		subtasks := parseStringSlice(args["subtasks"])
		tags := parseStringSlice(args["tags"])
		var projectIDPtr *int64
		if pid := parseInt(args["project_id"], 0); pid > 0 {
			p := int64(pid)
			projectIDPtr = &p
		}
		var dueTimePtr *time.Time
		if dd, ok := args["due_date"].(string); ok && strings.TrimSpace(dd) != "" {
			if t, err := parseDateTime(dd); err == nil {
				dueTimePtr = &t
			}
		}
		isRecurring, _ := args["is_recurring"].(bool)
		recurrencePattern, _ := args["recurrence_pattern"].(string)
		recurrenceInterval := parseInt(args["recurrence_interval"], 1)
		var recurrenceEndDatePtr *time.Time
		if red, ok := args["recurrence_end_date"].(string); ok && strings.TrimSpace(red) != "" {
			if t, err := parseDateTime(red); err == nil {
				recurrenceEndDatePtr = &t
			}
		}

		res, err := s.taskService.Create(ctx, userID, task.CreateTaskRequest{
			Title:              title,
			Description:        desc,
			Status:             status,
			Priority:           priority,
			DueDate:            dueTimePtr,
			IsFocus:            isFocus,
			Subtasks:           subtasks,
			Tags:               tags,
			ProjectID:          projectIDPtr,
			IsRecurring:        isRecurring,
			RecurrencePattern:  recurrencePattern,
			RecurrenceInterval: recurrenceInterval,
			RecurrenceEndDate:  recurrenceEndDatePtr,
		})
		if err != nil {
			return nil, err
		}
		jsonStr, _ := json.MarshalIndent(res, "", "  ")
		return &MCPToolResult{
			Content: []MCPToolContent{{Type: "text", Text: string(jsonStr)}},
		}, nil

	case "list_tasks":
		status, _ := args["status"].(string)
		priority, _ := args["priority"].(string)
		search, _ := args["search"].(string)
		limit := parseInt(args["limit"], 50)
		var isFocus *bool
		if f, ok := args["is_focus"].(bool); ok {
			isFocus = &f
		}
		var projectIDPtr *int64
		if pid := parseInt(args["project_id"], 0); pid > 0 {
			p := int64(pid)
			projectIDPtr = &p
		}

		res, err := s.taskService.FindAll(ctx, userID, task.TaskFilter{
			Status:    status,
			Priority:  priority,
			IsFocus:   isFocus,
			Search:    search,
			ProjectID: projectIDPtr,
			Limit:     limit,
		})
		if err != nil {
			return nil, err
		}
		jsonStr, _ := json.MarshalIndent(res, "", "  ")
		return &MCPToolResult{
			Content: []MCPToolContent{{Type: "text", Text: string(jsonStr)}},
		}, nil

	case "update_task_status":
		taskID := int64(parseInt(args["task_id"], 0))
		status, _ := args["status"].(string)

		res, err := s.taskService.Update(ctx, userID, taskID, task.UpdateTaskRequest{
			Status: &status,
		})
		if err != nil {
			return nil, err
		}
		jsonStr, _ := json.MarshalIndent(res, "", "  ")
		return &MCPToolResult{
			Content: []MCPToolContent{{Type: "text", Text: string(jsonStr)}},
		}, nil

	case "update_task":
		taskID := int64(parseInt(args["task_id"], 0))
		var titlePtr, descPtr, statusPtr, priorityPtr *string
		if t, ok := args["title"].(string); ok && t != "" {
			titlePtr = &t
		}
		if d, ok := args["description"].(string); ok {
			descPtr = &d
		}
		if s, ok := args["status"].(string); ok && s != "" {
			statusPtr = &s
		}
		if p, ok := args["priority"].(string); ok && p != "" {
			priorityPtr = &p
		}
		var dueTimePtr *time.Time
		if dd, ok := args["due_date"].(string); ok && strings.TrimSpace(dd) != "" {
			if t, err := parseDateTime(dd); err == nil {
				dueTimePtr = &t
			}
		}
		var isFocusPtr *bool
		if f, ok := args["is_focus"].(bool); ok {
			isFocusPtr = &f
		}
		var tagsPtr *[]string
		if rawTags, ok := args["tags"]; ok && rawTags != nil {
			parsed := parseStringSlice(rawTags)
			tagsPtr = &parsed
		}

		var projectIDPtr *int64
		hasProjectID := false
		if rawPID, exists := args["project_id"]; exists {
			hasProjectID = true
			if rawPID != nil {
				pid := int64(parseInt(rawPID, 0))
				if pid > 0 {
					projectIDPtr = &pid
				}
			}
		}

		var isRecurringPtr *bool
		if ir, ok := args["is_recurring"].(bool); ok {
			isRecurringPtr = &ir
		}
		var recurrencePatternPtr *string
		if rp, ok := args["recurrence_pattern"].(string); ok && rp != "" {
			recurrencePatternPtr = &rp
		}
		var recurrenceIntervalPtr *int
		if rawRI, exists := args["recurrence_interval"]; exists && rawRI != nil {
			ri := parseInt(rawRI, 1)
			recurrenceIntervalPtr = &ri
		}
		var recurrenceEndDatePtr *time.Time
		if red, ok := args["recurrence_end_date"].(string); ok && strings.TrimSpace(red) != "" {
			if t, err := parseDateTime(red); err == nil {
				recurrenceEndDatePtr = &t
			}
		}

		res, err := s.taskService.Update(ctx, userID, taskID, task.UpdateTaskRequest{
			Title:              titlePtr,
			Description:        descPtr,
			Status:             statusPtr,
			Priority:           priorityPtr,
			DueDate:            dueTimePtr,
			IsFocus:            isFocusPtr,
			Tags:               tagsPtr,
			ProjectID:          projectIDPtr,
			HasProjectID:       hasProjectID,
			IsRecurring:        isRecurringPtr,
			RecurrencePattern:  recurrencePatternPtr,
			RecurrenceInterval: recurrenceIntervalPtr,
			RecurrenceEndDate:  recurrenceEndDatePtr,
		})
		if err != nil {
			return nil, err
		}
		jsonStr, _ := json.MarshalIndent(res, "", "  ")
		return &MCPToolResult{
			Content: []MCPToolContent{{Type: "text", Text: string(jsonStr)}},
		}, nil

	case "delete_task":
		taskID := int64(parseInt(args["task_id"], 0))
		if err := s.taskService.Delete(ctx, userID, taskID); err != nil {
			return nil, err
		}
		out := map[string]interface{}{
			"message": "Task deleted successfully",
			"task_id": taskID,
		}
		jsonStr, _ := json.MarshalIndent(out, "", "  ")
		return &MCPToolResult{
			Content: []MCPToolContent{{Type: "text", Text: string(jsonStr)}},
		}, nil

	case "log_task_progress":
		taskID := int64(parseInt(args["task_id"], 0))
		progressNote, _ := args["progress_note"].(string)
		var statusPtr *string
		if st, ok := args["status"].(string); ok && st != "" {
			statusPtr = &st
		}

		var subtaskIDs []int64
		if rawSlice, ok := args["completed_subtask_ids"].([]interface{}); ok {
			for _, item := range rawSlice {
				if num, ok := item.(float64); ok {
					subtaskIDs = append(subtaskIDs, int64(num))
				}
			}
		}

		res, err := s.taskService.LogProgress(ctx, userID, taskID, task.LogProgressRequest{
			ProgressNote:        progressNote,
			Status:              statusPtr,
			CompletedSubtaskIDs: subtaskIDs,
		})
		if err != nil {
			return nil, err
		}
		jsonStr, _ := json.MarshalIndent(res, "", "  ")
		return &MCPToolResult{
			Content: []MCPToolContent{{Type: "text", Text: string(jsonStr)}},
		}, nil

	case "get_task_details":
		taskID := int64(parseInt(args["task_id"], 0))
		res, err := s.taskService.GetByID(ctx, userID, taskID)
		if err != nil {
			return nil, err
		}
		jsonStr, _ := json.MarshalIndent(res, "", "  ")
		return &MCPToolResult{
			Content: []MCPToolContent{{Type: "text", Text: string(jsonStr)}},
		}, nil

	case "create_reminder":
		if s.reminderService == nil {
			return nil, errors.New("reminder service is not available")
		}
		title, _ := args["title"].(string)
		desc, _ := args["description"].(string)
		remindAtStr, _ := args["remind_at"].(string)
		var taskID *int64
		if tid := parseInt(args["task_id"], 0); tid > 0 {
			t64 := int64(tid)
			taskID = &t64
		}

		remindAt, err := parseDateTime(remindAtStr)
		if err != nil {
			return nil, fmt.Errorf("invalid remind_at format: %v (expected ISO-8601/RFC3339, e.g. 2026-10-15T09:00:00Z)", err)
		}

		res, err := s.reminderService.Create(ctx, userID, reminder.CreateReminderRequest{
			Title:       title,
			Description: desc,
			RemindAt:    remindAt,
			TaskID:      taskID,
		})
		if err != nil {
			return nil, err
		}
		jsonStr, _ := json.MarshalIndent(res, "", "  ")
		return &MCPToolResult{
			Content: []MCPToolContent{{Type: "text", Text: string(jsonStr)}},
		}, nil

	case "list_reminders":
		if s.reminderService == nil {
			return nil, errors.New("reminder service is not available")
		}
		status, _ := args["status"].(string)
		limit := parseInt(args["limit"], 50)
		offset := parseInt(args["offset"], 0)
		var taskID *int64
		if tid := parseInt(args["task_id"], 0); tid > 0 {
			t64 := int64(tid)
			taskID = &t64
		}

		res, total, err := s.reminderService.List(ctx, userID, reminder.ReminderFilter{
			Status: status,
			TaskID: taskID,
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			return nil, err
		}
		out := map[string]interface{}{
			"total":     total,
			"reminders": res,
		}
		jsonStr, _ := json.MarshalIndent(out, "", "  ")
		return &MCPToolResult{
			Content: []MCPToolContent{{Type: "text", Text: string(jsonStr)}},
		}, nil

	case "snooze_reminder":
		if s.reminderService == nil {
			return nil, errors.New("reminder service is not available")
		}
		reminderID := int64(parseInt(args["reminder_id"], 0))
		minutes := parseInt(args["minutes"], 15)

		res, err := s.reminderService.Snooze(ctx, userID, reminderID, reminder.SnoozeReminderRequest{
			Minutes: minutes,
		})
		if err != nil {
			return nil, err
		}
		jsonStr, _ := json.MarshalIndent(res, "", "  ")
		return &MCPToolResult{
			Content: []MCPToolContent{{Type: "text", Text: string(jsonStr)}},
		}, nil

	case "dismiss_reminder":
		if s.reminderService == nil {
			return nil, errors.New("reminder service is not available")
		}
		reminderID := int64(parseInt(args["reminder_id"], 0))

		res, err := s.reminderService.Dismiss(ctx, userID, reminderID)
		if err != nil {
			return nil, err
		}
		jsonStr, _ := json.MarshalIndent(res, "", "  ")
		return &MCPToolResult{
			Content: []MCPToolContent{{Type: "text", Text: string(jsonStr)}},
		}, nil

	case "delete_reminder":
		if s.reminderService == nil {
			return nil, errors.New("reminder service is not available")
		}
		reminderID := int64(parseInt(args["reminder_id"], 0))

		if err := s.reminderService.Delete(ctx, userID, reminderID); err != nil {
			return nil, err
		}
		out := map[string]interface{}{
			"message":     "Reminder deleted successfully",
			"reminder_id": reminderID,
		}
		jsonStr, _ := json.MarshalIndent(out, "", "  ")
		return &MCPToolResult{
			Content: []MCPToolContent{{Type: "text", Text: string(jsonStr)}},
		}, nil

	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func parseDateTime(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}
	s = strings.TrimSpace(s)
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse date %q", s)
}

func parseStringSlice(v interface{}) []string {
	if slice, ok := v.([]interface{}); ok {
		res := make([]string, 0, len(slice))
		for _, s := range slice {
			if str, ok := s.(string); ok && strings.TrimSpace(str) != "" {
				res = append(res, strings.TrimSpace(str))
			}
		}
		return res
	}
	return []string{}
}

func parseInt(v interface{}, defaultVal int) int {
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case int32:
		return int(val)
	case float64:
		return int(val)
	case float32:
		return int(val)
	case string:
		if n, err := strconv.Atoi(strings.TrimSpace(val)); err == nil {
			return n
		}
	}
	return defaultVal
}
