package auth

import (
	"strings"

	"golang/pkg/auth"
	"golang/pkg/constants"
	customErrors "golang/pkg/errors"

	"github.com/gofiber/fiber/v2"
)

// bearerScheme is the Authorization header scheme used everywhere in the app.
const bearerScheme = "Bearer"

// AuthMiddleware creates a JWT authentication middleware
func AuthMiddleware(jwtService *auth.JWTService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var tokenString string

		// 1. Check Authorization header
		authHeader := c.Get("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == bearerScheme {
				tokenString = parts[1]
			} else {
				return customErrors.UnauthorizedWithCode("Invalid authorization header format", constants.CodeUnauthorized)
			}
		}

		// 2. Fallback to access_token cookie
		if tokenString == "" {
			tokenString = c.Cookies("access_token")
		}

		if tokenString == "" {
			return customErrors.UnauthorizedWithCode("Authorization token is required", constants.CodeUnauthorized)
		}

		// Validate access token
		claims, err := jwtService.ValidateToken(tokenString, auth.AccessToken)
		if err != nil {
			return err
		}

		// Check if user tokens are revoked
		if jwtService.IsUserRevoked(claims.UserID) {
			return customErrors.UnauthorizedWithCode("Token has been revoked", constants.CodeTokenInvalid)
		}

		// Store claims in context
		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("token_id", claims.TokenID)
		c.Locals("claims", claims)

		return c.Next()
	}
}

// OptionalAuthMiddleware is similar to AuthMiddleware but doesn't fail if token is missing
// Useful for endpoints that work differently for authenticated users
func OptionalAuthMiddleware(jwtService *auth.JWTService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var tokenString string

		authHeader := c.Get("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == bearerScheme {
				tokenString = parts[1]
			}
		}

		if tokenString == "" {
			tokenString = c.Cookies("access_token")
		}

		if tokenString == "" {
			return c.Next()
		}

		claims, err := jwtService.ValidateToken(tokenString, auth.AccessToken)
		if err != nil {
			return c.Next()
		}

		if !jwtService.IsUserRevoked(claims.UserID) {
			c.Locals("user_id", claims.UserID)
			c.Locals("email", claims.Email)
			c.Locals("token_id", claims.TokenID)
			c.Locals("claims", claims)
		}

		return c.Next()
	}
}

// RefreshTokenMiddleware validates refresh token from header or HttpOnly cookie
func RefreshTokenMiddleware(jwtService *auth.JWTService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var tokenString string

		authHeader := c.Get("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == bearerScheme {
				tokenString = parts[1]
			} else {
				return customErrors.UnauthorizedWithCode("Invalid authorization header format", constants.CodeUnauthorized)
			}
		}

		if tokenString == "" {
			tokenString = c.Cookies("refresh_token")
		}

		if tokenString == "" {
			return customErrors.UnauthorizedWithCode("Refresh token is required", constants.CodeRefreshTokenInvalid)
		}

		// Validate refresh token
		claims, err := jwtService.ValidateToken(tokenString, auth.RefreshToken)
		if err != nil {
			return err
		}

		if jwtService.IsUserRevoked(claims.UserID) {
			return customErrors.UnauthorizedWithCode("Token has been revoked", constants.CodeRefreshTokenRevoked)
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("token_id", claims.TokenID)
		c.Locals("claims", claims)
		c.Locals("refresh_token", tokenString)

		return c.Next()
	}
}

// GetUserID extracts user ID from context
func GetUserID(c *fiber.Ctx) (string, error) {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return "", customErrors.UnauthorizedWithCode("User not authenticated", constants.CodeUnauthorized)
	}
	return userID, nil
}

// GetEmail extracts email from context
func GetEmail(c *fiber.Ctx) (string, error) {
	email, ok := c.Locals("email").(string)
	if !ok || email == "" {
		return "", customErrors.Unauthorized("User not authenticated")
	}
	return email, nil
}

// GetClaims extracts JWT claims from context
func GetClaims(c *fiber.Ctx) (*auth.JWTClaims, error) {
	claims, ok := c.Locals("claims").(*auth.JWTClaims)
	if !ok {
		return nil, customErrors.Unauthorized("User not authenticated")
	}
	return claims, nil
}

// MustGetUserID extracts user ID or panics (use with recover middleware)
func MustGetUserID(c *fiber.Ctx) string {
	userID, err := GetUserID(c)
	if err != nil {
		panic(err)
	}
	return userID
}
