package logging

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	customErrors "golang/pkg/errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

// ContextKey (shared with the middleware).
const HeaderRequestID = "X-Request-ID"

// RequestIDMiddleware assigns a request ID to every request (honoring an
// optional X-Request-ID header) and stores a request-scoped logger in the
// context carrying it. Handlers can retrieve the logger with LoggerFromFiber
// or the ID with RequestIDFromFiber.
func RequestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		rid := c.Get(HeaderRequestID)
		if rid == "" {
			rid = newRequestID()
		}
		c.Set(HeaderRequestID, rid)
		c.Locals("request_id", rid)

		ctx := ContextWithRequestID(c.UserContext(), rid)

		logAttrs := []any{slog.String("request_id", rid)}
		if spanCtx := trace.SpanContextFromContext(c.UserContext()); spanCtx.IsValid() {
			logAttrs = append(logAttrs, slog.String("trace_id", spanCtx.TraceID().String()), slog.String("span_id", spanCtx.SpanID().String()))
		}

		ctx = ContextWithLogger(ctx, Logger().With(logAttrs...))
		c.SetUserContext(ctx)

		return c.Next()
	}
}

// newRequestID generates a unique request identifier.
func newRequestID() string {
	return uuid.NewString()
}

// RequestIDFromFiber returns the request ID for the current request.
func RequestIDFromFiber(c *fiber.Ctx) string {
	if v, ok := c.Locals("request_id").(string); ok {
		return v
	}
	return ""
}

// LoggerFromFiber returns a logger that already carries the request ID for
// the current request. Use it inside handlers instead of the global logger.
func LoggerFromFiber(c *fiber.Ctx) *slog.Logger {
	return LoggerFromContext(c.UserContext())
}

// ContextFromFiber returns the request context (useful for slog attrs).
func ContextFromFiber(c *fiber.Ctx) context.Context {
	return c.UserContext()
}

// AccessLogMiddleware logs every HTTP request with structured fields,
// including the request ID, path, status, and latency.
// Status >= 500 is logged at Error level, 4xx at Warn level, and 2xx/3xx at Info level.
func AccessLogMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		status := effectiveStatus(c, err)
		logger := LoggerFromFiber(c)
		attrs := []any{
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", status),
			slog.Duration("latency", time.Since(start)),
			slog.String("ip", c.IP()),
			slog.String("user_agent", string(c.Context().UserAgent())),
			Err(err),
		}

		switch {
		case status >= 500:
			logger.Error("http request", attrs...)
		case status >= 400:
			logger.Warn("http request", attrs...)
		default:
			logger.Info("http request", attrs...)
		}

		return err
	}
}

// effectiveStatus reports the final HTTP status for the request. When a
// handler returns an error, Fiber has not written the response status yet at
// this point, so the intended status is derived from the error type instead of
// reporting a misleading 200.
func effectiveStatus(c *fiber.Ctx, err error) int {
	if err == nil {
		return c.Response().StatusCode()
	}
	switch e := err.(type) {
	case *customErrors.AppError:
		return e.StatusCode
	case *fiber.Error:
		return e.Code
	case validator.ValidationErrors:
		return fiber.StatusBadRequest
	}
	if errors.Is(err, sql.ErrNoRows) {
		return fiber.StatusNotFound
	}
	return fiber.StatusInternalServerError
}
