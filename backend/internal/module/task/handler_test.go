package task

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

func setupTaskTestApp(repo Repository, userID string) *fiber.App {
	transactor := &mockTransactor{}
	svc := NewService(repo, transactor)
	handler := NewHandler(svc)

	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler(),
	})

	app.Use(func(c *fiber.Ctx) error {
		if userID != "" {
			c.Locals("user_id", userID)
		}
		return c.Next()
	})

	group := app.Group("/api/v1/tasks")
	group.Post("/", handler.Create)
	group.Get("/", handler.List)
	group.Get("/:id", handler.GetByID)
	group.Put("/:id", handler.Update)
	group.Delete("/:id", handler.Delete)
	group.Put("/:id/progress", handler.LogProgress)
	group.Patch("/:id/focus", handler.ToggleFocus)
	group.Get("/:id/activities", handler.GetActivities)
	group.Post("/:id/activities", handler.CreateActivity)

	return app
}

func doJSONTaskRequest(t *testing.T, app *fiber.App, method, path string, body interface{}) *http.Response {
	t.Helper()
	var bodyReader io.Reader
	if body != nil {
		if str, ok := body.(string); ok {
			bodyReader = strings.NewReader(str)
		} else {
			raw, err := json.Marshal(body)
			if err != nil {
				t.Fatalf("failed to marshal request body: %v", err)
			}
			bodyReader = bytes.NewReader(raw)
		}
	}

	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatalf("request %s %s failed: %v", method, path, err)
	}
	return resp
}

func parseJSONTaskResponse(t *testing.T, resp *http.Response) map[string]interface{} {
	t.Helper()
	defer resp.Body.Close()
	var res map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		t.Fatalf("failed to decode response JSON: %v", err)
	}
	return res
}

func TestTaskAPI_Create_AllFields(t *testing.T) {
	const testUser = "user-task-1"
	projID := int64(10)
	projName := "Marketing App"
	projColor := "#6366f1"

	repo := &mockRepository{
		createFunc: func(ctx context.Context, userID string, req CreateTaskRequest) (*TaskResponse, error) {
			now := time.Now()
			return &TaskResponse{
				ID:           1,
				UserID:       userID,
				Title:        req.Title,
				Description:  req.Description,
				Status:       req.Status,
				Priority:     req.Priority,
				DueDate:      req.DueDate,
				IsFocus:      req.IsFocus,
				Tags:         req.Tags,
				ProjectID:    req.ProjectID,
				ProjectName:  &projName,
				ProjectColor: &projColor,
				CreatedAt:    now,
				UpdatedAt:    now,
			}, nil
		},
	}

	app := setupTaskTestApp(repo, testUser)

	payload := map[string]interface{}{
		"title":       "Design Checkout Wireframes",
		"description": "Create Figma mocks for desktop & mobile checkout",
		"status":      "in_progress",
		"priority":    "high",
		"is_focus":    true,
		"project_id":  projID,
		"tags":        []string{"ui", "ux", "checkout"},
	}

	resp := doJSONTaskRequest(t, app, http.MethodPost, "/api/v1/tasks", payload)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d", resp.StatusCode)
	}

	res := parseJSONTaskResponse(t, resp)
	if res["code"] != "CREATED" {
		t.Errorf("expected code CREATED, got %v", res["code"])
	}

	data, ok := res["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %v", res["data"])
	}

	// Verify all returned fields
	if data["title"] != "Design Checkout Wireframes" {
		t.Errorf("expected title match, got %v", data["title"])
	}
	if data["description"] != "Create Figma mocks for desktop & mobile checkout" {
		t.Errorf("expected description match, got %v", data["description"])
	}
	if data["status"] != "in_progress" {
		t.Errorf("expected status in_progress, got %v", data["status"])
	}
	if data["priority"] != "high" {
		t.Errorf("expected priority high, got %v", data["priority"])
	}
	if data["is_focus"] != true {
		t.Errorf("expected is_focus true, got %v", data["is_focus"])
	}
	if data["project_id"] != float64(projID) {
		t.Errorf("expected project_id %d, got %v", projID, data["project_id"])
	}
	if data["project_name"] != "Marketing App" {
		t.Errorf("expected project_name Marketing App, got %v", data["project_name"])
	}
}

