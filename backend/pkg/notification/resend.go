package notification

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/resend/resend-go/v2"
)

type EmailSender interface {
	SendReminderEmail(ctx context.Context, toEmail, title, description string, taskTitle *string) error
}

type ResendEmailSender struct {
	client    *resend.Client
	fromEmail string
	logger    *slog.Logger
}

func NewResendEmailSender(apiKey, fromEmail string, logger *slog.Logger) *ResendEmailSender {
	if fromEmail == "" {
		fromEmail = "onboarding@resend.dev"
	}
	var client *resend.Client
	if apiKey != "" {
		client = resend.NewClient(apiKey)
	}
	return &ResendEmailSender{
		client:    client,
		fromEmail: fromEmail,
		logger:    logger,
	}
}

func (s *ResendEmailSender) SendReminderEmail(ctx context.Context, toEmail, title, description string, taskTitle *string) error {
	if s.client == nil {
		s.logger.Warn("resend email skipped: RESEND_API_KEY is not configured", slog.String("to", toEmail))
		return errors.New("resend api key is not configured")
	}

	subject := fmt.Sprintf("⏰ Reminder: %s", title)

	taskInfoHtml := ""
	if taskTitle != nil && *taskTitle != "" {
		taskInfoHtml = fmt.Sprintf(`
			<div style="background-color: #f1f5f9; padding: 12px 16px; border-radius: 8px; margin: 16px 0; border-left: 4px solid #6366f1;">
				<p style="margin: 0; font-size: 13px; color: #64748b; font-weight: 600;">LINKED TASK</p>
				<p style="margin: 4px 0 0 0; font-size: 15px; color: #1e293b; font-weight: 500;">%s</p>
			</div>
		`, *taskTitle)
	}

	htmlBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head><meta charset="utf-8"></head>
		<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f8fafc; padding: 24px; margin: 0;">
			<div style="max-width: 540px; margin: 0 auto; background: #ffffff; border-radius: 12px; border: 1px solid #e2e8f0; padding: 32px; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.05);">
				<div style="display: flex; align-items: center; margin-bottom: 20px;">
					<span style="font-size: 24px; margin-right: 8px;">⏰</span>
					<h2 style="margin: 0; font-size: 20px; font-weight: 700; color: #0f172a;">TaskHub Reminder</h2>
				</div>
				<h3 style="margin: 0 0 8px 0; font-size: 17px; color: #334155;">%s</h3>
				<p style="margin: 0; font-size: 14px; color: #475569; line-height: 1.5;">%s</p>
				%s
				<hr style="border: none; border-top: 1px solid #f1f5f9; margin: 24px 0;" />
				<p style="margin: 0; font-size: 12px; color: #94a3b8; text-align: center;">You received this notification because a reminder was scheduled in TaskHub.</p>
			</div>
		</body>
		</html>
	`, title, description, taskInfoHtml)

	params := &resend.SendEmailRequest{
		From:    s.fromEmail,
		To:      []string{toEmail},
		Subject: subject,
		Html:    htmlBody,
	}

	res, err := s.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		s.logger.Error("failed to send reminder email via resend", slog.String("to", toEmail), slog.String("error", err.Error()))
		return fmt.Errorf("resend send error: %w", err)
	}

	s.logger.Info("reminder email sent successfully", slog.String("id", res.Id), slog.String("to", toEmail))
	return nil
}
