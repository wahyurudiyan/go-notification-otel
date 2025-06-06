package notification

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wahyurudiyan/go-otel-context-propagation/pkg/telemetry"
	"go.uber.org/zap"
)

type httpHandler struct{}

type HTTPHandler interface {
	SendEmailNotification() fiber.Handler
}

func NewNotificationHTTPHandler() HTTPHandler {
	return &httpHandler{}
}

func (h *httpHandler) SendEmailNotification() fiber.Handler {
	return func(fiberCtx *fiber.Ctx) error {
		_, span := telemetry.StartSpan(fiberCtx.UserContext(), "httpHandler:SendEmailNotification")
		defer span.End()
		spanCtx := span.SpanContext()

		zap.L().Debug("body request", zap.ByteString("payload", fiberCtx.Body()))

		var req EmailNotificationRequest
		if err := fiberCtx.BodyParser(&req); err != nil {
			zap.L().Error("http.SendEmailNotification: error occur",
				zap.Error(err),
				zap.String("span.id", spanCtx.SpanID().String()),
				zap.String("trace.id", spanCtx.TraceID().String()),
			)
			return err
		}

		zap.L().Info("http.SendEmailNotification: span info",
			zap.String("span.id", spanCtx.SpanID().String()),
			zap.String("trace.id", spanCtx.TraceID().String()),
			zap.String("email.to", req.Email),
		)

		return fiberCtx.JSON(map[string]interface{}{
			"success":  true,
			"message":  "email sent",
			"payload":  req,
			"trace_id": spanCtx.TraceID().String(),
		})
	}
}
