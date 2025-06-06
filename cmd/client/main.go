package main

import (
	"context"
	"errors"
	"os/signal"
	"syscall"

	"github.com/wahyurudiyan/go-otel-context-propagation/pkg/graceful"
	"github.com/wahyurudiyan/go-otel-context-propagation/pkg/telemetry"
	"go.uber.org/zap"
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

	if err := graceful.Runner(ctx, boostrap); err != nil {
		zap.L().Fatal("Server cannot shutdown gracefully", zap.Error(err))
	}
}
