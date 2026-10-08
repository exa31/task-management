package mcp

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

// GenerateToken creates a new Personal Access Token for the authenticated user.
func (h *Handler) GenerateToken(c *fiber.Ctx) error {
	userID, err := authMiddleware.GetUserID(c)
	if err != nil {
		return err
	}

	var req GenerateTokenRequest
	if err := validation.Default.BindAndValidate(c, &req); err != nil {
		return err
	}

	result, err := h.service.CreateToken(c.UserContext(), userID, req.Name)
	if err != nil {
		return err
	}

	return response.CreatedResponse(c, result, "API Token generated successfully. Copy it now, it will not be shown again.")
}

// ListTokens lists all personal access tokens for the user.
func (h *Handler) ListTokens(c *fiber.Ctx) error {
	userID, err := authMiddleware.GetUserID(c)
	if err != nil {
		return err
	}

	tokens, err := h.service.ListTokens(c.UserContext(), userID)
	if err != nil {
		return err
	}

	return response.SuccessResponse(c, tokens, "Tokens listed successfully")
}

// DeleteToken revokes a token.
func (h *Handler) DeleteToken(c *fiber.Ctx) error {
	userID, err := authMiddleware.GetUserID(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid token ID")
	}

	if err := h.service.DeleteToken(c.UserContext(), userID, id); err != nil {
		return err
	}

	return response.SuccessMessageResponse(c, "Token revoked successfully")
}

// HandleMCP handles Model Context Protocol JSON-RPC 2.0 requests over HTTP.
// Authenticates via Personal Access Token (Bearer mcp_pat_...) or active JWT.
func (h *Handler) HandleMCP(c *fiber.Ctx) error {
	var userID string

	authHeader := c.Get("Authorization")
	parts := strings.Split(authHeader, " ")
	if len(parts) == 2 && parts[0] == "Bearer" {
		tokenStr := parts[1]
		if strings.HasPrefix(tokenStr, "mcp_pat_") {
			resolvedID, err := h.service.ValidateToken(c.UserContext(), tokenStr)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"jsonrpc": "2.0",
					"error": fiber.Map{
						"code":    -32000,
						"message": "Invalid or revoked MCP Personal Access Token",
					},
				})
			}
			userID = resolvedID
		}
	}

	// Fallback to JWT context if present (e.g. testing from web UI)
	if userID == "" {
		if jwtUserID, err := authMiddleware.GetUserID(c); err == nil && jwtUserID != "" {
			userID = jwtUserID
		}
	}

	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"jsonrpc": "2.0",
			"error": fiber.Map{
				"code":    -32000,
				"message": "Authentication required. Provide a valid 'Authorization: Bearer <mcp_pat_token>'",
			},
		})
	}

	var req JSONRPCRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"jsonrpc": "2.0",
			"error": fiber.Map{
				"code":    -32700,
				"message": "Parse error: invalid JSON-RPC payload",
			},
		})
	}

	res := h.service.HandleJSONRPC(c.UserContext(), userID, req)
	return c.JSON(res)
}

// Discovery returns metadata and instructions for setting up MCP clients.
func (h *Handler) Discovery(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"name":        "Task Management Model Context Protocol Server",
		"version":     "1.0.0",
		"endpoint":    "/api/v1/mcp",
		"auth_type":   "Bearer Personal Access Token",
		"description": "Connect AI tools (Antigravity, Cursor, Claude) to read/manage notes and tasks.",
	})
}
