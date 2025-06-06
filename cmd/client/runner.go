package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofiber/contrib/otelfiber/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/wahyurudiyan/go-otel-context-propagation/cmd/client/notification"
	"github.com/wahyurudiyan/go-otel-context-propagation/contract/notificationpb"
	"github.com/wahyurudiyan/go-otel-context-propagation/pkg/graceful"
	"github.com/wahyurudiyan/go-otel-context-propagation/pkg/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func boostrap(ctx context.Context) graceful.ShutdownCallback {
	httpServer := startHTTPServer()
	return func(ctx context.Context) error {
		if err := httpServer.ShutdownWithContext(ctx); err != nil {
			return err
		}
		return nil
	}
}

func startHTTPServer() *fiber.App {
	// Init HTTP Server
	mux := fiber.New()
	mux.Use(otelfiber.Middleware(otelfiber.WithCustomAttributes(
		func(ctx *fiber.Ctx) []attribute.KeyValue {
			return []attribute.KeyValue{
				semconv.ServiceName("notification:client"),
			}
		},
	)))
	router := mux.Group("/client")

	// Init HTTP and GRPC Client
	httpClient := newHTTPClient(time.Duration(0))
	conn, err := grpc.NewClient(
		NotificationGRPCHost,
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		zap.L().Fatal("Cannot listen to gRPC server", zap.String("server.host", NotificationGRPCHost), zap.Error(err))
	}
	grpcNotificationClient := notificationpb.NewNotificationServiceClient(conn)

	// Init notification logic
	notificationHandler := notification.NewNotificationHandler(httpClient, grpcNotificationClient)

	// Init Controller
	router.Post("/notifications/push", func(c *fiber.Ctx) error {
		ctx, span := telemetry.StartSpan(c.UserContext(), "controller:PushNotification")
		defer span.End()

		var req notification.PushNotificationRequest
		if err := c.BodyParser(&req); err != nil {
			zap.L().Error("Cannot unmarshal body", zap.ByteString("body", c.BodyRaw()), zap.Error(err))
			return err
		}

		if err := notificationHandler.SendPushNotification(ctx, req); err != nil {
			zap.L().Error("Unable to send notification", zap.Error(err))
			return err
		}

		return nil
	})

	router.Post("/notifications/email", func(c *fiber.Ctx) error {
		ctx, span := telemetry.StartSpan(c.UserContext(), "controller:EmailNotification")
		defer span.End()

		var req notification.EmailNotificationRequest
		if err := c.BodyParser(&req); err != nil {
			zap.L().Error("Cannot unmarshal body", zap.ByteString("body", c.BodyRaw()), zap.Error(err))
			return err
		}

		respBody, err := notificationHandler.SendEmailNotification(ctx, req)
		if err != nil {
			zap.L().Error("Unable to send notification", zap.Error(err))
			return err
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(respBody, &resp); err != nil {
			return err
		}

		return c.JSON(resp, "application/json")
	})

	// Run http server
	go func() {
		if err := mux.Listen(":8081"); err != nil {
			zap.L().Fatal("Cannot run http server", zap.Error(err))
		}
	}()

	return mux
}
