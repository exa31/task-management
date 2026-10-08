package authmodule

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestExtractRefreshTokens(t *testing.T) {
	app := fiber.New()

	t.Run("extracts from body and multiple cookies without duplicates", func(t *testing.T) {
		app.Post("/test", func(c *fiber.Ctx) error {
			tokens := extractRefreshTokens(c, "body-token")
			if len(tokens) != 3 {
				t.Fatalf("expected 3 tokens, got %d: %v", len(tokens), tokens)
			}
			if tokens[0] != "body-token" {
				t.Errorf("expected first token to be body-token, got %s", tokens[0])
			}
			if tokens[1] != "cookie-token-1" {
				t.Errorf("expected second token to be cookie-token-1, got %s", tokens[1])
			}
			if tokens[2] != "cookie-token-2" {
				t.Errorf("expected third token to be cookie-token-2, got %s", tokens[2])
			}
			return c.SendStatus(fiber.StatusOK)
		})

		req := httptest.NewRequest("POST", "/test", nil)
		req.Header.Set("Cookie", "refresh_token=cookie-token-1; other=abc; refresh_token=cookie-token-2")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("extracts when only cookie header exists", func(t *testing.T) {
		app.Post("/test-cookies-only", func(c *fiber.Ctx) error {
			tokens := extractRefreshTokens(c, "")
			if len(tokens) != 2 {
				t.Fatalf("expected 2 tokens, got %d: %v", len(tokens), tokens)
			}
			if tokens[0] != "tokenA" || tokens[1] != "tokenB" {
				t.Errorf("unexpected tokens: %v", tokens)
			}
			return c.SendStatus(fiber.StatusOK)
		})

		req := httptest.NewRequest("POST", "/test-cookies-only", nil)
		req.Header.Set("Cookie", "refresh_token=tokenA; refresh_token=tokenB")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})
}
