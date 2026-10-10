package reminder

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

type mockService struct {
	Service
	processReminderFunc func(ctx context.Context, id int64) error
}

func (m *mockService) ProcessReminder(ctx context.Context, id int64) error {
	if m.processReminderFunc != nil {
		return m.processReminderFunc(ctx, id)
	}
	return nil
}

func TestHandler_Trigger_MandatorySecret(t *testing.T) {
	secret := "test-secret-key-123"

	tests := []struct {
		name           string
		headers        map[string]string
		body           any
		mockProcessErr error
		expectedStatus int
	}{
		{
			name: "rejected when no secret header is provided",
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			body:           map[string]any{"reminder_id": 10},
			expectedStatus: fiber.StatusUnauthorized,
		},
		{
			name: "rejected when invalid secret is provided",
			headers: map[string]string{
				"Content-Type":     "application/json",
				"X-Webhook-Secret": "wrong-secret",
			},
			body:           map[string]any{"reminder_id": 10},
			expectedStatus: fiber.StatusUnauthorized,
		},
		{
			name: "success with valid X-Webhook-Secret header",
			headers: map[string]string{
				"Content-Type":     "application/json",
				"X-Webhook-Secret": secret,
			},
			body:           map[string]any{"reminder_id": 10},
			expectedStatus: fiber.StatusOK,
		},
		{
			name: "success with valid Authorization Bearer header",
			headers: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer " + secret,
			},
			body:           map[string]any{"id": 10},
			expectedStatus: fiber.StatusOK,
		},
		{
			name: "rejected when secret is valid but body has no reminder id",
			headers: map[string]string{
				"Content-Type":     "application/json",
				"X-Webhook-Secret": secret,
			},
			body:           map[string]any{},
			expectedStatus: fiber.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			mockSvc := &mockService{
				processReminderFunc: func(ctx context.Context, id int64) error {
					return tc.mockProcessErr
				},
			}
			h := NewHandler(mockSvc, secret)
			app.Post("/trigger", h.Trigger)

			bodyBytes, _ := json.Marshal(tc.body)
			req := httptest.NewRequest(http.MethodPost, "/trigger", bytes.NewReader(bodyBytes))
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}
