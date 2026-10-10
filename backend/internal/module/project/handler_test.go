package project

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

// setupProjectTestApp constructs an in-memory Fiber app with the project routes and mock authentication.
func setupProjectTestApp(repo Repository, userID string) *fiber.App {
	svc := NewService(repo)
	handler := NewHandler(svc)

	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler(),
	})

	// Simulate AuthMiddleware setting user_id in Fiber context locals
	app.Use(func(c *fiber.Ctx) error {
		if userID != "" {
			c.Locals("user_id", userID)
		}
		return c.Next()
	})

	group := app.Group("/api/v1/projects")
	group.Post("/", handler.Create)
	group.Get("/", handler.List)
	group.Get("/:id", handler.GetByID)
	group.Put("/:id", handler.Update)
	group.Delete("/:id", handler.Delete)

	return app
}

func doJSONRequest(t *testing.T, app *fiber.App, method, path string, body interface{}) *http.Response {
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

func parseJSONResponse(t *testing.T, resp *http.Response) map[string]interface{} {
	t.Helper()
	defer resp.Body.Close()
	var res map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		t.Fatalf("failed to decode response JSON: %v", err)
	}
	return res
}

func TestProjectAPI_Create_AllFields(t *testing.T) {
	const testUser = "user-101"
	repo := newMockRepository()
	app := setupProjectTestApp(repo, testUser)

	payload := map[string]interface{}{
		"name":        "E-Commerce Mobile App",
		"color":       "#ec4899",
		"description": "Cross-platform mobile application with shopping cart",
	}

	resp := doJSONRequest(t, app, http.MethodPost, "/api/v1/projects", payload)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d", resp.StatusCode)
	}

	res := parseJSONResponse(t, resp)
	if res["code"] != "CREATED" {
		t.Errorf("expected code CREATED, got %v", res["code"])
	}
	if res["success"] != true {
		t.Errorf("expected success true, got %v", res["success"])
	}

	data, ok := res["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %v", res["data"])
	}

	// Verify all returned fields
	if data["name"] != "E-Commerce Mobile App" {
		t.Errorf("expected name %q, got %q", "E-Commerce Mobile App", data["name"])
	}
	if data["color"] != "#ec4899" {
		t.Errorf("expected color %q, got %q", "#ec4899", data["color"])
	}
	if data["description"] != "Cross-platform mobile application with shopping cart" {
		t.Errorf("expected description %q, got %q", "Cross-platform mobile application with shopping cart", data["description"])
	}
	if data["task_count"] != float64(0) {
		t.Errorf("expected task_count 0, got %v", data["task_count"])
	}
	if data["user_id"] != testUser {
		t.Errorf("expected user_id %q, got %q", testUser, data["user_id"])
	}
	if data["id"] == nil || data["id"] == float64(0) {
		t.Errorf("expected valid positive id, got %v", data["id"])
	}
	if data["created_at"] == nil || data["created_at"] == "" {
		t.Errorf("expected non-empty created_at, got %v", data["created_at"])
	}
	if data["updated_at"] == nil || data["updated_at"] == "" {
		t.Errorf("expected non-empty updated_at, got %v", data["updated_at"])
	}
}

func TestProjectAPI_Create_DefaultValues(t *testing.T) {
	const testUser = "user-102"
	repo := newMockRepository()
	app := setupProjectTestApp(repo, testUser)

	payload := map[string]interface{}{
		"name": "Minimal Project",
	}

	resp := doJSONRequest(t, app, http.MethodPost, "/api/v1/projects", payload)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d", resp.StatusCode)
	}

	res := parseJSONResponse(t, resp)
	data := res["data"].(map[string]interface{})

	if data["name"] != "Minimal Project" {
		t.Errorf("expected name %q, got %q", "Minimal Project", data["name"])
	}
	if data["color"] != "indigo" {
		t.Errorf("expected default color %q, got %q", "indigo", data["color"])
	}
	if data["description"] != "" {
		t.Errorf("expected empty description, got %q", data["description"])
	}
}

