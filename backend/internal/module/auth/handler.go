package authmodule

import (
	"strings"
	"time"

	"golang/pkg/constants"
	"golang/pkg/errors"
	"golang/pkg/logging"
	"golang/pkg/response"
	"golang/pkg/validation"

	"github.com/gofiber/fiber/v2"
)

// AuthHandler exposes the auth endpoints.
type AuthHandler struct {
	service *AuthService
}

// NewAuthHandler creates the auth HTTP handler.
func NewAuthHandler(service *AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// OAuthLogin handles POST /api/v1/auth/oauth (Google & Apple sign-in).
func (h *AuthHandler) OAuthLogin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req OAuthLoginRequest
		if err := validation.Default.BindAndValidate(c, &req); err != nil {
			return err
		}

		res, err := h.service.OAuthLogin(&req)
		if err != nil {
			return err
		}

		setAuthCookies(c, res.Tokens)

		logging.LoggerFromFiber(c).Info("oauth user logged in", "user_id", res.User.ID, "provider", req.Provider)
		return response.SuccessResponse(c, res, "OAuth login successful")
	}
}

// AppleWebhook handles POST /api/v1/auth/apple/webhook (Apple Server-to-Server notification events).
func (h *AuthHandler) AppleWebhook() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req AppleWebhookRequest
		// Apple can send application/json or application/x-www-form-urlencoded
		if err := c.BodyParser(&req); err != nil {
			return errors.BadRequestWithCode("Failed to parse Apple webhook payload", constants.CodeBadRequest)
		}

		if req.Payload == "" {
			req.Payload = c.FormValue("payload")
			if req.Payload == "" {
				req.Payload = string(c.Body())
			}
		}

		if err := h.service.ProcessAppleWebhook(req.Payload); err != nil {
			return err
		}

		logging.LoggerFromFiber(c).Info("apple server-to-server webhook processed")
		return response.SuccessMessageResponse(c, "Apple event processed successfully")
	}
}

// DevLogin handles POST /api/v1/auth/dev-login.
func (h *AuthHandler) DevLogin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req DevLoginRequest
		if err := validation.Default.BindAndValidate(c, &req); err != nil {
			return err
		}

		res, err := h.service.DevLogin(&req)
		if err != nil {
			return err
		}

		setAuthCookies(c, res.Tokens)

		logging.LoggerFromFiber(c).Info("dev user logged in", "user_id", res.User.ID)
		return response.SuccessResponse(c, res, "Dev login successful")
	}
}

// Register handles POST /api/v1/auth/register.
func (h *AuthHandler) Register() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req RegisterRequest
		if err := validation.Default.BindAndValidate(c, &req); err != nil {
			return err
		}

		res, err := h.service.Register(&req)
		if err != nil {
			return err
		}

		setAuthCookies(c, res.Tokens)

		logging.LoggerFromFiber(c).Info("user registered", "email", req.Email)
		return response.CreatedResponse(c, res, "User registered successfully")
	}
}

// Login handles POST /api/v1/auth/login.
func (h *AuthHandler) Login() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req LoginRequest
		if err := validation.Default.BindAndValidate(c, &req); err != nil {
			return err
		}

		res, err := h.service.Login(&req)
		if err != nil {
			return err
		}

		setAuthCookies(c, res.Tokens)

		logging.LoggerFromFiber(c).Info("user logged in", "user_id", res.User.ID)
		return response.SuccessResponse(c, res, "Login successful")
	}
}

// RefreshToken handles POST /api/v1/auth/refresh.
func (h *AuthHandler) RefreshToken() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req RefreshTokenRequest
		_ = c.BodyParser(&req)

		candidates := extractRefreshTokens(c, req.RefreshToken)
		if len(candidates) == 0 {
			if token, ok := c.Locals("refresh_token").(string); ok && token != "" {
				candidates = append(candidates, token)
			}
		}

		if len(candidates) == 0 {
			return errors.UnauthorizedWithCode("Refresh token is required", constants.CodeRefreshTokenInvalid)
		}

		var (
			tokens  *TokenPair
			lastErr error
		)
		for _, candidate := range candidates {
			tok, err := h.service.RefreshToken(candidate)
			if err == nil {
				tokens = tok
				lastErr = nil
				break
			}
			lastErr = err
		}

		if tokens == nil {
			if lastErr != nil {
				return lastErr
			}
			return errors.UnauthorizedWithCode("Invalid token", constants.CodeTokenInvalid)
		}

		setAuthCookies(c, tokens)

		return response.SuccessResponse(c, tokens, "Token refreshed successfully")
	}
}

