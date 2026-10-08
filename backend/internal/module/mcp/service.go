package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"golang/database"
	"golang/internal/module/note"
	"golang/internal/module/task"

	"github.com/jmoiron/sqlx"
)

type Service interface {
	CreateToken(ctx context.Context, userID string, name string) (*GenerateTokenResponse, error)
	ListTokens(ctx context.Context, userID string) ([]APITokenResponse, error)
	DeleteToken(ctx context.Context, userID string, tokenID int64) error
	ValidateToken(ctx context.Context, rawToken string) (string, error)
	HandleJSONRPC(ctx context.Context, userID string, req JSONRPCRequest) JSONRPCResponse
}

type service struct {
	tokenRepo   TokenRepository
	noteService note.Service
	taskService task.Service
	db          database.Transactor
}

func NewService(tokenRepo TokenRepository, noteService note.Service, taskService task.Service, db database.Transactor) Service {
	return &service{
		tokenRepo:   tokenRepo,
		noteService: noteService,
		taskService: taskService,
		db:          db,
	}
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
			"name":        "convert_note_to_task",
			"description": "Promote an existing idea note into an actionable task with optional subtasks.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"note_id":  map[string]interface{}{"type": "number", "description": "ID of the note to convert"},
					"priority": map[string]interface{}{"type": "string", "enum": []string{"low", "medium", "high", "urgent"}, "description": "Task priority"},
					"subtasks": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Initial checklist steps"},
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
					"title":       map[string]interface{}{"type": "string", "description": "Task title"},
					"description": map[string]interface{}{"type": "string", "description": "Detailed description or markdown"},
					"status":      map[string]interface{}{"type": "string", "enum": []string{"backlog", "todo", "in_progress", "review", "done"}, "description": "Initial status"},
					"priority":    map[string]interface{}{"type": "string", "enum": []string{"low", "medium", "high", "urgent"}, "description": "Priority level"},
					"is_focus":    map[string]interface{}{"type": "boolean", "description": "Pin to Today's Focus"},
					"subtasks":    map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Checklist subtasks"},
					"tags":        map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Category tags"},
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
					"status":   map[string]interface{}{"type": "string", "enum": []string{"backlog", "todo", "in_progress", "review", "done"}},
					"priority": map[string]interface{}{"type": "string", "enum": []string{"low", "medium", "high", "urgent"}},
					"is_focus": map[string]interface{}{"type": "boolean", "description": "Filter by pinned Today's Focus tasks"},
					"search":   map[string]interface{}{"type": "string", "description": "Search keyword"},
					"limit":    map[string]interface{}{"type": "number", "description": "Limit count"},
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

	case "convert_note_to_task":
		noteID := int64(parseInt(args["note_id"], 0))
		priority, _ := args["priority"].(string)
		subtasks := parseStringSlice(args["subtasks"])

		n, taskID, err := s.noteService.ConvertToTask(ctx, userID, noteID, note.ConvertToTaskRequest{
			Priority: priority,
			Subtasks: subtasks,
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

		res, err := s.taskService.Create(ctx, userID, task.CreateTaskRequest{
			Title:       title,
			Description: desc,
			Status:      status,
			Priority:    priority,
			IsFocus:     isFocus,
			Subtasks:    subtasks,
			Tags:        tags,
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

		res, err := s.taskService.FindAll(ctx, userID, task.TaskFilter{
			Status:   status,
			Priority: priority,
			IsFocus:  isFocus,
			Search:   search,
			Limit:    limit,
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

	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
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
	if num, ok := v.(float64); ok {
		return int(num)
	}
	if str, ok := v.(string); ok {
		if n, err := strconv.Atoi(str); err == nil {
			return n
		}
	}
	return defaultVal
}
