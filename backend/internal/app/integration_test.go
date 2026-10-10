//go:build integration

// Package app_test contains real end-to-end integration tests. They boot the
// real application container (no mocks) against a live PostgreSQL instance and
// exercise the HTTP API the same way a client would.
//
// They are isolated behind the "integration" build tag so the regular
// `go test ./...` stays fast and database-free. Run them explicitly with:
//
//	go test -tags=integration ./internal/app/...
//
// Requirements:
//   - A reachable PostgreSQL server (override via TEST_DB_HOST, TEST_DB_PORT,
//     TEST_DB_NAME, TEST_DB_USER, TEST_DB_PASSWORD).
//   - Migrations applied against the test database:
//     go run ./cmd -action up   (with DB_* vars pointing at the test DB)
//
// If the database is unreachable the tests are skipped instead of failing.
package app_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"golang/config"
	"golang/database"
	"golang/database/migrate"
	"golang/internal/app"

	"github.com/gofiber/fiber/v2"
)

// ---------------------------------------------------------------------------
// Setup helpers
// ---------------------------------------------------------------------------

// loadCfg returns the database configuration used by the integration tests,
// honouring TEST_DB_* overrides. It also pushes those values into the
// environment so that app.Boot (which reads env) connects to the same DB.
func loadCfg(t *testing.T) *config.Config {
	t.Helper()
	cfg := config.Load()
	cfg.Database.Host = getEnv("TEST_DB_HOST", "127.0.0.1")
	cfg.Database.Port = getEnv("TEST_DB_PORT", "55432")
	cfg.Database.DBName = getEnv("TEST_DB_NAME", "golang_test")
	cfg.Database.User = getEnv("TEST_DB_USER", "postgres")
	cfg.Database.Password = getEnv("TEST_DB_PASSWORD", "")
	cfg.Database.SSLMode = "disable"

	// Optional integrations stay off unless a server is explicitly available
	// (TEST_REDIS_ENABLED=true). Realtime only needs the local hub.
	cfg.Feature.Redis = getEnvAsBool("TEST_REDIS_ENABLED", false)
	cfg.Feature.RabbitMQ = false
	cfg.Feature.Realtime = true

	t.Setenv("DB_HOST", cfg.Database.Host)
	t.Setenv("DB_PORT", cfg.Database.Port)
	t.Setenv("DB_NAME", cfg.Database.DBName)
	t.Setenv("DB_USER", cfg.Database.User)
	t.Setenv("DB_PASSWORD", cfg.Database.Password)
	t.Setenv("DB_SSLMODE", "disable")
	t.Setenv("REDIS_ENABLED", boolStr(cfg.Feature.Redis))
	t.Setenv("RABBITMQ_ENABLED", "false")
	t.Setenv("REALTIME_ENABLED", "true")

	return cfg
}

// dbReachable reports whether the configured PostgreSQL instance is reachable.
func dbReachable(t *testing.T, cfg *config.Config) bool {
	t.Helper()
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		return false
	}
	db.Close()
	return true
}

// startTestApp applies migrations, boots the real container and returns the
// Fiber application plus a shutdown function. Queries are sent through the
// in-process fiber.Test client.
func startTestApp(t *testing.T, cfg *config.Config) (*app.Container, *fiber.App) {
	t.Helper()

	if err := migrate.RunMigrations(&cfg.Database, "../../migrations"); err != nil {
		t.Fatalf("migrations failed: %v", err)
	}

	container, shutdown := app.Boot()
	t.Cleanup(func() { _ = shutdown() })

	return container, app.NewHTTP(container)
}

// resetUsers empties the users table between tests so each test starts clean.
func resetUsers(t *testing.T, container *app.Container) {
	t.Helper()
	if _, err := container.DB.DB.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate users: %v", err)
	}
}

// ---------------------------------------------------------------------------
// HTTP helpers
// ---------------------------------------------------------------------------

const jsonCT = "application/json"

// call performs an HTTP request against the Fiber test app.
func call(t *testing.T, httpApp *fiber.App, method, path, body, token string) *http.Response {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", jsonCT)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := httpApp.Test(req, 10000)
	if err != nil {
		t.Fatalf("request %s %s failed: %v", method, path, err)
	}
	return resp
}

