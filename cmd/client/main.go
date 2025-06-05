package main

import (
	"context"
	"errors"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/contrib/otelfiber/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/wahyurudiyan/go-notification-otel/cmd/client/notification"
	"github.com/wahyurudiyan/go-notification-otel/contract/notificationpb"
	"github.com/wahyurudiyan/go-notification-otel/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
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
	var logger = zap.L()

	// initialize graceful shutdown
	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGKILL, syscall.SIGQUIT)
	defer stop()

	shutdown, err := telemetry.SetupOTelSDK(ctx, "http.server")
	if err != nil {
		logger.Fatal("unable to setup OTelSDK", zap.Error(err))
	}

	defer func() {
		err = errors.Join(err, shutdown(context.Background()))
	}()

	httpMux := fiber.New()
	httpMux.Use(otelfiber.Middleware())

	conn, err := grpc.NewClient(
		NotificationGRPCHost,
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Fatal("cannot listen to gRPC server", zap.String("server.host", NotificationGRPCHost), zap.Error(err))
	}
	grpcNotificationClient := notificationpb.NewNotificationServiceClient(conn)

	notificationHandler := notification.NewNotificationHandler(grpcNotificationClient)
	httpMux.Post("/notifications/push", func(c *fiber.Ctx) error {
		ctx, span := telemetry.StartSpan(c.UserContext(), "controller:PushNotification")
		defer span.End()

		var req notification.PushNotificationRequest
		if err := c.BodyParser(&req); err != nil {
			logger.Error("unable to unmarshal body", zap.ByteString("body", c.BodyRaw()), zap.Error(err))
			return err
		}

		if err := notificationHandler.SendPushNotification(ctx, req); err != nil {
			logger.Error("unable to send notification", zap.Error(err))
			return err
		}

		return nil
	})

	// run http server
	go func() {
		if err := httpMux.Listen(":8080"); err != nil {
			logger.Fatal("unable to run http server", zap.Error(err))
		}
	}()

	<-ctx.Done()

	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Duration(30*time.Second))
	defer cancel()

	if err := httpMux.ShutdownWithContext(timeoutCtx); err != nil {
		logger.Fatal("http server cannot shutdown gracefully", zap.Error(err))
	}

	time.Sleep(time.Duration(10 * time.Second))
	logger.Info("http service shutting down gracefully")
}
