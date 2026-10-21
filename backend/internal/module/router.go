// Package module contains the feature modules of the application. Every
// module is self-contained: it owns its DTOs, repository, service and
// handler, and registers its own routes.
package module

import (
	"golang/config"
	"golang/database"
	authmodule "golang/internal/module/auth"
	graphqlmodule "golang/internal/module/graphql"
	"golang/internal/module/mcp"
	"golang/internal/module/note"
	"golang/internal/module/project"
	"golang/internal/module/reminder"
	"golang/internal/module/task"
	usermodule "golang/internal/module/user"
	"golang/pkg/auth"
	"golang/pkg/logging"
	authMiddleware "golang/pkg/middleware/auth"
	"golang/pkg/notification"
	"golang/pkg/realtime"
	"golang/pkg/storage"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// Router bundles module dependencies and registers routes under a base group.
type Router struct {
	DB              *database.PostgresDB
	JWTService      *auth.JWTService
	Hub             *realtime.Hub // nil when the realtime feature is disabled
	Config          *config.Config
	Storage         storage.Storage
	ReminderService reminder.Service
	AsynqScheduler  *notification.AsynqScheduler
	EmailSender     notification.EmailSender
	PushSender      notification.PushSender
	ReminderRepo    reminder.Repository
}

// NewRouter builds the module router.
func NewRouter(
	db *database.PostgresDB,
	jwtService *auth.JWTService,
	hub *realtime.Hub,
	cfg *config.Config,
	stor storage.Storage,
	reminderSvc reminder.Service,
	asynqSched *notification.AsynqScheduler,
	emailSender notification.EmailSender,
	pushSender notification.PushSender,
	reminderRepo reminder.Repository,
) *Router {
	return &Router{
		DB:              db,
		JWTService:      jwtService,
		Hub:             hub,
		Config:          cfg,
		Storage:         stor,
		ReminderService: reminderSvc,
		AsynqScheduler:  asynqSched,
		EmailSender:     emailSender,
		PushSender:      pushSender,
		ReminderRepo:    reminderRepo,
	}
}

// Register mounts every enabled module on the API router.
func (r *Router) Register(api fiber.Router) {
	r.registerAuth(api)
	r.registerUsers(api)
	r.registerNotesAndTasks(api)
	r.registerReminders(api)
	r.registerProtected(api)
	if r.Hub != nil {
		r.registerRealtime(api)
	}
}

func (r *Router) registerAuth(api fiber.Router) {
	handler := authmodule.NewAuthHandler(authmodule.NewAuthService(r.DB, r.JWTService, r.Config, r.Storage))

	auth := api.Group("/auth")
	auth.Post("/register", handler.Register())
	auth.Post("/login", handler.Login())
	auth.Post("/oauth", handler.OAuthLogin())
	auth.Post("/apple/webhook", handler.AppleWebhook())
	auth.Post("/dev-login", handler.DevLogin())
	auth.Post("/refresh", handler.RefreshToken())

	auth.Use(authMiddleware.AuthMiddleware(r.JWTService))
	auth.Get("/profile", handler.GetProfile())
	auth.Put("/profile", handler.UpdateProfile())
	auth.Post("/avatar", handler.UploadAvatar())
	auth.Post("/bind-oauth", handler.BindOAuth())
	auth.Post("/unbind-oauth", handler.UnbindOAuth())
	auth.Delete("/account", handler.DeleteAccount())
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
	if r.AsynqScheduler != nil {
		logger := logging.Logger()
		taskService.SetNotificationDependencies(r.AsynqScheduler, r.EmailSender, r.PushSender, r.ReminderRepo, r.Hub, logger)
	}
	taskHandler := task.NewHandler(taskService)

	// Cross-wire note conversion to task
	noteService.SetTaskCreator(taskService)

	// Initialize MCP
	mcpTokenRepo := mcp.NewTokenRepository(db)
	mcpService := mcp.NewService(mcpTokenRepo, noteService, taskService, r.DB)
	if r.ReminderService != nil {
		mcpService.SetReminderService(r.ReminderService)
	}
	mcpHandler := mcp.NewHandler(mcpService)

	webhookSecret := ""
	if r.Config != nil {
		webhookSecret = r.Config.Reminder.WebhookSecret
	}
	// External Asynq worker webhook trigger endpoint for task due warning (Public/Secret-protected)
	api.Post("/tasks/due-warning/trigger", taskHandler.TriggerDueWarning(webhookSecret))
	// External Asynq worker webhook trigger endpoint for recurring task midnight rollover (Public/Secret-protected)
	api.Post("/tasks/recurrence/trigger", taskHandler.TriggerRecurrence(webhookSecret))

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
	tasks.Get("/widgets/bundle", taskHandler.GetWidgetBundle)
	tasks.Get("/stats", taskHandler.GetWorkspaceStats)
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
	tasks.Get("/:id/activities", taskHandler.GetActivities)
	tasks.Post("/:id/activities", taskHandler.CreateActivity)

	// Recurring Tasks routes (Per-user isolated master templates)
	recurringTasks := api.Group("/recurring-tasks")
	recurringTasks.Use(authMiddleware.AuthMiddleware(r.JWTService))
	recurringTasks.Get("/", taskHandler.ListRecurringTasks)
	recurringTasks.Post("/", taskHandler.CreateRecurringTask)
	recurringTasks.Get("/:id", taskHandler.GetRecurringTask)
	recurringTasks.Put("/:id", taskHandler.UpdateRecurringTask)
	recurringTasks.Delete("/:id", taskHandler.DeleteRecurringTask)
	recurringTasks.Post("/:id/toggle-active", taskHandler.ToggleActiveRecurringTask)

	// Workspace routes (Per-user isolated)
	workspace := api.Group("/workspace")
	workspace.Use(authMiddleware.AuthMiddleware(r.JWTService))
	workspace.Get("/stats", taskHandler.GetWorkspaceStats)

	// Projects routes (Per-user isolated)
	projectRepo := project.NewRepository(db)
	projectService := project.NewService(projectRepo)
	projectHandler := project.NewHandler(projectService)
	mcpService.SetProjectService(projectService)

	projects := api.Group("/projects")
	projects.Use(authMiddleware.AuthMiddleware(r.JWTService))
	projects.Post("/", projectHandler.Create)
	projects.Get("/", projectHandler.List)
	projects.Get("/:id", projectHandler.GetByID)
	projects.Put("/:id", projectHandler.Update)
	projects.Delete("/:id", projectHandler.Delete)

	// MCP and API Tokens routes
	tokens := api.Group("/tokens")
	tokens.Use(authMiddleware.AuthMiddleware(r.JWTService))
	tokens.Post("/", mcpHandler.GenerateToken)
	tokens.Get("/", mcpHandler.ListTokens)
	tokens.Delete("/:id", mcpHandler.DeleteToken)

	mcpGroup := api.Group("/mcp")
	mcpGroup.Get("/", mcpHandler.Discovery)
	mcpGroup.Post("/", authMiddleware.OptionalAuthMiddleware(r.JWTService), mcpHandler.HandleMCP)

	// Hub modular life aggregator summary contract
	api.Get("/hub/summary", authMiddleware.OptionalAuthMiddleware(r.JWTService), mcpHandler.GetHubSummary)

	// GraphQL route (Playground on GET in browser, JSON execution on POST)
	userRepo := usermodule.NewUserRepository(db)
	reminderRepo := r.ReminderRepo
	if reminderRepo == nil {
		reminderRepo = reminder.NewRepository(db)
	}
	gqlSchema, err := graphqlmodule.BuildSchema(graphqlmodule.SchemaConfig{
		UserRepo:     userRepo,
		ProjectRepo:  projectRepo,
		TaskRepo:     taskRepo,
		NoteRepo:     noteRepo,
		ReminderRepo: reminderRepo,
		DB:           r.DB,
	})
	if err == nil {
		gqlHandler := graphqlmodule.NewHandler(gqlSchema, r.JWTService)
		api.Get("/graphql", gqlHandler.Handle)
		api.Post("/graphql", authMiddleware.AuthMiddleware(r.JWTService), gqlHandler.Handle)
	}
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
	sse.Get("/events", authMiddleware.AuthMiddleware(r.JWTService), sseHandler.Events)
	sse.Get("/subscribe", authMiddleware.AuthMiddleware(r.JWTService), sseHandler.Subscribe)
	sse.Post("/broadcast", authMiddleware.AuthMiddleware(r.JWTService), sseHandler.BroadcastToChannel)
	sse.Post("/send", authMiddleware.AuthMiddleware(r.JWTService), sseHandler.SendToUser)
	sse.Get("/stats", authMiddleware.AuthMiddleware(r.JWTService), sseHandler.GetStats)
}

func (r *Router) registerReminders(api fiber.Router) {
	if r.ReminderService == nil {
		return
	}
	webhookSecret := ""
	isDev := false
	if r.Config != nil {
		webhookSecret = r.Config.Reminder.WebhookSecret
		isDev = r.Config.App.Environment != "production"
	}
	handler := reminder.NewHandler(r.ReminderService, webhookSecret, isDev)
	reminders := api.Group("/reminders")

	// External Asynq worker webhook trigger endpoint (Public/Secret-protected)
	reminders.Post("/trigger", handler.Trigger)

	// User-authenticated endpoints
	reminders.Use(authMiddleware.AuthMiddleware(r.JWTService))
	reminders.Post("/", handler.Create)
	reminders.Get("/", handler.List)
	reminders.Get("/:id", handler.GetByID)
	reminders.Put("/:id", handler.Update)
	reminders.Delete("/:id", handler.Delete)
	reminders.Post("/:id/snooze", handler.Snooze)
	reminders.Post("/:id/dismiss", handler.Dismiss)
	reminders.Post("/device-token", handler.RegisterDeviceToken)
	reminders.Post("/test-fcm", handler.TestFCM)

	// Notifications alias for convenience
	notifications := api.Group("/notifications")
	notifications.Use(authMiddleware.AuthMiddleware(r.JWTService))
	notifications.Post("/test-fcm", handler.TestFCM)
	notifications.Post("/device-token", handler.RegisterDeviceToken)
}