// extractRefreshTokens retrieves candidate refresh tokens from request payload,
// headers, and all cookies matching "refresh_token". This handles cases where
// multiple cookies exist on localhost (e.g. from different sub-paths or projects).
func extractRefreshTokens(c *fiber.Ctx, reqToken string) []string {
	var candidates []string

	addCandidate := func(val string) {
		val = strings.TrimSpace(val)
		val = strings.Trim(val, "\"")
		if val == "" {
			return
		}
		for _, existing := range candidates {
			if existing == val {
				return
			}
		}
		candidates = append(candidates, val)
	}

	if reqToken != "" {
		addCandidate(reqToken)
	}

	// 1. Parse raw Cookie header to find all occurrences of refresh_token
	cookieHeader := c.Get("Cookie")
	if cookieHeader != "" {
		for _, part := range strings.Split(cookieHeader, ";") {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "refresh_token=") {
				addCandidate(strings.TrimPrefix(part, "refresh_token="))
			}
		}
	}

	// 2. Fallback to Fiber helper if raw cookie header was empty
	if len(candidates) == 0 {
		if cookieVal := c.Cookies("refresh_token"); cookieVal != "" {
			addCandidate(cookieVal)
		}
	}

	return candidates
}

// Logout handles POST /api/v1/auth/logout.
func (h *AuthHandler) Logout() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return errors.UnauthorizedWithCode("User not authenticated", constants.CodeUnauthorized)
		}

		token := extractBearer(c)
		if token != "" {
			if err := h.service.jwtService.RevokeToken(token, "access"); err != nil {
				logging.LoggerFromFiber(c).Warn("failed to revoke token on logout", logging.Err(err))
			}
		}

		var logoutReq struct {
			RefreshToken string `json:"refresh_token"`
		}
		_ = c.BodyParser(&logoutReq)

		// Delete the current active refresh token only (single-device session logout)
		candidates := extractRefreshTokens(c, logoutReq.RefreshToken)
		for _, rf := range candidates {
			h.service.RevokeCurrentToken(rf)
		}

		clearAuthCookies(c)

		return response.SuccessResponse(c, fiber.Map{"user_id": userID}, "Logout successful")
	}
}

// LogoutAll handles POST /api/v1/auth/logout-all.
func (h *AuthHandler) LogoutAll() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return errors.UnauthorizedWithCode("User not authenticated", constants.CodeUnauthorized)
		}

		_, _ = h.service.db.DB.Exec("UPDATE refresh_tokens SET is_revoked = true, updated_at = NOW() WHERE user_id = $1", userID)

		if err := h.service.jwtService.RevokeAllUserTokens(userID); err != nil {
			return err
		}

		clearAuthCookies(c)

		return response.SuccessResponse(c, fiber.Map{"user_id": userID}, "Logout successful")
	}
}

// GetProfile handles GET /api/v1/auth/profile.
func (h *AuthHandler) GetProfile() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return errors.UnauthorizedWithCode("User not authenticated", constants.CodeUnauthorized)
		}

		user, err := h.service.GetCurrentUser(userID)
		if err != nil {
			return err
		}

		return response.SuccessResponse(c, user, "Profile retrieved successfully")
	}
}

// UpdateProfile handles PUT /api/v1/auth/profile.
func (h *AuthHandler) UpdateProfile() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return errors.UnauthorizedWithCode("User not authenticated", constants.CodeUnauthorized)
		}

		var req UpdateProfileRequest
		if err := validation.Default.BindAndValidate(c, &req); err != nil {
			return err
		}

		user, err := h.service.UpdateProfile(userID, &req)
		if err != nil {
			return err
		}

		return response.SuccessResponse(c, user, "Profile updated successfully")
	}
}

// ChangePassword handles POST /api/v1/auth/change-password.
func (h *AuthHandler) ChangePassword() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return errors.UnauthorizedWithCode("User not authenticated", constants.CodeUnauthorized)
		}

		var req ChangePasswordRequest
		if err := validation.Default.BindAndValidate(c, &req); err != nil {
			return err
		}

		if err := h.service.ChangePassword(userID, &req); err != nil {
			return err
		}

		return response.SuccessResponse(c, fiber.Map{}, "Password changed successfully. Please login again.")
	}
}

