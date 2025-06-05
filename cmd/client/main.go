package main

import (
	"context"
	"errors"
	"os/signal"
	"syscall"

	"github.com/gofiber/contrib/otelfiber/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/wahyurudiyan/go-otel-context-propagation/cmd/client/notification"
	"github.com/wahyurudiyan/go-otel-context-propagation/contract/notificationpb"
	"github.com/wahyurudiyan/go-otel-context-propagation/pkg/graceful"
	"github.com/wahyurudiyan/go-otel-context-propagation/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	NotificationGRPCHost = "127.0.0.1:9090"
)

func init() {
	zapConfig := zap.NewDevelopmentConfig()
	zap.ReplaceGlobals(zap.Must(zapConfig.Build()))
}

func main() {
	// initialize graceful shutdown
	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGKILL, syscall.SIGQUIT)
	defer stop()

	shutdown, err := telemetry.SetupOTelSDK(ctx, "http.server")
	if err != nil {
		zap.L().Fatal("unable to setup OTelSDK", zap.Error(err))
	}

	defer func() {
		err = errors.Join(err, shutdown(context.Background()))
	}()

	if err := graceful.Runner(ctx, bootstrap); err != nil {
		zap.L().Fatal("Server cannot shutdown gracefully", zap.Error(err))
	}
}

func bootstrap(ctx context.Context) graceful.ShutdownCallback {
	httpServer := startHTTPServer()
	return func(ctx context.Context) error {
		if err := httpServer.ShutdownWithContext(ctx); err != nil {
			return err
		}
		return nil
	}
}

func startHTTPServer() *fiber.App {
	mux := fiber.New()
	mux.Use(otelfiber.Middleware(otelfiber.WithCustomAttributes(
		func(ctx *fiber.Ctx) []attribute.KeyValue {
			r := resource.NewSchemaless(
				semconv.ServiceName("notification:server"),
			)
			return r.Attributes()
		},
	)))
	conn, err := grpc.NewClient(
		NotificationGRPCHost,
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		zap.L().Fatal("Cannot listen to gRPC server", zap.String("server.host", NotificationGRPCHost), zap.Error(err))
	}
	grpcNotificationClient := notificationpb.NewNotificationServiceClient(conn)

	notificationHandler := notification.NewNotificationHandler(grpcNotificationClient)
	mux.Post("/notifications/push", func(c *fiber.Ctx) error {
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

	// run http server
	go func() {
		if err := mux.Listen(":8081"); err != nil {
			zap.L().Fatal("Cannot run http server", zap.Error(err))
		}
	}()

	return mux
}
