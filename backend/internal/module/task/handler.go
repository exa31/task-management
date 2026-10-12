package task

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	authMiddleware "golang/pkg/middleware/auth"
	"golang/pkg/response"
	"golang/pkg/validation"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(c *fiber.Ctx) error {
	userID, err := authMiddleware.GetUserID(c)
	if err != nil {
		return err
	}

	var req CreateTaskRequest
	if err := validation.Default.BindAndValidate(c, &req); err != nil {
		return err
	}

	result, err := h.service.Create(c.UserContext(), userID, req)
	if err != nil {
		return err
	}

	return response.CreatedResponse(c, result, "Task created successfully")
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	userID, err := authMiddleware.GetUserID(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid task ID")
	}

	result, err := h.service.GetByID(c.UserContext(), userID, id)
	if err != nil {
		return err
	}

	return response.SuccessResponse(c, result, "Task retrieved successfully")
}

func (h *Handler) List(c *fiber.Ctx) error {
	userID, err := authMiddleware.GetUserID(c)
	if err != nil {
		return err
	}

	var filter TaskFilter
	if err := validation.Default.BindQueryAndValidate(c, &filter); err != nil {
		return err
	}

	isCursorPaginated := filter.Paginate == "cursor" || filter.Cursor != nil
	pageSize := filter.Limit
	if isCursorPaginated {
		if pageSize <= 0 {
			pageSize = 15
		}
		// Request 1 extra item to check for next page
		filter.Limit = pageSize + 1
	}

	results, err := h.service.FindAll(c.UserContext(), userID, filter)
	if err != nil {
		return err
	}

	if isCursorPaginated {
		hasMore := false
		var nextCursor *int64
		if len(results) > pageSize {
			hasMore = true
			results = results[:pageSize]
			lastID := results[len(results)-1].ID
			nextCursor = &lastID
		}
		cursorResp := TaskCursorResponse{
			Items:      results,
			NextCursor: nextCursor,
			HasMore:    hasMore,
		}
		return response.SuccessResponse(c, cursorResp, "Tasks listed successfully")
	}

	return response.SuccessResponse(c, results, "Tasks listed successfully")
}

func (h *Handler) Update(c *fiber.Ctx) error {
	userID, err := authMiddleware.GetUserID(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid task ID")
	}

	var req UpdateTaskRequest
	if err := validation.Default.BindAndValidate(c, &req); err != nil {
		return err
	}

	result, err := h.service.Update(c.UserContext(), userID, id, req)
	if err != nil {
		return err
	}

	return response.SuccessResponse(c, result, "Task updated successfully")
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	userID, err := authMiddleware.GetUserID(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid task ID")
	}

	if err := h.service.Delete(c.UserContext(), userID, id); err != nil {
		return err
	}

	return response.SuccessMessageResponse(c, "Task deleted successfully")
}

func (h *Handler) LogProgress(c *fiber.Ctx) error {
	userID, err := authMiddleware.GetUserID(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid task ID")
	}

	var req LogProgressRequest
	if err := validation.Default.BindAndValidate(c, &req); err != nil {
		return err
	}

	result, err := h.service.LogProgress(c.UserContext(), userID, id, req)
	if err != nil {
		return err
	}

	return response.SuccessResponse(c, result, "Task progress logged successfully")
}

func (h *Handler) ToggleFocus(c *fiber.Ctx) error {
	userID, err := authMiddleware.GetUserID(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid task ID")
	}

	result, err := h.service.ToggleFocus(c.UserContext(), userID, id)
	if err != nil {
		return err
	}

	return response.SuccessResponse(c, result, "Task focus toggled successfully")
}

func (h *Handler) CreateSubtask(c *fiber.Ctx) error {
	userID, err := authMiddleware.GetUserID(c)
	if err != nil {
		return err
	}

	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid task ID")
	}

	var req CreateSubtaskRequest
	if err := validation.Default.BindAndValidate(c, &req); err != nil {
		return err
	}

	result, err := h.service.CreateSubtask(c.UserContext(), userID, taskID, req)
	if err != nil {
		return err
	}

	return response.CreatedResponse(c, result, "Subtask added successfully")
}

func (h *Handler) UpdateSubtask(c *fiber.Ctx) error {
	userID, err := authMiddleware.GetUserID(c)
	if err != nil {
		return err
	}

	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid task ID")
	}

	subtaskID, err := strconv.ParseInt(c.Params("subtaskId"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid subtask ID")
	}

	var req UpdateSubtaskRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid payload")
	}

	result, err := h.service.UpdateSubtask(c.UserContext(), userID, taskID, subtaskID, req)
	if err != nil {
		return err
	}

	return response.SuccessResponse(c, result, "Subtask updated successfully")
}

func (h *Handler) DeleteSubtask(c *fiber.Ctx) error {
	userID, err := authMiddleware.GetUserID(c)
	if err != nil {
		return err
	}

	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid task ID")
	}

	subtaskID, err := strconv.ParseInt(c.Params("subtaskId"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid subtask ID")
	}

	if err := h.service.DeleteSubtask(c.UserContext(), userID, taskID, subtaskID); err != nil {
		return err
	}

	return response.SuccessMessageResponse(c, "Subtask deleted successfully")
}

func (h *Handler) GetActivities(c *fiber.Ctx) error {
	userID, err := authMiddleware.GetUserID(c)
	if err != nil {
		return err
	}

	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid task ID")
	}

	results, err := h.service.GetActivities(c.UserContext(), userID, taskID)
	if err != nil {
		return err
	}

	return response.SuccessResponse(c, results, "Task activities retrieved successfully")
}

func (h *Handler) CreateActivity(c *fiber.Ctx) error {
	userID, err := authMiddleware.GetUserID(c)
	if err != nil {
		return err
	}

	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid task ID")
	}

	var req CreateTaskActivityRequest
	if err := validation.Default.BindAndValidate(c, &req); err != nil {
		return err
	}

	result, err := h.service.CreateActivity(c.UserContext(), userID, taskID, req)
	if err != nil {
		return err
	}

	return response.CreatedResponse(c, result, "Task activity recorded successfully")
}

// TriggerDueWarning is the webhook endpoint invoked by the external Asynq worker via HTTP POST.
func (h *Handler) TriggerDueWarning(webhookSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
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

		if incomingSecret == "" || webhookSecret == "" || incomingSecret != webhookSecret {
			return fiber.NewError(fiber.StatusUnauthorized, "Unauthorized: missing or invalid webhook secret")
		}

		var req TriggerDueWarningRequest
		_ = c.QueryParser(&req)
		_ = c.BodyParser(&req)

		targetID := req.TargetID()
		if targetID <= 0 {
			return fiber.NewError(fiber.StatusBadRequest, "task_id is required")
		}

		if err := h.service.ProcessDueWarning(c.UserContext(), targetID); err != nil {
			return err
		}

		return response.SuccessMessageResponse(c, "Task due warning processed successfully")
	}
}
