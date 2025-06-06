package notification

import (
	"context"

	"github.com/wahyurudiyan/go-otel-context-propagation/contract/notificationpb"
	"github.com/wahyurudiyan/go-otel-context-propagation/pkg/telemetry"
	"go.uber.org/zap"
)

type grpcHandler struct {
	notificationpb.UnimplementedNotificationServiceServer
}

func NewNotificationGRPCHandler() notificationpb.NotificationServiceServer {
	return &grpcHandler{}
}

func (h *grpcHandler) SendPushNotification(ctx context.Context, req *notificationpb.PushNotificationRequest) (*notificationpb.PushNotificationResponse, error) {
	_, span := telemetry.StartSpan(ctx, "grpcHandler:SendPushNotification")
	defer span.End()

	spanCtx := span.SpanContext()
	zap.L().Info("grpc.SendPushNotification: span info",
		zap.String("span.id", spanCtx.SpanID().String()),
		zap.String("trace.id", spanCtx.TraceID().String()),
	)

	return &notificationpb.PushNotificationResponse{
		Success: true,
		Message: "sent",
	}, nil
}
