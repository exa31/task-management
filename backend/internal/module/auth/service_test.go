package authmodule

import (
	"golang/config"
	"testing"
)

func TestAppleAllowedAudiences(t *testing.T) {
	t.Run("matches when both AppleClientID and AppleBundleID are set", func(t *testing.T) {
		service := &AuthService{
			cfg: &config.Config{
				OAuth: config.OAuthConfig{
					AppleClientID: "cloud.eka-dev.task-hub.service",
					AppleBundleID: "cloud.eka-dev.task-hub",
				},
			},
		}

		if !service.isAllowedAppleAudience("cloud.eka-dev.task-hub.service") {
			t.Errorf("expected web audience to be allowed")
		}

		if !service.isAllowedAppleAudience("cloud.eka-dev.task-hub") {
			t.Errorf("expected mobile bundle ID to be allowed")
		}

		if service.isAllowedAppleAudience("com.other.app") {
			t.Errorf("expected unrelated audience to be rejected")
		}
	})

	t.Run("matches comma-separated AppleClientID", func(t *testing.T) {
		service := &AuthService{
			cfg: &config.Config{
				OAuth: config.OAuthConfig{
					AppleClientID: "cloud.eka-dev.task-hub.service, cloud.eka-dev.task-hub",
				},
			},
		}

		if !service.isAllowedAppleAudience("cloud.eka-dev.task-hub.service") {
			t.Errorf("expected first audience to be allowed")
		}

		if !service.isAllowedAppleAudience("cloud.eka-dev.task-hub") {
			t.Errorf("expected second audience to be allowed")
		}

		if service.isAllowedAppleAudience("random.client") {
			t.Errorf("expected random audience to be rejected")
		}
	})

	t.Run("allows all when no apple client ID or bundle ID is configured", func(t *testing.T) {
		service := &AuthService{
			cfg: &config.Config{},
		}

		if !service.isAllowedAppleAudience("any.audience") {
			t.Errorf("expected any audience to be allowed if unconfigured")
		}
	})
}

func TestHashToken(t *testing.T) {
	tok1 := "test-token-value-1"
	tok2 := "test-token-value-2"

	h1 := hashToken(tok1)
	h2 := hashToken(tok2)

	if h1 == "" || h2 == "" {
		t.Fatal("hashToken should produce non-empty hashes")
	}
	if h1 == h2 {
		t.Fatal("different tokens should produce different hashes")
	}
	if hashToken(tok1) != h1 {
		t.Fatal("hashToken should be deterministic")
	}
}
