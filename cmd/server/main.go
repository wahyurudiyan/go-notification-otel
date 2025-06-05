package main

import (
	"context"
	"errors"
	"net"
	"sync"

	"github.com/gofiber/contrib/otelfiber/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/wahyurudiyan/go-otel-context-propagation/cmd/server/notification"
	"github.com/wahyurudiyan/go-otel-context-propagation/contract/notificationpb"
	"github.com/wahyurudiyan/go-otel-context-propagation/pkg/graceful"
	"github.com/wahyurudiyan/go-otel-context-propagation/pkg/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const (
	HTTPServerAddr = "0.0.0.0:8080"
	GRPCServerAddr = "0.0.0.0:9090"
)

func init() {
	zapConfig := zap.NewDevelopmentConfig()
	zap.ReplaceGlobals(zap.Must(zapConfig.Build()))
}

func main() {
	zap.L().Info("Starting GRPC server")

	// initialize graceful shutdown
	ctx := context.Background()

	shutdown, err := telemetry.SetupOTelSDK(ctx, "server.grpc")
	if err != nil {
		zap.L().Fatal("unable to setup OTelSDK", zap.Error(err))
	}

	defer func() {
		err = errors.Join(err, shutdown(context.Background()))
		if err != nil {
			zap.L().Error("Open Telemetry shutdown failed!", zap.Error(err))
		}
	}()

	if err := graceful.Runner(ctx, bootstrapServer); err != nil {
		zap.L().Fatal("Server cannot shutting down gracefully!", zap.Error(err))
	}
}

func bootstrapServer(ctx context.Context) graceful.ShutdownCallback {
	httpServer := startHTTPServer()
	grpcServer := startGRPCServer()
	return func(ctx context.Context) error {
		var wg sync.WaitGroup
		chanErr := make(chan error)

		go func(ctx context.Context) {
			wg.Add(1)
			defer wg.Done()

			if err := httpServer.ShutdownWithContext(ctx); err != nil {
				chanErr <- err
			}
		}(context.Background())

		go func() {
			wg.Add(1)
			defer wg.Done()

			grpcServer.GracefulStop()
		}()
		wg.Wait()

		select {
		case err := <-chanErr:
			return err
		case <-ctx.Done():
			zap.L().Warn("Shutting down timeout exceeded!")
			return nil
		}
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

	handler := notification.NewNotificationHTTPHandler()
	mux.Post("/notifications/email", handler.SendEmailNotification())

	zap.L().Info("HTTP server is running!", zap.String("http.address", HTTPServerAddr))
	go func() {
		if err := mux.Listen(HTTPServerAddr); err != nil {
			zap.L().Fatal("Cannot start HTTP server!", zap.Error(err))
		}
	}()

	return mux
}

func startGRPCServer() *grpc.Server {
	// initialize tcp network listener
	lst, err := net.Listen("tcp", GRPCServerAddr)
	if err != nil {
		zap.L().Fatal("Failed to listen tcp", zap.Error(err))
	}

	// initialize grpc server
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

	// initialize handler and register into server
	notificationGRPCHandler := notification.NewNotificationGRPCHandler()
	notificationpb.RegisterNotificationServiceServer(grpcServer, notificationGRPCHandler)

	// run grpc server
	zap.L().Info("GRPC server is running!", zap.String("grpc.address", GRPCServerAddr))
	go func() {
		if err := grpcServer.Serve(lst); err != nil {
			zap.L().Fatal("Cannot start GRPC server", zap.Error(err))
		}
	}()

	return grpcServer
}