func TestProjectAPI_Create_ValidationErrors(t *testing.T) {
	const testUser = "user-103"
	repo := newMockRepository()
	app := setupProjectTestApp(repo, testUser)

	tests := []struct {
		name       string
		payload    string
		wantStatus int
		wantCode   string
	}{
		{
			name:       "Empty Name",
			payload:    `{"name":""}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   "VALIDATION_ERROR",
		},
		{
			name:       "Name Too Short (min=2)",
			payload:    `{"name":"x"}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   "VALIDATION_ERROR",
		},
		{
			name:       "Color Too Long (max=30)",
			payload:    `{"name":"Valid Name","color":"this_color_hex_or_name_is_way_too_long_for_the_field"}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   "VALIDATION_ERROR",
		},
		{
			name:       "Malformed JSON",
			payload:    `{"name": invalid json`,
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := doJSONRequest(t, app, http.MethodPost, "/api/v1/projects", tt.payload)
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, resp.StatusCode)
			}
			res := parseJSONResponse(t, resp)
			if res["code"] != tt.wantCode {
				t.Errorf("expected code %q, got %q (payload: %#v)", tt.wantCode, res["code"], res)
			}
		})
	}
}

func TestProjectAPI_Update_AllFields(t *testing.T) {
	const testUser = "user-104"
	repo := newMockRepository()
	app := setupProjectTestApp(repo, testUser)

	// 1. Create project
	createResp := doJSONRequest(t, app, http.MethodPost, "/api/v1/projects", map[string]interface{}{
		"name":        "Initial Project Name",
		"color":       "#6366f1",
		"description": "Initial Description",
	})
	createData := parseJSONResponse(t, createResp)["data"].(map[string]interface{})
	projectID := int64(createData["id"].(float64))

	// 2. Update all fields via PUT
	updatePayload := map[string]interface{}{
		"name":        "Brand New Title",
		"color":       "#10b981",
		"description": "Completely revised project description",
	}
	updateResp := doJSONRequest(t, app, http.MethodPut, fmt.Sprintf("/api/v1/projects/%d", projectID), updatePayload)
	if updateResp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", updateResp.StatusCode)
	}

	updateRes := parseJSONResponse(t, updateResp)
	if updateRes["code"] != "SUCCESS" {
		t.Errorf("expected code SUCCESS, got %v", updateRes["code"])
	}

	updatedData := updateRes["data"].(map[string]interface{})
	if updatedData["name"] != "Brand New Title" {
		t.Errorf("expected updated name %q, got %q", "Brand New Title", updatedData["name"])
	}
	if updatedData["color"] != "#10b981" {
		t.Errorf("expected updated color %q, got %q", "#10b981", updatedData["color"])
	}
	if updatedData["description"] != "Completely revised project description" {
		t.Errorf("expected updated description %q, got %q", "Completely revised project description", updatedData["description"])
	}

	// 3. GET project to ensure persistence
	getResp := doJSONRequest(t, app, http.MethodGet, fmt.Sprintf("/api/v1/projects/%d", projectID), nil)
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 OK on GET, got %d", getResp.StatusCode)
	}
	getData := parseJSONResponse(t, getResp)["data"].(map[string]interface{})
	if getData["name"] != "Brand New Title" {
		t.Errorf("expected persistent name %q, got %q", "Brand New Title", getData["name"])
	}
}

func TestProjectAPI_Update_PartialFields(t *testing.T) {
	const testUser = "user-105"
	repo := newMockRepository()
	app := setupProjectTestApp(repo, testUser)

	// 1. Create project
	createResp := doJSONRequest(t, app, http.MethodPost, "/api/v1/projects", map[string]interface{}{
		"name":        "Stable Project",
		"color":       "#8b5cf6",
		"description": "Preserved Description",
	})
	createData := parseJSONResponse(t, createResp)["data"].(map[string]interface{})
	projectID := int64(createData["id"].(float64))

	// 2. Update only name
	updateResp := doJSONRequest(t, app, http.MethodPut, fmt.Sprintf("/api/v1/projects/%d", projectID), map[string]interface{}{
		"name": "Renamed Stable Project",
	})
	if updateResp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", updateResp.StatusCode)
	}

	updatedData := parseJSONResponse(t, updateResp)["data"].(map[string]interface{})
	if updatedData["name"] != "Renamed Stable Project" {
		t.Errorf("expected updated name, got %q", updatedData["name"])
	}
	// Verify color and description remained intact!
	if updatedData["color"] != "#8b5cf6" {
		t.Errorf("expected preserved color %q, got %q", "#8b5cf6", updatedData["color"])
	}
	if updatedData["description"] != "Preserved Description" {
		t.Errorf("expected preserved description %q, got %q", "Preserved Description", updatedData["description"])
	}
}

func TestProjectAPI_GetAndList(t *testing.T) {
	const testUser = "user-106"
	repo := newMockRepository()
	app := setupProjectTestApp(repo, testUser)

	// Create 2 projects
	_ = doJSONRequest(t, app, http.MethodPost, "/api/v1/projects", map[string]interface{}{"name": "Alpha Project"})
	_ = doJSONRequest(t, app, http.MethodPost, "/api/v1/projects", map[string]interface{}{"name": "Beta Project"})

	// List
	listResp := doJSONRequest(t, app, http.MethodGet, "/api/v1/projects", nil)
	if listResp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", listResp.StatusCode)
	}

	listRes := parseJSONResponse(t, listResp)
	items, ok := listRes["data"].([]interface{})
	if !ok || len(items) != 2 {
		t.Fatalf("expected 2 items, got %v", items)
	}

	// Verify NotFound
	notFoundResp := doJSONRequest(t, app, http.MethodGet, "/api/v1/projects/9999", nil)
	if notFoundResp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", notFoundResp.StatusCode)
	}
	notFoundRes := parseJSONResponse(t, notFoundResp)
	if notFoundRes["code"] != "PROJECT_NOT_FOUND" {
		t.Errorf("expected code PROJECT_NOT_FOUND, got %v", notFoundRes["code"])
	}

	// Verify Invalid ID parameter
	badIDResp := doJSONRequest(t, app, http.MethodGet, "/api/v1/projects/not-an-id", nil)
	if badIDResp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400 for non-numeric ID, got %d", badIDResp.StatusCode)
	}
}

func TestProjectAPI_Delete(t *testing.T) {
	const testUser = "user-107"
	repo := newMockRepository()
	app := setupProjectTestApp(repo, testUser)

	// 1. Create project
	createResp := doJSONRequest(t, app, http.MethodPost, "/api/v1/projects", map[string]interface{}{
		"name": "Project to Delete",
	})
	createData := parseJSONResponse(t, createResp)["data"].(map[string]interface{})
	projectID := int64(createData["id"].(float64))

	// 2. Delete project
	delResp := doJSONRequest(t, app, http.MethodDelete, fmt.Sprintf("/api/v1/projects/%d", projectID), nil)
	if delResp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", delResp.StatusCode)
	}
	delRes := parseJSONResponse(t, delResp)
	if delRes["success"] != true {
		t.Errorf("expected success true, got %v", delRes["success"])
	}

	// 3. Subsequent GET -> 404 PROJECT_NOT_FOUND
	getResp := doJSONRequest(t, app, http.MethodGet, fmt.Sprintf("/api/v1/projects/%d", projectID), nil)
	if getResp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404 Not Found, got %d", getResp.StatusCode)
	}
	getRes := parseJSONResponse(t, getResp)
	if getRes["code"] != "PROJECT_NOT_FOUND" {
		t.Errorf("expected code PROJECT_NOT_FOUND, got %v", getRes["code"])
	}

	// 4. Subsequent DELETE -> 404 PROJECT_NOT_FOUND
	delResp2 := doJSONRequest(t, app, http.MethodDelete, fmt.Sprintf("/api/v1/projects/%d", projectID), nil)
	if delResp2.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404 on deleting non-existent project, got %d", delResp2.StatusCode)
	}
}
