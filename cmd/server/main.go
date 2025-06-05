package main

import (
	"context"
	"errors"
	"net"
	"os/signal"
	"syscall"
	"time"

	"github.com/wahyurudiyan/go-notification-otel/cmd/server/notification"
	"github.com/wahyurudiyan/go-notification-otel/contract/notificationpb"
	"github.com/wahyurudiyan/go-notification-otel/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func init() {
	zapConfig := zap.NewDevelopmentConfig()
	zap.ReplaceGlobals(zap.Must(zapConfig.Build()))
}

func main() {
	var logger = zap.L()
	logger.Info("Starting GRPC server")

	// initialize graceful shutdown
	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGKILL, syscall.SIGQUIT)
	defer stop()

	shutdown, err := telemetry.SetupOTelSDK(ctx, "server.grpc")
	if err != nil {
		logger.Fatal("unable to setup OTelSDK", zap.Error(err))
	}

	defer func() {
		err = errors.Join(err, shutdown(context.Background()))
	}()

	// initialize tcp network listener
	addr := "0.0.0.0:9090"
	lst, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatal("failed to listen tcp", zap.Error(err))
	}

	// initialize grpc server
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

	// initialize handler and register into server
	notificationHandler := notification.NewNotificationHandler()
	notificationpb.RegisterNotificationServiceServer(grpcServer, notificationHandler)

	// run grpc server
	logger.Info("GRPC server is running!", zap.String("address", addr))
	go func() {
		if err := grpcServer.Serve(lst); err != nil {
			logger.Fatal("unable to run grpc server", zap.Error(err))
		}
	}()

	<-ctx.Done()
	grpcServer.GracefulStop()

	time.Sleep(time.Duration(10 * time.Second))
	logger.Info("grpc service shutting down gracefully")
}
