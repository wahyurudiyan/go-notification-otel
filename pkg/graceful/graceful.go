package graceful

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

type ShutdownCallback func(context.Context) error

func Runner(ctx context.Context, callback func(ctx context.Context) ShutdownCallback) (err error) {
	logger := zap.L()
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGKILL, syscall.SIGQUIT)
	defer stop()

	shutdown := callback(ctx)
	<-ctx.Done()

	if err = shutdown(ctx); err != nil {
		return
	}

	time.Sleep(time.Duration(10 * time.Second))
	logger.Info("system shutting down gracefully")
	return
}
