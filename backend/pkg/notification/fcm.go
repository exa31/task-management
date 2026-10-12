package notification

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type PushResult struct {
	SuccessCount int      `json:"success_count"`
	FailureCount int      `json:"failure_count"`
	TotalTokens  int      `json:"total_tokens"`
	Errors       []string `json:"errors,omitempty"`
}

type PushSender interface {
	SendPushNotification(ctx context.Context, tokens []string, title, body string, data map[string]string) error
	SendPushNotificationWithResult(ctx context.Context, tokens []string, title, body string, data map[string]string) (*PushResult, error)
}

type FCMPushSender struct {
	client *messaging.Client
	logger *slog.Logger
}

func NewFCMPushSender(ctx context.Context, projectID, credsFile, credsJSON, credsBase64 string, logger *slog.Logger) *FCMPushSender {
	var opts []option.ClientOption
	var credsBytes []byte

	if credsBase64 != "" {
		decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(credsBase64))
		if err != nil {
			logger.Error("failed to decode FCM credentials base64 string", slog.String("error", err.Error()))
		} else {
			credsBytes = decoded
		}
	} else if credsJSON != "" {
		trimmed := strings.TrimSpace(credsJSON)
		// Check if credsJSON is itself a base64 encoded string
		if decoded, err := base64.StdEncoding.DecodeString(trimmed); err == nil && len(decoded) > 0 && (decoded[0] == '{' || decoded[0] == '[') {
			credsBytes = decoded
		} else {
			credsBytes = []byte(trimmed)
		}
	}

	if len(credsBytes) > 0 {
		opts = append(opts, option.WithCredentialsJSON(credsBytes))
	} else if credsFile != "" {
		opts = append(opts, option.WithCredentialsFile(credsFile))
	}

	cfg := &firebase.Config{}
	if projectID != "" {
		cfg.ProjectID = projectID
	}

	// If no credentials supplied at all, do not initialize client
	if len(opts) == 0 && projectID == "" {
		logger.Warn("fcm push sender initialized without credentials (push notifications disabled until configured)")
		return &FCMPushSender{
			client: nil,
			logger: logger,
		}
	}

	app, err := firebase.NewApp(ctx, cfg, opts...)
	if err != nil {
		logger.Warn("failed to initialize firebase app", slog.String("error", err.Error()))
		return &FCMPushSender{client: nil, logger: logger}
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		logger.Warn("failed to initialize firebase messaging client", slog.String("error", err.Error()))
		return &FCMPushSender{client: nil, logger: logger}
	}

	logger.Info("fcm push sender initialized successfully")
	return &FCMPushSender{
		client: client,
		logger: logger,
	}
}

func (s *FCMPushSender) SendPushNotification(ctx context.Context, tokens []string, title, body string, data map[string]string) error {
	_, err := s.SendPushNotificationWithResult(ctx, tokens, title, body, data)
	return err
}

func (s *FCMPushSender) SendPushNotificationWithResult(ctx context.Context, tokens []string, title, body string, data map[string]string) (*PushResult, error) {
	if s.client == nil {
		s.logger.Warn("fcm push skipped: FCM credentials not configured")
		return nil, errors.New("fcm credentials not configured")
	}

	if len(tokens) == 0 {
		s.logger.Debug("fcm push skipped: no device tokens registered for user")
		return &PushResult{SuccessCount: 0, FailureCount: 0, TotalTokens: 0}, nil
	}

	msg := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: "default",
				},
			},
		},
	}

	resp, err := s.client.SendEachForMulticast(ctx, msg)
	if err != nil {
		s.logger.Error("failed to send fcm multicast message", slog.String("error", err.Error()))
		return nil, fmt.Errorf("fcm send error: %w", err)
	}

	res := &PushResult{
		SuccessCount: resp.SuccessCount,
		FailureCount: resp.FailureCount,
		TotalTokens:  len(tokens),
	}

	if resp.FailureCount > 0 {
		for i, r := range resp.Responses {
			if !r.Success && r.Error != nil {
				errMsg := fmt.Sprintf("token[%d]: %v", i, r.Error)
				res.Errors = append(res.Errors, errMsg)
				s.logger.Warn("fcm token delivery failed",
					slog.Int("index", i),
					slog.String("error", r.Error.Error()),
				)
			}
		}
	}

	s.logger.Info("fcm push notification dispatched",
		slog.Int("success_count", resp.SuccessCount),
		slog.Int("failure_count", resp.FailureCount),
		slog.Int("total_tokens", len(tokens)),
	)

	return res, nil
}