func TestTaskAPI_Create_ValidationErrors(t *testing.T) {
	repo := &mockRepository{}
	app := setupTaskTestApp(repo, "user-task-2")

	tests := []struct {
		name       string
		payload    string
		wantStatus int
		wantCode   string
	}{
		{
			name:       "Missing Title",
			payload:    `{"status":"todo"}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   "VALIDATION_ERROR",
		},
		{
			name:       "Invalid Status",
			payload:    `{"title":"Some Task","status":"not_a_valid_status"}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   "VALIDATION_ERROR",
		},
		{
			name:       "Invalid Priority",
			payload:    `{"title":"Some Task","priority":"extreme_super_urgent"}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := doJSONTaskRequest(t, app, http.MethodPost, "/api/v1/tasks", tt.payload)
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, resp.StatusCode)
			}
			res := parseJSONTaskResponse(t, resp)
			if res["code"] != tt.wantCode {
				t.Errorf("expected code %q, got %q (res: %#v)", tt.wantCode, res["code"], res)
			}
		})
	}
}

func TestTaskAPI_Update_Fields(t *testing.T) {
	const testUser = "user-task-3"
	updatedTitle := "Updated Task Title"
	updatedStatus := "done"
	updatedPriority := "urgent"
	updatedProjectID := int64(42)

	repo := &mockRepository{
		findByIDFunc: func(ctx context.Context, userID string, id int64) (*TaskResponse, error) {
			now := time.Now()
			return &TaskResponse{
				ID:        id,
				UserID:    userID,
				Title:     "Old Title",
				Status:    "todo",
				Priority:  "low",
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
		updateFunc: func(ctx context.Context, userID string, id int64, req UpdateTaskRequest) (*TaskResponse, error) {
			now := time.Now()
			res := &TaskResponse{
				ID:        id,
				UserID:    userID,
				Title:     "Old Title",
				Status:    "todo",
				Priority:  "low",
				CreatedAt: now,
				UpdatedAt: now,
			}
			if req.Title != nil {
				res.Title = *req.Title
			}
			if req.Status != nil {
				res.Status = *req.Status
			}
			if req.Priority != nil {
				res.Priority = *req.Priority
			}
			if req.ProjectID != nil {
				res.ProjectID = req.ProjectID
			}
			return res, nil
		},
	}

	app := setupTaskTestApp(repo, testUser)

	payload := map[string]interface{}{
		"title":      updatedTitle,
		"status":     updatedStatus,
		"priority":   updatedPriority,
		"project_id": updatedProjectID,
	}

	resp := doJSONTaskRequest(t, app, http.MethodPut, "/api/v1/tasks/1", payload)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", resp.StatusCode)
	}

	res := parseJSONTaskResponse(t, resp)
	data := res["data"].(map[string]interface{})

	if data["title"] != updatedTitle {
		t.Errorf("expected updated title %q, got %q", updatedTitle, data["title"])
	}
	if data["status"] != updatedStatus {
		t.Errorf("expected updated status %q, got %q", updatedStatus, data["status"])
	}
	if data["priority"] != updatedPriority {
		t.Errorf("expected updated priority %q, got %q", updatedPriority, data["priority"])
	}
	if data["project_id"] != float64(updatedProjectID) {
		t.Errorf("expected updated project_id %d, got %v", updatedProjectID, data["project_id"])
	}
}

func TestTaskAPI_Update_ClearDueDate(t *testing.T) {
	const testUser = "user-task-clear-due"
	initialDue := time.Now().Add(24 * time.Hour)

	var receivedReq UpdateTaskRequest

	repo := &mockRepository{
		findByIDFunc: func(ctx context.Context, userID string, id int64) (*TaskResponse, error) {
			now := time.Now()
			return &TaskResponse{
				ID:        id,
				UserID:    userID,
				Title:     "Task With Due Date",
				Status:    "todo",
				Priority:  "medium",
				DueDate:   &initialDue,
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
		updateFunc: func(ctx context.Context, userID string, id int64, req UpdateTaskRequest) (*TaskResponse, error) {
			receivedReq = req
			now := time.Now()
			res := &TaskResponse{
				ID:        id,
				UserID:    userID,
				Title:     "Task With Due Date",
				Status:    "todo",
				Priority:  "medium",
				CreatedAt: now,
				UpdatedAt: now,
			}
			if req.HasDueDate {
				res.DueDate = req.DueDate
			} else {
				res.DueDate = &initialDue
			}
			return res, nil
		},
	}

	app := setupTaskTestApp(repo, testUser)

	payload := map[string]interface{}{
		"due_date": nil,
	}

	resp := doJSONTaskRequest(t, app, http.MethodPut, "/api/v1/tasks/1", payload)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", resp.StatusCode)
	}

	if !receivedReq.HasDueDate {
		t.Errorf("expected HasDueDate to be true when due_date is null in JSON")
	}
	if receivedReq.DueDate != nil {
		t.Errorf("expected DueDate to be nil when cleared, got %v", receivedReq.DueDate)
	}

	res := parseJSONTaskResponse(t, resp)
	data := res["data"].(map[string]interface{})
	if data["due_date"] != nil {
		t.Errorf("expected due_date in response to be nil, got %v", data["due_date"])
	}
}

func TestTaskAPI_GetByID_And_NotFound(t *testing.T) {
	const testUser = "user-task-4"

	repo := &mockRepository{
		findByIDFunc: func(ctx context.Context, userID string, id int64) (*TaskResponse, error) {
			if id == 1 {
				return &TaskResponse{
					ID:     1,
					UserID: userID,
					Title:  "Existing Task",
					Status: "todo",
				}, nil
			}
			return nil, nil // not found
		},
	}

	app := setupTaskTestApp(repo, testUser)

	// Existing task
	resp := doJSONTaskRequest(t, app, http.MethodGet, "/api/v1/tasks/1", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", resp.StatusCode)
	}
	res := parseJSONTaskResponse(t, resp)
	data := res["data"].(map[string]interface{})
	if data["title"] != "Existing Task" {
		t.Errorf("expected title Existing Task, got %v", data["title"])
	}

	// Missing task -> 404
	resp404 := doJSONTaskRequest(t, app, http.MethodGet, "/api/v1/tasks/999", nil)
	if resp404.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404 Not Found, got %d", resp404.StatusCode)
	}
	res404 := parseJSONTaskResponse(t, resp404)
	if res404["code"] != "TASK_NOT_FOUND" {
		t.Errorf("expected code TASK_NOT_FOUND, got %v", res404["code"])
	}
}

func TestTaskAPI_List_WithProjectFilter(t *testing.T) {
	const testUser = "user-task-5"
	targetProjectID := int64(77)

	var capturedFilter TaskFilter
	repo := &mockRepository{
		findAllFunc: func(ctx context.Context, userID string, filter TaskFilter) ([]TaskResponse, error) {
			capturedFilter = filter
			return []TaskResponse{
				{
					ID:        1,
					UserID:    userID,
					Title:     "Project Task 1",
					ProjectID: &targetProjectID,
				},
			}, nil
		},
	}

	app := setupTaskTestApp(repo, testUser)

	resp := doJSONTaskRequest(t, app, http.MethodGet, fmt.Sprintf("/api/v1/tasks?project_id=%d", targetProjectID), nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", resp.StatusCode)
	}

	if capturedFilter.ProjectID == nil || *capturedFilter.ProjectID != targetProjectID {
		t.Errorf("expected captured filter project_id %d, got %v", targetProjectID, capturedFilter.ProjectID)
	}

	res := parseJSONTaskResponse(t, resp)
	data := res["data"].([]interface{})
	if len(data) != 1 {
		t.Fatalf("expected 1 task in response, got %d", len(data))
	}
}

func TestTaskAPI_List_CursorPagination(t *testing.T) {
	const testUser = "user-task-cursor"
	var capturedFilter TaskFilter

	repo := &mockRepository{
		findAllFunc: func(ctx context.Context, userID string, filter TaskFilter) ([]TaskResponse, error) {
			capturedFilter = filter
			// Generate 16 items to simulate having more than limit 15
			items := make([]TaskResponse, 16)
			for i := 0; i < 16; i++ {
				items[i] = TaskResponse{
					ID:     int64(i + 1),
					UserID: userID,
					Title:  fmt.Sprintf("Task %d", i+1),
				}
			}
			return items, nil
		},
	}

	app := setupTaskTestApp(repo, testUser)

	// Page 1
	resp := doJSONTaskRequest(t, app, http.MethodGet, "/api/v1/tasks?paginate=cursor&limit=15", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", resp.StatusCode)
	}

	res := parseJSONTaskResponse(t, resp)
	data := res["data"].(map[string]interface{})
	items := data["items"].([]interface{})
	if len(items) != 15 {
		t.Errorf("expected 15 items on page 1, got %d", len(items))
	}
	if data["has_more"] != true {
		t.Errorf("expected has_more to be true, got %v", data["has_more"])
	}
	nextCursor, ok := data["next_cursor"].(float64)
	if !ok || int64(nextCursor) != 15 {
		t.Errorf("expected next_cursor to be 15, got %v", data["next_cursor"])
	}

	// Page 2 request with cursor=15
	resp2 := doJSONTaskRequest(t, app, http.MethodGet, "/api/v1/tasks?paginate=cursor&limit=15&cursor=15", nil)
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 OK on page 2, got %d", resp2.StatusCode)
	}
	if capturedFilter.Offset != 15 {
		t.Errorf("expected captured filter offset 15, got %d", capturedFilter.Offset)
	}
}

func TestTaskHandler_Activities(t *testing.T) {
	testUser := "user-uuid-123"
	taskID := int64(10)

	repo := &mockRepository{
		findByIDFunc: func(ctx context.Context, userID string, id int64) (*TaskResponse, error) {
			return &TaskResponse{
				ID:     id,
				UserID: userID,
				Title:  "Task with Activities",
			}, nil
		},
		createActivityFunc: func(ctx context.Context, tid int64, uid string, actType string, desc string, changes []byte) (*TaskActivityResponse, error) {
			return &TaskActivityResponse{
				ID:           1,
				TaskID:       tid,
				UserID:       uid,
				ActivityType: actType,
				Description:  desc,
				CreatedAt:    time.Now(),
			}, nil
		},
		findActivitiesFunc: func(ctx context.Context, uid string, tid int64) ([]TaskActivityResponse, error) {
			return []TaskActivityResponse{
				{
					ID:           1,
					TaskID:       tid,
					UserID:       uid,
					ActivityType: "progress_note",
					Description:  "Testing activity handler",
					CreatedAt:    time.Now(),
				},
			}, nil
		},
	}

	app := setupTaskTestApp(repo, testUser)

	// Test GET /api/v1/tasks/:id/activities
	getResp := doJSONTaskRequest(t, app, http.MethodGet, fmt.Sprintf("/api/v1/tasks/%d/activities", taskID), nil)
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("expected GET status 200, got %d", getResp.StatusCode)
	}
	getRes := parseJSONTaskResponse(t, getResp)
	actData, ok := getRes["data"].([]interface{})
	if !ok || len(actData) != 1 {
		t.Fatalf("expected 1 activity item, got %v", getRes["data"])
	}

	// Test POST /api/v1/tasks/:id/activities
	postBody := map[string]string{
		"description": "Progress catatan manual",
	}
	postResp := doJSONTaskRequest(t, app, http.MethodPost, fmt.Sprintf("/api/v1/tasks/%d/activities", taskID), postBody)
	if postResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected POST status 201, got %d", postResp.StatusCode)
	}
}