// BindOAuth handles POST /api/v1/auth/bind-oauth.
func (h *AuthHandler) BindOAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return errors.UnauthorizedWithCode("User not authenticated", constants.CodeUnauthorized)
		}

		var req BindOAuthRequest
		if err := validation.Default.BindAndValidate(c, &req); err != nil {
			return err
		}

		res, err := h.service.BindOAuth(userID, &req)
		if err != nil {
			return err
		}

		return response.SuccessResponse(c, res, "OAuth account linked successfully")
	}
}

// UnbindOAuth handles POST /api/v1/auth/unbind-oauth.
func (h *AuthHandler) UnbindOAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return errors.UnauthorizedWithCode("User not authenticated", constants.CodeUnauthorized)
		}

		var req UnbindOAuthRequest
		if err := validation.Default.BindAndValidate(c, &req); err != nil {
			return err
		}

		res, err := h.service.UnbindOAuth(userID, &req)
		if err != nil {
			return err
		}

		return response.SuccessResponse(c, res, "OAuth account unlinked successfully")
	}
}

// UploadAvatar handles POST /api/v1/auth/avatar.
func (h *AuthHandler) UploadAvatar() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return errors.UnauthorizedWithCode("User not authenticated", constants.CodeUnauthorized)
		}

		fileHeader, err := c.FormFile("avatar")
		if err != nil {
			// Fallback check field name 'file'
			fileHeader, err = c.FormFile("file")
			if err != nil {
				return errors.BadRequestWithCode("Avatar file is required (form-data: avatar)", constants.CodeAvatarFileRequired)
			}
		}

		// Max 5MB
		if fileHeader.Size > 5*1024*1024 {
			return errors.BadRequestWithCode("Avatar file size exceeds 5MB limit", constants.CodeAvatarFileTooLarge)
		}

		file, err := fileHeader.Open()
		if err != nil {
			return errors.InternalError("Failed to read avatar file", err)
		}
		defer file.Close()

		contentType := fileHeader.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "image/webp"
		}

		res, err := h.service.UploadAvatar(userID, file, fileHeader.Size, contentType)
		if err != nil {
			return err
		}

		return response.SuccessResponse(c, res, "Profile avatar updated successfully")
	}
}

// DeleteAccount handles DELETE /api/v1/auth/account.
func (h *AuthHandler) DeleteAccount() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return errors.UnauthorizedWithCode("User not authenticated", constants.CodeUnauthorized)
		}

		var req DeleteAccountRequest
		if err := validation.Default.BindAndValidate(c, &req); err != nil {
			return err
		}

		if err := h.service.DeleteAccount(userID, &req); err != nil {
			return err
		}

		clearAuthCookies(c)

		return response.SuccessMessageResponse(c, "Account deleted permanently")
	}
}

// extractBearer strips the "Bearer " prefix from the Authorization header or checks the access_token cookie.
func extractBearer(c *fiber.Ctx) string {
	header := c.Get("Authorization")
	if len(header) > 7 && header[:7] == "Bearer " {
		return header[7:]
	}
	if cookie := c.Cookies("access_token"); cookie != "" {
		return cookie
	}
	return ""
}

// setAuthCookies configures the HttpOnly refresh token cookie and the access token cookie.
func setAuthCookies(c *fiber.Ctx, tokens *TokenPair) {
	if tokens == nil {
		return
	}

	secure := c.Protocol() == "https"

	// 1. Refresh token: HttpOnly cookie (inaccessible via JavaScript)
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		Path:     "/",
		HTTPOnly: true,
		SameSite: "Lax",
		MaxAge:   7 * 24 * 3600, // 7 days
		Secure:   secure,
	})

	// 2. Access token: standard cookie ("klo access taruh di cookie aja")
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    tokens.AccessToken,
		Path:     "/",
		HTTPOnly: false,
		SameSite: "Lax",
		MaxAge:   15 * 60, // 15 mins
		Secure:   secure,
	})
}

// clearAuthCookies removes both the refresh token and access token cookies across common paths.
func clearAuthCookies(c *fiber.Ctx) {
	paths := []string{"/", "/api", "/api/v1", "/api/v1/auth"}
	for _, p := range paths {
		c.Cookie(&fiber.Cookie{
			Name:     "refresh_token",
			Value:    "",
			Path:     p,
			HTTPOnly: true,
			SameSite: "Lax",
			MaxAge:   -1,
			Expires:  time.Unix(0, 0),
		})
		c.Cookie(&fiber.Cookie{
			Name:     "access_token",
			Value:    "",
			Path:     p,
			HTTPOnly: false,
			SameSite: "Lax",
			MaxAge:   -1,
			Expires:  time.Unix(0, 0),
		})
	}
}
