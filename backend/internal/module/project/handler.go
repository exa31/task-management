package project

import (
	"strconv"

	authMiddleware "golang/pkg/middleware/auth"
	"golang/pkg/response"
	"golang/pkg/validation"

	"github.com/gofiber/fiber/v2"
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

	var req CreateProjectRequest
	if err := validation.Default.BindAndValidate(c, &req); err != nil {
		return err
	}

	res, err := h.service.Create(c.UserContext(), userID, &req)
	if err != nil {
		return err
	}

	return response.CreatedResponse(c, res, "Project created successfully")
}

func (h *Handler) List(c *fiber.Ctx) error {
	userID, err := authMiddleware.GetUserID(c)
	if err != nil {
		return err
	}

	items, err := h.service.List(c.UserContext(), userID)
	if err != nil {
		return err
	}

	return response.SuccessResponse(c, items, "Projects listed successfully")
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	userID, err := authMiddleware.GetUserID(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid project ID")
	}

	res, err := h.service.GetByID(c.UserContext(), userID, id)
	if err != nil {
		return err
	}

	return response.SuccessResponse(c, res, "Project retrieved successfully")
}

func (h *Handler) Update(c *fiber.Ctx) error {
	userID, err := authMiddleware.GetUserID(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid project ID")
	}

	var req UpdateProjectRequest
	if err := validation.Default.BindAndValidate(c, &req); err != nil {
		return err
	}

	res, err := h.service.Update(c.UserContext(), userID, id, &req)
	if err != nil {
		return err
	}

	return response.SuccessResponse(c, res, "Project updated successfully")
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	userID, err := authMiddleware.GetUserID(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid project ID")
	}

	if err := h.service.Delete(c.UserContext(), userID, id); err != nil {
		return err
	}

	return response.SuccessMessageResponse(c, "Project deleted successfully")
}
