package notification

import (
	"context"
	"strconv"

	"github.com/wahyurudiyan/go-otel-context-propagation/contract/notificationpb"
	"github.com/wahyurudiyan/go-otel-context-propagation/telemetry"
	"go.uber.org/zap"
)

type handler struct {
	logger          *zap.Logger
	notificationCli notificationpb.NotificationServiceClient
}

type Handler interface {
	SendPushNotification(ctx context.Context, data PushNotificationRequest) error
}

func NewNotificationHandler(
	notificationCli notificationpb.NotificationServiceClient,
) Handler {
	return &handler{
		logger:          zap.L(),
		notificationCli: notificationCli,
	}
}

func (h *handler) SendPushNotification(ctx context.Context, data PushNotificationRequest) error {
	ctx, span := telemetry.StartSpan(ctx, "handler:SendPushNotification")
	defer span.End()

	spanCtx := span.SpanContext()
	h.logger.Info("http.SendPushNotification: span info",
		zap.String("span.id", spanCtx.SpanID().String()),
		zap.String("trace.id", spanCtx.TraceID().String()),
	)

	rpcReq := &notificationpb.PushNotificationRequest{
		Data:   data.Data,
		Body:   data.Body,
		Title:  data.Title,
		UserId: strconv.Itoa(int(data.UserId)),
	}
	rpcRes, err := h.notificationCli.SendPushNotification(ctx, rpcReq)
	if err != nil {
		h.logger.Error("failed to call rpc SendPushNotification", zap.Error(err))
		return err
	}

	h.logger.Debug("RPC payload response", zap.Any("grpc.response", rpcRes))

	return nil
}
