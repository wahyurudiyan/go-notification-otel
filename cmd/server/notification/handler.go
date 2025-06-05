package notification

import (
	"context"

	"github.com/wahyurudiyan/go-notification-otel/contract/notificationpb"
	"github.com/wahyurudiyan/go-notification-otel/telemetry"
	"go.uber.org/zap"
)

type handler struct {
	logger *zap.Logger
	notificationpb.UnimplementedNotificationServiceServer
}

func NewNotificationHandler() notificationpb.NotificationServiceServer {
	return &handler{
		logger: zap.L(),
	}
}

func (h *handler) SendPushNotification(ctx context.Context, req *notificationpb.PushNotificationRequest) (*notificationpb.PushNotificationResponse, error) {
	_, span := telemetry.StartSpan(ctx, "handler:SendPushNotification")
	defer span.End()

	spanCtx := span.SpanContext()
	h.logger.Info("grpc.SendPushNotification: span info",
		zap.String("span.id", spanCtx.SpanID().String()),
		zap.String("trace.id", spanCtx.TraceID().String()),
	)

	return &notificationpb.PushNotificationResponse{
		Success: true,
		Message: "sent",
	}, nil
}
