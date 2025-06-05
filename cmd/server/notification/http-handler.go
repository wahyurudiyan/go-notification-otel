package notification

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wahyurudiyan/go-otel-context-propagation/contract/notificationpb"
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
		ctx := fiberCtx.UserContext()
		_, span := telemetry.StartSpan(ctx, "httpHandler:SendEmailNotification")
		defer span.End()

		spanCtx := span.SpanContext()
		zap.L().Info("http.SendEmailNotification: span info",
			zap.String("span.id", spanCtx.SpanID().String()),
			zap.String("trace.id", spanCtx.TraceID().String()),
		)

		return fiberCtx.JSON(&notificationpb.PushNotificationResponse{
			Success: true,
			Message: "sent",
		})
	}
}
