package note

import (
	"strconv"

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

	var req CreateNoteRequest
	if err := validation.Default.BindAndValidate(c, &req); err != nil {
		return err
	}

	result, err := h.service.Create(c.UserContext(), userID, req)
	if err != nil {
		return err
	}

	return response.CreatedResponse(c, result, "Note created successfully")
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	userID, err := authMiddleware.GetUserID(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid note ID")
	}

	result, err := h.service.GetByID(c.UserContext(), userID, id)
	if err != nil {
		return err
	}

	return response.SuccessResponse(c, result, "Note retrieved successfully")
}

func (h *Handler) List(c *fiber.Ctx) error {
	userID, err := authMiddleware.GetUserID(c)
	if err != nil {
		return err
	}

	var filter NoteFilter
	if err := validation.Default.BindQueryAndValidate(c, &filter); err != nil {
		return err
	}

	results, err := h.service.FindAll(c.UserContext(), userID, filter)
	if err != nil {
		return err
	}

	return response.SuccessResponse(c, results, "Notes listed successfully")
}

func (h *Handler) Update(c *fiber.Ctx) error {
	userID, err := authMiddleware.GetUserID(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid note ID")
	}

	var req UpdateNoteRequest
	if err := validation.Default.BindAndValidate(c, &req); err != nil {
		return err
	}

	result, err := h.service.Update(c.UserContext(), userID, id, req)
	if err != nil {
		return err
	}

	return response.SuccessResponse(c, result, "Note updated successfully")
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	userID, err := authMiddleware.GetUserID(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid note ID")
	}

	if err := h.service.Delete(c.UserContext(), userID, id); err != nil {
		return err
	}

	return response.SuccessMessageResponse(c, "Note deleted successfully")
}

func (h *Handler) ConvertToTask(c *fiber.Ctx) error {
	userID, err := authMiddleware.GetUserID(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid note ID")
	}

	var req ConvertToTaskRequest
	if err := validation.Default.BindAndValidate(c, &req); err != nil {
		return err
	}

	note, taskID, err := h.service.ConvertToTask(c.UserContext(), userID, id, req)
	if err != nil {
		return err
	}

	return response.SuccessResponse(c, fiber.Map{
		"note":    note,
		"task_id": taskID,
	}, "Note converted to task successfully")
}
