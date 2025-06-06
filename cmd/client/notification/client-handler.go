package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/wahyurudiyan/go-otel-context-propagation/contract/notificationpb"
	"github.com/wahyurudiyan/go-otel-context-propagation/pkg/telemetry"
	"go.uber.org/zap"
)

type handler struct {
	httpClient      *http.Client
	notificationCli notificationpb.NotificationServiceClient
}

type Handler interface {
	SendPushNotification(ctx context.Context, data PushNotificationRequest) error
	SendEmailNotification(ctx context.Context, data EmailNotificationRequest) ([]byte, error)
}

func NewNotificationHandler(
	httpClient *http.Client,
	notificationCli notificationpb.NotificationServiceClient,
) Handler {
	return &handler{
		httpClient:      httpClient,
		notificationCli: notificationCli,
	}
}

func (h *handler) SendPushNotification(ctx context.Context, data PushNotificationRequest) error {
	ctx, span := telemetry.StartSpan(ctx, "handler:SendPushNotification")
	defer span.End()

	spanCtx := span.SpanContext()
	zap.L().Info("http.SendPushNotification: span info",
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
		zap.L().Error("failed to call rpc SendPushNotification", zap.Error(err))
		return err
	}

	zap.L().Debug("RPC payload response", zap.Any("grpc.response", rpcRes))

	return nil
}

func (h *handler) SendEmailNotification(ctx context.Context, data EmailNotificationRequest) ([]byte, error) {
	ctx, span := telemetry.StartSpan(ctx, "handler:SendEmailNotification")
	defer span.End()

	spanCtx := span.SpanContext()
	zap.L().Info("http.SendPushNotification: span info",
		zap.String("span.id", spanCtx.SpanID().String()),
		zap.String("trace.id", spanCtx.TraceID().String()),
	)

	requestBody, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	notificationServiceURL := "http://localhost:8080/server/notifications/email"
	httpRequest, err := http.NewRequestWithContext(
		ctx,
		"POST", notificationServiceURL,
		bytes.NewBuffer(requestBody),
	)
	if err != nil {
		return nil, err
	}

	httpRequest.Header.Add("Content-Type", "application/json")
	resp, err := h.httpClient.Do(httpRequest)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("http call error occur: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	zap.L().Debug("HTTP payload response", zap.ByteString("grpc.response", requestBody))

	return body, nil
}
