// Package module contains the feature modules of the application. Every
// module is self-contained: it owns its DTOs, repository, service and
// handler, and registers its own routes.
package module

import (
	"golang/config"
	"golang/database"
	authmodule "golang/internal/module/auth"
	"golang/internal/module/mcp"
	"golang/internal/module/note"
	"golang/internal/module/task"
	usermodule "golang/internal/module/user"
	"golang/pkg/auth"
	authMiddleware "golang/pkg/middleware/auth"
	"golang/pkg/realtime"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// Router bundles module dependencies and registers routes under a base group.
type Router struct {
	DB         *database.PostgresDB
	JWTService *auth.JWTService
	Hub        *realtime.Hub // nil when the realtime feature is disabled
	Config     *config.Config
}

// NewRouter builds the module router.
func NewRouter(db *database.PostgresDB, jwtService *auth.JWTService, hub *realtime.Hub, cfg *config.Config) *Router {
	return &Router{DB: db, JWTService: jwtService, Hub: hub, Config: cfg}
}

// Register mounts every enabled module on the API router.
func (r *Router) Register(api fiber.Router) {
	r.registerAuth(api)
	r.registerUsers(api)
	r.registerNotesAndTasks(api)
	r.registerProtected(api)
	if r.Hub != nil {
		r.registerRealtime(api)
	}
}

func (r *Router) registerAuth(api fiber.Router) {
	handler := authmodule.NewAuthHandler(authmodule.NewAuthService(r.DB, r.JWTService, r.Config))

	auth := api.Group("/auth")
	auth.Post("/oauth", handler.OAuthLogin())
	auth.Post("/apple/webhook", handler.AppleWebhook())
	auth.Post("/dev-login", handler.DevLogin())
	auth.Post("/refresh", handler.RefreshToken())

	auth.Use(authMiddleware.AuthMiddleware(r.JWTService))
	auth.Get("/profile", handler.GetProfile())
	auth.Put("/profile", handler.UpdateProfile())
	auth.Post("/logout", handler.Logout())
	auth.Post("/logout-all", handler.LogoutAll())
}

func (r *Router) registerUsers(api fiber.Router) {
	repo := usermodule.NewUserRepository(r.DB.GetDB())
	service := usermodule.NewUserService(repo)
	handler := usermodule.NewUserHandler(service)

	users := api.Group("/users")
	users.Get("/", handler.GetUsers())
	users.Get("/:id", handler.GetUser())
	users.Post("/", handler.CreateUser())
	users.Put("/:id", handler.UpdateUser())
	users.Delete("/:id", handler.DeleteUser())

	admin := users.Group("/admin")
	admin.Delete("/:id", handler.HardDeleteUser())
}

func (r *Router) registerNotesAndTasks(api fiber.Router) {
	db := r.DB.GetDB()

	// Initialize repos & services
	noteRepo := note.NewRepository(db)
	noteService := note.NewService(noteRepo, r.DB)
	noteHandler := note.NewHandler(noteService)

	taskRepo := task.NewRepository(db)
	taskService := task.NewService(taskRepo, r.DB)
	taskHandler := task.NewHandler(taskService)

	// Cross-wire note conversion to task
	noteService.SetTaskCreator(taskService)

	// Initialize MCP
	mcpTokenRepo := mcp.NewTokenRepository(db)
	mcpService := mcp.NewService(mcpTokenRepo, noteService, taskService, r.DB)
	mcpHandler := mcp.NewHandler(mcpService)

	// Notes routes (Per-user isolated)
	notes := api.Group("/notes")
	notes.Use(authMiddleware.AuthMiddleware(r.JWTService))
	notes.Post("/", noteHandler.Create)
	notes.Get("/", noteHandler.List)
	notes.Get("/:id", noteHandler.GetByID)
	notes.Put("/:id", noteHandler.Update)
	notes.Delete("/:id", noteHandler.Delete)
	notes.Post("/:id/convert", noteHandler.ConvertToTask)

	// Tasks routes (Per-user isolated)
	tasks := api.Group("/tasks")
	tasks.Use(authMiddleware.AuthMiddleware(r.JWTService))
	tasks.Post("/", taskHandler.Create)
	tasks.Get("/", taskHandler.List)
	tasks.Get("/:id", taskHandler.GetByID)
	tasks.Put("/:id", taskHandler.Update)
	tasks.Delete("/:id", taskHandler.Delete)
	tasks.Post("/:id/progress", taskHandler.LogProgress)
	tasks.Post("/:id/focus", taskHandler.ToggleFocus)
	tasks.Post("/:id/subtasks", taskHandler.CreateSubtask)
	tasks.Put("/:id/subtasks/:subtaskId", taskHandler.UpdateSubtask)
	tasks.Delete("/:id/subtasks/:subtaskId", taskHandler.DeleteSubtask)

	// MCP and API Tokens routes
	tokens := api.Group("/tokens")
	tokens.Use(authMiddleware.AuthMiddleware(r.JWTService))
	tokens.Post("/", mcpHandler.GenerateToken)
	tokens.Get("/", mcpHandler.ListTokens)
	tokens.Delete("/:id", mcpHandler.DeleteToken)

	mcpGroup := api.Group("/mcp")
	mcpGroup.Get("/", mcpHandler.Discovery)
	mcpGroup.Post("/", authMiddleware.OptionalAuthMiddleware(r.JWTService), mcpHandler.HandleMCP)
}

func (r *Router) registerProtected(api fiber.Router) {
	protected := api.Group("/protected")
	protected.Use(authMiddleware.AuthMiddleware(r.JWTService))
	protected.Get("/data", func(c *fiber.Ctx) error {
		userID, _ := authMiddleware.GetUserID(c)
		email, _ := authMiddleware.GetEmail(c)
		return c.JSON(fiber.Map{
			"message": "This is protected data",
			"user_id": userID,
			"email":   email,
		})
	})

	api.Get("/public-or-private", authMiddleware.OptionalAuthMiddleware(r.JWTService), func(c *fiber.Ctx) error {
		if userID, err := authMiddleware.GetUserID(c); err == nil {
			return c.JSON(fiber.Map{"message": "Private data", "authenticated": true, "user_id": userID})
		}
		return c.JSON(fiber.Map{"message": "Public data", "authenticated": false})
	})
}

func (r *Router) registerRealtime(api fiber.Router) {
	rt := api.Group("/realtime")

	if r.Config.Feature.RealtimeWS {
		r.registerWebSocketRoutes(rt)
	}
	if r.Config.Feature.RealtimeSSE {
		r.registerSSERoutes(rt)
	}
}

func (r *Router) registerWebSocketRoutes(rt fiber.Router) {
	wsHandler := NewWSHandler(r.Hub)
	ws := rt.Group("/ws")
	ws.Get("/connect", upgradeWS(), websocket.New(wsHandler.Handle))
	ws.Get("/stats", authMiddleware.AuthMiddleware(r.JWTService), wsHandler.GetStats)
}

func (r *Router) registerSSERoutes(rt fiber.Router) {
	sseHandler := NewSSEHandler(r.Hub)
	sse := rt.Group("/sse")
	sse.Get("/events", sseHandler.Events)
	sse.Get("/subscribe", sseHandler.Subscribe)
	sse.Post("/broadcast", authMiddleware.AuthMiddleware(r.JWTService), sseHandler.BroadcastToChannel)
	sse.Post("/send", authMiddleware.AuthMiddleware(r.JWTService), sseHandler.SendToUser)
	sse.Get("/stats", authMiddleware.AuthMiddleware(r.JWTService), sseHandler.GetStats)
}
