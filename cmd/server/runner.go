package main

import (
	"context"
	"net"
	"sync"

	"github.com/gofiber/contrib/otelfiber/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/wahyurudiyan/go-otel-context-propagation/cmd/server/notification"
	"github.com/wahyurudiyan/go-otel-context-propagation/contract/notificationpb"
	"github.com/wahyurudiyan/go-otel-context-propagation/pkg/graceful"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func wrapServer(ctx context.Context) graceful.ShutdownCallback {
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
			return nil
		}
	}
}

func startHTTPServer() *fiber.App {
	mux := fiber.New()
	mux.Use(otelfiber.Middleware(otelfiber.WithCustomAttributes(
		func(ctx *fiber.Ctx) []attribute.KeyValue {
			return []attribute.KeyValue{
				semconv.ServiceName("notification:server"),
			}
		},
	)))
	router := mux.Group("/server")

	handler := notification.NewNotificationHTTPHandler()
	router.Post("/notifications/email", handler.SendEmailNotification())

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
