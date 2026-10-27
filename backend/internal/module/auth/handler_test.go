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

	t.Run("extracts with X-Refresh-Token header and URL-encoded cookie", func(t *testing.T) {
		app.Post("/test-header-urlencode", func(c *fiber.Ctx) error {
			tokens := extractRefreshTokens(c, "")
			if len(tokens) < 2 {
				t.Fatalf("expected at least 2 tokens, got %d: %v", len(tokens), tokens)
			}
			if tokens[0] != "header-token-123" {
				t.Errorf("expected first token from header, got %s", tokens[0])
			}
			foundUnescaped := false
			for _, tok := range tokens {
				if tok == "token value with spaces" {
					foundUnescaped = true
					break
				}
			}
			if !foundUnescaped {
				t.Errorf("expected URL-unescaped token in candidates: %v", tokens)
			}
			return c.SendStatus(fiber.StatusOK)
		})

		req := httptest.NewRequest("POST", "/test-header-urlencode", nil)
		req.Header.Set("X-Refresh-Token", "header-token-123")
		req.Header.Set("Cookie", "refresh_token=token%20value%20with%20spaces")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})
}

func TestSetAuthCookies_ProxySecure(t *testing.T) {
	app := fiber.New()

	app.Get("/set-cookie-test", func(c *fiber.Ctx) error {
		setAuthCookies(c, &TokenPair{
			AccessToken:  "test-acc",
			RefreshToken: "test-ref",
		})
		return c.SendStatus(fiber.StatusOK)
	})

	t.Run("sets Secure=true when behind HTTPS reverse proxy", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/set-cookie-test", nil)
		req.Header.Set("X-Forwarded-Proto", "https")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test failed: %v", err)
		}

		cookies := resp.Cookies()
		var foundRef, foundAcc bool
		for _, ck := range cookies {
			if ck.Name == "refresh_token" {
				foundRef = true
				if !ck.Secure {
					t.Errorf("expected refresh_token to have Secure=true behind HTTPS proxy")
				}
				if !ck.HttpOnly {
					t.Errorf("expected refresh_token to be HttpOnly")
				}
			}
			if ck.Name == "access_token" {
				foundAcc = true
				if !ck.Secure {
					t.Errorf("expected access_token to have Secure=true behind HTTPS proxy")
				}
			}
		}
		if !foundRef || !foundAcc {
			t.Errorf("missing expected auth cookies in response: %v", cookies)
		}
	})
}

func TestVerifyAppleAuth(t *testing.T) {
	svc := &AuthService{}

	t.Run("verifies apple with direct email and name", func(t *testing.T) {
		req := &OAuthLoginRequest{
			Provider: "apple",
			Email:    "test@privaterelay.appleid.com",
			Name:     "Apple Tester",
			Code:     "auth-code-123",
		}
		info, err := svc.verifyAppleAuth(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.Email != req.Email {
			t.Errorf("expected email %s, got %s", req.Email, info.Email)
		}
		if info.Name != req.Name {
			t.Errorf("expected name %s, got %s", req.Name, info.Name)
		}
		if info.Sub != req.Code {
			t.Errorf("expected sub %s, got %s", req.Code, info.Sub)
		}
	})

	t.Run("fails when email and id_token are missing", func(t *testing.T) {
		req := &OAuthLoginRequest{
			Provider: "apple",
		}
		_, err := svc.verifyAppleAuth(req)
		if err == nil {
			t.Fatal("expected error for missing email, got nil")
		}
	})
}
