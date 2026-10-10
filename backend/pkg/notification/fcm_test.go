package notification

import (
	"context"
	"encoding/base64"
	"io"
	"log/slog"
	"testing"
)

func TestNewFCMPushSender_Base64Handling(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("Empty credentials returns nil client gracefully", func(t *testing.T) {
		sender := NewFCMPushSender(ctx, "", "", "", "", logger)
		if sender.client != nil {
			t.Errorf("expected client to be nil when no credentials provided")
		}
	})

	t.Run("Invalid base64 string does not panic", func(t *testing.T) {
		sender := NewFCMPushSender(ctx, "", "", "", "invalid_base64_!@#$", logger)
		if sender.client != nil {
			t.Errorf("expected client to be nil on invalid credentials")
		}
	})

	t.Run("Base64 encoded JSON string is decoded and attempted", func(t *testing.T) {
		dummyJSON := `{"type": "service_account", "project_id": "test-project"}`
		encoded := base64.StdEncoding.EncodeToString([]byte(dummyJSON))

		// Firebase Init will fail on fake dummy keys, but it should decode without panicking
		sender := NewFCMPushSender(ctx, "test-project", "", "", encoded, logger)
		if sender == nil {
			t.Fatalf("expected sender instance to not be nil")
		}
	})
}