func get(t *testing.T, httpApp *fiber.App, path string) *http.Response {
	t.Helper()
	return call(t, httpApp, http.MethodGet, path, "", "")
}

func post(t *testing.T, httpApp *fiber.App, path, body string) *http.Response {
	t.Helper()
	return call(t, httpApp, http.MethodPost, path, body, "")
}

func callAuth(t *testing.T, httpApp *fiber.App, method, path, body, token string) *http.Response {
	t.Helper()
	if token == "" {
		t.Fatalf("callAuth requires a token for %s %s", method, path)
	}
	return call(t, httpApp, method, path, body, token)
}

func requireStatus(t *testing.T, resp *http.Response, want int) {
	t.Helper()
	if resp.StatusCode != want {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status %d, got %d. body: %s", want, resp.StatusCode, string(body))
	}
}

// decodeJSON reads and unmarshals a JSON body into a generic map.
func decodeJSON(t *testing.T, reader io.Reader) map[string]interface{} {
	t.Helper()
	raw, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("failed to decode JSON %q: %v", string(raw), err)
	}
	return out
}

// dataOf returns the "data" member of a response as a map.
func dataOf(t *testing.T, payload map[string]interface{}) map[string]interface{} {
	t.Helper()
	data, _ := payload["data"].(map[string]interface{})
	return data
}

func requireCode(t *testing.T, payload map[string]interface{}, want string) {
	t.Helper()
	if got, _ := payload["code"].(string); got != want {
		t.Fatalf("expected response code %q, got %q (payload: %#v)", want, got, payload)
	}
}

