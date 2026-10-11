package reminder

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	authMiddleware "golang/pkg/middleware/auth"
	"golang/pkg/response"
	"golang/pkg/validation"
)

type Handler struct {
	service       Service
	webhookSecret string
}

func NewHandler(service Service, webhookSecret string) *Handler {
	return &Handler{
		service:       service,
		webhookSecret: webhookSecret,
	}
}

// Trigger is the webhook endpoint invoked by the external Asynq worker via HTTP POST.
// Secret verification from headers is MANDATORY (not optional).
func (h *Handler) Trigger(c *fiber.Ctx) error {
	incomingSecret := c.Get("X-Internal-Token")
	if incomingSecret == "" {
		incomingSecret = c.Get("X-Webhook-Secret")
	}
	if incomingSecret == "" {
		incomingSecret = c.Get("X-Reminder-Secret")
	}
	if incomingSecret == "" {
		authHeader := c.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			incomingSecret = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			incomingSecret = authHeader
		}
	}

	if incomingSecret == "" || h.webhookSecret == "" || incomingSecret != h.webhookSecret {
		return fiber.NewError(fiber.StatusUnauthorized, "Unauthorized: missing or invalid webhook secret")
	}

	var req TriggerReminderRequest
	_ = c.QueryParser(&req)
	_ = c.BodyParser(&req)

	targetID := req.TargetID()
	if targetID <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "reminder_id is required")
	}

	if err := h.service.ProcessReminder(c.UserContext(), targetID); err != nil {
		return err
	}

	return response.SuccessMessageResponse(c, "Reminder processed successfully")
}

func (h *Handler) Create(c *fiber.Ctx) error {
	userID, err := authMiddleware.GetUserID(c)
	if err != nil {
		return err
	}

	var req CreateReminderRequest
	if err := validation.Default.BindAndValidate(c, &req); err != nil {
		return err
	}

	result, err := h.service.Create(c.UserContext(), userID, req)
	if err != nil {
		return err
	}

	return response.CreatedResponse(c, result, "Reminder created successfully")
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	userID, err := authMiddleware.GetUserID(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid reminder ID")
	}

	result, err := h.service.GetByID(c.UserContext(), userID, id)
	if err != nil {
		return err
	}

	return response.SuccessResponse(c, result, "Reminder retrieved successfully")
}

func (h *Handler) List(c *fiber.Ctx) error {
	userID, err := authMiddleware.GetUserID(c)
	if err != nil {
		return err
	}

	var filter ReminderFilter
	if err := validation.Default.BindQueryAndValidate(c, &filter); err != nil {
		return err
	}

	results, total, err := h.service.List(c.UserContext(), userID, filter)
	if err != nil {
		return err
	}

	page := 1
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	totalPages := 0
	if limit > 0 {
		totalPages = (total + limit - 1) / limit
	}

	return response.PaginatedSuccessResponse(c, response.PaginatedResponse[ReminderResponse]{
		Items:      results,
		Total:      int64(total),
		Page:       page,
		PerPage:    limit,
		TotalPages: totalPages,
	}, "Reminders retrieved successfully")
}

func (h *Handler) Update(c *fiber.Ctx) error {
	userID, err := authMiddleware.GetUserID(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid reminder ID")
	}

	var req UpdateReminderRequest
	if err := validation.Default.BindAndValidate(c, &req); err != nil {
		return err
	}

	result, err := h.service.Update(c.UserContext(), userID, id, req)
	if err != nil {
		return err
	}

	return response.SuccessResponse(c, result, "Reminder updated successfully")
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	userID, err := authMiddleware.GetUserID(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid reminder ID")
	}

	if err := h.service.Delete(c.UserContext(), userID, id); err != nil {
		return err
	}

	return response.SuccessMessageResponse(c, "Reminder deleted successfully")
}

func (h *Handler) Snooze(c *fiber.Ctx) error {
	userID, err := authMiddleware.GetUserID(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid reminder ID")
	}

	var req SnoozeReminderRequest
	if err := validation.Default.BindAndValidate(c, &req); err != nil {
		return err
	}

	result, err := h.service.Snooze(c.UserContext(), userID, id, req)
	if err != nil {
		return err
	}

	return response.SuccessResponse(c, result, "Reminder snoozed successfully")
}

func (h *Handler) Dismiss(c *fiber.Ctx) error {
	userID, err := authMiddleware.GetUserID(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid reminder ID")
	}

	result, err := h.service.Dismiss(c.UserContext(), userID, id)
	if err != nil {
		return err
	}

	return response.SuccessResponse(c, result, "Reminder dismissed successfully")
}

func (h *Handler) RegisterDeviceToken(c *fiber.Ctx) error {
	userID, err := authMiddleware.GetUserID(c)
	if err != nil {
		return err
	}

	var req RegisterDeviceTokenRequest
	if err := validation.Default.BindAndValidate(c, &req); err != nil {
		return err
	}

	if err := h.service.RegisterDeviceToken(c.UserContext(), userID, req); err != nil {
		return err
	}

	return response.SuccessMessageResponse(c, "Device token registered successfully")
}
