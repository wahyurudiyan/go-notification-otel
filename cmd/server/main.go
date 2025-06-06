package main

import (
	"context"
	"errors"

	"github.com/wahyurudiyan/go-otel-context-propagation/pkg/graceful"
	"github.com/wahyurudiyan/go-otel-context-propagation/pkg/telemetry"
	"go.uber.org/zap"
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

	if err := graceful.Runner(ctx, wrapServer); err != nil {
		zap.L().Fatal("Server cannot shutting down gracefully!", zap.Error(err))
	}
}