func requireErrorCode(t *testing.T, resp *http.Response, want string) {
	t.Helper()
	defer resp.Body.Close()
	payload := decodeJSON(t, resp.Body)
	if success, _ := payload["success"].(bool); success {
		t.Fatalf("expected non-success response for %q, got success payload: %#v", want, payload)
	}
	if got, _ := payload["code"].(string); got != want {
		t.Fatalf("expected error code %q, got %q (payload: %#v)", want, got, payload)
	}
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestHealthEndpoint(t *testing.T) {
	cfg := loadCfg(t)
	if !dbReachable(t, cfg) {
		t.Skip("PostgreSQL not reachable; skipping integration test")
	}
	_, httpApp := startTestApp(t, cfg)

	resp := get(t, httpApp, "/health")
	defer resp.Body.Close()
	requireStatus(t, resp, http.StatusOK)

	payload := decodeJSON(t, resp.Body)
	if got, _ := payload["status"].(string); got != "healthy" {
		t.Fatalf("expected healthy status, got %#v", payload)
	}
	features, ok := payload["features"].(map[string]interface{})
	if !ok {
		t.Fatalf("health payload missing features map: %#v", payload)
	}
	if got := features["realtime"]; got != true {
		t.Errorf("expected realtime=true, got %v", got)
	}
	if got := features["redis"]; got != false {
		t.Errorf("expected redis=false (disabled in tests), got %v", got)
	}
}

func TestAuthLifecycle(t *testing.T) {
	cfg := loadCfg(t)
	if !dbReachable(t, cfg) {
		t.Skip("PostgreSQL not reachable; skipping integration test")
	}
	container, httpApp := startTestApp(t, cfg)
	resetUsers(t, container)

	email := fmt.Sprintf("user-%s@example.com", randSuffix())
	password := "SuperSecret1"

	// 1. Register.
	regResp := post(t, httpApp, "/api/v1/auth/register",
		fmt.Sprintf(`{"name":"Test User","email":%q,"password":%q}`, email, password))
	requireStatus(t, regResp, http.StatusCreated)
	reg := decodeJSON(t, regResp.Body)
	requireCode(t, reg, "CREATED")
	regData := dataOf(t, reg)
	tokens, _ := regData["tokens"].(map[string]interface{})
	accessToken, _ := tokens["access_token"].(string)
	refreshToken, _ := tokens["refresh_token"].(string)
	if accessToken == "" || refreshToken == "" {
		t.Fatalf("register response missing tokens: %#v", regData)
	}
	regResp.Body.Close()

	// 2. Login with wrong password -> 401 INVALID_CREDENTIALS.
	badLogin := post(t, httpApp, "/api/v1/auth/login",
		fmt.Sprintf(`{"email":%q,"password":"wrong"}`, email))
	requireStatus(t, badLogin, http.StatusUnauthorized)
	requireErrorCode(t, badLogin, "INVALID_CREDENTIALS")

	// 3. Login.
	loginResp := post(t, httpApp, "/api/v1/auth/login",
		fmt.Sprintf(`{"email":%q,"password":%q}`, email, password))
	requireStatus(t, loginResp, http.StatusOK)
	requireCode(t, decodeJSON(t, loginResp.Body), "SUCCESS")
	loginResp.Body.Close()

	// 4. Protected profile with the access token.
	profileResp := callAuth(t, httpApp, http.MethodGet, "/api/v1/auth/profile", "", accessToken)
	requireStatus(t, profileResp, http.StatusOK)
	profile := dataOf(t, decodeJSON(t, profileResp.Body))
	if got, _ := profile["email"].(string); got != email {
		t.Errorf("expected profile email %q, got %#v", email, profile)
	}
	profileResp.Body.Close()

	// 5. Refresh the access token.
	refreshResp := post(t, httpApp, "/api/v1/auth/refresh",
		fmt.Sprintf(`{"refresh_token":%q}`, refreshToken))
	requireStatus(t, refreshResp, http.StatusOK)
	refreshData := dataOf(t, decodeJSON(t, refreshResp.Body))
	newAccess, _ := refreshData["access_token"].(string)
	if newAccess == "" {
		t.Fatalf("refresh did not return a new access token: %#v", refreshData)
	}
	refreshResp.Body.Close()

	// 6. Update profile.
	updResp := callAuth(t, httpApp, http.MethodPut, "/api/v1/auth/profile", `{"name":"Updated Name"}`, newAccess)
	requireStatus(t, updResp, http.StatusOK)
	upd := dataOf(t, decodeJSON(t, updResp.Body))
	if got, _ := upd["name"].(string); got != "Updated Name" {
		t.Errorf("expected updated name, got %#v", upd)
	}
	updResp.Body.Close()

	// 7. Logout returns success.
	logoutResp := callAuth(t, httpApp, http.MethodPost, "/api/v1/auth/logout", `{}`, accessToken)
	requireStatus(t, logoutResp, http.StatusOK)
	logoutResp.Body.Close()
}

func TestUserCRUD(t *testing.T) {
	cfg := loadCfg(t)
	if !dbReachable(t, cfg) {
		t.Skip("PostgreSQL not reachable; skipping integration test")
	}
	container, httpApp := startTestApp(t, cfg)
	resetUsers(t, container)

	jsonBody := `{"name": "Alice", "email": "alice@example.com", "phone": "0812345678"}`

	// Create.
	createResp := call(t, httpApp, "POST", "/api/v1/users", jsonBody, "")
	requireStatus(t, createResp, http.StatusCreated)
	created := decodeJSON(t, createResp.Body)
	requireCode(t, created, "CREATED")
	id, _ := dataOf(t, created)["id"].(string)
	if id == "" {
		t.Fatalf("expected created user id, %#v", created)
	}
	createResp.Body.Close()

	// Duplicate email -> 409 (returns DUPLICATE_ENTRY).
	dupResp := call(t, httpApp, "POST", "/api/v1/users", jsonBody, "")
	requireStatus(t, dupResp, http.StatusConflict)
	requireErrorCode(t, dupResp, "DUPLICATE_ENTRY")

	// Validation failure -> 400 VALIDATION_ERROR.
	badResp := call(t, httpApp, "POST", "/api/v1/users", `{"name":"","email":"nope"}`, "")
	requireStatus(t, badResp, http.StatusBadRequest)
	requireErrorCode(t, badResp, "VALIDATION_ERROR")

	// Get by ID.
	getResp := get(t, httpApp, "/api/v1/users/"+id)
	requireStatus(t, getResp, http.StatusOK)
	got := dataOf(t, decodeJSON(t, getResp.Body))
	if got["email"] != "alice@example.com" {
		t.Errorf("expected alice@example.com, got %#v", got)
	}
	getResp.Body.Close()

	// Update.
	updResp := call(t, httpApp, "PUT", "/api/v1/users/"+id, `{"name":"Alice Updated"}`, "")
	requireStatus(t, updResp, http.StatusOK)
	upd := dataOf(t, decodeJSON(t, updResp.Body))
	if got, _ := upd["name"].(string); got != "Alice Updated" {
		t.Errorf("expected updated name, got %#v", upd)
	}
	updResp.Body.Close()

	// Soft delete.
	delResp := call(t, httpApp, "DELETE", "/api/v1/users/"+id, "", "")
	requireStatus(t, delResp, http.StatusOK)
	delResp.Body.Close()

	// Gone after deletion -> 404 USER_NOT_FOUND.
	goneResp := get(t, httpApp, "/api/v1/users/"+id)
	requireStatus(t, goneResp, http.StatusNotFound)
	requireErrorCode(t, goneResp, "USER_NOT_FOUND")

	// Invalid UUID -> 400 VALIDATION_ERROR.
	invalidResp := get(t, httpApp, "/api/v1/users/not-a-uuid")
	requireStatus(t, invalidResp, http.StatusBadRequest)
	requireErrorCode(t, invalidResp, "VALIDATION_ERROR")
}

func TestProtectedRouteRequiresAuth(t *testing.T) {
	cfg := loadCfg(t)
	if !dbReachable(t, cfg) {
		t.Skip("PostgreSQL not reachable; skipping integration test")
	}
	_, httpApp := startTestApp(t, cfg)

	for _, path := range []string{"/api/v1/auth/profile", "/api/v1/realtime/sse/stats", "/api/v1/protected/data"} {
		resp := get(t, httpApp, path)
		requireStatus(t, resp, http.StatusUnauthorized)
		requireErrorCode(t, resp, "UNAUTHORIZED")
	}
}

func TestRouteNotFound(t *testing.T) {
	cfg := loadCfg(t)
	if !dbReachable(t, cfg) {
		t.Skip("PostgreSQL not reachable; skipping integration test")
	}
	_, httpApp := startTestApp(t, cfg)
	resp := get(t, httpApp, "/api/v1/does-not-exist")
	requireStatus(t, resp, http.StatusNotFound)
	requireErrorCode(t, resp, "NOT_FOUND")
}

func TestRequestIDHeader(t *testing.T) {
	cfg := loadCfg(t)
	if !dbReachable(t, cfg) {
		t.Skip("PostgreSQL not reachable; skipping integration test")
	}
	_, httpApp := startTestApp(t, cfg)
	resp := get(t, httpApp, "/health")
	defer resp.Body.Close()
	requireStatus(t, resp, http.StatusOK)
	if rid := resp.Header.Get("X-Request-ID"); rid == "" {
		t.Error("expected X-Request-ID header on responses")
	}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("X-Request-ID", "my-custom-id")
	resp2, err := httpApp.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	if got := resp2.Header.Get("X-Request-ID"); got != "my-custom-id" {
		t.Errorf("expected echoed request id, got %q", got)
	}
}

func TestProjectAndTaskLifecycle(t *testing.T) {
	cfg := loadCfg(t)
	if !dbReachable(t, cfg) {
		t.Skip("PostgreSQL not reachable; skipping integration test")
	}
	container, httpApp := startTestApp(t, cfg)
	resetUsers(t, container)

	email := fmt.Sprintf("proj-%s@example.com", randSuffix())
	password := "TestPassword123!"

	// 1. Register user & obtain JWT token
	regResp := post(t, httpApp, "/api/v1/auth/register",
		fmt.Sprintf(`{"name":"Project Tester","email":%q,"password":%q}`, email, password))
	requireStatus(t, regResp, http.StatusCreated)
	reg := decodeJSON(t, regResp.Body)
	regData := dataOf(t, reg)
	tokens := regData["tokens"].(map[string]interface{})
	token := tokens["access_token"].(string)
	regResp.Body.Close()

	// 2. Create Project (POST) - verify all fields
	projBody := `{"name":"Backend Integration","color":"#6366f1","description":"Refactoring backend and API tests"}`
	projResp := callAuth(t, httpApp, http.MethodPost, "/api/v1/projects", projBody, token)
	requireStatus(t, projResp, http.StatusCreated)
	projCreated := decodeJSON(t, projResp.Body)
	requireCode(t, projCreated, "CREATED")
	projData := dataOf(t, projCreated)
	projIDFloat := projData["id"].(float64)
	projID := int64(projIDFloat)
	if projData["name"] != "Backend Integration" {
		t.Errorf("expected name %q, got %v", "Backend Integration", projData["name"])
	}
	if projData["color"] != "#6366f1" {
		t.Errorf("expected color %q, got %v", "#6366f1", projData["color"])
	}
	if projData["description"] != "Refactoring backend and API tests" {
		t.Errorf("expected description match, got %v", projData["description"])
	}
	if projData["task_count"] != float64(0) {
		t.Errorf("expected task_count 0, got %v", projData["task_count"])
	}
	projResp.Body.Close()

	// 3. Create Task 1 linked to Project
	task1Body := fmt.Sprintf(`{"title":"Write Integration Tests","status":"in_progress","priority":"high","project_id":%d}`, projID)
	task1Resp := callAuth(t, httpApp, http.MethodPost, "/api/v1/tasks", task1Body, token)
	requireStatus(t, task1Resp, http.StatusCreated)
	task1Data := dataOf(t, decodeJSON(t, task1Resp.Body))
	if task1Data["title"] != "Write Integration Tests" {
		t.Errorf("expected task title match, got %v", task1Data["title"])
	}
	if task1Data["project_id"] != float64(projID) {
		t.Errorf("expected task project_id %d, got %v", projID, task1Data["project_id"])
	}
	task1Resp.Body.Close()

	// 4. Create Task 2 linked to Project
	task2Body := fmt.Sprintf(`{"title":"Deploy to Staging","status":"todo","priority":"medium","project_id":%d}`, projID)
	task2Resp := callAuth(t, httpApp, http.MethodPost, "/api/v1/tasks", task2Body, token)
	requireStatus(t, task2Resp, http.StatusCreated)
	task2Resp.Body.Close()

	// 5. GET /api/v1/projects -> verify real SQL count aggregation: task_count == 2
	listResp := callAuth(t, httpApp, http.MethodGet, "/api/v1/projects", "", token)
	requireStatus(t, listResp, http.StatusOK)
	listPayload := decodeJSON(t, listResp.Body)
	items, _ := listPayload["data"].([]interface{})
	if len(items) != 1 {
		t.Fatalf("expected 1 project in list, got %d", len(items))
	}
	item0 := items[0].(map[string]interface{})
	if item0["task_count"] != float64(2) {
		t.Errorf("expected task_count 2, got %v", item0["task_count"])
	}
	listResp.Body.Close()

	// 6. Update Project (PUT) - verify all updated fields & task_count preserved
	updateBody := `{"name":"Backend Integration V2","color":"#10b981","description":"Updated description"}`
	updateResp := callAuth(t, httpApp, http.MethodPut, fmt.Sprintf("/api/v1/projects/%d", projID), updateBody, token)
	requireStatus(t, updateResp, http.StatusOK)
	updateData := dataOf(t, decodeJSON(t, updateResp.Body))
	if updateData["name"] != "Backend Integration V2" {
		t.Errorf("expected updated name, got %v", updateData["name"])
	}
	if updateData["color"] != "#10b981" {
		t.Errorf("expected updated color, got %v", updateData["color"])
	}
	if updateData["description"] != "Updated description" {
		t.Errorf("expected updated description, got %v", updateData["description"])
	}
	if updateData["task_count"] != float64(2) {
		t.Errorf("expected task_count to remain 2, got %v", updateData["task_count"])
	}
	updateResp.Body.Close()

	// 7. Delete Project (DELETE)
	delResp := callAuth(t, httpApp, http.MethodDelete, fmt.Sprintf("/api/v1/projects/%d", projID), "", token)
	requireStatus(t, delResp, http.StatusOK)
	delResp.Body.Close()

	// 8. Verify Project is deleted
	getGone := callAuth(t, httpApp, http.MethodGet, fmt.Sprintf("/api/v1/projects/%d", projID), "", token)
	requireStatus(t, getGone, http.StatusNotFound)
	requireErrorCode(t, getGone, "PROJECT_NOT_FOUND")

	// 9. Verify tasks still exist
	tasksResp := callAuth(t, httpApp, http.MethodGet, "/api/v1/tasks", "", token)
	requireStatus(t, tasksResp, http.StatusOK)
	tasksData := decodeJSON(t, tasksResp.Body)["data"].([]interface{})
	if len(tasksData) != 2 {
		t.Errorf("expected 2 tasks to still exist after project deletion, got %d", len(tasksData))
	}
	tasksResp.Body.Close()
}

func TestNotesLifecycle(t *testing.T) {
	cfg := loadCfg(t)
	if !dbReachable(t, cfg) {
		t.Skip("PostgreSQL not reachable; skipping integration test")
	}
	container, httpApp := startTestApp(t, cfg)
	resetUsers(t, container)

	email := fmt.Sprintf("notes-%s@example.com", randSuffix())
	password := "TestPassword123!"

	// 1. Register user & obtain JWT token
	regResp := post(t, httpApp, "/api/v1/auth/register",
		fmt.Sprintf(`{"name":"Notes Tester","email":%q,"password":%q}`, email, password))
	requireStatus(t, regResp, http.StatusCreated)
	tokens := dataOf(t, decodeJSON(t, regResp.Body))["tokens"].(map[string]interface{})
	token := tokens["access_token"].(string)
	regResp.Body.Close()

	// 2. Create Note (POST)
	noteBody := `{"title":"Integration Architecture Idea","content":"Use testcontainers for real DB verification","color":"emerald","tags":["architecture","testing"]}`
	noteResp := callAuth(t, httpApp, http.MethodPost, "/api/v1/notes", noteBody, token)
	requireStatus(t, noteResp, http.StatusCreated)
	noteData := dataOf(t, decodeJSON(t, noteResp.Body))
	noteID := int64(noteData["id"].(float64))
	if noteData["title"] != "Integration Architecture Idea" {
		t.Errorf("expected title match, got %v", noteData["title"])
	}
	if noteData["color"] != "emerald" {
		t.Errorf("expected color emerald, got %v", noteData["color"])
	}
	noteResp.Body.Close()

	// 3. Convert Note to Task (POST)
	convertBody := `{"priority":"urgent"}`
	convertResp := callAuth(t, httpApp, http.MethodPost, fmt.Sprintf("/api/v1/notes/%d/convert", noteID), convertBody, token)
	requireStatus(t, convertResp, http.StatusOK)
	convertData := dataOf(t, decodeJSON(t, convertResp.Body))
	taskID := int64(convertData["task_id"].(float64))
	noteObj := convertData["note"].(map[string]interface{})
	if noteObj["title"] != "Integration Architecture Idea" {
		t.Errorf("expected note title match, got %v", noteObj["title"])
	}
	convertResp.Body.Close()

	// 4. Verify Note now references the converted task
	getNoteResp := callAuth(t, httpApp, http.MethodGet, fmt.Sprintf("/api/v1/notes/%d", noteID), "", token)
	requireStatus(t, getNoteResp, http.StatusOK)
	getNoteData := dataOf(t, decodeJSON(t, getNoteResp.Body))
	if getNoteData["converted_task_id"] != float64(taskID) {
		t.Errorf("expected converted_task_id %d, got %v", taskID, getNoteData["converted_task_id"])
	}
	getNoteResp.Body.Close()

	// 5. Delete Note
	delResp := callAuth(t, httpApp, http.MethodDelete, fmt.Sprintf("/api/v1/notes/%d", noteID), "", token)
	requireStatus(t, delResp, http.StatusOK)
	delResp.Body.Close()
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func randSuffix() string {
	const alpha = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	var seed uint64 = 0x0badc0de
	for i := len(b) - 1; i >= 0; i-- {
		seed = seed*6364136223846793005 + 1442695040888963407
		b[i] = alpha[(seed>>33)%uint64(len(alpha))]
	}
	return string(b)
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func getEnvAsBool(key string, def bool) bool {
	switch os.Getenv(key) {
	case "1", "true", "TRUE", "yes":
		return true
	case "0", "false", "FALSE", "no":
		return false
	}
	return def
}
